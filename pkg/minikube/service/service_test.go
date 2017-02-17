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
	"strings"
	"testing"

	"github.com/pkg/errors"
	"k8s.io/client-go/1.5/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
)

func TestCheckEndpointReady(t *testing.T) {
	endpointNoSubsets := &v1.Endpoints{}
	if err := checkEndpointReady(endpointNoSubsets); err == nil {
		t.Fatalf("Endpoint had no subsets but checkEndpointReady did not return an error")
	}

	endpointNotReady := &v1.Endpoints{
		Subsets: []v1.EndpointSubset{
			{Addresses: []v1.EndpointAddress{},
				NotReadyAddresses: []v1.EndpointAddress{
					{IP: "1.1.1.1"},
					{IP: "2.2.2.2"},
					{IP: "3.3.3.3"},
				}}}}
	if err := checkEndpointReady(endpointNotReady); err == nil {
		t.Fatalf("Endpoint had no Addresses but checkEndpointReady did not return an error")
	}

	endpointReady := &v1.Endpoints{
		Subsets: []v1.EndpointSubset{
			{Addresses: []v1.EndpointAddress{
				{IP: "1.1.1.1"},
				{IP: "2.2.2.2"},
			},
				NotReadyAddresses: []v1.EndpointAddress{},
			}},
	}
	if err := checkEndpointReady(endpointReady); err != nil {
		t.Fatalf("Endpoint was ready with at least one Address, but checkEndpointReady returned an error")
	}
}

type ServiceInterfaceMock struct {
	fake.FakeServices
	ServiceList *v1.ServiceList
}

func (s ServiceInterfaceMock) List(opts api.ListOptions) (*v1.ServiceList, error) {
	serviceList := &v1.ServiceList{
		Items: []v1.Service{},
	}
	keyValArr := strings.Split(opts.LabelSelector.String(), "=")
	for _, service := range s.ServiceList.Items {
		if service.Spec.Selector[keyValArr[0]] == keyValArr[1] {
			serviceList.Items = append(serviceList.Items, service)
		}
	}
	return serviceList, nil
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
	serviceIface := ServiceInterfaceMock{
		ServiceList: serviceList,
	}
	if _, err := getServiceListFromServicesByLabel(&serviceIface, "nothing", "nothing"); err != nil {
		t.Fatalf("Service had no label match, but getServiceListFromServicesByLabel returned an error")
	}

	if _, err := getServiceListFromServicesByLabel(&serviceIface, "foo", "bar"); err != nil {
		t.Fatalf("Endpoint was ready with at least one Address, but getServiceListFromServicesByLabel returned an error")
	}
}

type MockServiceGetter struct {
	services map[string]v1.Service
}

func NewMockServiceGetter() *MockServiceGetter {
	return &MockServiceGetter{
		services: make(map[string]v1.Service),
	}
}

func (mockServiceGetter *MockServiceGetter) Get(name string) (*v1.Service, error) {
	service, ok := mockServiceGetter.services[name]
	if !ok {
		return nil, errors.Errorf("Error getting %s service from mockServiceGetter", name)
	}
	return &service, nil
}

func (mockServiceGetter *MockServiceGetter) List(options api.ListOptions) (*v1.ServiceList, error) {
	services := v1.ServiceList{
		TypeMeta: unversioned.TypeMeta{Kind: "ServiceList", APIVersion: "v1"},
		ListMeta: unversioned.ListMeta{},
	}

	for _, svc := range mockServiceGetter.services {
		services.Items = append(services.Items, svc)
	}
	return &services, nil
}

func TestGetServiceURLs(t *testing.T) {
	mockServiceGetter := NewMockServiceGetter()
	expected := []int32{1111, 2222}
	mockDashboardService := v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					NodePort: expected[0],
				}, {
					NodePort: expected[1],
				}},
		},
	}
	mockServiceGetter.services["mock-service"] = mockDashboardService

	ports, err := getServicePortsFromServiceGetter(mockServiceGetter, "mock-service")
	if err != nil {
		t.Fatalf("Error getting mock-service ports from api: Error: %s", err)
	}
	for i := range ports {
		if ports[i] != expected[i] {
			t.Fatalf("Error getting mock-service port from api: Expected: %d, Got: %d", ports[0], expected)
		}
	}
}

func TestGetServiceURLWithoutNodePort(t *testing.T) {
	mockServiceGetter := NewMockServiceGetter()
	mockDashboardService := v1.Service{}
	mockServiceGetter.services["mock-service"] = mockDashboardService

	_, err := getServicePortsFromServiceGetter(mockServiceGetter, "mock-service")
	if err == nil {
		t.Fatalf("Expected error getting service with no node port")
	}
}
