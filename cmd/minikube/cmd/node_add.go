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

package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/cmd/minikube/cmd/flags"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var (
	cpNode              bool
	workerNode          bool
	deleteNodeOnFailure bool
	osType              string

	osTypeLong = "This flag should only be used when adding a windows node to a cluster.\n\n" +
		"Specify the OS of the node to add in the format 'os=OS_TYPE,version=VERSION'.\n" +
		"This means that the node to be added will be a Windows node and the version of Windows OS to use for that node is Windows Server 2022.\n" +
		"Example: $ minikube node add --os='os=windows,version=2022'\n" +
		"Valid options for OS_TYPE are: linux, windows. If not specified, the default value is linux.\n" +
		"You do not need to specify the --os flag if you are adding a linux node."
)

var nodeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a node to the given cluster.",
	Long:  "Adds a node to the given cluster config, and starts it.",
	Run: func(cmd *cobra.Command, _ []string) {

		osType, windowsVersion, err := parseOSFlag(osType)
		if err != nil {
			exit.Message(reason.Usage, "{{.err}}", out.V{"err": err})
		}

		if err := validateOSandVersion(osType, windowsVersion); err != nil {
			exit.Message(reason.Usage, "{{.err}}", out.V{"err": err})
		}

		if osType == "windows" && cpNode {
			exit.Message(reason.Usage, "Windows node cannot be used as control-plane nodes.")
		}

		options := flags.CommandOptions()

		co := mustload.Healthy(ClusterFlagValue(), options)
		cc := co.Config

		if driver.BareMetal(cc.Driver) {
			out.FailureT("none driver does not support multi-node clusters")
		}

		if cpNode && !config.IsHA(*cc) {
			out.FailureT("Adding a control-plane node to a non-HA (non-multi-control plane) cluster is not currently supported. Please first delete the cluster and use 'minikube start --ha' to create new one.")
		}

		roles := []string{}
		if workerNode {
			roles = append(roles, "worker")
		}
		if cpNode {
			roles = append(roles, "control-plane")
		}

		// calculate appropriate new node name with id following the last existing one
		lastID, err := node.ID(cc.Nodes[len(cc.Nodes)-1].Name)
		if err != nil {
			lastID = len(cc.Nodes)
			out.ErrLn("determining last node index (will assume %d): %v", lastID, err)
		}
		name := node.Name(lastID + 1)

		out.Step(style.Happy, "Adding node {{.name}} to cluster {{.cluster}} as {{.roles}}", out.V{"name": name, "cluster": cc.Name, "roles": roles})
		n := config.Node{
			Name:              name,
			Worker:            workerNode,
			ControlPlane:      cpNode,
			KubernetesVersion: cc.KubernetesConfig.KubernetesVersion,
		}

		// Make sure to decrease the default amount of memory we use per VM if this is the first worker node
		if len(cc.Nodes) == 1 {
			if viper.GetString(memory) == "" {
				cc.Memory = 2200
			}

			if !cc.MultiNodeRequested || cni.IsDisabled(*cc) {
				warnAboutMultiNodeCNI()
			}
		}

		register.Reg.SetStep(register.InitialSetup)
		if err := node.Add(cc, n, deleteNodeOnFailure, options); err != nil {
			_, err := maybeDeleteAndRetry(cmd, *cc, n, nil, err, options)
			if err != nil {
				exit.Error(reason.GuestNodeAdd, "failed to add node", err)
			}
		}

		if err := config.SaveProfile(cc.Name, cc); err != nil {
			exit.Error(reason.HostSaveProfile, "failed to save config", err)
		}

		out.Step(style.Ready, "Successfully added {{.name}} to {{.cluster}}!", out.V{"name": name, "cluster": cc.Name})
	},
}

func init() {
	nodeAddCmd.Flags().BoolVar(&cpNode, "control-plane", false, "If set, added node will become a control-plane. Defaults to false. Currently only supported for existing HA (multi-control plane) clusters.")
	nodeAddCmd.Flags().BoolVar(&workerNode, "worker", true, "If set, added node will be available as worker. Defaults to true.")
	nodeAddCmd.Flags().BoolVar(&deleteNodeOnFailure, "delete-on-failure", false, "If set, delete the current cluster if start fails and try again. Defaults to false.")
	nodeAddCmd.Flags().StringVar(&osType, "os", "linux", osTypeLong)

	nodeCmd.AddCommand(nodeAddCmd)
}

// parseOSFlag parses the --os flag value , 'os=OS_TYPE,version=VERSION', and returns the os type and version
// For example, 'os=windows,version=2022' The output will be os: 'windows' and version: '2022' respectively
func parseOSFlag(osFlagValue string) (string, string, error) {
	// Remove all spaces from the input string
	osFlagValue = strings.ReplaceAll(osFlagValue, " ", "")
	parts := strings.Split(osFlagValue, ",")
	osInfo := map[string]string{
		"os":      "linux", // default value
		"version": "",
	}

	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			return "", "", errors.Errorf("Invalid format for --os flag: %s", osFlagValue)
		}
		osInfo[kv[0]] = kv[1]
	}

	// if os is specified to linux, set the version to empty string as it is not required
	if osInfo["os"] == "linux" {
		if osInfo["version"] != "" {
			out.WarningT("Ignoring version flag for linux os. You do not need to specify the version for linux os.")
		}
		osInfo["version"] = ""
	}

	// if os is specified to windows and version is not specified, set the default version to 2022(Windows Server 2022)
	if osInfo["os"] == "windows" && osInfo["version"] == "" {
		osInfo["version"] = "2022"
	}

	return osInfo["os"], osInfo["version"], nil
}
