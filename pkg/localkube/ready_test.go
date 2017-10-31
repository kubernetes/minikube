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
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/util"
)

func TestBasicHealthCheck(t *testing.T) {

	tcs := []struct {
		body          string
		statusCode    int
		shouldSucceed bool
	}{
		{"ok", 200, true},
		{"notok", 200, false},
	}

	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)
	_, ipnet, err := net.ParseCIDR(util.DefaultServiceCIDR)
	if err != nil {
		t.Fatalf("Error parsing default service cidr range: %s", err)
	}
	lk := LocalkubeServer{
		LocalkubeDirectory:    tempDir,
		ServiceClusterIPRange: *ipnet,
	}
	lk.GenerateCerts()

	cert, err := tls.LoadX509KeyPair(lk.GetPublicKeyCertPath(), lk.GetPrivateKeyCertPath())
	if err != nil {
		t.Fatalf("Unable to load server certs.")
	}

	caCert, err := ioutil.ReadFile(lk.GetCAPublicKeyCertPath())
	if err != nil {
		t.Fatalf("Unable to load CA certs.")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tls := tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
	}

	tls.BuildNameToCertificate()

	for _, tc := range tcs {
		// Do this in a func so we can use defer.
		doTest := func() {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				io.WriteString(w, tc.body)
			}
			server := httptest.NewUnstartedServer(http.HandlerFunc(handler))
			defer server.Close()
			server.TLS = &tls
			server.StartTLS()

			hcFunc := healthCheck(server.URL, lk)
			result := hcFunc()
			if result != tc.shouldSucceed {
				t.Errorf("Expected healthcheck to return %v. Got %v", result, tc.shouldSucceed)
			}
		}
		doTest()
	}
}
