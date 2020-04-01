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
	"runtime"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"k8s.io/minikube/pkg/drivers/kic/oci"
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

type ControlPlane struct {
	Host          *host.Host
	Node          *config.Node
	Runner        command.Runner
	ForwardedIP   net.IP
	ForwardedPort int
}

// Partial is a cmd-friendly way to load a cluster which may or may not be running
func Partial(name string) (libmachine.API, *config.ClusterConfig) {
	glog.Infof("Loading cluster: %s", name)
	api, err := machine.NewAPIClient()
	if err != nil {
		exit.WithError("libmachine failed", err)
	}

	cc, err := config.Load(name)
	if err != nil {
		if config.IsNotExist(err) {
			out.T(out.Shrug, `There is no local cluster named "{{.cluster}}"`, out.V{"cluster": name})
			exitTip("start", name, exit.Data)
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
		exitTip("start", name, exit.Unavailable)
	}

	if hs == state.Stopped.String() {
		out.T(out.Shrug, `The control plane node must be running for this command`)
		exitTip("start", name, exit.Unavailable)
	}

	if hs != state.Running.String() {
		out.T(out.Shrug, `The control plane node is not running (state={{.state}})`, out.V{"name": cp.Name, "state": hs})
		exitTip("start", name, exit.Unavailable)
	}

	host, err := machine.LoadHost(api, name)
	if err != nil {
		exit.WithError("Unable to load host", err)
	}

	cr, err := machine.CommandRunner(host)
	if err != nil {
		exit.WithError("Unable to get command runner", err)
	}

	ipStr, err := host.Driver.GetIP()
	if err != nil {
		exit.WithError("Unable to get driver IP", err)
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		exit.WithCodeT(exit.Software, fmt.Sprintf("Unable to parse driver IP: %q", ipStr))
	}

	cpIP := cp.IP
	cpPort := cp.Port
	if driver.IsKIC(host.DriverName) && runtime.GOOS != "linux" {
		cpIP = oci.DefaultBindIPV4
		cpPort, err = oci.ForwardedPort(cc.Driver, cc.Name, cp.Port)
		if err != nil {
			exit.WithError("Unable to get forwarded port", err)
		}
	}

	return ClusterController{
		API:    api,
		Config: cc,
		CP: ControlPlane{
			Runner:        cr,
			Host:          host,
			Node:          &cp,
			ForwardedIP:   net.ParseIP(cpIP),
			ForwardedPort: cpPort,
		},
	}
}

// Healthy is a cmd-friendly way to load a healthy cluster
func Healthy(name string) ClusterController {
	co := Running(name)

	as, err := kverify.APIServerStatus(co.CP.Runner, co.CP.ForwardedIP, co.CP.ForwardedPort)
	if err != nil {
		out.T(out.FailureType, `Unable to get control plane status: {{.error}}`, out.V{"error": err})
		exitTip("delete", name, exit.Unavailable)
	}

	if as == state.Paused {
		out.T(out.Shrug, `The control plane for "{{.name}}" is paused!`, out.V{"name": name})
		exitTip("unpause", name, exit.Unavailable)
	}

	if as != state.Running {
		out.T(out.Shrug, `This control plane is not running! (state={{.state}})`, out.V{"state": as.String()})
		out.T(out.Warning, `This is unusual - you may want to investigate using "{{.command}}"`, out.V{"command": ExampleCmd(name, "logs")})
		exitTip("start", name, exit.Unavailable)
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
