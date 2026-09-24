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

package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
)

// parseNodeSpecs parses each occurrence independently; commas separate fields, not nodes.
func parseNodeSpecs(specs []string) ([]config.Node, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("--node requires at least one node specification")
	}
	var result []config.Node
	for i, spec := range specs {
		n := config.Node{Worker: true}
		seen := map[string]bool{}
		for _, field := range strings.Split(spec, ",") {
			key, value, ok := strings.Cut(field, "=")
			key = strings.ToLower(strings.TrimSpace(key))
			value = strings.ToLower(strings.TrimSpace(value))
			if !ok || key == "" || value == "" {
				return nil, fmt.Errorf("--node entry %d: expected nonempty key=value fields", i+1)
			}
			if seen[key] {
				return nil, fmt.Errorf("--node entry %d: duplicate key %q", i+1, key)
			}
			seen[key] = true
			switch key {
			case "role":
				switch value {
				case "control-plane":
					n.ControlPlane = true
				case "worker":
				default:
					return nil, fmt.Errorf("--node entry %d: unsupported role %q; use control-plane or worker", i+1, value)
				}
			case "os":
				switch value {
				case "linux":
					// An empty Guest.Name is the existing Linux representation.
				case config.GuestOSWindows:
					n.Guest.Name = config.GuestOSWindows
				default:
					return nil, fmt.Errorf("--node entry %d: unsupported OS %q; use linux or windows", i+1, value)
				}
			default:
				return nil, fmt.Errorf("--node entry %d: unknown key %q; use role and os", i+1, key)
			}
		}
		if !seen["role"] {
			return nil, fmt.Errorf("--node entry %d: role is required", i+1)
		}
		result = append(result, n)
	}
	// Make the control plane first while preserving worker order.
	slices.SortStableFunc(result, func(a, b config.Node) int {
		if a.ControlPlane == b.ControlPlane {
			return 0
		}
		if a.ControlPlane {
			return -1
		}
		return 1
	})
	return result, nil
}

// resolveNodes handles explicit --node inputs only, leaving the existing
// --nodes/--ha and saved-profile startup paths unchanged.
func resolveNodes(cmd *cobra.Command, existing *config.ClusterConfig, v *viper.Viper) ([]config.Node, error) {
	if !cmd.Flags().Changed(nodeSpec) {
		return nil, nil
	}
	if cmd.Flags().Changed(nodes) || cmd.Flags().Changed(ha) {
		return nil, fmt.Errorf("--node cannot be combined with --nodes/-n or --ha")
	}
	if v.GetBool(ha) {
		return nil, fmt.Errorf("--node requires HA to be disabled in the configuration or environment")
	}
	specs, err := cmd.Flags().GetStringArray(nodeSpec)
	if err != nil {
		return nil, err
	}
	requested, err := parseNodeSpecs(specs)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if len(existing.Nodes) == 0 || !existing.Nodes[0].ControlPlane {
			return nil, fmt.Errorf("profile %q has no primary control-plane node", existing.Name)
		}
		if len(requested) != len(existing.Nodes) {
			return nil, fmt.Errorf("--node cannot change the topology of existing profile %q; use a new profile", existing.Name)
		}
		for i, n := range requested {
			saved := existing.Nodes[i]
			if n.ControlPlane != saved.ControlPlane || n.Guest.IsWindows() != saved.Guest.IsWindows() {
				return nil, fmt.Errorf("--node cannot change the topology of existing profile %q; use a new profile", existing.Name)
			}
		}
		return slices.Clone(existing.Nodes), nil
	}
	return requested, nil
}

func validateNodeTopology(ns []config.Node, hostOS string) error {
	controlPlanes := 0
	for _, n := range ns {
		if n.ControlPlane {
			controlPlanes++
			if n.Guest.IsWindows() {
				return fmt.Errorf("Windows control-plane nodes are not supported")
			}
		}
	}
	if controlPlanes == 0 {
		return fmt.Errorf("node topology requires a Linux control-plane node")
	}
	if controlPlanes != 1 {
		return fmt.Errorf("--node currently supports one Linux control-plane node; use --ha without --node for Linux HA clusters")
	}
	if config.HasWindowsNodes(ns) {
		if hostOS != "windows" {
			return fmt.Errorf("Windows nodes require a Windows host with Hyper-V")
		}
		if len(ns) != 2 || controlPlanes != 1 {
			return fmt.Errorf("mixed-OS clusters currently support one Linux control-plane node and one Windows worker")
		}
	}
	return nil
}

// applyNodeDefaults validates --node before setting its count and Windows defaults.
// Without --node, existing flags and configuration are left untouched.
func applyNodeDefaults(cmd *cobra.Command, existing *config.ClusterConfig, ns []config.Node, hostOS string, v *viper.Viper) error {
	if len(ns) == 0 {
		return nil
	}
	if err := validateNodeTopology(ns, hostOS); err != nil {
		return err
	}
	required := []struct {
		flagName, want, saved string
	}{
		{"driver", driver.HyperV, ""},
		{cniFlag, "flannel", ""},
		{containerRuntime, constants.Containerd, ""},
	}
	if config.HasWindowsNodes(ns) {
		if v.GetBool(noKubernetes) {
			return fmt.Errorf("--no-kubernetes is not supported for mixed-OS clusters")
		}
		if cmd.Flags().Changed("vm-driver") && v.GetString("vm-driver") != driver.HyperV {
			return fmt.Errorf("Windows nodes require --driver=hyperv; conflicting --vm-driver=%s", v.GetString("vm-driver"))
		}
		if existing != nil {
			required[0].saved = existing.Driver
			required[1].saved = existing.KubernetesConfig.CNI
			required[2].saved = existing.KubernetesConfig.ContainerRuntime
		}
		for _, r := range required {
			if existing != nil && r.saved != r.want {
				return fmt.Errorf("Windows nodes require %s=%s, but the saved profile uses %q", r.flagName, r.want, r.saved)
			}
			if cmd.Flags().Changed(r.flagName) && v.GetString(r.flagName) != r.want {
				return fmt.Errorf("Windows nodes require --%s=%s, but --%s=%s was specified", r.flagName, r.want, r.flagName, v.GetString(r.flagName))
			}
		}
		for _, r := range required {
			if !cmd.Flags().Changed(r.flagName) {
				v.Set(r.flagName, r.want)
				if existing == nil {
					out.Infof("Windows nodes: automatically selecting --{{.flag}}={{.value}}", out.V{"flag": r.flagName, "value": r.want})
				}
			}
		}
	}
	v.Set(nodes, len(ns))
	return nil
}

func configureNodeSpecs(definitions []config.Node, primary config.Node) []config.Node {
	ns := slices.Clone(definitions)
	for i := range ns {
		n := &ns[i]
		n.Port = primary.Port
		n.KubernetesVersion = primary.KubernetesVersion
		n.ContainerRuntime = primary.ContainerRuntime
		if i > 0 {
			n.Name = node.Name(i + 1)
		}
		if n.Guest.IsWindows() {
			n.Guest.Version = constants.DefaultWindowsNodeVersion
			n.Guest.URL = viper.GetString(windowsVhdURL)
		}
	}
	return ns
}
