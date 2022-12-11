//go:build integration

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
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
)

// TestNetworkPlugins tests all supported CNI options
// Options tested: kubenet, bridge, flannel, kindnet, calico, cilium
// Flags tested: enable-default-cni (legacy), false (CNI off), auto-detection
func TestNetworkPlugins(t *testing.T) {
	// generate reasonably unique profile name suffix to be used for all tests
	suffix := UniqueProfileName("")

	MaybeParallel(t)
	if NoneDriver() {
		t.Skip("skipping since test for none driver")
	}

	t.Run("group", func(t *testing.T) {
		tests := []struct {
			name          string
			args          []string
			kubeletPlugin string
			podLabel      string
			namespace     string
			hairpin       bool
		}{
			// for containerd and crio runtimes kindnet CNI is used by default and hairpin is enabled
			{"auto", []string{}, "", "", "", ContainerRuntime() != "docker"},
			{"kubenet", []string{"--network-plugin=kubenet"}, "kubenet", "", "", true},
			{"bridge", []string{"--cni=bridge"}, "cni", "", "", true},
			{"enable-default-cni", []string{"--enable-default-cni=true"}, "cni", "", "", true},
			{"flannel", []string{"--cni=flannel"}, "cni", "app=flannel", "kube-flannel", true},
			{"kindnet", []string{"--cni=kindnet"}, "cni", "app=kindnet", "kube-system", true},
			{"false", []string{"--cni=false"}, "", "", "", false},
			{"custom-flannel", []string{fmt.Sprintf("--cni=%s", filepath.Join(*testdataDir, "kube-flannel.yaml"))}, "cni", "", "kube-flannel", true},
			{"calico", []string{"--cni=calico"}, "cni", "k8s-app=calico-node", "kube-system", true},
			{"cilium", []string{"--cni=cilium"}, "cni", "k8s-app=cilium", "kube-system", true},
		}

		for _, tc := range tests {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				profile := tc.name + suffix

				ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
				defer CleanupWithLogs(t, profile, cancel)

				if ContainerRuntime() != "docker" && tc.name == "false" {
					// CNI is required for current container runtime
					validateFalseCNI(ctx, t, profile)
					return
				}

				if ContainerRuntime() != "docker" && tc.name == "kubenet" {
					// CNI is disabled when --network-plugin=kubenet option is passed. See cni.New(..) function
					// But for containerd/crio CNI has to be configured
					t.Skipf("Skipping the test as %s container runtimes requires CNI", ContainerRuntime())
				}

				start := time.Now()
				MaybeParallel(t)

				startArgs := append([]string{"start", "-p", profile, "--memory=3072", "--alsologtostderr", "--wait=true", "--wait-timeout=20m"}, tc.args...)
				startArgs = append(startArgs, StartArgs()...)

				t.Run("Start", func(t *testing.T) {
					_, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
					if err != nil {
						t.Fatalf("failed start: %v", err)
					}
				})

				if !t.Failed() && tc.podLabel != "" {
					t.Run("ControllerPod", func(t *testing.T) {
						if _, err := PodWait(ctx, t, profile, tc.namespace, tc.podLabel, Minutes(10)); err != nil {
							t.Fatalf("failed waiting for %s labeled pod: %v", tc.podLabel, err)
						}
					})
				}
				if !t.Failed() {
					t.Run("KubeletFlags", func(t *testing.T) {
						var rr *RunResult
						var err error
						if NoneDriver() {
							// none does not support 'minikube ssh'
							rr, err = Run(t, exec.CommandContext(ctx, "pgrep", "-a", "kubelet"))
						} else {
							rr, err = Run(t, exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "pgrep -a kubelet"))
						}
						if err != nil {
							t.Fatalf("ssh failed: %v", err)
						}
						out := rr.Stdout.String()
						c, err := config.Load(profile)
						if err != nil {
							t.Errorf("failed to load cluster config: %v", err)
						}
						verifyKubeletFlagsOutput(t, c.KubernetesConfig.KubernetesVersion, tc.kubeletPlugin, out)
					})
				}

				if !t.Failed() {
					t.Run("NetCatPod", func(t *testing.T) {
						_, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "netcat-deployment.yaml")))
						if err != nil {
							t.Errorf("failed to apply netcat manifest: %v", err)
						}

						client, err := kapi.Client(profile)
						if err != nil {
							t.Fatalf("failed to get Kubernetes client for %s: %v", profile, err)
						}

						if err := kapi.WaitForDeploymentToStabilize(client, "default", "netcat", Minutes(15)); err != nil {
							t.Errorf("failed waiting for netcat deployment to stabilize: %v", err)
						}

						if _, err := PodWait(ctx, t, profile, "default", "app=netcat", Minutes(15)); err != nil {
							t.Fatalf("failed waiting for netcat pod: %v", err)
						}

					})
				}

				if strings.Contains(tc.name, "weave") {
					t.Skipf("skipping remaining tests for weave, as results can be unpredictable")
				}

				if !t.Failed() {
					t.Run("DNS", func(t *testing.T) {
						var rr *RunResult
						var err error

						nslookup := func() error {
							rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nslookup", "kubernetes.default"))
							return err
						}

						// If the coredns process was stable, this retry wouldn't be necessary.
						if err := retry.Expo(nslookup, 1*time.Second, Minutes(6)); err != nil {
							t.Errorf("failed to do nslookup on kubernetes.default: %v", err)
						}

						want := []byte("10.96.0.1")
						if !bytes.Contains(rr.Stdout.Bytes(), want) {
							start := time.Now()
							t.Logf(">>> debugDNS logs:\n%s\n<<< debugDNS took %v", debugDNS(t, profile), time.Since(start))
							t.Errorf("failed nslookup: got=%q, want=*%q*", rr.Stdout.Bytes(), want)
						}
					})
				}

				if !t.Failed() {
					t.Run("Localhost", func(t *testing.T) {
						tryLocal := func() error {
							_, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", "deployment/netcat", "--", "/bin/sh", "-c", "nc -w 5 -i 5 -z localhost 8080"))
							return err
						}

						if err := retry.Expo(tryLocal, 1*time.Second, Seconds(60)); err != nil {
							t.Errorf("failed to connect via localhost: %v", err)
						}
					})
				}

				if !t.Failed() {
					t.Run("HairPin", func(t *testing.T) {
						validateHairpinMode(ctx, t, profile, tc.hairpin)
					})
				}

				t.Logf("%q test finished in %s, failed=%v", tc.name, time.Since(start), t.Failed())
			})
		}
	})
}

// validateFalseCNI checks that minikube returns and error
// if container runtime is "containerd" or "crio"
// and --cni=false
func validateFalseCNI(ctx context.Context, t *testing.T, profile string) {
	cr := ContainerRuntime()

	startArgs := []string{"start", "-p", profile, "--memory=2048", "--alsologtostderr", "--cni=false"}
	startArgs = append(startArgs, StartArgs()...)

	mkCmd := exec.CommandContext(ctx, Target(), startArgs...)
	rr, err := Run(t, mkCmd)
	if err == nil {
		t.Errorf("%s expected to fail", mkCmd)
	}
	if rr.ExitCode != reason.Usage.ExitCode {
		t.Errorf("Expected %d exit code, got %d", reason.Usage.ExitCode, rr.ExitCode)
	}
	expectedMsg := fmt.Sprintf("The %q container runtime requires CNI", cr)
	if !strings.Contains(rr.Output(), expectedMsg) {
		t.Errorf("Expected %q line not found in output %s", expectedMsg, rr.Output())
	}
}

// validateHairpinMode makes sure the hairpinning (https://en.wikipedia.org/wiki/Hairpinning) is correctly configured for given CNI
// try to access deployment/netcat pod using external, obtained from 'netcat' service dns resolution, IP address
// should fail if hairpinMode is off
func validateHairpinMode(ctx context.Context, t *testing.T, profile string, hairpin bool) {
	tryHairPin := func() error {
		_, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", "deployment/netcat", "--", "/bin/sh", "-c", "nc -w 5 -i 5 -z netcat 8080"))
		return err
	}
	if hairpin {
		if err := retry.Expo(tryHairPin, 1*time.Second, Seconds(60)); err != nil {
			t.Errorf("failed to connect via pod host: %v", err)
		}
	} else {
		if tryHairPin() == nil {
			t.Fatalf("hairpin connection unexpectedly succeeded - misconfigured test?")
		}
	}
}

func verifyKubeletFlagsOutput(t *testing.T, k8sVersion, kubeletPlugin, out string) {
	version, err := util.ParseKubernetesVersion(k8sVersion)
	if err != nil {
		t.Errorf("failed to parse kubernetes version %s: %v", k8sVersion, err)
	}
	if version.GTE(semver.MustParse("1.24.0-alpha.2")) {
		return
	}
	if kubeletPlugin == "" {
		if strings.Contains(out, "--network-plugin") && ContainerRuntime() == "docker" {
			t.Errorf("expected no network plug-in, got %s", out)
		}
		if !strings.Contains(out, "--network-plugin=cni") && ContainerRuntime() != "docker" {
			t.Errorf("expected cni network plugin with conatinerd/crio, got %s", out)
		}
	} else if !strings.Contains(out, fmt.Sprintf("--network-plugin=%s", kubeletPlugin)) {
		t.Errorf("expected --network-plugin=%s, got %s", kubeletPlugin, out)
	}
}

func debugDNS(t *testing.T, profile string) string {
	var output strings.Builder

	cmd := exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nslookup", "-type=a", "kubernetes.default")
	out, _ := cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: nslookup type A kubernetes.default:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nslookup", "-debug", "kubernetes.default")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: nslookup kubernetes.default:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nslookup", "-debug", "kubernetes.default.svc.cluster.local")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: nslookup kubernetes.default.svc.cluster.local:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "dig", "kubernetes.default")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: dig kubernetes.default:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "dig", "kubernetes.default.svc.cluster.local")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: dig kubernetes.default.svc.cluster.local:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "dig", "@10.96.0.10", "kubernetes.default")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: dig @10.96.0.10 kubernetes.default:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "dig", "@10.96.0.10", "kubernetes.default.svc.cluster.local")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: dig @10.96.0.10 kubernetes.default.svc.cluster.local:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "ping", "-c", "1", "-w", "1", "kubernetes.default")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: ping kubernetes.default:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "ping", "-c", "1", "-w", "1", "kubernetes.default.svc.cluster.local")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: ping kubernetes.default.svc.cluster.local:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nc", "-z", "-w", "5", "-v", "10.96.0.10", "53")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: nc 10.96.0.10 tcp/53:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nc", "-z", "-w", "5", "-u", "-v", "10.96.0.10", "53")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: nc 10.96.0.10 udp/53:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "cat", "/etc/nsswitch.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: /etc/nsswitch.conf:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "cat", "/etc/hosts")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: /etc/hosts:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "cat", "/etc/resolv.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: /etc/resolv.conf:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/nsswitch.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/nsswitch.conf:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/hosts")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/hosts:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/resolv.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/resolv.conf:\n%s\n", out))

	cmd = exec.Command("kubectl", "config", "view")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: kubectl config:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "get", "cm", "-A", "-oyaml")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: cms:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "get", "svc", "-A", "-owide")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: svcs:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "get", "pods", "-A", "-owide")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: pods:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/systemd/system/kubelet.service.d/10-kubeadm.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/systemd/system/kubelet.service.d/10-kubeadm.conf:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /var/lib/kubelet/config.yaml")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /var/lib/kubelet/config.yaml:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo", "find /etc/cni -type f -exec sh -c 'echo {}; cat {}' \\;")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/cni:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl status docker | cat")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: docker svc:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl status containerd | cat")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: containerd svc:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl status cri-docker | cat")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: cri-docker svc:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cri-dockerd --version")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: cri-dockerd version:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo ip a s")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: ip a s:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo iptables-save")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: iptables-save:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo iptables -t nat -L")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: iptables table nat:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "-n", "kube-system", "logs", "--selector=k8s-app=kube-proxy", "--tail=-1")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: kube-proxy logs:\n%s\n", out))

	return output.String()
}
