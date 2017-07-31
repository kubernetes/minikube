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
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/tests"
)

type MockClientGetter struct {
	servicesMap map[string]corev1.ServiceInterface
}

func (m *MockClientGetter) GetCoreClient() (corev1.CoreV1Interface, error) {
	return &MockCoreClient{
		servicesMap: m.servicesMap,
	}, nil
}

type MockCoreClient struct {
	fake.FakeCoreV1
	servicesMap map[string]corev1.ServiceInterface
}

var serviceNamespaces = map[string]corev1.ServiceInterface{
	"default": defaultNamespaceServiceInterface,
}

var defaultNamespaceServiceInterface = &MockServiceInterface{
	ServiceList: &v1.ServiceList{
		Items: []v1.Service{
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      "mock-dashboard",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{NodePort: int32(1111)},
						{NodePort: int32(2222)},
					},
				},
			},
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      "mock-dashboard-no-ports",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{},
				},
			},
		},
	},
}

func (m *MockCoreClient) Endpoints(namespace string) corev1.EndpointsInterface {
	return &MockEndpointsInterface{}
}

func (m *MockCoreClient) Services(namespace string) corev1.ServiceInterface {
	return m.servicesMap[namespace]
}

type MockEndpointsInterface struct {
	fake.FakeEndpoints
	Endpoints *v1.Endpoints
}

var endpointMap = map[string]*v1.Endpoints{
	"no-subsets": {},
	"not-ready": {
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{},
				NotReadyAddresses: []v1.EndpointAddress{
					{IP: "1.1.1.1"},
					{IP: "2.2.2.2"},
				},
			},
		},
	},
	"one-ready": {
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{IP: "1.1.1.1"},
				},
				NotReadyAddresses: []v1.EndpointAddress{
					{IP: "2.2.2.2"},
				},
			},
		},
	},
}

func (e MockEndpointsInterface) Get(name string, _ meta_v1.GetOptions) (*v1.Endpoints, error) {
	endpoint, ok := endpointMap[name]
	if !ok {
		return nil, errors.New("Endpoint not found")
	}
	return endpoint, nil
}

func TestCheckEndpointReady(t *testing.T) {
	var tests = []struct {
		description string
		service     string
		err         bool
	}{
		{
			description: "Endpoint with no subsets should return an error",
			service:     "no-subsets",
			err:         true,
		},
		{
			description: "Endpoint with no ready endpoints should return an error",
			service:     "not-ready",
			err:         true,
		},
		{
			description: "Endpoint with at least one ready endpoint should not return an error",
			service:     "one-ready",
			err:         false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			err := checkEndpointReady(&MockEndpointsInterface{}, test.service)
			if err != nil && !test.err {
				t.Errorf("Check endpoints returned an error: %+v", err)
			}
			if err == nil && test.err {
				t.Errorf("Check endpoints should have returned an error but returned nil")
			}
		})
	}
}

type MockServiceInterface struct {
	fake.FakeServices
	ServiceList *v1.ServiceList
}

func (s MockServiceInterface) List(opts meta_v1.ListOptions) (*v1.ServiceList, error) {
	serviceList := &v1.ServiceList{
		Items: []v1.Service{},
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

func (s MockServiceInterface) Get(name string, _ meta_v1.GetOptions) (*v1.Service, error) {
	for _, svc := range s.ServiceList.Items {
		if svc.ObjectMeta.Name == name {
			return &svc, nil
		}
	}

	return nil, nil
}

func TestGetServiceListFromServicesByLabel(t *testing.T) {
	serviceList := &v1.ServiceList{
		Items: []v1.Service{
			{
				Spec: v1.ServiceSpec{
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
		servicesMap: serviceNamespaces,
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
			urls, err := printURLsForService(client, "127.0.0.1", test.serviceName, test.namespace, test.tmpl)
			if err != nil && !test.err {
				t.Errorf("Error: %s", err)
			}
			if err == nil && test.err {
				t.Errorf("Expected error but got none")
			}
			if !reflect.DeepEqual(urls, test.expectedOutput) {
				t.Errorf("\nExpected %v \nActual: %v \n\n", test.expectedOutput, urls)
			}
		})
	}
}

func TestGetServiceURLs(t *testing.T) {
	defaultAPI := &tests.MockAPI{
		Hosts: map[string]*host.Host{
			config.GetMachineName(): {
				Name:   config.GetMachineName(),
				Driver: &tests.MockDriver{},
			},
		},
	}
	defaultTemplate := template.Must(template.New("svc-template").Parse("http://{{.IP}}:{{.Port}}"))

	var tests = []struct {
		description string
		api         libmachine.API
		namespace   string
		expected    ServiceURLs
		err         bool
	}{
		{
			description: "no host",
			api: &tests.MockAPI{
				Hosts: make(map[string]*host.Host),
			},
			err: true,
		},
		{
			description: "correctly return serviceURLs",
			namespace:   "default",
			api:         defaultAPI,
			expected: []ServiceURL{
				{
					Namespace: "default",
					Name:      "mock-dashboard",
					URLs:      []string{"http://127.0.0.1:1111", "http://127.0.0.1:2222"},
				},
				{
					Namespace: "default",
					Name:      "mock-dashboard-no-ports",
					URLs:      []string{},
				},
			},
		},
	}

	defer revertK8sClient(k8s)
	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()

			k8s = &MockClientGetter{
				servicesMap: serviceNamespaces,
			}
			urls, err := GetServiceURLs(test.api, test.namespace, defaultTemplate)
			if err != nil && !test.err {
				t.Errorf("Error GetServiceURLs %s", err)
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
		Hosts: map[string]*host.Host{
			config.GetMachineName(): {
				Name:   config.GetMachineName(),
				Driver: &tests.MockDriver{},
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
		err         bool
	}{
		{
			description: "no host",
			api: &tests.MockAPI{
				Hosts: make(map[string]*host.Host),
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

	defer revertK8sClient(k8s)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			k8s = &MockClientGetter{
				servicesMap: serviceNamespaces,
			}
			urls, err := GetServiceURLsForService(test.api, test.namespace, test.service, defaultTemplate)
			if err != nil && !test.err {
				t.Errorf("Error GetServiceURLsForService %s", err)
			}
			if err == nil && test.err {
				t.Errorf("Test should have failed, but didn't")
			}
			if !reflect.DeepEqual(urls, test.expected) {
				t.Errorf("URLs did not match, expected %+v \n\n got %+v", test.expected, urls)
			}
		})
	}
}

func revertK8sClient(k K8sClient) {
	k8s = k
}
