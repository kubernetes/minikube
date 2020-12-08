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
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	credentialsPath = "/var/lib/minikube/google_application_credentials.json"
	projectPath     = "/var/lib/minikube/google_cloud_project"
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
	// Grab command runner from running cluster
	cc := mustload.Running(cfg.Name)
	r := cc.CP.Runner

	// Grab credentials from where GCP would normally look
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		exit.Message(reason.InternalCredsNotFound, "Could not find any GCP credentials. Either run `gcloud auth application-default login` or set the GOOGLE_APPLICATION_CREDENTIALS environment variable to the path of your credentials file.")
	}

	if creds.JSON == nil {
		// Cloud Shell sends credential files to an unusual location, let's check that location
		// For example, CLOUDSDK_CONFIG=/tmp/tmp.cflmvysoQE
		if e := os.Getenv("CLOUDSDK_CONFIG"); e != "" {
			credFile := path.Join(e, "application_default_credentials.json")
			b, err := ioutil.ReadFile(credFile)
			if err != nil {
				exit.Message(reason.InternalCredsNotFound, "Could not find any GCP credentials. Either run `gcloud auth application-default login` or set the GOOGLE_APPLICATION_CREDENTIALS environment variable to the path of your credentials file.")
			}
			creds.JSON = b
		} else {
			// We don't currently support authentication through the metadata server
			exit.Message(reason.InternalCredsNotFound, "Could not find any GCP credentials. Either run `gcloud auth application-default login` or set the GOOGLE_APPLICATION_CREDENTIALS environment variable to the path of your credentials file.")
		}
	}

	f := assets.NewMemoryAssetTarget(creds.JSON, credentialsPath, "0444")

	err = r.Copy(f)
	if err != nil {
		return err
	}

	// First check if the project env var is explicitly set
	projectEnv := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectEnv != "" {
		f := assets.NewMemoryAssetTarget([]byte(projectEnv), projectPath, "0444")
		return r.Copy(f)
	}

	// We're currently assuming gcloud is installed and in the user's path
	project, err := exec.Command("gcloud", "config", "get-value", "project").Output()
	if err == nil && len(project) > 0 {
		f := assets.NewMemoryAssetTarget(bytes.TrimSpace(project), projectPath, "0444")
		return r.Copy(f)
	}

	out.WarningT("Could not determine a Google Cloud project, which might be ok.")
	out.Step(style.Tip, false, `To set your Google Cloud project,  run: 

		gcloud config set project <project name>

or set the GOOGLE_CLOUD_PROJECT environment variable.`)

	// Copy an empty file in to avoid errors about missing files
	emptyFile := assets.NewMemoryAssetTarget([]byte{}, projectPath, "0444")
	return r.Copy(emptyFile)
}

func disableAddon(cfg *config.ClusterConfig) error {
	// Grab command runner from running cluster
	cc := mustload.Running(cfg.Name)
	r := cc.CP.Runner

	// Clean up the files generated when enabling the addon
	creds := assets.NewMemoryAssetTarget([]byte{}, credentialsPath, "0444")
	err := r.Remove(creds)
	if err != nil {
		return err
	}

	project := assets.NewMemoryAssetTarget([]byte{}, projectPath, "0444")
	err = r.Remove(project)
	if err != nil {
		return err
	}

	return nil
}
