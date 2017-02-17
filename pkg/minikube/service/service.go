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

package service

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"k8s.io/client-go/1.5/kubernetes"
	corev1 "k8s.io/client-go/1.5/kubernetes/typed/core/v1"
	kubeapi "k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"

	"text/template"

	"k8s.io/client-go/1.5/pkg/labels"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/util"
)

type ServiceURL struct {
	Namespace string
	Name      string
	URLs      []string
}

type ServiceURLs []ServiceURL

func GetServiceURLs(api libmachine.API, namespace string, t *template.Template) (ServiceURLs, error) {
	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		return nil, err
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, err
	}

	client, err := cluster.GetKubernetesClient()
	if err != nil {
		return nil, err
	}

	getter := client.Services(namespace)

	svcs, err := getter.List(kubeapi.ListOptions{})
	if err != nil {
		return nil, err
	}

	var serviceURLs []ServiceURL

	for _, svc := range svcs.Items {
		urls, err := getServiceURLsWithClient(client, ip, svc.Namespace, svc.Name, t)
		if err != nil {
			if _, ok := err.(MissingNodePortError); ok {
				serviceURLs = append(serviceURLs, ServiceURL{Namespace: svc.Namespace, Name: svc.Name})
				continue
			}
			return nil, err
		}
		serviceURLs = append(serviceURLs, ServiceURL{Namespace: svc.Namespace, Name: svc.Name, URLs: urls})
	}

	return serviceURLs, nil
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
	endpoint, err := endpoints.Get(service)
	if err != nil {
		return &util.RetriableError{Err: err}
	}
	return checkEndpointReady(endpoint)
}

func checkEndpointReady(endpoint *v1.Endpoints) error {
	const notReadyMsg = "Waiting, endpoint for service is not ready yet...\n"
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

func WaitAndMaybeOpenService(api libmachine.API, namespace string, service string, urlTemplate *template.Template, urlMode bool, https bool) {
	if err := util.RetryAfter(20, func() error { return CheckService(namespace, service) }, 6*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "Could not find finalized endpoint being pointed to by %s: %s\n", service, err)
		os.Exit(1)
	}

	urls, err := GetServiceURLsForService(api, namespace, service, urlTemplate)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Check that minikube is running and that you have specified the correct namespace (-n flag).")
		os.Exit(1)
	}
	for _, url := range urls {
		if https {
			url = strings.Replace(url, "http", "https", 1)
		}
		if urlMode || !strings.HasPrefix(url, "http") {
			fmt.Fprintln(os.Stdout, url)
		} else {
			fmt.Fprintln(os.Stdout, "Opening kubernetes service "+namespace+"/"+service+" in default browser...")
			browser.OpenURL(url)
		}
	}
}

func GetServiceListByLabel(namespace string, key string, value string) (*v1.ServiceList, error) {
	client, err := cluster.GetKubernetesClient()
	if err != nil {
		return &v1.ServiceList{}, &util.RetriableError{Err: err}
	}
	services := client.Services(namespace)
	if err != nil {
		return &v1.ServiceList{}, &util.RetriableError{Err: err}
	}
	return getServiceListFromServicesByLabel(services, key, value)
}

func getServiceListFromServicesByLabel(services corev1.ServiceInterface, key string, value string) (*v1.ServiceList, error) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{key: value}))
	serviceList, err := services.List(kubeapi.ListOptions{LabelSelector: selector})
	if err != nil {
		return &v1.ServiceList{}, &util.RetriableError{Err: err}
	}

	return serviceList, nil
}

func getServicePortsFromServiceGetter(services serviceGetter, service string) ([]int32, error) {
	svc, err := services.Get(service)
	if err != nil {
		return nil, fmt.Errorf("Error getting %s service: %s", service, err)
	}
	var nodePorts []int32
	if len(svc.Spec.Ports) > 0 {
		for _, port := range svc.Spec.Ports {
			if port.NodePort > 0 {
				nodePorts = append(nodePorts, port.NodePort)
			}
		}
	}
	if len(nodePorts) == 0 {
		return nil, MissingNodePortError{svc}
	}
	return nodePorts, nil
}

type serviceGetter interface {
	Get(name string) (*v1.Service, error)
	List(kubeapi.ListOptions) (*v1.ServiceList, error)
}

func getServicePorts(client *kubernetes.Clientset, namespace, service string) ([]int32, error) {
	services := client.Services(namespace)
	return getServicePortsFromServiceGetter(services, service)
}

type MissingNodePortError struct {
	service *v1.Service
}

func (e MissingNodePortError) Error() string {
	return fmt.Sprintf("Service %s/%s does not have a node port. To have one assigned automatically, the service type must be NodePort or LoadBalancer, but this service is of type %s.", e.service.Namespace, e.service.Name, e.service.Spec.Type)
}

func GetServiceURLsForService(api libmachine.API, namespace, service string, t *template.Template) ([]string, error) {
	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		return nil, errors.Wrap(err, "Error checking if api exist and loading it")
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ip from host")
	}

	client, err := cluster.GetKubernetesClient()
	if err != nil {
		return nil, err
	}

	return getServiceURLsWithClient(client, ip, namespace, service, t)
}

func getServiceURLsWithClient(client *kubernetes.Clientset, ip, namespace, service string, t *template.Template) ([]string, error) {
	if t == nil {
		return nil, errors.New("Error, attempted to generate service url with nil --format template")
	}

	ports, err := getServicePorts(client, namespace, service)
	if err != nil {
		return nil, err
	}
	urls := []string{}
	for _, port := range ports {

		var doc bytes.Buffer
		err = t.Execute(&doc, struct {
			IP   string
			Port int32
		}{
			ip,
			port,
		})
		if err != nil {
			return nil, err
		}

		u, err := url.Parse(doc.String())
		if err != nil {
			return nil, err
		}

		urls = append(urls, u.String())
	}
	return urls, nil
}

func ValidateService(namespace string, service string) error {
	client, err := cluster.GetKubernetesClient()
	if err != nil {
		return errors.Wrap(err, "error validating input service name")
	}
	services := client.Services(namespace)
	if _, err = services.Get(service); err != nil {
		return errors.Wrapf(err, "service '%s' could not be found running in namespace '%s' within kubernetes", service, namespace)
	}
	return nil
}
