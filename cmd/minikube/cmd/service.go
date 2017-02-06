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
	"text/template"

	"github.com/docker/machine/libmachine"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

var (
	namespace          string
	https              bool
	serviceURLMode     bool
	serviceURLFormat   string
	serviceURLTemplate *template.Template
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service [flags] SERVICE",
	Short: "Gets the kubernetes URL(s) for the specified service in your local cluster",
	Long:  `Gets the kubernetes URL(s) for the specified service in your local cluster.  In the case of multiple URLs they will be printed one at a time`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		t, err := template.New("serviceURL").Parse(serviceURLFormat)
		if err != nil {
			fmt.Fprintln(os.Stderr, "The value passed to --format is invalid:\n\n", err)
			os.Exit(1)
		}
		serviceURLTemplate = t
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			errText := "Please specify a service name."
			fmt.Fprintln(os.Stderr, errText)
			os.Exit(1)
		}

		service := args[0]
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		cluster.EnsureMinikubeRunningOrExit(api, 1)
		if err := validateService(namespace, service); err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("service '%s' could not be found running in namespace '%s' within kubernetes",
				service, namespace))
			os.Exit(1)
		}
		cluster.WaitAndMaybeOpenService(api, namespace, service, serviceURLTemplate, serviceURLMode, https)
	},
}

const defaultServiceFormatTemplate = "http://{{.IP}}:{{.Port}}"

func init() {
	serviceCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "The service namespace")
	serviceCmd.Flags().BoolVar(&serviceURLMode, "url", false, "Display the kubernetes service URL in the CLI instead of opening it in the default browser")
	serviceCmd.Flags().BoolVar(&https, "https", false, "Open the service URL with https instead of http")

	serviceCmd.PersistentFlags().StringVar(&serviceURLFormat, "format", defaultServiceFormatTemplate, "Format to output service URL in.  This format will be applied to each url individually and they will be printed one at a time.")

	RootCmd.AddCommand(serviceCmd)
}

func validateService(namespace string, service string) error {
	client, err := cluster.GetKubernetesClient()
	if err != nil {
		return errors.Wrap(err, "error validating input service name")
	}
	services := client.Services(namespace)
	if _, err = services.Get(service, meta_v1.GetOptions{}); err != nil {
		return errors.Wrapf(err, "service '%s' could not be found running in namespace '%s' within kubernetes", service, namespace)
	}
	return nil
}

// CheckService waits for the specified service to be ready by returning an error until the service is up
// The check is done by polling the endpoint associated with the service and when the endpoint exists, returning no error->service-online
func CheckService(namespace string, service string) error {
	client, err := cluster.GetKubernetesClient()
	if err != nil {
		return &util.RetriableError{Err: err}
	}
	endpoints := client.Endpoints(namespace)
	if err != nil {
		return &util.RetriableError{Err: err}
	}
	endpoint, err := endpoints.Get(service, meta_v1.GetOptions{})
	if err != nil {
		return &util.RetriableError{Err: err}
	}
	return CheckEndpointReady(endpoint)
}

const notReadyMsg = "Waiting, endpoint for service is not ready yet...\n"

func CheckEndpointReady(endpoint *v1.Endpoints) error {
	if len(endpoint.Subsets) == 0 {
		fmt.Fprintf(os.Stderr, notReadyMsg)
		return &util.RetriableError{Err: errors.New("Endpoint for service is not ready yet")}
	}
	for _, subset := range endpoint.Subsets {
		if len(subset.Addresses) == 0 {
			fmt.Fprintf(os.Stderr, notReadyMsg)
			return &util.RetriableError{Err: errors.New("No endpoints for service are ready yet")}
		}
	}
	return nil
}
