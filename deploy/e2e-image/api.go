/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const (
	serviceAccountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
	hostVar            = "KUBERNETES_SERVICE_HOST"
	portVar            = "KUBERNETES_PORT_443_TCP_PORT"
)

var (
	tokenPath = filepath.Join(serviceAccountPath, "token")
)

func main() {
	host := os.Getenv(hostVar)
	port := os.Getenv(portVar)
	address := fmt.Sprintf("https://%s:%s", host, port)

	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading token: %s", err)
		os.Exit(1)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %s", err)
		os.Exit(1)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %s", err)
		os.Exit(1)
	}
	fmt.Println(string(body))

}
