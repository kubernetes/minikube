/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGeneratePodmanScripts(t *testing.T) {
	var tests = []struct {
		shell         string
		config        PodmanEnvConfig
		noProxyGetter *FakeNoProxyGetter
		wantSet       string
		wantUnset     string
	}{
		{
			"bash",
			PodmanEnvConfig{profile: "bash", driver: "kvm2", ssh: false, hostIP: "127.0.0.1", port: 2376, certsDir: "/certs", noProxy: false},
			nil,
			`export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_TLS_VERIFY="1"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_PODMAN="bash"

# To point your shell to minikube's Podman Docker-compatible service, run:
# eval $(minikube -p bash podman-env)
`,
			`unset DOCKER_HOST;
unset DOCKER_TLS_VERIFY;
unset DOCKER_CERT_PATH;
unset MINIKUBE_ACTIVE_PODMAN;
`,
		},
		{
			"bash",
			PodmanEnvConfig{profile: "bash", driver: "kvm2", ssh: true, hostIP: "127.0.0.1", port: 2376, certsDir: "/certs", noProxy: false},
			nil,
			`export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_TLS_VERIFY="1"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_PODMAN="bash"

# To point your shell to minikube's Podman Docker-compatible service, run:
# eval $(minikube -p bash podman-env)
`,
			`unset DOCKER_HOST;
unset DOCKER_TLS_VERIFY;
unset DOCKER_CERT_PATH;
unset MINIKUBE_ACTIVE_PODMAN;
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.config.profile, func(t *testing.T) {
			tc.config.EnvConfig.Shell = tc.shell
			var b []byte
			buf := bytes.NewBuffer(b)
			if err := podmanSetScript(tc.config, buf); err != nil {
				t.Errorf("setScript(%+v) error: %v", tc.config, err)
			}
			got := buf.String()
			if diff := cmp.Diff(tc.wantSet, got); diff != "" {
				t.Errorf("setScript(%+v) mismatch (-want +got):\n%s\n\nraw output:\n%s\nquoted: %q", tc.config, diff, got, got)
			}

			buf = bytes.NewBuffer(b)
			if err := podmanUnsetScript(tc.config, buf); err != nil {
				t.Errorf("unsetScript(%+v) error: %v", tc.config, err)
			}
			got = buf.String()
			if diff := cmp.Diff(tc.wantUnset, got); diff != "" {
				t.Errorf("unsetScript(%+v) mismatch (-want +got):\n%s\n\nraw output:\n%s\nquoted: %q", tc.config, diff, got, got)
			}

		})
	}
}