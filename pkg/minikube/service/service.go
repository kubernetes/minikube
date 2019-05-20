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
	"strings"
	"text/template"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/proxy"
	"k8s.io/minikube/pkg/util"
)

// K8sClient represents a kubernetes client
type K8sClient interface {
	GetCoreClient() (typed_core.CoreV1Interface, error)
	GetClientset(timeout time.Duration) (*kubernetes.Clientset, error)
}

// K8sClientGetter can get a K8sClient
type K8sClientGetter struct{}

// K8s is the current K8sClient
var K8s K8sClient

func init() {
	K8s = &K8sClientGetter{}
}

// GetCoreClient returns a core client
func (k *K8sClientGetter) GetCoreClient() (typed_core.CoreV1Interface, error) {
	client, err := k.GetClientset(constants.DefaultK8sClientTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "getting clientset")
	}
	return client.CoreV1(), nil
}

// GetClientset returns a clientset
func (*K8sClientGetter) GetClientset(timeout time.Duration) (*kubernetes.Clientset, error) {
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
		return nil, fmt.Errorf("kubeConfig: %v", err)
	}
	clientConfig.Timeout = timeout
	clientConfig = proxy.UpdateTransport(clientConfig)
	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.Wrap(err, "client from config")
	}

	return client, nil
}

// URL represents service URL
type URL struct {
	Namespace string
	Name      string
	URLs      []string
}

// URLs represents a list of URL
type URLs []URL

// GetServiceURLs returns all the node port URLs for every service in a particular namespace
// Accepts a template for formatting
func GetServiceURLs(api libmachine.API, namespace string, t *template.Template) (URLs, error) {
	host, err := cluster.CheckIfHostExistsAndLoad(api, config.GetMachineName())
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

	svcs, err := serviceInterface.List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	var serviceURLs []URL
	for _, svc := range svcs.Items {
		urls, err := printURLsForService(client, ip, svc.Name, svc.Namespace, t)
		if err != nil {
			return nil, err
		}
		serviceURLs = append(serviceURLs, URL{Namespace: svc.Namespace, Name: svc.Name, URLs: urls})
	}

	return serviceURLs, nil
}

// GetServiceURLsForService returns all the node ports for a service in a namespace
// with optional formatting
func GetServiceURLsForService(api libmachine.API, namespace, service string, t *template.Template) ([]string, error) {
	host, err := cluster.CheckIfHostExistsAndLoad(api, config.GetMachineName())
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

func printURLsForService(c typed_core.CoreV1Interface, ip, service, namespace string, t *template.Template) ([]string, error) {
	if t == nil {
		return nil, errors.New("Error, attempted to generate service url with nil --format template")
	}

	s := c.Services(namespace)
	svc, err := s.Get(service, meta.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "service '%s' could not be found running", service)
	}

	e := c.Endpoints(namespace)
	endpoints, err := e.Get(service, meta.GetOptions{})
	m := make(map[int32]string)
	if err == nil && endpoints != nil && len(endpoints.Subsets) > 0 {
		for _, ept := range endpoints.Subsets {
			for _, p := range ept.Ports {
				m[p.Port] = p.Name
			}
		}
	}

	urls := []string{}
	for _, port := range svc.Spec.Ports {
		if port.NodePort > 0 {
			var doc bytes.Buffer
			err = t.Execute(&doc, struct {
				IP   string
				Port int32
				Name string
			}{
				ip,
				port.NodePort,
				m[port.TargetPort.IntVal],
			})
			if err != nil {
				return nil, err
			}
			urls = append(urls, doc.String())
		}
	}
	return urls, nil
}

// CheckService checks if a service is listening on a port.
func CheckService(namespace string, service string) error {
	client, err := K8s.GetCoreClient()
	if err != nil {
		return errors.Wrap(err, "Error getting kubernetes client")
	}

	svc, err := client.Services(namespace).Get(service, meta.GetOptions{})
	if err != nil {
		return &util.RetriableError{
			Err: errors.Wrapf(err, "Error getting service %s", service),
		}
	}
	if len(svc.Spec.Ports) == 0 {
		return fmt.Errorf("%s:%s has no ports", namespace, service)
	}
	glog.Infof("Found service: %+v", svc)
	return nil
}

// OptionallyHTTPSFormattedURLString returns a formatted URL string, optionally HTTPS
func OptionallyHTTPSFormattedURLString(bareURLString string, https bool) (string, bool) {
	httpsFormattedString := bareURLString
	isHTTPSchemedURL := false

	if u, parseErr := url.Parse(bareURLString); parseErr == nil {
		isHTTPSchemedURL = u.Scheme == "http"
	}

	if isHTTPSchemedURL && https {
		httpsFormattedString = strings.Replace(bareURLString, "http", "https", 1)
	}

	return httpsFormattedString, isHTTPSchemedURL
}

// WaitAndMaybeOpenService waits for a service, and opens it when running
func WaitAndMaybeOpenService(api libmachine.API, namespace string, service string, urlTemplate *template.Template, urlMode bool, https bool,
	wait int, interval int) error {
	if err := util.RetryAfter(wait, func() error { return CheckService(namespace, service) }, time.Duration(interval)*time.Second); err != nil {
		return errors.Wrapf(err, "Could not find finalized endpoint being pointed to by %s", service)
	}

	urls, err := GetServiceURLsForService(api, namespace, service, urlTemplate)
	if err != nil {
		return errors.Wrap(err, "Check that minikube is running and that you have specified the correct namespace")
	}
	for _, bareURLString := range urls {
		urlString, isHTTPSchemedURL := OptionallyHTTPSFormattedURLString(bareURLString, https)

		if urlMode || !isHTTPSchemedURL {
			console.OutLn(urlString)
		} else {
			console.OutStyle("celebrate", "Opening kubernetes service %s/%s in default browser...", namespace, service)
			if err := browser.OpenURL(urlString); err != nil {
				console.Err("browser failed to open url: %v", err)
			}
		}
	}
	return nil
}

// GetServiceListByLabel returns a ServiceList by label
func GetServiceListByLabel(namespace string, key string, value string) (*core.ServiceList, error) {
	client, err := K8s.GetCoreClient()
	if err != nil {
		return &core.ServiceList{}, &util.RetriableError{Err: err}
	}
	return getServiceListFromServicesByLabel(client.Services(namespace), key, value)
}

func getServiceListFromServicesByLabel(services typed_core.ServiceInterface, key string, value string) (*core.ServiceList, error) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{key: value}))
	serviceList, err := services.List(meta.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return &core.ServiceList{}, &util.RetriableError{Err: err}
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
	secret, _ := secrets.Get(name, meta.GetOptions{})

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
	secretObj := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Data: data,
		Type: core.SecretTypeOpaque,
	}

	_, err = secrets.Create(secretObj)
	if err != nil {
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
	err = secrets.Delete(name, &meta.DeleteOptions{})
	if err != nil {
		return &util.RetriableError{Err: err}
	}

	return nil
}
