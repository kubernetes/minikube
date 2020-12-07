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
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util/retry"
)

const (
	// DefaultWait is the default wait time, in seconds
	DefaultWait = 2
	// DefaultInterval is the default interval, in seconds
	DefaultInterval = 1
)

// K8sClient represents a Kubernetes client
type K8sClient interface {
	GetCoreClient(string) (typed_core.CoreV1Interface, error)
}

// K8sClientGetter can get a K8sClient
type K8sClientGetter struct{}

// K8s is the current K8sClient
var K8s K8sClient

func init() {
	K8s = &K8sClientGetter{}
}

// GetCoreClient returns a core client
func (k *K8sClientGetter) GetCoreClient(context string) (typed_core.CoreV1Interface, error) {
	client, err := kapi.Client(context)
	if err != nil {
		return nil, errors.Wrap(err, "client")
	}
	return client.CoreV1(), nil
}

// SvcURL represents a service URL. Each item in the URLs field combines the service URL with one of the configured
// node ports. The PortNames field contains the configured names of the ports in the URLs field (sorted correspondingly -
// first item in PortNames belongs to the first item in URLs).
type SvcURL struct {
	Namespace string
	Name      string
	URLs      []string
	PortNames []string
}

// URLs represents a list of URL
type URLs []SvcURL

// GetServiceURLs returns a SvcURL object for every service in a particular namespace.
// Accepts a template for formatting
func GetServiceURLs(api libmachine.API, cname string, namespace string, t *template.Template) (URLs, error) {
	host, err := machine.LoadHost(api, cname)
	if err != nil {
		return nil, err
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, err
	}

	client, err := K8s.GetCoreClient(cname)
	if err != nil {
		return nil, err
	}

	serviceInterface := client.Services(namespace)

	svcs, err := serviceInterface.List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	var serviceURLs []SvcURL
	for _, svc := range svcs.Items {
		svcURL, err := printURLsForService(client, ip, svc.Name, svc.Namespace, t)
		if err != nil {
			return nil, err
		}
		serviceURLs = append(serviceURLs, svcURL)
	}

	return serviceURLs, nil
}

// GetServiceURLsForService returns a SvcURL object for a service in a namespace. Supports optional formatting.
func GetServiceURLsForService(api libmachine.API, cname string, namespace, service string, t *template.Template) (SvcURL, error) {
	host, err := machine.LoadHost(api, cname)
	if err != nil {
		return SvcURL{}, errors.Wrap(err, "Error checking if api exist and loading it")
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return SvcURL{}, errors.Wrap(err, "Error getting ip from host")
	}

	client, err := K8s.GetCoreClient(cname)
	if err != nil {
		return SvcURL{}, err
	}

	return printURLsForService(client, ip, service, namespace, t)
}

func printURLsForService(c typed_core.CoreV1Interface, ip, service, namespace string, t *template.Template) (SvcURL, error) {
	if t == nil {
		return SvcURL{}, errors.New("Error, attempted to generate service url with nil --format template")
	}

	svc, err := c.Services(namespace).Get(service, meta.GetOptions{})
	if err != nil {
		return SvcURL{}, errors.Wrapf(err, "service '%s' could not be found running", service)
	}

	endpoints, err := c.Endpoints(namespace).Get(service, meta.GetOptions{})
	m := make(map[int32]string)
	if err == nil && endpoints != nil && len(endpoints.Subsets) > 0 {
		for _, ept := range endpoints.Subsets {
			for _, p := range ept.Ports {
				m[p.Port] = p.Name
			}
		}
	}

	urls := []string{}
	portNames := []string{}
	for _, port := range svc.Spec.Ports {

		if port.Name != "" {
			m[port.TargetPort.IntVal] = fmt.Sprintf("%s/%d", port.Name, port.Port)
		} else {
			m[port.TargetPort.IntVal] = strconv.Itoa(int(port.Port))
		}

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
				return SvcURL{}, err
			}
			urls = append(urls, doc.String())
			portNames = append(portNames, m[port.TargetPort.IntVal])
		}
	}
	return SvcURL{Namespace: svc.Namespace, Name: svc.Name, URLs: urls, PortNames: portNames}, nil
}

// CheckService checks if a service is listening on a port.
func CheckService(cname string, namespace string, service string) error {
	client, err := K8s.GetCoreClient(cname)
	if err != nil {
		return errors.Wrap(err, "Error getting Kubernetes client")
	}

	svc, err := client.Services(namespace).Get(service, meta.GetOptions{})
	if err != nil {
		return &retry.RetriableError{
			Err: errors.Wrapf(err, "Error getting service %s", service),
		}
	}
	if len(svc.Spec.Ports) == 0 {
		return fmt.Errorf("%s:%s has no ports", namespace, service)
	}
	klog.Infof("Found service: %+v", svc)
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

// PrintServiceList prints a list of services as a table which has
// "Namespace", "Name" and "URL" columns to a writer
func PrintServiceList(writer io.Writer, data [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Namespace", "Name", "Target Port", "URL"})
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()
}

// SVCNotFoundError error type handles 'service not found' scenarios
type SVCNotFoundError struct {
	Err error
}

// Error method for SVCNotFoundError type
func (t SVCNotFoundError) Error() string {
	return "Service not found"
}

// WaitForService waits for a service, and return the urls when available
func WaitForService(api libmachine.API, cname string, namespace string, service string, urlTemplate *template.Template, urlMode bool, https bool,
	wait int, interval int) ([]string, error) {
	var urlList []string
	// Convert "Amount of time to wait" and "interval of each check" to attempts
	if interval == 0 {
		interval = 1
	}

	err := CheckService(cname, namespace, service)
	if err != nil {
		return nil, &SVCNotFoundError{err}
	}

	chkSVC := func() error { return CheckService(cname, namespace, service) }

	if err := retry.Expo(chkSVC, time.Duration(interval)*time.Second, time.Duration(wait)*time.Second); err != nil {
		return nil, &SVCNotFoundError{err}
	}

	serviceURL, err := GetServiceURLsForService(api, cname, namespace, service, urlTemplate)
	if err != nil {
		return urlList, errors.Wrap(err, "Check that minikube is running and that you have specified the correct namespace")
	}

	if !urlMode {
		var data [][]string
		if len(serviceURL.URLs) == 0 {
			data = append(data, []string{namespace, service, "", "No node port"})
		} else {
			data = append(data, []string{namespace, service, strings.Join(serviceURL.PortNames, "\n"), strings.Join(serviceURL.URLs, "\n")})
		}
		PrintServiceList(os.Stdout, data)
	}

	if len(serviceURL.URLs) == 0 {
		out.Step(style.Sad, "service {{.namespace_name}}/{{.service_name}} has no node port", false, out.V{"namespace_name": namespace, "service_name": service})
		return urlList, nil
	}

	for _, bareURLString := range serviceURL.URLs {
		url, _ := OptionallyHTTPSFormattedURLString(bareURLString, https)
		urlList = append(urlList, url)
	}
	return urlList, nil
}

// GetServiceListByLabel returns a ServiceList by label
func GetServiceListByLabel(cname string, namespace string, key string, value string) (*core.ServiceList, error) {
	client, err := K8s.GetCoreClient(cname)
	if err != nil {
		return &core.ServiceList{}, &retry.RetriableError{Err: err}
	}
	return getServiceListFromServicesByLabel(client.Services(namespace), key, value)
}

func getServiceListFromServicesByLabel(services typed_core.ServiceInterface, key string, value string) (*core.ServiceList, error) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{key: value}))
	serviceList, err := services.List(meta.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return &core.ServiceList{}, &retry.RetriableError{Err: err}
	}

	return serviceList, nil
}

// CreateSecret creates or modifies secrets
func CreateSecret(cname string, namespace, name string, dataValues map[string]string, labels map[string]string) error {
	client, err := K8s.GetCoreClient(cname)
	if err != nil {
		return &retry.RetriableError{Err: err}
	}

	secrets := client.Secrets(namespace)
	secret, err := secrets.Get(name, meta.GetOptions{})
	if err != nil {
		klog.Infof("Failed to retrieve existing secret: %v", err)
	}

	// Delete existing secret
	if len(secret.Name) > 0 {
		err = DeleteSecret(cname, namespace, name)
		if err != nil {
			return &retry.RetriableError{Err: err}
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
		return &retry.RetriableError{Err: err}
	}

	return nil
}

// DeleteSecret deletes a secret from a namespace
func DeleteSecret(cname string, namespace, name string) error {
	client, err := K8s.GetCoreClient(cname)
	if err != nil {
		return &retry.RetriableError{Err: err}
	}

	secrets := client.Secrets(namespace)
	err = secrets.Delete(name, &meta.DeleteOptions{})
	if err != nil {
		return &retry.RetriableError{Err: err}
	}

	return nil
}
