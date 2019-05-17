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
	"net"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/util"
)

type clusterInspector struct {
	machineAPI   libmachine.API
	configLoader config.Loader
	machineName  string
}

func (m *clusterInspector) getStateAndHost() (HostState, *host.Host, error) {

	h, err := cluster.CheckIfHostExistsAndLoad(m.machineAPI, m.machineName)

	if err != nil {
		err = errors.Wrapf(err, "error loading docker-machine host for: %s", m.machineName)
		return Unknown, nil, err
	}

	var s state.State
	s, err = h.Driver.GetState()
	if err != nil {
		err = errors.Wrapf(err, "error getting host status for %s", m.machineName)
		return Unknown, nil, err
	}

	if s == state.Running {
		return Running, h, nil
	}

	return Stopped, h, nil
}

func (m *clusterInspector) getStateAndRoute() (HostState, *Route, error) {
	hostState, h, err := m.getStateAndHost()
	defer m.machineAPI.Close()
	if err != nil {
		return hostState, nil, err
	}
	var c *config.Config
	c, err = m.configLoader.LoadConfigFromFile(m.machineName)
	if err != nil {
		err = errors.Wrapf(err, "error loading config for %s", m.machineName)
		return hostState, nil, err
	}

	var route *Route
	route, err = getRoute(h, *c)
	if err != nil {
		err = errors.Wrapf(err, "error getting route info for %s", m.machineName)
		return hostState, nil, err
	}
	return hostState, route, nil
}

func getRoute(host *host.Host, clusterConfig config.Config) (*Route, error) {
	hostDriverIP, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting host IP for %s", host.Name)
	}

	_, ipNet, err := net.ParseCIDR(clusterConfig.KubernetesConfig.ServiceCIDR)
	if err != nil {
		return nil, fmt.Errorf("error parsing service CIDR: %s", err)
	}
	ip := net.ParseIP(hostDriverIP)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP for host %s", hostDriverIP)
	}
	dnsIP, err := util.GetDNSIP(ipNet.String())
	if err != nil {
		return nil, err
	}
	return &Route{
		Gateway:       ip,
		DestCIDR:      ipNet,
		ClusterDomain: clusterConfig.KubernetesConfig.DNSDomain,
		ClusterDNSIP:  dnsIP,
	}, nil
}
