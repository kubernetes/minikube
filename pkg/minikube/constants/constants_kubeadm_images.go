/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package constants

var (
	KubeadmImages = map[string]map[string]string{
		"v1.22.0-rc.0": {
			"k8s.gcr.io/coredns/coredns":         "v1.8.4",
			"k8s.gcr.io/etcd":                    "3.4.13-0",
			"k8s.gcr.io/kube-apiserver":          "v1.21.3",
			"k8s.gcr.io/kube-controller-manager": "v1.21.3",
			"k8s.gcr.io/kube-proxy":              "v1.21.3",
			"k8s.gcr.io/kube-scheduler":          "v1.21.3",
			"k8s.gcr.io/pause":                   "3.5",
		},
		"v1.21.3": {
			"k8s.gcr.io/coredns/coredns":         "v1.8.0",
			"k8s.gcr.io/etcd":                    "3.4.13-0",
			"k8s.gcr.io/kube-apiserver":          "v1.21.3",
			"k8s.gcr.io/kube-controller-manager": "v1.21.3",
			"k8s.gcr.io/kube-proxy":              "v1.21.3",
			"k8s.gcr.io/kube-scheduler":          "v1.21.3",
			"k8s.gcr.io/pause":                   "3.4.1",
		},
		"v1.21.0": {
			"k8s.gcr.io/coredns/coredns":         "v1.8.0",
			"k8s.gcr.io/etcd":                    "3.4.13-0",
			"k8s.gcr.io/kube-apiserver":          "v1.21.3",
			"k8s.gcr.io/kube-controller-manager": "v1.21.3",
			"k8s.gcr.io/kube-proxy":              "v1.21.3",
			"k8s.gcr.io/kube-scheduler":          "v1.21.3",
			"k8s.gcr.io/pause":                   "3.4.1",
		},
		"v1.20.0": {
			"k8s.gcr.io/coredns":                 "1.7.0",
			"k8s.gcr.io/etcd":                    "3.4.13-0",
			"k8s.gcr.io/kube-apiserver":          "v1.20.9",
			"k8s.gcr.io/kube-controller-manager": "v1.20.9",
			"k8s.gcr.io/kube-proxy":              "v1.20.9",
			"k8s.gcr.io/kube-scheduler":          "v1.20.9",
			"k8s.gcr.io/pause":                   "3.2",
		},
		"v1.19.0": {
			"k8s.gcr.io/coredns":                 "1.7.0",
			"k8s.gcr.io/etcd":                    "3.4.9-1",
			"k8s.gcr.io/kube-apiserver":          "v1.19.13",
			"k8s.gcr.io/kube-controller-manager": "v1.19.13",
			"k8s.gcr.io/kube-proxy":              "v1.19.13",
			"k8s.gcr.io/kube-scheduler":          "v1.19.13",
			"k8s.gcr.io/pause":                   "3.2",
		},
		"v1.18.0": {
			"k8s.gcr.io/coredns":                 "1.6.7",
			"k8s.gcr.io/etcd":                    "3.4.3-0",
			"k8s.gcr.io/kube-apiserver":          "v1.18.20",
			"k8s.gcr.io/kube-controller-manager": "v1.18.20",
			"k8s.gcr.io/kube-proxy":              "v1.18.20",
			"k8s.gcr.io/kube-scheduler":          "v1.18.20",
			"k8s.gcr.io/pause":                   "3.2",
		},
		"v1.17.0": {
			"k8s.gcr.io/coredns":                 "1.6.5",
			"k8s.gcr.io/etcd":                    "3.4.3-0",
			"k8s.gcr.io/kube-apiserver":          "v1.17.17",
			"k8s.gcr.io/kube-controller-manager": "v1.17.17",
			"k8s.gcr.io/kube-proxy":              "v1.17.17",
			"k8s.gcr.io/kube-scheduler":          "v1.17.17",
			"k8s.gcr.io/pause":                   "3.1",
		},
		"v1.16.1": {
			"mirror.k8s.io/kube-proxy":              "v1.16.1",
			"mirror.k8s.io/kube-scheduler":          "v1.16.1",
			"mirror.k8s.io/kube-controller-manager": "v1.16.1",
			"mirror.k8s.io/kube-apiserver":          "v1.16.1",
			"mirror.k8s.io/coredns":                 "1.6.2",
			"mirror.k8s.io/etcd":                    "3.3.15-0",
			"mirror.k8s.io/pause":                   "3.1",
		},
		"v1.16.0": {
			"k8s.gcr.io/coredns":                 "1.6.2",
			"k8s.gcr.io/etcd":                    "3.3.15-0",
			"k8s.gcr.io/kube-apiserver":          "v1.16.15",
			"k8s.gcr.io/kube-controller-manager": "v1.16.15",
			"k8s.gcr.io/kube-proxy":              "v1.16.15",
			"k8s.gcr.io/kube-scheduler":          "v1.16.15",
			"k8s.gcr.io/pause":                   "3.1",
		},
		"v1.15.0": {
			"k8s.gcr.io/coredns":                 "1.3.1",
			"k8s.gcr.io/etcd":                    "3.3.10",
			"k8s.gcr.io/kube-apiserver":          "v1.15.12",
			"k8s.gcr.io/kube-controller-manager": "v1.15.12",
			"k8s.gcr.io/kube-proxy":              "v1.15.12",
			"k8s.gcr.io/kube-scheduler":          "v1.15.12",
			"k8s.gcr.io/pause":                   "3.1",
		},
		"v1.14.0": {
			"k8s.gcr.io/coredns":                 "1.3.1",
			"k8s.gcr.io/etcd":                    "3.3.10",
			"k8s.gcr.io/kube-apiserver":          "v1.14.10",
			"k8s.gcr.io/kube-controller-manager": "v1.14.10",
			"k8s.gcr.io/kube-proxy":              "v1.14.10",
			"k8s.gcr.io/kube-scheduler":          "v1.14.10",
			"k8s.gcr.io/pause":                   "3.1",
		},
		"v1.13.0": {
			"k8s.gcr.io/coredns":                 "1.2.6",
			"k8s.gcr.io/etcd":                    "3.2.24",
			"k8s.gcr.io/kube-apiserver":          "v1.13.12",
			"k8s.gcr.io/kube-controller-manager": "v1.13.12",
			"k8s.gcr.io/kube-proxy":              "v1.13.12",
			"k8s.gcr.io/kube-scheduler":          "v1.13.12",
			"k8s.gcr.io/pause":                   "3.1",
		},
		"v1.12.0": {
			"k8s.gcr.io/coredns":                 "1.2.2",
			"k8s.gcr.io/etcd":                    "3.2.24",
			"k8s.gcr.io/kube-apiserver":          "v1.21.3",
			"k8s.gcr.io/kube-controller-manager": "v1.21.3",
			"k8s.gcr.io/kube-proxy":              "v1.21.3",
			"k8s.gcr.io/kube-scheduler":          "v1.21.3",
			"k8s.gcr.io/pause":                   "3.1",
		},
		"v1.11.0": {
			"k8s.gcr.io/coredns":                       "1.1.3",
			"k8s.gcr.io/etcd-amd64":                    "3.2.18",
			"k8s.gcr.io/kube-apiserver-amd64":          "v1.11.10",
			"k8s.gcr.io/kube-controller-manager-amd64": "v1.11.10",
			"k8s.gcr.io/kube-proxy-amd64":              "v1.11.10",
			"k8s.gcr.io/kube-scheduler-amd64":          "v1.11.10",
			"k8s.gcr.io/pause-amd64":                   "3.1",
		},
	}
)
