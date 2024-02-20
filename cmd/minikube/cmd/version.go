/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"encoding/json"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/version"
)

var (
	versionOutput          string
	shortVersion           bool
	listComponentsVersions bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of minikube",
	Long:  `Print the version of minikube.`,
	Run: func(_ *cobra.Command, _ []string) {
		minikubeVersion := version.GetVersion()
		gitCommitID := version.GetGitCommitID()
		data := map[string]interface{}{
			"minikubeVersion": minikubeVersion,
			"commit":          gitCommitID,
		}

		if listComponentsVersions && !shortVersion {
			co := mustload.Running(ClusterFlagValue())
			runner := co.CP.Runner
			versionCMDS := map[string]*exec.Cmd{
				"docker":      exec.Command("docker", "--version"),
				"dockerd":     exec.Command("dockerd", "--version"),
				"cri-dockerd": exec.Command("cri-dockerd", "--version"),
				"containerd":  exec.Command("containerd", "--version"),
				"crio":        exec.Command("crio", "--version"),
				"podman":      exec.Command("sudo", "podman", "--version"),
				"crictl":      exec.Command("sudo", "crictl", "--version"),
				"buildctl":    exec.Command("buildctl", "--version"),
				"ctr":         exec.Command("ctr", "--version"),
				"runc":        exec.Command("runc", "--version"),
				"crun":        exec.Command("crun", "--version"),
			}
			for k, v := range versionCMDS {
				rr, err := runner.RunCmd(v)
				if err != nil {
					klog.Warningf("error getting %s's version: %v", k, err)
					data[k] = "error"
				} else {
					version := rr.Stdout.String()
					// remove extra lines after the version
					version = strings.Split(version, "\n")[0]
					data[k] = strings.TrimSpace(version)
				}

			}

		}

		switch versionOutput {
		case "":
			if !shortVersion {
				out.Ln("minikube version: %v", minikubeVersion)
				if gitCommitID != "" {
					out.Ln("commit: %v", gitCommitID)
				}
				keys := make([]string, 0, len(data))
				for k := range data {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					v := data[k]
					// for backward compatibility we keep displaying the old way for these two
					if k == "minikubeVersion" || k == "commit" {
						continue
					}
					if v != "" {
						out.Ln("\n%s:\n%s", k, v)
					}
				}
			} else {
				out.Ln("%v", minikubeVersion)
			}
		case "json":
			json, err := json.Marshal(data)
			if err != nil {
				exit.Error(reason.InternalJSONMarshal, "version json failure", err)
			}
			out.Ln(string(json))
		case "yaml":
			yaml, err := yaml.Marshal(data)
			if err != nil {
				exit.Error(reason.InternalYamlMarshal, "version yaml failure", err)
			}
			out.Ln(string(yaml))
		default:
			exit.Message(reason.InternalOutputUsage, "error: --output must be 'yaml' or 'json'")
		}
	},
}

func init() {
	versionCmd.Flags().StringVarP(&versionOutput, "output", "o", "", "One of 'yaml' or 'json'.")
	versionCmd.Flags().BoolVar(&shortVersion, "short", false, "Print just the version number.")
	versionCmd.Flags().BoolVar(&listComponentsVersions, "components", false, "list versions of all components included with minikube. (the cluster must be running)")
}
