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

package deployer

type MiniTestBoskosConfig struct {
	GCPZone       string `json:"GCPZone,omitempty"`
	InstanceImage string `json:"InstanceImage,omitempty"`
	InstanceType  string `json:"InstanceType,omitempty"`
	DiskGiB       int    `json:"DiskGiB,omitempty"`
	// Boskos flags correspond to https://github.com/kubernetes-sigs/kubetest2/blob/71238a9645df6fbd7eaac9a36f635c22f1566168/kubetest2-gce/deployer/deployer.go
	BoskosAcquireTimeoutSeconds    int
	BoskosHeartbeatIntervalSeconds int
	BoskosLocation                 string
}
