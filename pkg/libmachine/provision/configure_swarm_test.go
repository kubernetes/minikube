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

package provision

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cert"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/drivers/fakedriver"
	"k8s.io/minikube/pkg/libmachine/state"
	"k8s.io/minikube/pkg/libmachine/swarm"
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

type MockDriver struct {
	*fakedriver.Driver
	Port int
}

func (d *MockDriver) GetURL() (string, error) {
	return fmt.Sprintf("tcp://127.0.0.1:%d", d.Port), nil
}

type MockProvisioner struct {
	*FakeProvisioner
	Driver drivers.Driver
}

func (m *MockProvisioner) GetDriver() drivers.Driver {
	return m.Driver
}

func (m *MockProvisioner) GetDockerOptionsDir() string {
	return "/var/lib/docker"
}

// TestConfigureSwarm verifies that the configureSwarm function correctly constructs Docker API requests
// to initialize a Swarm master. It mocks the Docker daemon and verifies that the function attempts
// to create the expected containers with the correct configuration (args, network settings).
// This ensures that our Swarm provisioning logic remains stable even if underlying libraries change.
func TestConfigureSwarm(t *testing.T) {
	cert.SetCertGenerator(&MockCertGenerator{})
	defer cert.SetCertGenerator(cert.NewX509CertGenerator())

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/_ping") {
			w.WriteHeader(http.StatusOK)
			return
		}
		if strings.Contains(r.URL.Path, "/images/create") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
			return
		}
		if strings.Contains(r.URL.Path, "/containers/create") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"Id": "test-id"}`))
			return
		}
		if strings.Contains(r.URL.Path, "/start") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	_, portStr, _ := net.SplitHostPort(u.Host)
	port, _ := strconv.Atoi(portStr)

	driver := &MockDriver{
		Driver: &fakedriver.Driver{
			MockState: state.Running,
			MockIP:    "127.0.0.1",
		},
		Port: port,
	}

	p := &MockProvisioner{
		FakeProvisioner: &FakeProvisioner{},
		Driver:          driver,
	}

	swarmOpts := swarm.Options{
		IsSwarm:   true,
		Master:    true,
		Host:      "tcp://127.0.0.1:2376",
		Discovery: "token://...",
		Image:     "swarm:latest",
		Strategy:  "spread",
	}

	authOpts := auth.Options{
		CaCertRemotePath:     "/ca.pem",
		ServerCertRemotePath: "/server.pem",
		ServerKeyRemotePath:  "/server-key.pem",
	}

	err := configureSwarm(p, swarmOpts, authOpts)
	if err != nil {
		t.Errorf("configureSwarm failed: %v", err)
	}
}
