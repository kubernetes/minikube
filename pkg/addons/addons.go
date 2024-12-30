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
	"context"
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blang/semver/v4"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
)

// Force is used to override checks for addons
var Force = false

// Refresh is used to refresh pods in specific cases when an addon is enabled
// Currently only used for gcp-auth
var Refresh = false

// ErrSkipThisAddon is a special error that tells us to not error out, but to also not mark the addon as enabled
var ErrSkipThisAddon = errors.New("skipping this addon")

// RunCallbacks runs all actions associated to an addon, but does not set it (thread-safe)
func RunCallbacks(cc *config.ClusterConfig, name string, value string) error {
	klog.Infof("Setting %s=%s in profile %q", name, value, cc.Name)
	a, valid := isAddonValid(name)
	if !valid {
		return errors.Errorf("%s is not a valid addon", name)
	}

	// Run any additional validations for this property
	if err := run(cc, name, value, a.validations); err != nil {
		if errors.Is(err, ErrSkipThisAddon) {
			return err
		}
		return errors.Wrap(err, "running validations")
	}

	preStartMessages(name, value)

	// Run any callbacks for this property
	if err := run(cc, name, value, a.callbacks); err != nil {
		if errors.Is(err, ErrSkipThisAddon) {
			return err
		}
		return errors.Wrap(err, "running callbacks")
	}

	postStartMessages(cc, name, value)

	return nil
}

func preStartMessages(name, value string) {
	if value != "true" {
		return
	}
	switch name {
	case "ambassador":
		out.Styled(style.Warning, "The ambassador addon has stopped working as of v1.23.0, for more details visit: https://github.com/datawire/ambassador-operator/issues/73")
	case "olm":
		out.Styled(style.Warning, "The OLM addon has stopped working, for more details visit: https://github.com/operator-framework/operator-lifecycle-manager/issues/2534")
	case "nvidia-gpu-device-plugin":
		out.Styled(style.Warning, "The nvidia-gpu-device-plugin addon is deprecated and it's functionality is merged inside of nvidia-device-plugin addon. It will be removed in a future release. Please use the nvidia-device-plugin addon instead. For more details, visit: https://github.com/kubernetes/minikube/issues/19114.")
	}
}

func postStartMessages(cc *config.ClusterConfig, name, value string) {
	if value != "true" {
		return
	}
	clusterName := cc.Name
	tipProfileArg := ""
	if clusterName != constants.DefaultClusterName {
		tipProfileArg = fmt.Sprintf(" -p %s", clusterName)
	}
	switch name {
	case "dashboard":
		out.Styled(style.Tip, `Some dashboard features require the metrics-server addon. To enable all features please run:

	minikube{{.profileArg}} addons enable metrics-server
`, out.V{"profileArg": tipProfileArg})
	case "headlamp":
		out.Styled(style.Tip, `To access Headlamp, use the following command:

	minikube{{.profileArg}} service headlamp -n headlamp
`, out.V{"profileArg": tipProfileArg})
		tokenGenerationTip := "To authenticate in Headlamp, fetch the Authentication Token using the following command:"
		createSvcAccountToken := "kubectl create token headlamp --duration 24h -n headlamp"
		getSvcAccountToken := `export SECRET=$(kubectl get secrets --namespace headlamp -o custom-columns=":metadata.name" | grep "headlamp-token")
kubectl get secret $SECRET --namespace headlamp --template=\{\{.data.token\}\} | base64 --decode`

		clusterVersion := cc.KubernetesConfig.KubernetesVersion
		parsedClusterVersion, err := util.ParseKubernetesVersion(clusterVersion)
		if err != nil {
			tokenGenerationTip = fmt.Sprintf("%s\nIf Kubernetes Version is <1.24:\n%s\n\nIf Kubernetes Version is >=1.24:\n%s\n", tokenGenerationTip, createSvcAccountToken, getSvcAccountToken)
		} else {
			if parsedClusterVersion.GTE(semver.Version{Major: 1, Minor: 24}) {
				tokenGenerationTip = fmt.Sprintf("%s\n\n        %s", tokenGenerationTip, createSvcAccountToken)
			} else {
				tokenGenerationTip = fmt.Sprintf("%s\n\n        %s", tokenGenerationTip, getSvcAccountToken)
			}
		}
		out.Styled(style.Tip, fmt.Sprintf("%s\n", tokenGenerationTip))
		out.Styled(style.Tip, `Headlamp can display more detailed information when metrics-server is installed. To install it, run:

	minikube{{.profileArg}} addons enable metrics-server
`, out.V{"profileArg": tipProfileArg})
	case "yakd":
		out.Styled(style.Tip, `To access YAKD - Kubernetes Dashboard, wait for Pod to be ready and run the following command:

	minikube{{.profileArg}} service yakd-dashboard -n yakd-dashboard
`, out.V{"profileArg": tipProfileArg})
	}
}

// Deprecations if the selected addon is deprecated return the replacement addon, otherwise return the passed in addon
func Deprecations(name string) (bool, string, string) {
	switch name {
	case "heapster":
		return true, "metrics-server", "using metrics-server addon, heapster is deprecated"
	case "efk":
		return true, "", "The current images used in the efk addon contain Log4j vulnerabilities, the addon will be disabled until images are updated, see: https://github.com/kubernetes/minikube/issues/15280"
	case "nvidia-gpu-device-plugin":
		return true, "nvidia-device-plugin", "The nvidia-gpu-device-plugin addon is deprecated and it's functionality is merged inside of nvidia-device-plugin addon. It will be removed in a future release. Please use the nvidia-device-plugin addon instead. For more details, visit: https://github.com/kubernetes/minikube/issues/19114."
	}
	return false, "", ""
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
		if errors.Is(err, ErrSkipThisAddon) {
			return err
		}
		return errors.Wrap(err, "run callbacks")
	}

	if err := Set(cc, name, value); err != nil {
		return errors.Wrap(err, "set")
	}

	klog.Infof("Writing out %q config to set %s=%v...", profile, name, value)
	return config.Write(profile, cc)
}

// Runs all the validation or callback functions and collects errors
func run(cc *config.ClusterConfig, name string, value string, fns []setFn) error {
	var errs []error
	for _, fn := range fns {
		err := fn(cc, name, value)
		if err != nil {
			if errors.Is(err, ErrSkipThisAddon) {
				return ErrSkipThisAddon
			}
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
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

// EnableOrDisableAddon updates addon status executing any commands necessary
func EnableOrDisableAddon(cc *config.ClusterConfig, name string, val string) error {
	klog.Infof("Setting addon %s=%s in %q", name, val, cc.Name)
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	addon := assets.Addons[name]

	// check addon status before enabling/disabling it
	if isAddonAlreadySet(cc, addon, enable) {
		klog.Warningf("addon %s should already be in state %v", name, val)
		if !enable {
			return nil
		}
	}

	api, err := machine.NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "machine client")
	}
	defer api.Close()

	cp, err := config.ControlPlane(*cc)
	if err != nil {
		exit.Error(reason.GuestCpConfig, "Error getting control-plane node", err)
	}

	// maintain backwards compatibility for ingress and ingress-dns addons with k8s < v1.19
	if strings.HasPrefix(name, "ingress") && enable {
		if err := supportLegacyIngress(addon, *cc); err != nil {
			return err
		}
	}

	// Persist images even if the machine is running so starting gets the correct images.
	images, customRegistries, err := assets.SelectAndPersistImages(addon, cc)
	if err != nil {
		exit.Error(reason.HostSaveProfile, "Failed to persist images", err)
	}

	if cc.KubernetesConfig.ImageRepository == constants.AliyunMirror {
		images, customRegistries = assets.FixAddonImagesAndRegistries(addon, images, customRegistries)
	}

	mName := config.MachineName(*cc, cp)
	host, err := machine.LoadHost(api, mName)
	if err != nil || !machine.IsRunning(api, mName) {
		klog.Warningf("%q is not running, setting %s=%v and skipping enablement (err=%v)", mName, addon.Name(), enable, err)
		return nil
	}

	runner, err := machine.CommandRunner(host)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	bail, err := addonSpecificChecks(cc, name, enable, runner)
	if err != nil {
		return err
	}
	if bail {
		return nil
	}

	var networkInfo assets.NetworkInfo
	if len(cc.Nodes) >= 1 {
		networkInfo.ControlPlaneNodeIP = cc.Nodes[0].IP
		networkInfo.ControlPlaneNodePort = cc.Nodes[0].Port
	} else {
		out.WarningT("At least needs control plane nodes to enable addon")
	}

	data := assets.GenerateTemplateData(addon, cc, networkInfo, images, customRegistries, enable)
	return enableOrDisableAddonInternal(cc, addon, runner, data, enable)
}

func addonSpecificChecks(cc *config.ClusterConfig, name string, enable bool, runner command.Runner) (bool, error) {
	// to match both ingress and ingress-dns addons
	if strings.HasPrefix(name, "ingress") && enable {
		if driver.IsKIC(cc.Driver) {
			if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
				out.Styled(style.Tip, `After the addon is enabled, please run "minikube tunnel" and your ingress resources would be available at "127.0.0.1"`)
			}
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

	if name == "registry" {
		if driver.NeedsPortForward(cc.Driver) {
			port, err := oci.ForwardedPort(cc.Driver, cc.Name, constants.RegistryAddonPort)
			if err != nil {
				return false, errors.Wrap(err, "registry port")
			}
			if enable {
				out.Boxed(`Registry addon with {{.driver}} driver uses port {{.port}} please use that instead of default port 5000`, out.V{"driver": cc.Driver, "port": port})
			}
			out.Styled(style.Documentation, `For more information see: https://minikube.sigs.k8s.io/docs/drivers/{{.driver}}`, out.V{"driver": cc.Driver})
		}
		return false, nil
	}

	if name == "auto-pause" && !enable { // needs to be disabled before deleting the service file in the internal disable
		if err := sysinit.New(runner).DisableNow("auto-pause"); err != nil {
			klog.ErrorS(err, "failed to disable", "service", "auto-pause")
		}
		return false, nil
	}

	// If the gcp-auth credentials haven't been mounted in, don't start the pods
	if name == "gcp-auth" && enable {
		rr, err := runner.RunCmd(exec.Command("cat", credentialsPath))
		if err != nil || rr.Stdout.String() == "" {
			return true, nil
		}
	}

	// we cannot use volcano for crio
	if name == "volcano" && cc.KubernetesConfig.ContainerRuntime == constants.CRIO && enable {
		return false, fmt.Errorf("volcano addon does not support crio")
	}

	return false, nil
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

// maintain backwards compatibility for ingress and ingress-dns addons with k8s < v1.19 by replacing default addons' images with compatible versions
func supportLegacyIngress(addon *assets.Addon, cc config.ClusterConfig) error {
	v, err := util.ParseKubernetesVersion(cc.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kubernetes version")
	}
	if semver.MustParseRange("<1.19.0")(v) {
		if addon.Name() == "ingress" {
			addon.Images = map[string]string{
				// https://github.com/kubernetes/ingress-nginx/blob/0a2ec01eb4ec0e1b29c4b96eb838a2e7bfe0e9f6/deploy/static/provider/kind/deploy.yaml#L328
				"IngressController": "ingress-nginx/controller:v0.49.3@sha256:35fe394c82164efa8f47f3ed0be981b3f23da77175bbb8268a9ae438851c8324",
				// issues: https://github.com/kubernetes/ingress-nginx/issues/7418 and https://github.com/jet/kube-webhook-certgen/issues/30
				"KubeWebhookCertgenCreate": "docker.io/jettech/kube-webhook-certgen:v1.5.1@sha256:950833e19ade18cd389d647efb88992a7cc077abedef343fa59e012d376d79b7",
				"KubeWebhookCertgenPatch":  "docker.io/jettech/kube-webhook-certgen:v1.5.1@sha256:950833e19ade18cd389d647efb88992a7cc077abedef343fa59e012d376d79b7",
			}
			addon.Registries = map[string]string{
				"IngressController": "registry.k8s.io",
			}
			return nil
		}
		if addon.Name() == "ingress-dns" {
			addon.Images = map[string]string{
				"IngressDNS": "cryptexlabs/minikube-ingress-dns:0.3.0@sha256:e252d2a4c704027342b303cc563e95d2e71d2a0f1404f55d676390e28d5093ab",
			}
			addon.Registries = nil
			return nil
		}
		return fmt.Errorf("supportLegacyIngress called for unexpected addon %q - nothing to do here", addon.Name())
	}

	return nil
}

func enableOrDisableAddonInternal(cc *config.ClusterConfig, addon *assets.Addon, runner command.Runner, data interface{}, enable bool) error {
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
			klog.Infof("installing %s", fPath)
			if err := runner.Copy(f); err != nil {
				return err
			}
		} else {
			klog.Infof("Removing %+v", fPath)
			defer func() {
				if err := runner.Remove(f); err != nil {
					klog.Warningf("error removing %s; addon should still be disabled as expected", fPath)
				}
			}()
		}
		if strings.HasSuffix(fPath, ".yaml") {
			deployFiles = append(deployFiles, fPath)
		}
	}

	// on the first attempt try without force, but on subsequent attempts use force
	force := false

	// Retry, because sometimes we race against an apiserver restart
	apply := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, err := runner.RunCmd(kubectlCommand(ctx, cc, deployFiles, enable, force))
		if err != nil {
			klog.Warningf("apply failed, will retry: %v", err)
			force = true
		}
		return err
	}

	return retry.Expo(apply, 250*time.Millisecond, 2*time.Minute)
}

func verifyAddonStatus(cc *config.ClusterConfig, name string, val string) error {
	ns := "kube-system"
	if name == "ingress" {
		ns = "ingress-nginx"
	}
	return verifyAddonStatusInternal(cc, name, val, ns)
}

func verifyAddonStatusInternal(cc *config.ClusterConfig, name string, val string, ns string) error {
	klog.Infof("Verifying addon %s=%s in %q", name, val, cc.Name)
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}

	label, ok := addonPodLabels[name]
	if ok && enable {
		out.Step(style.HealthCheck, "Verifying {{.addon_name}} addon...", out.V{"addon_name": name})
		client, err := kapi.Client(viper.GetString(config.ProfileName))
		if err != nil {
			return errors.Wrapf(err, "get kube-client to validate %s addon: %v", name, err)
		}

		// This timeout includes image pull time, which can take a few minutes. 3 is not enough.
		err = kapi.WaitForPods(client, ns, label, time.Minute*6)
		if err != nil {
			return errors.Wrapf(err, "waiting for %s pods", label)
		}

	}
	return nil
}

// Enable tries to enable the default addons for a profile plus any additional, and returns a single slice of all successfully enabled addons via channel (thread-safe).
// Since Enable is called asynchronously (so is not thread-safe for concurrent addons map updating/reading), to avoid race conditions,
// ToEnable should be called synchronously before Enable to get complete list of addons to enable, and
// UpdateConfig should be called synchronously after Enable to update the config with successfully enabled addons.
func Enable(wg *sync.WaitGroup, cc *config.ClusterConfig, toEnable map[string]bool, enabled chan<- []string) {
	defer wg.Done()

	start := time.Now()
	klog.Infof("enable addons start: toEnable=%v", toEnable)
	var enabledAddons []string
	defer func() {
		klog.Infof("duration metric: took %s for enable addons: enabled=%v", time.Since(start), enabledAddons)
	}()

	toEnableList := []string{}
	for k, v := range toEnable {
		if v {
			toEnableList = append(toEnableList, k)
		}
	}
	sort.Strings(toEnableList)

	var awg sync.WaitGroup

	defer func() { // making it show after verifications (see #7613)
		register.Reg.SetStep(register.EnablingAddons)
		out.Step(style.AddonEnable, "Enabled addons: {{.addons}}", out.V{"addons": strings.Join(enabledAddons, ", ")})
	}()
	for _, a := range toEnableList {
		awg.Add(1)
		go func(name string) {
			err := RunCallbacks(cc, name, "true")
			if err != nil && !errors.Is(err, ErrSkipThisAddon) {
				out.WarningT("Enabling '{{.name}}' returned an error: {{.error}}", out.V{"name": name, "error": err})
			} else {
				enabledAddons = append(enabledAddons, name)
			}
			awg.Done()
		}(a)
	}

	// Wait until all of the addons are enabled
	awg.Wait()

	// send the slice of all successfully enabled addons to channel and close
	enabled <- enabledAddons
	close(enabled)
}

// ToEnable returns the final list of addons to enable (not thread-safe).
func ToEnable(cc *config.ClusterConfig, existing map[string]bool, additional []string) map[string]bool {
	// start from existing
	enable := map[string]bool{}
	for k, v := range existing {
		enable[k] = v
	}

	// Get the default values of any addons not saved to our config
	for name, a := range assets.Addons {
		if _, exists := existing[name]; !exists {
			enable[name] = a.IsEnabledOrDefault(cc)
		}
	}

	// Apply new addons
	for _, name := range additional {
		isDeprecated, replacement, msg := Deprecations(name)
		if isDeprecated && replacement == "" {
			out.FailureT(msg)
			continue
		} else if isDeprecated {
			out.Styled(style.Waiting, msg)
			name = replacement
		}
		// if the specified addon doesn't exist, skip enabling
		if _, e := isAddonValid(name); e {
			enable[name] = true
		}
	}

	return enable
}

// UpdateConfigToEnable tries to update config with all enabled addons (not thread-safe).
// Any error will be logged and it will continue.
func UpdateConfigToEnable(cc *config.ClusterConfig, enabled []string) {
	for _, a := range enabled {
		if err := Set(cc, a, "true"); err != nil {
			klog.Errorf("store failed: %v", err)
		}
	}
}

func UpdateConfigToDisable(cc *config.ClusterConfig) {
	for name := range assets.Addons {
		if err := Set(cc, name, "false"); err != nil {
			klog.Errorf("store failed: %v", err)
		}
	}
}

// VerifyNotPaused verifies the cluster is not paused before enable/disable an addon.
func VerifyNotPaused(profile string, enable bool) error {
	klog.Info("checking whether the cluster is paused")

	cc, err := config.Load(profile)
	if err != nil {
		return errors.Wrap(err, "loading profile")
	}

	api, err := machine.NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "machine client")
	}
	defer api.Close()

	cp, err := config.ControlPlane(*cc)
	if err != nil {
		return errors.Wrap(err, "get control-plane node")
	}

	host, err := machine.LoadHost(api, config.MachineName(*cc, cp))
	if err != nil {
		return errors.Wrap(err, "get host")
	}

	s, err := host.Driver.GetState()
	if err != nil {
		return errors.Wrap(err, "get state")
	}
	if s != state.Running {
		// can't check the status of pods on a non-running cluster
		return nil
	}

	runner, err := machine.CommandRunner(host)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	crName := cc.KubernetesConfig.ContainerRuntime
	cr, err := cruntime.New(cruntime.Config{Type: crName, Runner: runner})
	if err != nil {
		return errors.Wrap(err, "container runtime")
	}
	runtimePaused, err := cluster.CheckIfPaused(cr, []string{"kube-system"})
	if err != nil {
		return errors.Wrap(err, "check paused")
	}
	if !runtimePaused {
		return nil
	}
	action := "disable"
	if enable {
		action = "enable"
	}
	msg := fmt.Sprintf("Can't %s addon on a paused cluster, please unpause the cluster first.", action)
	out.Styled(style.Shrug, msg)
	return errors.New(msg)
}
