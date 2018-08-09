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

package types

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"net"
)

type TunnelState struct {
	MinikubeState HostState
	MinikubeError error

	Route      *Route
	RouteError error

	PatchedServices          []string
	LoadBalancerPatcherError error
}

func (t *TunnelState) Clone() *TunnelState {
	return &TunnelState{
		MinikubeState:            t.MinikubeState,
		MinikubeError:            t.MinikubeError,
		Route:                    t.Route,
		RouteError:               t.RouteError,
		PatchedServices:          t.PatchedServices,
		LoadBalancerPatcherError: t.LoadBalancerPatcherError,
	}
}

func (t *TunnelState) String() string {
	return fmt.Sprintf("minikube(%s, e:%s), route(%s, e:%s), services(%s, e:%s)",
		t.MinikubeState,
		t.MinikubeError,
		t.Route,
		t.RouteError,
		t.PatchedServices,
		t.LoadBalancerPatcherError)
}

type Route struct {
	Gateway  net.IP
	DestCIDR *net.IPNet
}

func (r *Route) String() string {
	return fmt.Sprintf("%s -> %s", r.DestCIDR.String(), r.Gateway.String())
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

// State represents the state of a host
type HostState int

const (
	Unkown HostState = iota
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
