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

package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestWindowsNode validates a hybrid Linux+Windows minikube cluster using the Hyper-V driver.
// It starts a 2-node cluster (one Linux control-plane, one Windows worker), verifies both
// nodes are correctly labelled, deploys workloads targeting each OS, and then cleans up.
//
// Prerequisites:
//   - Must run on a Windows host with Hyper-V enabled (physical or Azure nested-virt VM)
//   - The minikube binary under test must be built from the feature/windows-node-support branch
//
// Optional env vars:
//
//	WINDOWS_VHD_URL - path or URL to a custom Windows Server VHDX image.
//	                  If unset, minikube uses the default bundled VHD.
func TestWindowsNode(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("TestWindowsNode requires a Windows host with Hyper-V")
	}

	profile := UniqueProfileName("windows-node")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(90))
	defer CleanupWithLogs(t, profile, cancel)

	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator func(context.Context, *testing.T, string)
		}{
			{"Start", validateWindowsNodeStart},
			{"NodeCount", validateWindowsNodeCount},
			{"NodeLabels", validateWindowsNodeLabels},
			{"NodeOSVersion", validateWindowsNodeOSVersion},
			{"KubeletService", validateWindowsKubeletService},
			{"KubeletArgs", validateWindowsKubeletArgs},
			{"NodeCertificates", validateWindowsNodeCertificates},
			{"ContainerdConfig", validateWindowsContainerdConfig},
			{"WindowsWorkload", validateWindowsWorkload},
			{"LinuxWorkload", validateLinuxWorkload},
			{"WorkloadsOnCorrectOS", validateWorkloadsOnCorrectOS},
			{"DNSFromWindowsPod", validateWindowsPodDNS},
			{"CrossNodeConnectivity", validateCrossNodePodConnectivity},
			{"WebServerConnectivity", validateWebServerConnectivity},
			{"NodeDiagnostics", validateWindowsNodeDiagnostics},
		}
		startPassed := false
		for _, tc := range tests {
			tc := tc
			if tc.name != "Start" && !startPassed {
				t.Run(tc.name, func(t *testing.T) {
					t.Skip("skipping because Start failed")
				})
				continue
			}

			passed := t.Run(tc.name, func(t *testing.T) {
				defer PostMortemLogs(t, profile)
				tc.validator(ctx, t, profile)
			})

			if tc.name == "Start" {
				startPassed = passed
			}
		}
	})
}

// windowsNodeName returns the name of the Windows worker node in the cluster.
func windowsNodeName(ctx context.Context, t *testing.T, profile string) string {
	t.Helper()
	rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
		"--context", profile,
		"get", "nodes",
		"-l", "kubernetes.io/os=windows",
		"-o", "jsonpath={.items[0].metadata.name}"))
	if err != nil {
		t.Fatalf("failed to get Windows node name: %v\n%s", err, rr.Stderr.String())
	}
	name := strings.TrimSpace(rr.Stdout.String())
	if name == "" {
		t.Fatal("no Windows node found in cluster")
	}
	return name
}

// podName returns the name of the first pod matching the given label selector.
func podName(ctx context.Context, t *testing.T, profile, selector string) string {
	t.Helper()
	rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
		"--context", profile,
		"get", "pods",
		"-l", selector,
		"-o", "jsonpath={.items[0].metadata.name}"))
	if err != nil {
		t.Fatalf("failed to get pod name for %s: %v\n%s", selector, err, rr.Stderr.String())
	}
	name := strings.TrimSpace(rr.Stdout.String())
	if name == "" {
		t.Fatalf("no pod found matching selector %s", selector)
	}
	return name
}

// winSSH runs a PowerShell command on the Windows node via minikube ssh.
func winSSH(ctx context.Context, t *testing.T, profile, node, command string) (string, error) {
	t.Helper()
	rr, err := Run(t, exec.CommandContext(ctx, Target(),
		"-p", profile, "ssh", "-n", node, "--", command))
	stdout := strings.TrimSpace(rr.Stdout.String())
	if err != nil {
		return stdout, fmt.Errorf("windows ssh command failed: %w\ncommand: %s\nstdout:\n%s\nstderr:\n%s", err, command, rr.Stdout.String(), rr.Stderr.String())
	}
	return stdout, nil
}

// validateWindowsNodeStart starts a 2-node hybrid cluster with one Linux and one Windows node.
func validateWindowsNodeStart(ctx context.Context, t *testing.T, profile string) {
	t.Helper()

	args := []string{
		"start", "-p", profile,
		"--nodes=2",
		"--node-os=[linux,windows]",
		"--kubernetes-version=v1.35.0",
		"--driver=hyperv",
		"--wait=true",
		"--wait-timeout=25m",
		"-v=5",
		"--alsologtostderr",
	}

	if vhdURL := os.Getenv("WINDOWS_VHD_URL"); vhdURL != "" {
		args = append(args, "--windows-vhd-url="+vhdURL)
	}
	if vsw := os.Getenv("HYPERV_VIRTUAL_SWITCH"); vsw != "" {
		args = append(args, "--hyperv-virtual-switch="+vsw)
	}

	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start hybrid cluster: %v\n%s", err, rr.Stderr.String())
	}
}

// validateWindowsNodeCount asserts exactly 2 nodes are present in the cluster.
func validateWindowsNodeCount(ctx context.Context, t *testing.T, profile string) {
	t.Helper()

	rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
		"--context", profile, "get", "nodes", "--no-headers"))
	if err != nil {
		t.Fatalf("kubectl get nodes failed: %v\n%s", err, rr.Stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(rr.Stdout.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 nodes, got %d:\n%s", len(lines), rr.Stdout.String())
	}
}

// validateWindowsNodeLabels checks that the cluster contains one Linux node and one Windows node,
// identified by the kubernetes.io/os label applied automatically by the kubelet.
func validateWindowsNodeLabels(ctx context.Context, t *testing.T, profile string) {
	t.Helper()

	for _, os := range []string{"linux", "windows"} {
		rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
			"--context", profile,
			"get", "nodes",
			"-l", "kubernetes.io/os="+os,
			"--no-headers"))
		if err != nil {
			t.Fatalf("kubectl get nodes -l kubernetes.io/os=%s failed: %v\n%s", os, err, rr.Stderr.String())
		}
		if strings.TrimSpace(rr.Stdout.String()) == "" {
			t.Errorf("expected at least one %s node but found none", os)
		}
	}
}

// validateWindowsNodeOSVersion reads the Windows version from the registry and confirms
// the node is running Windows Server 2025.
func validateWindowsNodeOSVersion(ctx context.Context, t *testing.T, profile string) {
	t.Helper()
	node := windowsNodeName(ctx, t, profile)

	out, err := winSSH(ctx, t, profile, node,
		`powershell -Command "(Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion').ProductName"`)
	if err != nil {
		t.Fatalf("failed to read OS version from Windows node: %v", err)
	}
	t.Logf("Windows node OS: %s", out)
	if !strings.Contains(out, "Windows Server 2025") {
		t.Errorf("expected Windows Server 2025, got: %q", out)
	}
}

// validateWindowsKubeletService verifies that kubelet and containerd Windows services are
// Running and configured to start automatically.
func validateWindowsKubeletService(ctx context.Context, t *testing.T, profile string) {
	t.Helper()
	node := windowsNodeName(ctx, t, profile)

	for _, svc := range []string{"kubelet", "containerd"} {
		status, err := winSSH(ctx, t, profile, node,
			fmt.Sprintf(`powershell -Command "(Get-Service %s).Status"`, svc))
		if err != nil {
			t.Errorf("failed to query service %s status: %v", svc, err)
			continue
		}
		if !strings.EqualFold(status, "Running") {
			t.Errorf("service %s: expected Running, got %q", svc, status)
		}

		startMode, err := winSSH(ctx, t, profile, node,
			fmt.Sprintf(`powershell -Command "(Get-CimInstance Win32_Service -Filter \"Name='%s'\").StartMode"`, svc))
		if err != nil {
			t.Errorf("failed to query service %s start mode: %v", svc, err)
			continue
		}
		if !strings.EqualFold(startMode, "Auto") {
			t.Errorf("service %s: expected StartMode Auto, got %q", svc, startMode)
		}
	}
}

// validateWindowsKubeletArgs reads the kubelet process command line via WMI and asserts
// that required flags are present.
func validateWindowsKubeletArgs(ctx context.Context, t *testing.T, profile string) {
	t.Helper()
	node := windowsNodeName(ctx, t, profile)

	cmdLine, err := winSSH(ctx, t, profile, node,
		`powershell -Command "(Get-CimInstance Win32_Process -Filter \"Name='kubelet.exe'\").CommandLine"`)
	if err != nil {
		t.Fatalf("failed to read kubelet command line: %v", err)
	}
	t.Logf("kubelet command line: %s", cmdLine)

	for _, flag := range []string{"--container-runtime-endpoint", "--kubeconfig"} {
		if !strings.Contains(cmdLine, flag) {
			t.Errorf("kubelet command line missing flag %q", flag)
		}
	}
}

// validateWindowsNodeCertificates confirms that kubelet client certificates are present
// on the Windows node. Checks the two paths minikube may provision.
func validateWindowsNodeCertificates(ctx context.Context, t *testing.T, profile string) {
	t.Helper()
	node := windowsNodeName(ctx, t, profile)

	paths := []string{
		`C:\var\lib\kubelet\pki\kubelet-client-current.pem`,
		`C:\k\pki\kubelet-client-current.pem`,
	}

	for _, path := range paths {
		out, err := winSSH(ctx, t, profile, node,
			fmt.Sprintf(`powershell -Command "Test-Path '%s'"`, path))
		if err != nil {
			continue
		}
		if strings.EqualFold(out, "True") {
			t.Logf("kubelet client certificate found at %s", path)
			return
		}
	}
	t.Errorf("kubelet client certificate not found at any expected path: %v", paths)
}

// validateWindowsContainerdConfig reads containerd's config.toml from the Windows node and
// asserts the pause image, runhcs-wcow-process runtime handler, and endpoint are present.
func validateWindowsContainerdConfig(ctx context.Context, t *testing.T, profile string) {
	t.Helper()
	node := windowsNodeName(ctx, t, profile)

	config, err := winSSH(ctx, t, profile, node,
		`powershell -Command "Get-Content 'C:\ProgramData\containerd\config.toml'"`)
	if err != nil {
		t.Fatalf("failed to read containerd config.toml: %v", err)
	}
	t.Logf("containerd config.toml:\n%s", config)

	checks := map[string]string{
		"pause image":           "pause",
		"runhcs-wcow-process":   "runhcs-wcow-process",
		"containerd named pipe": "npipe://",
	}
	for desc, want := range checks {
		if !strings.Contains(config, want) {
			t.Errorf("containerd config.toml missing %s (expected to contain %q)", desc, want)
		}
	}
}

// validateWindowsWorkload deploys a Windows pause container with a nodeSelector targeting the
// Windows node and waits for it to reach Running state.
func validateWindowsWorkload(ctx context.Context, t *testing.T, profile string) {
	t.Helper()

	rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
		"--context", profile, "apply", "-f", "./testdata/windows-node-workload.yaml"))
	if err != nil {
		t.Fatalf("failed to apply windows workload: %v\n%s", err, rr.Stderr.String())
	}

	// Windows containers can take several minutes to pull and start on first run.
	if _, err := PodWait(ctx, t, profile, "default", "app=win-webserver", Minutes(15)); err != nil {
		t.Errorf("win-webserver pod did not reach Running state: %v", err)
	}
}

// validateLinuxWorkload deploys a busybox container with a nodeSelector targeting the Linux node
// and waits for it to reach Running state.
func validateLinuxWorkload(ctx context.Context, t *testing.T, profile string) {
	t.Helper()

	rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
		"--context", profile, "apply", "-f", "./testdata/linux-node-workload.yaml"))
	if err != nil {
		t.Fatalf("failed to apply linux workload: %v\n%s", err, rr.Stderr.String())
	}

	if _, err := PodWait(ctx, t, profile, "default", "app=linux-test", Minutes(5)); err != nil {
		t.Errorf("linux-test pod did not reach Running state: %v", err)
	}
}

// validateWorkloadsOnCorrectOS confirms that each workload pod is scheduled on a node
// whose kubernetes.io/os label matches the pod's nodeSelector.
func validateWorkloadsOnCorrectOS(ctx context.Context, t *testing.T, profile string) {
	t.Helper()

	cases := []struct {
		podSelector string
		expectedOS  string
	}{
		{"app=win-webserver", "windows"},
		{"app=linux-test", "linux"},
	}

	for _, tc := range cases {
		// Get the node name the pod is running on.
		rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
			"--context", profile,
			"get", "pods", "-l", tc.podSelector,
			"-o", "jsonpath={.items[0].spec.nodeName}"))
		if err != nil {
			t.Errorf("failed to get node for pod %s: %v\n%s", tc.podSelector, err, rr.Stderr.String())
			continue
		}
		nodeName := strings.TrimSpace(rr.Stdout.String())
		if nodeName == "" {
			t.Errorf("pod %s has no nodeName assigned", tc.podSelector)
			continue
		}

		// Get the kubernetes.io/os label from that node.
		rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(),
			"--context", profile,
			"get", "node", nodeName,
			"-o", "jsonpath={.metadata.labels.kubernetes\\.io/os}"))
		if err != nil {
			t.Errorf("failed to get OS label for node %s: %v\n%s", nodeName, err, rr.Stderr.String())
			continue
		}
		nodeOS := strings.TrimSpace(rr.Stdout.String())

		if nodeOS != tc.expectedOS {
			t.Errorf("pod %s scheduled on node %s with OS=%q, want %q", tc.podSelector, nodeName, nodeOS, tc.expectedOS)
		}
	}
}

// validateWindowsPodDNS verifies DNS resolution works from inside the Windows pod by running
// Resolve-DnsName against the cluster DNS service name.
func validateWindowsPodDNS(ctx context.Context, t *testing.T, profile string) {
	t.Helper()
	winPod := podName(ctx, t, profile, "app=win-webserver")

	deadline := time.Now().Add(2 * time.Minute)
	var lastErr error
	for time.Now().Before(deadline) {
		rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
			"--context", profile,
			"exec", winPod, "--",
			"powershell", "-Command",
			"Resolve-DnsName kubernetes.default.svc.cluster.local"))
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Second)
			continue
		}
		out := rr.Stdout.String()
		t.Logf("Resolve-DnsName output: %s", out)
		if strings.Contains(out, "Address") || strings.Contains(out, "IP4Address") {
			t.Logf("DNS resolution from Windows pod succeeded")
			return
		}
		lastErr = fmt.Errorf("unexpected output: %s", out)
		time.Sleep(5 * time.Second)
	}
	t.Errorf("DNS resolution from Windows pod failed after 2 minutes: %v", lastErr)
}

// validateCrossNodePodConnectivity verifies Linux → Windows pod connectivity by wget-ing
// the win-webserver ClusterIP from the linux-test pod.
func validateCrossNodePodConnectivity(ctx context.Context, t *testing.T, profile string) {
	t.Helper()

	// Get the ClusterIP of the win-webserver service.
	rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
		"--context", profile,
		"get", "svc", "win-webserver",
		"-o", "jsonpath={.spec.clusterIP}"))
	if err != nil {
		t.Fatalf("failed to get win-webserver ClusterIP: %v\n%s", err, rr.Stderr.String())
	}
	clusterIP := strings.TrimSpace(rr.Stdout.String())
	if clusterIP == "" {
		t.Fatal("win-webserver ClusterIP is empty")
	}

	linuxPod := podName(ctx, t, profile, "app=linux-test")
	url := fmt.Sprintf("http://%s/", clusterIP)
	t.Logf("Testing cross-node connectivity: linux-test → win-webserver at %s", url)

	deadline := time.Now().Add(2 * time.Minute)
	var lastErr error
	for time.Now().Before(deadline) {
		rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(),
			"--context", profile,
			"exec", linuxPod, "--",
			"wget", "-qO-", "--timeout=10", url))
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Second)
			continue
		}
		t.Logf("Cross-node connectivity succeeded: got response from win-webserver")
		return
	}
	t.Errorf("cross-node connectivity linux→windows failed after 2 minutes: %v", lastErr)
}

// validateWebServerConnectivity retrieves the NodePort assigned to win-webserver, gets the
// Windows node's internal IP, and verifies the web server returns HTTP 200 with the expected
// response body.
func validateWebServerConnectivity(ctx context.Context, t *testing.T, profile string) {
	t.Helper()

	// Get the NodePort assigned to the win-webserver service.
	rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(),
		"--context", profile,
		"get", "svc", "win-webserver",
		"-o", "jsonpath={.spec.ports[0].nodePort}"))
	if err != nil {
		t.Fatalf("failed to get NodePort: %v\n%s", err, rr.Stderr.String())
	}
	nodePort := strings.TrimSpace(rr.Stdout.String())
	if nodePort == "" {
		t.Fatal("NodePort is empty")
	}

	// Get the Windows node's internal IP.
	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(),
		"--context", profile,
		"get", "nodes",
		"-l", "kubernetes.io/os=windows",
		"-o", "jsonpath={.items[0].status.addresses[?(@.type==\"InternalIP\")].address}"))
	if err != nil {
		t.Fatalf("failed to get Windows node IP: %v\n%s", err, rr.Stderr.String())
	}
	nodeIP := strings.TrimSpace(rr.Stdout.String())
	if nodeIP == "" {
		t.Fatal("Windows node IP is empty")
	}

	url := fmt.Sprintf("http://%s:%s/", nodeIP, nodePort)
	t.Logf("Checking web server at %s", url)

	client := &http.Client{Timeout: 10 * time.Second}

	// Retry for up to 2 minutes — the web server may still be starting.
	deadline := time.Now().Add(2 * time.Minute)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("web server returned HTTP %d, want 200", resp.StatusCode)
			return
		}
		t.Logf("web server responded with HTTP 200")
		return
	}
	t.Errorf("web server at %s did not respond after 2 minutes: %v", url, lastErr)
}

// validateWindowsNodeDiagnostics checks that the Windows log collection script is present
// on the node and executes without error.
func validateWindowsNodeDiagnostics(ctx context.Context, t *testing.T, profile string) {
	t.Helper()
	node := windowsNodeName(ctx, t, profile)

	// Check the script is present.
	out, err := winSSH(ctx, t, profile, node,
		`powershell -Command "Test-Path 'c:\k\debug\collect-windows-logs.ps1'"`)
	if err != nil {
		t.Fatalf("failed to check diagnostics script presence: %v", err)
	}
	if !strings.EqualFold(out, "True") {
		t.Errorf("diagnostics script not found at c:\\k\\debug\\collect-windows-logs.ps1")
		return
	}

	// Execute the script and confirm it runs without error.
	_, err = winSSH(ctx, t, profile, node,
		`powershell -Command "& 'c:\k\debug\collect-windows-logs.ps1'; exit $LASTEXITCODE"`)
	if err != nil {
		t.Errorf("diagnostics script failed to execute cleanly: %v", err)
	}
}
