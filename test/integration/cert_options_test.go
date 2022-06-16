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
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/drivers/kic/oci"
)

// TestCertOptions makes sure minikube certs respect the --apiserver-ips and --apiserver-names parameters
func TestCertOptions(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping: none driver does not support ssh or bundle docker")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("cert-options")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--memory=2048", "--apiserver-ips=127.0.0.1", "--apiserver-ips=192.168.15.15", "--apiserver-names=localhost", "--apiserver-names=www.google.com", "--apiserver-port=8555"}, StartArgs()...)

	// We can safely override --apiserver-name with
	if NeedsPortForward() {
		args = append(args, "--apiserver-name=localhost")
	}

	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	// verify that the alternate names/ips are included in the apiserver cert
	// in minikube vm, run - openssl x509 -text -noout -in /var/lib/minikube/certs/apiserver.crt
	// to inspect the apiserver cert

	// can filter further with '-certopt no_subject,no_header,no_version,no_serial,no_signame,no_validity,no_issuer,no_pubkey,no_sigdump,no_aux'
	apiserverCertCmd := "openssl x509 -text -noout -in /var/lib/minikube/certs/apiserver.crt"
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", apiserverCertCmd))
	if err != nil {
		t.Errorf("failed to read apiserver cert inside minikube. args %q: %v", rr.Command(), err)
	}

	extraNamesIps := [4]string{"127.0.0.1", "192.168.15.15", "localhost", "www.google.com"}

	for _, eni := range extraNamesIps {
		if !strings.Contains(rr.Stdout.String(), eni) {
			t.Errorf("apiserver cert does not include %s in SAN.", eni)
		}
	}

	// verify that the apiserver is serving on port 8555
	if NeedsPortForward() { // in case of docker/podman on non-linux the port will be a "random assigned port" in kubeconfig
		bin := "docker"
		if PodmanDriver() {
			bin = "podman"
		}

		port, err := oci.ForwardedPort(bin, profile, 8555)
		if err != nil {
			t.Errorf("failed to inspect container for the port %v", err)
		}
		if port == 0 {
			t.Errorf("expected to get a non-zero forwarded port but got %d", port)
		}
	} else {
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "config", "view"))
		if err != nil {
			t.Errorf("failed to get kubectl config. args %q : %v", rr.Command(), err)
		}
		if !strings.Contains(rr.Stdout.String(), "8555") {
			t.Errorf("Kubeconfig apiserver server port incorrect. Output of \n 'kubectl config view' = %q", rr.Output())
		}
	}

	// Also check the kubeconfig inside minikube using SSH
	// located at /etc/kubernetes/admin.conf
	args = []string{"ssh", "-p", profile, "--", "sudo cat /etc/kubernetes/admin.conf"}
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to SSH to minikube with args: %q : %v", rr.Command(), err)
	}

	if !strings.Contains(rr.Stdout.String(), "8555") {
		t.Errorf("Internal minikube kubeconfig (admin.conf) does not contains the right api port. %s", rr.Output())
	}

}

// TestCertExpiration makes sure minikube can start after its profile certs have expired.
// It does this by configuring minikube certs to expire after 3 minutes, then waiting 3 minutes, then starting again.
// It also makes sure minikube prints a cert expiration warning to the user.
func TestCertExpiration(t *testing.T) {
	MaybeParallel(t)

	profile := UniqueProfileName("cert-expiration")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--memory=2048", "--cert-expiration=3m"}, StartArgs()...)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	// Now wait 3 minutes for the certs to expire and make sure minikube starts properly
	time.Sleep(time.Minute * 3)
	args = append([]string{"start", "-p", profile, "--memory=2048", "--cert-expiration=8760h"}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube after cert expiration: %q : %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Output(), "expired") {
		t.Errorf("minikube start output did not warn about expired certs: %v", rr.Output())
	}
}
