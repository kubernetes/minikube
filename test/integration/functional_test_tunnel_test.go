// +build integration

/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
)

var tunnelSession StartSession

var (
	hostname = ""
	domain   = "nginx-svc.default.svc.cluster.local."
)

// validateTunnelCmd makes sure the minikube tunnel command works as expected
func validateTunnelCmd(ctx context.Context, t *testing.T, profile string) {
	ctx, cancel := context.WithTimeout(ctx, Minutes(20))
	type validateFunc func(context.Context, *testing.T, string)
	defer cancel()

	// Serial tests
	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"StartTunnel", validateTunnelStart},                   // Start tunnel
			{"WaitService", validateServiceStable},                 // Wait for service is stable
			{"AccessDirect", validateAccessDirect},                 // Access test for loadbalancer IP
			{"DNSResolutionByDig", validateDNSDig},                 // DNS forwarding test by dig
			{"DNSResolutionByDscacheutil", validateDNSDscacheutil}, // DNS forwarding test by dscacheutil
			{"AccessThroughDNS", validateAccessDNS},                // Access test for absolute dns name
			{"DeleteTunnel", validateTunnelDelete},                 // Stop tunnel and delete cluster
		}
		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile)
			})
		}
	})
}

// checkRoutePassword skips tunnel test if sudo password required for route
func checkRoutePassword(t *testing.T) {
	if !KicDriver() && runtime.GOOS != "windows" {
		if err := exec.Command("sudo", "-n", "ifconfig").Run(); err != nil {
			t.Skipf("password required to execute 'route', skipping testTunnel: %v", err)
		}
	}
}

// checkDNSForward skips DNS forwarding test if runtime is not supported
func checkDNSForward(t *testing.T) {
	// Not all platforms support DNS forwarding
	if runtime.GOOS != "darwin" {
		t.Skip("DNS forwarding is supported for darwin only now, skipping test DNS forwarding")
	}
}

// getKubeDNSIP returns kube-dns ClusterIP
func getKubeDNSIP(t *testing.T, profile string) string {
	// Load ClusterConfig
	c, err := config.Load(profile)
	if err != nil {
		t.Errorf("failed to load cluster config: %v", err)
	}
	// Get ipNet
	_, ipNet, err := net.ParseCIDR(c.KubernetesConfig.ServiceCIDR)
	if err != nil {
		t.Errorf("failed to parse service CIDR: %v", err)
	}
	// Get kube-dns ClusterIP
	ip, err := util.GetDNSIP(ipNet.String())
	if err != nil {
		t.Errorf("failed to get kube-dns IP: %v", err)
	}

	return ip.String()
}

// validateTunnelStart starts `minikube tunnel`
func validateTunnelStart(ctx context.Context, t *testing.T, profile string) {
	checkRoutePassword(t)

	args := []string{"-p", profile, "tunnel", "--alsologtostderr"}
	ss, err := Start(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start a tunnel: args %q: %v", args, err)
	}
	tunnelSession = *ss
}

// validateServiceStable starts nginx pod, nginx service and waits nginx having loadbalancer ingress IP
func validateServiceStable(ctx context.Context, t *testing.T, profile string) {
	if GithubActionRunner() && runtime.GOOS == "darwin" {
		t.Skip("The test WaitService is broken on github actions in macos https://github.com/kubernetes/minikube/issues/8434")
	}
	checkRoutePassword(t)

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("failed to get Kubernetes client for %q: %v", profile, err)
	}

	// Start the "nginx" pod.
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "apply", "-f", filepath.Join(*testdataDir, "testsvc.yaml")))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}
	if _, err := PodWait(ctx, t, profile, "default", "run=nginx-svc", Minutes(4)); err != nil {
		t.Fatalf("wait: %v", err)
	}

	if err := kapi.WaitForService(client, "default", "nginx-svc", true, 1*time.Second, Minutes(2)); err != nil {
		t.Fatal(errors.Wrap(err, "Error waiting for nginx service to be up"))
	}

	t.Run("IngressIP", func(t *testing.T) {
		if HyperVDriver() {
			t.Skip("The test WaitService/IngressIP is broken on hyperv https://github.com/kubernetes/minikube/issues/8381")
		}
		// Wait until the nginx-svc has a loadbalancer ingress IP
		err = wait.PollImmediate(5*time.Second, Minutes(3), func() (bool, error) {
			rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "svc", "nginx-svc", "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}"))
			if err != nil {
				return false, err
			}
			if len(rr.Stdout.String()) > 0 {
				hostname = rr.Stdout.String()
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			t.Errorf("nginx-svc svc.status.loadBalancer.ingress never got an IP: %v", err)
			rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "svc", "nginx-svc"))
			if err != nil {
				t.Errorf("%s failed: %v", rr.Command(), err)
			}
			t.Logf("failed to kubectl get svc nginx-svc:\n%s", rr.Output())
		}
	})
}

// validateAccessDirect validates if the test service can be accessed with LoadBalancer IP from host
func validateAccessDirect(ctx context.Context, t *testing.T, profile string) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping: access direct test is broken on windows: https://github.com/kubernetes/minikube/issues/8304")
	}
	if GithubActionRunner() && runtime.GOOS == "darwin" {
		t.Skip("skipping: access direct test is broken on github actions on macos https://github.com/kubernetes/minikube/issues/8434")
	}

	checkRoutePassword(t)

	got := []byte{}
	url := fmt.Sprintf("http://%s", hostname)

	fetch := func() error {
		h := &http.Client{Timeout: time.Second * 10}
		resp, err := h.Get(url)
		if err != nil {
			return &retry.RetriableError{Err: err}
		}
		if resp.Body == nil {
			return &retry.RetriableError{Err: fmt.Errorf("no body")}
		}
		defer resp.Body.Close()
		got, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return &retry.RetriableError{Err: err}
		}
		return nil
	}

	// Check if the nginx service can be accessed
	if err := retry.Expo(fetch, 3*time.Second, Minutes(2), 13); err != nil {
		t.Errorf("failed to hit nginx at %q: %v", url, err)

		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "svc", "nginx-svc"))
		if err != nil {
			t.Errorf("%s failed: %v", rr.Command(), err)
		}
		t.Logf("failed to kubectl get svc nginx-svc:\n%s", rr.Stdout)
	}

	want := "Welcome to nginx!"
	if strings.Contains(string(got), want) {
		t.Logf("tunnel at %s is working!", url)
	} else {
		t.Errorf("expected body to contain %q, but got *%q*", want, got)
	}
}

// validateDNSDig validates if the DNS forwarding works by dig command DNS lookup
// NOTE: DNS forwarding is experimental: https://minikube.sigs.k8s.io/docs/handbook/accessing/#dns-resolution-experimental
func validateDNSDig(ctx context.Context, t *testing.T, profile string) {
	if GithubActionRunner() && runtime.GOOS == "darwin" {
		t.Skip("skipping: access direct test is broken on github actions on macos https://github.com/kubernetes/minikube/issues/8434")
	}

	checkRoutePassword(t)
	checkDNSForward(t)

	ip := getKubeDNSIP(t, profile)
	dnsIP := fmt.Sprintf("@%s", ip)

	// Check if the dig DNS lookup works toward kube-dns IP
	rr, err := Run(t, exec.CommandContext(ctx, "dig", "+time=5", "+tries=3", dnsIP, domain, "A"))
	// dig command returns its output for stdout only. So we don't check stderr output.
	if err != nil {
		t.Errorf("failed to resolve DNS name: %v", err)
	}

	want := "ANSWER: 1"
	if strings.Contains(rr.Stdout.String(), want) {
		t.Logf("DNS resolution by dig for %s is working!", domain)
	} else {
		t.Errorf("expected body to contain %q, but got *%q*", want, rr.Stdout.String())

		// debug DNS configuration
		rr, err := Run(t, exec.CommandContext(ctx, "scutil", "--dns"))
		if err != nil {
			t.Errorf("%s failed: %v", rr.Command(), err)
		}
		t.Logf("debug for DNS configuration:\n%s", rr.Stdout.String())
	}
}

// validateDNSDscacheutil validates if the DNS forwarding works by dscacheutil command DNS lookup
// NOTE: DNS forwarding is experimental: https://minikube.sigs.k8s.io/docs/handbook/accessing/#dns-resolution-experimental
func validateDNSDscacheutil(ctx context.Context, t *testing.T, profile string) {
	if GithubActionRunner() && runtime.GOOS == "darwin" {
		t.Skip("skipping: access direct test is broken on github actions on macos https://github.com/kubernetes/minikube/issues/8434")
	}

	checkRoutePassword(t)
	checkDNSForward(t)

	// Check if the dscacheutil DNS lookup works toward target domain
	rr, err := Run(t, exec.CommandContext(ctx, "dscacheutil", "-q", "host", "-a", "name", domain))
	// If dscacheutil cannot lookup dns record, it returns no output. So we don't check stderr output.
	if err != nil {
		t.Errorf("failed to resolve DNS name: %v", err)
	}

	want := hostname
	if strings.Contains(rr.Stdout.String(), want) {
		t.Logf("DNS resolution by dscacheutil for %s is working!", domain)
	} else {
		t.Errorf("expected body to contain %q, but got *%q*", want, rr.Stdout.String())
	}
}

// validateAccessDNS validates if the test service can be accessed with DNS forwarding from host
// NOTE: DNS forwarding is experimental: https://minikube.sigs.k8s.io/docs/handbook/accessing/#dns-resolution-experimental
func validateAccessDNS(ctx context.Context, t *testing.T, profile string) {
	if GithubActionRunner() && runtime.GOOS == "darwin" {
		t.Skip("skipping: access direct test is broken on github actions on macos https://github.com/kubernetes/minikube/issues/8434")
	}

	checkRoutePassword(t)
	checkDNSForward(t)

	got := []byte{}
	url := fmt.Sprintf("http://%s", domain)

	ip := getKubeDNSIP(t, profile)
	dnsIP := fmt.Sprintf("%s:53", ip)

	// Set kube-dns dial
	kubeDNSDial := func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "udp", dnsIP)
	}

	// Set kube-dns resolver
	r := net.Resolver{
		PreferGo: true,
		Dial:     kubeDNSDial,
	}
	dialer := net.Dialer{Resolver: &r}

	// Use kube-dns resolver
	transport := &http.Transport{
		Dial:        dialer.Dial,
		DialContext: dialer.DialContext,
	}

	fetch := func() error {
		h := &http.Client{Timeout: time.Second * 10, Transport: transport}
		resp, err := h.Get(url)
		if err != nil {
			return &retry.RetriableError{Err: err}
		}
		if resp.Body == nil {
			return &retry.RetriableError{Err: fmt.Errorf("no body")}
		}
		defer resp.Body.Close()
		got, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return &retry.RetriableError{Err: err}
		}
		return nil
	}

	// Access nginx-svc through DNS resolution
	if err := retry.Expo(fetch, 3*time.Second, Seconds(30), 10); err != nil {
		t.Errorf("failed to hit nginx with DNS forwarded %q: %v", url, err)
	}

	want := "Welcome to nginx!"
	if strings.Contains(string(got), want) {
		t.Logf("tunnel at %s is working!", url)
	} else {
		t.Errorf("expected body to contain %q, but got *%q*", want, got)
	}
}

// validateTunnelDelete stops `minikube tunnel`
func validateTunnelDelete(ctx context.Context, t *testing.T, profile string) {
	checkRoutePassword(t)
	// Stop tunnel
	tunnelSession.Stop(t)
}
