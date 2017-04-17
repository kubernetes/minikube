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

package localkube

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicHealthCheck(t *testing.T) {

	tests := []struct {
		body          string
		statusCode    int
		shouldSucceed bool
	}{
		{"ok", 200, true},
		{"notok", 200, false},
	}

	for _, tc := range tests {
		// Do this in a func so we can use defer.
		doTest := func() {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				io.WriteString(w, tc.body)
			}
			server := httptest.NewServer(http.HandlerFunc(handler))
			defer server.Close()

			hcFunc := healthCheck(server.URL)
			result := hcFunc()
			if result != tc.shouldSucceed {
				t.Errorf("Expected healthcheck to return %v. Got %v", result, tc.shouldSucceed)
			}
		}
		doTest()
	}
}
