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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1/fake"
	testing_fake "k8s.io/client-go/testing"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/tests"
)

// Mock Kubernetes client getter - NOT THREAD SAFE
type MockClientGetter struct {
	servicesMap  map[string]typed_core.ServiceInterface
	endpointsMap map[string]typed_core.EndpointsInterface
	secretsMap   map[string]typed_core.SecretInterface
	Fake         fake.FakeCoreV1
}

// Force GetCoreClient to fail
var getCoreClientFail bool

func (m *MockClientGetter) GetCoreClient(string) (typed_core.CoreV1Interface, error) {
	if getCoreClientFail {
		return nil, fmt.Errorf("test Error - Mocked Get")
	}
	return &MockCoreClient{
		servicesMap:  m.servicesMap,
		endpointsMap: m.endpointsMap,
		secretsMap:   m.secretsMap}, nil
}

func (m *MockCoreClient) Secrets(ns string) typed_core.SecretInterface {
	return &fake.FakeSecrets{Fake: &fake.FakeCoreV1{Fake: &testing_fake.Fake{}}}
}

func (m *MockCoreClient) Services(namespace string) typed_core.ServiceInterface {
	return m.servicesMap[namespace]
}

// Mock Kubernetes client - NOT THREAD SAFE
type MockCoreClient struct {
	fake.FakeCoreV1
	servicesMap  map[string]typed_core.ServiceInterface
	endpointsMap map[string]typed_core.EndpointsInterface
	secretsMap   map[string]typed_core.SecretInterface
}

var secretsNamespaces = map[string]typed_core.SecretInterface{
	"default": defaultNamespaceSecretsInterface,
}

var defaultNamespaceSecretsInterface = &MockSecretInterface{
	SecretsList: &core.SecretList{
		Items: []core.Secret{
			{},
		},
	},
}

var serviceNamespaces = map[string]typed_core.ServiceInterface{
	"default": defaultNamespaceServiceInterface,
}

var serviceNamespaceOther = map[string]typed_core.ServiceInterface{
	"default": nondefaultNamespaceServiceInterface,
}

var nondefaultNamespaceServiceInterface = &MockServiceInterface{
	ServiceList: &core.ServiceList{
		Items: []core.Service{
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "non-namespace-dashboard-no-ports",
					Namespace: "cannot_be_found_namespace",
					Labels:    map[string]string{"mock": "mock"},
				},
				Spec: core.ServiceSpec{
					Ports: []core.ServicePort{},
				},
			},
		},
	},
}

var defaultNamespaceServiceInterface = &MockServiceInterface{
	ServiceList: &core.ServiceList{
		Items: []core.Service{
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "mock-dashboard",
					Namespace: "default",
					Labels:    map[string]string{"mock": "mock"},
				},
				Spec: core.ServiceSpec{
					Ports: []core.ServicePort{
						{
							Name:     "port1",
							NodePort: int32(1111),
							Port:     int32(11111),
							TargetPort: intstr.IntOrString{
								IntVal: int32(11111),
							},
						},
						{
							Name:     "port2",
							NodePort: int32(2222),
							Port:     int32(22222),
							TargetPort: intstr.IntOrString{
								IntVal: int32(22222),
							},
						},
					},
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "mock-dashboard-no-ports",
					Namespace: "default",
					Labels:    map[string]string{"mock": "mock"},
				},
				Spec: core.ServiceSpec{
					Ports: []core.ServicePort{},
				},
			},
		},
	},
}

var endpointNamespaces = map[string]typed_core.EndpointsInterface{
	"default": defaultNamespaceEndpointInterface,
}

var defaultNamespaceEndpointInterface = &MockEndpointsInterface{}

func (m *MockCoreClient) Endpoints(namespace string) typed_core.EndpointsInterface {
	return m.endpointsMap[namespace]
}

type MockEndpointsInterface struct {
	fake.FakeEndpoints
	Endpoints *core.Endpoints
}

var endpointMap = map[string]*core.Endpoints{
	"no-subsets": {},
	"not-ready": {
		Subsets: []core.EndpointSubset{
			{
				Addresses: []core.EndpointAddress{},
				NotReadyAddresses: []core.EndpointAddress{
					{IP: "1.1.1.1"},
					{IP: "2.2.2.2"},
				},
			},
		},
	},
	"one-ready": {
		Subsets: []core.EndpointSubset{
			{
				Addresses: []core.EndpointAddress{
					{IP: "1.1.1.1"},
				},
				NotReadyAddresses: []core.EndpointAddress{
					{IP: "2.2.2.2"},
				},
			},
		},
	},
	"mock-dashboard": {
		Subsets: []core.EndpointSubset{
			{
				Ports: []core.EndpointPort{
					{
						Name: "port1",
						Port: int32(11111),
					},
					{
						Name: "port2",
						Port: int32(22222),
					},
				},
			},
		},
	},
}

func (e MockEndpointsInterface) Get(ctx context.Context, name string, _ meta.GetOptions) (*core.Endpoints, error) {
	endpoint, ok := endpointMap[name]
	if !ok {
		return nil, errors.New("Endpoint not found")
	}
	return endpoint, nil
}

type MockServiceInterface struct {
	fake.FakeServices
	ServiceList *core.ServiceList
}

type MockSecretInterface struct {
	fake.FakeSecrets
	SecretsList *core.SecretList
}

func (s MockServiceInterface) List(ctx context.Context, opts meta.ListOptions) (*core.ServiceList, error) {
	serviceList := &core.ServiceList{
		Items: []core.Service{},
	}
	if opts.LabelSelector != "" {
		keyValArr := strings.Split(opts.LabelSelector, "=")

		for _, service := range s.ServiceList.Items {
			if service.Spec.Selector[keyValArr[0]] == keyValArr[1] {
				serviceList.Items = append(serviceList.Items, service)
			}
		}

		return serviceList, nil
	}

	return s.ServiceList, nil
}

func (s MockServiceInterface) Get(ctx context.Context, name string, _ meta.GetOptions) (*core.Service, error) {
	for _, svc := range s.ServiceList.Items {
		if svc.ObjectMeta.Name == name {
			return &svc, nil
		}
	}

	return nil, nil
}

func TestGetServiceListFromServicesByLabel(t *testing.T) {
	serviceList := &core.ServiceList{
		Items: []core.Service{
			{
				Spec: core.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}
	serviceIface := MockServiceInterface{
		ServiceList: serviceList,
	}
	if _, err := getServiceListFromServicesByLabel(&serviceIface, "nothing", "nothing"); err != nil {
		t.Fatalf("Service had no label match, but getServiceListFromServicesByLabel returned an error")
	}

	if _, err := getServiceListFromServicesByLabel(&serviceIface, "foo", "bar"); err != nil {
		t.Fatalf("Endpoint was ready with at least one Address, but getServiceListFromServicesByLabel returned an error")
	}
}

func TestPrintURLsForService(t *testing.T) {
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))
	client := &MockCoreClient{
		servicesMap:  serviceNamespaces,
		endpointsMap: endpointNamespaces,
	}
	var tests = []struct {
		description    string
		serviceName    string
		namespace      string
		tmpl           *template.Template
		expectedOutput []string
		err            bool
	}{
		{
			description:    "should get all node ports",
			serviceName:    "mock-dashboard",
			namespace:      "default",
			tmpl:           defaultTemplate,
			expectedOutput: []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
		},
		{
			description:    "should get all node ports with arbitrary format",
			serviceName:    "mock-dashboard",
			namespace:      "default",
			tmpl:           template.Must(template.New("svc-arbitrary-template").Parse("{{.IP}}:{{.Port}}")),
			expectedOutput: []string{"127.0.0.1:1111", "127.0.0.1:2222"},
		},
		{
			description:    "should get the name of all target ports with arbitrary format",
			serviceName:    "mock-dashboard",
			namespace:      "default",
			tmpl:           template.Must(template.New("svc-arbitrary-template").Parse("{{.Name}}={{.IP}}:{{.Port}}")),
			expectedOutput: []string{"port1/11111=127.0.0.1:1111", "port2/22222=127.0.0.1:2222"},
		},
		{
			description:    "empty slice for no node ports",
			serviceName:    "mock-dashboard-no-ports",
			namespace:      "default",
			tmpl:           defaultTemplate,
			expectedOutput: []string{},
		},
		{
			description: "throw error without template",
			err:         true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			svcURL, err := printURLsForService(client, "127.0.0.1", test.serviceName, test.namespace, test.tmpl)
			if err != nil && !test.err {
				t.Errorf("Error: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none")
			}
			if !reflect.DeepEqual(svcURL.URLs, test.expectedOutput) {
				t.Errorf("\nExpected %v \nActual: %v \n\n", test.expectedOutput, svcURL.URLs)
			}
		})
	}
}

func TestOptionallyHttpsFormattedUrlString(t *testing.T) {

	var tests = []struct {
		description                     string
		bareURLString                   string
		https                           bool
		expectedHTTPSFormattedURLString string
		expectedIsHTTPSchemedURL        bool
	}{
		{
			description:                     "no https for http schemed with no https option",
			bareURLString:                   "http://192.168.99.100:30563",
			https:                           false,
			expectedHTTPSFormattedURLString: "http://192.168.99.100:30563",
			expectedIsHTTPSchemedURL:        true,
		},
		{
			description:                     "no https for non-http schemed with no https option",
			bareURLString:                   "xyz.http.myservice:30563",
			https:                           false,
			expectedHTTPSFormattedURLString: "xyz.http.myservice:30563",
			expectedIsHTTPSchemedURL:        false,
		},
		{
			description:                     "https for http schemed with https option",
			bareURLString:                   "http://192.168.99.100:30563",
			https:                           true,
			expectedHTTPSFormattedURLString: "https://192.168.99.100:30563",
			expectedIsHTTPSchemedURL:        true,
		},
		{
			description:                     "no https for non-http schemed with https option and http substring",
			bareURLString:                   "xyz.http.myservice:30563",
			https:                           true,
			expectedHTTPSFormattedURLString: "xyz.http.myservice:30563",
			expectedIsHTTPSchemedURL:        false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			httpsFormattedURLString, isHTTPSchemedURL := OptionallyHTTPSFormattedURLString(test.bareURLString, test.https)

			if httpsFormattedURLString != test.expectedHTTPSFormattedURLString {
				t.Errorf("\nhttpsFormattedURLString, Expected %v \nActual: %v \n\n", test.expectedHTTPSFormattedURLString, httpsFormattedURLString)
			}

			if isHTTPSchemedURL != test.expectedIsHTTPSchemedURL {
				t.Errorf("\nisHTTPSchemedURL, Expected %v \nActual: %v \n\n",
					test.expectedHTTPSFormattedURLString, httpsFormattedURLString)
			}
		})
	}
}

func TestGetServiceURLs(t *testing.T) {
	defaultAPI := &tests.MockAPI{
		FakeStore: tests.FakeStore{
			Hosts: map[string]*host.Host{
				constants.DefaultClusterName: {
					Name:   constants.DefaultClusterName,
					Driver: &tests.MockDriver{},
				},
			},
		},
	}
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))
	viper.Set(config.ProfileName, constants.DefaultClusterName)

	var tests = []struct {
		description string
		api         libmachine.API
		namespace   string
		expected    URLs
		err         bool
	}{
		{
			description: "no host",
			api: &tests.MockAPI{
				FakeStore: tests.FakeStore{
					Hosts: make(map[string]*host.Host),
				},
			},
			err: true,
		},
		{
			description: "correctly return serviceURLs",
			namespace:   "default",
			api:         defaultAPI,
			expected: []SvcURL{
				{
					Namespace: "default",
					Name:      "mock-dashboard",
					URLs:      []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
					PortNames: []string{"port1/11111", "port2/22222"},
				},
				{
					Namespace: "default",
					Name:      "mock-dashboard-no-ports",
					URLs:      []string{},
					PortNames: []string{},
				},
			},
		},
	}

	defer revertK8sClient(K8s)
	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaces,
				endpointsMap: endpointNamespaces,
			}
			urls, err := GetServiceURLs(test.api, "minikube", test.namespace, defaultTemplate)
			if err != nil && !test.err {
				t.Errorf("Error GetServiceURLs %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Test should have failed, but didn't")
			}
			if !reflect.DeepEqual(urls, test.expected) {
				t.Errorf("URLs did not match, expected %v \n\n got %v", test.expected, urls)
			}
		})
	}
}

func TestGetServiceURLsForService(t *testing.T) {
	defaultAPI := &tests.MockAPI{
		FakeStore: tests.FakeStore{
			Hosts: map[string]*host.Host{
				constants.DefaultClusterName: {
					Name:   constants.DefaultClusterName,
					Driver: &tests.MockDriver{},
				},
			},
		},
	}
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))
	viper.Set(config.ProfileName, constants.DefaultClusterName)

	var tests = []struct {
		description string
		api         libmachine.API
		namespace   string
		service     string
		expected    []string
		err         bool
	}{
		{
			description: "no host",
			api: &tests.MockAPI{
				FakeStore: tests.FakeStore{
					Hosts: make(map[string]*host.Host),
				},
			},
			err: true,
		},
		{
			description: "correctly return serviceURLs",
			namespace:   "default",
			service:     "mock-dashboard",
			api:         defaultAPI,
			expected:    []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
		},
		{
			description: "correctly return empty serviceURLs",
			namespace:   "default",
			service:     "mock-dashboard-no-ports",
			api:         defaultAPI,
			expected:    []string{},
		},
	}

	defer revertK8sClient(K8s)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaces,
				endpointsMap: endpointNamespaces,
			}
			svcURL, err := GetServiceURLsForService(test.api, "minikube", test.namespace, test.service, defaultTemplate)
			if err != nil && !test.err {
				t.Errorf("Error GetServiceURLsForService %v", err)
			}
			if err == nil && test.err {
				t.Errorf("Test should have failed, but didn't")
			}
			if !reflect.DeepEqual(svcURL.URLs, test.expected) {
				t.Errorf("URLs did not match, expected %+v \n\n got %+v", test.expected, svcURL.URLs)
			}
		})
	}
}

func revertK8sClient(k K8sClient) {
	K8s = k
	getCoreClientFail = false
}

func TestGetCoreClient(t *testing.T) {
	originalEnv := os.Getenv("KUBECONFIG")
	defer func() {
		err := os.Setenv("KUBECONFIG", originalEnv)
		if err != nil {
			t.Fatalf("Error reverting env KUBECONFIG to its original value. Got err (%s)", err)
		}
	}()

	mockK8sConfig := `apiVersion: v1
clusters:
- cluster:
    server: https://192.168.99.102:8443
  name: minikube
contexts:
- context:
    cluster: minikube
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
- name: minikube
`
	var tests = []struct {
		description    string
		kubeconfigPath string
		config         string
		err            bool
	}{
		{
			description:    "ok",
			kubeconfigPath: "/tmp/kube_config",
			config:         mockK8sConfig,
			err:            false,
		},
		{
			description:    "empty config",
			kubeconfigPath: "/tmp/kube_config",
			config:         "",
			err:            true,
		},
		{
			description:    "broken config",
			kubeconfigPath: "/tmp/kube_config",
			config:         "this**is&&not: yaml::valid: file",
			err:            true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mockK8sConfigByte := []byte(test.config)
			mockK8sConfigPath := test.kubeconfigPath
			err := ioutil.WriteFile(mockK8sConfigPath, mockK8sConfigByte, 0644)
			defer os.Remove(mockK8sConfigPath)
			if err != nil {
				t.Fatalf("Unexpected error when writing to file %v. Error: %v", test.kubeconfigPath, err)
			}
			os.Setenv("KUBECONFIG", mockK8sConfigPath)

			k8s := K8sClientGetter{}
			_, err = k8s.GetCoreClient("minikube")
			if err != nil && !test.err {
				t.Fatalf("GetCoreClient returned unexpected error: %v", err)
			}
			if err == nil && test.err {
				t.Fatal("GetCoreClient expected to return error but got nil")
			}
		})
	}
}

func TestPrintServiceList(t *testing.T) {
	var buf bytes.Buffer
	out := &buf
	input := [][]string{{"foo", "bar", "baz", "nah"}}
	PrintServiceList(out, input)
	expected := `|-----------|------|-------------|-----|
| NAMESPACE | NAME | TARGET PORT | URL |
|-----------|------|-------------|-----|
| foo       | bar  | baz         | nah |
|-----------|------|-------------|-----|
`
	got := out.String()
	if got != expected {
		t.Fatalf("PrintServiceList(%v) expected to return %v but got \n%v", input, expected, got)
	}
}

func TestGetServiceListByLabel(t *testing.T) {
	var tests = []struct {
		description, ns, name, label string
		items                        int
		failedGetClient, err         bool
	}{
		{
			description: "ok",
			name:        "mock-dashboard",
			ns:          "default",
			items:       2,
		},
		{
			description:     "failed get client",
			name:            "mock-dashboard",
			ns:              "default",
			failedGetClient: true,
			err:             true,
		},
		{
			description: "no matches",
			name:        "mock-dashboard-no-ports",
			ns:          "default",
			label:       "foo",
			items:       0,
		},
	}

	defer revertK8sClient(K8s)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaces,
				endpointsMap: endpointNamespaces,
				secretsMap:   secretsNamespaces,
			}
			getCoreClientFail = test.failedGetClient
			svcs, err := GetServiceListByLabel("minikube", test.ns, test.name, test.label)
			if err != nil && !test.err {
				t.Fatalf("Test %v got unexpected error: %v", test.description, err)
			}
			if err == nil && test.err {
				t.Fatalf("Test %v expected error but got nil", test.description)
			}
			if err == nil {
				if len(svcs.Items) != test.items {
					t.Fatalf("GetServiceListByLabel for test: %v data should return %d elements, but got: %d", test.description, test.items, len(svcs.Items))
				}
			}
		})
	}
}

func TestCheckService(t *testing.T) {
	var tests = []struct {
		description, ns, name string
		failedGetClient, err  bool
	}{
		{
			description: "ok",
			name:        "mock-dashboard",
			ns:          "default",
		},
		{
			description:     "failed get client",
			name:            "mock-dashboard",
			ns:              "default",
			failedGetClient: true,
			err:             true,
		},
		{
			description: "svc no ports",
			name:        "mock-dashboard-no-ports",
			ns:          "default",
			err:         true,
		},
	}

	defer revertK8sClient(K8s)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaces,
				endpointsMap: endpointNamespaces,
				secretsMap:   secretsNamespaces,
			}
			getCoreClientFail = test.failedGetClient
			err := CheckService("minikube", test.ns, test.name)
			if err == nil && test.err {
				t.Fatalf("Test %v expected error but got nil", test.description)
			}
			if err != nil && !test.err {
				t.Fatalf("Test %v got unexpected error: %v", test.description, err)
			}
		})
	}
}

func TestDeleteSecret(t *testing.T) {
	var tests = []struct {
		description, ns, name string
		failedGetClient, err  bool
	}{
		{
			description: "ok",
			name:        "foo",
			ns:          "foo",
		},
		{
			description:     "failed get client",
			name:            "foo",
			ns:              "foo",
			failedGetClient: true,
			err:             true,
		},
	}

	defer revertK8sClient(K8s)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaces,
				endpointsMap: endpointNamespaces,
				secretsMap:   secretsNamespaces,
			}
			getCoreClientFail = test.failedGetClient
			err := DeleteSecret("minikube", test.ns, test.name)
			if err == nil && test.err {
				t.Fatalf("Test %v expected error but got nil", test.description)
			}
			if err != nil && !test.err {
				t.Fatalf("Test %v got unexpected error: %v", test.description, err)
			}
		})
	}
}

func TestCreateSecret(t *testing.T) {
	var tests = []struct {
		description, ns, name string
		failedGetClient, err  bool
	}{
		{
			description: "ok",
			name:        "foo",
			ns:          "foo",
		},
		{
			description:     "failed get client",
			name:            "foo",
			ns:              "foo",
			failedGetClient: true,
			err:             true,
		},
	}

	defer revertK8sClient(K8s)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaces,
				endpointsMap: endpointNamespaces,
				secretsMap:   secretsNamespaces,
			}
			getCoreClientFail = test.failedGetClient
			err := CreateSecret("minikube", test.ns, test.name, map[string]string{"ns": "secret"}, map[string]string{"ns": "baz"})
			if err == nil && test.err {
				t.Fatalf("Test %v expected error but got nil", test.description)
			}
			if err != nil && !test.err {
				t.Fatalf("Test %v got unexpected error: %v", test.description, err)
			}
		})
	}
}

func TestWaitAndMaybeOpenService(t *testing.T) {
	defaultAPI := &tests.MockAPI{
		FakeStore: tests.FakeStore{
			Hosts: map[string]*host.Host{
				constants.DefaultClusterName: {
					Name:   constants.DefaultClusterName,
					Driver: &tests.MockDriver{},
				},
			},
		},
	}
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))

	var tests = []struct {
		description string
		api         libmachine.API
		namespace   string
		service     string
		expected    []string
		urlMode     bool
		https       bool
		err         bool
	}{
		/*	{
			description: "no host",
			api: &tests.mockapi{
				fakestore: tests.fakestore{
					hosts: make(map[string]*host.host),
				},
			},
			err: true,
		}, */
		{
			description: "correctly return serviceURLs, https, no url mode",
			namespace:   "default",
			service:     "mock-dashboard",
			api:         defaultAPI,
			https:       true,
			expected:    []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
		},
		{
			description: "correctly return serviceURLs, no https, no url mode",
			namespace:   "default",
			service:     "mock-dashboard",
			api:         defaultAPI,
			expected:    []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
		},
		{
			description: "correctly return serviceURLs, no https, url mode",
			namespace:   "default",
			service:     "mock-dashboard",
			api:         defaultAPI,
			urlMode:     true,
			expected:    []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
		},
		{
			description: "correctly return serviceURLs, https, url mode",
			namespace:   "default",
			service:     "mock-dashboard",
			api:         defaultAPI,
			urlMode:     true,
			https:       true,
			expected:    []string{"https://127.0.0.1:1111", "https://127.0.0.1:2222"},
		},
		{
			description: "correctly return serviceURLs, http, url mode",
			namespace:   "default",
			service:     "mock-dashboard",
			api:         defaultAPI,
			urlMode:     true,
			https:       false,
			expected:    []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
		},
		{
			description: "correctly return empty serviceURLs",
			namespace:   "default",
			service:     "mock-dashboard-no-ports",
			api:         defaultAPI,
			expected:    []string{},
			err:         true,
		},
	}
	defer revertK8sClient(K8s)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaces,
				endpointsMap: endpointNamespaces,
			}

			var urlList []string
			urlList, err := WaitForService(test.api, "minikube", test.namespace, test.service, defaultTemplate, test.urlMode, test.https, 1, 0)
			if test.err && err == nil {
				t.Fatalf("WaitForService expected to fail for test: %v", test)
			}
			if !test.err && err != nil {
				t.Fatalf("WaitForService not expected to fail but got err: %v", err)
			}

			if test.urlMode {
				// check the size of the url slices
				if len(urlList) != len(test.expected) {
					t.Fatalf("WaitForService returned [%d] urls while expected is [%d] url", len(urlList), len(test.expected))
				}

				// check content of the expected url
				for i, v := range test.expected {
					if v != urlList[i] {
						t.Fatalf("WaitForService returned [%s] urls while expected is [%s] url", urlList[i], v)
					}
				}
			}

		})
	}
}

func TestWaitAndMaybeOpenServiceForNotDefaultNamspace(t *testing.T) {
	defaultAPI := &tests.MockAPI{
		FakeStore: tests.FakeStore{
			Hosts: map[string]*host.Host{
				constants.DefaultClusterName: {
					Name:   constants.DefaultClusterName,
					Driver: &tests.MockDriver{},
				},
			},
		},
	}
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))

	var tests = []struct {
		description string
		api         libmachine.API
		namespace   string
		service     string
		expected    []string
		urlMode     bool
		https       bool
		err         bool
	}{
		{
			description: "correctly return empty serviceURLs",
			namespace:   "default",
			service:     "non-namespace-dashboard-no-ports",
			api:         defaultAPI,
			expected:    []string{},
			err:         true,
		},
	}
	defer revertK8sClient(K8s)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			K8s = &MockClientGetter{
				servicesMap:  serviceNamespaceOther,
				endpointsMap: endpointNamespaces,
			}
			_, err := WaitForService(test.api, "minikube", test.namespace, test.service, defaultTemplate, test.urlMode, test.https, 1, 0)
			if test.err && err == nil {
				t.Fatalf("WaitForService expected to fail for test: %v", test)
			}
			if !test.err && err != nil {
				t.Fatalf("WaitForService not expected to fail but got err: %v", err)
			}

		})
	}
}
