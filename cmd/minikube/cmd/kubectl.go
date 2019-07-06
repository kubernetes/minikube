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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

var (
	kubectlInstallMode   bool
	kubectlInstallPrefix string
	kubectlPathMode      bool
)

const defaultPrefix = "/usr/local"

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

		if kubectlInstallMode {
			installKubectl(binary, version, path, kubectlInstallPrefix)
			return
		}

		if kubectlPathMode {
			console.OutLn(path)
			return
		}

		glog.Infof("Running %s %v", path, args)
		c := exec.Command(path, args...)
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

func installKubectl(binary, version, path, prefix string) {
	bindir := filepath.Join(prefix, "bin")
	installpath := filepath.Join(bindir, binary)
	glog.Infof("Installing %s (%s)", installpath, version)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		exit.WithError("Failed to read kubectl", err)
	}
	err = os.MkdirAll(bindir, 0755)
	if err != nil {
		exit.WithError("Failed to create directory", err)
	}
	err = ioutil.WriteFile(installpath, data, 0755)
	if err != nil {
		exit.WithError("Failed to write kubectl", err)
	}
}

func init() {
	kubectlCmd.Flags().BoolVar(&kubectlInstallMode, "install", false, "Install kubectl and exit")
	kubectlCmd.Flags().StringVar(&kubectlInstallPrefix, "prefix", defaultPrefix, "Installation prefix")
	kubectlCmd.Flags().BoolVar(&kubectlPathMode, "path", false, "Display kubectl path instead of running it")
	RootCmd.AddCommand(kubectlCmd)
}
