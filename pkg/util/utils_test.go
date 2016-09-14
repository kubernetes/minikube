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

package util

import (
	"testing"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
)

// Returns a function that will return n errors, then return successfully forever.
func errorGenerator(n int) func() error {
	errorCount := 0
	return func() (err error) {
		if errorCount < n {
			errorCount += 1
			return errors.New("Error!")
		}
		return nil
	}
}

func TestErrorGenerator(t *testing.T) {
	errors := 3
	f := errorGenerator(errors)
	for i := 0; i < errors-1; i++ {
		if err := f(); err == nil {
			t.Fatalf("Error should have been thrown at iteration %v", i)
		}
	}
	if err := f(); err == nil {
		t.Fatalf("Error should not have been thrown this call!")
	}
}

func TestRetry(t *testing.T) {

	f := errorGenerator(4)
	if err := Retry(5, f); err != nil {
		t.Fatalf("Error should not have been raised by retry.")
	}

	f = errorGenerator(5)
	if err := Retry(4, f); err == nil {
		t.Fatalf("Error should have been raised by retry.")
	}
}

type getLocalkubeArgs struct {
	input         string
	expected      string
	expectedError bool
}

func TestGetLocalkubeDownloadURL(t *testing.T) {
	argsList := [...]getLocalkubeArgs{
		{"v1.3.0",
			"https://storage.googleapis.com/minikube/k8sReleases/v1.3.0/localkube-linux-amd64", false},
		{"v1.3.3",
			"https://storage.googleapis.com/minikube/k8sReleases/v1.3.3/localkube-linux-amd64", false},
		{"http://www.example.com/my-localkube", "http://www.example.com/my-localkube", false},
		{"abc", "", true},
		{"1.2.3.4", "", true},
	}
	for _, args := range argsList {
		url, err := GetLocalkubeDownloadURL(args.input, constants.LocalkubeLinuxFilename)
		wasError := err != nil
		if wasError != args.expectedError {
			t.Errorf("GetLocalkubeDownloadURL Expected error was: %t, Actual Error was: %s",
				args.expectedError, err)
		}
		if url != args.expected {
			t.Errorf("GetLocalkubeDownloadURL: Expected %s, Actual: %s", args.expected, url)
		}
	}
}

func TestMultiError(t *testing.T) {
	m := MultiError{}

	m.Collect(errors.New("Error 1"))
	m.Collect(errors.New("Error 2"))

	err := m.ToError()
	expected := `Error 1
Error 2`
	if err.Error() != expected {
		t.Fatalf("%s != %s", err, expected)
	}

	m = MultiError{}
	if err := m.ToError(); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}
