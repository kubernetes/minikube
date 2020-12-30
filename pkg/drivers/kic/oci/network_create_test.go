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
var dockerInspectGetterMock = func(name string) (*RunResult, error) {
	var responseInBytes bytes.Buffer
	responseInBytes.WriteString(dockerResponse)
	response := &RunResult{Stdout: responseInBytes}

	return response, nil
}

func TestDockerInspectWithMTU(t *testing.T) {
	dockerInspectResponseWithMtu := `{"Name": "m2","Driver": "bridge","Subnet": "172.19.0.0/16","Gateway": "172.19.0.1","MTU": 9216, "ContainerIPs": []}`

	// setting up mock funcs
	dockerResponse = dockerInspectResponseWithMtu
	dockerInsepctGetter = dockerInspectGetterMock

	netInfo, err := dockerNetworkInspect("m2")

	if err != nil {
		t.Errorf("Expected not to have error but got %v", err)
	}

	if netInfo.mtu != 9216 {
		t.Errorf("Expected not to have MTU as 9216 but got %v", netInfo.mtu)
	}

	if !netInfo.gateway.Equal(net.ParseIP("172.19.0.1")) {
		t.Errorf("Expected not to have gateway as 172.19.0.1 but got %v", netInfo.gateway)
	}

	if !netInfo.subnet.IP.Equal(net.ParseIP("172.19.0.0")) {
		t.Errorf("Expected not to have subnet as 172.19.0.0 but got %v", netInfo.gateway)
	}
}

func TestDockerInspectWithoutMTU(t *testing.T) {
	dockerInspectResponseWithMtu := `{"Name": "m2","Driver": "bridge","Subnet": "172.19.0.0/16","Gateway": "172.19.0.1","MTU": 0, "ContainerIPs": []}`

	// setting up mock funcs
	dockerResponse = dockerInspectResponseWithMtu
	dockerInsepctGetter = dockerInspectGetterMock

	netInfo, err := dockerNetworkInspect("m2")

	if err != nil {
		t.Errorf("Expected not to have error but got %v", err)
	}

	if netInfo.mtu != 0 {
		t.Errorf("Expected not to have MTU as 0 but got %v", netInfo.mtu)
	}

	if !netInfo.gateway.Equal(net.ParseIP("172.19.0.1")) {
		t.Errorf("Expected not to have gateway as 172.19.0.1 but got %v", netInfo.gateway)
	}

	if !netInfo.subnet.IP.Equal(net.ParseIP("172.19.0.0")) {
		t.Errorf("Expected not to have subnet as 172.19.0.0 but got %v", netInfo.gateway)
	}
}
