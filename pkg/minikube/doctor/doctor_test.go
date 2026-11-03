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

package doctor

import (
	"errors"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/registry"
)

func healthyCluster() *config.ClusterConfig {
	return &config.ClusterConfig{
		Name: "minikube", Driver: "docker", CPUs: 2, Memory: 2048,
		KubernetesConfig: config.KubernetesConfig{KubernetesVersion: "v1.35.0", ContainerRuntime: "docker"},
	}
}

func running(names ...string) []*cluster.Status {
	var s []*cluster.Status
	for i, n := range names {
		st := &cluster.Status{Name: n, Host: "Running", Kubelet: "Running", APIServer: "Running", Kubeconfig: cluster.Configured}
		if i > 0 {
			st.APIServer, st.Kubeconfig, st.Worker = cluster.Irrelevant, cluster.Irrelevant, true
		}
		s = append(s, st)
	}
	return s
}

// fakeHost returns a host where everything is healthy; tests override one
// field at a time.
func fakeHost(cc *config.ClusterConfig) Host {
	return Host{
		LoadConfig:    func(string) (*config.ClusterConfig, error) { return cc, nil },
		ListProfiles:  func() ([]*config.Profile, []*config.Profile, error) { return nil, nil, nil },
		DriverStatus:  func(string) (registry.State, bool) { return registry.State{Installed: true, Healthy: true}, true },
		ClusterStatus: func(*config.ClusterConfig) ([]*cluster.Status, error) { return running("minikube"), nil },
		LookPath:      func(string) (string, error) { return "/usr/bin/kubectl", nil },
		CurrentContext: func() (string, error) {
			if cc == nil {
				return "minikube", nil
			}
			return cc.Name, nil
		},
	}
}

func find(t *testing.T, r Report, name string) Check {
	t.Helper()
	for _, c := range r.Checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no check %q in %+v", name, r.Checks)
	return Check{}
}

func TestDefaultClusterIsHealthy(t *testing.T) {
	// `minikube start` defaults: 2 CPUs and about a quarter of host RAM.
	r := Run("minikube", fakeHost(healthyCluster()))
	for _, c := range r.Checks {
		if c.Status != Pass {
			t.Errorf("%s: %s %q; a default cluster must not be flagged", c.Name, c.Status, c.Message)
		}
	}
	if r.Driver != "docker" || r.Runtime != "docker" || r.Kubernetes != "v1.35.0" {
		t.Errorf("report header = %+v", r)
	}
}

func TestMissingProfileIsReportedNotFatal(t *testing.T) {
	h := fakeHost(nil)
	h.LoadConfig = func(string) (*config.ClusterConfig, error) {
		return nil, &config.ErrNotExist{}
	}
	r := Run("dev", h)
	p := find(t, r, "Profile")
	if p.Status != Fail || !strings.Contains(p.Fix, "minikube start -p dev") {
		t.Errorf("profile = %+v", p)
	}
	if d := find(t, r, "Driver"); d.Status != Skip {
		t.Errorf("driver should be skipped, got %+v", d)
	}
	if k := find(t, r, "kubectl"); k.Status != Pass {
		t.Error("host checks still run without a profile")
	}
}

func TestDriverUsesRegistryStatus(t *testing.T) {
	cc := healthyCluster()
	cc.Driver = "qemu2"
	h := fakeHost(cc)
	h.DriverStatus = func(name string) (registry.State, bool) {
		if name != "qemu2" {
			t.Errorf("status asked for %q", name)
		}
		return registry.State{Installed: true, Healthy: true, Version: "9.1.0"}, true
	}
	if d := find(t, Run("minikube", h), "Driver"); d.Status != Pass {
		t.Errorf("qemu2 installed and healthy must pass: %+v", d)
	}

	h.DriverStatus = func(string) (registry.State, bool) {
		return registry.State{Installed: true, Healthy: false, Error: errors.New("docker daemon not running"), Fix: "Start Docker"}, true
	}
	r := Run("minikube", h)
	d := find(t, r, "Driver")
	if d.Status != Fail || d.Fix != "Start Docker" || d.Details != "docker daemon not running" {
		t.Errorf("unhealthy driver = %+v", d)
	}
	if c := find(t, r, "Cluster status"); c.Status != Skip {
		t.Errorf("cluster checks must be skipped when the driver is unusable: %+v", c)
	}

	h.DriverStatus = func(string) (registry.State, bool) { return registry.State{}, false }
	if d := find(t, Run("minikube", h), "Driver"); d.Status != Fail || !strings.Contains(d.Message, "not supported") {
		t.Errorf("unknown driver = %+v", d)
	}
}

func TestStoppedClusterSkipsDependentChecks(t *testing.T) {
	h := fakeHost(healthyCluster())
	calls := 0
	h.ClusterStatus = func(*config.ClusterConfig) ([]*cluster.Status, error) {
		calls++
		return []*cluster.Status{{Name: "minikube", Host: "Stopped"}}, nil
	}
	r := Run("minikube", h)
	if c := find(t, r, "Cluster status"); c.Status != Fail || c.Message != "cluster is stopped" {
		t.Errorf("cluster = %+v", c)
	}
	for _, n := range []string{"API server", "Nodes", "Kubeconfig"} {
		if c := find(t, r, n); c.Status != Skip {
			t.Errorf("%s = %+v", n, c)
		}
	}
	if calls != 1 {
		t.Errorf("cluster status queried %d times; want once", calls)
	}
}

func TestMultiNodeAndNoKubernetes(t *testing.T) {
	h := fakeHost(healthyCluster())
	h.ClusterStatus = func(*config.ClusterConfig) ([]*cluster.Status, error) {
		s := running("minikube", "minikube-m02", "minikube-m03")
		s[2].Kubelet = "Stopped"
		return s, nil
	}
	r := Run("minikube", h)
	if n := find(t, r, "Nodes"); n.Status != Fail || !strings.Contains(n.Message, "minikube-m03") || strings.Contains(n.Message, "m02") {
		t.Errorf("nodes = %+v", n)
	}
	if k := find(t, r, "Kubeconfig"); k.Status != Pass {
		t.Errorf("worker nodes' Irrelevant kubeconfig must not fail: %+v", k)
	}

	cc := healthyCluster()
	cc.KubernetesConfig.KubernetesVersion = constants.NoKubernetesVersion
	h = fakeHost(cc)
	h.ClusterStatus = func(*config.ClusterConfig) ([]*cluster.Status, error) {
		return []*cluster.Status{{Name: "minikube", Host: "Running", Kubelet: cluster.Irrelevant, APIServer: cluster.Irrelevant, Kubeconfig: cluster.Irrelevant}}, nil
	}
	r = Run("minikube", h)
	if r.Count(Fail) != 0 {
		t.Errorf("--no-kubernetes cluster must not fail: %+v", r.Checks)
	}
}

func TestResourcesMatchStartLimits(t *testing.T) {
	cases := []struct {
		cpus, mem        int
		cpuWant, memWant Status
	}{
		{2, 2048, Pass, Pass},
		{0, 0, Pass, Pass}, // --cpus=no-limit --memory=no-limit
		{1, 1850, Fail, Warn},
		{4, 1700, Pass, Fail},
	}
	for _, tc := range cases {
		cc := healthyCluster()
		cc.CPUs, cc.Memory = tc.cpus, tc.mem
		rs := resources(cc)
		if rs[0].Status != tc.cpuWant || rs[1].Status != tc.memWant {
			t.Errorf("cpus=%d mem=%d: got %s/%s, want %s/%s", tc.cpus, tc.mem, rs[0].Status, rs[1].Status, tc.cpuWant, tc.memWant)
		}
	}
}

func TestKubectlMissingIsAWarning(t *testing.T) {
	h := fakeHost(healthyCluster())
	h.LookPath = func(string) (string, error) { return "", errors.New("not found") }
	k := find(t, Run("minikube", h), "kubectl")
	if k.Status != Warn || !strings.Contains(k.Fix, "minikube kubectl") {
		t.Errorf("kubectl = %+v", k)
	}
}

func TestInvalidOtherProfileIsAWarning(t *testing.T) {
	h := fakeHost(healthyCluster())
	h.ListProfiles = func() ([]*config.Profile, []*config.Profile, error) {
		return nil, []*config.Profile{{Name: "old"}}, nil
	}
	p := find(t, Run("minikube", h), "Profiles")
	if p.Status != Warn || !strings.Contains(p.Message, "old") {
		t.Errorf("profiles = %+v", p)
	}
}

func TestDeletedContainerIsExplained(t *testing.T) {
	h := fakeHost(healthyCluster())
	h.ClusterStatus = func(*config.ClusterConfig) ([]*cluster.Status, error) {
		return nil, errors.New("host: state: unknown state \"minikube\": docker container inspect minikube: exit status 1\nstderr:\nError response from daemon: No such container: minikube")
	}
	c := find(t, Run("minikube", h), "Cluster status")
	if !strings.Contains(c.Message, "no longer exists") || strings.Contains(c.Details, "\n") || !strings.Contains(c.Fix, "minikube delete -p minikube") {
		t.Errorf("cluster = %+v", c)
	}
}

func TestDriverMessageWithoutVersion(t *testing.T) {
	if d := driver("docker", fakeHost(nil)); d.Message != "docker is healthy" {
		t.Errorf("message = %q", d.Message)
	}
}

func TestKubectlUsingAnotherContextIsAWarning(t *testing.T) {
	h := fakeHost(healthyCluster())
	h.CurrentContext = func() (string, error) { return "kind-production-lab", nil }
	k := find(t, Run("minikube", h), "Kubeconfig")
	if k.Status != Warn || !strings.Contains(k.Message, "kind-production-lab") || k.Fix != "Switch with: kubectl config use-context minikube" {
		t.Errorf("kubeconfig = %+v", k)
	}
}

func TestUnhealthyDriverAlwaysHasAFix(t *testing.T) {
	h := fakeHost(healthyCluster())
	h.DriverStatus = func(string) (registry.State, bool) {
		return registry.State{Installed: true, Error: errors.New("cannot connect")}, true
	}
	if d := find(t, Run("minikube", h), "Driver"); d.Fix != "Start docker, then run `minikube doctor` again" {
		t.Errorf("fix = %q", d.Fix)
	}
}
