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
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/browser"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/tunnel/kic"
)

const defaultServiceFormatTemplate = "http://{{.IP}}:{{.Port}}"

var (
	namespace          string
	all                bool
	https              bool
	serviceURLMode     bool
	serviceURLFormat   string
	serviceURLTemplate *template.Template
	wait               int
	interval           int
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service [flags] SERVICE",
	Short: "Returns a URL to connect to a service",
	Long:  `Returns the Kubernetes URL(s) for service(s) in your local cluster. In the case of multiple URLs they will be printed one at a time.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		t, err := template.New("serviceURL").Parse(serviceURLFormat)
		if err != nil {
			exit.Error(reason.InternalFormatUsage, "The value passed to --format is invalid", err)
		}
		serviceURLTemplate = t

		RootCmd.PersistentPreRun(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 && !all || (len(args) > 0 && all) {
			exit.Message(reason.Usage, "You must specify service name(s) or --all")
		}

		svcArgs := make(map[string]bool)
		for _, v := range args {
			svcArgs[v] = true
		}

		cname := ClusterFlagValue()
		co := mustload.Healthy(cname)
		var services service.URLs
		services, err := service.GetServiceURLs(co.API, co.Config.Name, namespace, serviceURLTemplate)
		if err != nil {
			out.FatalT("Failed to get service URL: {{.error}}", out.V{"error": err})
			out.ErrT(style.Notice, "Check that minikube is running and that you have specified the correct namespace (-n flag) if required.")
			os.Exit(reason.ExSvcUnavailable)
		}

		if len(args) >= 1 {
			var newServices service.URLs
			for _, svc := range services {
				if _, ok := svcArgs[svc.Name]; ok {
					newServices = append(newServices, svc)
				}
			}
			services = newServices
		}

		var data [][]string
		var openUrls []string
		for _, svc := range services {
			openUrls, err := service.WaitForService(co.API, co.Config.Name, namespace, svc.Name, serviceURLTemplate, true, https, wait, interval)

			if err != nil {
				var s *service.SVCNotFoundError
				if errors.As(err, &s) {
					exit.Message(reason.SvcNotFound, `Service '{{.service}}' was not found in '{{.namespace}}' namespace.
You may select another namespace by using 'minikube service {{.service}} -n <namespace>'. Or list out all the services using 'minikube service list'`, out.V{"service": svc, "namespace": namespace})
				}
				exit.Error(reason.SvcTimeout, "Error opening service", err)
			}

			if len(openUrls) == 0 {
				data = append(data, []string{svc.Namespace, svc.Name, "No node port"})
			} else {
				servicePortNames := strings.Join(svc.PortNames, "\n")
				serviceURLs := strings.Join(openUrls, "\n")

				// if we are running Docker on OSX we empty the internal service URLs
				if runtime.GOOS == "darwin" && co.Config.Driver == oci.Docker {
					serviceURLs = ""
				}

				data = append(data, []string{svc.Namespace, svc.Name, servicePortNames, serviceURLs})
			}
		}

		if (!serviceURLMode && serviceURLFormat != defaultServiceFormatTemplate && !all) || all {
			service.PrintServiceList(os.Stdout, data)
		} else if serviceURLMode && !all {
			for _, u := range data {
				out.String(fmt.Sprintf("%s\n", u[3]))
			}
		}

		if driver.NeedsPortForward(co.Config.Driver) {
			startKicServiceTunnel(args, services, cname, co.Config.Driver)
			return
		}

		if !serviceURLMode && !all && len(args) == 1 {
			openURLs(args[0], openUrls)
		}
	},
}

func shouldOpen(args []string) bool {
	if !serviceURLMode && !all && len(args) == 1 {
		return true
	}
	return false
}

func init() {
	serviceCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "The service namespace")
	serviceCmd.Flags().BoolVar(&serviceURLMode, "url", false, "Display the Kubernetes service URL in the CLI instead of opening it in the default browser")
	serviceCmd.Flags().BoolVar(&all, "all", false, "Forwards all services in a namespace (defaults to \"false\")")
	serviceCmd.Flags().BoolVar(&https, "https", false, "Open the service URL with https instead of http (defaults to \"false\")")
	serviceCmd.Flags().IntVar(&wait, "wait", service.DefaultWait, "Amount of time to wait for a service in seconds")
	serviceCmd.Flags().IntVar(&interval, "interval", service.DefaultInterval, "The initial time interval for each check that wait performs in seconds")

	serviceCmd.PersistentFlags().StringVar(&serviceURLFormat, "format", defaultServiceFormatTemplate, "Format to output service URL in. This format will be applied to each url individually and they will be printed one at a time.")
}

func startKicServiceTunnel(args []string, services service.URLs, configName, driverName string) {
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)

	clientset, err := kapi.Client(configName)
	if err != nil {
		exit.Error(reason.InternalKubernetesClient, "error creating clientset", err)
	}

	var data [][]string
	for _, svc := range services {
		port, err := oci.ForwardedPort(driverName, configName, 22)
		if err != nil {
			exit.Error(reason.DrvPortForward, "error getting ssh port", err)
		}
		sshPort := strconv.Itoa(port)
		sshKey := filepath.Join(localpath.MiniPath(), "machines", configName, "id_rsa")

		serviceTunnel := kic.NewServiceTunnel(sshPort, sshKey, clientset.CoreV1())
		urls, err := serviceTunnel.Start(svc.Name, namespace)
		if err != nil {
			exit.Error(reason.SvcTunnelStart, "error starting tunnel", err)
		}
		defer serviceTunnel.Stop()
		data = append(data, []string{namespace, svc.Name, "", strings.Join(urls, "\n")})
	}

	time.Sleep(1 * time.Second)

	if !serviceURLMode && serviceURLFormat != defaultServiceFormatTemplate && !all {
		service.PrintServiceList(os.Stdout, data)
	}

	if shouldOpen(args) {
		openURLs(services[0].Name, services[0].URLs)
	}

	out.WarningT("Because you are using a Docker driver on {{.operating_system}}, the terminal needs to be open to run it.", out.V{"operating_system": runtime.GOOS})

	<-ctrlC
}

func openURLs(svc string, urls []string) {
	for _, u := range urls {
		_, err := url.Parse(u)
		if err != nil {
			klog.Warningf("failed to parse url %q: %v (will not open)", u, err)
			out.String(fmt.Sprintf("%s\n", u))
			continue
		}

		if serviceURLMode {
			out.String(fmt.Sprintf("%s\n", u))
			continue
		}

		out.Styled(style.Celebrate, "Opening service {{.namespace_name}}/{{.service_name}} in default browser...", out.V{"namespace_name": namespace, "service_name": svc})
		if err := browser.OpenURL(u); err != nil {
			exit.Error(reason.HostBrowser, fmt.Sprintf("open url failed: %s", u), err)
		}
	}
}
