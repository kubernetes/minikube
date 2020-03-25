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
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/storageclass"
	"k8s.io/minikube/pkg/util/retry"
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

	cc, err := config.Load(profile)
	if err != nil {
		return errors.Wrap(err, "loading profile")
	}

	// Run any additional validations for this property
	if err := run(cc, name, value, a.validations); err != nil {
		return errors.Wrap(err, "running validations")
	}

	if err := a.set(cc, name, value); err != nil {
		return errors.Wrap(err, "setting new value of addon")
	}

	// Run any callbacks for this property
	if err := run(cc, name, value, a.callbacks); err != nil {
		return errors.Wrap(err, "running callbacks")
	}

	glog.Infof("Writing out %q config to set %s=%v...", profile, name, value)
	return config.Write(profile, cc)
}

// Runs all the validation or callback functions and collects errors
func run(cc *config.ClusterConfig, name string, value string, fns []setFn) error {
	var errors []error
	for _, fn := range fns {
		err := fn(cc, name, value)
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
func SetBool(cc *config.ClusterConfig, name string, val string) error {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	if cc.Addons == nil {
		cc.Addons = map[string]bool{}
	}
	cc.Addons[name] = b
	return nil
}

// enableOrDisableAddon updates addon status executing any commands necessary
func enableOrDisableAddon(cc *config.ClusterConfig, name string, val string) error {
	glog.Infof("Setting addon %s=%s in %q", name, val, cc.Name)
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	addon := assets.Addons[name]

	// check addon status before enabling/disabling it
	alreadySet, err := isAddonAlreadySet(addon, enable, cc.Name)
	if err != nil {
		out.ErrT(out.Conflict, "{{.error}}", out.V{"error": err})
		return err
	}

	if alreadySet {
		glog.Warningf("addon %s should already be in state %v", name, val)
		if !enable {
			return nil
		}
	}

	if strings.HasPrefix(name, "istio") && enable {
		minMem := 8192
		minCPUs := 4
		if cc.Memory < minMem {
			out.WarningT("Istio needs {{.minMem}}MB of memory -- your configuration only allocates {{.memory}}MB", out.V{"minMem": minMem, "memory": cc.Memory})
		}
		if cc.CPUs < minCPUs {
			out.WarningT("Istio needs {{.minCPUs}} CPUs -- your configuration only allocates {{.cpus}} CPUs", out.V{"minCPUs": minCPUs, "cpus": cc.CPUs})
		}
	}

	// TODO(r2d4): config package should not reference API, pull this out
	api, err := machine.NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "machine client")
	}
	defer api.Close()

	cp, err := config.PrimaryControlPlane(cc)
	if err != nil {
		exit.WithError("Error getting primary control plane", err)
	}

	mName := driver.MachineName(*cc, cp)
	host, err := machine.LoadHost(api, mName)
	if err != nil || !machine.IsRunning(api, mName) {
		glog.Warningf("%q is not running, writing %s=%v to disk and skipping enablement (err=%v)", mName, addon.Name(), enable, err)
		return nil
	}

	cmd, err := machine.CommandRunner(host)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	data := assets.GenerateTemplateData(cc.KubernetesConfig)
	return enableOrDisableAddonInternal(cc, addon, cmd, data, enable)
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

func enableOrDisableAddonInternal(cc *config.ClusterConfig, addon *assets.Addon, cmd command.Runner, data interface{}, enable bool) error {
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

	command := kubectlCommand(cc, deployFiles, enable)

	// Retry, because sometimes we race against an apiserver restart
	apply := func() error {
		_, err := cmd.RunCmd(command)
		if err != nil {
			glog.Warningf("apply failed, will retry: %v", err)
		}
		return err
	}

	return retry.Expo(apply, 1*time.Second, time.Second*30)
}

// enableOrDisableStorageClasses enables or disables storage classes
func enableOrDisableStorageClasses(cc *config.ClusterConfig, name string, val string) error {
	glog.Infof("enableOrDisableStorageClasses %s=%v on %q", name, val, cc.Name)
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

	cp, err := config.PrimaryControlPlane(cc)
	if err != nil {
		return errors.Wrap(err, "getting control plane")
	}
	if !machine.IsRunning(api, driver.MachineName(*cc, cp)) {
		glog.Warningf("%q is not running, writing %s=%v to disk and skipping enablement", driver.MachineName(*cc, cp), name, val)
		return enableOrDisableAddon(cc, name, val)
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

	return enableOrDisableAddon(cc, name, val)
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
