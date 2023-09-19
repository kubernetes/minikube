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
	"k8s.io/minikube/pkg/minikube/constants"
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
			// kindnet CNI is used by default and hairpin is enabled
			{"auto", []string{}, "", "", "", usingCNI()},
			{"kubenet", []string{"--network-plugin=kubenet"}, "kubenet", "", "", true},
			{"bridge", []string{"--cni=bridge"}, "cni", "", "", true},
			{"enable-default-cni", []string{"--enable-default-cni=true"}, "cni", "", "", true},
			{"flannel", []string{"--cni=flannel"}, "cni", "app=flannel", "kube-flannel", true},
			{"kindnet", []string{"--cni=kindnet"}, "cni", "app=kindnet", "kube-system", true},
			{"false", []string{"--cni=false"}, "", "", "", true},
			{"custom-flannel", []string{fmt.Sprintf("--cni=%s", filepath.Join(*testdataDir, "kube-flannel.yaml"))}, "cni", "", "kube-flannel", true},
			{"calico", []string{"--cni=calico"}, "cni", "k8s-app=calico-node", "kube-system", true},
			{"cilium", []string{"--cni=cilium"}, "cni", "k8s-app=cilium", "kube-system", true},
		}

		for _, tc := range tests {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				profile := tc.name + suffix

				ctx, cancel := context.WithTimeout(context.Background(), Minutes(90))
				defer CleanupWithLogs(t, profile, cancel)
				// collect debug logs
				defer debugLogs(t, profile)

				if ContainerRuntime() != "docker" && tc.name == "false" {
					// CNI is required for current container runtime
					validateFalseCNI(ctx, t, profile)
					return
				}

				if ContainerRuntime() != "docker" && tc.name == "kubenet" {
					// CNI is disabled when --network-plugin=kubenet option is passed. See cni.New(..) function
					t.Skipf("Skipping the test as %s container runtimes requires CNI", ContainerRuntime())
				}

				// (current) cilium is known to mess up the system when interfering with other network tests, so we disable it for now - probably needs updating?
				// hint: most probably the problem is in combination of: containerd + (outdated) cgroup_v1(cgroupfs) + (outdated) cilium, on systemd it should work
				// unfortunately, cilium changed how cni is deployed and does not provide manifests anymore (since v1.9) so that we can "just update" ours
				// ref: https://docs.cilium.io/en/stable/gettingstarted/k8s-install-default/
				// ref: https://docs.cilium.io/en/stable/gettingstarted/k8s-install-kubeadm/
				if tc.name == "cilium" {
					t.Skip("Skipping the test as it's interfering with other tests and is outdated")
				}

				start := time.Now()
				MaybeParallel(t)

				startArgs := append([]string{"start", "-p", profile, "--memory=3072", "--alsologtostderr", "--wait=true", "--wait-timeout=15m"}, tc.args...)
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

// usingCNI checks if not using dockershim
func usingCNI() bool {
	if ContainerRuntime() != "docker" {
		return true
	}
	version, err := util.ParseKubernetesVersion(constants.DefaultKubernetesVersion)
	if err != nil {
		return false
	}
	if version.GTE(semver.MustParse("1.24.0-alpha.2")) {
		return true
	}
	return false
}

// validateFalseCNI checks that minikube returns and error
// if container runtime is "containerd" or "crio"
// and --cni=false
func validateFalseCNI(ctx context.Context, t *testing.T, profile string) {
	cr := ContainerRuntime()

	// override cri-o name
	if cr == "cri-o" {
		cr = "crio"
	}

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
			t.Errorf("hairpin connection unexpectedly succeeded - misconfigured test?")
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

// debug logs for dns and other network issues
func debugLogs(t *testing.T, profile string) {
	t.Helper()

	start := time.Now()

	var output strings.Builder
	output.WriteString(fmt.Sprintf("----------------------- debugLogs start: %s [pass: %v] --------------------------------", profile, !t.Failed()))

	// basic nslookup
	cmd := exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nslookup", "-timeout=5", "kubernetes.default")
	out, err := cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: nslookup kubernetes.default:\n%s\n", out))
	// skip some checks if no issues or lower-level connectivity issues
	if err == nil && !strings.Contains(string(out), "10.96.0.1") || err != nil && !strings.Contains(string(out), ";; connection timed out; no servers could be reached") { // for both nslookup and dig
		// nslookup trace search
		cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nslookup", "-timeout=5", "-debug", "-type=a", "kubernetes.default")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> netcat: nslookup debug kubernetes.default a-records:\n%s\n", out))

		// dig trace search udp
		cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "dig", "+timeout=5", "+search", "+showsearch", "kubernetes.default")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> netcat: dig search kubernetes.default:\n%s\n", out))
		// dig trace direct udp
		cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "dig", "+timeout=5", "@10.96.0.10", "kubernetes.default.svc.cluster.local")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> netcat: dig @10.96.0.10 kubernetes.default.svc.cluster.local udp/53:\n%s\n", out))
		// dig trace direct tcp
		cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "dig", "+timeout=5", "@10.96.0.10", "+tcp", "kubernetes.default.svc.cluster.local")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> netcat: dig @10.96.0.10 kubernetes.default.svc.cluster.local tcp/53:\n%s\n", out))
	}

	// check udp connectivity
	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nc", "-w", "5", "-z", "-n", "-v", "-u", "10.96.0.10", "53")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: nc 10.96.0.10 udp/53:\n%s\n", out))
	// check tcp connectivity
	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nc", "-w", "5", "-z", "-n", "-v", "10.96.0.10", "53")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: nc 10.96.0.10 tcp/53:\n%s\n", out))

	// pod's dns env
	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "cat", "/etc/nsswitch.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: /etc/nsswitch.conf:\n%s\n", out))
	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "cat", "/etc/hosts")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: /etc/hosts:\n%s\n", out))
	cmd = exec.Command("kubectl", "--context", profile, "exec", "deployment/netcat", "--", "cat", "/etc/resolv.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> netcat: /etc/resolv.conf:\n%s\n", out))

	// "host's" dns env
	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/nsswitch.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/nsswitch.conf:\n%s\n", out))
	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/hosts")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/hosts:\n%s\n", out))
	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/resolv.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/resolv.conf:\n%s\n", out))

	// k8s resources overview
	cmd = exec.Command("kubectl", "--context", profile, "get", "node,svc,ep,ds,deploy,pods", "-A", "-owide")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: nodes, services, endpoints, daemon sets, deployments and pods, :\n%s\n", out))

	// crictl pods overview
	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo crictl pods")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: crictl pods:\n%s\n", out))
	// crictl containers overview
	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo crictl ps --all")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: crictl containers:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "describe", "deployment", "-n", "default", "--selector=app=netcat")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: describe netcat deployment:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-n", "default", "--selector=app=netcat")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: describe netcat pod(s):\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "logs", "-n", "default", "--selector=app=netcat", "--tail=-1")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: netcat logs:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "describe", "deployment", "-n", "kube-system", "--selector=k8s-app=kube-dns")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: describe coredns deployment:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-n", "kube-system", "--selector=k8s-app=kube-dns")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: describe coredns pods:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "logs", "-n", "kube-system", "--selector=k8s-app=kube-dns", "--tail=-1")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: coredns logs:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-n", "kube-system", "--selector=component=kube-apiserver")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: describe api server pod(s):\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "logs", "-n", "kube-system", "--selector=component=kube-apiserver", "--tail=-1")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: api server logs:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo find /etc/cni -type f -exec sh -c 'echo {}; cat {}' \\;")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/cni:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo ip a s")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: ip a s:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo ip r s")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: ip r s:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo iptables-save")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: iptables-save:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo iptables -t nat -L -n -v")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: iptables table nat:\n%s\n", out))

	if strings.Contains(profile, "flannel") {
		cmd = exec.Command("kubectl", "--context", profile, "describe", "ds", "-A", "--selector=app=flannel")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe flannel daemon set:\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-A", "--selector=app=flannel")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe flannel pod(s):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-flannel", "--selector=app=flannel", "--all-containers", "--prefix", "--ignore-errors")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: flannel container(s) logs (current):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-flannel", "--selector=app=flannel", "--all-containers", "--prefix", "--ignore-errors", "--previous")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: flannel container(s) logs (previous):\n%s\n", out))

		cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /run/flannel/subnet.env")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> host: /run/flannel/subnet.env:\n%s\n", out))

		cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/kube-flannel/cni-conf.json")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> host: /etc/kube-flannel/cni-conf.json:\n%s\n", out))
	}

	if strings.Contains(profile, "calico") {
		cmd = exec.Command("kubectl", "--context", profile, "describe", "ds", "-A", "--selector=k8s-app=calico-node")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe calico daemon set:\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-A", "--selector=k8s-app=calico-node")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe calico daemon set pod(s):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=k8s-app=calico-node", "--all-containers", "--prefix", "--ignore-errors")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: calico daemon set container(s) logs (current):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=k8s-app=calico-node", "--all-containers", "--prefix", "--ignore-errors", "--previous")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: calico daemon set container(s) logs (previous):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "describe", "deploy", "-A", "--selector=k8s-app=calico-kube-controllers")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe calico deployment:\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-A", "--selector=k8s-app=calico-kube-controllers")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe calico deployment pod(s):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=k8s-app=calico-kube-controllers", "--all-containers", "--prefix", "--ignore-errors")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: calico deployment container(s) logs (current):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=k8s-app=calico-kube-controllers", "--all-containers", "--prefix", "--ignore-errors", "--previous")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: calico deployment container(s) logs (previous):\n%s\n", out))
	}

	if strings.Contains(profile, "cilium") {
		cmd = exec.Command("kubectl", "--context", profile, "describe", "ds", "-A", "--selector=k8s-app=cilium")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe cilium daemon set:\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-A", "--selector=k8s-app=cilium")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe cilium daemon set pod(s):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=k8s-app=cilium", "--all-containers", "--prefix", "--ignore-errors")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: cilium daemon set container(s) logs (current):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=k8s-app=cilium", "--all-containers", "--prefix", "--ignore-errors", "--previous")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: cilium daemon set container(s) logs (previous):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "describe", "deploy", "-A", "--selector=name=cilium-operator")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe cilium deployment:\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-A", "--selector=name=cilium-operator")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe cilium deployment pod(s):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=name=cilium-operator", "--all-containers", "--prefix", "--ignore-errors")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: cilium deployment container(s) logs (current):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=name=cilium-operator", "--all-containers", "--prefix", "--ignore-errors", "--previous")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: cilium deployment container(s) logs (previous):\n%s\n", out))
	}

	if strings.Contains(profile, "kindnet") {
		cmd = exec.Command("kubectl", "--context", profile, "describe", "ds", "-A", "--selector=app=kindnet")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe kindnet daemon set:\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-A", "--selector=app=kindnet")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: describe kindnet pod(s):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=app=kindnet", "--all-containers", "--prefix", "--ignore-errors")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: kindnet container(s) logs (current):\n%s\n", out))

		cmd = exec.Command("kubectl", "--context", profile, "logs", "--namespace=kube-system", "--selector=app=kindnet", "--all-containers", "--prefix", "--ignore-errors", "--previous")
		out, _ = cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("\n>>> k8s: kindnet container(s) logs (previous):\n%s\n", out))
	}

	cmd = exec.Command("kubectl", "--context", profile, "describe", "ds", "-n", "kube-system", "--selector=k8s-app=kube-proxy")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: describe kube-proxy daemon set:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "describe", "pods", "-n", "kube-system", "--selector=k8s-app=kube-proxy")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: describe kube-proxy pod(s):\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "logs", "-n", "kube-system", "--selector=k8s-app=kube-proxy", "--tail=-1")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: kube-proxy logs:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl status kubelet --all --full --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: kubelet daemon status:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl cat kubelet --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: kubelet daemon config:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo journalctl -xeu kubelet --all --full --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: kubelet logs:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/kubernetes/kubelet.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/kubernetes/kubelet.conf:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /var/lib/kubelet/config.yaml")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /var/lib/kubelet/config.yaml:\n%s\n", out))

	cmd = exec.Command("kubectl", "config", "view")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: kubectl config:\n%s\n", out))

	cmd = exec.Command("kubectl", "--context", profile, "get", "cm", "-A", "-oyaml")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> k8s: cms:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl status docker --all --full --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: docker daemon status:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl cat docker --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: docker daemon config:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/docker/daemon.json")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/docker/daemon.json:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo docker system info")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: docker system info:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl status cri-docker --all --full --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: cri-docker daemon status:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl cat cri-docker --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: cri-docker daemon config:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/systemd/system/cri-docker.service.d/10-cni.conf")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/systemd/system/cri-docker.service.d/10-cni.conf:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /usr/lib/systemd/system/cri-docker.service")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /usr/lib/systemd/system/cri-docker.service:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cri-dockerd --version")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: cri-dockerd version:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl status containerd --all --full --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: containerd daemon status:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl cat containerd --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: containerd daemon config:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /lib/systemd/system/containerd.service")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /lib/systemd/system/containerd.service:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo cat /etc/containerd/config.toml")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/containerd/config.toml:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo containerd config dump")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: containerd config dump:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl status crio --all --full --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: crio daemon status:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo systemctl cat crio --no-pager")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: crio daemon config:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo find /etc/crio -type f -exec sh -c 'echo {}; cat {}' \\;")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: /etc/crio:\n%s\n", out))

	cmd = exec.Command(Target(), "ssh", "-p", profile, "sudo crio config")
	out, _ = cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("\n>>> host: crio config:\n%s\n", out))

	output.WriteString(fmt.Sprintf("----------------------- debugLogs end: %s [took: %v] --------------------------------", profile, time.Since(start)))
	t.Logf("\n%s\n", output.String())
}
