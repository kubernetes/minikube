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
