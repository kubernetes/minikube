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

package kubeconfig

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/minikube/pkg/minikube/constants"

	"k8s.io/client-go/tools/clientcmd"
)

var kubeConfigWithoutHTTPS = []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/la-croix/apiserver.crt
    server: 192.168.1.1:8080
  name: la-croix
contexts:
- context:
    cluster: la-croix
    user: la-croix
  name: la-croix
current-context: la-croix
kind: Config
preferences: {}
users:
- name: la-croix
  user:
    client-certificate: /home/la-croix/apiserver.crt
    client-key: /home/la-croix/apiserver.key
`)

var kubeConfig192 = []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/la-croix/apiserver.crt
    server: https://192.168.10.100:8443
  name: minikube
contexts:
- context:
    cluster: la-croix
    user: la-croix
  name: la-croix
current-context: la-croix
kind: Config
preferences: {}
users:
- name: la-croix
  user:
    client-certificate: /home/la-croix/apiserver.crt
    client-key: /home/la-croix/apiserver.key
`)

var kubeConfigLocalhost = []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/la-croix/apiserver.crt
    server: https://127.0.0.1:8443
  name: minikube
contexts:
- context:
    cluster: la-croix
    user: la-croix
  name: la-croix
current-context: la-croix
kind: Config
preferences: {}
users:
- name: la-croix
  user:
    client-certificate: /home/la-croix/apiserver.crt
    client-key: /home/la-croix/apiserver.key
`)

var kubeConfigLocalhost12345 = []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/la-croix/apiserver.crt
    server: https://127.0.0.1:12345
  name: minikube
contexts:
- context:
    cluster: la-croix
    user: la-croix
  name: la-croix
current-context: la-croix
kind: Config
preferences: {}
users:
- name: la-croix
  user:
    client-certificate: /home/la-croix/apiserver.crt
    client-key: /home/la-croix/apiserver.key
`)

func TestUpdate(t *testing.T) {
	setupCfg := &Settings{
		ClusterName:          "test",
		ClusterServerAddress: "192.168.1.1:8080",
		ClientCertificate:    "/home/apiserver.crt",
		ClientKey:            "/home/apiserver.key",
		CertificateAuthority: "/home/apiserver.crt",
		KeepContext:          false,
	}

	var tests = []struct {
		description string
		cfg         *Settings
		existingCfg []byte
		expected    api.Config
		err         bool
	}{
		{
			description: "new kube config",
			cfg:         setupCfg,
		},
		{
			description: "add to kube config",
			cfg:         setupCfg,
			existingCfg: kubeConfigWithoutHTTPS,
		},
		{
			description: "use config env var",
			cfg:         setupCfg,
		},
		{
			description: "keep context",
			cfg: &Settings{
				ClusterName:          "test",
				ClusterServerAddress: "192.168.1.1:8080",
				ClientCertificate:    "/home/apiserver.crt",
				ClientKey:            "/home/apiserver.key",
				CertificateAuthority: "/home/apiserver.crt",
				KeepContext:          true,
			},
			existingCfg: kubeConfigWithoutHTTPS,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("Error making temp directory %v", err)
			}
			defer func() { //clean up tempdir
				err := os.RemoveAll(tmpDir)
				if err != nil {
					t.Errorf("failed to clean up temp folder  %q", tmpDir)
				}
			}()

			test.cfg.SetPath(filepath.Join(tmpDir, "kubeconfig"))
			if len(test.existingCfg) != 0 {
				if err := ioutil.WriteFile(test.cfg.filePath(), test.existingCfg, 0600); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			}
			err = Update(test.cfg)
			if err != nil && !test.err {
				t.Errorf("Got unexpected error: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none")
			}
			config, err := readOrNew(test.cfg.filePath())
			if err != nil {
				t.Errorf("Error reading kubeconfig file: %v", err)
			}
			if test.cfg.KeepContext && config.CurrentContext == test.cfg.ClusterName {
				t.Errorf("Context was changed even though KeepContext was true")
			}
			if !test.cfg.KeepContext && config.CurrentContext != test.cfg.ClusterName {
				t.Errorf("Context was not switched")
			}

			os.RemoveAll(tmpDir)
		})

	}
}

func TestVerifyEndpoint(t *testing.T) {

	var tests = []struct {
		description string
		hostname    string
		port        int
		existing    []byte
		err         bool
		status      bool
	}{
		{
			description: "empty hostname",
			hostname:    "",
			port:        8443,
			existing:    kubeConfigWithoutHTTPS,
			err:         true,
		},
		{
			description: "no minikube cluster",
			hostname:    "192.168.10.100",
			port:        8443,
			existing:    kubeConfigWithoutHTTPS,
			err:         true,
		},
		{
			description: "exactly matching hostname/port",
			hostname:    "192.168.10.100",
			port:        8443,
			existing:    kubeConfig192,
			status:      true,
		},
		{
			description: "different hostnames",
			hostname:    "192.168.10.100",
			port:        8443,
			existing:    kubeConfigLocalhost,
			err:         true,
		},
		{
			description: "different hostname",
			hostname:    "",
			port:        8443,
			existing:    kubeConfigLocalhost,
			err:         true,
		},
		{
			description: "different ports",
			hostname:    "127.0.0.1",
			port:        84430,
			existing:    kubeConfigLocalhost,
			err:         true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			configFilename := tempFile(t, test.existing)
			defer os.Remove(configFilename)
			err := VerifyEndpoint("minikube", test.hostname, test.port, configFilename)
			if err != nil && !test.err {
				t.Errorf("Got unexpected error: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none: %v", err)
			}
		})

	}
}

func TestUpdateIP(t *testing.T) {

	var tests = []struct {
		description string
		hostname    string
		port        int
		existing    []byte
		err         bool
		status      bool
		expCfg      []byte
	}{
		{
			description: "empty hostname",
			hostname:    "",
			port:        8443,
			existing:    kubeConfig192,
			err:         true,
			expCfg:      kubeConfig192,
		},
		{
			description: "no minikube cluster",
			hostname:    "192.168.10.100",
			port:        8080,
			existing:    kubeConfigWithoutHTTPS,
			err:         true,
			expCfg:      kubeConfigWithoutHTTPS,
		},
		{
			description: "same IP",
			hostname:    "192.168.10.100",
			port:        8443,
			existing:    kubeConfig192,
			expCfg:      kubeConfig192,
		},
		{
			description: "different IP",
			hostname:    "127.0.0.1",
			port:        8443,
			existing:    kubeConfig192,
			status:      true,
			expCfg:      kubeConfigLocalhost,
		},
		{
			description: "different port",
			hostname:    "127.0.0.1",
			port:        12345,
			existing:    kubeConfigLocalhost,
			status:      true,
			expCfg:      kubeConfigLocalhost12345,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			configFilename := tempFile(t, test.existing)
			defer os.Remove(configFilename)
			statusActual, err := UpdateEndpoint("minikube", test.hostname, test.port, configFilename)
			if err != nil && !test.err {
				t.Errorf("Got unexpected error: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none: %v", err)
			}
			if test.status != statusActual {
				t.Errorf("Expected status %t, but got %t", test.status, statusActual)
			}

			actual, err := readOrNew(configFilename)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := decode(test.expCfg)
			if err != nil {
				t.Fatal(err)
			}
			if !configEquals(actual, expected) {
				t.Errorf("Expected cfg %v, but got %v", expected, actual)
			}
		})

	}
}

func TestEmptyConfig(t *testing.T) {
	tmp := tempFile(t, []byte{})
	defer os.Remove(tmp)

	cfg, err := readOrNew(tmp)
	if err != nil {
		t.Fatalf("could not read config: %v", err)
	}

	if len(cfg.AuthInfos) != 0 {
		t.Fail()
	}

	if len(cfg.Clusters) != 0 {
		t.Fail()
	}

	if len(cfg.Contexts) != 0 {
		t.Fail()
	}
}

func TestNewConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", ".kube")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Errorf("Failed to remove dir %q: %v", dir, err)
		}
	}()

	// setup minikube config
	expected := api.NewConfig()
	minikubeConfig(expected)

	// write actual
	filename := filepath.Join(dir, "config")
	err = writeToFile(expected, filename)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := readOrNew(filename)
	if err != nil {
		t.Fatal(err)
	}

	if !configEquals(actual, expected) {
		t.Fatal("configs did not match")
	}
}

func Test_Endpoint(t *testing.T) {

	var tests = []struct {
		description string
		cfg         []byte
		hostname    string
		port        int
		err         bool
	}{
		{
			description: "normal IP",
			cfg:         kubeConfig192,
			hostname:    "192.168.10.100",
			port:        8443,
		},
		{
			description: "no minikube cluster",
			cfg:         kubeConfigWithoutHTTPS,
			err:         true,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			configFilename := tempFile(t, test.cfg)
			defer os.Remove(configFilename)
			hostname, port, err := Endpoint("minikube", configFilename)
			if err != nil && !test.err {
				t.Errorf("Got unexpected error: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none: %v", err)
			}
			if hostname != test.hostname {
				t.Errorf("got hostname = %q, want hostname = %q", hostname, test.hostname)
			}
			if port != test.port {
				t.Errorf("got port = %q, want port = %q", port, test.port)
			}
		})
	}
}

// tempFile creates a temporary with the provided bytes as its contents.
// The caller is responsible for deleting file after use.
func tempFile(t *testing.T, data []byte) string {
	tmp, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		t.Fatal(err)
	}

	if len(data) > 0 {
		if _, err := tmp.Write(data); err != nil {
			t.Fatal(err)
		}
	}

	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	return tmp.Name()
}

// minikubeConfig returns a k8s cluster config
func minikubeConfig(config *api.Config) {
	// cluster
	clusterName := "minikube"
	cluster := api.NewCluster()
	cluster.Server = "https://192.168.99.100:" + strconv.Itoa(constants.APIServerPort)
	cluster.CertificateAuthority = "/home/tux/.minikube/apiserver.crt"
	config.Clusters[clusterName] = cluster

	// user
	userName := "minikube"
	user := api.NewAuthInfo()
	user.ClientCertificate = "/home/tux/.minikube/apiserver.crt"
	user.ClientKey = "/home/tux/.minikube/apiserver.key"
	config.AuthInfos[userName] = user

	// context
	contextName := "minikube"
	context := api.NewContext()
	context.Cluster = clusterName
	context.AuthInfo = userName
	config.Contexts[contextName] = context

	config.CurrentContext = contextName
}

// configEquals checks if configs are identical
func configEquals(a, b *api.Config) bool {
	if a.Kind != b.Kind {
		return false
	}

	if a.APIVersion != b.APIVersion {
		return false
	}

	if a.Preferences.Colors != b.Preferences.Colors {
		return false
	}
	if len(a.Extensions) != len(b.Extensions) {
		return false
	}

	// clusters
	if !clustersEquals(a, b) {
		return false
	}

	// users
	if !authInfosEquals(a, b) {
		return false
	}

	// contexts
	if !contextsEquals(a, b) {
		return false
	}

	return true
}

func clustersEquals(a, b *api.Config) bool {
	if len(a.Clusters) != len(b.Clusters) {
		return false
	}
	for k, aCluster := range a.Clusters {
		bCluster, exists := b.Clusters[k]
		if !exists {
			return false
		}

		if !clusterEquals(aCluster, bCluster) {
			return false
		}
	}
	return true
}

func clusterEquals(aCluster, bCluster *api.Cluster) bool {
	if aCluster.LocationOfOrigin != bCluster.LocationOfOrigin ||
		aCluster.Server != bCluster.Server ||
		aCluster.InsecureSkipTLSVerify != bCluster.InsecureSkipTLSVerify ||
		aCluster.CertificateAuthority != bCluster.CertificateAuthority ||
		len(aCluster.CertificateAuthorityData) != len(bCluster.CertificateAuthorityData) ||
		len(aCluster.Extensions) != len(bCluster.Extensions) {
		return false
	}
	return true
}

func authInfosEquals(a, b *api.Config) bool {
	if len(a.AuthInfos) != len(b.AuthInfos) {
		return false
	}
	for k, aAuth := range a.AuthInfos {
		bAuth, exists := b.AuthInfos[k]
		if !exists {
			return false
		}
		if !authInfoEquals(aAuth, bAuth) {
			return false
		}
	}
	return true
}

func authInfoEquals(aAuth, bAuth *api.AuthInfo) bool {
	if aAuth.LocationOfOrigin != bAuth.LocationOfOrigin ||
		aAuth.ClientCertificate != bAuth.ClientCertificate ||
		len(aAuth.ClientCertificateData) != len(bAuth.ClientCertificateData) ||
		aAuth.ClientKey != bAuth.ClientKey ||
		len(aAuth.ClientKeyData) != len(bAuth.ClientKeyData) ||
		aAuth.Token != bAuth.Token ||
		aAuth.Username != bAuth.Username ||
		aAuth.Password != bAuth.Password ||
		len(aAuth.Extensions) != len(bAuth.Extensions) {
		return false
	}
	return true
}

func contextsEquals(a, b *api.Config) bool {
	if len(a.Contexts) != len(b.Contexts) {
		return false
	}
	for k, aContext := range a.Contexts {
		bContext, exists := b.Contexts[k]
		if !exists {
			return false
		}
		if !contextEquals(aContext, bContext) {
			return false
		}
	}
	return true
}

func contextEquals(aContext, bContext *api.Context) bool {
	if aContext.LocationOfOrigin != bContext.LocationOfOrigin ||
		aContext.Cluster != bContext.Cluster ||
		aContext.AuthInfo != bContext.AuthInfo ||
		aContext.Namespace != bContext.Namespace ||
		len(aContext.Extensions) != len(bContext.Extensions) {
		return false
	}
	return true
}

func TestGetKubeConfigPath(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{
			input: "/home/fake/.kube/.kubeconfig",
			want:  "/home/fake/.kube/.kubeconfig",
		},
		{
			input: "/home/fake/.kube/.kubeconfig:/home/fake2/.kubeconfig",
			want:  "/home/fake/.kube/.kubeconfig",
		},
		{
			input: ":/home/fake/.kube/.kubeconfig:/home/fake2/.kubeconfig",
			want:  "/home/fake/.kube/.kubeconfig",
		},
		{
			input: ":",
			want:  "$HOME/.kube/config",
		},
		{
			input: "",
			want:  "$HOME/.kube/config",
		},
	}

	for _, test := range tests {
		os.Setenv(clientcmd.RecommendedConfigPathEnvVar, test.input)
		if result := PathFromEnv(); result != os.ExpandEnv(test.want) {
			t.Errorf("Expected first split chunk, got: %s", result)
		}
	}
}
