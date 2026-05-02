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
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/cmd/minikube/cmd/flags"
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
	dashboardProvider    string
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
		options := flags.CommandOptions()
		cname := ClusterFlagValue()
		co := mustload.Healthy(cname, options)

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
		addonName := dashboardAddonName(dashboardProvider) 
		addon := assets.Addons[addonName]
		enabled := addon.IsEnabled(co.Config)

		if !enabled {
			// Send status messages to stderr for folks reusing this output.
			out.ErrT(style.Enabling, "Enabling dashboard ...")
			// Enable the dashboard add-on
			err = addons.SetAndSave(cname, addonName, "true", options)
			if err != nil {
				exit.Error(reason.InternalAddonEnable, "Unable to enable dashboard", err)
			}
		}

		ns := dashboardAddon_resourceName(dashboardProvider)
		svc := dashboardAddon_resourceName(dashboardProvider)
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
		url := dashboardURL(hostPort, ns, svc, co)	

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
			out.Styled(style.Celebrate, `Opening {{.url}} in your default browser...`, out.V{"url": url})
			addons.PostStartMessages(co.Config, addonName, "true")
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
		return nil, "", fmt.Errorf("cmd stdout: %w", err)
	}

	klog.Infof("Executing: %s %s", cmd.Path, cmd.Args)
	if err := cmd.Start(); err != nil {
		return nil, "", fmt.Errorf("proxy start: %w", err)
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

//gives name mapped in addon, ex: yakd -> dashboard, headlamp, etc
func dashboardAddonName(provider string) string{
	switch(provider){
		case "yakd":  		  return "dashboard"
		case "headlamp": 	  return "headlamp" 
	}
	return "dashboard"
}

//gives namespace & service name for that dashboard addon
func dashboardAddon_resourceName(provider string) string {
	switch(provider) {
		case "yakd":   return "kubernetes-dashboard"	
		case "headlamp":   return "headlamp"	
	}
	return "kubernetes-dashboard"
}

// dashboardURL generates a URL for accessing the dashboard service
func dashboardURL(addr string, ns string, svc string, co mustload.ClusterController) string {
	switch (svc) {
		default:
			return fmt.Sprintf("http://%s/api/v1/namespaces/%s/services/http:%s:/proxy/", addr, ns, svc)
		case "headlamp":
			serviceURLTemplate := template.Must(template.New("serviceURL").Parse("http://{{.IP}}:{{.Port}}"))  
			serviceURLs, err := service.GetServiceURLsForService(co.API, co.Config.Name, ns, svc, serviceURLTemplate)  
			if err != nil {  
				exit.Error(reason.SvcTimeout, "Error getting service URL", err)  
			}
			if len(serviceURLs.URLs) == 0 {  
				exit.Message(reason.SvcNotFound, fmt.Sprintf("No URL found for %s service", svc))  
			}  
			serviceURL := serviceURLs.URLs[0]  
			return serviceURL
	}
}

// checkURL checks if a URL returns 200 HTTP OK
func checkURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("hitting URL:%q\n response: %+v: %w", url, resp, err)
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
	dashboardCmd.Flags().StringVar(&dashboardProvider, "provider", "yakd", "Which dashboard to run, such as headlamp, yakd(default = kubernetes dashboard, i.e: yakd)")
}
