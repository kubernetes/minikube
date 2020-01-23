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

package addons

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/storageclass"
	pkgutil "k8s.io/minikube/pkg/util"
)

// defaultStorageClassProvisioner is the name of the default storage class provisioner
const defaultStorageClassProvisioner = "standard"

// Set sets a value
func Set(name, value, profile string) error {
	a, valid := isAddonValid(name)
	if !valid {
		return errors.Errorf("%s is not a valid addon", name)
	}

	// Run any additional validations for this property
	if err := run(name, value, profile, a.validations); err != nil {
		return errors.Wrap(err, "running validations")
	}

	// Set the value
	c, err := config.Load(profile)
	if err != nil {
		return errors.Wrap(err, "loading profile")
	}

	if err := a.set(c, name, value); err != nil {
		return errors.Wrap(err, "setting new value of addon")
	}

	// Run any callbacks for this property
	if err := run(name, value, profile, a.callbacks); err != nil {
		return errors.Wrap(err, "running callbacks")
	}

	// Write the value
	return config.Write(profile, c)
}

// Runs all the validation or callback functions and collects errors
func run(name, value, profile string, fns []setFn) error {
	var errors []error
	for _, fn := range fns {
		err := fn(name, value, profile)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%v", errors)
	}
	return nil
}

// SetBool sets a bool value
func SetBool(m *config.MachineConfig, name string, val string) error {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	if m.Addons == nil {
		m.Addons = map[string]bool{}
	}
	m.Addons[name] = b
	return nil
}

// enableOrDisableAddon updates addon status executing any commands necessary
func enableOrDisableAddon(name, val, profile string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	addon := assets.Addons[name]

	// check addon status before enabling/disabling it
	alreadySet, err := isAddonAlreadySet(addon, enable)
	if err != nil {
		out.ErrT(out.Conflict, "{{.error}}", out.V{"error": err})
		return err
	}
	//if addon is already enabled or disabled, do nothing
	if alreadySet {
		return nil
	}

	if name == "istio" && enable {
		minMem := 8192
		minCpus := 4
		memorySizeMB := pkgutil.CalculateSizeInMB(viper.GetString("memory"))
		cpuCount := viper.GetInt("cpus")
		if memorySizeMB < minMem || cpuCount < minCpus {
			out.WarningT("Enable istio needs {{.minMem}} MB of memory and {{.minCpus}} CPUs.", out.V{"minMem": minMem, "minCpus": minCpus})
		}
	}

	// TODO(r2d4): config package should not reference API, pull this out
	api, err := machine.NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "machine client")
	}
	defer api.Close()

	//if minikube is not running, we return and simply update the value in the addon
	//config and rewrite the file
	if !cluster.IsMinikubeRunning(api) {
		return nil
	}

	cfg, err := config.Load(profile)
	if err != nil && !os.IsNotExist(err) {
		exit.WithCodeT(exit.Data, "Unable to load config: {{.error}}", out.V{"error": err})
	}

	host, err := cluster.CheckIfHostExistsAndLoad(api, cfg.Name)
	if err != nil {
		return errors.Wrap(err, "getting host")
	}

	cmd, err := machine.CommandRunner(host)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	data := assets.GenerateTemplateData(cfg.KubernetesConfig)
	return enableOrDisableAddonInternal(addon, cmd, data, enable, profile)
}

func isAddonAlreadySet(addon *assets.Addon, enable bool) (bool, error) {
	addonStatus, err := addon.IsEnabled()

	if err != nil {
		return false, errors.Wrap(err, "get the addon status")
	}

	if addonStatus && enable {
		return true, nil
	} else if !addonStatus && !enable {
		return true, nil
	}

	return false, nil
}

func enableOrDisableAddonInternal(addon *assets.Addon, cmd command.Runner, data interface{}, enable bool, profile string) error {
	files := []string{}
	for _, addon := range addon.Assets {
		var addonFile assets.CopyableFile
		var err error
		if addon.IsTemplate() {
			addonFile, err = addon.Evaluate(data)
			if err != nil {
				return errors.Wrapf(err, "evaluate bundled addon %s asset", addon.GetAssetName())
			}

		} else {
			addonFile = addon
		}
		if enable {
			if err := cmd.Copy(addonFile); err != nil {
				return err
			}
		} else {
			defer func() {
				if err := cmd.Remove(addonFile); err != nil {
					glog.Warningf("error removing %s; addon should still be disabled as expected", addonFile)
				}
			}()
		}
		files = append(files, filepath.Join(addonFile.GetTargetDir(), addonFile.GetTargetName()))
	}
	command, err := kubectlCommand(profile, files, enable)
	if err != nil {
		return err
	}
	if result, err := cmd.RunCmd(command); err != nil {
		return errors.Wrapf(err, "error updating addon:\n%s", result.Output())
	}
	return nil
}

// enableOrDisableStorageClasses enables or disables storage classes
func enableOrDisableStorageClasses(name, val, profile string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrap(err, "Error parsing boolean")
	}

	class := defaultStorageClassProvisioner
	if name == "storage-provisioner-gluster" {
		class = "glusterfile"
	}
	storagev1, err := storageclass.GetStoragev1()
	if err != nil {
		return errors.Wrapf(err, "Error getting storagev1 interface %v ", err)
	}

	if enable {
		// Only StorageClass for 'name' should be marked as default
		err = storageclass.SetDefaultStorageClass(storagev1, class)
		if err != nil {
			return errors.Wrapf(err, "Error making %s the default storage class", class)
		}
	} else {
		// Unset the StorageClass as default
		err := storageclass.DisableDefaultStorageClass(storagev1, class)
		if err != nil {
			return errors.Wrapf(err, "Error disabling %s as the default storage class", class)
		}
	}

	return enableOrDisableAddon(name, val, profile)
}
