// +build linux

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

package kvm2

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/machine/libmachine/drivers"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

const (
	docURL = "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.KVM2,
		Config:   configure,
		Status:   status,
		Priority: registry.Preferred,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

// This is duplicate of kvm.Driver. Avoids importing the kvm2 driver, which requires cgo & libvirt.
type kvmDriver struct {
	*drivers.BaseDriver

	Memory         int
	DiskSize       int
	CPU            int
	Network        string
	PrivateNetwork string
	ISO            string
	Boot2DockerURL string
	DiskPath       string
	GPU            bool
	Hidden         bool
	ConnectionURI  string
}

func configure(mc config.MachineConfig) interface{} {
	name := config.GetMachineName()
	return kvmDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: name,
			StorePath:   localpath.MiniPath(),
			SSHUser:     "docker",
		},
		Memory:         mc.Memory,
		CPU:            mc.CPUs,
		Network:        mc.KVMNetwork,
		PrivateNetwork: "minikube-net",
		Boot2DockerURL: mc.Downloader.GetISOFileURI(mc.MinikubeISO),
		DiskSize:       mc.DiskSize,
		DiskPath:       filepath.Join(localpath.MiniPath(), "machines", name, fmt.Sprintf("%s.rawdisk", name)),
		ISO:            filepath.Join(localpath.MiniPath(), "machines", name, "boot2docker.iso"),
		GPU:            mc.KVMGPU,
		Hidden:         mc.KVMHidden,
		ConnectionURI:  mc.KVMQemuURI,
	}
}

func status() registry.State {
	path, err := exec.LookPath("virsh")
	if err != nil {
		return registry.State{Error: err, Fix: "Install libvirt", Doc: docURL}
	}

	cmd := exec.Command(path, "domcapabilities", "--virttype", "kvm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return registry.State{
			Installed: true,
			Error:     fmt.Errorf("%s failed:\n%s", strings.Join(cmd.Args, " "), strings.TrimSpace(string(out))),
			Fix:       "Follow your Linux distribution instructions for configuring KVM",
			Doc:       docURL,
		}
	}

	cmd = exec.Command("virsh", "list")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return registry.State{
			Installed: true,
			Error:     fmt.Errorf("%s failed:\n%s", strings.Join(cmd.Args, " "), strings.TrimSpace(string(out))),
			Fix:       "Check that libvirtd is properly installed and that you are a member of the appropriate libvirt group",
			Doc:       docURL,
		}
	}
	return registry.State{Installed: true, Healthy: true}
}
