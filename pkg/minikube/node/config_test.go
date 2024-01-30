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

package node

import (
	"testing"
)

func Test_maskProxyPassword(t *testing.T) {
	type dockerOptTest struct {
		input  string
		output string
	}
	var tests = []dockerOptTest{
		{
			input:  "cats",
			output: "cats",
		},
		{
			input:  "myDockerOption=value",
			output: "myDockerOption=value",
		},
		{
			input:  "http_proxy=http://minikube.sigs.k8s.io",
			output: "HTTP_PROXY=http://minikube.sigs.k8s.io",
		},
		{
			input:  "https_proxy=http://jdoe@minikube.sigs.k8s.io:8080",
			output: "HTTPS_PROXY=http://jdoe@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "https_proxy=https://mary:iam$Fake!password@minikube.sigs.k8s.io:8080",
			output: "HTTPS_PROXY=https://mary:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http_proxy=http://jdoe:%n0tRe@al:Password!@minikube.sigs.k8s.io:8080",
			output: "HTTP_PROXY=http://jdoe:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http_proxy=http://jo@han:n0tRe@al:&Password!@minikube.sigs.k8s.io:8080",
			output: "HTTP_PROXY=http://jo@han:*****@minikube.sigs.k8s.io:8080",
		},
		{
			input:  "http_proxy=http://k@r3n!:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "HTTP_PROXY=http://k@r3n!:*****@minikube.sigs.k8s.io",
		},
		{
			input:  "https_proxy=https://fr@ank5t3in:an0th3erF@akeP@55word@minikube.sigs.k8s.io",
			output: "HTTPS_PROXY=https://fr@ank5t3in:*****@minikube.sigs.k8s.io",
		},
	}
	for _, test := range tests {
		got := maskProxyPassword(test.input)
		if got != test.output {
			t.Errorf("maskProxyPassword(\"%v\"): got %v, expected %v", test.input, got, test.output)
		}
	}
}
