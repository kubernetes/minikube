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
	"k8s.io/apimachinery/pkg/types"
	"net"
)

type TunnelStatus struct {
	TunnelID TunnelID

	MinikubeState HostState
	MinikubeError error

	RouterError error

	PatchedServices           []string
	LoadBalancerEmulatorError error
}

func (t *TunnelStatus) Clone() *TunnelStatus {
	return &TunnelStatus{
		TunnelID:                  t.TunnelID,
		MinikubeState:             t.MinikubeState,
		MinikubeError:             t.MinikubeError,
		RouterError:               t.RouterError,
		PatchedServices:           t.PatchedServices,
		LoadBalancerEmulatorError: t.LoadBalancerEmulatorError,
	}
}

func (t *TunnelStatus) String() string {
	return fmt.Sprintf("id(%v), minikube(%s, e:%s), route(%s, e:%s), services(%s, e:%s)",
		t.TunnelID,
		t.MinikubeState,
		t.MinikubeError,
		t.TunnelID.Route,
		t.RouterError,
		t.PatchedServices,
		t.LoadBalancerEmulatorError)
}

type Route struct {
	Gateway  net.IP
	DestCIDR *net.IPNet
}

func (r *Route) String() string {
	return fmt.Sprintf("%s -> %s", r.DestCIDR.String(), r.Gateway.String())
}

func (r *Route) Equal(other *Route) bool {
	return other != nil && r.DestCIDR.IP.Equal(other.DestCIDR.IP) &&
		r.DestCIDR.Mask.String() == other.DestCIDR.Mask.String() &&
		r.Gateway.Equal(other.Gateway)
}

type Patch struct {
	Type         types.PatchType
	NameSpace    string
	NameSpaceSet bool
	Resource     string
	Subresource  string
	ResourceName string
	BodyContent  string
}

// State represents the status of a host
type HostState int

const (
	Unknown HostState = iota
	Running
	Stopped
)

var states = []string{
	"Unknown",
	"Running",
	"Stopped",
}

func (h HostState) String() string {
	return states[h]
}
