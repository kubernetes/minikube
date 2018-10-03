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

package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/service"

	"k8s.io/minikube/pkg/util"
)

var (
	dashboardURLMode bool
	// Matches: 127.0.0.1:8001
	hostPortRe = regexp.MustCompile(`127.0.0.1:\d{4,}`)
)

// dashboardCmd represents the dashboard command
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Access the kubernetes dashboard running within the minikube cluster",
	Long:  `Access the kubernetes dashboard running within the minikube cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		defer func() {
			err := api.Close()
			if err != nil {
				glog.Warningf("Failed to close API: %v", err)
			}
		}()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
			os.Exit(1)
		}
		cluster.EnsureMinikubeRunningOrExit(api, 1)

		ns := "kube-system"
		svc := "kubernetes-dashboard"
		if err = util.RetryAfter(30, func() error { return service.CheckService(ns, svc) }, 1*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "%s:%s is not running: %s\n", ns, svc, err)
			os.Exit(1)
		}

		p, hostPort, err := kubectlProxy()
		if err != nil {
			glog.Fatalf("kubectl proxy: %v", err)
		}
		url := dashboardURL(hostPort, ns, svc)

		if err = util.RetryAfter(30, func() error { return checkURL(url) }, 1*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "%s is not responding properly: %s\n", url, err)
			os.Exit(1)
		}

		if dashboardURLMode {
			fmt.Fprintln(os.Stdout, url)
			return
		}
		fmt.Fprintln(os.Stdout, fmt.Sprintf("Opening %s in your default browser...", url))
		if err = browser.OpenURL(url); err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("failed to open browser: %v", err))
		}
		glog.Infof("Waiting forever for kubectl proxy to exit ...")
		if err = p.Wait(); err != nil {
			glog.Errorf("Wait: %v", err)
		}
	},
}

// kubectlProxy runs "kubectl proxy", returning host:port
func kubectlProxy() (*exec.Cmd, string, error) {
	path, err := exec.LookPath("kubectl")
	if err != nil {
		return nil, "", errors.Wrap(err, "Unable to find kubectl in PATH")
	}
	cmd := exec.Command(path, "proxy", "--port=0")
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", errors.Wrap(err, "stdout")
	}

	glog.Infof("Executing: %s %s", cmd.Path, cmd.Args)
	if err := cmd.Start(); err != nil {
		return nil, "", errors.Wrap(err, "start")
	}
	reader := bufio.NewReader(stdoutPipe)
	glog.Infof("Started proxy, now reading stdout pipe ...")
	out, err := reader.ReadString('\n')
	if err != nil {
		return nil, "", errors.Wrap(err, "read")
	}
	glog.Infof("Output: %s ...", out)
	return cmd, hostPortRe.FindString(out), nil
}

func dashboardURL(proxy string, ns string, svc string) string {
	// Reference: https://github.com/kubernetes/dashboard/wiki/Accessing-Dashboard---1.7.X-and-above
	return fmt.Sprintf("http://%s/api/v1/nss/%s/services/http:%s:/proxy/", proxy, ns, svc)
}

func checkURL(url string) error {
	resp, err := http.Get(url)
	glog.Infof("%s response: %s %+v", url, err, resp)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return &util.RetriableError{
			Err: fmt.Errorf("unexpected response code: %d", resp.StatusCode),
		}
	}
	return nil
}

func init() {
	dashboardCmd.Flags().BoolVar(&dashboardURLMode, "url", false, "Display dashboard URL instead of opening a browser")
	RootCmd.AddCommand(dashboardCmd)
}
