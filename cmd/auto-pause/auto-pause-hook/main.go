/*
Copyright 2021 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/mattbaird/jsonpatch"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/constants"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

var targetIP *string

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling a request")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Printf("Wrong content type. Got: %s", contentType)
		return
	}

	admReq := v1.AdmissionReview{}
	admResp := v1.AdmissionReview{}

	if _, _, err := deserializer.Decode(body, nil, &admReq); err != nil {
		log.Printf("Could not decode body: %v", err)
		admResp.Response = admissionError(err)
	} else {
		admResp.Response = AdmissionDecision(&admReq)
	}

	admResp.APIVersion = "admission.k8s.io/v1"
	admResp.Kind = "AdmissionReview"
	resp, err := json.Marshal(admResp)
	if err != nil {
		log.Printf("error marshalling decision: %v", err)
	}
	log.Printf("%s", string(resp))
	if _, err := w.Write(resp); err != nil {
		log.Printf("error writing response %v", err)
	}
}

func admissionError(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{Message: err.Error()},
	}
}

// AdmissionDecision creates the admission decision for the request
func AdmissionDecision(admReq *v1.AdmissionReview) *v1.AdmissionResponse {
	req := admReq.Request
	var pod corev1.Pod

	err := json.Unmarshal(req.Object.Raw, &pod)
	if err != nil {
		log.Printf("Could not unmarshal raw object: %v", err)
		return admissionError(err)
	}

	log.Printf("AdmissionReview for Kind=%v Namespace=%v Name=%v UID=%v Operation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	patch, err := patchConfig(&pod)

	if err != nil {
		log.Printf("Error creating conduit patch: %v", err)
		return admissionError(err)
	}

	jsonPatchType := v1.PatchTypeJSONPatch

	return &v1.AdmissionResponse{
		Allowed:   true,
		Patch:     patch,
		PatchType: &jsonPatchType,
		UID:       req.UID,
	}
}

func patchConfig(pod *corev1.Pod) ([]byte, error) {
	var patch []jsonpatch.JsonPatchOperation

	configEnv := []corev1.EnvVar{
		{Name: "KUBERNETES_SERVICE_HOST", Value: *targetIP},
		{Name: "KUBERNETES_SERVICE_PORT", Value: strconv.Itoa(constants.AutoPauseProxyPort)}}
	for idx, container := range pod.Spec.Containers {
		patch = append(patch, addEnv(container.Env, configEnv, fmt.Sprintf("/spec/containers/%d/env", idx))...)
	}
	return json.Marshal(patch)
}

// addEnv performs the mutation(s) needed to add the extra environment variables to the target
// resource
func addEnv(target, envVars []corev1.EnvVar, basePath string) (patch []jsonpatch.JsonPatchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, envVar := range envVars {
		value = envVar
		path := basePath
		if first {
			first = false
			value = []corev1.EnvVar{envVar}
		} else {
			path += "/-"
		}
		patch = append(patch, jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      path,
			Value:     value,
		})
	}
	return patch
}

func main() {
	addr := flag.String("addr", ":8080", "address to serve on")
	targetIP = flag.String("targetIP", "192.168.49.2", "The reverse proxy IP")

	http.HandleFunc("/", handler)

	flag.Parse()

	log.Printf("Starting HTTPS webhook server on %+v and target ip is %v", *addr, *targetIP)
	cacert, serverCert, serverKey := gencerts()
	clientset := client()
	server := &http.Server{
		Addr:      *addr,
		TLSConfig: configTLS(clientset, serverCert, serverKey),
	}
	go selfRegistration(clientset, cacert)
	err := server.ListenAndServeTLS("", "")
	if err != nil {
		klog.Fatalf("Start https server failed with %s", err)
	}
}
