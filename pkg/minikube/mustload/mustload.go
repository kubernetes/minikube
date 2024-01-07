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

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
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
	klog.Infof("Loading cluster: %s", name)
	api, err := machine.NewAPIClient(miniHome...)
	if err != nil {
		exit.Error(reason.NewAPIClient, "libmachine failed", err)
	}

	cc, err := config.Load(name, miniHome...)
	if err != nil {
		if config.IsNotExist(err) {
			out.Styled(style.Shrug, `Profile "{{.cluster}}" not found. Run "minikube profile list" to view all profiles.`, out.V{"cluster": name})
			exitTip("start", name, reason.ExGuestNotFound)
		}
		exit.Error(reason.HostConfigLoad, "Error getting cluster config", err)
	}

	return api, cc
}

// Running is a cmd-friendly way to load a running cluster
func Running(name string) ClusterController {
	ctrls, err := running(name, true)
	if err != nil {
		out.WarningT(`This is unusual - you may want to investigate using "{{.command}}"`, out.V{"command": ExampleCmd(name, "logs")})
		exit.Message(reason.GuestCpConfig, "Unable to get running control-plane nodes")
	}

	if len(ctrls) == 0 {
		out.WarningT(`This is unusual - you may want to investigate using "{{.command}}"`, out.V{"command": ExampleCmd(name, "logs")})
		exit.Message(reason.GuestCpConfig, "Unable to find any running control-plane nodes")
	}
	return ctrls[0]
}

// running returns first or all running ClusterControllers found or an error.
func running(name string, first bool) ([]ClusterController, error) {
	api, cc := Partial(name)

	cps := config.ControlPlanes(*cc)
	if len(cps) == 0 {
		return nil, fmt.Errorf("unable to find any control-plane nodes")
	}

	running := []ClusterController{}
	for _, cp := range cps {
		machineName := config.MachineName(*cc, cp)

		status, err := machine.Status(api, machineName)
		if err != nil {
			out.Styled(style.Shrug, `Unable to get control-plane node {{.name}} status (will continue): {{.err}}`, out.V{"name": machineName, "err": err})
			continue
		}

		if status == state.None.String() {
			out.Styled(style.Shrug, `The control-plane node {{.name}} does not exist (will continue)`, out.V{"name": machineName})
			continue
		}

		if status != state.Running.String() {
			out.Styled(style.Shrug, `The control-plane node {{.name}} is not running (will continue): state={{.state}}`, out.V{"name": machineName, "state": status})
			continue
		}

		host, err := machine.LoadHost(api, machineName)
		if err != nil {
			out.Styled(style.Shrug, `Unable to load control-plane node {{.name}} host (will continue): {{.err}}`, out.V{"name": machineName, "err": err})
			continue
		}

		cr, err := machine.CommandRunner(host)
		if err != nil {
			out.Styled(style.Shrug, `Unable to get control-plane node {{.name}} command runner (will continue): {{.err}}`, out.V{"name": machineName, "err": err})
			continue
		}

		hostname, ip, port, err := driver.ControlPlaneEndpoint(cc, &cp, host.DriverName)
		if err != nil {
			out.Styled(style.Shrug, `Unable to get control-plane node {{.name}} endpoint (will continue): {{.err}}`, out.V{"name": machineName, "err": err})
			continue
		}

		running = append(running, ClusterController{
			API:    api,
			Config: cc,
			CP: ControlPlane{
				Runner:   cr,
				Host:     host,
				Node:     &cp,
				Hostname: hostname,
				IP:       ip,
				Port:     port,
			}})
		if first {
			return running, nil
		}
	}
	return running, nil
}

// Healthy is a cmd-friendly way to load a healthy cluster
func Healthy(name string) ClusterController {
	ctrls, err := running(name, false)
	if err != nil {
		out.WarningT(`This is unusual - you may want to investigate using "{{.command}}"`, out.V{"command": ExampleCmd(name, "logs")})
		exit.Message(reason.GuestCpConfig, "Unable to get running control-plane nodes")
	}

	if len(ctrls) == 0 {
		out.WarningT(`This is unusual - you may want to investigate using "{{.command}}"`, out.V{"command": ExampleCmd(name, "logs")})
		exit.Message(reason.GuestCpConfig, "Unable to find any running control-plane nodes")
	}

	for _, ctrl := range ctrls {
		machineName := config.MachineName(*ctrl.Config, *ctrl.CP.Node)

		as, err := kverify.APIServerStatus(ctrl.CP.Runner, ctrl.CP.Hostname, ctrl.CP.Port)
		if err != nil {
			out.Styled(style.Shrug, `Unable to get control-plane node {{.name}} apiserver status: {{.error}}`, out.V{"name": machineName, "error": err})
			continue
		}

		if as == state.Paused {
			out.Styled(style.Shrug, `The control-plane node {{.name}} apiserver is paused (will continue)`, out.V{"name": machineName})
			continue
		}

		if as != state.Running {
			out.Styled(style.Shrug, `The control-plane node {{.name}} apiserver is not running (will continue): (state={{.state}})`, out.V{"name": machineName, "state": as.String()})
			continue
		}

		return ctrl
	}
	out.WarningT(`This is unusual - you may want to investigate using "{{.command}}"`, out.V{"command": ExampleCmd(name, "logs")})
	exit.Message(reason.GuestCpConfig, "Unable to find any healthy control-plane nodes")
	return ClusterController{}
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
	out.Styled(style.Workaround, `To start a cluster, run: "{{.command}}"`, out.V{"command": command})
	exit.Code(code)
}
