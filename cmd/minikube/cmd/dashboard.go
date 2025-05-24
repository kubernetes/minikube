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
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
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
	dashboardURLMode     bool
	dashboardExposedPort int
	// Matches: "127.0.0.1:8001" or "127.0.0.1 40012" etc.
	// TODO(tstromberg): Get kubectl to implement a stable supported output format.
	hostPortRe = regexp.MustCompile(`127\.0\.0\.1(:| )\d{4,}`)
)

// dashboardCmd represents the dashboard command
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Access the Kubernetes dashboard running within the minikube cluster",
	Long:  `Access the Kubernetes dashboard running within the minikube cluster`,
	Run: func(_ *cobra.Command, _ []string) {
		cname := ClusterFlagValue()
		co := mustload.Healthy(cname)

		for _, n := range co.Config.Nodes {
			if err := proxy.ExcludeIP(n.IP); err != nil {
				klog.Errorf("Error excluding IP from proxy: %s", err)
			}
		}

		if dashboardExposedPort < 0 || dashboardExposedPort > 65535 {
			exit.Message(reason.HostKubectlProxy, "Invalid port")
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
		p, hostPort, err := kubectlProxy(kubectlVersion, co.Config.BinaryMirror, cname, dashboardExposedPort)
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
			out.Styled(style.Celebrate, "Opening {{.url}} in your default browser...", out.V{"url": url})
			if err = browser.OpenURL(url); err != nil {
				exit.Message(reason.HostBrowser, "failed to open browser: {{.error}}", out.V{"error": err})
			}
		}

		klog.Infof("Success! I will now quietly sit around until kubectl proxy exits!")
		if err = p.Wait(); err != nil {
			klog.Errorf("Wait: %v", err)
		}
	},
}

// kubectlProxy runs "kubectl proxy", returning host:port
func kubectlProxy(kubectlVersion string, binaryURL string, contextName string, port int) (*exec.Cmd, string, error) {
	// port=0 picks a random system port

	kubectlArgs := []string{"--context", contextName, "proxy", "--port", strconv.Itoa(port)}

	var cmd *exec.Cmd
	if kubectl, err := exec.LookPath("kubectl"); err == nil {
		cmd = exec.Command(kubectl, kubectlArgs...)
	} else if cmd, err = KubectlCommand(kubectlVersion, binaryURL, kubectlArgs...); err != nil {
		return nil, "", err
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", errors.Wrap(err, "cmd stdout")
	}

	klog.Infof("Executing: %s %s", cmd.Path, cmd.Args)
	if err := cmd.Start(); err != nil {
		return nil, "", errors.Wrap(err, "proxy start")
	}

	klog.Infof("Waiting for kubectl to output host:port ...")
	reader := bufio.NewReader(stdoutPipe)

	var outData []byte
	for {
		r, timedOut, err := readByteWithTimeout(reader, 5*time.Second)
		if err != nil {
			return cmd, "", fmt.Errorf("readByteWithTimeout: %v", err)
		}
		if r == byte('\n') {
			break
		}
		if timedOut {
			klog.Infof("timed out waiting for input: possibly due to an old kubectl version.")
			break
		}
		outData = append(outData, r)
	}
	klog.Infof("proxy stdout: %s", string(outData))
	return cmd, hostPortRe.FindString(string(outData)), nil
}

// readByteWithTimeout returns a byte from a reader or an indicator that a timeout has occurred.
func readByteWithTimeout(r io.ByteReader, timeout time.Duration) (byte, bool, error) {
	bc := make(chan byte, 1)
	ec := make(chan error, 1)
	defer func() {
		close(bc)
		close(ec)
	}()
	go func() {
		b, err := r.ReadByte()
		if err != nil {
			ec <- err
		} else {
			bc <- b
		}
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
func dashboardURL(addr string, ns string, svc string) string {
	// Reference: https://github.com/kubernetes/dashboard/wiki/Accessing-Dashboard---1.7.X-and-above
	return fmt.Sprintf("http://%s/api/v1/namespaces/%s/services/http:%s:/proxy/", addr, ns, svc)
}

// checkURL checks if a URL returns 200 HTTP OK
func checkURL(url string) error {
	resp, err := http.Get(url)
	klog.Infof("%s response: %v %+v", url, err, resp)
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
	dashboardCmd.Flags().IntVar(&dashboardExposedPort, "port", 0, "Exposed port of the proxyfied dashboard. Set to 0 to pick a random port.")
}
