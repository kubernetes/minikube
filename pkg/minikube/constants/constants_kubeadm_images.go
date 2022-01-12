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
		"v1.24": {
			"coredns/coredns":         "v1.8.6",
			"etcd":                    "3.5.1-0",
			"kube-apiserver":          "v1.23.1",
			"kube-controller-manager": "v1.23.1",
			"kube-proxy":              "v1.23.1",
			"kube-scheduler":          "v1.23.1",
			"pause":                   "3.6",
		},
		"v1.23": {
			"coredns/coredns":         "v1.8.6",
			"etcd":                    "3.5.1-0",
			"kube-apiserver":          "v1.23.1",
			"kube-controller-manager": "v1.23.1",
			"kube-proxy":              "v1.23.1",
			"kube-scheduler":          "v1.23.1",
			"pause":                   "3.6",
		},
		"v1.22": {
			"coredns/coredns":         "v1.8.4",
			"etcd":                    "3.5.0-0",
			"kube-apiserver":          "v1.22.4",
			"kube-controller-manager": "v1.22.4",
			"kube-proxy":              "v1.22.4",
			"kube-scheduler":          "v1.22.4",
			"pause":                   "3.5",
		},
		"v1.21": {
			"coredns/coredns":         "v1.8.0",
			"etcd":                    "3.4.13-0",
			"kube-apiserver":          "v1.21.6",
			"kube-controller-manager": "v1.21.6",
			"kube-proxy":              "v1.21.6",
			"kube-scheduler":          "v1.21.6",
			"pause":                   "3.4.1",
		},
		"v1.20": {
			"coredns":                 "1.7.0",
			"etcd":                    "3.4.13-0",
			"kube-apiserver":          "v1.20.12",
			"kube-controller-manager": "v1.20.12",
			"kube-proxy":              "v1.20.12",
			"kube-scheduler":          "v1.20.12",
			"pause":                   "3.2",
		},
		"v1.19": {
			"coredns":                 "1.7.0",
			"etcd":                    "3.4.9-1",
			"kube-apiserver":          "v1.19.16",
			"kube-controller-manager": "v1.19.16",
			"kube-proxy":              "v1.19.16",
			"kube-scheduler":          "v1.19.16",
			"pause":                   "3.2",
		},
		"v1.18": {
			"coredns":                 "1.6.7",
			"etcd":                    "3.4.3-0",
			"kube-apiserver":          "v1.18.20",
			"kube-controller-manager": "v1.18.20",
			"kube-proxy":              "v1.18.20",
			"kube-scheduler":          "v1.18.20",
			"pause":                   "3.2",
		},
		"v1.17": {
			"coredns":                 "1.6.5",
			"etcd":                    "3.4.3-0",
			"kube-apiserver":          "v1.17.17",
			"kube-controller-manager": "v1.17.17",
			"kube-proxy":              "v1.17.17",
			"kube-scheduler":          "v1.17.17",
			"pause":                   "3.1",
		},
		"v1.16": {
			"coredns":                 "1.6.2",
			"etcd":                    "3.3.15-0",
			"kube-apiserver":          "v1.16.15",
			"kube-controller-manager": "v1.16.15",
			"kube-proxy":              "v1.16.15",
			"kube-scheduler":          "v1.16.15",
			"pause":                   "3.1",
		},
		"v1.15": {
			"coredns":                 "1.3.1",
			"etcd":                    "3.3.10",
			"kube-apiserver":          "v1.15.12",
			"kube-controller-manager": "v1.15.12",
			"kube-proxy":              "v1.15.12",
			"kube-scheduler":          "v1.15.12",
			"pause":                   "3.1",
		},
		"v1.14": {
			"coredns":                 "1.3.1",
			"etcd":                    "3.3.10",
			"kube-apiserver":          "v1.14.10",
			"kube-controller-manager": "v1.14.10",
			"kube-proxy":              "v1.14.10",
			"kube-scheduler":          "v1.14.10",
			"pause":                   "3.1",
		},
		"v1.13": {
			"coredns":                 "1.2.6",
			"etcd":                    "3.2.24",
			"kube-apiserver":          "v1.13.12",
			"kube-controller-manager": "v1.13.12",
			"kube-proxy":              "v1.13.12",
			"kube-scheduler":          "v1.13.12",
			"pause":                   "3.1",
		},
		"v1.12": {
			"coredns":                 "1.2.2",
			"etcd":                    "3.2.24",
			"kube-apiserver":          "v1.22.3",
			"kube-controller-manager": "v1.22.3",
			"kube-proxy":              "v1.22.3",
			"kube-scheduler":          "v1.22.3",
			"pause":                   "3.1",
		},
		"v1.11": {
			"coredns":                       "1.1.3",
			"etcd-amd64":                    "3.2.18",
			"kube-apiserver-amd64":          "v1.11.10",
			"kube-controller-manager-amd64": "v1.11.10",
			"kube-proxy-amd64":              "v1.11.10",
			"kube-scheduler-amd64":          "v1.11.10",
			"pause-amd64":                   "3.1",
		},
	}
)
