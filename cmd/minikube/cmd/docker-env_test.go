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

type FakeNoProxyGetter struct {
	NoProxyVar   string
	NoProxyValue string
}

func (f FakeNoProxyGetter) GetNoProxyVar() (string, string) {
	return f.NoProxyVar, f.NoProxyValue
}

func TestGenerateDockerScripts(t *testing.T) {
	var tests = []struct {
		shell         string
		config        DockerEnvConfig
		noProxyGetter *FakeNoProxyGetter
		wantSet       string
		wantUnset     string
	}{
		{
			"bash",
			DockerEnvConfig{profile: "dockerdriver", driver: "docker", hostIP: "127.0.0.1", port: 32842, certsDir: "/certs"},
			nil,
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:32842"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="dockerdriver"

# To point your shell to minikube's docker-daemon, run:
# eval $(minikube -p dockerdriver docker-env)
`,
			`unset DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH MINIKUBE_ACTIVE_DOCKERD
`,
		},
		{
			"bash",
			DockerEnvConfig{profile: "dockerdriver", driver: "docker", ssh: true, username: "root", hostname: "host", sshport: 22},
			nil,
			`export DOCKER_HOST="ssh://root@host:22"
export MINIKUBE_ACTIVE_DOCKERD="dockerdriver"

# To point your shell to minikube's docker-daemon, run:
# eval $(minikube -p dockerdriver docker-env --ssh-host)
`,
			`unset DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH MINIKUBE_ACTIVE_DOCKERD
`,
		},
		{
			"bash",
			DockerEnvConfig{profile: "bash", driver: "kvm2", hostIP: "127.0.0.1", port: 2376, certsDir: "/certs"},
			nil,
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="bash"

# To point your shell to minikube's docker-daemon, run:
# eval $(minikube -p bash docker-env)
`,
			`unset DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH MINIKUBE_ACTIVE_DOCKERD
`,
		},
		{
			"bash",
			DockerEnvConfig{profile: "ipv6", driver: "kvm2", hostIP: "fe80::215:5dff:fe00:a903", port: 2376, certsDir: "/certs"},
			nil,
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://[fe80::215:5dff:fe00:a903]:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="ipv6"

# To point your shell to minikube's docker-daemon, run:
# eval $(minikube -p ipv6 docker-env)
`,
			`unset DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH MINIKUBE_ACTIVE_DOCKERD
`,
		},
		{
			"fish",
			DockerEnvConfig{profile: "fish", driver: "kvm2", hostIP: "127.0.0.1", port: 2376, certsDir: "/certs"},
			nil,
			`set -gx DOCKER_TLS_VERIFY "1";
set -gx DOCKER_HOST "tcp://127.0.0.1:2376";
set -gx DOCKER_CERT_PATH "/certs";
set -gx MINIKUBE_ACTIVE_DOCKERD "fish";

# To point your shell to minikube's docker-daemon, run:
# minikube -p fish docker-env | source
`,
			`set -e DOCKER_TLS_VERIFY;
set -e DOCKER_HOST;
set -e DOCKER_CERT_PATH;
set -e MINIKUBE_ACTIVE_DOCKERD;
`,
		},
		{
			"powershell",
			DockerEnvConfig{profile: "powershell", driver: "hyperv", hostIP: "192.168.0.1", port: 2376, certsDir: "/certs"},
			nil,
			`$Env:DOCKER_TLS_VERIFY = "1"
$Env:DOCKER_HOST = "tcp://192.168.0.1:2376"
$Env:DOCKER_CERT_PATH = "/certs"
$Env:MINIKUBE_ACTIVE_DOCKERD = "powershell"
# To point your shell to minikube's docker-daemon, run:
# & minikube -p powershell docker-env | Invoke-Expression
`,

			`Remove-Item Env:\\DOCKER_TLS_VERIFY Env:\\DOCKER_HOST Env:\\DOCKER_CERT_PATH Env:\\MINIKUBE_ACTIVE_DOCKERD
`,
		},
		{
			"cmd",
			DockerEnvConfig{profile: "cmd", driver: "hyperv", hostIP: "192.168.0.1", port: 2376, certsDir: "/certs"},
			nil,
			`SET DOCKER_TLS_VERIFY=1
SET DOCKER_HOST=tcp://192.168.0.1:2376
SET DOCKER_CERT_PATH=/certs
SET MINIKUBE_ACTIVE_DOCKERD=cmd
REM To point your shell to minikube's docker-daemon, run:
REM @FOR /f "tokens=*" %i IN ('minikube -p cmd docker-env') DO @%i
`,

			`SET DOCKER_TLS_VERIFY=
SET DOCKER_HOST=
SET DOCKER_CERT_PATH=
SET MINIKUBE_ACTIVE_DOCKERD=
`,
		},
		{
			"emacs",
			DockerEnvConfig{profile: "emacs", driver: "hyperv", hostIP: "192.168.0.1", port: 2376, certsDir: "/certs"},
			nil,
			`(setenv "DOCKER_TLS_VERIFY" "1")
(setenv "DOCKER_HOST" "tcp://192.168.0.1:2376")
(setenv "DOCKER_CERT_PATH" "/certs")
(setenv "MINIKUBE_ACTIVE_DOCKERD" "emacs")
;; To point your shell to minikube's docker-daemon, run:
;; (with-temp-buffer (shell-command "minikube -p emacs docker-env" (current-buffer)) (eval-buffer))
`,
			`(setenv "DOCKER_TLS_VERIFY" nil)
(setenv "DOCKER_HOST" nil)
(setenv "DOCKER_CERT_PATH" nil)
(setenv "MINIKUBE_ACTIVE_DOCKERD" nil)
`,
		},
		{
			"bash",
			DockerEnvConfig{profile: "bash-no-proxy", driver: "kvm2", hostIP: "127.0.0.1", port: 2376, certsDir: "/certs", noProxy: true},
			&FakeNoProxyGetter{"NO_PROXY", "127.0.0.1"},
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="bash-no-proxy"
export NO_PROXY="127.0.0.1"

# To point your shell to minikube's docker-daemon, run:
# eval $(minikube -p bash-no-proxy docker-env)
`,

			`unset DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH MINIKUBE_ACTIVE_DOCKERD NO_PROXY
`,
		},
		{
			"bash",
			DockerEnvConfig{profile: "bash-no-proxy-lower", driver: "kvm2", hostIP: "127.0.0.1", port: 2376, certsDir: "/certs", noProxy: true},
			&FakeNoProxyGetter{"no_proxy", "127.0.0.1"},
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="bash-no-proxy-lower"
export no_proxy="127.0.0.1"

# To point your shell to minikube's docker-daemon, run:
# eval $(minikube -p bash-no-proxy-lower docker-env)
`,

			`unset DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH MINIKUBE_ACTIVE_DOCKERD no_proxy
`,
		},
		{
			"powershell",
			DockerEnvConfig{profile: "powershell-no-proxy-idempotent", driver: "hyperv", hostIP: "192.168.0.1", port: 2376, certsDir: "/certs", noProxy: true},
			&FakeNoProxyGetter{"no_proxy", "192.168.0.1"},
			`$Env:DOCKER_TLS_VERIFY = "1"
$Env:DOCKER_HOST = "tcp://192.168.0.1:2376"
$Env:DOCKER_CERT_PATH = "/certs"
$Env:MINIKUBE_ACTIVE_DOCKERD = "powershell-no-proxy-idempotent"
$Env:no_proxy = "192.168.0.1"
# To point your shell to minikube's docker-daemon, run:
# & minikube -p powershell-no-proxy-idempotent docker-env | Invoke-Expression
`,

			`Remove-Item Env:\\DOCKER_TLS_VERIFY Env:\\DOCKER_HOST Env:\\DOCKER_CERT_PATH Env:\\MINIKUBE_ACTIVE_DOCKERD Env:\\no_proxy
`,
		},
		{
			"bash",
			DockerEnvConfig{profile: "sh-no-proxy-add", driver: "kvm2", hostIP: "127.0.0.1", port: 2376, certsDir: "/certs", noProxy: true},
			&FakeNoProxyGetter{"NO_PROXY", "192.168.0.1,10.0.0.4"},
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="sh-no-proxy-add"
export NO_PROXY="192.168.0.1,10.0.0.4,127.0.0.1"

# To point your shell to minikube's docker-daemon, run:
# eval $(minikube -p sh-no-proxy-add docker-env)
`,

			`unset DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH MINIKUBE_ACTIVE_DOCKERD NO_PROXY
`,
		},
		{
			"none",
			DockerEnvConfig{profile: "noneshell", driver: "docker", hostIP: "127.0.0.1", port: 32842, certsDir: "/certs"},
			nil,
			`DOCKER_TLS_VERIFY=1
DOCKER_HOST=tcp://127.0.0.1:32842
DOCKER_CERT_PATH=/certs
MINIKUBE_ACTIVE_DOCKERD=noneshell
`,
			`DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH MINIKUBE_ACTIVE_DOCKERD
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.config.profile, func(t *testing.T) {
			tc.config.EnvConfig.Shell = tc.shell
			defaultNoProxyGetter = tc.noProxyGetter
			var b []byte
			buf := bytes.NewBuffer(b)
			if err := dockerSetScript(tc.config, buf); err != nil {
				t.Errorf("setScript(%+v) error: %v", tc.config, err)
			}
			got := buf.String()
			if diff := cmp.Diff(tc.wantSet, got); diff != "" {
				t.Errorf("setScript(%+v) mismatch (-want +got):\n%s\n\nraw output:\n%s\nquoted: %q", tc.config, diff, got, got)
			}

			buf = bytes.NewBuffer(b)
			if err := dockerUnsetScript(tc.config, buf); err != nil {
				t.Errorf("unsetScript(%+v) error: %v", tc.config, err)
			}
			got = buf.String()
			if diff := cmp.Diff(tc.wantUnset, got); diff != "" {
				t.Errorf("unsetScript(%+v) mismatch (-want +got):\n%s\n\nraw output:\n%s\nquoted: %q", tc.config, diff, got, got)
			}

		})
	}
}
