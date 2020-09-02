/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package out

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
)

type buffFd struct {
	bytes.Buffer
	uptr uintptr
}

func (b buffFd) Fd() uintptr { return b.uptr }

func TestDisplayProblem(t *testing.T) {
	buffErr := buffFd{}
	SetErrFile(&buffErr)
	tests := []struct {
		description string
		issue       reason.Kind
		expected    string
	}{
		{
			issue: reason.Kind{
				ID:  "example",
				URL: "example.com",
			},
			description: "url, id and err",
			expected: `X Something failed
* Documentation: example.com
`,
		},
		{
			issue:       reason.Kind{ID: "example", URL: "example.com", Issues: []int{0, 1}, Advice: "you need a hug"},
			description: "with 2 issues and suggestion",
			expected: `X Something failed
* Suggestion: you need a hug
* Documentation: example.com
* Related issues:
  - https://github.com/kubernetes/minikube/issues/0
  - https://github.com/kubernetes/minikube/issues/1
`,
		},
		{
			issue:       reason.Kind{ID: "example", URL: "example.com", Issues: []int{0, 1}},
			description: "with 2 issues",
			expected: `X Something failed
* Documentation: example.com
* Related issues:
  - https://github.com/kubernetes/minikube/issues/0
  - https://github.com/kubernetes/minikube/issues/1
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			buffErr.Truncate(0)
			JSON = false
			Error(tc.issue, "Something failed")
			errStr := buffErr.String()
			if strings.TrimSpace(errStr) != strings.TrimSpace(tc.expected) {
				t.Fatalf("Expected errString:\n%v\ngot:\n%v\n", tc.expected, errStr)
			}
		})
	}
}

func TestDisplayJSON(t *testing.T) {
	defer SetJSON(false)
	SetJSON(true)

	tcs := []struct {
		k        *reason.Kind
		expected string
	}{
		{
			k: &reason.Kind{

				ID:       "BUG",
				ExitCode: 4,
				Advice:   "fix me!",
				Issues:   []int{1, 2},
				URL:      "url",
			},
			expected: `{"data":{"advice":"fix me!","exitcode":"4","issues":"https://github.com/kubernetes/minikube/issues/1,https://github.com/kubernetes/minikube/issues/2","message":"my error","name":"BUG","url":"url"},"datacontenttype":"application/json","id":"random-id","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.error"}
`,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.k.ID, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			register.SetOutputFile(buf)
			defer func() { register.SetOutputFile(os.Stdout) }()

			register.GetUUID = func() string {
				return "random-id"
			}

			JSON = true
			Error(*tc.k, "my error")
			actual := buf.String()
			if actual != tc.expected {
				t.Fatalf("expected didn't match actual:\nExpected:\n%v\n\nActual:\n%v", tc.expected, actual)
			}
		})
	}
}
