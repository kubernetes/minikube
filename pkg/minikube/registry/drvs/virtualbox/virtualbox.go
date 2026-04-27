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
	"runtime"
	"strconv"
	"strings"
	"time"

	"k8s.io/minikube/pkg/libmachine/drivers"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/virtualbox"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/run"
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
		Init:     func(_ *run.CommandOptions) drivers.Driver { return virtualbox.NewDriver("", "") },
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
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" && !d.NoShare {
		klog.Infof("darwin/arm64: forcing NoShare=true; arm64 minikube ISO lacks VirtualBox Guest Additions")
		d.NoShare = true
	}
	d.NoVTXCheck = cc.NoVTXCheck
	d.NatNicType = cc.NatNicType
	d.HostOnlyNicType = cc.HostOnlyNicType
	d.DNSProxy = cc.DNSProxy
	d.HostDNSResolver = cc.HostDNSResolver
	return d, nil
}

func status(_ *run.CommandOptions) registry.State {
	// Reuse this function as it's particularly helpful for Windows
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

	major, minor, perr := parseVboxVersion(version)
	if perr != nil {
		klog.Warningf("unable to parse virtualbox version %q: %v", version, perr)
		if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
			return registry.State{
				Installed: true,
				Healthy:   false,
				Error:     fmt.Errorf("unable to parse VirtualBox version %q: %w", version, perr),
				Fix:       "Ensure VirtualBox 7.1 or later is installed (7.2+ recommended for Apple Silicon)",
				Doc:       docURL,
				Version:   version,
			}
		}
	} else {
		if major < vboxSupportedMajorVersion {
			out.WarningT("Minimum VirtualBox Version supported: {{.vers}}, current VirtualBox version: {{.cvers}}", out.V{"vers": vboxSupportedMajorVersion, "cvers": major})
		}
		if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
			healthy, warn := vboxArm64Policy(major, minor)
			if !healthy {
				return registry.State{
					Installed: true,
					Healthy:   false,
					Error:     fmt.Errorf("VirtualBox %d.%d is too old for Apple Silicon; host support was added in VirtualBox 7.1", major, minor),
					Fix:       "Upgrade to VirtualBox 7.1 or later (7.2+ recommended)",
					Doc:       docURL,
					Version:   version,
				}
			}
			if warn {
				out.WarningT("VirtualBox {{.v}} works on Apple Silicon but 7.2 or later is recommended for stability", out.V{"v": fmt.Sprintf("%d.%d", major, minor)})
			}
		}
	}

	klog.Infof("virtual box version: %s", version)

	return registry.State{Installed: true, Healthy: true, Version: version}
}

// parseVboxVersion extracts major and minor version numbers from the output of
// `VBoxManage --version`. The reported string typically looks like
// "7.2.6" or "7.1.12_Ubuntur169389" (a build suffix is appended to the patch
// component on some distro builds). Only the major and minor are returned
// because the arm64 gate only cares about >=7.1 and a soft warn below 7.2.
func parseVboxVersion(v string) (major, minor int, err error) {
	v = strings.TrimSpace(v)
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("unexpected VirtualBox version format: %q", v)
	}
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse major from %q: %w", v, err)
	}
	// Strip any trailing non-digit suffix from the minor component.
	minorStr := parts[1]
	for i, r := range minorStr {
		if r < '0' || r > '9' {
			minorStr = minorStr[:i]
			break
		}
	}
	if minorStr == "" {
		return 0, 0, fmt.Errorf("no digits in minor component of %q", v)
	}
	minor, err = strconv.Atoi(minorStr)
	if err != nil {
		return 0, 0, fmt.Errorf("parse minor from %q: %w", v, err)
	}
	return major, minor, nil
}

// vboxArm64Policy reports whether the given VirtualBox version is acceptable
// for use on darwin/arm64 (Apple Silicon). Apple Silicon host support was
// added in VirtualBox 7.1; 7.2+ is recommended for stability. Below 7.1 the
// driver cannot start at all — callers should surface this as an unhealthy
// registry.State rather than letting VBoxManage fail at runtime.
func vboxArm64Policy(major, minor int) (healthy bool, warn bool) {
	if major < 7 || (major == 7 && minor < 1) {
		return false, false
	}
	if major == 7 && minor < 2 {
		return true, true
	}
	return true, false
}
