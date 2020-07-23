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
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/state"
	"k8s.io/minikube/pkg/util/retry"
)

// WaitForAPIServerProcess waits for api server to be healthy returns error if it doesn't
func WaitForAPIServerProcess(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner, start time.Time, timeout time.Duration) error {
	glog.Infof("waiting for apiserver process to appear ...")
	err := wait.PollImmediate(time.Millisecond*500, timeout, func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during process check")
		}

		if time.Since(start) > minLogCheckTime {
			announceProblems(r, bs, cfg, cr)
			time.Sleep(kconst.APICallRetryInterval * 5)
		}

		if _, ierr := apiServerPID(cr); ierr != nil {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("apiserver process never appeared")
	}
	glog.Infof("duration metric: took %s to wait for apiserver process to appear ...", time.Since(start))
	return nil
}

// apiServerPID returns our best guess to the apiserver pid
func apiServerPID(cr command.Runner) (int, error) {
	rr, err := cr.RunCmd(exec.Command("sudo", "pgrep", "-xnf", "kube-apiserver.*minikube.*"))
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(rr.Stdout.String())
	return strconv.Atoi(s)
}

// WaitForHealthyAPIServer waits for api server status to be running
func WaitForHealthyAPIServer(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner, client *kubernetes.Clientset, start time.Time, hostname string, port int, timeout time.Duration) error {
	glog.Infof("waiting for apiserver healthz status ...")
	hStart := time.Now()

	healthz := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during healthz check")
		}

		if time.Since(start) > minLogCheckTime {
			announceProblems(r, bs, cfg, cr)
			time.Sleep(kconst.APICallRetryInterval * 5)
		}

		st, err := apiServerHealthzNow(hostname, port)
		if err != nil {
			glog.Warningf("status: %v", err)
			return false, nil
		}
		if st.Condition != state.OK {
			return false, nil
		}
		return true, nil
	}

	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, healthz); err != nil {
		return fmt.Errorf("apiserver healthz never reported healthy: %v", err)
	}

	vcheck := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during version check")
		}
		if err := APIServerVersionMatch(client, cfg.KubernetesConfig.KubernetesVersion); err != nil {
			glog.Warningf("api server version match failed: %v", err)
			return false, nil
		}
		return true, nil
	}

	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, vcheck); err != nil {
		return fmt.Errorf("controlPlane never updated to %s", cfg.KubernetesConfig.KubernetesVersion)
	}

	glog.Infof("duration metric: took %s to wait for apiserver health ...", time.Since(hStart))
	return nil
}

// APIServerVersionMatch checks if the server version matches the expected
func APIServerVersionMatch(client *kubernetes.Clientset, expected string) error {
	vi, err := client.ServerVersion()
	if err != nil {
		return errors.Wrap(err, "server version")
	}
	glog.Infof("control plane version: %s", vi)
	if version.CompareKubeAwareVersionStrings(vi.String(), expected) != 0 {
		return fmt.Errorf("controlPane = %q, expected: %q", vi.String(), expected)
	}
	return nil
}

// APIServerStatus returns apiserver state
func APIServerStatus(cr command.Runner, hostname string, port int) (state.Component, error) {
	glog.Infof("Checking apiserver status ...")
	st := state.Component{Name: "apiserver", Kind: "service"}

	pid, err := apiServerPID(cr)
	if err != nil {
		st.Errors = append(st.Errors, fmt.Sprintf("pid: %v", err))
		st.Condition = state.Unavailable
		return st, nil
	}

	// Get the freezer cgroup entry for this pid
	rr, err := cr.RunCmd(exec.Command("sudo", "egrep", "^[0-9]+:freezer:", fmt.Sprintf("/proc/%d/cgroup", pid)))
	if err != nil {
		w := fmt.Sprintf("unable to find freezer cgroup: %v", err)
		glog.Warning(w)
		st.Warnings = append(st.Warnings, w)
		hst, err := apiServerHealthz(hostname, port)
		return st.Merge(hst), err
	}

	freezer := strings.TrimSpace(rr.Stdout.String())
	glog.Infof("apiserver freezer: %q", freezer)
	fparts := strings.Split(freezer, ":")
	if len(fparts) != 3 {
		w := fmt.Sprintf("unable to parse freezer - found %d parts: %s", len(fparts), freezer)
		glog.Warning(w)
		st.Warnings = append(st.Warnings, w)
		hst, err := apiServerHealthz(hostname, port)
		return st.Merge(hst), err

	}

	rr, err = cr.RunCmd(exec.Command("sudo", "cat", path.Join("/sys/fs/cgroup/freezer", fparts[2], "freezer.state")))
	if err != nil {
		// example error from github action:
		// cat: /sys/fs/cgroup/freezer/actions_job/e62ef4349cc5a70f4b49f8a150ace391da6ad6df27073c83ecc03dbf81fde1ce/kubepods/burstable/poda1de58db0ce81d19df7999f6808def1b/5df53230fe3483fd65f341923f18a477fda92ae9cd71061168130ef164fe479c/freezer.state: No such file or directory\n"*
		// TODO: #7770 investigate how to handle this error better.
		stderr := rr.Stderr.String()
		if strings.Contains(stderr, "freezer.state: No such file or directory\n") {
			glog.Infof("unable to get freezer state (might be okay and be related to #770): %s", stderr)
		} else {
			w := fmt.Sprintf("unable to get freezer state: %v (stderr=%s)", err, stderr)
			glog.Warning(w)
			st.Warnings = append(st.Warnings, w)
		}
		hst, err := apiServerHealthz(hostname, port)
		return st.Merge(hst), err
	}

	fs := strings.TrimSpace(rr.Stdout.String())
	glog.Infof("freezer state: %q", fs)

	if fs == "FREEZING" {
		st.Condition = state.Pausing
		return st, nil
	}

	if fs == "FROZEN" {
		st.Condition = state.Paused
		return st, nil
	}

	hst, err := apiServerHealthz(hostname, port)
	return st.Merge(hst), err

}

// apiServerHealthz checks apiserver in a patient and tolerant manner
func apiServerHealthz(hostname string, port int) (state.Component, error) {
	var st state.Component
	var err error
	start := time.Now()

	check := func() error {
		// etcd gets upset sometimes and causes healthz to report a failure. Be tolerant of it.
		st, err = apiServerHealthzNow(hostname, port)
		if err != nil {
			return err
		}
		if st.Condition != state.OK {
			return fmt.Errorf("state is %q", st)
		}
		return nil
	}

	err = retry.Local(check, 5*time.Second)

	// Don't propagate 'Stopped' upwards as an error message, as clients may interpret the err
	// as an inability to get status. We need it for retry.Local, however.
	if st.Condition == state.Stopped {
		return st, nil
	}

	if st.Condition == state.Unavailable {
		st.Messages = append(st.Messages, fmt.Sprintf("healthz returned within %s", time.Since(start)))
	}

	return st, err
}

// apiServerHealthzNow hits the /healthz endpoint and returns libmachine style state.Component
func apiServerHealthzNow(hostname string, port int) (state.Component, error) {
	url := fmt.Sprintf("https://%s/healthz", net.JoinHostPort(hostname, fmt.Sprint(port)))
	glog.Infof("Checking apiserver healthz at %s ...", url)
	// To avoid: x509: certificate signed by unknown authority
	tr := &http.Transport{
		Proxy:           nil, // Avoid using a proxy to speak to a local host
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	// Connection refused, usually.
	if err != nil {
		glog.Infof("not running: %s: %v", url, err)
		return state.Component{Condition: state.Unavailable, Errors: []string{err.Error()}}, nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Warningf("unable to read response body: %s", err)
	}

	glog.Infof("%s returned %d:\n%s", url, resp.StatusCode, body)
	if resp.StatusCode == http.StatusUnauthorized {
		err := fmt.Errorf("%s returned code %d (unauthorized). Check your apiserver authorization settings:\n%s", url, resp.StatusCode, body)
		return state.Component{Condition: state.Unauthorized, Errors: []string{err.Error()}}, err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("%s returned code %d (unauthorized). Check your apiserver authorization settings:\n%s", url, resp.StatusCode, body)
		return state.Component{Condition: state.Unauthorized, Errors: []string{err.Error()}}, err
	}

	return state.Component{Condition: state.OK}, nil
}
