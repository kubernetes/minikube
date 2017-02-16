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

package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/sshutil"
)

// Runs all the validation or callback functions and collects errors
func run(name string, value string, fns []setFn) error {
	var errors []error
	for _, fn := range fns {
		err := fn(name, value)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%v", errors)
	}
	return nil
}

func findSetting(name string) (Setting, error) {
	for _, s := range settings {
		if name == s.name {
			return s, nil
		}
	}
	return Setting{}, fmt.Errorf("Property name %s not found", name)
}

// Set Functions

func SetString(m config.MinikubeConfig, name string, val string) error {
	m[name] = val
	return nil
}

func SetInt(m config.MinikubeConfig, name string, val string) error {
	i, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	m[name] = i
	return nil
}

func SetBool(m config.MinikubeConfig, name string, val string) error {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	m[name] = b
	return nil
}

func GetClientType() machine.ClientType {
	if viper.GetBool(useVendoredDriver) {
		return machine.ClientTypeLocal
	}
	return machine.ClientTypeRPC
}

func EnableOrDisableAddon(name string, val string) error {

	enable, err := strconv.ParseBool(val)
	if err != nil {
		errors.Wrapf(err, "error attempted to parse enabled/disable value addon %s", name)
	}

	// allows for additional prompting of information when enabling addons
	if enable {
		switch name {
		case "registry-creds":
			posResponses := []string{"yes", "y"}
			negResponses := []string{"no", "n"}

			enableAWSECR := AskForYesNoConfirmation("\nDo you want to enable AWS Elastic Container Registry?", posResponses, negResponses)
			if enableAWSECR {
				awsAccessID := AskForStaticValue("-- Enter AWS Access Key ID: ")
				awsAccessKey := AskForStaticValue("-- Enter AWS Secret Access Key: ")
				awsRegion := AskForStaticValue("-- Enter AWS Region: ")
				awsAccount := AskForStaticValue("-- Enter 12 digit AWS Account ID: ")

				cluster.CreateSecret(
					"kube-system",
					"registry-creds-ecr",
					map[string]string{
						"AWS_ACCESS_KEY_ID":     awsAccessID,
						"AWS_SECRET_ACCESS_KEY": awsAccessKey,
						"aws-account":           awsAccount,
						"awsregion":             awsRegion,
					},
					map[string]string{
						"app":   "registry-creds",
						"cloud": "ecr",
						"kubernetes.io/minikube-addons": "registry-creds",
					})
			}

			enableGCR := AskForYesNoConfirmation("\nDo you want to enable Google Container Registry?", posResponses, negResponses)
			if enableGCR {
				fmt.Println("-- Enter applicatoin_default_credentials.json as base64 by running following command:")
				gcrApplicationDefaultCredentials := AskForStaticValue("  base64 -w 0 $HOME/.config/gcloud/application_default_credentials.json: ")

				cluster.CreateSecret(
					"kube-system",
					"registry-creds-gcr",
					map[string]string{
						"application_default_credentials.json": gcrApplicationDefaultCredentials,
					},
					map[string]string{
						"app":   "registry-creds",
						"cloud": "gcr",
						"kubernetes.io/minikube-addons": "registry-creds",
					})
			}
			break
		}
	} else {
		// Cleanup existing secrets
		cluster.DeleteSecret("kube-system", "registry-creds-ecr")
		cluster.DeleteSecret("kube-system", "registry-creds-gcr")
	}

	//TODO(r2d4): config package should not reference API, pull this out
	api, err := machine.NewAPIClient(GetClientType())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
		os.Exit(1)
	}
	defer api.Close()
	cluster.EnsureMinikubeRunningOrExit(api, 0)

	addon, _ := assets.Addons[name] // validation done prior
	if err != nil {
		return err
	}
	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if enable {
		if err = transferAddonViaDriver(addon, host.Driver); err != nil {
			return errors.Wrapf(err, "Error transferring addon %s to VM", name)
		}
	} else {
		if err = deleteAddonViaDriver(addon, host.Driver); err != nil {
			return errors.Wrapf(err, "Error deleting addon %s from VM", name)
		}
	}
	return nil
}

func deleteAddonViaDriver(addon *assets.Addon, d drivers.Driver) error {
	client, err := sshutil.NewSSHClient(d)
	if err != nil {
		return err
	}
	if err := sshutil.DeleteAddon(addon, client); err != nil {
		return err
	}
	return nil
}

func transferAddonViaDriver(addon *assets.Addon, d drivers.Driver) error {
	client, err := sshutil.NewSSHClient(d)
	if err != nil {
		return err
	}
	if err := sshutil.TransferAddon(addon, client); err != nil {
		return err
	}
	return nil
}
