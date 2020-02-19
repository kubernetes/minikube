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
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/assets"
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
	glog.Infof("Setting %s=%s in profile %q", name, value, profile)
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

	glog.Infof("Writing out %q config to set %s=%v...", profile, name, value)
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
func SetBool(m *config.ClusterConfig, name string, val string) error {
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
	glog.Infof("Setting addon %s=%s in %q", name, val, profile)
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	addon := assets.Addons[name]

	// check addon status before enabling/disabling it
	alreadySet, err := isAddonAlreadySet(addon, enable, profile)
	if err != nil {
		out.ErrT(out.Conflict, "{{.error}}", out.V{"error": err})
		return err
	}

	if alreadySet {
		glog.Warningf("addon %s should already be in state %v", name, val)
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

	cfg, err := config.Load(profile)
	if err != nil && !config.IsNotExist(err) {
		exit.WithCodeT(exit.Data, "Unable to load config: {{.error}}", out.V{"error": err})
	}

	host, err := machine.CheckIfHostExistsAndLoad(api, profile)
	if err != nil || !machine.IsHostRunning(api, profile) {
		glog.Warningf("%q is not running, writing %s=%v to disk and skipping enablement (err=%v)", profile, addon.Name(), enable, err)
		return nil
	}

	cmd, err := machine.CommandRunner(host)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	data := assets.GenerateTemplateData(cfg.KubernetesConfig)
	return enableOrDisableAddonInternal(addon, cmd, data, enable, profile)
}

func isAddonAlreadySet(addon *assets.Addon, enable bool, profile string) (bool, error) {
	addonStatus, err := addon.IsEnabled(profile)
	if err != nil {
		return false, errors.Wrap(err, "is enabled")
	}

	if addonStatus && enable {
		return true, nil
	} else if !addonStatus && !enable {
		return true, nil
	}

	return false, nil
}

func enableOrDisableAddonInternal(addon *assets.Addon, cmd command.Runner, data interface{}, enable bool, profile string) error {
	deployFiles := []string{}

	for _, addon := range addon.Assets {
		var f assets.CopyableFile
		var err error
		if addon.IsTemplate() {
			f, err = addon.Evaluate(data)
			if err != nil {
				return errors.Wrapf(err, "evaluate bundled addon %s asset", addon.GetAssetName())
			}

		} else {
			f = addon
		}
		fPath := path.Join(f.GetTargetDir(), f.GetTargetName())

		if enable {
			glog.Infof("installing %s", fPath)
			if err := cmd.Copy(f); err != nil {
				return err
			}
		} else {
			glog.Infof("Removing %+v", fPath)
			defer func() {
				if err := cmd.Remove(f); err != nil {
					glog.Warningf("error removing %s; addon should still be disabled as expected", fPath)
				}
			}()
		}
		if strings.HasSuffix(fPath, ".yaml") {
			deployFiles = append(deployFiles, fPath)
		}
	}

	command, err := kubectlCommand(profile, deployFiles, enable)
	if err != nil {
		return err
	}
	glog.Infof("Running: %v", command)
	rr, err := cmd.RunCmd(command)
	if err != nil {
		return errors.Wrapf(err, "addon apply")
	}
	glog.Infof("output:\n%s", rr.Output())
	return nil
}

// enableOrDisableStorageClasses enables or disables storage classes
func enableOrDisableStorageClasses(name, val, profile string) error {
	glog.Infof("enableOrDisableStorageClasses %s=%v on %q", name, val, profile)
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

	api, err := machine.NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "machine client")
	}
	defer api.Close()

	if !machine.IsHostRunning(api, profile) {
		glog.Warningf("%q is not running, writing %s=%v to disk and skipping enablement", profile, name, val)
		return enableOrDisableAddon(name, val, profile)
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

// Start enables the default addons for a profile, plus any additional
func Start(profile string, toEnable map[string]bool, additional []string) {
	start := time.Now()
	glog.Infof("enableAddons start: toEnable=%v, additional=%s", toEnable, additional)
	defer func() {
		glog.Infof("enableAddons completed in %s", time.Since(start))
	}()

	// Get the default values of any addons not saved to our config
	for name, a := range assets.Addons {
		defaultVal, err := a.IsEnabled(profile)
		if err != nil {
			glog.Errorf("is-enabled failed for %q: %v", a.Name(), err)
			continue
		}

		_, exists := toEnable[name]
		if !exists {
			toEnable[name] = defaultVal
		}
	}

	// Apply new addons
	for _, name := range additional {
		toEnable[name] = true
	}

	toEnableList := []string{}
	for k, v := range toEnable {
		if v {
			toEnableList = append(toEnableList, k)
		}
	}
	sort.Strings(toEnableList)

	out.T(out.AddonEnable, "Enabling addons: {{.addons}}", out.V{"addons": strings.Join(toEnableList, ", ")})
	for _, a := range toEnableList {
		err := Set(a, "true", profile)
		if err != nil {
			// Intentionally non-fatal
			out.WarningT("Enabling '{{.name}}' returned an error: {{.error}}", out.V{"name": a, "error": err})
		}
	}
}
