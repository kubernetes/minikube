/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package addons

import (
	"context"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/assets"
)

func TestParseImageRepository(t *testing.T) {
	tests := []struct {
		description string
		input       string
		wantRegis   string
		wantOrg     string
		wantTag     string
		wantErr     bool
	}{
		{
			description: "ghcr.io with latest tag",
			input:       "ghcr.io/volcano-sh:latest",
			wantRegis:   "ghcr.io",
			wantOrg:     "volcano-sh",
			wantTag:     "latest",
			wantErr:     false,
		},
		{
			description: "docker.io with version tag",
			input:       "docker.io/volcanosh:v1.12.0",
			wantRegis:   "docker.io",
			wantOrg:     "volcanosh",
			wantTag:     "v1.12.0",
			wantErr:     false,
		},
		{
			description: "no tag defaults to latest",
			input:       "my-mirror.cn/volcanosh",
			wantRegis:   "my-mirror.cn",
			wantOrg:     "volcanosh",
			wantTag:     "latest",
			wantErr:     false,
		},
		{
			description: "empty string returns nil",
			input:       "",
			wantRegis:   "",
			wantOrg:     "",
			wantTag:     "",
			wantErr:     false,
		},
		{
			description: "registry only - error",
			input:       "docker.io",
			wantErr:     true,
		},
		{
			description: "trailing slash - error",
			input:       "docker.io/",
			wantErr:     true,
		},
		{
			description: "registry with port and org",
			input:       "somecompany.com:5000/myorg:v1.0",
			wantRegis:   "somecompany.com:5000",
			wantOrg:     "myorg",
			wantTag:     "v1.0",
			wantErr:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := parseImageRepository(test.input)
			if test.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q but got none", test.input)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if test.input == "" {
				if result != nil {
					t.Errorf("expected nil result for empty input")
				}
				return
			}
			if result.Registry != test.wantRegis {
				t.Errorf("registry mismatch: want %q, got %q", test.wantRegis, result.Registry)
			}
			if result.Org != test.wantOrg {
				t.Errorf("org mismatch: want %q, got %q", test.wantOrg, result.Org)
			}
			if result.Tag != test.wantTag {
				t.Errorf("tag mismatch: want %q, got %q", test.wantTag, result.Tag)
			}
		})
	}
}

func TestHelmCommand(t *testing.T) {
	tests := []struct {
		description     string
		chart           *assets.HelmChart
		enable          bool
		imageRepository string
		expected        string
		mode            string
	}{
		{
			description: "enable an addon with default registry",
			chart: &assets.HelmChart{
				Name:       "addon-name",
				Repo:       "addon-repo/addon-chart",
				Namespace:  "addon-namespace",
				Values:     []string{"key=value"},
				ValueFiles: []string{"/etc/kubernetes/addons/values.yaml"},
			},
			enable:          true,
			imageRepository: "",
			expected:        "sudo KUBECONFIG=/var/lib/minikube/kubeconfig helm upgrade --install addon-name addon-repo/addon-chart --create-namespace --namespace addon-namespace --set key=value --values /etc/kubernetes/addons/values.yaml",
		},
		{
			description: "enable an addon without namespace",
			chart: &assets.HelmChart{
				Name:       "addon-name",
				Repo:       "addon-repo/addon-chart",
				Values:     []string{"key=value"},
				ValueFiles: []string{"/etc/kubernetes/addons/values.yaml"},
			},
			enable:          true,
			imageRepository: "",
			expected:        "sudo KUBECONFIG=/var/lib/minikube/kubeconfig helm upgrade --install addon-name addon-repo/addon-chart --create-namespace --set key=value --values /etc/kubernetes/addons/values.yaml",
		},
		{
			description: "disable an addon",
			chart: &assets.HelmChart{
				Name:      "addon-name",
				Namespace: "addon-namespace",
			},
			enable:          false,
			imageRepository: "",
			expected:        "sudo KUBECONFIG=/var/lib/minikube/kubeconfig helm uninstall addon-name --namespace addon-namespace",
			mode:            "cpu",
		},
		{
			description: "enable addon with simple registry override (no ImageNameKeys)",
			chart: &assets.HelmChart{
				Name:        "metrics-server",
				Repo:        "metrics-server/metrics-server",
				Namespace:   "kube-system",
				ImageSetKey: "image.repository",
			},
			enable:          true,
			imageRepository: "my-registry.example.com/images",
			expected:        "sudo KUBECONFIG=/var/lib/minikube/kubeconfig helm upgrade --install metrics-server metrics-server/metrics-server --create-namespace --namespace kube-system --set image.repository=my-registry.example.com/images",
		},
		{
			description: "enable addon with image repository but no ImageSetKey",
			chart: &assets.HelmChart{
				Name:      "addon-no-imagesetkey",
				Repo:      "some-repo/chart",
				Namespace: "default",
			},
			enable:          true,
			imageRepository: "my-registry.example.com/images",
			expected:        "sudo KUBECONFIG=/var/lib/minikube/kubeconfig helm upgrade --install addon-no-imagesetkey some-repo/chart --create-namespace --namespace default",
		},
		{
			description: "volcano with ghcr.io/volcano-sh:latest",
			chart: &assets.HelmChart{
				Name:        "volcano",
				Repo:        "volcano/volcano",
				Namespace:   "volcano-system",
				ImageSetKey: "basic.image_registry",
				TagSetKey:   "basic.image_tag_version",
				ImageNameKeys: map[string]string{
					"vc-controller-manager": "basic.controller_image_name",
					"vc-scheduler":          "basic.scheduler_image_name",
					"vc-webhook-manager":    "basic.admission_image_name",
				},
			},
			enable:          true,
			imageRepository: "ghcr.io/volcano-sh:latest",
			// Note: map iteration order is not guaranteed, so we check contains instead
			expected: "sudo KUBECONFIG=/var/lib/minikube/kubeconfig helm upgrade --install volcano volcano/volcano --create-namespace --namespace volcano-system --set basic.image_registry=ghcr.io",
		},
		{
			description: "volcano with docker.io/volcanosh:v1.12.0",
			chart: &assets.HelmChart{
				Name:        "volcano",
				Repo:        "volcano/volcano",
				Namespace:   "volcano-system",
				ImageSetKey: "basic.image_registry",
				TagSetKey:   "basic.image_tag_version",
				ImageNameKeys: map[string]string{
					"vc-controller-manager": "basic.controller_image_name",
					"vc-scheduler":          "basic.scheduler_image_name",
					"vc-webhook-manager":    "basic.admission_image_name",
				},
			},
			enable:          true,
			imageRepository: "docker.io/volcanosh:v1.12.0",
			expected:        "sudo KUBECONFIG=/var/lib/minikube/kubeconfig helm upgrade --install volcano volcano/volcano --create-namespace --namespace volcano-system --set basic.image_registry=docker.io",
		},
		{
			description: "volcano no flag uses defaults",
			chart: &assets.HelmChart{
				Name:        "volcano",
				Repo:        "volcano/volcano",
				Namespace:   "volcano-system",
				ImageSetKey: "basic.image_registry",
				TagSetKey:   "basic.image_tag_version",
				ImageNameKeys: map[string]string{
					"vc-controller-manager": "basic.controller_image_name",
				},
			},
			enable:          true,
			imageRepository: "",
			expected:        "sudo KUBECONFIG=/var/lib/minikube/kubeconfig helm upgrade --install volcano volcano/volcano --create-namespace --namespace volcano-system",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			command := helmUninstallOrInstall(context.Background(), test.chart, test.enable, test.imageRepository)
			actual := strings.Join(command.Args, " ")

			// For tests with ImageNameKeys, check that expected parts are present
			// (map iteration order is not guaranteed)
			if test.chart.ImageNameKeys != nil && test.imageRepository != "" {
				if !strings.Contains(actual, test.expected) {
					t.Errorf("helm command missing expected prefix:\nwant contains: %s\nactual:        %s", test.expected, actual)
				}
				// Verify all image name keys are set
				parsed, _ := parseImageRepository(test.imageRepository)
				if parsed != nil {
					for imageName := range test.chart.ImageNameKeys {
						expectedSet := parsed.Org + "/" + imageName
						if !strings.Contains(actual, expectedSet) {
							t.Errorf("missing image name set: %s", expectedSet)
						}
					}
					// Verify tag is set
					if test.chart.TagSetKey != "" {
						expectedTag := test.chart.TagSetKey + "=" + parsed.Tag
						if !strings.Contains(actual, expectedTag) {
							t.Errorf("missing tag set: %s", expectedTag)
						}
					}
				}
			} else {
				if actual != test.expected {
					t.Errorf("helm command mismatch:\nexpected: %s\nactual:   %s", test.expected, actual)
				}
			}
		})
	}
}
