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

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
)

// kubectlCmd represents the kubectl command
var kubectlCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "Run a kubectl binary matching the cluster version",
	Long: `Run the Kubernetes client, download it if necessary. Remember -- after kubectl!

Examples:
minikube kubectl -- --help
minikube kubectl -- get pods --namespace kube-system`,
	Run: func(cmd *cobra.Command, args []string) {
		co := mustload.Healthy(ClusterFlagValue())

		version := co.Config.KubernetesConfig.KubernetesVersion
		c, err := KubectlCommand(version, args...)
		if err != nil {
			out.ErrLn("Error caching kubectl: %v", err)
		}

		klog.InfoS("Running", "path", c.Path, "args", args)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			var rc int
			if exitError, ok := err.(*exec.ExitError); ok {
				waitStatus := exitError.Sys().(syscall.WaitStatus)
				rc = waitStatus.ExitStatus()
			} else {
				fmt.Fprintf(os.Stderr, "Error running %s: %v\n", path, err)
				rc = 1
			}
			os.Exit(rc)
		}
	},
}

// KubectlCommand will return kubectl command with a version matching the cluster
func KubectlCommand(version string, args ...string) (*exec.Cmd, error) {
	if version == "" {
		version = constants.DefaultKubernetesVersion
	}

	path, err := node.CacheKubectlBinary(version)
	if err != nil {
		return nil, err
	}

	return exec.Command(path, args...), nil
}
