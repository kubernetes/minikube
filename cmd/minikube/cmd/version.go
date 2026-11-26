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
	"maps"
	"os/exec"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
	"k8s.io/minikube/cmd/minikube/cmd/flags"
	"k8s.io/minikube/pkg/drivers/kic"
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

// versionInfo contains the data printed by the version command. It holds
// collected version data only; formatting is done separately.
type versionInfo struct {
	Version    string
	Commit     string
	Kicbase    []string
	Components map[string]string
}

// values returns the machine-readable version data. Values are interface{}
// because the output will also include lists such as kicbase and iso images.
func (d versionInfo) values() map[string]interface{} {
	values := map[string]interface{}{
		"minikubeVersion": d.Version,
		"commit":          d.Commit,
	}
	if len(d.Kicbase) > 0 {
		values["kicbase"] = d.Kicbase
	}
	for key, value := range d.Components {
		values[key] = value
	}
	return values
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of minikube",
	Long:  `Print the version of minikube.`,
	Run: func(_ *cobra.Command, _ []string) {
		switch versionOutput {
		case "", "yaml", "json":
		default:
			exit.Message(reason.InternalOutputUsage, "error: --output must be 'yaml' or 'json'")
		}

		options := flags.CommandOptions()
		info := versionInfo{
			Version: version.GetVersion(),
			Commit:  version.GetGitCommitID(),
			Kicbase: kic.DefaultImages(),
		}

		if listComponentsVersions && !shortVersion {
			co := mustload.Running(ClusterFlagValue(), options)
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
			info.Components = make(map[string]string, len(versionCMDS))
			for k, v := range versionCMDS {
				rr, err := runner.RunCmd(v)
				if err != nil {
					klog.Warningf("error getting %s's version: %v", k, err)
					info.Components[k] = "error"
				} else {
					version := rr.Stdout.String()
					// remove extra lines after the version
					version = strings.Split(version, "\n")[0]
					info.Components[k] = strings.TrimSpace(version)
				}

			}

		}

		output, err := formatVersion(info, versionOutput, shortVersion)
		if err != nil {
			exit.Error(reason.InternalError, "version failure", err)
		}
		out.String(output)
	},
}

// formatVersion returns the complete command output for the given data and
// options. It does not read flags or global state, and it does not exit.
func formatVersion(info versionInfo, output string, short bool) (string, error) {
	switch output {
	case "":
		if short {
			return info.Version + "\n", nil
		}
		ordered := yaml.MapSlice{
			{Key: "minikube version", Value: info.Version},
		}
		if info.Commit != "" {
			ordered = append(ordered, yaml.MapItem{Key: "commit", Value: info.Commit})
		}
		if len(info.Kicbase) > 0 {
			ordered = append(ordered, yaml.MapItem{Key: "kicbase", Value: info.Kicbase})
		}
		for _, key := range slices.Sorted(maps.Keys(info.Components)) {
			value := info.Components[key]
			if value != "" {
				ordered = append(ordered, yaml.MapItem{Key: key, Value: value})
			}
		}
		b, err := yaml.Marshal(ordered)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case "json":
		b, err := json.Marshal(info.values())
		if err != nil {
			return "", err
		}
		return string(b) + "\n", nil
	case "yaml":
		b, err := yaml.Marshal(info.values())
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		panic("invalid output format: " + output)
	}
}

func init() {
	versionCmd.Flags().StringVarP(&versionOutput, "output", "o", "", "One of 'yaml' or 'json'.")
	versionCmd.Flags().BoolVar(&shortVersion, "short", false, "Print just the version number.")
	versionCmd.Flags().BoolVar(&listComponentsVersions, "components", false, "list versions of all components included with minikube. (the cluster must be running)")
}
