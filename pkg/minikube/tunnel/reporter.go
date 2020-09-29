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

	"k8s.io/klog/v2"
)

// reporter that reports the status of a tunnel
type reporter interface {
	Report(tunnelState *Status)
}

type simpleReporter struct {
	out       io.Writer
	lastState *Status
}

const noErrors = "no errors"

func (r *simpleReporter) Report(tunnelState *Status) {
	if r.lastState == tunnelState {
		return
	}
	r.lastState = tunnelState
	minikubeState := tunnelState.MinikubeState.String()

	managedServices := fmt.Sprintf("[%s]", strings.Join(tunnelState.PatchedServices, ", "))

	lbError := noErrors
	if tunnelState.LoadBalancerEmulatorError != nil {
		lbError = tunnelState.LoadBalancerEmulatorError.Error()
	}

	minikubeError := noErrors
	if tunnelState.MinikubeError != nil {
		minikubeError = tunnelState.MinikubeError.Error()
	}

	routerError := noErrors
	if tunnelState.RouteError != nil {
		routerError = tunnelState.RouteError.Error()
	}

	errors := fmt.Sprintf(`    errors: 
		minikube: %s
		router: %s
		loadbalancer emulator: %s
`, minikubeError, routerError, lbError)

	_, err := r.out.Write([]byte(fmt.Sprintf(
		`Status:	
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
	if err != nil {
		klog.Errorf("failed to report state %s", err)
	}
}

func newReporter(out io.Writer) reporter {
	return &simpleReporter{
		out: out,
	}
}
