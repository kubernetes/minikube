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
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"k8s.io/klog/v2"
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

var defaultAddons = map[string]bool{
	"default-storageclass": true,
	"storage-provisioner":  true,
}

// AddonImage represents a single image in an addon
type AddonImage struct {
	name     string
	registry string
	image    string
}

// Name get the name of the addon image
func (a AddonImage) Name() string {
	return a.name
}

// Registry get the registry containing the image
func (a AddonImage) Registry() string {
	return a.registry
}

// Image get the path of the image
func (a AddonImage) Image() string {
	return a.image
}

// AddonPackage represents a whole addon
type AddonPackage struct {
	maintainer string
	name       string
	assets     []assets.Asset
	images     map[string]AddonImage
}

// Name get the addon name
func (a *AddonPackage) Name() string {
	return a.name
}

// Maintainer get the addon maintainer
func (a *AddonPackage) Maintainer() string {
	return a.maintainer
}

// GetImages gets a list of the addon images
func (a *AddonPackage) GetImages() map[string]AddonImage {
	return a.images
}

// GetAssets gets a list of the addon assets
func (a *AddonPackage) GetAssets() []assets.Asset {
	return a.assets
}

// IsEnabled checks if an Addon is enabled for the given profile
func (a *AddonPackage) IsEnabled(cc *config.ClusterConfig) bool {
	status, ok := cc.Addons[a.name]
	if ok {
		return status
	}

	// Return the default unconfigured state of the addon
	status, ok = defaultAddons[a.name]
	return ok && status
}

// WithImages overrides the images and returns a new AddonPackage
func (a AddonPackage) WithImages(images map[string]AddonImage) AddonPackage {
	return AddonPackage{
		name:       a.name,
		maintainer: a.maintainer,
		assets:     a.assets,
		images:     images,
	}
}

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

	// Run any callbacks for this property
	if err := run(cc, name, value, a.callbacks); err != nil {
		if errors.Is(err, ErrSkipThisAddon) {
			return err
		}
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
	addon := Addons[name]

	// check addon status before enabling/disabling it
	if isAddonAlreadySet(cc, addon, enable) {
		if addon.Name() == "gcp-auth" {
			return nil
		}
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

	cp, err := config.PrimaryControlPlane(cc)
	if err != nil {
		exit.Error(reason.GuestCpConfig, "Error getting primary control plane", err)
	}

	// maintain backwards compatibility for ingress and ingress-dns addons with k8s < v1.19
	if strings.HasPrefix(name, "ingress") && enable {
		if err := supportLegacyIngress(addon, *cc); err != nil {
			return err
		}
	}

	// Persist images even if the machine is running so starting gets the correct images.
	images, customRegistries, err := selectAndPersistImages(addon, cc)
	if err != nil {
		exit.Error(reason.HostSaveProfile, "Failed to persist images", err)
	}

	if cc.KubernetesConfig.ImageRepository == "registry.cn-hangzhou.aliyuncs.com/google_containers" {
		images, customRegistries = fixAddonImagesAndRegistries(addon, images, customRegistries)
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

	data := generateTemplateData(addon, cc.KubernetesConfig, networkInfo, images, customRegistries, enable)
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

	return false, nil
}

func isAddonAlreadySet(cc *config.ClusterConfig, addon *AddonPackage, enable bool) bool {
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
func supportLegacyIngress(addon *AddonPackage, cc config.ClusterConfig) error {
	v, err := util.ParseKubernetesVersion(cc.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kubernetes version")
	}
	if semver.MustParseRange("<1.19.0")(v) {
		if addon.Name() == "ingress" {
			addon.images = map[string]AddonImage{
				"IngressController": {
					name:     "IngressController",
					image:    "ingress-nginx/controller:v0.49.3@sha256:35fe394c82164efa8f47f3ed0be981b3f23da77175bbb8268a9ae438851c8324",
					registry: "",
				},
				"KubeWebhookCertgenCreate": {
					name:     "KubeWebhookCertgenCreate",
					image:    "jettech/kube-webhook-certgen:v1.5.1@sha256:950833e19ade18cd389d647efb88992a7cc077abedef343fa59e012d376d79b7",
					registry: "docker.io",
				},
				"KubeWebhookCertgenPatch": {
					name:     "KubeWebhookCertgenPatch",
					image:    "jettech/kube-webhook-certgen:v1.5.1@sha256:950833e19ade18cd389d647efb88992a7cc077abedef343fa59e012d376d79b7",
					registry: "",
				},
			}
			return nil
		}
		if addon.Name() == "ingress-dns" {
			addon.images = map[string]AddonImage{
				"IngressDNS": {
					name:     "IngressDNS",
					image:    "cryptexlabs/minikube-ingress-dns:0.3.0@sha256:e252d2a4c704027342b303cc563e95d2e71d2a0f1404f55d676390e28d5093ab",
					registry: "",
				},
			}
			return nil
		}
		return fmt.Errorf("supportLegacyIngress called for unexpected addon %q - nothing to do here", addon.Name())
	}

	return nil
}

func enableOrDisableAddonInternal(cc *config.ClusterConfig, add *AddonPackage, runner command.Runner, data interface{}, enable bool) error {
	deployFiles := []string{}

	for _, asset := range add.GetAssets() {
		var f assets.CopyableFile
		if asset.IsTemplate() {
			d, err := asset.Evaluate(data)
			if err != nil {
				return errors.Wrapf(err, "evaluate bundled addon %s asset", asset.GetSourcePath())
			}
			f = d
		} else {
			f = asset
		}

		fPath := asset.GetTargetPath()

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

	// Retry, because sometimes we race against an apiserver restart
	apply := func() error {
		_, err := runner.RunCmd(kubectlCommand(cc, deployFiles, enable))
		if err != nil {
			klog.Warningf("apply failed, will retry: %v", err)
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

// Start enables the default addons for a profile, plus any additional
func Start(wg *sync.WaitGroup, cc *config.ClusterConfig, toEnable map[string]bool, additional []string) {
	defer wg.Done()

	start := time.Now()
	klog.Infof("enableAddons start: toEnable=%v, additional=%s", toEnable, additional)
	defer func() {
		klog.Infof("enableAddons completed in %s", time.Since(start))
	}()

	// Get the default values of any addons not saved to our config
	for name, a := range Addons {
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

	var enabledAddons []string

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

	// Wait until all of the addons are enabled before updating the config (not thread safe)
	awg.Wait()

	for _, a := range enabledAddons {
		if err := Set(cc, a, "true"); err != nil {
			klog.Errorf("store failed: %v", err)
		}
	}
}

// GenerateTemplateData generates template data for template assets
func generateTemplateData(addon *AddonPackage, cfg config.KubernetesConfig, netInfo assets.NetworkInfo, images, customRegistries map[string]string, enable bool) interface{} {

	a := runtime.GOARCH
	// Some legacy docker images still need the -arch suffix
	// for  less common architectures blank suffix for amd64
	ea := ""
	if runtime.GOARCH != "amd64" {
		ea = "-" + runtime.GOARCH
	}

	var registries = make(map[string]string, len(addon.GetImages()))
	for _, image := range addon.GetImages() {
		registries[image.Name()] = image.Registry()
	}

	opts := struct {
		PreOneTwentyKubernetes bool
		Arch                   string
		ExoticArch             string
		ImageRepository        string
		LoadBalancerStartIP    string
		LoadBalancerEndIP      string
		CustomIngressCert      string
		IngressAPIVersion      string
		ContainerRuntime       string
		Images                 map[string]string
		Registries             map[string]string
		CustomRegistries       map[string]string
		NetworkInfo            map[string]string
	}{
		PreOneTwentyKubernetes: false,
		Arch:                   a,
		ExoticArch:             ea,
		ImageRepository:        cfg.ImageRepository,
		LoadBalancerStartIP:    cfg.LoadBalancerStartIP,
		LoadBalancerEndIP:      cfg.LoadBalancerEndIP,
		CustomIngressCert:      cfg.CustomIngressCert,
		IngressAPIVersion:      "v1", // api version for ingress (eg, "v1beta1"; defaults to "v1" for k8s 1.19+)
		ContainerRuntime:       cfg.ContainerRuntime,
		Images:                 images,
		Registries:             registries,
		CustomRegistries:       customRegistries,
		NetworkInfo:            make(map[string]string),
	}
	if opts.ImageRepository != "" && !strings.HasSuffix(opts.ImageRepository, "/") {
		opts.ImageRepository += "/"
	}
	if opts.Registries == nil {
		opts.Registries = make(map[string]string)
	}

	// maintain backwards compatibility with k8s < v1.19
	// by using v1beta1 instead of v1 api version for ingress
	v, err := util.ParseKubernetesVersion(cfg.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kubernetes version")
	}
	if semver.MustParseRange("<1.19.0")(v) {
		opts.IngressAPIVersion = "v1beta1"
	}
	if semver.MustParseRange("<1.20.0")(v) {
		opts.PreOneTwentyKubernetes = true
	}

	// Network info for generating template
	opts.NetworkInfo["ControlPlaneNodeIP"] = netInfo.ControlPlaneNodeIP
	opts.NetworkInfo["ControlPlaneNodePort"] = fmt.Sprint(netInfo.ControlPlaneNodePort)

	// Append postfix "/" to registries
	for k, v := range opts.Registries {
		if v != "" && !strings.HasSuffix(v, "/") {
			opts.Registries[k] = v + "/"
		}
	}

	for k, v := range opts.CustomRegistries {
		if v != "" && !strings.HasSuffix(v, "/") {
			opts.CustomRegistries[k] = v + "/"
		}
	}

	for name, image := range opts.Images {
		if _, ok := opts.Registries[name]; !ok {
			opts.Registries[name] = "" // Avoid nil access when rendering
		}

		if override, ok := opts.CustomRegistries[name]; ok {
			out.Infof("Using image {{.registry}}{{.image}}", out.V{
				"registry": override,
				// removing the SHA from UI
				// SHA example gcr.io/k8s-minikube/gcp-auth-webhook:v0.0.4@sha256:65e9e69022aa7b0eb1e390e1916e3bf67f75ae5c25987f9154ef3b0e8ab8528b
				"image": strings.Split(image, "@")[0],
			})
		} else if opts.ImageRepository != "" {
			out.Infof("Using image {{.registry}}{{.image}} (global image repository)", out.V{
				"registry": opts.ImageRepository,
				"image":    image,
			})
		} else {
			out.Infof("Using image {{.registry}}{{.image}}", out.V{
				"registry": opts.Registries[name],
				"image":    strings.Split(image, "@")[0],
			})
		}

		if enable {
			if override, ok := opts.CustomRegistries[name]; ok {
				out.Infof("Using image {{.registry}}{{.image}}", out.V{
					"registry": override,
					// removing the SHA from UI
					// SHA example gcr.io/k8s-minikube/gcp-auth-webhook:v0.0.4@sha256:65e9e69022aa7b0eb1e390e1916e3bf67f75ae5c25987f9154ef3b0e8ab8528b
					"image": strings.Split(image, "@")[0],
				})
			} else if opts.ImageRepository != "" {
				out.Infof("Using image {{.registry}}{{.image}} (global image repository)", out.V{
					"registry": opts.ImageRepository,
					"image":    image,
				})
			} else {
				out.Infof("Using image {{.registry}}{{.image}}", out.V{
					"registry": opts.Registries[name],
					"image":    strings.Split(image, "@")[0],
				})
			}
		}
	}
	return opts
}

// SelectAndPersistImages selects which images to use based on addon default images, previously persisted images, and newly requested images - which are then persisted for future enables.
func selectAndPersistImages(addon *AddonPackage, cc *config.ClusterConfig) (images, customRegistries map[string]string, err error) {
	addonDefaultImages := addon.GetImages()
	addonImages := make(map[string]string, len(addonDefaultImages))

	for _, image := range addonDefaultImages {
		addonImages[image.Name()] = image.Image()
	}

	// Use previously configured custom images.
	images = overrideDefaults(addonImages, cc.CustomAddonImages)
	if viper.IsSet(config.AddonImages) {
		// Parse the AddonImages flag if present.
		newImages := parseMapString(viper.GetString(config.AddonImages))
		for name, image := range newImages {
			if image == "" {
				out.WarningT("Ignoring empty custom image {{.name}}", out.V{"name": name})
				delete(newImages, name)
				continue
			}
			if _, ok := addonImages[name]; !ok {
				out.WarningT("Ignoring unknown custom image {{.name}}", out.V{"name": name})
			}
		}
		// Use newly configured custom images.
		images = overrideDefaults(addonImages, newImages)
		// Store custom addon images to be written.
		cc.CustomAddonImages = mergeMaps(cc.CustomAddonImages, images)
	}

	// Use previously configured custom registries.
	customRegistries = filterKeySpace(addonImages, cc.CustomAddonRegistries) // filter by images map because registry map may omit default registry.
	if viper.IsSet(config.AddonRegistries) {
		// Parse the AddonRegistries flag if present.
		customRegistries = parseMapString(viper.GetString(config.AddonRegistries))
		for name := range customRegistries {
			if _, ok := addonImages[name]; !ok { // check images map because registry map may omitted default registry
				out.WarningT("Ignoring unknown custom registry {{.name}}", out.V{"name": name})
				delete(customRegistries, name)
			}
		}
		// Since registry map may omit default registry, any previously set custom registries for these images must be cleared.
		for name := range addonImages {
			delete(cc.CustomAddonRegistries, name)
		}
		// Merge newly set registries into custom addon registries to be written.
		cc.CustomAddonRegistries = mergeMaps(cc.CustomAddonRegistries, customRegistries)
	}

	err = nil
	// If images or registries were specified, save the config afterward.
	if viper.IsSet(config.AddonImages) || viper.IsSet(config.AddonRegistries) {
		// Since these values are only set when a user enables an addon, it is safe to refer to the profile name.
		err = config.Write(viper.GetString(config.ProfileName), cc)
		// Whether err is nil or not we still return here.
	}
	return images, customRegistries, err
}
