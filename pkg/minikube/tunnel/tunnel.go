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
	"os"

	"os/exec"
	"regexp"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/minikube/pkg/minikube/config"
)

//tunnel represents the basic API for a tunnel: periodically the state of the tunnel
//can be updated and when the tunnel is not needed, it can be cleaned up
//It was mostly introduced for testability.
type controller interface {
	cleanup() *Status
	update() *Status
}

func errorTunnelAlreadyExists(id *ID) error {
	return fmt.Errorf("there is already a running tunnel for this machine: %s", id)
}

func newTunnel(machineName string, machineAPI libmachine.API, configLoader config.Loader, v1Core typed_core.CoreV1Interface, registry *persistentRegistry, router router) (*tunnel, error) {
	ci := &clusterInspector{
		machineName:  machineName,
		machineAPI:   machineAPI,
		configLoader: configLoader,
	}
	state, route, err := ci.getStateAndRoute()

	if err != nil {
		return nil, fmt.Errorf("unable to determine cluster info: %s", err)
	}
	id := ID{
		Route:       route,
		MachineName: machineName,
		Pid:         getPid(),
	}
	runningTunnel, err := registry.IsAlreadyDefinedAndRunning(&id)
	if err != nil {
		return nil, fmt.Errorf("unable to check tunnel registry for conflict: %s", err)
	}
	if runningTunnel != nil {
		return nil, fmt.Errorf("another tunnel is already running, shut it down first: %s", runningTunnel)
	}

	return &tunnel{
		clusterInspector:     ci,
		router:               router,
		registry:             registry,
		loadBalancerEmulator: newLoadBalancerEmulator(v1Core),
		status: &Status{
			TunnelID:      id,
			MinikubeState: state,
		},
		reporter: &simpleReporter{
			out: os.Stdout,
		},
	}, nil

}

type tunnel struct {
	//collaborators
	clusterInspector     *clusterInspector
	router               router
	loadBalancerEmulator loadBalancerEmulator
	reporter             reporter
	registry             *persistentRegistry

	status *Status
}

func (t *tunnel) cleanup() *Status {
	glog.V(3).Infof("cleaning up %s", t.status.TunnelID.Route)
	err := t.router.Cleanup(t.status.TunnelID.Route)
	if err != nil {
		t.status.RouteError = errors.Errorf("error cleaning up route: %v", err)
		glog.V(3).Infof(t.status.RouteError.Error())
	} else {
		err = t.registry.Remove(t.status.TunnelID.Route)
		if err != nil {
			glog.V(3).Infof("error removing route from registry: %v", err)
		}
	}
	if t.status.MinikubeState == Running {
		t.status.PatchedServices, t.status.LoadBalancerEmulatorError = t.loadBalancerEmulator.Cleanup()
	}
	return t.status
}

func (t *tunnel) update() *Status {
	glog.V(3).Info("updating tunnel status...")
	var h *host.Host
	t.status.MinikubeState, h, t.status.MinikubeError = t.clusterInspector.getStateAndHost()
	defer t.clusterInspector.machineAPI.Close()
	if t.status.MinikubeState == Running {
		glog.V(3).Infof("minikube is running, trying to add route%s", t.status.TunnelID.Route)
		setupRoute(t, h)
		if t.status.RouteError == nil {
			t.status.PatchedServices, t.status.LoadBalancerEmulatorError = t.loadBalancerEmulator.PatchServices()
		}
	}
	glog.V(3).Infof("sending report %s", t.status)
	t.reporter.Report(t.status.Clone())
	return t.status
}

func setupRoute(t *tunnel, h *host.Host) {
	exists, conflict, _, err := t.router.Inspect(t.status.TunnelID.Route)
	if err != nil {
		t.status.RouteError = fmt.Errorf("error checking for route state: %s", err)
		return
	}

	if !exists && len(conflict) == 0 {
		t.status.RouteError = t.router.EnsureRouteIsAdded(t.status.TunnelID.Route)
		if t.status.RouteError != nil {
			return
		}
		//the route was added successfully, we need to make sure the registry has it too
		//this might fail in race conditions, when another process created this tunnel
		if err := t.registry.Register(&t.status.TunnelID); err != nil {
			glog.Errorf("failed to register tunnel: %s", err)
			t.status.RouteError = err
			return
		}

		if h.DriverName == "hyperkit" {
			//the virtio-net interface acts up with ip tunnels :(
			setupBridge(t)
			if t.status.RouteError != nil {
				return
			}
		}
	}

	// error scenarios

	if len(conflict) > 0 {
		t.status.RouteError = fmt.Errorf("conflicting route: %s", conflict)
		return
	}

	//the route exists, make sure that this process owns it in the registry
	existingTunnel, err := t.registry.IsAlreadyDefinedAndRunning(&t.status.TunnelID)
	if err != nil {
		glog.Errorf("failed to check for other tunnels: %s", err)
		t.status.RouteError = err
		return
	}

	if existingTunnel == nil {
		//the route exists, but "orphaned", this process will "own it" in the registry
		if err := t.registry.Register(&t.status.TunnelID); err != nil {
			glog.Errorf("failed to register tunnel: %s", err)
			t.status.RouteError = err
		}
		return
	}

	if existingTunnel.Pid != getPid() {
		//another running process owns the tunnel
		t.status.RouteError = errorTunnelAlreadyExists(existingTunnel)
		return
	}

}

func setupBridge(t *tunnel) {
	command := exec.Command("ifconfig", "bridge100")
	glog.Infof("About to run command: %s\n", command.Args)
	response, err := command.CombinedOutput()
	if err != nil {
		t.status.RouteError = fmt.Errorf("running %v: %v", command.Args, err)
		return
	}
	iface := string(response)
	pattern := regexp.MustCompile(`.*member: (en\d+) flags=.*`)
	submatch := pattern.FindStringSubmatch(iface)
	if len(submatch) != 2 {
		t.status.RouteError = fmt.Errorf("couldn't find member in bridge100 interface: %s", iface)
		return
	}

	member := submatch[1]
	command = exec.Command("sudo", "ifconfig", "bridge100", "deletem", member)
	glog.Infof("About to run command: %s\n", command.Args)
	response, err = command.CombinedOutput()
	glog.Infof(string(response))
	if err != nil {
		t.status.RouteError = fmt.Errorf("couldn't remove member %s: %s", member, err)
		return
	}

	command = exec.Command("sudo", "ifconfig", "bridge100", "addm", member)
	glog.Infof("About to run command: %s\n", command.Args)
	response, err = command.CombinedOutput()
	glog.Infof(string(response))
	if err != nil {
		t.status.RouteError = fmt.Errorf("couldn't re-add member %s: %s", member, err)
		return
	}
}
