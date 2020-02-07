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

func TestGenerateScripts(t *testing.T) {
	var tests = []struct {
		config        EnvConfig
		noProxyGetter *FakeNoProxyGetter
		wantSet       string
		wantUnset     string
	}{
		{
			EnvConfig{profile: "bash", shell: "bash", driver: "kvm2", hostIP: "127.0.0.1", certsDir: "/certs"},
			nil,
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="bash"

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p bash docker-env)
`,
			`unset DOCKER_TLS_VERIFY
unset DOCKER_HOST
unset DOCKER_CERT_PATH
unset MINIKUBE_ACTIVE_DOCKERD

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p bash docker-env)
`,
		},
		{
			EnvConfig{profile: "ipv6", shell: "bash", driver: "kvm2", hostIP: "fe80::215:5dff:fe00:a903", certsDir: "/certs"},
			nil,
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://[fe80::215:5dff:fe00:a903]:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="ipv6"

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p ipv6 docker-env)
`,
			`unset DOCKER_TLS_VERIFY
unset DOCKER_HOST
unset DOCKER_CERT_PATH
unset MINIKUBE_ACTIVE_DOCKERD

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p ipv6 docker-env)
`,
		},
		{
			EnvConfig{profile: "fish", shell: "fish", driver: "kvm2", hostIP: "127.0.0.1", certsDir: "/certs"},
			nil,
			`set -gx DOCKER_TLS_VERIFY "1";
set -gx DOCKER_HOST "tcp://127.0.0.1:2376";
set -gx DOCKER_CERT_PATH "/certs";
set -gx MINIKUBE_ACTIVE_DOCKERD "fish";

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval (minikube -p fish docker-env)
`,
			`set -e DOCKER_TLS_VERIFY;
set -e DOCKER_HOST;
set -e DOCKER_CERT_PATH;
set -e MINIKUBE_ACTIVE_DOCKERD;

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval (minikube -p fish docker-env)
`,
		},
		{
			EnvConfig{profile: "powershell", shell: "powershell", driver: "hyperv", hostIP: "192.168.0.1", certsDir: "/certs"},
			nil,
			`$Env:DOCKER_TLS_VERIFY = "1"
$Env:DOCKER_HOST = "tcp://192.168.0.1:2376"
$Env:DOCKER_CERT_PATH = "/certs"
$Env:MINIKUBE_ACTIVE_DOCKERD = "powershell"

# Please run command bellow to point your shell to minikube's docker-daemon :
# & minikube -p powershell docker-env | Invoke-Expression
	`,

			`Remove-Item Env:\\DOCKER_TLS_VERIFY
Remove-Item Env:\\DOCKER_HOST
Remove-Item Env:\\DOCKER_CERT_PATH
Remove-Item Env:\\MINIKUBE_ACTIVE_DOCKERD

# Please run command bellow to point your shell to minikube's docker-daemon :
# & minikube -p powershell docker-env | Invoke-Expression
	`,
		},
		{
			EnvConfig{profile: "cmd", shell: "cmd", driver: "hyperv", hostIP: "192.168.0.1", certsDir: "/certs"},
			nil,
			`SET DOCKER_TLS_VERIFY=1
SET DOCKER_HOST=tcp://192.168.0.1:2376
SET DOCKER_CERT_PATH=/certs
SET MINIKUBE_ACTIVE_DOCKERD=cmd

REM Please run command bellow to point your shell to minikube's docker-daemon :
REM @FOR /f "tokens=*" %i IN ('minikube -p cmd docker-env') DO @%i
	`,

			`SET DOCKER_TLS_VERIFY=
SET DOCKER_HOST=
SET DOCKER_CERT_PATH=
SET MINIKUBE_ACTIVE_DOCKERD=

REM Please run command bellow to point your shell to minikube's docker-daemon :
REM @FOR /f "tokens=*" %i IN ('minikube -p cmd docker-env') DO @%i
	`,
		},
		{
			EnvConfig{profile: "emacs", shell: "emacs", driver: "hyperv", hostIP: "192.168.0.1", certsDir: "/certs"},
			nil,
			`(setenv "DOCKER_TLS_VERIFY" "1")
(setenv "DOCKER_HOST" "tcp://192.168.0.1:2376")
(setenv "DOCKER_CERT_PATH" "/certs")
(setenv "MINIKUBE_ACTIVE_DOCKERD" "emacs")

;; Please run command bellow to point your shell to minikube's docker-daemon :
;; (with-temp-buffer (shell-command "minikube -p emacs docker-env" (current-buffer)) (eval-buffer))
	`,
			`(setenv "DOCKER_TLS_VERIFY" nil)
(setenv "DOCKER_HOST" nil)
(setenv "DOCKER_CERT_PATH" nil)
(setenv "MINIKUBE_ACTIVE_DOCKERD" nil)

;; Please run command bellow to point your shell to minikube's docker-daemon :
;; (with-temp-buffer (shell-command "minikube -p emacs docker-env" (current-buffer)) (eval-buffer))
	`,
		},
		{
			EnvConfig{profile: "bash-no-proxy", shell: "bash", driver: "kvm2", hostIP: "127.0.0.1", certsDir: "/certs", noProxy: true},
			&FakeNoProxyGetter{"NO_PROXY", "127.0.0.1"},
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="bash-no-proxy"
export NO_PROXY="127.0.0.1"

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p bash-no-proxy docker-env)
`,

			`unset DOCKER_TLS_VERIFY
unset DOCKER_HOST
unset DOCKER_CERT_PATH
unset MINIKUBE_ACTIVE_DOCKERD
unset NO_PROXY127.0.0.1

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p bash-no-proxy docker-env)
`,
		},
		{
			EnvConfig{profile: "bash-no-proxy-lower", shell: "bash", driver: "kvm2", hostIP: "127.0.0.1", certsDir: "/certs", noProxy: true},
			&FakeNoProxyGetter{"no_proxy", "127.0.0.1"},
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="bash-no-proxy-lower"
export no_proxy="127.0.0.1"

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p bash-no-proxy-lower docker-env)
`,

			`unset DOCKER_TLS_VERIFY
unset DOCKER_HOST
unset DOCKER_CERT_PATH
unset MINIKUBE_ACTIVE_DOCKERD
unset no_proxy127.0.0.1

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p bash-no-proxy-lower docker-env)
`,
		},
		{
			EnvConfig{profile: "bash-no-proxy-idempotent", shell: "bash", driver: "kvm2", hostIP: "127.0.0.1", certsDir: "/certs", noProxy: true},
			&FakeNoProxyGetter{"no_proxy", "127.0.0.1"},
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="bash-no-proxy-idempotent"
export no_proxy="127.0.0.1"

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p bash-no-proxy-idempotent docker-env)
`,

			`unset DOCKER_TLS_VERIFY
unset DOCKER_HOST
unset DOCKER_CERT_PATH
unset MINIKUBE_ACTIVE_DOCKERD
unset no_proxy127.0.0.1

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p bash-no-proxy-idempotent docker-env)
`,
		},
		{
			EnvConfig{profile: "sh-no-proxy-add", shell: "bash", driver: "kvm2", hostIP: "127.0.0.1", certsDir: "/certs", noProxy: true},
			&FakeNoProxyGetter{"NO_PROXY", "192.168.0.1,10.0.0.4"},
			`export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:2376"
export DOCKER_CERT_PATH="/certs"
export MINIKUBE_ACTIVE_DOCKERD="sh-no-proxy-add"
export NO_PROXY="192.168.0.1,10.0.0.4,127.0.0.1"

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p sh-no-proxy-add docker-env)
`,

			`unset DOCKER_TLS_VERIFY
unset DOCKER_HOST
unset DOCKER_CERT_PATH
unset MINIKUBE_ACTIVE_DOCKERD
unset NO_PROXY192.168.0.1,10.0.0.4

# Please run command bellow to point your shell to minikube's docker-daemon :
# eval $(minikube -p sh-no-proxy-add docker-env)
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.config.profile, func(t *testing.T) {
			defaultNoProxyGetter = tc.noProxyGetter
			var b []byte
			buf := bytes.NewBuffer(b)
			if err := setScript(tc.config, buf); err != nil {
				t.Errorf("setScript(%+v) error: %v", tc.config, err)
			}
			got := buf.String()
			if diff := cmp.Diff(tc.wantSet, got); diff != "" {
				t.Errorf("setScript(%+v) mismatch (-want +got):\n%s\n\nraw output:\n%s\nquoted: %q", tc.config, diff, got, got)
			}

			buf = bytes.NewBuffer(b)
			if err := unsetScript(tc.config, buf); err != nil {
				t.Errorf("unsetScript(%+v) error: %v", tc.config, err)
			}
			got = buf.String()
			if diff := cmp.Diff(tc.wantUnset, got); diff != "" {
				t.Errorf("unsetScript(%+v) mismatch (-want +got):\n%s\n\nraw output:\n%s\nquoted: %q", tc.config, diff, got, got)
			}

		})
	}
}
