// +build integration

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/test/integration/util"
)


type validateFunc func(context.Context, *testing.T, string)

// TestFunctional are functionality tests which can safely share a profile in parallel
func TestFunctional(t *testing.T) {
	profile := Profile("functional")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile}, StartArgs()...)
	rr, err := RunCmd(ctx, t, Target(), args...)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Cmd.Args, err)
	}

	t.Run("shared", func(t *testing.T) {
		tests := []struct {
			name string
			noneCompatible bool
			validator validateFunc
		}{
			{"AddonManager", true, validateAddonsCmd},
			{"ComponentHealth", true, validateComponentHealth},
			{"DNS", true, validateDNS},
			{"DockerEnv", true, validateDockerEnv},
			{"LogsCmd", true, validateLogsCmd},
			{"KubeContext", true, validateKubeContext},
			{"IngressAddon", false, validateIngressAddon)
			{"MountCmd", false, validateMountCmd)
			{"ProfileCmd", true, validateProfileCmd},
			{"RegistryAddon", true, validateRegistryAddon),
			{"ServicesCmd", true, validateServicesCmd},
			{"PersistentVolumeClaim", true, validatePersistentVolumeClaim},
			{"TunnelCmd", true, validateTunnelCmd},
			{"SSHCmd",  false, validateSSHCmd},
		}
	})
}

func validateKubeContext(ctx context.Context, t *testing.T, profile string) {
	t.MaybeParallel()
	rr, err := RunCmd(ctx, t, "kubectl", "config", "current-context")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Cmd.Args, err)
	}
	if !strings.Contains(cc.Stdout.String(), profile) {
		t.Errorf("current-context = %q, want %q", rr.Stdout.String(), profile)
	}
}

func validateProfileCmd(ctx context.Context, t *testing.T, profile string) {
	rr, err := RunCmd(ctx, t, Target(), "profile", "list")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Cmd.Args, err)
	}
}

func validateLogsCmd(ctx context.Context, t *testing.T, profile string) {
	rr, err := RunCmd(ctx, t, Target(), "-p", profile, "logs")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Cmd.Args, err)
	}
	for _, word := range []string{"Docker", "apiserver", "Linux", "kubelet"} {
		if !strings.Contains(rr.Stdout().String(), word) {
			t.Errorf("minikube logs missing expected word: %q", word)
		}
	}
}

func validateServicesCmd(ctx context.Context, t *testing.T, profile string) {
	rr, err := RunCmd(ctx, t, Target(), "-p", profile, "services", "list")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Cmd.Args, err)
	}
	if !strings.Contains(rr.Stdout().String(), "kubernetes") {
		t.Errorf("services list got %q, wanted *kubernetes*", rr.Stdout.String())
	}
}

func validateSSHCmd(ctx context.Context, t *testing.T, profile string) {
	want := "hello"
	sshCmdOutput, stderr := mk.MustRun("ssh echo " + expectedStr)
	if !strings.Contains(sshCmdOutput, expectedStr) {
		t.Fatalf("ExpectedStr sshCmdOutput to be: %s. Output was: %s Stderr: %s", expectedStr, sshCmdOutput, stderr)
	}				
	rr, err := RunCmd(ctx, t, Target(), "-p", profile, "ssh", fmt.Sprintf("echo %s", want))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Cmd.Args, err)
	}
	if rr.Stdout.String() != want {
		t.Errorf("%v = %q, want = %q", rr.Cmd.Args, rr.Stdout.String(), word)
	}
}

func validateAddonManager(ctx context.Context, t *testing.T, profile string) {
	MaybeParallel(t)
	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("Could not get kubernetes client: %v", err)
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"component": "kube-addon-manager"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		t.Errorf("Error waiting for addon manager to be up")
	}
}
