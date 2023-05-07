/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package machine

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/libmachine/libmachine"
	"k8s.io/minikube/pkg/libmachine/libmachine/host"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
)

// Machine contains information about a machine
type Machine struct {
	*host.Host
}

// IsValid checks if the machine has the essential info needed for a machine
func (h *Machine) IsValid() bool {
	if h == nil {
		return false
	}

	if h.Host == nil {
		return false
	}

	if h.Host.Name == "" {
		return false
	}

	if h.Host.Driver == nil {
		return false
	}

	if h.Host.HostOptions == nil {
		return false
	}

	if h.Host.RawDriver == nil {
		return false
	}
	return true
}

// LoadMachine returns a Machine abstracting a libmachine.Host
func LoadMachine(name string) (*Machine, error) {
	api, err := NewAPIClient()
	if err != nil {
		return nil, err
	}

	h, err := LoadHost(api, name)
	if err != nil {
		return nil, err
	}

	var mm Machine
	if h != nil {
		mm.Host = h
	} else {
		return nil, errors.New("host is nil")
	}
	return &mm, nil
}

// provisionDockerMachine provides fast provisioning of a docker machine
func provisionMachine(h *host.Host) error {
	klog.Infof("provisioning machine ...")
	start := time.Now()
	defer func() {
		klog.Infof("provisioned machine in %s", time.Since(start))
	}()

	p, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return errors.Wrap(err, "while detecting provisioner")
	}

	return p.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
}

// saveHost is a wrapper around libmachine's Save function to proactively update the node's IP whenever a host is saved
func saveHost(api libmachine.API, h *host.Host, cfg *config.ClusterConfig, n *config.Node) error {
	if err := api.Save(h); err != nil {
		return errors.Wrap(err, "save")
	}

	// Save IP to config file for subsequent use
	ip, err := h.Driver.GetIP()
	if err != nil {
		return err
	}
	if ip == "127.0.0.1" && driver.IsQEMU(h.Driver.DriverName()) {
		ip = "10.0.2.15"
	}
	n.IP = ip
	return config.SaveNode(cfg, n)
}
