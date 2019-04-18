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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/pkg/errors"
)

// Returns a function that will return n errors, then return successfully forever.
func errorGenerator(n int, retryable bool) func() error {
	errorCount := 0
	return func() (err error) {
		if errorCount < n {
			errorCount++
			e := errors.New("Error")
			if retryable {
				return &RetriableError{Err: e}
			}
			return e

		}

		return nil
	}
}

func TestErrorGenerator(t *testing.T) {
	errors := 3
	f := errorGenerator(errors, false)
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
	f := errorGenerator(4, true)
	if err := Retry(5, f); err != nil {
		t.Fatalf("Error should not have been raised by retry.")
	}

	f = errorGenerator(5, true)
	if err := Retry(4, f); err == nil {
		t.Fatalf("Error should have been raised by retry.")
	}
}

func TestRetryNotRetriableError(t *testing.T) {
	f := errorGenerator(4, false)
	if err := Retry(5, f); err == nil {
		t.Fatalf("Error should have been raised by retry.")
	}

	f = errorGenerator(5, false)
	if err := Retry(4, f); err == nil {
		t.Fatalf("Error should have been raised by retry.")
	}
	f = errorGenerator(0, false)
	if err := Retry(5, f); err != nil {
		t.Fatalf("Error should not have been raised by retry.")
	}
}

type getTestArgs struct {
	input         string
	expected      string
	expectedError bool
}

var testSHAString = "test"

func TestParseSHAFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, testSHAString)
	}))
	serverBadResponse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 HTTP status code returned!"))
	}))

	argsList := [...]getTestArgs{
		{server.URL, testSHAString, false},
		{serverBadResponse.URL, "", true},
		{"abc", "", true},
	}
	for _, args := range argsList {
		url, err := ParseSHAFromURL(args.input)
		wasError := err != nil
		if wasError != args.expectedError {
			t.Errorf("ParseSHAFromURL Expected error was: %t, Actual Error was: %s",
				args.expectedError, err)
		}
		if url != args.expected {
			t.Errorf("ParseSHAFromURL: Expected %s, Actual: %s", args.expected, url)
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
		t.Fatalf("%s != %s", err.Error(), expected)
	}

	m = MultiError{}
	if err := m.ToError(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGetBinaryDownloadURL(t *testing.T) {
	testData := []struct {
		version     string
		platform    string
		expectedURL string
	}{
		{"v0.0.1", "linux", "https://storage.googleapis.com/minikube/releases/v0.0.1/minikube-linux-amd64"},
		{"v0.0.1", "darwin", "https://storage.googleapis.com/minikube/releases/v0.0.1/minikube-darwin-amd64"},
		{"v0.0.1", "windows", "https://storage.googleapis.com/minikube/releases/v0.0.1/minikube-windows-amd64.exe"},
	}

	for _, tt := range testData {
		url := GetBinaryDownloadURL(tt.version, tt.platform)
		if url != tt.expectedURL {
			t.Fatalf("Expected '%s' but got '%s'", tt.expectedURL, url)
		}
	}

}

func TestTeePrefix(t *testing.T) {
	var in bytes.Buffer
	var out bytes.Buffer
	var logged strings.Builder

	logSink := func(format string, args ...interface{}) {
		logged.WriteString("(" + fmt.Sprintf(format, args...) + ")")
	}

	// Simulate the primary use case: tee in the background. This also helps avoid I/O races.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		TeePrefix(":", &in, &out, logSink)
		wg.Done()
	}()

	in.Write([]byte("goo"))
	in.Write([]byte("\n"))
	in.Write([]byte("g\r\n\r\n"))
	in.Write([]byte("le"))
	wg.Wait()

	gotBytes := out.Bytes()
	wantBytes := []byte("goo\ng\r\n\r\nle")
	if !bytes.Equal(gotBytes, wantBytes) {
		t.Errorf("output=%q, want: %q", gotBytes, wantBytes)
	}

	gotLog := logged.String()
	wantLog := "(:goo)(:g)(:le)"
	if gotLog != wantLog {
		t.Errorf("log=%q, want: %q", gotLog, wantLog)
	}
}

func TestReplaceChars(t *testing.T) {
	testData := []struct {
		src         []string
		replacer    *strings.Replacer
		expectedRes []string
	}{
		{[]string{"abc%def", "%Y%"}, strings.NewReplacer("%", "X"), []string{"abcXdef", "XYX"}},
	}

	for _, tt := range testData {
		res := ReplaceChars(tt.src, tt.replacer)
		for i, val := range res {
			if val != tt.expectedRes[i] {
				t.Fatalf("Expected '%s' but got '%s'", tt.expectedRes, res)
			}
		}
	}
}

func TestConcatStrings(t *testing.T) {
	testData := []struct {
		src         []string
		prefix      string
		postfix     string
		expectedRes []string
	}{
		{[]string{"abc", ""}, "xx", "yy", []string{"xxabcyy", "xxyy"}},
		{[]string{"abc", ""}, "", "", []string{"abc", ""}},
	}

	for _, tt := range testData {
		res := ConcatStrings(tt.src, tt.prefix, tt.postfix)
		for i, val := range res {
			if val != tt.expectedRes[i] {
				t.Fatalf("Expected '%s' but got '%s'", tt.expectedRes, res)
			}
		}
	}
}
