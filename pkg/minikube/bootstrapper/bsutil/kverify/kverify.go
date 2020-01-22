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

// Package kverify verifies a running kubernetes cluster is healthy
package kverify

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/minikube/command"
)

// APIServerProcess waits for api server to be healthy returns error if it doesn't
func APIServerProcess(runner command.Runner, start time.Time, timeout time.Duration) error {
	glog.Infof("waiting for apiserver process to appear ...")
	err := wait.PollImmediate(time.Second*1, timeout, func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during process check")
		}
		rr, ierr := runner.RunCmd(exec.Command("sudo", "pgrep", "kube-apiserver"))
		if ierr != nil {
			glog.Warningf("pgrep apiserver: %v cmd: %s", ierr, rr.Command())
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

// SystemPods verifies essential pods for running kurnetes is running
func SystemPods(client *kubernetes.Clientset, start time.Time, timeout time.Duration) error {
	glog.Info("waiting for kube-system pods to appear ...")
	pStart := time.Now()
	podStart := time.Time{}
	podList := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during pod check")
		}
		// Wait for any system pod, as waiting for apiserver may block until etcd
		pods, err := client.CoreV1().Pods("kube-system").List(meta.ListOptions{})
		if len(pods.Items) < 2 {
			podStart = time.Time{}
			return false, nil
		}
		if err != nil {
			podStart = time.Time{}
			return false, nil
		}
		if podStart.IsZero() {
			podStart = time.Now()
		}

		glog.Infof("%d kube-system pods found since %s", len(pods.Items), podStart)
		if time.Since(podStart) > 2*kconst.APICallRetryInterval {
			glog.Infof("stability requirement met, returning")
			return true, nil
		}
		return false, nil
	}
	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, podList); err != nil {
		return fmt.Errorf("apiserver never returned a pod list")
	}
	glog.Infof("duration metric: took %s to wait for pod list to return data ...", time.Since(pStart))
	return nil
}

// APIServerIsRunning waits for api server status to be running
func APIServerIsRunning(start time.Time, ip string, port int, timeout time.Duration) error {
	glog.Infof("waiting for apiserver healthz status ...")
	hStart := time.Now()
	healthz := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during healthz check")
		}

		status, err := APIServerStatus(net.ParseIP(ip), port)
		if err != nil {
			glog.Warningf("status: %v", err)
			return false, nil
		}
		if status != "Running" {
			return false, nil
		}
		return true, nil
	}

	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, healthz); err != nil {
		return fmt.Errorf("apiserver healthz never reported healthy")
	}
	glog.Infof("duration metric: took %s to wait for apiserver healthz status ...", time.Since(hStart))
	return nil
}

// APIServerStatus hits the /healthz endpoint and returns libmachine style state.State
func APIServerStatus(ip net.IP, apiserverPort int) (string, error) {
	url := fmt.Sprintf("https://%s/healthz", net.JoinHostPort(ip.String(), fmt.Sprint(apiserverPort)))
	// To avoid: x509: certificate signed by unknown authority
	tr := &http.Transport{
		Proxy:           nil, // To avoid connectiv issue if http(s)_proxy is set.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	// Connection refused, usually.
	if err != nil {
		return state.Stopped.String(), nil
	}
	if resp.StatusCode != http.StatusOK {
		glog.Warningf("%s response: %v %+v", url, err, resp)
		return state.Error.String(), nil
	}
	return state.Running.String(), nil
}
