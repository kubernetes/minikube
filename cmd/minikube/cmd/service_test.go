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
	"testing"

	kubeApi "k8s.io/kubernetes/pkg/api"
)

func TestCheckEndpointReady(t *testing.T) {
	endpointNoSubsets := &kubeApi.Endpoints{}
	if err := CheckEndpointReady(endpointNoSubsets); err == nil {
		t.Fatalf("Endpoint had no subsets but CheckEndpointReady did not return an error")
	}

	endpointNotReady := &kubeApi.Endpoints{
		Subsets: []kubeApi.EndpointSubset{
			{Addresses: []kubeApi.EndpointAddress{
				{IP: "1.1.1.1"},
				{IP: "2.2.2.2"}},
				NotReadyAddresses: []kubeApi.EndpointAddress{
					{IP: "3.3.3.3"},
				}}}}
	if err := CheckEndpointReady(endpointNotReady); err == nil {
		t.Fatalf("Endpoint had NotReadyAddresses but CheckEndpointReady did not return an error")
	}

	endpointReady := &kubeApi.Endpoints{
		Subsets: []kubeApi.EndpointSubset{
			{Addresses: []kubeApi.EndpointAddress{
				{IP: "1.1.1.1"},
				{IP: "2.2.2.2"},
			},
				NotReadyAddresses: []kubeApi.EndpointAddress{},
			}},
	}
	if err := CheckEndpointReady(endpointReady); err != nil {
		t.Fatalf("Endpoint was ready with no NotReadyAddresses but CheckEndpointReady returned an error")
	}
}
