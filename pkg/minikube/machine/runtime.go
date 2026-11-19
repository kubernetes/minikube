/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
)

// containerRuntimeUser is implemented by drivers that create a container
// runtime manager inside the driver (ssh and none). Other drivers do not use
// the runtime to stop containers, so a missing ContainerRuntime field is harmless.
type containerRuntimeUser interface {
	UseContainerRuntime(name string) error
}

// driverNamer is the part of libmachine's Driver used to find the profile.
type driverNamer interface {
	DriverName() string
}

// applyProfileContainerRuntime points ssh and none drivers at the runtime
// recorded in the profile when the machine config does not name one.
//
// Those drivers build the runtime in NewDriver. Load and NewHost start from an
// empty config, which is the Docker runtime, then unmarshal the machine config.
// UnmarshalJSON can switch runtimes only when that config has ContainerRuntime.
// Configs written before the field existed omit it, so a containerd or cri-o
// cluster comes back as Docker. KubernetesConfig.ContainerRuntime has been
// stored on the profile for a long time, and it is the value configure() passed
// to the driver when the cluster was created.
func applyProfileContainerRuntime(drv driverNamer, raw []byte, machineName, miniHome string) error {
	user, ok := drv.(containerRuntimeUser)
	if !ok {
		return nil
	}
	if _, recorded := containerRuntimeFromRaw(raw); recorded {
		return nil
	}

	rt, ok := profileContainerRuntime(miniHome, machineName, drv.DriverName())
	if !ok {
		return nil
	}
	klog.Infof("machine %q (%s) has no ContainerRuntime; using %q from the profile config", machineName, drv.DriverName(), rt)
	if err := user.UseContainerRuntime(rt); err != nil {
		return fmt.Errorf("container runtime %q from profile: %w", rt, err)
	}
	return nil
}

// containerRuntimeFromRaw reports the ContainerRuntime value stored in a driver
// config. The boolean is false when the field is absent. An explicit empty
// string is present: older docker clusters saved that zero value on purpose.
func containerRuntimeFromRaw(raw []byte) (string, bool) {
	if len(strings.TrimSpace(string(raw))) == 0 {
		return "", false
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		return "", false
	}
	var encoded json.RawMessage
	found := false
	for key, value := range probe {
		if strings.EqualFold(key, "ContainerRuntime") {
			encoded = value
			found = true
			break
		}
	}
	if !found {
		return "", false
	}
	var name string
	if err := json.Unmarshal(encoded, &name); err != nil {
		return "", false
	}
	return name, true
}

// profileContainerRuntime returns the profile runtime for machineName.
// The boolean is false when no matching profile records a runtime. An empty
// runtime is not a match: profiles from before the default runtime was made
// explicit stored "" for Docker, and cruntime.New already maps "" to Docker.
func profileContainerRuntime(miniHome, machineName, machineDriver string) (string, bool) {
	if rt, owned := runtimeFromProfile(miniHome, machineName, machineName, machineDriver); owned {
		if rt == "" {
			return "", false
		}
		return rt, true
	}

	entries, err := os.ReadDir(filepath.Join(miniHome, "profiles"))
	if err != nil {
		klog.V(2).Infof("profiles dir for %q: %v", machineName, err)
		return "", false
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == machineName {
			continue
		}
		rt, owned := runtimeFromProfile(miniHome, entry.Name(), machineName, machineDriver)
		if !owned {
			continue
		}
		if rt == "" {
			return "", false
		}
		return rt, true
	}
	return "", false
}

// runtimeFromProfile reports the runtime of one profile and whether that
// profile owns machineName. owned is true even when the runtime string is
// empty, so the caller does not keep scanning after the owning profile.
func runtimeFromProfile(miniHome, profileName, machineName, machineDriver string) (runtime string, owned bool) {
	cc, err := config.Load(profileName, miniHome)
	if err != nil || cc == nil {
		return "", false
	}
	if !profileDriverMatches(cc.Driver, machineDriver) {
		return "", false
	}
	if !profileOwnsMachine(profileName, cc, machineName) {
		return "", false
	}
	return cc.KubernetesConfig.ContainerRuntime, true
}

func profileDriverMatches(profileDriver, machineDriver string) bool {
	if profileDriver == "" {
		return true
	}
	return canonicalDriver(profileDriver) == canonicalDriver(machineDriver)
}

func canonicalDriver(name string) string {
	switch name {
	case driver.AliasSSH:
		return driver.SSH
	case driver.AliasNative:
		return driver.None
	default:
		return name
	}
}

func profileOwnsMachine(profileName string, cc *config.ClusterConfig, machineName string) bool {
	cfg := *cc
	if cfg.Name == "" {
		cfg.Name = profileName
	}
	if len(cfg.Nodes) == 0 {
		return cfg.Name == machineName
	}
	for _, n := range cfg.Nodes {
		if config.MachineName(cfg, n) == machineName {
			return true
		}
	}
	return false
}
