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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	kubeApi "k8s.io/kubernetes/pkg/api"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"

	commonutil "k8s.io/minikube/pkg/util"
)

var (
	namespace      string
	https          bool
	serviceURLMode bool
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service [flags] SERVICE",
	Short: "Gets the kubernetes URL for the specified service in your local cluster",
	Long:  `Gets the kubernetes URL for the specified service in your local cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Please specify a service name.")
			os.Exit(1)
		}

		service := args[0]
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		cluster.EnsureMinikubeRunningOrExit(api)
		if err := commonutil.RetryAfter(20, func() error { return CheckService(namespace, service) }, 6*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "Could not find finalized endpoint being pointed to by %s: %s\n", service, err)
			os.Exit(1)
		}

		url, err := cluster.GetServiceURL(api, namespace, service)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "Check that minikube is running and that you have specified the correct namespace (-n flag).")
			os.Exit(1)
		}
		if https {
			url = strings.Replace(url, "http", "https", 1)
		}
		if serviceURLMode {
			fmt.Fprintln(os.Stdout, url)
		} else {
			fmt.Fprintln(os.Stdout, "Opening kubernetes service "+namespace+"/"+service+" in default browser...")
			browser.OpenURL(url)
		}
	},
}

func init() {
	serviceCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "The service namespace")
	serviceCmd.Flags().BoolVar(&serviceURLMode, "url", false, "Display the kubernetes service URL in the CLI instead of opening it in the default browser")
	serviceCmd.Flags().BoolVar(&https, "https", false, "Open the service URL with https instead of http")
	RootCmd.AddCommand(serviceCmd)
}

// CheckService waits for the specified service to be ready by returning an error until the service is up
// The check is done by polling the endpoint associated with the service and when the endpoint exists, returning no error->service-online
func CheckService(namespace string, service string) error {
	endpoints, err := cluster.GetKubernetesEndpointsWithNamespace(namespace)
	if err != nil {
		return err
	}
	endpoint, err := endpoints.Get(service)
	if err != nil {
		return err
	}
	return CheckEndpointReady(endpoint)
}

func CheckEndpointReady(endpoint *kubeApi.Endpoints) error {
	if len(endpoint.Subsets) == 0 {
		fmt.Fprintf(os.Stderr, "Waiting, endpoint for service is not ready yet...")
		return errors.New("Endpoint for service is not ready yet")
	}
	for _, subset := range endpoint.Subsets {
		if len(subset.NotReadyAddresses) != 0 {
			fmt.Fprintf(os.Stderr, "Waiting, endpoint for service is not ready yet...")
			return errors.New("Endpoint for service is not ready yet")
		}
	}
	return nil
}
