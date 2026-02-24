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

package kubeadm

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
)

func TestConfigureDNSSearch_Empty(t *testing.T) {
	fcr := command.NewFakeCommandRunner()
	k := &Bootstrapper{c: fcr}

	cfg := config.ClusterConfig{
		DNSSearch: nil,
	}

	if err := k.configureDNSSearch(cfg); err != nil {
		t.Errorf("configureDNSSearch() with empty search returned error: %v", err)
	}
}

func TestConfigureDNSSearch_SingleDomain(t *testing.T) {
	fcr := command.NewFakeCommandRunner()
	fcr.SetCommandToOutput(map[string]string{
		`sudo sh -c "{ echo 'search corp.example.com'; sed '/^search /d' /etc/resolv.conf; } > /tmp/resolv.tmp && cp /tmp/resolv.tmp /etc/resolv.conf && rm /tmp/resolv.tmp"`: "",
	})

	k := &Bootstrapper{c: fcr}
	cfg := config.ClusterConfig{
		DNSSearch: []string{"corp.example.com"},
		KubernetesConfig: config.KubernetesConfig{
			KubernetesVersion: "v1.28.0",
		},
	}

	if err := k.configureDNSSearch(cfg); err != nil {
		t.Errorf("configureDNSSearch() returned error: %v", err)
	}
}

func TestConfigureDNSSearch_MultipleDomains(t *testing.T) {
	fcr := command.NewFakeCommandRunner()
	fcr.SetCommandToOutput(map[string]string{
		`sudo sh -c "{ echo 'search corp.example.com eng.example.com'; sed '/^search /d' /etc/resolv.conf; } > /tmp/resolv.tmp && cp /tmp/resolv.tmp /etc/resolv.conf && rm /tmp/resolv.tmp"`: "",
	})

	k := &Bootstrapper{c: fcr}
	cfg := config.ClusterConfig{
		DNSSearch: []string{"corp.example.com", "eng.example.com"},
		KubernetesConfig: config.KubernetesConfig{
			KubernetesVersion: "v1.28.0",
		},
	}

	if err := k.configureDNSSearch(cfg); err != nil {
		t.Errorf("configureDNSSearch() returned error: %v", err)
	}
}

func TestConfigureDNSSearch_K8s125Regression(t *testing.T) {
	// K8s v1.25.0 has the resolv.conf search regression, so configureDNSSearch
	// should also update /etc/kubelet-resolv.conf
	fcr := command.NewFakeCommandRunner()
	fcr.SetCommandToOutput(map[string]string{
		`sudo sh -c "{ echo 'search corp.example.com'; sed '/^search /d' /etc/resolv.conf; } > /tmp/resolv.tmp && cp /tmp/resolv.tmp /etc/resolv.conf && rm /tmp/resolv.tmp"`:                         "",
		`sudo sh -c "{ echo 'search corp.example.com'; sed '/^search /d' /etc/kubelet-resolv.conf; } > /tmp/kubelet-resolv.tmp && cp /tmp/kubelet-resolv.tmp /etc/kubelet-resolv.conf && rm /tmp/kubelet-resolv.tmp"`: "",
	})

	k := &Bootstrapper{c: fcr}
	cfg := config.ClusterConfig{
		DNSSearch: []string{"corp.example.com"},
		KubernetesConfig: config.KubernetesConfig{
			KubernetesVersion: "v1.25.0",
		},
	}

	if err := k.configureDNSSearch(cfg); err != nil {
		t.Errorf("configureDNSSearch() with K8s v1.25.0 returned error: %v", err)
	}
}

func TestConfigureDNSSearch_K8s125NotAffected(t *testing.T) {
	// K8s v1.25.3 does NOT have the regression, so only /etc/resolv.conf should be updated.
	// If kubelet-resolv.conf commands are called, the fake runner will error on unknown commands.
	fcr := command.NewFakeCommandRunner()
	fcr.SetCommandToOutput(map[string]string{
		`sudo sh -c "{ echo 'search corp.example.com'; sed '/^search /d' /etc/resolv.conf; } > /tmp/resolv.tmp && cp /tmp/resolv.tmp /etc/resolv.conf && rm /tmp/resolv.tmp"`: "",
	})

	k := &Bootstrapper{c: fcr}
	cfg := config.ClusterConfig{
		DNSSearch: []string{"corp.example.com"},
		KubernetesConfig: config.KubernetesConfig{
			KubernetesVersion: "v1.25.3",
		},
	}

	if err := k.configureDNSSearch(cfg); err != nil {
		t.Errorf("configureDNSSearch() with K8s v1.25.3 returned error: %v", err)
	}
}

func TestConfigureDNSSearch_NoneDriverUnsupported(t *testing.T) {
	fcr := command.NewFakeCommandRunner()
	k := &Bootstrapper{c: fcr}
	cfg := config.ClusterConfig{
		Driver:    driver.None,
		DNSSearch: []string{"corp.example.com"},
		KubernetesConfig: config.KubernetesConfig{
			KubernetesVersion: "v1.28.0",
		},
	}

	if err := k.configureDNSSearch(cfg); err == nil {
		t.Error("configureDNSSearch() should fail for none driver")
	}
}

func TestConfigureDNSSearch_ShellError(t *testing.T) {
	// Don't register any commands so the shell command fails
	fcr := command.NewFakeCommandRunner()
	k := &Bootstrapper{c: fcr}
	cfg := config.ClusterConfig{
		DNSSearch: []string{"corp.example.com"},
		KubernetesConfig: config.KubernetesConfig{
			KubernetesVersion: "v1.28.0",
		},
	}

	err := k.configureDNSSearch(cfg)
	if err == nil {
		t.Error("configureDNSSearch() should have returned error when command fails")
	}
}
