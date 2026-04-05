/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package mcndockerclient

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cert"
)

type MockCertGenerator struct{}

func (m *MockCertGenerator) GenerateCACertificate(certFile, keyFile, org string, bits int) error {
	return nil
}

func (m *MockCertGenerator) GenerateCert(opts *cert.Options) error {
	return nil
}

func (m *MockCertGenerator) ReadTLSConfig(addr string, authOptions *auth.Options) (*tls.Config, error) {
	return &tls.Config{InsecureSkipVerify: true}, nil
}

func (m *MockCertGenerator) ValidateCertificate(addr string, authOptions *auth.Options) (bool, error) {
	return true, nil
}

type MockDockerHost struct {
	URLStr string
	AuthOp *auth.Options
}

func (m *MockDockerHost) URL() (string, error) {
	return m.URLStr, nil
}

func (m *MockDockerHost) AuthOptions() *auth.Options {
	return m.AuthOp
}

// TestDockerClient verifies that we can create a client and communicate with a (mock) docker daemon.
// This is important to ensure our client setup (TLS, transport) matches what the official SDK expects
// and that we can successfully negotiate the API version.
func TestDockerClient(t *testing.T) {
	cert.SetCertGenerator(&MockCertGenerator{})
	defer cert.SetCertGenerator(cert.NewX509CertGenerator())

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/version") {
			w.Header().Set("Content-Type", "application/json")
			// Return a version response compatible with Docker API
			w.Write([]byte(`{"Version": "1.2.3", "ApiVersion": "1.41"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	host := &MockDockerHost{
		URLStr: server.URL,
		AuthOp: &auth.Options{},
	}

	client, err := DockerClient(host)
	if err != nil {
		t.Fatalf("DockerClient failed: %v", err)
	}

	v, err := client.ServerVersion(context.Background())
	if err != nil {
		t.Fatalf("ServerVersion failed: %v", err)
	}

	if v.Version != "1.2.3" {
		t.Errorf("Expected version 1.2.3, got %s", v.Version)
	}
}

// TestCreateContainer verifies that we can pull an image and create/start a container using our helper function.
// It uses a mock HTTP server to simulate the Docker Engine, allowing us to test the entire flow
// (pull -> create -> start) without needing a real Docker daemon installed or running.
func TestCreateContainer(t *testing.T) {
	cert.SetCertGenerator(&MockCertGenerator{})
	defer cert.SetCertGenerator(cert.NewX509CertGenerator())

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle Ping
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Mock Pull
		if strings.Contains(r.URL.Path, "/images/create") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
			return
		}
		// Mock Create
		if strings.Contains(r.URL.Path, "/containers/create") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"Id": "test-container-id"}`))
			return
		}
		// Mock Start
		if strings.Contains(r.URL.Path, "/containers/test-container-id/start") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		t.Logf("Unexpected request: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	}))
	defer server.Close()

	host := &MockDockerHost{
		URLStr: server.URL,
		AuthOp: &auth.Options{},
	}

	config := &container.Config{Image: "test-image"}
	hostConfig := &container.HostConfig{}

	err := CreateContainer(host, config, hostConfig, "test-name")
	if err != nil {
		t.Errorf("CreateContainer failed: %v", err)
	}
}
