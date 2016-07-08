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

package kubeconfig

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"k8s.io/minikube/pkg/minikube/constants"
)

func TestEmptyConfig(t *testing.T) {
	tmp := tempFile(t, []byte{})
	defer os.Remove(tmp)

	cfg, err := ReadConfigOrNew(tmp)
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
	defer os.RemoveAll(dir)

	// setup minikube config
	expected := api.NewConfig()
	minikubeConfig(expected)

	// write actual
	filename := filepath.Join(dir, "config")
	err = WriteConfig(expected, filename)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := ReadConfigOrNew(filename)
	if err != nil {
		t.Fatal(err)
	}

	if !configEquals(actual, expected) {
		t.Fatal("configs did not match")
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

// minikubeConfig returns a config that reasonably approximates a localkube cluster
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
	if len(a.Clusters) != len(b.Clusters) {
		return false
	}
	for k, aCluster := range a.Clusters {
		bCluster, exists := b.Clusters[k]
		if !exists {
			return false
		}

		if aCluster.LocationOfOrigin != bCluster.LocationOfOrigin ||
			aCluster.Server != bCluster.Server ||
			aCluster.APIVersion != bCluster.APIVersion ||
			aCluster.InsecureSkipTLSVerify != bCluster.InsecureSkipTLSVerify ||
			aCluster.CertificateAuthority != bCluster.CertificateAuthority ||
			len(aCluster.CertificateAuthorityData) != len(bCluster.CertificateAuthorityData) ||
			len(aCluster.Extensions) != len(bCluster.Extensions) {
			return false
		}
	}

	// users
	if len(a.AuthInfos) != len(b.AuthInfos) {
		return false
	}
	for k, aAuth := range a.AuthInfos {
		bAuth, exists := b.AuthInfos[k]
		if !exists {
			return false
		}
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

	}

	// contexts
	if len(a.Contexts) != len(b.Contexts) {
		return false
	}
	for k, aContext := range a.Contexts {
		bContext, exists := b.Contexts[k]
		if !exists {
			return false
		}
		if aContext.LocationOfOrigin != bContext.LocationOfOrigin ||
			aContext.Cluster != bContext.Cluster ||
			aContext.AuthInfo != bContext.AuthInfo ||
			aContext.Namespace != bContext.Namespace ||
			len(aContext.Extensions) != len(aContext.Extensions) {
			return false
		}

	}
	return true
}
