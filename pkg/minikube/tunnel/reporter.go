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
	"strings"
)

type reporter interface {
	Report(tunnelState *TunnelState)
}

type simpleReporter struct {
	out       io.Writer
	lastState *TunnelState
}

func (r *simpleReporter) Report(tunnelState *TunnelState) {
	if r.lastState == tunnelState {
		return
	}
	r.lastState = tunnelState
	minikubeState := tunnelState.MinikubeState.String()

	var managedServices string
	managedServices = fmt.Sprintf("[%s]", strings.Join(tunnelState.PatchedServices, ", "))

	lbError :=  "no errors"
	if tunnelState.LoadBalancerEmulatorError != nil {
		lbError = tunnelState.LoadBalancerEmulatorError.Error()
	}

	minikubeError :=  "no errors"
	if tunnelState.MinikubeError != nil {
		minikubeError = tunnelState.MinikubeError.Error()
	}

	routerError :=  "no errors"
	if tunnelState.RouterError != nil {
		routerError = tunnelState.RouterError.Error()
	}

	errors := fmt.Sprintf(`    errors: 
		minikube: %s
		router: %s
		loadbalancer emulator: %s
`, minikubeError, routerError, lbError)

	r.out.Write([]byte(fmt.Sprintf(
		`TunnelState:	
	machine: %s
	pid: %d
	route: %s
	minikube: %s
	services: %s
%s`, tunnelState.TunnelID.MachineName,
		tunnelState.TunnelID.Pid,
		tunnelState.TunnelID.Route,
		minikubeState,
		managedServices,
		errors)))
}

func NewReporter(out io.Writer) reporter {
	return &simpleReporter{
		out: out,
	}
}
