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
	"io"
	"net/http"
	"os/exec"
	"os/user"
	"regexp"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/style"

	"k8s.io/minikube/pkg/minikube/browser"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/proxy"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/util/retry"
)

var (
	dashboardURLMode bool
	// Matches: 127.0.0.1:8001
	// TODO(tstromberg): Get kubectl to implement a stable supported output format.
	hostPortRe = regexp.MustCompile(`127.0.0.1:\d{4,}`)
)

// dashboardCmd represents the dashboard command
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Access the Kubernetes dashboard running within the minikube cluster",
	Long:  `Access the Kubernetes dashboard running within the minikube cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		cname := ClusterFlagValue()
		co := mustload.Healthy(cname)

		for _, n := range co.Config.Nodes {
			if err := proxy.ExcludeIP(n.IP); err != nil {
				glog.Errorf("Error excluding IP from proxy: %s", err)
			}
		}

		kubectlVersion := co.Config.KubernetesConfig.KubernetesVersion
		var err error

		// Check dashboard status before enabling it
		addon := assets.Addons["dashboard"]
		enabled := addon.IsEnabled(co.Config)

		if !enabled {
			// Send status messages to stderr for folks re-using this output.
			out.ErrT(style.Enabling, "Enabling dashboard ...")
			// Enable the dashboard add-on
			err = addons.SetAndSave(cname, "dashboard", "true")
			if err != nil {
				exit.Error(reason.InternalAddonEnable, "Unable to enable dashboard", err)
			}
		}

		ns := "kubernetes-dashboard"
		svc := "kubernetes-dashboard"
		out.ErrT(style.Verifying, "Verifying dashboard health ...")
		checkSVC := func() error { return service.CheckService(cname, ns, svc) }
		// for slow machines or parallels in CI to avoid #7503
		if err = retry.Expo(checkSVC, 100*time.Microsecond, time.Minute*10); err != nil {
			exit.Message(reason.SvcCheckTimeout, "dashboard service is not running: {{.error}}", out.V{"error": err})
		}

		out.ErrT(style.Launch, "Launching proxy ...")
		p, hostPort, err := kubectlProxy(kubectlVersion, cname)
		if err != nil {
			exit.Error(reason.HostKubectlProxy, "kubectl proxy", err)
		}
		url := dashboardURL(hostPort, ns, svc)

		out.ErrT(style.Verifying, "Verifying proxy health ...")
		chkURL := func() error { return checkURL(url) }
		if err = retry.Expo(chkURL, 100*time.Microsecond, 10*time.Minute); err != nil {
			exit.Message(reason.SvcURLTimeout, "{{.url}} is not accessible: {{.error}}", out.V{"url": url, "error": err})
		}

		// check if current user is root
		user, err := user.Current()
		if err != nil {
			exit.Error(reason.HostCurrentUser, "Unable to get current user", err)
		}
		if dashboardURLMode || user.Uid == "0" {
			out.Ln(url)
		} else {
			out.T(style.Celebrate, "Opening {{.url}} in your default browser...", out.V{"url": url})
			if err = browser.OpenURL(url); err != nil {
				exit.Message(reason.HostBrowser, "failed to open browser: {{.error}}", out.V{"error": err})
			}
		}

		glog.Infof("Success! I will now quietly sit around until kubectl proxy exits!")
		if err = p.Wait(); err != nil {
			glog.Errorf("Wait: %v", err)
		}
	},
}

// kubectlProxy runs "kubectl proxy", returning host:port
func kubectlProxy(kubectlVersion string, contextName string) (*exec.Cmd, string, error) {
	// port=0 picks a random system port

	kubectlArgs := []string{"--context", contextName, "proxy", "--port=0"}

	var cmd *exec.Cmd
	if kubectl, err := exec.LookPath("kubectl"); err == nil {
		cmd = exec.Command(kubectl, kubectlArgs...)
	} else if cmd, err = KubectlCommand(kubectlVersion, kubectlArgs...); err != nil {
		return nil, "", err
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", errors.Wrap(err, "cmd stdout")
	}

	glog.Infof("Executing: %s %s", cmd.Path, cmd.Args)
	if err := cmd.Start(); err != nil {
		return nil, "", errors.Wrap(err, "proxy start")
	}

	glog.Infof("Waiting for kubectl to output host:port ...")
	reader := bufio.NewReader(stdoutPipe)

	var out []byte
	for {
		r, timedOut, err := readByteWithTimeout(reader, 5*time.Second)
		if err != nil {
			return cmd, "", fmt.Errorf("readByteWithTimeout: %v", err)
		}
		if r == byte('\n') {
			break
		}
		if timedOut {
			glog.Infof("timed out waiting for input: possibly due to an old kubectl version.")
			break
		}
		out = append(out, r)
	}
	glog.Infof("proxy stdout: %s", string(out))
	return cmd, hostPortRe.FindString(string(out)), nil
}

// readByteWithTimeout returns a byte from a reader or an indicator that a timeout has occurred.
func readByteWithTimeout(r io.ByteReader, timeout time.Duration) (byte, bool, error) {
	bc := make(chan byte)
	ec := make(chan error)
	go func() {
		b, err := r.ReadByte()
		if err != nil {
			ec <- err
		} else {
			bc <- b
		}
		close(bc)
		close(ec)
	}()
	select {
	case b := <-bc:
		return b, false, nil
	case err := <-ec:
		return byte(' '), false, err
	case <-time.After(timeout):
		return byte(' '), true, nil
	}
}

// dashboardURL generates a URL for accessing the dashboard service
func dashboardURL(proxy string, ns string, svc string) string {
	// Reference: https://github.com/kubernetes/dashboard/wiki/Accessing-Dashboard---1.7.X-and-above
	return fmt.Sprintf("http://%s/api/v1/namespaces/%s/services/http:%s:/proxy/", proxy, ns, svc)
}

// checkURL checks if a URL returns 200 HTTP OK
func checkURL(url string) error {
	resp, err := http.Get(url)
	glog.Infof("%s response: %v %+v", url, err, resp)
	if err != nil {
		return errors.Wrap(err, "checkURL")
	}
	if resp.StatusCode != http.StatusOK {
		return &retry.RetriableError{
			Err: fmt.Errorf("unexpected response code: %d", resp.StatusCode),
		}
	}
	return nil
}

func init() {
	dashboardCmd.Flags().BoolVar(&dashboardURLMode, "url", false, "Display dashboard URL instead of opening a browser")
}
