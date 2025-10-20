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

// Package kverify verifies a running Kubernetes cluster is healthy
package kverify

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"k8s.io/minikube/pkg/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/util/retry"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

// WaitForAPIServerProcess waits for api server to be healthy returns error if it doesn't
func WaitForAPIServerProcess(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner, start time.Time, timeout time.Duration) error {
	klog.Infof("waiting for apiserver process to appear ...")
	err := wait.PollUntilContextTimeout(context.Background(), time.Millisecond*500, timeout, true, func(_ context.Context) (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during process check")
		}

		if time.Since(start) > minLogCheckTime {
			announceProblems(r, bs, cfg, cr)
			time.Sleep(kconst.APICallRetryInterval * 5)
		}

		if _, ierr := APIServerPID(cr); ierr != nil {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("apiserver process never appeared")
	}
	klog.Infof("duration metric: took %s to wait for apiserver process to appear ...", time.Since(start))
	return nil
}

// APIServerPID returns our best guess to the apiserver pid
func APIServerPID(cr command.Runner) (int, error) {
	rr, err := cr.RunCmd(exec.Command("sudo", "pgrep", "-xnf", "kube-apiserver.*minikube.*"))
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(rr.Stdout.String())
	return strconv.Atoi(s)
}

// WaitForHealthyAPIServer waits for api server status to be running
func WaitForHealthyAPIServer(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner, client *kubernetes.Clientset, start time.Time, hostname string, port int, timeout time.Duration) error {
	klog.Infof("waiting for apiserver healthz status ...")
	hStart := time.Now()

	healthz := func(_ context.Context) (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during healthz check")
		}

		if time.Since(start) > minLogCheckTime {
			announceProblems(r, bs, cfg, cr)
			time.Sleep(kconst.APICallRetryInterval * 5)
		}

		status, err := apiServerHealthzNow(hostname, port)
		if err != nil {
			klog.Warningf("status: %v", err)
			return false, nil
		}
		if status != state.Running {
			return false, nil
		}
		return true, nil
	}

	if err := wait.PollUntilContextTimeout(context.Background(), kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, true, healthz); err != nil {
		return fmt.Errorf("apiserver healthz never reported healthy: %v", err)
	}

	vcheck := func(_ context.Context) (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during version check")
		}
		if err := APIServerVersionMatch(client, cfg.KubernetesConfig.KubernetesVersion); err != nil {
			klog.Warningf("api server version match failed: %v", err)
			return false, nil
		}
		return true, nil
	}

	if err := wait.PollUntilContextTimeout(context.Background(), kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, true, vcheck); err != nil {
		return fmt.Errorf("controlPlane never updated to %s", cfg.KubernetesConfig.KubernetesVersion)
	}

	klog.Infof("duration metric: took %s to wait for apiserver health ...", time.Since(hStart))
	return nil
}

// APIServerVersionMatch checks if the server version matches the expected
func APIServerVersionMatch(client *kubernetes.Clientset, expected string) error {
	vi, err := client.ServerVersion()
	if err != nil {
		return errors.Wrap(err, "server version")
	}
	klog.Infof("control plane version: %s", vi)
	if version.CompareKubeAwareVersionStrings(vi.String(), expected) != 0 {
		return fmt.Errorf("controlPane = %q, expected: %q", vi.String(), expected)
	}
	return nil
}

// WaitForAPIServerStatus waits for 'to' duration to get apiserver pod running or stopped
// this functions is intended to use in situations where apiserver process can be recreated
// by container runtime restart for example and there is a gap before it comes back
func WaitForAPIServerStatus(cr command.Runner, to time.Duration, hostname string, port int) (state.State, error) {
	var st state.State
	err := wait.PollUntilContextTimeout(context.Background(), 500*time.Millisecond, to, true, func(_ context.Context) (bool, error) {
		var err error
		st, err = APIServerStatus(cr, hostname, port)
		if st == state.Stopped {
			return false, nil
		}
		return true, err
	})
	return st, err
}

// APIServerStatus returns apiserver status in libmachine style state.State
func APIServerStatus(cr command.Runner, hostname string, port int) (state.State, error) {
	klog.Infof("Checking apiserver status ...")

	pid, err := APIServerPID(cr)
	if err != nil {
		klog.Warningf("stopped: unable to get apiserver pid: %v", err)
		return state.Stopped, nil
	}

	// Get the freezer cgroup entry for this pid
	rr, err := cr.RunCmd(exec.Command("sudo", "egrep", "^[0-9]+:freezer:", fmt.Sprintf("/proc/%d/cgroup", pid)))
	if err != nil {
		klog.Warningf("unable to find freezer cgroup: %v", err)
		return nonFreezerServerStatus(cr, hostname, port)

	}
	freezer := strings.TrimSpace(rr.Stdout.String())
	klog.Infof("apiserver freezer: %q", freezer)
	fparts := strings.Split(freezer, ":")
	if len(fparts) != 3 {
		klog.Warningf("unable to parse freezer - found %d parts: %s", len(fparts), freezer)
		return nonFreezerServerStatus(cr, hostname, port)
	}

	rr, err = cr.RunCmd(exec.Command("sudo", "cat", path.Join("/sys/fs/cgroup/freezer", fparts[2], "freezer.state")))
	if err != nil {
		// example error from github action:
		// cat: /sys/fs/cgroup/freezer/actions_job/e62ef4349cc5a70f4b49f8a150ace391da6ad6df27073c83ecc03dbf81fde1ce/kubepods/burstable/poda1de58db0ce81d19df7999f6808def1b/5df53230fe3483fd65f341923f18a477fda92ae9cd71061168130ef164fe479c/freezer.state: No such file or directory\n"*
		// TODO: #7770 investigate how to handle this error better.
		if strings.Contains(rr.Stderr.String(), "freezer.state: No such file or directory\n") {
			klog.Infof("unable to get freezer state (might be okay and be related to #770): %s", rr.Stderr.String())
		} else {
			klog.Warningf("unable to get freezer state: %s", rr.Stderr.String())
		}

		return nonFreezerServerStatus(cr, hostname, port)
	}

	fs := strings.TrimSpace(rr.Stdout.String())
	klog.Infof("freezer state: %q", fs)
	if fs == "FREEZING" || fs == "FROZEN" {
		return state.Paused, nil
	}
	return apiServerHealthz(hostname, port)
}

// nonFreezerServerStatus is the alternative flow if the guest does not have the freezer cgroup so different methods to detect the apiserver status are used
func nonFreezerServerStatus(cr command.Runner, hostname string, port int) (state.State, error) {
	rr, err := cr.RunCmd(exec.Command("ls"))
	if err != nil {
		return state.None, err
	}
	if strings.Contains(rr.Stdout.String(), "paused") {
		return state.Paused, nil
	}
	return apiServerHealthz(hostname, port)
}

// apiServerHealthz checks apiserver in a patient and tolerant manner
func apiServerHealthz(hostname string, port int) (state.State, error) {
	var st state.State
	var err error

	check := func() error {
		// etcd gets upset sometimes and causes healthz to report a failure. Be tolerant of it.
		st, err = apiServerHealthzNow(hostname, port)
		if err != nil {
			return err
		}
		if st != state.Running {
			return fmt.Errorf("state is %q", st)
		}
		return nil
	}

	err = retry.Local(check, 15*time.Second)

	// Don't propagate 'Stopped' upwards as an error message, as clients may interpret the err
	// as an inability to get status. We need it for retry.Local, however.
	if st == state.Stopped {
		return st, nil
	}
	return st, err
}

// apiServerHealthzNow hits the /healthz endpoint and returns libmachine style state.State
func apiServerHealthzNow(hostname string, port int) (state.State, error) {
	url := fmt.Sprintf("https://%s/healthz", net.JoinHostPort(hostname, fmt.Sprint(port)))
	klog.Infof("Checking apiserver healthz at %s ...", url)
	cert, err := os.ReadFile(localpath.CACert())
	if err != nil {
		klog.Infof("ca certificate: %v", err)
		return state.Stopped, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(cert)
	tr := &http.Transport{
		Proxy:           nil, // Avoid using a proxy to speak to a local host
		TLSClientConfig: &tls.Config{RootCAs: pool},
	}
	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	// Connection refused, usually.
	if err != nil {
		klog.Infof("stopped: %s: %v", url, err)
		return state.Stopped, nil
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Warningf("unable to read response body: %s", err)
	}

	klog.Infof("%s returned %d:\n%s", url, resp.StatusCode, body)
	if resp.StatusCode == http.StatusUnauthorized {
		return state.Error, fmt.Errorf("%s returned code %d (unauthorized). Check your apiserver authorization settings:\n%s", url, resp.StatusCode, body)
	}
	if resp.StatusCode != http.StatusOK {
		return state.Error, fmt.Errorf("%s returned error %d:\n%s", url, resp.StatusCode, body)
	}
	return state.Running, nil
}
