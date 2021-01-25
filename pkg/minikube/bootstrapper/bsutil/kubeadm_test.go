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

package bsutil

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/mod/semver"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

func getExtraOpts() []config.ExtraOption {
	return config.ExtraOptionSlice{
		config.ExtraOption{
			Component: Apiserver,
			Key:       "fail-no-swap",
			Value:     "true",
		},
		config.ExtraOption{
			Component: ControllerManager,
			Key:       "kube-api-burst",
			Value:     "32",
		},
		config.ExtraOption{
			Component: Scheduler,
			Key:       "scheduler-name",
			Value:     "mini-scheduler",
		},
		config.ExtraOption{
			Component: Kubeadm,
			Key:       "ignore-preflight-errors",
			Value:     "true",
		},
		config.ExtraOption{
			Component: Kubeadm,
			Key:       "dry-run",
			Value:     "true",
		},
		config.ExtraOption{
			Component: Kubeproxy,
			Key:       "mode",
			Value:     "iptables",
		},
	}
}

func getExtraOptsPodCidr() []config.ExtraOption {
	return config.ExtraOptionSlice{
		config.ExtraOption{
			Component: Kubeadm,
			Key:       "pod-network-cidr",
			Value:     "192.168.32.0/20",
		},
	}
}

// recentReleases returns a dynamic list of up to n recent testdata versions, sorted from newest to older.
// If n > 0, returns at most n versions.
// If n <= 0, returns all the versions.
// It will error if no testdata are available or in absence of testdata for newest and default minor k8s versions.
func recentReleases(n int) ([]string, error) {
	path := "testdata"
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to list testdata directory %s: %w", path, err)
	}
	var versions []string
	for _, file := range files {
		if file.IsDir() {
			versions = append(versions, file.Name())
		}
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] > versions[j] })
	if n <= 0 || n > len(versions) {
		n = len(versions)
	}
	versions = versions[0:n]

	foundNewest := false
	foundDefault := false

	for _, v := range versions {
		if strings.HasPrefix(constants.NewestKubernetesVersion, v) {
			foundNewest = true
		}
		if strings.HasPrefix(constants.DefaultKubernetesVersion, v) {
			foundDefault = true
		}
	}

	if !foundNewest {
		return nil, fmt.Errorf("No tests exist yet for newest minor version: %s", constants.NewestKubernetesVersion)
	}

	if !foundDefault {
		return nil, fmt.Errorf("No tests exist yet for default minor version: %s", constants.DefaultKubernetesVersion)
	}

	return versions, nil
}

/**
Need a separate test function to test the DNS server IP
as v1.11 yaml file is very different compared to v1.12+.
This test case has only 1 thing to test and that is the
networking/dnsDomain value
*/
func TestGenerateKubeadmYAMLDNS(t *testing.T) {
	// test all testdata releases greater than v1.11
	versions, err := recentReleases(0)
	if err != nil {
		t.Errorf("versions: %v", err)
	}
	for i, v := range versions {
		if semver.Compare(v, "v1.11") <= 0 {
			versions = versions[0:i]
			break
		}
	}
	fcr := command.NewFakeCommandRunner()
	fcr.SetCommandToOutput(map[string]string{
		"docker info --format {{.CgroupDriver}}": "systemd\n",
	})
	tests := []struct {
		name      string
		runtime   string
		shouldErr bool
		cfg       config.ClusterConfig
	}{
		{"dns", "docker", false, config.ClusterConfig{Name: "mk", KubernetesConfig: config.KubernetesConfig{DNSDomain: "minikube.local"}}},
	}
	for _, version := range versions {
		for _, tc := range tests {
			runtime, err := cruntime.New(cruntime.Config{Type: tc.runtime, Runner: fcr})
			if err != nil {
				t.Fatalf("runtime: %v", err)
			}
			tname := tc.name + "_" + version
			t.Run(tname, func(t *testing.T) {
				cfg := tc.cfg
				cfg.Nodes = []config.Node{
					{
						IP:           "1.1.1.1",
						Name:         "mk",
						ControlPlane: true,
					},
				}
				cfg.KubernetesConfig.KubernetesVersion = version + ".0"
				// if version+".0" does not yet have a stable release, use NewestKubernetesVersion
				// ie, 'v1.20.0-beta.1' NewestKubernetesVersion indicates that 'v1.20.0' is not yet released as stable
				if semver.Compare(cfg.KubernetesConfig.KubernetesVersion, constants.NewestKubernetesVersion) == 1 {
					cfg.KubernetesConfig.KubernetesVersion = constants.NewestKubernetesVersion
				}
				cfg.KubernetesConfig.ClusterName = "kubernetes"

				got, err := GenerateKubeadmYAML(cfg, cfg.Nodes[0], runtime)
				if err != nil && !tc.shouldErr {
					t.Fatalf("got unexpected error generating config: %v", err)
				}
				if err == nil && tc.shouldErr {
					t.Fatalf("expected error but got none, config: %s", got)
				}
				if tc.shouldErr {
					return
				}
				expected, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/%s.yaml", version, tc.name))
				if err != nil {
					t.Fatalf("unable to read testdata: %v", err)
				}
				diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(string(expected)),
					B:        difflib.SplitLines(string(got)),
					FromFile: "Expected",
					ToFile:   "Got",
					Context:  1,
				})
				if err != nil {
					t.Fatalf("diff error: %v", err)
				}
				if diff != "" {
					t.Errorf("unexpected diff:\n%s\n===== [RAW OUTPUT] =====\n%s", diff, got)
				}
			})
		}
	}
}

func TestGenerateKubeadmYAML(t *testing.T) {
	extraOpts := getExtraOpts()
	extraOptsPodCidr := getExtraOptsPodCidr()
	// test the 6 most recent releases
	versions, err := recentReleases(6)
	if err != nil {
		t.Errorf("versions: %v", err)
	}
	fcr := command.NewFakeCommandRunner()
	fcr.SetCommandToOutput(map[string]string{
		"docker info --format {{.CgroupDriver}}": "systemd\n",
		"crio config":                            "cgroup_manager = \"systemd\"\n",
		"sudo crictl info":                       "{\"config\": {\"systemdCgroup\": true}}",
	})
	tests := []struct {
		name      string
		runtime   string
		shouldErr bool
		cfg       config.ClusterConfig
	}{
		{"default", "docker", false, config.ClusterConfig{Name: "mk"}},
		{"containerd", "containerd", false, config.ClusterConfig{Name: "mk"}},
		{"crio", "crio", false, config.ClusterConfig{Name: "mk"}},
		{"options", "docker", false, config.ClusterConfig{Name: "mk", KubernetesConfig: config.KubernetesConfig{ExtraOptions: extraOpts}}},
		{"crio-options-gates", "crio", false, config.ClusterConfig{Name: "mk", KubernetesConfig: config.KubernetesConfig{ExtraOptions: extraOpts, FeatureGates: "a=b"}}},
		{"unknown-component", "docker", true, config.ClusterConfig{Name: "mk", KubernetesConfig: config.KubernetesConfig{ExtraOptions: config.ExtraOptionSlice{config.ExtraOption{Component: "not-a-real-component", Key: "killswitch", Value: "true"}}}}},
		{"containerd-api-port", "containerd", false, config.ClusterConfig{Name: "mk", Nodes: []config.Node{{Port: 12345}}}},
		{"containerd-pod-network-cidr", "containerd", false, config.ClusterConfig{Name: "mk", KubernetesConfig: config.KubernetesConfig{ExtraOptions: extraOptsPodCidr}}},
		{"image-repository", "docker", false, config.ClusterConfig{Name: "mk", KubernetesConfig: config.KubernetesConfig{ImageRepository: "test/repo"}}},
	}
	for _, version := range versions {
		for _, tc := range tests {
			runtime, err := cruntime.New(cruntime.Config{Type: tc.runtime, Runner: fcr})
			if err != nil {
				t.Fatalf("runtime: %v", err)
			}
			tname := tc.name + "_" + version
			t.Run(tname, func(t *testing.T) {
				cfg := tc.cfg

				if len(cfg.Nodes) > 0 {
					cfg.Nodes[0].IP = "1.1.1.1"
					cfg.Nodes[0].Name = "mk"
					cfg.Nodes[0].ControlPlane = true
				} else {
					cfg.Nodes = []config.Node{
						{
							IP:           "1.1.1.1",
							Name:         "mk",
							ControlPlane: true,
						},
					}
				}
				cfg.KubernetesConfig.KubernetesVersion = version + ".0"
				// if version+".0" does not yet have a stable release, use NewestKubernetesVersion
				// ie, 'v1.20.0-beta.1' NewestKubernetesVersion indicates that 'v1.20.0' is not yet released as stable
				if semver.Compare(cfg.KubernetesConfig.KubernetesVersion, constants.NewestKubernetesVersion) == 1 {
					cfg.KubernetesConfig.KubernetesVersion = constants.NewestKubernetesVersion
				}
				cfg.KubernetesConfig.ClusterName = "kubernetes"

				got, err := GenerateKubeadmYAML(cfg, cfg.Nodes[0], runtime)
				if err != nil && !tc.shouldErr {
					t.Fatalf("got unexpected error generating config: %v", err)
				}
				if err == nil && tc.shouldErr {
					t.Fatalf("expected error but got none, config: %s", got)
				}
				if tc.shouldErr {
					return
				}
				expected, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/%s.yaml", version, tc.name))
				if err != nil {
					t.Fatalf("unable to read testdata: %v", err)
				}
				diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(string(expected)),
					B:        difflib.SplitLines(string(got)),
					FromFile: "Expected",
					ToFile:   "Got",
					Context:  1,
				})
				if err != nil {
					t.Fatalf("diff error: %v", err)
				}
				if diff != "" {
					t.Errorf("unexpected diff:\n%s\n", diff)
				}
			})
		}
	}
}

func TestEtcdExtraArgs(t *testing.T) {
	expected := map[string]string{
		"key": "value",
	}
	extraOpts := append(getExtraOpts(), config.ExtraOption{
		Component: Etcd,
		Key:       "key",
		Value:     "value",
	})
	actual := etcdExtraArgs(extraOpts)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("machines mismatch (-want +got):\n%s", diff)
	}
}
