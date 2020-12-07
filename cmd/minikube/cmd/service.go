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
	Long:  `Returns the Kubernetes URL for a service in your local cluster. In the case of multiple URLs they will be printed one at a time.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		t, err := template.New("serviceURL").Parse(serviceURLFormat)
		if err != nil {
			exit.Error(reason.InternalFormatUsage, "The value passed to --format is invalid", err)
		}
		serviceURLTemplate = t

		RootCmd.PersistentPreRun(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			exit.Message(reason.Usage, "You must specify a service name")
		}

		svc := args[0]

		cname := ClusterFlagValue()
		co := mustload.Healthy(cname)

		urls, err := service.WaitForService(co.API, co.Config.Name, namespace, svc, serviceURLTemplate, serviceURLMode, https, wait, interval)
		if err != nil {
			var s *service.SVCNotFoundError
			if errors.As(err, &s) {
				exit.Message(reason.SvcNotFound, `Service '{{.service}}' was not found in '{{.namespace}}' namespace.
You may select another namespace by using 'minikube service {{.service}} -n <namespace>'. Or list out all the services using 'minikube service list'`, out.V{"service": svc, "namespace": namespace})
			}
			exit.Error(reason.SvcTimeout, "Error opening service", err)
		}

		if driver.NeedsPortForward(co.Config.Driver) {
			startKicServiceTunnel(svc, cname)
			return
		}

		openURLs(svc, urls)
	},
}

func init() {
	serviceCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "The service namespace")
	serviceCmd.Flags().BoolVar(&serviceURLMode, "url", false, "Display the Kubernetes service URL in the CLI instead of opening it in the default browser")
	serviceCmd.Flags().BoolVar(&https, "https", false, "Open the service URL with https instead of http (defaults to \"false\")")
	serviceCmd.Flags().IntVar(&wait, "wait", service.DefaultWait, "Amount of time to wait for a service in seconds")
	serviceCmd.Flags().IntVar(&interval, "interval", service.DefaultInterval, "The initial time interval for each check that wait performs in seconds")

	serviceCmd.PersistentFlags().StringVar(&serviceURLFormat, "format", defaultServiceFormatTemplate, "Format to output service URL in. This format will be applied to each url individually and they will be printed one at a time.")
}

func startKicServiceTunnel(svc, configName string) {
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)

	clientset, err := kapi.Client(configName)
	if err != nil {
		exit.Error(reason.InternalKubernetesClient, "error creating clientset", err)
	}

	port, err := oci.ForwardedPort(oci.Docker, configName, 22)
	if err != nil {
		exit.Error(reason.DrvPortForward, "error getting ssh port", err)
	}
	sshPort := strconv.Itoa(port)
	sshKey := filepath.Join(localpath.MiniPath(), "machines", configName, "id_rsa")

	serviceTunnel := kic.NewServiceTunnel(sshPort, sshKey, clientset.CoreV1())
	urls, err := serviceTunnel.Start(svc, namespace)
	if err != nil {
		exit.Error(reason.SvcTunnelStart, "error starting tunnel", err)
	}

	// wait for tunnel to come up
	time.Sleep(1 * time.Second)

	data := [][]string{{namespace, svc, "", strings.Join(urls, "\n")}}
	service.PrintServiceList(os.Stdout, data)

	openURLs(svc, urls)
	out.WarningT("Because you are using a Docker driver on {{.operating_system}}, the terminal needs to be open to run it.", out.V{"operating_system": runtime.GOOS})

	<-ctrlC

	err = serviceTunnel.Stop()
	if err != nil {
		exit.Error(reason.SvcTunnelStop, "error stopping tunnel", err)
	}
}

func openURLs(svc string, urls []string) {
	for _, u := range urls {
		_, err := url.Parse(u)
		if err != nil {
			klog.Warningf("failed to parse url %q: %v (will not open)", u, err)
			out.String(fmt.Sprintf("%s\n", u), false)
			continue
		}

		if serviceURLMode {
			out.String(fmt.Sprintf("%s\n", u), false)
			continue
		}

		out.Step(style.Celebrate, "Opening service {{.namespace_name}}/{{.service_name}} in default browser...", false, out.V{"namespace_name": namespace, "service_name": svc})
		if err := browser.OpenURL(u); err != nil {
			exit.Error(reason.HostBrowser, fmt.Sprintf("open url failed: %s", u), err)
		}
	}
}
