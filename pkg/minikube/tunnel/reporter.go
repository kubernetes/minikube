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

	"io"
	"k8s.io/minikube/pkg/minikube/tunnel/types"
	"strings"
)

type Reporter interface {
	Report(tunnelState *types.TunnelState)
}

type SimpleReporter struct {
	out       io.Writer
	lastState *types.TunnelState
}

func (r *SimpleReporter) Report(tunnelState *types.TunnelState) {
	if r.lastState == tunnelState {
		return
	}
	r.lastState = tunnelState
	var minikubeState string
	if tunnelState.MinikubeError == nil {
		minikubeState = tunnelState.MinikubeState.String()
	} else {
		minikubeState = tunnelState.MinikubeError.Error()
	}

	var routeState string
	if tunnelState.RouteError == nil {
		routeState = tunnelState.Route.String()
	} else {
		routeState = tunnelState.RouteError.Error()
	}

	var managedServices string
	if tunnelState.LoadBalancerPatcherError == nil {
		managedServices = fmt.Sprintf("[%s]", strings.Join(tunnelState.PatchedServices, ", "))
	} else {
		managedServices = tunnelState.LoadBalancerPatcherError.Error()
	}

	r.out.Write([]byte(fmt.Sprintf(
		`TunnelState:
	minikube: %s
	Route: %s
	services: %s
`, minikubeState, routeState, managedServices)))
}

func NewReporter(out io.Writer) Reporter {
	return &SimpleReporter{
		out: out,
	}
}
