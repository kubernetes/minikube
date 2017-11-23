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
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"

	"text/template"

	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/labels"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/util"
)

type K8sClient interface {
	GetCoreClient() (corev1.CoreV1Interface, error)
	GetClientset() (*kubernetes.Clientset, error)
}

type K8sClientGetter struct{}

var K8s K8sClient

func init() {
	K8s = &K8sClientGetter{}
}

func (k *K8sClientGetter) GetCoreClient() (corev1.CoreV1Interface, error) {
	client, err := k.GetClientset()
	if err != nil {
		return nil, errors.Wrap(err, "getting clientset")
	}
	return client.Core(), nil
}

func (*K8sClientGetter) GetClientset() (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	profile := viper.GetString(config.MachineProfile)
	configOverrides := &clientcmd.ConfigOverrides{
		Context: clientcmdapi.Context{
			Cluster:  profile,
			AuthInfo: profile,
		},
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Error creating kubeConfig: %s", err)
	}
	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new client from kubeConfig.ClientConfig()")
	}

	return client, nil
}

type ServiceURL struct {
	Namespace string
	Name      string
	URLs      []string
}

type ServiceURLs []ServiceURL

// Returns all the node port URLs for every service in a particular namespace
// Accepts a template for formating
func GetServiceURLs(api libmachine.API, namespace string, t *template.Template) (ServiceURLs, error) {
	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		return nil, err
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, err
	}

	client, err := K8s.GetCoreClient()
	if err != nil {
		return nil, err
	}

	serviceInterface := client.Services(namespace)

	svcs, err := serviceInterface.List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var serviceURLs []ServiceURL
	for _, svc := range svcs.Items {
		urls, err := printURLsForService(client, ip, svc.Name, svc.Namespace, t)
		if err != nil {
			return nil, err
		}
		serviceURLs = append(serviceURLs, ServiceURL{Namespace: svc.Namespace, Name: svc.Name, URLs: urls})
	}

	return serviceURLs, nil
}

// Returns all the node ports for a service in a namespace
// with optional formatting
func GetServiceURLsForService(api libmachine.API, namespace, service string, t *template.Template) ([]string, error) {
	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		return nil, errors.Wrap(err, "Error checking if api exist and loading it")
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ip from host")
	}

	client, err := K8s.GetCoreClient()
	if err != nil {
		return nil, err
	}

	return printURLsForService(client, ip, service, namespace, t)
}

func printURLsForService(c corev1.CoreV1Interface, ip, service, namespace string, t *template.Template) ([]string, error) {
	if t == nil {
		return nil, errors.New("Error, attempted to generate service url with nil --format template")
	}

	s := c.Services(namespace)
	svc, err := s.Get(service, meta_v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "service '%s' could not be found running", service)
	}
	var nodePorts []int32
	if len(svc.Spec.Ports) > 0 {
		for _, port := range svc.Spec.Ports {
			if port.NodePort > 0 {
				nodePorts = append(nodePorts, port.NodePort)
			}
		}
	}
	urls := []string{}
	for _, port := range nodePorts {
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

// CheckService waits for the specified service to be ready by returning an error until the service is up
// The check is done by polling the endpoint associated with the service and when the endpoint exists, returning no error->service-online
func CheckService(namespace string, service string) error {
	client, err := K8s.GetCoreClient()
	if err != nil {
		return errors.Wrap(err, "Error getting kubernetes client")
	}
	services := client.Services(namespace)
	err = validateService(services, service)
	if err != nil {
		return errors.Wrap(err, "Error validating service")
	}
	endpoints := client.Endpoints(namespace)
	return checkEndpointReady(endpoints, service)
}

func validateService(s corev1.ServiceInterface, service string) error {
	if _, err := s.Get(service, meta_v1.GetOptions{}); err != nil {
		return errors.Wrapf(err, "Error getting service %s", service)
	}
	return nil
}

func checkEndpointReady(endpoints corev1.EndpointsInterface, service string) error {
	endpoint, err := endpoints.Get(service, meta_v1.GetOptions{})
	if err != nil {
		return &util.RetriableError{Err: errors.Errorf("Error getting endpoints for service %s", service)}
	}
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

func WaitAndMaybeOpenService(api libmachine.API, namespace string, service string, urlTemplate *template.Template, urlMode bool, https bool,
	wait int, interval int) error {
	if err := util.RetryAfter(wait, func() error { return CheckService(namespace, service) }, time.Duration(interval)*time.Second); err != nil {
		return errors.Wrapf(err, "Could not find finalized endpoint being pointed to by %s", service)
	}

	urls, err := GetServiceURLsForService(api, namespace, service, urlTemplate)
	if err != nil {
		return errors.Wrap(err, "Check that minikube is running and that you have specified the correct namespace")
	}
	for _, url := range urls {
		if https {
			url = strings.Replace(url, "http", "https", 1)
		}
		if urlMode || !strings.HasPrefix(url, "http") {
			fmt.Fprintln(os.Stdout, url)
		} else {
			fmt.Fprintln(os.Stderr, "Opening kubernetes service "+namespace+"/"+service+" in default browser...")
			browser.OpenURL(url)
		}
	}
	return nil
}

func GetServiceListByLabel(namespace string, key string, value string) (*v1.ServiceList, error) {
	client, err := K8s.GetCoreClient()
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
	serviceList, err := services.List(meta_v1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return &v1.ServiceList{}, &util.RetriableError{Err: err}
	}

	return serviceList, nil
}

// CreateSecret creates or modifies secrets
func CreateSecret(namespace, name string, dataValues map[string]string, labels map[string]string) error {
	client, err := K8s.GetCoreClient()
	if err != nil {
		return &util.RetriableError{Err: err}
	}
	secrets := client.Secrets(namespace)
	if err != nil {
		return &util.RetriableError{Err: err}
	}

	secret, _ := secrets.Get(name, meta_v1.GetOptions{})

	// Delete existing secret
	if len(secret.Name) > 0 {
		err = DeleteSecret(namespace, name)
		if err != nil {
			return &util.RetriableError{Err: err}
		}
	}

	// convert strings to data secrets
	data := map[string][]byte{}
	for key, value := range dataValues {
		data[key] = []byte(value)
	}

	// Create Secret
	secretObj := &v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Data: data,
		Type: v1.SecretTypeOpaque,
	}

	_, err = secrets.Create(secretObj)
	if err != nil {
		fmt.Println("err: ", err)
		return &util.RetriableError{Err: err}
	}

	return nil
}

// DeleteSecret deletes a secret from a namespace
func DeleteSecret(namespace, name string) error {
	client, err := K8s.GetCoreClient()
	if err != nil {
		return &util.RetriableError{Err: err}
	}

	secrets := client.Secrets(namespace)
	if err != nil {
		return &util.RetriableError{Err: err}
	}

	err = secrets.Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return &util.RetriableError{Err: err}
	}

	return nil
}
