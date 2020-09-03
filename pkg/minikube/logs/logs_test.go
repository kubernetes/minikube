/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package logs

import (
	"testing"
)

func TestIsProblem(t *testing.T) {
	var tests = []struct {
		name  string
		want  bool
		input string
	}{
		{"almost", false, "F2350 I would love to be an unknown flag, but I am not -- :( --"},
		{"apiserver-required-flag #1962", true, "error: [service-account-issuer is a required flag when BoundServiceAccountTokenVolume is enabled, --service-account-signing-key-file and --service-account-issuer are required flags"},
		{"kubelet-eviction #3611", true, `eviction_manager.go:187] eviction manager: pods kube-proxy-kfs8p_kube-system(27fd6b4b-33cf-11e9-ae1d-00155d4b0144) evicted, waiting for pod to be cleaned up`},
		{"kubelet-unknown-flag #3655", true, "F0212 14:55:46.443031    2693 server.go:148] unknown flag: --AllowedUnsafeSysctls"},
		{"apiserver-auth-mode #2852", true, `{"log":"Error: unknown flag: --Authorization.Mode\n","stream":"stderr","time":"2018-06-17T22:16:35.134161966Z"}`},
		{"apiserver-admission #3524", true, "error: unknown flag: --GenericServerRunOptions.AdmissionControl"},
		{"no-providers-available #3818", true, ` kubelet.go:1662] Failed creating a mirror pod for "kube-apiserver-minikube_kube-system(c7d572aebd3d33b17fa78ae6395b6d0a)": pods "kube-apiserver-minikube" is forbidden: no providers available to validate pod request`},
		{"no-objects-passed-to-apply #4010", false, "error: no objects passed to apply"},
		{"bad-certificate #4251", true, "log.go:172] http: TLS handshake error from 127.0.0.1:49200: remote error: tls: bad certificate"},
		{"ephemeral-eviction #5355", true, " eviction_manager.go:419] eviction manager: unexpected error when attempting to reduce ephemeral-storage pressure: wanted to free 9223372036854775807 bytes, but freed 0 bytes space with errors in image deletion"},
		{"anonymous-auth", true, "AnonymousAuth is not allowed with the AlwaysAllow authorizer. Resetting AnonymousAuth to false. You should use a different authorizer"},
		{"disk-pressure #7073", true, "eviction_manager.go:159] Failed to admit pod kindnet-jpzzf_kube-system(b63b1ee0-0fc6-428f-8e67-e357464f579c) - node has conditions: [DiskPressure]"},
		{"csi timeout", true, `Failed to initialize CSINodeInfo: error updating CSINode annotation: timed out waiting for the condition; caused by: csinodes.storage.k8s.io "m01" is forbidden: User "system:node:m01" cannot get resource "csinodes" in API group "storage.k8s.io" at the cluster scope`},
		{"node registration permissions", true, `Unable to register node "m01" with API server: nodes is forbidden: User "system:node:m01" cannot create resource "nodes" in API group "" at the cluster scope`},
		{"regular kubelet refused", false, `kubelet_node_status.go:92] Unable to register node "m01" with API server: Post https://localhost:8443/api/v1/nodes: dial tcp 127.0.0.1:8443: connect: connection refused`},
		{"regular csi refused", false, `Failed to initialize CSINodeInfo: error updating CSINode annotation: timed out waiting for the condition; caused by: Get https://localhost:8443/apis/storage.k8s.io/v1/csinodes/m01: dial tcp 127.0.0.1:8443: connect: connection refused`},
		{"apiserver crashloop", true, `pod_workers.go:191] Error syncing pod 9f8ee739bd14e8733f807eb2be99768f ("kube-apiserver-m01_kube-system(9f8ee739bd14e8733f807eb2be99768f)"), skipping: failed to "StartContainer" for "kube-apiserver" with CrashLoopBackOff: "back-off 10s restarting failed container=kube-apiserver pod=kube-apiserver-m01_kube-system(9f8ee739bd14e8733f807eb2be99768f)`},
		{"kubelet node timeout", false, `failed to ensure node lease exists, will retry in 6.4s, error: Get https://localhost:8443/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/m01?timeout=10s: dial tcp 127.0.0.1:8443: connect: connection refused`},
		{"rbac misconfiguration", true, `leases.coordination.k8s.io "m01" is forbidden: User "system:node:m01" cannot get resource "leases" in API group "coordination.k8s.io" in the namespace "kube-node-lease"`},
		{"regular controller init", false, `error retrieving resource lock kube-system/kube-controller-manager: endpoints "kube-controller-manager" is forbidden: User "system:kube-controller-manager" cannot get resource "endpoints" in API group "" in the namespace "kube-system"`},
		{"regular scheduler services init", false, ` k8s.io/client-go/informers/factory.go:135: Failed to list *v1.Service: services is forbidden: User "system:kube-scheduler" cannot list resource "services" in API group "" at the cluster scope`},
		{"regular scheduler nodes init", false, `k8s.io/client-go/informers/factory.go:135: Failed to list *v1.Node: nodes is forbidden: User "system:kube-scheduler" cannot list resource "nodes" in API group "" at the cluster scope`},
		{"kubelet rbac fail", true, `k8s.io/kubernetes/pkg/kubelet/kubelet.go:526: Failed to list *v1.Node: nodes "m01" is forbidden: User "system:node:m01" cannot list resource "nodes" in API group "" at the cluster scope`},
		{"kubelet pids cgroup", true, `Failed to start ContainerManager failed to initialize top level QOS containers: failed to update top level Burstable QOS cgroup : failed to set supported cgroup subsystems for cgroup [kubepods burstable]: failed to find subsystem mount for required subsystem: pids`},
		{"docker cgroups v2 fail", true, `failed to start daemon: Devices cgroup isn't mounted`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsProblem(tc.input)
			if got != tc.want {
				t.Fatalf("IsProblem(%s)=%v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
