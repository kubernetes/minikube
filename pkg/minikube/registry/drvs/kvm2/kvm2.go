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
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
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
		Alias:    []string{driver.AliasKVM},
		Config:   configure,
		Status:   status,
		Default:  true,
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
	NUMANodeCount  int
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	name := config.MachineName(cc, n)
	return kvmDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: name,
			StorePath:   localpath.MiniPath(),
			SSHUser:     "docker",
		},
		Memory:         cc.Memory,
		CPU:            cc.CPUs,
		Network:        cc.KVMNetwork,
		PrivateNetwork: privateNetwork(cc),
		Boot2DockerURL: download.LocalISOResource(cc.MinikubeISO),
		DiskSize:       cc.DiskSize,
		DiskPath:       filepath.Join(localpath.MiniPath(), "machines", name, fmt.Sprintf("%s.rawdisk", name)),
		ISO:            filepath.Join(localpath.MiniPath(), "machines", name, "boot2docker.iso"),
		GPU:            cc.KVMGPU,
		Hidden:         cc.KVMHidden,
		ConnectionURI:  cc.KVMQemuURI,
		NUMANodeCount:  cc.KVMNUMACount,
	}, nil
}

// if network is not user-defined it defaults to "mk-<cluster_name>"
func privateNetwork(cc config.ClusterConfig) string {
	if cc.Network == "" {
		return fmt.Sprintf("mk-%s", cc.KubernetesConfig.ClusterName)
	}
	return cc.Network
}

// defaultURI returns the QEMU URI to connect to for health checks
func defaultURI() string {
	u := os.Getenv("LIBVIRT_DEFAULT_URI")
	if u != "" {
		return u
	}
	return "qemu:///system"
}

func status() registry.State {
	// Allow no more than 6 seconds for querying state
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	path, err := exec.LookPath("virsh")
	if err != nil {
		return registry.State{Error: err, Fix: "Install libvirt", Doc: docURL}
	}

	member, err := isCurrentUserLibvirtGroupMember()
	if err != nil {
		return registry.State{
			Installed: true,
			Running:   true,
			// keep the error messsage in sync with reason.providerIssues(Kind.ID: "PR_KVM_USER_PERMISSION") regexp
			Error:  fmt.Errorf("libvirt group membership check failed:\n%v", err.Error()),
			Reason: "PR_KVM_USER_PERMISSION",
			Fix:    "Check that libvirtd is properly installed and that you are a member of the appropriate libvirt group (remember to relogin for group changes to take effect!)",
			Doc:    docURL,
		}
	}
	if !member {
		return registry.State{
			Installed: true,
			Running:   true,
			// keep the error messsage in sync with reason.providerIssues(Kind.ID: "PR_KVM_USER_PERMISSION") regexp
			Error:  fmt.Errorf("libvirt group membership check failed:\nuser is not a member of the appropriate libvirt group"),
			Reason: "PR_KVM_USER_PERMISSION",
			Fix:    "Check that libvirtd is properly installed and that you are a member of the appropriate libvirt group (remember to relogin for group changes to take effect!)",
			Doc:    docURL,
		}
	}

	// On Ubuntu 19.10 (libvirt 5.4), this fails if LIBVIRT_DEFAULT_URI is unset
	cmd := exec.CommandContext(ctx, path, "domcapabilities", "--virttype", "kvm")
	cmd.Env = append(os.Environ(), fmt.Sprintf("LIBVIRT_DEFAULT_URI=%s", defaultURI()))
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return registry.State{
			Installed: true,
			Running:   false,
			Error:     fmt.Errorf("%s timed out", strings.Join(cmd.Args, " ")),
			Fix:       "Check that the libvirtd service is running and the socket is ready",
			Doc:       docURL,
		}
	}
	if err != nil {
		return registry.State{
			Installed: true,
			Running:   true,
			Error:     fmt.Errorf("%s failed:\n%s\n%v", strings.Join(cmd.Args, " "), strings.TrimSpace(string(out)), err),
			Fix:       "Follow your Linux distribution instructions for configuring KVM",
			Doc:       docURL,
		}
	}

	cmd = exec.CommandContext(ctx, "virsh", "list")
	cmd.Env = append(os.Environ(), fmt.Sprintf("LIBVIRT_DEFAULT_URI=%s", defaultURI()))
	out, err = cmd.CombinedOutput()
	if err != nil {
		return registry.State{
			Installed: true,
			Running:   true,
			Error:     fmt.Errorf("%s failed:\n%s", strings.Join(cmd.Args, " "), strings.TrimSpace(string(out))),
			Fix:       "Check that libvirtd is properly installed and that you are a member of the appropriate libvirt group (remember to relogin for group changes to take effect!)",
			Doc:       docURL,
		}
	}

	return registry.State{Installed: true, Healthy: true, Running: true}
}

// isCurrentUserLibvirtGroupMember returns if the current user is a member of "libvirt*" group.
func isCurrentUserLibvirtGroupMember() (bool, error) {
	usr, err := user.Current()
	if err != nil {
		return false, fmt.Errorf("error getting current user: %w", err)
	}
	gids, err := usr.GroupIds()
	if err != nil {
		return false, fmt.Errorf("error getting current user's GIDs: %w", err)
	}
	for _, gid := range gids {
		grp, err := user.LookupGroupId(gid)
		if err != nil {
			return false, fmt.Errorf("error getting current user's group with GID %q: %w", gid, err)
		}
		if strings.HasPrefix(grp.Name, "libvirt") {
			return true, nil
		}
	}
	return false, nil
}
