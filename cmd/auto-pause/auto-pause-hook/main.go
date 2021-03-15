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
	"strings"

	"github.com/mattbaird/jsonpatch"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// TODO(https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

// the Path of the JSON patch is a JSON pointer value
// so we need to escape any "/"s in the key we add to the annotation
// https://tools.ietf.org/html/rfc6901
func escapeJSONPointer(s string) string {
	esc := strings.Replace(s, "~", "~0", -1)
	esc = strings.Replace(esc, "/", "~1", -1)
	return esc
}

var minikubeSystemNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
	"auto-pause",
}

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
		admResp.Response = getAdmissionDecision(&admReq)
	}

	admResp.APIVersion = "admission.k8s.io/v1"
	admResp.Kind = "AdmissionReview"
	resp, err := json.Marshal(admResp)
	if err != nil {
		log.Printf("error marshalling decision: %v", err)
	}
	log.Printf(string(resp))
	if _, err := w.Write(resp); err != nil {
		log.Printf("error writing response %v", err)
	}
}

func admissionError(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{Message: err.Error()},
	}
}

func getAdmissionDecision(admReq *v1.AdmissionReview) *v1.AdmissionResponse {
	req := admReq.Request
	var pod corev1.Pod

	err := json.Unmarshal(req.Object.Raw, &pod)
	if err != nil {
		log.Printf("Could not unmarshal raw object: %v", err)
		return admissionError(err)
	}

	log.Printf("AdmissionReview for Kind=%v Namespace=%v Name=%v UID=%v Operation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	if !shouldInject(&pod.ObjectMeta) {
		log.Printf("Skipping inject for %s %s", pod.Namespace, pod.Name)
		return &v1.AdmissionResponse{
			Allowed: true,
			UID:     req.UID,
		}
	}

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
		{Name: "KUBERNETES_SERVICE_HOST", Value: "192.168.49.2"},
		{Name: "KUBERNETES_SERVICE_PORT", Value: "32443"}}
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
			path = path + "/-"
		}
		patch = append(patch, jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      path,
			Value:     value,
		})
	}
	return patch
}

func shouldInject(metadata *metav1.ObjectMeta) bool {
	shouldInject := true

	// don't attempt to inject pods in the Kubernetes system namespaces
	for _, ns := range minikubeSystemNamespaces {
		if metadata.Namespace == ns {
			shouldInject = false
		}
	}

	return shouldInject
}

func main() {
	addr := flag.String("addr", ":8080", "address to serve on")

	http.HandleFunc("/", handler)

	flag.CommandLine.Parse([]string{}) // hack fix for https://github.com/kubernetes/kubernetes/issues/17162

	log.Printf("Starting HTTPS webhook server on %+v", *addr)
	cacert, serverCert, serverKey := gencerts()
	clientset := getClient()
	server := &http.Server{
		Addr:      *addr,
		TLSConfig: configTLS(clientset, serverCert, serverKey),
	}
	go selfRegistration(clientset, cacert)
	server.ListenAndServeTLS("", "")
}
