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
			description: "enable addon with custom image repository",
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
			description: "enable addon with Aliyun mirror",
			chart: &assets.HelmChart{
				Name:        "metrics-server",
				Repo:        "metrics-server/metrics-server",
				Namespace:   "kube-system",
				ImageSetKey: "image.repository",
			},
			enable:          true,
			imageRepository: "registry.cn-hangzhou.aliyuncs.com/google_containers",
			expected:        "sudo KUBECONFIG=/var/lib/minikube/kubeconfig helm upgrade --install metrics-server metrics-server/metrics-server --create-namespace --namespace kube-system --set image.repository=registry.cn-hangzhou.aliyuncs.com/google_containers",
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
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			command := helmUninstallOrInstall(context.Background(), test.chart, test.enable, test.imageRepository)
			actual := strings.Join(command.Args, " ")
			if actual != test.expected {
				t.Errorf("helm command mismatch:\nexpected: %s\nactual:   %s", test.expected, actual)
			}
		})
	}
}
