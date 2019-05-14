/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"testing"

	"reflect"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"
	fake "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/rest"
)

type stubCoreClient struct {
	fake.FakeCoreV1
	servicesList *core.ServiceList
	restClient   *rest.RESTClient
}

func (c *stubCoreClient) Services(namespace string) typed_core.ServiceInterface {
	return &stubServices{
		fake.FakeServices{Fake: &c.FakeCoreV1},
		c.servicesList,
	}
}

func (c *stubCoreClient) RESTClient() rest.Interface {
	return c.restClient
}

type stubServices struct {
	fake.FakeServices
	servicesList *core.ServiceList
}

func (s *stubServices) List(opts meta.ListOptions) (*core.ServiceList, error) {
	return s.servicesList, nil
}

func newStubCoreClient(servicesList *core.ServiceList) *stubCoreClient {
	if servicesList == nil {
		servicesList = &core.ServiceList{
			Items: []core.Service{}}
	}
	return &stubCoreClient{
		servicesList: servicesList,
		restClient:   nil,
	}
}

type countingRequestSender struct {
	requests int
}

func (s *countingRequestSender) send(request *rest.Request) (result []byte, err error) {
	s.requests++
	return nil, nil
}

type recordingPatchConverter struct {
	patches []*Patch
}

func (r *recordingPatchConverter) convert(restClient rest.Interface, patch *Patch) *rest.Request {
	r.patches = append(r.patches, patch)
	return nil
}

func TestEmptyListOfServicesDoesNothing(t *testing.T) {
	client := newStubCoreClient(&core.ServiceList{
		Items: []core.Service{}})

	patcher := newLoadBalancerEmulator(client)

	serviceNames, err := patcher.PatchServices()

	if len(serviceNames) > 0 || err != nil {
		t.Errorf("Expected: [], nil\n Got: %v, %s", serviceNames, err)
	}

}

func TestServicesWithNoLoadbalancerType(t *testing.T) {
	client := newStubCoreClient(&core.ServiceList{
		Items: []core.Service{
			{
				Spec: core.ServiceSpec{
					Type: "ClusterIP",
				},
			},
			{
				Spec: core.ServiceSpec{
					Type: "NodeIP",
				},
			},
		},
	})

	patcher := newLoadBalancerEmulator(client)

	serviceNames, err := patcher.PatchServices()

	if len(serviceNames) > 0 || err != nil {
		t.Errorf("Expected: [], nil\n Got: %v, %s", serviceNames, err)
	}

}

func TestServicesWithLoadbalancerType(t *testing.T) {
	client := newStubCoreClient(&core.ServiceList{
		Items: []core.Service{
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "svc1-up-to-date",
					Namespace: "ns1",
				},
				Spec: core.ServiceSpec{
					Type:      "LoadBalancer",
					ClusterIP: "10.96.0.3",
				},
				Status: core.ServiceStatus{
					LoadBalancer: core.LoadBalancerStatus{
						Ingress: []core.LoadBalancerIngress{
							{
								IP: "10.96.0.3",
							},
						},
					},
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "svc2-out-of-date",
					Namespace: "ns2",
				},
				Spec: core.ServiceSpec{
					Type:      "LoadBalancer",
					ClusterIP: "10.96.0.4",
				},
				Status: core.ServiceStatus{
					LoadBalancer: core.LoadBalancerStatus{
						Ingress: []core.LoadBalancerIngress{
							{
								IP: "10.96.0.5",
							},
						},
					},
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "svc3-empty-ingress",
					Namespace: "ns3",
				},
				Spec: core.ServiceSpec{
					Type:      "LoadBalancer",
					ClusterIP: "10.96.0.2",
				},
				Status: core.ServiceStatus{
					LoadBalancer: core.LoadBalancerStatus{
						Ingress: []core.LoadBalancerIngress{},
					},
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					Name: "svc4-not-lb",
				},
				Spec: core.ServiceSpec{
					Type: "NodeIP",
				},
			},
		},
	})

	expectedPatches := []*Patch{
		{
			Type:         "application/json-patch+json",
			NameSpace:    "ns2",
			NameSpaceSet: true,
			Resource:     "services",
			Subresource:  "status",
			ResourceName: "svc2-out-of-date",
			BodyContent:  `[{"op": "add", "path": "/status/loadBalancer/ingress", "value":  [ { "ip": "10.96.0.4" } ] }]`,
		},
		{
			Type:         "application/json-patch+json",
			NameSpace:    "ns3",
			NameSpaceSet: true,
			Resource:     "services",
			Subresource:  "status",
			ResourceName: "svc3-empty-ingress",
			BodyContent:  `[{"op": "add", "path": "/status/loadBalancer/ingress", "value":  [ { "ip": "10.96.0.2" } ] }]`,
		},
	}

	requestSender := &countingRequestSender{}
	patchConverter := &recordingPatchConverter{}

	patcher := newLoadBalancerEmulator(client)
	patcher.requestSender = requestSender
	patcher.patchConverter = patchConverter

	serviceNames, err := patcher.PatchServices()

	expectedServices := []string{"svc1-up-to-date", "svc2-out-of-date", "svc3-empty-ingress"}

	if !reflect.DeepEqual(serviceNames, expectedServices) || err != nil {
		t.Errorf("error.\nExpected: %s, <nil>\nGot: %v, %v", expectedServices, serviceNames, err)
	}

	if !reflect.DeepEqual(patchConverter.patches, expectedPatches) {
		t.Errorf("error in patches.\nExpected: %v, <nil>\nGot: %v", expectedPatches, patchConverter.patches)
	}

	if requestSender.requests != 2 {
		t.Errorf("error in number of requests sent.\nExpected: %v, <nil>\nGot: %v", 2, requestSender.requests)
	}

}

func TestCleanupPatchedIPs(t *testing.T) {
	expectedPatches := []*Patch{
		{
			Type:         "application/json-patch+json",
			NameSpace:    "ns1",
			NameSpaceSet: true,
			Resource:     "services",
			Subresource:  "status",
			ResourceName: "svc1-up-to-date",
			BodyContent:  `[{"op": "remove", "path": "/status/loadBalancer/ingress" }]`,
		},

		{
			Type:         "application/json-patch+json",
			NameSpace:    "ns2",
			NameSpaceSet: true,
			Resource:     "services",
			Subresource:  "status",
			ResourceName: "svc2-out-of-date",
			BodyContent:  `[{"op": "remove", "path": "/status/loadBalancer/ingress" }]`,
		},
	}

	client := newStubCoreClient(&core.ServiceList{
		Items: []core.Service{
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "svc1-up-to-date",
					Namespace: "ns1",
				},
				Spec: core.ServiceSpec{
					Type:      "LoadBalancer",
					ClusterIP: "10.96.0.3",
				},
				Status: core.ServiceStatus{
					LoadBalancer: core.LoadBalancerStatus{
						Ingress: []core.LoadBalancerIngress{
							{
								IP: "10.96.0.3",
							},
						},
					},
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "svc2-out-of-date",
					Namespace: "ns2",
				},
				Spec: core.ServiceSpec{
					Type:      "LoadBalancer",
					ClusterIP: "10.96.0.4",
				},
				Status: core.ServiceStatus{
					LoadBalancer: core.LoadBalancerStatus{
						Ingress: []core.LoadBalancerIngress{
							{
								IP: "10.96.0.5",
							},
						},
					},
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					Name:      "svc3-empty-ingress",
					Namespace: "ns3",
				},
				Spec: core.ServiceSpec{
					Type:      "LoadBalancer",
					ClusterIP: "10.96.0.2",
				},
				Status: core.ServiceStatus{
					LoadBalancer: core.LoadBalancerStatus{
						Ingress: []core.LoadBalancerIngress{},
					},
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					Name: "svc4-not-lb",
				},
				Spec: core.ServiceSpec{
					Type: "NodeIP",
				},
			},
		},
	})

	requestSender := &countingRequestSender{}
	patchConverter := &recordingPatchConverter{}

	patcher := newLoadBalancerEmulator(client)
	patcher.requestSender = requestSender
	patcher.patchConverter = patchConverter

	serviceNames, err := patcher.Cleanup()
	expectedServices := []string{"svc1-up-to-date", "svc2-out-of-date", "svc3-empty-ingress"}

	if !reflect.DeepEqual(serviceNames, expectedServices) || err != nil {
		t.Errorf("error.\nExpected: %s, <nil>\nGot: %v, %v", expectedServices, serviceNames, err)
	}
	if !reflect.DeepEqual(patchConverter.patches, expectedPatches) {
		t.Errorf("error in patches.\nExpected: %v, <nil>\nGot: %v", expectedPatches, patchConverter.patches)
	}
	if requestSender.requests != 2 {
		t.Errorf("error in number of requests sent.\nExpected: %v, <nil>\nGot: %v", 2, requestSender.requests)
	}
}
