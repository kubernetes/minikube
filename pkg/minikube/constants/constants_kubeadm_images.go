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
		"v1.30.0-alpha.3": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.12-0",
			"pause":           "3.9",
		},
		"v1.29.2": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.28.7": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.27.11": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.26.14": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.30.0-alpha.2": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.12-0",
			"pause":           "3.9",
		},
		"v1.30.0-alpha.1": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.11-0",
			"pause":           "3.9",
		},
		"v1.29.1": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.28.6": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.27.10": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.26.13": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.28.5": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.27.9": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.26.12": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.29.0": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.29.0-rc.2": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.29.0-rc.1": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.29.0-rc.0": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.28.4": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.27.8": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.26.11": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.25.16": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.9-0",
			"pause":           "3.8",
		},
		"v1.29.0-alpha.3": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.10-0",
			"pause":           "3.9",
		},
		"v1.28.3": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.27.7": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.26.10": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.25.15": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.9-0",
			"pause":           "3.8",
		},
		"v1.29.0-alpha.2": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.29.0-alpha.1": {
			"coredns/coredns": "v1.11.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.28.2": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.27.6": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.26.9": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.14": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.28.1": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.27.5": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.26.8": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.13": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.24.17": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.7",
		},
		"v1.28.0": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.28.0-rc.1": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.28.0-rc.0": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.28.0-beta.0": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.27.4": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.26.7": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.12": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.24.16": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.7",
		},
		"v1.28.0-alpha.4": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.28.0-alpha.3": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.9-0",
			"pause":           "3.9",
		},
		"v1.27.3": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.26.6": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.11": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.24.15": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.7",
		},
		"v1.28.0-alpha.2": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.8-0",
			"pause":           "3.9",
		},
		"v1.28.0-alpha.1": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.8-0",
			"pause":           "3.9",
		},
		"v1.27.2": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.26.5": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.10": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.24.14": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.7",
		},
		"v1.27.1": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.26.4": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.9": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.24.13": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.7",
		},
		"v1.27.0": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.27.0-rc.1": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.27.0-rc.0": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.27.0-beta.0": {
			"coredns/coredns": "v1.10.1",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.26.3": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.8": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.24.12": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.7",
		},
		"v1.27.0-alpha.3": {
			"coredns/coredns": "v1.10.0",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.26.2": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.7": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.24.11": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.7",
		},
		"v1.23.17": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.6",
		},
		"v1.27.0-alpha.2": {
			"coredns/coredns": "v1.10.0",
			"etcd":            "3.5.7-0",
			"pause":           "3.9",
		},
		"v1.27.0-alpha.1": {
			"coredns/coredns": "v1.10.0",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.26.1": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.6": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.24.10": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.7",
		},
		"v1.23.16": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.6",
		},
		"v1.26.0": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.25.5": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.8",
		},
		"v1.24.9": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.7",
		},
		"v1.23.15": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.6-0",
			"pause":           "3.6",
		},
		"v1.22.17": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.6-0",
			"pause":           "3.5",
		},
		"v1.26.0-rc.1": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.6-0",
			"pause":           "3.9",
		},
		"v1.26.0-rc.0": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.5-0",
			"pause":           "3.9",
		},
		"v1.26.0-beta.0": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.5-0",
			"pause":           "3.8",
		},
		"v1.25.4": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.5-0",
			"pause":           "3.8",
		},
		"v1.24.8": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.5-0",
			"pause":           "3.7",
		},
		"v1.23.14": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.5-0",
			"pause":           "3.6",
		},
		"v1.22.16": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.26.0-alpha.3": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.5-0",
			"pause":           "3.8",
		},
		"v1.25.3": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.8",
		},
		"v1.24.7": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.23.13": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.26.0-alpha.2": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.5-0",
			"pause":           "3.8",
		},
		"v1.25.2": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.8",
		},
		"v1.24.6": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.23.12": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.15": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.26.0-alpha.1": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.5-0",
			"pause":           "3.8",
		},
		"v1.25.1": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.8",
		},
		"v1.24.5": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.23.11": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.14": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.25.0": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.8",
		},
		"v1.24.4": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.23.10": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.13": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.25.0-rc.1": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.8",
		},
		"v1.25.0-rc.0": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.8",
		},
		"v1.25.0-beta.0": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.8",
		},
		"v1.25.0-alpha.3": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.7",
		},
		"v1.24.3": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.23.9": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.12": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.25.0-alpha.2": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.7",
		},
		"v1.23.8": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.11": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.25.0-alpha.1": {
			"coredns/coredns": "v1.9.3",
			"etcd":            "3.5.4-0",
			"pause":           "3.7",
		},
		"v1.24.2": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.21.14": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.24.1": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.23.7": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.10": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.13": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.24.0": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.24.0-rc.1": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.24.0-rc.0": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.3-0",
			"pause":           "3.7",
		},
		"v1.23.6": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.9": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.12": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.24.0-beta.0": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.7",
		},
		"v1.24.0-alpha.4": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.23.5": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.8": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.11": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.23.4": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.7": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.10": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.24.0-alpha.3": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.24.0-alpha.2": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.23.3": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.23.2": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.6": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.9": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.15": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.23.1": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.5": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.8": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.14": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.24.0-alpha.1": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.23.0": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.23.0-rc.1": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.23.0-rc.0": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.23.0-beta.0": {
			"coredns/coredns": "v1.8.6",
			"etcd":            "3.5.1-0",
			"pause":           "3.6",
		},
		"v1.22.4": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.7": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.13": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.23.0-alpha.4": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.6",
		},
		"v1.22.3": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.6": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.12": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.16": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.23.0-alpha.3": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.6",
		},
		"v1.22.2": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.5": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.11": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.15": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.23.0-alpha.2": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.6",
		},
		"v1.23.0-alpha.1": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.22.1": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.4": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.10": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.14": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.22.0": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.22.0-rc.0": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.21.3": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.9": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.13": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.22.0-beta.2": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.22.0-beta.1": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-0",
			"pause":           "3.5",
		},
		"v1.22.0-beta.0": {
			"coredns/coredns": "v1.8.4",
			"etcd":            "3.5.0-rc.0-0",
			"pause":           "3.5",
		},
		"v1.21.2": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.8": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.12": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.20": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.22.0-alpha.3": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-3",
			"pause":           "3.5",
		},
		"v1.22.0-alpha.2": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-3",
			"pause":           "3.4.1",
		},
		"v1.21.1": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.7": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.11": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.19": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.22.0-alpha.1": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.6": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.10": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.18": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.21.0": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.21.0-rc.0": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.5": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.9": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.17": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.21.0-beta.1": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.21.0-beta.0": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.20.4": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.20.3": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.8": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.16": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.21.0-alpha.3": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.4.1",
		},
		"v1.21.0-alpha.2": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.2",
		},
		"v1.21.0-alpha.1": {
			"coredns/coredns": "v1.8.0",
			"etcd":            "3.4.13-0",
			"pause":           "3.2",
		},
		"v1.20.2": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.7": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.15": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.17": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.20.1": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.6": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.14": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.16": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.19.5": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.13": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.15": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.20.0": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.20.0-rc.0": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.20.0-beta.2": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.12": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.19.4": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.17.14": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.20.0-beta.1": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.20.0-beta.0": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.20.0-alpha.3": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.10": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.13": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.19.3": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.20.0-alpha.2": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.20.0-alpha.1": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.19.2": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.18.9": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.12": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.19.1": {
			"coredns": "1.7.0",
			"etcd":    "3.4.13-0",
			"pause":   "3.2",
		},
		"v1.16.15": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.19.0": {
			"coredns": "1.7.0",
			"etcd":    "3.4.9-1",
			"pause":   "3.2",
		},
		"v1.18.8": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.11": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.14": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.19.0-rc.4": {
			"coredns": "1.7.0",
			"etcd":    "3.4.9-1",
			"pause":   "3.2",
		},
		"v1.19.0-rc.3": {
			"coredns": "1.7.0",
			"etcd":    "3.4.9-1",
			"pause":   "3.2",
		},
		"v1.19.0-rc.2": {
			"coredns": "1.7.0",
			"etcd":    "3.4.9-1",
			"pause":   "3.2",
		},
		"v1.18.6": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.9": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.13": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.19.0-rc.1": {
			"coredns": "1.7.0",
			"etcd":    "3.4.7-0",
			"pause":   "3.2",
		},
		"v1.18.5": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.8": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.12": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.18.5-rc.1": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.8-rc.1": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.12-rc.1": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.18.4": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.7": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.11": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.19.0-beta.2": {
			"coredns": "1.6.7",
			"etcd":    "3.4.7-0",
			"pause":   "3.2",
		},
		"v1.19.0-beta.1": {
			"coredns": "1.6.7",
			"etcd":    "3.4.7-0",
			"pause":   "3.2",
		},
		"v1.18.3": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.6": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.10": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.19.0-beta.0": {
			"coredns": "1.6.7",
			"etcd":    "3.4.7-0",
			"pause":   "3.2",
		},
		"v1.19.0-alpha.3": {
			"coredns": "1.6.7",
			"etcd":    "3.4.7-0",
			"pause":   "3.2",
		},
		"v1.19.0-alpha.2": {
			"coredns": "1.6.7",
			"etcd":    "3.4.7-0",
			"pause":   "3.2",
		},
		"v1.18.2": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.5": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.9": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.18.1": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.19.0-alpha.1": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.18.0": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.18.0-rc.1": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.17.4": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.8": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.18.0-beta.2": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.18.0-beta.1": {
			"coredns": "1.6.7",
			"etcd":    "3.4.3-0",
			"pause":   "3.2",
		},
		"v1.18.0-alpha.5": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.17.3": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.7": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.18.0-alpha.3": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.18.0-alpha.2": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.17.2": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.6": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.16.5": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.17.1": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.18.0-alpha.1": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.4": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.17.0": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.17.0-rc.2": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.17.0-rc.1": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.17.0-beta.2": {
			"coredns": "1.6.5",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.16.3": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.17.0-beta.1": {
			"coredns": "1.6.2",
			"etcd":    "3.4.3-0",
			"pause":   "3.1",
		},
		"v1.17.0-alpha.3": {
			"coredns": "1.6.2",
			"etcd":    "3.3.17-0",
			"pause":   "3.1",
		},
		"v1.17.0-alpha.2": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.16.2": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.17.0-alpha.1": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.16.1": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
		"v1.16.0": {
			"coredns": "1.6.2",
			"etcd":    "3.3.15-0",
			"pause":   "3.1",
		},
	}
)
