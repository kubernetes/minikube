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
	"fmt"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"net"
)

type minikubeInspector struct {
	machineAPI   libmachine.API
	configLoader config.ConfigLoader
	machineName  string
}

func (m *minikubeInspector) getStateAndHost() (HostState, *host.Host, error) {
	hostState := Unknown

	h, e := cluster.CheckIfHostExistsAndLoad(m.machineAPI, m.machineName)

	if e != nil {
		e = errors.Wrapf(e, "error loading docker-machine host for: %s", m.machineName)
		return hostState, nil, e
	}

	var s state.State
	s, e = h.Driver.GetState()
	if e != nil {
		e = errors.Wrapf(e, "error getting host status for %s", m.machineName)
		return hostState, nil, e
	}

	if s == state.Running {
		hostState = Running
	} else {
		hostState = Stopped
	}

	return hostState, h, nil
}

func (m *minikubeInspector) getStateAndRoute() (HostState, *Route, error) {
	hostState, h, e := m.getStateAndHost()
	defer m.machineAPI.Close()
	if e != nil {
		return hostState, nil, e
	}
	var c config.Config
	c, e = m.configLoader.LoadConfigFromFile(m.machineName)
	if e != nil {
		e = errors.Wrapf(e, "error loading config for %s", m.machineName)
		return hostState, nil, e
	}

	var route *Route
	route, e = getRoute(h, c)
	if e != nil {
		e = errors.Wrapf(e, "error getting Route info for %s", m.machineName)
		return hostState, nil, e
	}
	return hostState, route, nil
}

func getRoute(host *host.Host, clusterConfig config.Config) (*Route, error) {
	hostDriverIP, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting host IP for %s", host.Name)
	}

	_, ipNet, e := net.ParseCIDR(clusterConfig.KubernetesConfig.ServiceCIDR)
	if e != nil {
		return nil, fmt.Errorf("error parsing service CIDR: %s", e)
	}
	ip := net.ParseIP(hostDriverIP)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP for host %s", hostDriverIP)
	}

	return &Route{
		Gateway:  ip,
		DestCIDR: ipNet,
	}, nil
}
