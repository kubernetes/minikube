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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/minikube/pkg/version"

	"github.com/pkg/errors"
)

func TestFormatError(t *testing.T) {
	var testErr error
	if _, err := FormatError(testErr); err == nil {
		t.Fatalf("FormatError should have errored with a nil error input")
	}
	testErr = fmt.Errorf("Not a valid error to format as there is no stacktrace")

	if out, err := FormatError(testErr); err == nil {
		t.Fatalf("FormatError should have errored with a non pkg/errors error (no stacktrace info): %s", out)
	}

	testErr = errors.New("TestFormatError 1")
	errors.Wrap(testErr, "TestFormatError 2")
	errors.Wrap(testErr, "TestFormatError 3")

	_, err := FormatError(testErr)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestMarshallError(t *testing.T) {
	testErr := errors.New("TestMarshallError 1")
	errors.Wrap(testErr, "TestMarshallError 2")
	errors.Wrap(testErr, "TestMarshallError 3")

	errMsg, _ := FormatError(testErr)
	if _, err := MarshallError(errMsg, "default", version.GetVersion()); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestUploadError(t *testing.T) {
	testErr := errors.New("TestUploadError 1")
	errors.Wrap(testErr, "TestUploadError 2")
	errors.Wrap(testErr, "TestUploadError 3")
	errMsg, _ := FormatError(testErr)
	jsonErrMsg, _ := MarshallError(errMsg, "default", version.GetVersion())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!")
	}))

	if err := UploadError(jsonErrMsg, server.URL); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed to write report", 400)
	}))
	if err := UploadError(jsonErrMsg, server.URL); err == nil {
		t.Fatalf("UploadError should have errored from a 400 response")
	}
}
