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

	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/minikube/tunnel/types"
	"net"
	"reflect"
	"strings"
)

func TestAPIError(t *testing.T) {
	machineName := "nonexistentmachine"

	machineStore := &tests.FakeStore{}
	configLoader := &stubConfigLoader{}
	inspector := &minikubeInspector{
		machineStore, configLoader, machineName,
	}

	s, r, e := inspector.Inspect()

	if e == nil || !strings.Contains(e.Error(), "Machine does not exist") {
		t.Fatalf("minikube monitor should propagate errors from API, inspectCluster() returned \"%v, %v\", %v", s, r, e)
		t.Fail()
	}
}

func TestMinikubeCheckReturnsHostInformation(t *testing.T) {
	machineName := "testmachine"

	machineStore := &tests.FakeStore{
		Hosts: map[string]*host.Host{
			machineName: {
				Driver: &tests.MockDriver{
					CurrentState: state.Running,
					IP:           "1.2.3.4",
				},
			},
		},
	}
	configLoader := &stubConfigLoader{
		c: config.Config{
			KubernetesConfig: config.KubernetesConfig{
				ServiceCIDR: "96.0.0.0/12",
			},
		},
	}
	inspector := &minikubeInspector{
		machineStore, configLoader, machineName,
	}

	s, r, e := inspector.Inspect()

	if e != nil {
		t.Errorf("`error` is not nil")
	}

	ip := net.ParseIP("1.2.3.4")
	_, ipNet, _ := net.ParseCIDR("96.0.0.0/12")

	expectedRoute := &types.Route{
		Gateway:  ip,
		DestCIDR: ipNet,
	}

	if s != types.Running {
		t.Errorf("expected running, got %s", s)
	}
	if !reflect.DeepEqual(r, expectedRoute) {
		t.Errorf("expected %v, got %v", expectedRoute, r)
	}
}

func TestUnparseableCIDR(t *testing.T) {
	cfg := config.Config{
		KubernetesConfig: config.KubernetesConfig{
			ServiceCIDR: "bad.cidr.0.0/12",
		}}
	h := &host.Host{
		Driver: &tests.MockDriver{
			IP: "192.168.1.1",
		},
	}

	_, e := toRoute(h, cfg)

	if e == nil {
		t.Errorf("expected non nil error, instead got %s", e)
		t.Fail()
	}
}

func TestRouteIPDetection(t *testing.T) {
	expectedTargetCIDR := "10.96.0.0/12"

	cfg := config.Config{
		KubernetesConfig: config.KubernetesConfig{
			ServiceCIDR: expectedTargetCIDR,
		},
	}

	expectedGatewayIP := "192.168.1.1"
	h := &host.Host{
		Driver: &tests.MockDriver{
			IP: expectedGatewayIP,
		},
	}

	routerConfig, err := toRoute(h, cfg)

	if err != nil {
		t.Errorf("expected no errors but got: %s", err)
	}

	if routerConfig.DestCIDR.String() != expectedTargetCIDR {
		t.Errorf("addTargetCIDR doesn't match, expected '%s', got '%s'", expectedTargetCIDR, routerConfig.DestCIDR)
	}

	if routerConfig.Gateway.String() != expectedGatewayIP {
		t.Errorf("add gateway IP doesn't match, expected '%s', got '%s'", expectedGatewayIP, routerConfig.Gateway)
	}

}
