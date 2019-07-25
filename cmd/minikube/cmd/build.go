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
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a container image",
	Long: `Run the docker client, download it if necessary.
Examples:
minikube build .`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %v\n", err)
			os.Exit(1)
		}
		defer api.Close()

		version := constants.DefaultBuildDockerVersion
		if runtime.GOOS == "windows" {
			version = constants.FallbackBuildDockerVersion
		}
		archive, err := machine.CacheDockerArchive("docker", version, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			exit.WithError("Failed to download docker", err)
		}

		binary := "docker"
		if runtime.GOOS == "windows" {
			binary = "docker.exe"
		}
		path := filepath.Join(filepath.Dir(archive), binary)

		err = machine.ExtractBinary(archive, path, fmt.Sprintf("docker/%s", binary))
		if err != nil {
			exit.WithError("Failed to extract docker", err)
		}

		envMap, err := cluster.GetHostDockerEnv(api)
		if err != nil {
			exit.WithError("Failed to get docker env", err)
		}

		tlsVerify := envMap["DOCKER_TLS_VERIFY"]
		certPath := envMap["DOCKER_CERT_PATH"]
		dockerHost := envMap["DOCKER_HOST"]

		options := []string{}
		if tlsVerify != "" {
			options = append(options, "--tlsverify")
		}
		if certPath != "" {
			options = append(options, "--tlscacert", filepath.Join(certPath, "ca.pem"))
			options = append(options, "--tlscert", filepath.Join(certPath, "cert.pem"))
			options = append(options, "--tlskey", filepath.Join(certPath, "key.pem"))
		}
		options = append(options, "-H", dockerHost)

		options = append(options, "build")
		args = append(options, args...)

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

func init() {
	RootCmd.AddCommand(buildCmd)
}
