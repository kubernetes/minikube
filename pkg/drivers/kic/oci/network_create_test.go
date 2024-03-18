/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"bytes"
	"net"
	"testing"
)

var dockerResponse string
var dockerInspectGetterMock = func(_ string) (*RunResult, error) {
	var responseInBytes bytes.Buffer
	responseInBytes.WriteString(dockerResponse)
	response := &RunResult{Stdout: responseInBytes}

	return response, nil
}

func TestDockerInspect(t *testing.T) {
	var tests = []struct {
		name                  string
		dockerInspectResponse string
		gateway               string
		subnetIP              string
		mtu                   int
	}{
		{
			name:                  "withMTU",
			dockerInspectResponse: `{"Name": "m2","Driver": "bridge","Subnet": "172.19.0.0/16","Gateway": "172.19.0.1","MTU": 9216, "ContainerIPs": []}`,
			gateway:               "172.19.0.1",
			subnetIP:              "172.19.0.0",
			mtu:                   9216,
		},
		{
			name:                  "withoutMTU",
			dockerInspectResponse: `{"Name": "m2","Driver": "bridge","Subnet": "172.19.0.0/16","Gateway": "172.19.0.1","MTU": 0, "ContainerIPs": []}`,
			gateway:               "172.19.0.1",
			subnetIP:              "172.19.0.0",
			mtu:                   0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dockerInspectResponseWithMtu := tc.dockerInspectResponse

			// setting up mock funcs
			dockerResponse = dockerInspectResponseWithMtu
			dockerInspectGetter = dockerInspectGetterMock

			netInfo, err := dockerNetworkInspect("m2")

			if err != nil {
				t.Errorf("Expected not to have error but got %v", err)
			}

			if netInfo.mtu != tc.mtu {
				t.Errorf("Expected MTU to be %v but got %v", tc.mtu, netInfo.mtu)
			}

			if !netInfo.gateway.Equal(net.ParseIP(tc.gateway)) {
				t.Errorf("Expected gateway to be %v but got %v", tc.gateway, netInfo.gateway)
			}

			if !netInfo.subnet.IP.Equal(net.ParseIP(tc.subnetIP)) {
				t.Errorf("Expected subnet to be %v but got %v", tc.subnetIP, netInfo.subnet.IP)
			}
		})
	}
}

var podmanResponse string
var podmanInspectGetterMock = func(_ string) (*RunResult, error) {
	var responseInBytes bytes.Buffer
	responseInBytes.WriteString(podmanResponse)
	response := &RunResult{Stdout: responseInBytes}

	return response, nil
}

func TestPodmanInspect(t *testing.T) {
	var emptyGateway net.IP
	gateway := net.ParseIP("172.17.0.1")
	_, subnetIP, err := net.ParseCIDR("172.17.0.0/16")
	if err != nil {
		t.Fatalf("failed to parse CIDR: %v", err)
	}

	var tests = []struct {
		name                  string
		podmanInspectResponse string
		gateway               net.IP
		subnetIP              string
	}{
		{
			name:                  "WithGateway",
			podmanInspectResponse: "172.17.0.0/16,172.17.0.1",
			gateway:               gateway,
			subnetIP:              subnetIP.String(),
		},
		{
			name:                  "WithoutGateway",
			podmanInspectResponse: "172.17.0.0/16",
			gateway:               emptyGateway,
			subnetIP:              subnetIP.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			podmanInspectResponse := tc.podmanInspectResponse

			// setting up mock funcs
			podmanResponse = podmanInspectResponse
			podmanInspectGetter = podmanInspectGetterMock

			netInfo, err := podmanNetworkInspect("m2")
			if err != nil {
				t.Errorf("Expected not to have error but got %v", err)
			}

			if !netInfo.gateway.Equal(tc.gateway) {
				t.Errorf("Expected gateway to be %v but got %v", tc.gateway, netInfo.gateway)
			}

			if netInfo.subnet.String() != tc.subnetIP {
				t.Errorf("Expected subnet to be %v but got %v", tc.subnetIP, netInfo.subnet)
			}
		})
	}
}

func TestIsNetworkNotFound(t *testing.T) {
	tests := []struct {
		output     string
		isNotFound bool
	}{
		{"Error: No such network: cat", true},
		{"Error response from daemon: network cat not found", true},
		{`[
    {
        "Name": "abcde123",
        "Id": "4683c88eb412f2744e9763a4bebcb5e3b73a11dbcc79d6d9ab64ab2f10e08faa",
        "Created": "2023-09-29T17:12:11.774716834Z",
        "Scope": "local",
        "Driver": "bridge",
        "EnableIPv6": false,
        "IPAM": {
            "Driver": "default",
            "Options": {},
            "Config": [
                {
                    "Subnet": "192.168.49.0/24",
                    "Gateway": "192.168.49.1"
                }
            ]
        },
        "Internal": false,
        "Attachable": false,
        "Ingress": false,
        "ConfigFrom": {
            "Network": ""
        },
        "ConfigOnly": false,
        "Containers": {
            "b6954f226ccfdb7d190e3792be8d569e4bc5e3c44833d9e274835212fca4f4d2": {
                "Name": "p2",
                "EndpointID": "30fd6525dab2b0a4f1953a3c8cae7485be272e09938dffe3d6de81e84c574826",
                "MacAddress": "02:42:c0:a8:31:02",
                "IPv4Address": "192.168.49.2/24",
                "IPv6Address": ""
            }
        },
        "Options": {
            "--icc": "",
            "--ip-masq": "",
            "com.docker.network.driver.mtu": "65535"
        },
        "Labels": {
            "created_by.minikube.sigs.k8s.io": "true",
            "name.minikube.sigs.k8s.io": "minikube"
        }
    }
]`, false},
	}

	for _, tc := range tests {
		got := isNetworkNotFound(tc.output)
		if got != tc.isNotFound {
			t.Errorf("isNetworkNotFound(%s) = %t; want = %t", tc.output, got, tc.isNotFound)
		}
	}
}
