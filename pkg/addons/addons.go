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
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/storageclass"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util/retry"
)

// defaultStorageClassProvisioner is the name of the default storage class provisioner
const defaultStorageClassProvisioner = "standard"

// RunCallbacks runs all actions associated to an addon, but does not set it (thread-safe)
func RunCallbacks(cc *config.ClusterConfig, name string, value string) error {
	glog.Infof("Setting %s=%s in profile %q", name, value, cc.Name)
	a, valid := isAddonValid(name)
	if !valid {
		return errors.Errorf("%s is not a valid addon", name)
	}

	// Run any additional validations for this property
	if err := run(cc, name, value, a.validations); err != nil {
		return errors.Wrap(err, "running validations")
	}

	// Run any callbacks for this property
	if err := run(cc, name, value, a.callbacks); err != nil {
		return errors.Wrap(err, "running callbacks")
	}
	return nil
}

// Set sets a value in the config (not threadsafe)
func Set(cc *config.ClusterConfig, name string, value string) error {
	a, valid := isAddonValid(name)
	if !valid {
		return errors.Errorf("%s is not a valid addon", name)
	}
	return a.set(cc, name, value)
}

// SetAndSave sets a value and saves the config
func SetAndSave(profile string, name string, value string) error {
	cc, err := config.Load(profile)
	if err != nil {
		return errors.Wrap(err, "loading profile")
	}

	if err := RunCallbacks(cc, name, value); err != nil {
		return errors.Wrap(err, "run callbacks")
	}

	if err := Set(cc, name, value); err != nil {
		return errors.Wrap(err, "set")
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

// SetBool sets a bool value in the config (not threadsafe)
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
	if isAddonAlreadySet(cc, addon, enable) {
		glog.Warningf("addon %s should already be in state %v", name, val)
		if !enable {
			return nil
		}
	}

	// to match both ingress and ingress-dns addons
	if strings.HasPrefix(name, "ingress") && enable {
		if driver.IsKIC(cc.Driver) && runtime.GOOS != "linux" {
			exit.Message(reason.Usage, `Due to networking limitations of driver {{.driver_name}} on {{.os_name}}, {{.addon_name}} addon is not supported.
Alternatively to use this addon you can use a vm-based driver:

	'minikube start --vm=true'

To track the update on this work in progress feature please check:
https://github.com/kubernetes/minikube/issues/7332`, out.V{"driver_name": cc.Driver, "os_name": runtime.GOOS, "addon_name": name})
		} else if driver.BareMetal(cc.Driver) {
			exit.Message(reason.Usage, `Due to networking limitations of driver {{.driver_name}}, {{.addon_name}} addon is not supported. Try using a different driver.`,
				out.V{"driver_name": cc.Driver, "addon_name": name})
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
		exit.Error(reason.GuestCpConfig, "Error getting primary control plane", err)
	}

	mName := driver.MachineName(*cc, cp)
	host, err := machine.LoadHost(api, mName)
	if err != nil || !machine.IsRunning(api, mName) {
		glog.Warningf("%q is not running, setting %s=%v and skipping enablement (err=%v)", mName, addon.Name(), enable, err)
		return nil
	}

	if name == "registry" {
		if driver.NeedsPortForward(cc.Driver) {
			port, err := oci.ForwardedPort(cc.Driver, cc.Name, constants.RegistryAddonPort)
			if err != nil {
				return errors.Wrap(err, "registry port")
			}
			out.T(style.Tip, `Registry addon on with {{.driver}} uses {{.port}} please use that instead of default 5000`, out.V{"driver": cc.Driver, "port": port})
			out.T(style.Documentation, `For more information see: https://minikube.sigs.k8s.io/docs/drivers/{{.driver}}`, out.V{"driver": cc.Driver})
		}
	}

	cmd, err := machine.CommandRunner(host)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	data := assets.GenerateTemplateData(cc.KubernetesConfig)
	return enableOrDisableAddonInternal(cc, addon, cmd, data, enable)
}

func isAddonAlreadySet(cc *config.ClusterConfig, addon *assets.Addon, enable bool) bool {
	enabled := addon.IsEnabled(cc)
	if enabled && enable {
		return true
	}

	if !enabled && !enable {
		return true
	}

	return false
}

func enableOrDisableAddonInternal(cc *config.ClusterConfig, addon *assets.Addon, cmd command.Runner, data interface{}, enable bool) error {
	deployFiles := []string{}

	for _, addon := range addon.Assets {
		var f assets.CopyableFile
		var err error
		if addon.IsTemplate() {
			f, err = addon.Evaluate(data)
			if err != nil {
				return errors.Wrapf(err, "evaluate bundled addon %s asset", addon.GetSourcePath())
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

	// Retry, because sometimes we race against an apiserver restart
	apply := func() error {
		_, err := cmd.RunCmd(kubectlCommand(cc, deployFiles, enable))
		if err != nil {
			glog.Warningf("apply failed, will retry: %v", err)
		}
		return err
	}

	return retry.Expo(apply, 250*time.Millisecond, 2*time.Minute)
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

	storagev1, err := storageclass.GetStoragev1(cc.Name)
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

	return enableOrDisableAddon(cc, name, val)
}

func verifyAddonStatus(cc *config.ClusterConfig, name string, val string) error {
	return verifyAddonStatusInternal(cc, name, val, "kube-system")
}

func verifyGCPAuthAddon(cc *config.ClusterConfig, name string, val string) error {
	return verifyAddonStatusInternal(cc, name, val, "gcp-auth")
}

func verifyAddonStatusInternal(cc *config.ClusterConfig, name string, val string, ns string) error {
	glog.Infof("Verifying addon %s=%s in %q", name, val, cc.Name)
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}

	label, ok := addonPodLabels[name]
	if ok && enable {
		out.T(style.HealthCheck, "Verifying {{.addon_name}} addon...", out.V{"addon_name": name})
		client, err := kapi.Client(viper.GetString(config.ProfileName))
		if err != nil {
			return errors.Wrapf(err, "get kube-client to validate %s addon: %v", name, err)
		}

		err = kapi.WaitForPods(client, ns, label, time.Minute*3)
		if err != nil {
			return errors.Wrapf(err, "verifying %s addon pods : %v", name, err)
		}

	}
	return nil
}

// Start enables the default addons for a profile, plus any additional
func Start(wg *sync.WaitGroup, cc *config.ClusterConfig, toEnable map[string]bool, additional []string) {
	wg.Add(1)
	defer wg.Done()

	start := time.Now()
	glog.Infof("enableAddons start: toEnable=%v, additional=%s", toEnable, additional)
	defer func() {
		glog.Infof("enableAddons completed in %s", time.Since(start))
	}()

	// Get the default values of any addons not saved to our config
	for name, a := range assets.Addons {
		defaultVal := a.IsEnabled(cc)

		_, exists := toEnable[name]
		if !exists {
			toEnable[name] = defaultVal
		}
	}

	// Apply new addons
	for _, name := range additional {
		// replace heapster as metrics-server because heapster is deprecated
		if name == "heapster" {
			name = "metrics-server"
		}
		// if the specified addon doesn't exist, skip enabling
		_, e := isAddonValid(name)
		if e {
			toEnable[name] = true
		}
	}

	toEnableList := []string{}
	for k, v := range toEnable {
		if v {
			toEnableList = append(toEnableList, k)
		}
	}
	sort.Strings(toEnableList)

	var awg sync.WaitGroup

	defer func() { // making it show after verifications( not perfect till #7613 is closed)
		register.Reg.SetStep(register.EnablingAddons)
		out.T(style.AddonEnable, "Enabled addons: {{.addons}}", out.V{"addons": strings.Join(toEnableList, ", ")})
	}()
	for _, a := range toEnableList {
		awg.Add(1)
		go func(name string) {
			err := RunCallbacks(cc, name, "true")
			if err != nil {
				out.WarningT("Enabling '{{.name}}' returned an error: {{.error}}", out.V{"name": name, "error": err})
			}
			awg.Done()
		}(a)
	}

	// Wait until all of the addons are enabled before updating the config (not thread safe)
	awg.Wait()
	for _, a := range toEnableList {
		if err := Set(cc, a, "true"); err != nil {
			glog.Errorf("store failed: %v", err)
		}
	}
}
