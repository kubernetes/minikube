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
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/node"
)

// parseNodeSpecs parses each occurrence independently; commas separate fields, not nodes.
func parseNodeSpecs(specs []string) ([]config.Node, error) {
	if len(specs) == 0 {
		return nil, errors.New("--node requires at least one node specification")
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
					// An empty Guest.Name is the default Linux representation.
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

// resolveNodes validates explicit --node inputs without mutating settings or profiles.
// Legacy --nodes/--ha requests continue through their existing startup paths.
func resolveNodes(cmd *cobra.Command, existing *config.ClusterConfig, v *viper.Viper) ([]config.Node, error) {
	if !cmd.Flags().Changed(nodeSpec) {
		return nil, nil
	}
	if cmd.Flags().Changed(nodes) || cmd.Flags().Changed(ha) {
		return nil, errors.New("--node cannot be combined with --nodes/-n or --ha")
	}
	if v.GetBool(ha) {
		return nil, errors.New("--node requires HA to be disabled in the configuration or environment")
	}
	specs, err := cmd.Flags().GetStringArray(nodeSpec)
	if err != nil {
		return nil, err
	}
	requested, err := parseNodeSpecs(specs)
	if err != nil {
		return nil, err
	}
	if err := validateNodeTopology(requested); err != nil {
		return nil, err
	}
	if existing != nil {
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

// applyNodeSpecs updates the node count only after all explicit specifications are valid.
func applyNodeSpecs(cmd *cobra.Command, existing *config.ClusterConfig, v *viper.Viper) ([]config.Node, error) {
	ns, err := resolveNodes(cmd, existing, v)
	if err != nil {
		return nil, err
	}
	if ns != nil {
		v.Set(nodes, len(ns))
	}
	return ns, nil
}

func validateNodeTopology(ns []config.Node) error {
	controlPlanes := 0
	for _, n := range ns {
		if n.ControlPlane {
			controlPlanes++
			if n.Guest.IsWindows() {
				return errors.New("Windows control-plane nodes are not supported")
			}
		}
	}
	if controlPlanes == 0 {
		return errors.New("node topology requires a Linux control-plane node")
	}
	if controlPlanes != 1 {
		return errors.New("--node currently supports one Linux control-plane node; use --ha without --node for Linux HA clusters")
	}
	for _, n := range ns {
		if n.Guest.IsWindows() {
			return errors.New("Windows worker provisioning is not available in this build")
		}
	}
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
	}
	return ns
}
