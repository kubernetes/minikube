/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package kubeadm

import (
	"errors"
	"io"
	"os/exec"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
)

type identityRunner struct {
	*command.FakeCommandRunner
	existing map[string]bool
	denied   bool
	copied   assets.CopyableFile
}

// RunCmd simulates explicit metadata probes and directory creation only.
func (r *identityRunner) RunCmd(cmd *exec.Cmd) (*command.RunResult, error) {
	result := &command.RunResult{Args: cmd.Args}
	if r.denied {
		result.ExitCode = 255
		return result, errors.New("synthetic transport denial")
	}
	if len(cmd.Args) == 4 && cmd.Args[1] == "test" && cmd.Args[2] == "-s" {
		if r.existing[cmd.Args[3]] {
			return result, nil
		}
		result.ExitCode = 1
		return result, errors.New("synthetic absent file")
	}
	return result, nil
}

// Copy records one private configuration write without reading client key bytes.
func (r *identityRunner) Copy(file assets.CopyableFile) error {
	r.copied = file
	return nil
}

// TestRestoreManagedWorkerIdentity covers retained credentials, new join,
// preserved custom configuration, missing CA and transport denial.
func TestRestoreManagedWorkerIdentity(t *testing.T) {
	for _, scenario := range []string{"restore", "new", "existing", "missing-ca", "denied"} {
		t.Run(scenario, func(t *testing.T) {
			r := &identityRunner{FakeCommandRunner: command.NewFakeCommandRunner(), existing: map[string]bool{
				workerClientIdentityPath: true, "/var/lib/minikube/certs/ca.crt": true,
			}}
			switch scenario {
			case "new":
				r.existing[workerClientIdentityPath] = false
			case "existing":
				r.existing[workerKubeconfigPath] = true
			case "missing-ca":
				r.existing["/var/lib/minikube/certs/ca.crt"] = false
			case "denied":
				r.denied = true
			}
			b := &Bootstrapper{c: r}
			cfg := config.ClusterConfig{APIServerPort: 8443}
			err := b.restoreManagedKubeletConfig(cfg)
			if (err != nil) != (scenario == "denied" || scenario == "missing-ca") {
				t.Fatalf("unexpected outcome: %v", err)
			}
			if scenario != "restore" {
				if r.copied != nil {
					t.Fatal("unexpected configuration replacement")
				}
				return
			}
			if r.copied == nil || r.copied.GetTargetPath() != workerKubeconfigPath || r.copied.GetPermissions() != "0600" {
				t.Fatal("expected a private kubelet configuration write")
			}
			payload, err := io.ReadAll(r.copied)
			if err != nil {
				t.Fatal(err)
			}
			restored, err := clientcmd.Load(payload)
			if err != nil {
				t.Fatal(err)
			}
			context := restored.Contexts[restored.CurrentContext]
			identity := restored.AuthInfos[context.AuthInfo]
			cluster := restored.Clusters[context.Cluster]
			if identity.ClientCertificate != workerClientIdentityPath || identity.ClientKey != workerClientIdentityPath || len(identity.ClientKeyData) != 0 {
				t.Fatal("original rotating client identity was replaced or copied")
			}
			if cluster.Server != "https://control-plane.minikube.internal:8443" || cluster.InsecureSkipTLSVerify || cluster.CertificateAuthority != "/var/lib/minikube/certs/ca.crt" {
				t.Fatal("original cluster verification was weakened")
			}
		})
	}
}
