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
	"bytes"
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
	pkgnetwork "k8s.io/minikube/pkg/network"
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
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 && !all || (len(args) > 0 && all) {
			exit.Message(reason.Usage, "You must specify service name(s) or --all")
		}

		svcArgs := make(map[string]bool)
		for _, v := range args {
			svcArgs[v] = true
		}

		cname := ClusterFlagValue()
		co := mustload.Healthy(cname)

		if driver.IsQEMU(co.Config.Driver) && pkgnetwork.IsBuiltinQEMU(co.Config.Network) {
			msg := "minikube service is not currently implemented with the builtin network on QEMU"
			if runtime.GOOS == "darwin" {
				msg += ", try starting minikube with '--network=socket_vmnet'"
			}
			exit.Message(reason.Unimplemented, msg)
		}

		var services service.URLs
		services, err := service.GetServiceURLs(co.API, co.Config.Name, namespace, serviceURLTemplate)
		if err != nil {
			out.ErrT(style.Fatal, "Failed to get service URL - check that minikube is running and that you have specified the correct namespace (-n flag) if required: {{.error}}", out.V{"error": err})
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

		if len(services) == 0 && all {
			exit.Message(reason.SvcNotFound, `No services were found in the '{{.namespace}}' namespace.
You may select another namespace by using 'minikube service --all -n <namespace>'`, out.V{"namespace": namespace})
		} else if len(services) == 0 {
			exit.Message(reason.SvcNotFound, `Service '{{.service}}' was not found in '{{.namespace}}' namespace.
You may select another namespace by using 'minikube service {{.service}} -n <namespace>'. Or list out all the services using 'minikube service list'`, out.V{"service": args[0], "namespace": namespace})
		}

		var data [][]string
		var noNodePortServices service.URLs

		for _, svc := range services {
			openUrls, err := service.WaitForService(co.API, co.Config.Name, namespace, svc.Name, serviceURLTemplate, serviceURLMode, https, wait, interval)

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
				noNodePortServices = append(noNodePortServices, svc)
			} else {
				servicePortNames := strings.Join(svc.PortNames, "\n")
				serviceURLs := strings.Join(openUrls, "\n")

				// if we are running Docker on OSX we empty the internal service URLs
				if runtime.GOOS == "darwin" && co.Config.Driver == oci.Docker {
					serviceURLs = ""
				}

				data = append(data, []string{svc.Namespace, svc.Name, servicePortNames, serviceURLs})

				if serviceURLMode && !driver.NeedsPortForward(co.Config.Driver) {
					out.Stringf("%s\n", serviceURLs)
				}
			}
			// check whether there are running pods for this service
			if err := service.CheckServicePods(cname, svc.Name, namespace); err != nil {
				exit.Error(reason.SvcUnreachable, "service not available", err)
			}
		}

		noNodePortSvcNames := []string{}
		for _, svc := range noNodePortServices {
			noNodePortSvcNames = append(noNodePortSvcNames, fmt.Sprintf("%s/%s", svc.Namespace, svc.Name))
		}
		if len(noNodePortServices) != 0 {
			out.WarningT("Services {{.svc_names}} have type \"ClusterIP\" not meant to be exposed, however for local development minikube allows you to access this !", out.V{"svc_names": noNodePortSvcNames})
		}

		if driver.NeedsPortForward(co.Config.Driver) && services != nil {
			startKicServiceTunnel(services, cname, co.Config.Driver)
		} else if !serviceURLMode {
			openURLs(data)
			if len(noNodePortServices) != 0 {
				startKicServiceTunnel(noNodePortServices, cname, co.Config.Driver)
			}

		}
	},
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

func startKicServiceTunnel(services service.URLs, configName, driverName string) {
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

		serviceTunnel := kic.NewServiceTunnel(sshPort, sshKey, clientset.CoreV1(), serviceURLMode)
		urls, err := serviceTunnel.Start(svc.Name, namespace)

		if err != nil {
			exit.Error(reason.SvcTunnelStart, "error starting tunnel", err)
		}
		// mutate response urls to HTTPS if needed
		urls, err = mutateURLs(svc.Name, urls)

		if err != nil {
			exit.Error(reason.SvcTunnelStart, "error creating urls", err)
		}

		defer serviceTunnel.Stop()
		svc.URLs = urls
		data = append(data, []string{namespace, svc.Name, "", strings.Join(urls, "\n")})
	}

	time.Sleep(1 * time.Second)

	if !serviceURLMode {
		service.PrintServiceList(os.Stdout, data)
	} else {
		for _, row := range data {
			out.Stringf("%s\n", row[3])
		}
	}

	if !serviceURLMode {
		openURLs(data)
	}

	out.WarningT("Because you are using a Docker driver on {{.operating_system}}, the terminal needs to be open to run it.", out.V{"operating_system": runtime.GOOS})

	<-ctrlC
}

func mutateURLs(serviceName string, urls []string) ([]string, error) {
	formattedUrls := make([]string, 0)
	for _, rawURL := range urls {
		var doc bytes.Buffer
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			exit.Error(reason.SvcTunnelStart, "No valid URL found for tunnel.", err)
		}
		port, err := strconv.Atoi(parsedURL.Port())
		if err != nil {
			exit.Error(reason.SvcTunnelStart, "No valid port found for tunnel.", err)
		}
		err = serviceURLTemplate.Execute(&doc, struct {
			IP   string
			Port int32
			Name string
		}{
			parsedURL.Hostname(),
			int32(port),
			serviceName,
		})

		if err != nil {
			return nil, err
		}

		httpsURL, _ := service.OptionallyHTTPSFormattedURLString(doc.String(), https)
		formattedUrls = append(formattedUrls, httpsURL)
	}

	return formattedUrls, nil
}

func openURLs(urls [][]string) {
	for _, u := range urls {

		if len(u) < 4 {
			klog.Warning("No URL found")
			continue
		}

		_, err := url.Parse(u[3])
		if err != nil {
			klog.Warningf("failed to parse url %q: %v (will not open)", u[3], err)
			out.Stringf("%s\n", u)
			continue
		}

		if serviceURLMode {
			out.Stringf("%s\n", u)
			continue
		}

		out.Styled(style.Celebrate, "Opening service {{.namespace_name}}/{{.service_name}} in default browser...", out.V{"namespace_name": namespace, "service_name": u[1]})
		if err := browser.OpenURL(u[3]); err != nil {
			exit.Error(reason.HostBrowser, fmt.Sprintf("open url failed: %s", u), err)
		}
	}
}
