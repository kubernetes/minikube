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

package virtualbox

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine/drivers"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/registry"
)

const (
	docURL                    = "https://minikube.sigs.k8s.io/docs/reference/drivers/virtualbox/"
	vboxSupportedMajorVersion = 5
)

func init() {
	err := registry.Register(registry.DriverDef{
		Name:     driver.VirtualBox,
		Config:   configure,
		Status:   status,
		Default:  true,
		Priority: registry.Fallback,
		Init:     func() drivers.Driver { return virtualbox.NewDriver("", "") },
	})
	if err != nil {
		panic(fmt.Sprintf("unable to register: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	d := virtualbox.NewDriver(config.MachineName(cc, n), localpath.MiniPath())
	d.Boot2DockerURL = download.LocalISOResource(cc.MinikubeISO)
	d.Memory = cc.Memory
	d.CPU = cc.CPUs
	d.DiskSize = cc.DiskSize
	d.HostOnlyCIDR = cc.HostOnlyCIDR
	d.NoShare = cc.DisableDriverMounts
	d.NoVTXCheck = cc.NoVTXCheck
	d.NatNicType = cc.NatNicType
	d.HostOnlyNicType = cc.HostOnlyNicType
	d.DNSProxy = cc.DNSProxy
	d.HostDNSResolver = cc.HostDNSResolver
	return d, nil
}

func status() registry.State {
	// Re-use this function as it's particularly helpful for Windows
	tryPath := driver.VBoxManagePath()
	path, err := exec.LookPath(tryPath)
	if err != nil {
		return registry.State{
			Error:     fmt.Errorf("unable to find VBoxManage in $PATH"),
			Fix:       "Install VirtualBox",
			Installed: false,
			Doc:       docURL,
		}
	}

	// Allow no more than 4 seconds for querying state
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	version := ""
	cmd := exec.CommandContext(ctx, path, "--version")
	stdout, verr := cmd.Output()
	// Basic timeout
	if ctx.Err() == context.DeadlineExceeded {
		klog.Warningf("%q timed out. ", strings.Join(cmd.Args, " "))
		return registry.State{Error: err, Installed: true, Running: false, Healthy: false, Fix: "Restart VirtualBox", Doc: docURL}
	}

	if verr != nil {
		klog.Warningf("unable to get virtualbox version, returned: %v", err)
		return registry.State{
			Installed: true,
			Error:     fmt.Errorf(`%q returned: %v`, strings.Join(cmd.Args, " "), err),
			Fix:       "Restart VirtualBox, or upgrade to the latest version of VirtualBox",
			Doc:       docURL,
		}
	}

	version = string(stdout)
	// some warnings related to VirtualBox configs are always printed along with version string
	// ex:
	// WARNING: The character device /dev/vboxdrv does not exist.
	// Please install the virtualbox-dkms package and the appropriate
	// headers, most likely linux-headers-generic.
	// You will not be able to start VMs until this problem is fixed.
	// 6.1.26_Ubuntur145957
	tempSlice := strings.Split(version, "\n")
	if len(tempSlice) > 2 {
		return registry.State{
			Installed: true,
			Error:     fmt.Errorf(`warning from virtualbox %s`, version),
			Fix:       "Read the docs for resolution",
			Doc:       docURL,
		}
	}

	majorVers := (strings.Split(version, "."))[0]
	majorVersInt, cerr := strconv.Atoi(majorVers)
	if cerr != nil {
		klog.Warningf("unable to convert major version: %s of vbox to an int %v", majorVers, cerr)
	} else if majorVersInt < vboxSupportedMajorVersion {
		out.WarningT("Minimum VirtualBox Version supported: {{.vers}}, current VirtualBox version: {{.cvers}}", out.V{"vers": vboxSupportedMajorVersion, "cvers": majorVersInt})
	}

	klog.Infof("virtual box version: %s", version)

	return registry.State{Installed: true, Healthy: true, Version: version}
}
