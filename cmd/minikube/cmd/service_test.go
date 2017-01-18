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

	"k8s.io/client-go/pkg/api"
)

func TestCheckEndpointReady(t *testing.T) {
	endpointNoSubsets := &api.Endpoints{}
	if err := CheckEndpointReady(endpointNoSubsets); err == nil {
		t.Fatalf("Endpoint had no subsets but CheckEndpointReady did not return an error")
	}

	endpointNotReady := &api.Endpoints{
		Subsets: []api.EndpointSubset{
			{Addresses: []api.EndpointAddress{},
				NotReadyAddresses: []api.EndpointAddress{
					{IP: "1.1.1.1"},
					{IP: "2.2.2.2"},
					{IP: "3.3.3.3"},
				}}}}
	if err := CheckEndpointReady(endpointNotReady); err == nil {
		t.Fatalf("Endpoint had no Addresses but CheckEndpointReady did not return an error")
	}

	endpointReady := &api.Endpoints{
		Subsets: []api.EndpointSubset{
			{Addresses: []api.EndpointAddress{
				{IP: "1.1.1.1"},
				{IP: "2.2.2.2"},
			},
				NotReadyAddresses: []api.EndpointAddress{},
			}},
	}
	if err := CheckEndpointReady(endpointReady); err != nil {
		t.Fatalf("Endpoint was ready with at least one Address, but CheckEndpointReady returned an error")
	}
}
