/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/style"
)

var configUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update machine config",
	Long:  `Update machine config, for e.g. when the minikube home folder is moved to other location.`,
	Run: func(_ *cobra.Command, _ []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.Message(reason.NewAPIClient, "NewAPIClient")
		}

		cname := ClusterFlagValue()
		_, cc := mustload.Partial(cname)

		for _, n := range cc.Nodes {
			h, err := api.Load(config.MachineName(*cc, n))
			if err != nil {
				exit.Message(reason.HostConfigLoad, "Error loading existing host: {{.error}}", out.V{"error": err})
			}

			nHost, err := newHost(api, cc, &n)
			if err != nil {
				out.Styled(style.Failure, "Error creating new host: {{.error}}", out.V{"error": err})
				os.Exit(1)
			}

			fixAuthOptionsPaths(h, nHost)

			err = fixDriverPaths(h, nHost)
			if err != nil {
				out.Styled(style.Failure, "Error fixing driver paths: {{.error}}", out.V{"error": err})
				os.Exit(1)
			}

			if err := api.Save(h); err != nil {
				exit.Message(reason.HostSaveProfile, "Save host failed: {{.error}}", out.V{"error": err})
			}
		}

		out.Styled(style.Celebrate, "Machine config paths has been updated")
	},
}

func init() {
	ConfigCmd.AddCommand(configUpdateCmd)
}

// newHost creates a new Host
func newHost(api libmachine.API, cfg *config.ClusterConfig, n *config.Node) (*host.Host, error) {
	def := registry.Driver(cfg.Driver)
	if def.Empty() {
		return nil, fmt.Errorf("unsupported/missing driver: %s", cfg.Driver)
	}

	dd, err := def.Config(*cfg, *n)
	if err != nil {
		return nil, errors.Wrap(err, "config")
	}

	data, err := json.Marshal(dd)
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}

	h, err := api.NewHost(cfg.Driver, data)
	if err != nil {
		return nil, errors.Wrap(err, "new host")
	}

	return h, nil
}

// fixAuthOptionsPaths updates AuthOptions paths
func fixAuthOptionsPaths(h, newHost *host.Host) {
	h.HostOptions.AuthOptions = newHost.HostOptions.AuthOptions
	h.HostOptions.AuthOptions.CertDir = localpath.MiniPath()
	h.HostOptions.AuthOptions.StorePath = localpath.MiniPath()
}

// fixDriverPaths update paths for driver
func fixDriverPaths(h, newHost *host.Host) error {
	var result map[string]interface{}

	err := json.Unmarshal(h.RawDriver, &result)
	if err != nil {
		return err
	}

	result["StorePath"] = localpath.MiniPath()
	result["SSHKeyPath"] = newHost.Driver.GetSSHKeyPath()

	if driver.IsKIC(h.DriverName) {
		nodeConfig := result["NodeConfig"].(map[string]interface{})
		nodeConfig["StorePath"] = localpath.MiniPath()
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, h.Driver)
}
