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

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/storageclass"
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

func EnableOrDisableAddon(name string, val string) error {

	enable, err := strconv.ParseBool(val)
	if err != nil {
		errors.Wrapf(err, "error attempted to parse enabled/disable value addon %s", name)
	}

	//TODO(r2d4): config package should not reference API, pull this out
	api, err := machine.NewAPIClient()
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
	if err != nil {
		return errors.Wrap(err, "getting host")
	}
	cmd, err := machine.GetCommandRunner(host)
	if err != nil {
		return errors.Wrap(err, "getting command runner")
	}
	if enable {
		for _, addon := range addon.Assets {
			cmd.Copy(addon)
		}
	} else {
		for _, addon := range addon.Assets {
			cmd.Remove(addon)
		}
	}
	return nil
}

func EnableOrDisableDefaultStorageClass(name, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrap(err, "Error parsing boolean")
	}

	// Special logic to disable the default storage class
	if !enable {
		err := storageclass.DisableDefaultStorageClass()
		if err != nil {
			return errors.Wrap(err, "Error disabling default storage class")
		}
	}
	return EnableOrDisableAddon(name, val)
}
