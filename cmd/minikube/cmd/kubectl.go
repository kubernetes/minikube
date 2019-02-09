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
	"runtime"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	pkg_config "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
)

// kubectlCmd represents the kubectl command
var kubectlCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "Run kubectl",
	Long:  `Run the kubernetes client, download it if necessary.`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %v\n", err)
			os.Exit(1)
		}
		defer api.Close()

		cc, err := pkg_config.Load()
		if err != nil && !os.IsNotExist(err) {
			console.ErrLn("Error loading profile config: %v", err)
		}

		binary := "kubectl"
		if runtime.GOOS == "windows" {
			binary = "kubectl.exe"
		}

		version := constants.DefaultKubernetesVersion
		if cc != nil {
			version = cc.KubernetesConfig.KubernetesVersion
		}

		path, err := machine.CacheBinary(binary, version, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			exit.WithError("Failed to download kubectl", err)
		}

		glog.Infof("Running %s %v", path, args)
		c := exec.Command(path, args...)
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

func init() {
	RootCmd.AddCommand(kubectlCmd)
}
