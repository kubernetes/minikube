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

package gcpauth

import (
	"context"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
)

// EnableOrDisable enables or disables the metadata addon depending on the val parameter
func EnableOrDisable(cfg *config.ClusterConfig, name string, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	if enable {
		return enableAddon(cfg)
	}
	return disableAddon(cfg)

}

func enableAddon(cfg *config.ClusterConfig) error {
	// Grab credentials from where GCP would normally look
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return err
	}

	f := assets.NewMemoryAssetTarget(creds.JSON, "/tmp/google_application_credentials.json", "0444")

	api, err := machine.NewAPIClient()
	if err != nil {
		return err
	}

	host, err := machine.LoadHost(api, driver.MachineName(*cfg, cfg.Nodes[0]))
	if err != nil {
		return err
	}

	r, err := machine.CommandRunner(host)
	if err != nil {
		return err
	}

	err = r.Copy(f)
	if err != nil {
		return err
	}

	// We're currently assuming gcloud is installed and in the user's path
	project, err := exec.Command("gcloud", "config", "get-value", "project").Output()
	if err == nil && len(project) > 0 {
		f := assets.NewMemoryAssetTarget(project, "/tmp/google_cloud_project", "0444")
		return r.Copy(f)
	}

	return nil
}

func disableAddon(cfg *config.ClusterConfig) error {
	api, err := machine.NewAPIClient()
	if err != nil {
		return err
	}

	host, err := machine.LoadHost(api, driver.MachineName(*cfg, cfg.Nodes[0]))
	if err != nil {
		return err
	}

	r, err := machine.CommandRunner(host)
	if err != nil {
		return err
	}

	// Clean up the files generated when enabling the addon
	creds := assets.NewMemoryAssetTarget([]byte{}, "/tmp/google_application_credentials.json", "0444")
	err = r.Remove(creds)
	if err != nil {
		return err
	}

	project := assets.NewMemoryAssetTarget([]byte{}, "/tmp/google_cloud_project", "0444")
	err = r.Remove(project)
	if err != nil {
		return err
	}

	return nil
}
