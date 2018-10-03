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

	commonutil "k8s.io/minikube/pkg/util"
)

var (
	dashboardURLMode bool
	// Matches: 127.0.0.1:8001
	hostPortRe = regexp.MustCompile(`127.0.0.1:\d{4,}`)
)

// dashboardCmd represents the dashboard command
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Opens/displays the kubernetes dashboard URL for your local cluster",
	Long:  `Opens/displays the kubernetes dashboard URL for your local cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		glog.Infof("Setting up dashboard ...")
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()

		cluster.EnsureMinikubeRunningOrExit(api, 1)
		namespace := "kube-system"
		svc := "kubernetes-dashboard"

		if err = commonutil.RetryAfter(20, func() error { return service.CheckService(namespace, svc) }, 6*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "Could not find finalized endpoint being pointed to by %s: %s\n", svc, err)
			os.Exit(1)
		}

		p, hostPort, err := kubectlProxy()
		if err != nil {
			glog.Fatalf("kubectl proxy: %v", err)
		}
		url := dashboardURL(hostPort, namespace, svc)
		if dashboardURLMode {
			fmt.Fprintln(os.Stdout, url)
			return
		}
		fmt.Fprintln(os.Stdout, fmt.Sprintf("Opening %s in your default browser...", url))
		browser.OpenURL(url)
		p.Wait()
	},
}

// kubectlProxy runs "kubectl proxy", returning host:port
func kubectlProxy() (*exec.Cmd, string, error) {
	glog.Infof("Searching for kubectl ...")
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
	glog.Infof("proxy should be running ...")
	reader := bufio.NewReader(stdoutPipe)
	glog.Infof("Reading stdout pipe ...")
	out, err := reader.ReadString('\n')
	if err != nil {
		return nil, "", errors.Wrap(err, "read")
	}
	return cmd, parseHostPort(out), nil
}

func dashboardURL(proxy string, ns string, svc string) string {
	// Reference: https://github.com/kubernetes/dashboard/wiki/Accessing-Dashboard---1.7.X-and-above
	return fmt.Sprintf("http://%s/api/v1/namespaces/%s/services/http:%s:/proxy/", proxy, ns, svc)
}

func parseHostPort(out string) string {
	// Starting to serve on 127.0.0.1:8001
	glog.Infof("Parsing: %s ...", out)
	return hostPortRe.FindString(out)
}

func init() {
	dashboardCmd.Flags().BoolVar(&dashboardURLMode, "url", false, "Display the kubernetes dashboard in the CLI instead of opening it in the default browser")
	RootCmd.AddCommand(dashboardCmd)
}
