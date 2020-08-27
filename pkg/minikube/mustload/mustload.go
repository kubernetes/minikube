/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

// Package mustload loads minikube clusters, exiting with user-friendly messages
package mustload

import (
	"fmt"
	"net"
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
)

// ClusterController holds all the needed information for a minikube cluster
type ClusterController struct {
	Config *config.ClusterConfig
	API    libmachine.API
	CP     ControlPlane
}

// ControlPlane holds all the needed information for the k8s control plane
type ControlPlane struct {
	// Host is the libmachine host object
	Host *host.Host
	// Node is our internal control object
	Node *config.Node
	// Runner provides command execution
	Runner command.Runner
	// Hostname is the host-accesible target for the apiserver
	Hostname string
	// Port is the host-accessible port for the apiserver
	Port int
	// IP is the host-accessible IP for the control plane
	IP net.IP
}

// Partial is a cmd-friendly way to load a cluster which may or may not be running
func Partial(name string, miniHome ...string) (libmachine.API, *config.ClusterConfig) {
	glog.Infof("Loading cluster: %s", name)
	api, err := machine.NewAPIClient(miniHome...)
	if err != nil {
		exit.WithError("libmachine failed", err)
	}
	cc, err := config.Load(name, miniHome...)
	if err != nil {
		if config.IsNotExist(err) {
			out.T(out.Shrug, `There is no local cluster named "{{.cluster}}"`, out.V{"cluster": name})
			exitTip("start", name, exit.GuestNotFound)
		}
		exit.WithError("Error getting cluster config", err)
	}

	return api, cc
}

// Running is a cmd-friendly way to load a running cluster
func Running(name string) ClusterController {
	api, cc := Partial(name)

	cp, err := config.PrimaryControlPlane(cc)
	if err != nil {
		exit.WithError("Unable to find control plane", err)
	}

	machineName := driver.MachineName(*cc, cp)
	hs, err := machine.Status(api, machineName)
	if err != nil {
		exit.WithError("Unable to get machine status", err)
	}

	if hs == state.None.String() {
		out.T(out.Shrug, `The control plane node "{{.name}}" does not exist.`, out.V{"name": cp.Name})
		exitTip("start", name, exit.GuestNotFound)
	}

	if hs == state.Stopped.String() {
		out.T(out.Shrug, `The control plane node must be running for this command`)
		exitTip("start", name, exit.GuestUnavailable)
	}

	if hs != state.Running.String() {
		out.T(out.Shrug, `The control plane node is not running (state={{.state}})`, out.V{"name": cp.Name, "state": hs})
		exitTip("start", name, exit.ServiceUnavailable)
	}

	host, err := machine.LoadHost(api, name)
	if err != nil {
		exit.WithError("Unable to load host", err)
	}

	cr, err := machine.CommandRunner(host)
	if err != nil {
		exit.WithError("Unable to get command runner", err)
	}

	hostname, ip, port, err := driver.ControlPlaneEndpoint(cc, &cp, host.DriverName)
	if err != nil {
		exit.WithError("Unable to get forwarded endpoint", err)
	}

	return ClusterController{
		API:    api,
		Config: cc,
		CP: ControlPlane{
			Runner:   cr,
			Host:     host,
			Node:     &cp,
			Hostname: hostname,
			IP:       ip,
			Port:     port,
		},
	}
}

// Healthy is a cmd-friendly way to load a healthy cluster
func Healthy(name string) ClusterController {
	co := Running(name)

	as, err := kverify.APIServerStatus(co.CP.Runner, co.CP.Hostname, co.CP.Port)
	if err != nil {
		out.FailureT(`Unable to get control plane status: {{.error}}`, out.V{"error": err})
		exitTip("delete", name, exit.ServiceError)
	}

	if as == state.Paused {
		out.T(out.Shrug, `The control plane for "{{.name}}" is paused!`, out.V{"name": name})
		exitTip("unpause", name, exit.ServiceConfig)
	}

	if as != state.Running {
		out.T(out.Shrug, `This control plane is not running! (state={{.state}})`, out.V{"state": as.String()})
		out.WarningT(`This is unusual - you may want to investigate using "{{.command}}"`, out.V{"command": ExampleCmd(name, "logs")})
		exitTip("start", name, exit.ServiceUnavailable)
	}
	return co
}

// ExampleCmd Return a minikube command containing the current profile name
func ExampleCmd(cname string, action string) string {
	if cname != constants.DefaultClusterName {
		return fmt.Sprintf("minikube %s -p %s", action, cname)
	}
	return fmt.Sprintf("minikube %s", action)
}

// exitTip returns an action tip and exits
func exitTip(action string, profile string, code int) {
	command := ExampleCmd(profile, action)
	out.T(out.Workaround, `To fix this, run: "{{.command}}"`, out.V{"command": command})
	os.Exit(code)
}
