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
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/cmd/minikube/cmd/flags"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/driver"
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
		addon := assets.Addons["dashboard"]
		enabled := addon.IsEnabled(co.Config)

		if !enabled {
			// Send status messages to stderr for folks reusing this output.
			out.ErrT(style.Enabling, "Enabling dashboard ...")
			// Enable the dashboard add-on
			err = addons.SetAndSave(cname, "dashboard", "true", options)
			if err != nil {
				exit.Error(reason.InternalAddonEnable, "Unable to enable dashboard", err)
			}
		}

		ns := "kubernetes-dashboard"
		svc := "kubernetes-dashboard-kong-proxy"
		out.ErrT(style.Verifying, "Verifying dashboard health ...")
		checkSVC := func() error { return service.CheckService(cname, ns, svc) }
		// for slow machines or parallels in CI to avoid #7503
		if err = retry.Expo(checkSVC, 100*time.Microsecond, time.Minute*10); err != nil {
			exit.Message(reason.SvcCheckTimeout, "dashboard service is not running: {{.error}}", out.V{"error": err})
		}

		// Attempt to finding the service URL directly (e.g. NodePort)
		// If found, we usage it directly and skip the proxy
		// The user requested HTTPS URL
		tmpl := template.Must(template.New("svc-template").Parse("https://{{.IP}}:{{.Port}}"))
		urls, err := service.GetServiceURLs(co.API, cname, ns, tmpl)

		if err != nil {
			// If we can't find the service, exit.
			exit.Message(reason.SvcNotFound, "dashboard service not found: {{.error}}", out.V{"error": err})
		}

		if len(urls) == 0 {
			exit.Message(reason.SvcNotFound, "dashboard service not found")
		}

		// If we need port forwarding (e.g. Docker on Mac), we must start the tunnel
		if driver.NeedsPortForward(co.Config.Driver) {
			startKicServiceTunnel(urls, cname, co.Config.Driver, ns, dashboardURLMode, true, tmpl)
			return
		}

		// Otherwise, we can likely access NodePort directly (linux, vmware, etc)
		// Check if we found any URLs
		found := false
		for _, u := range urls {
			if u.Name == svc && len(u.URLs) > 0 {
				url := u.URLs[0]
				printDashboardToken(kubectlVersion, co.Config.BinaryMirror, cname, ns)
				if dashboardURLMode {
					out.Ln(url)
				} else {
					out.Styled(style.Celebrate, "Opening {{.url}} in your default browser...", out.V{"url": url})
					if err = browser.OpenURL(url); err != nil {
						exit.Message(reason.HostBrowser, "failed to open browser: {{.error}}", out.V{"error": err})
					}
				}
				found = true
				break
			}
		}

		if !found {
			exit.Message(reason.SvcNotFound, "dashboard service has no node port")
		}
	},

}

func printDashboardToken(kubectlVersion, binaryURL, cname, ns string) {
	// 5 = the previous max-attempts for retry.Expo
	token, err := getDashboardToken(kubectlVersion, binaryURL, cname, ns, 5)
	if err != nil {
		out.WarningT("Unable to generate dashboard token: {{.error}}", out.V{"error": err})
		return
	}
	out.Styled(style.Permissions, "Dashboard Token:")
	out.Ln(token)
}


func runKubectl(kubectlVersion, binaryURL, cxtName, ns string, args ...string) (string, error) {
	fullArgs := append([]string{"--context", cxtName, "-n", ns}, args...)
	var cmd *exec.Cmd
	var err error
	if kubectl, errLookup := exec.LookPath("kubectl"); errLookup == nil {
		cmd = exec.Command(kubectl, fullArgs...)
	} else if cmd, err = KubectlCommand(kubectlVersion, binaryURL, fullArgs...); err != nil {
		return "", err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func getDashboardToken(kubectlVersion, binaryURL, ctxName, ns string, attempts int) (string, error) {
	sa := "admin-user" 
	// docs https://github.com/kubernetes/dashboard/blob/master/docs/user/access-control/creating-sample-user.md#creating-a-service-account
	// kubectl -n kubernetes-dashboard create token admin-user
	output, err := runKubectl(kubectlVersion, binaryURL, ctxName, ns, "create", "token", sa, "--duration=24h")
	if err != nil {
		// If the SA doesn't exist (race?), try retry?
		if strings.Contains(output, "not found") && attempts > 0 {
			time.Sleep(500 * time.Millisecond)
			return getDashboardToken(kubectlVersion, binaryURL, ctxName, ns, attempts-1)
		}
		return "", errors.Wrapf(err, "kubectl create token: %s", output)
	}
	return strings.TrimSpace(output), nil
}



func init() {
	dashboardCmd.Flags().BoolVar(&dashboardURLMode, "url", false, "Display dashboard URL instead of opening a browser")
	dashboardCmd.Flags().IntVar(&dashboardExposedPort, "port", 0, "Exposed port of the proxyfied dashboard. Set to 0 to pick a random port.")
}
