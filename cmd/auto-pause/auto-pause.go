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
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

func main() {
	http.HandleFunc("/", handler) // each request calls handler
	fmt.Printf("Starting server at port 8080\n")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

// handler echoes the Path component of the requested URL.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Receive request uri %s at port 8080\n", r.RequestURI)
	out, err := exec.Command("docker", "ps").Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("docker ps output:\n%s\n", string(out))
	fmt.Fprintf(w, "allow")
}
