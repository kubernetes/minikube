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

package kubeadm

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/util"
)

// enum to differentiate kubeadm command line parameters from kubeadm config file parameters (see the
// KubeadmExtraArgsWhitelist variable below for more info)
const (
	KubeadmCmdParam    = iota
	KubeadmConfigParam = iota
)

// KubeadmExtraArgsWhitelist is a whitelist of supported kubeadm params that can be supplied to kubeadm through
// minikube's ExtraArgs parameter. The list is split into two parts - params that can be supplied as flags on the
// command line and params that have to be inserted into the kubeadm config file. This is because of a kubeadm
// constraint which allows only certain params to be provided from the command line when the --config parameter
// is specified
var KubeadmExtraArgsWhitelist = map[int][]string{
	KubeadmCmdParam: {
		"ignore-preflight-errors",
		"dry-run",
		"kubeconfig",
		"kubeconfig-dir",
		"node-name",
		"cri-socket",
		"experimental-upload-certs",
		"certificate-key",
		"rootfs",
	},
	KubeadmConfigParam: {
		"pod-network-cidr",
	},
}

type pod struct {
	// Human friendly name
	name  string
	key   string
	value string
}

// PodsByLayer are queries we run when health checking, sorted roughly by dependency layer
var PodsByLayer = []pod{
	{"proxy", "k8s-app", "kube-proxy"},
	{"etcd", "component", "etcd"},
	{"scheduler", "component", "kube-scheduler"},
	{"controller", "component", "kube-controller-manager"},
	{"dns", "k8s-app", "kube-dns"},
}

// SkipAdditionalPreflights are additional preflights we skip depending on the runtime in use.
var SkipAdditionalPreflights = map[string][]string{}

// Bootstrapper is a bootstrapper using kubeadm
type Bootstrapper struct {
	c bootstrapper.CommandRunner
}

// NewKubeadmBootstrapper creates a new kubeadm.Bootstrapper
func NewKubeadmBootstrapper(api libmachine.API) (*Bootstrapper, error) {
	h, err := api.Load(config.GetMachineName())
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return nil, errors.Wrap(err, "command runner")
	}
	return &Bootstrapper{c: runner}, nil
}

// GetKubeletStatus returns the kubelet status
func (k *Bootstrapper) GetKubeletStatus() (string, error) {
	statusCmd := `sudo systemctl is-active kubelet`
	status, err := k.c.CombinedOutput(statusCmd)
	if err != nil {
		return "", errors.Wrap(err, "getting status")
	}
	s := strings.TrimSpace(status)
	switch s {
	case "active":
		return state.Running.String(), nil
	case "inactive":
		return state.Stopped.String(), nil
	case "activating":
		return state.Starting.String(), nil
	}
	return state.Error.String(), nil
}

// GetAPIServerStatus returns the api-server status
func (k *Bootstrapper) GetAPIServerStatus(ip net.IP, apiserverPort int) (string, error) {
	url := fmt.Sprintf("https://%s:%d/healthz", ip, apiserverPort)
	// To avoid: x509: certificate signed by unknown authority
	tr := &http.Transport{
		Proxy:           nil, // To avoid connectiv issue if http(s)_proxy is set.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	glog.Infof("%s response: %v %+v", url, err, resp)
	// Connection refused, usually.
	if err != nil {
		return state.Stopped.String(), nil
	}
	if resp.StatusCode != http.StatusOK {
		return state.Error.String(), nil
	}
	return state.Running.String(), nil
}

// LogCommands returns a map of log type to a command which will display that log.
func (k *Bootstrapper) LogCommands(o bootstrapper.LogOptions) map[string]string {
	var kubelet strings.Builder
	kubelet.WriteString("journalctl -u kubelet")
	if o.Lines > 0 {
		kubelet.WriteString(fmt.Sprintf(" -n %d", o.Lines))
	}
	if o.Follow {
		kubelet.WriteString(" -f")
	}

	var dmesg strings.Builder
	dmesg.WriteString("sudo dmesg -PH -L=never --level warn,err,crit,alert,emerg")
	if o.Follow {
		dmesg.WriteString(" --follow")
	}
	if o.Lines > 0 {
		dmesg.WriteString(fmt.Sprintf(" | tail -n %d", o.Lines))
	}
	return map[string]string{
		"kubelet": kubelet.String(),
		"dmesg":   dmesg.String(),
	}
}

// createFlagsFromExtraArgs converts kubeadm extra args into flags to be supplied from the commad linne
func createFlagsFromExtraArgs(extraOptions util.ExtraOptionSlice) string {
	kubeadmExtraOpts := extraOptions.AsMap().Get(Kubeadm)

	// kubeadm allows only a small set of parameters to be supplied from the command line when the --config param
	// is specified, here we remove those that are not allowed
	for opt := range kubeadmExtraOpts {
		if !util.ContainsString(KubeadmExtraArgsWhitelist[KubeadmCmdParam], opt) {
			// kubeadmExtraOpts is a copy so safe to delete
			delete(kubeadmExtraOpts, opt)
		}
	}
	return convertToFlags(kubeadmExtraOpts)
}

// StartCluster starts the cluster
func (k *Bootstrapper) StartCluster(k8s config.KubernetesConfig) error {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	extraFlags := createFlagsFromExtraArgs(k8s.ExtraOptions)

	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime})
	if err != nil {
		return err
	}

	ignore := []string{
		"DirAvailable--etc-kubernetes-manifests", // Addons are stored in /etc/kubernetes/manifests
		"DirAvailable--data-minikube",
		"FileAvailable--etc-kubernetes-manifests-kube-scheduler.yaml",
		"FileAvailable--etc-kubernetes-manifests-kube-apiserver.yaml",
		"FileAvailable--etc-kubernetes-manifests-kube-controller-manager.yaml",
		"FileAvailable--etc-kubernetes-manifests-etcd.yaml",
		"Port-10250", // For "none" users who already have a kubelet online
		"Swap",       // For "none" users who have swap configured
	}
	ignore = append(ignore, SkipAdditionalPreflights[r.Name()]...)

	// Allow older kubeadm versions to function with newer Docker releases.
	if version.LT(semver.MustParse("1.13.0")) {
		glog.Infof("Older Kubernetes release detected (%s), disabling SystemVerification check.", version)
		ignore = append(ignore, "SystemVerification")
	}

	cmd := fmt.Sprintf("sudo /usr/bin/kubeadm init --config %s %s --ignore-preflight-errors=%s",
		constants.KubeadmConfigFile, extraFlags, strings.Join(ignore, ","))
	out, err := k.c.CombinedOutput(cmd)
	if err != nil {
		return errors.Wrapf(err, "cmd failed: %s\n%s\n", cmd, out)
	}

	if version.LT(semver.MustParse("1.10.0-alpha.0")) {
		//TODO(r2d4): get rid of global here
		master = k8s.NodeName
		if err := util.RetryAfter(200, unmarkMaster, time.Second*1); err != nil {
			return errors.Wrap(err, "timed out waiting to unmark master")
		}
	}

	glog.Infof("Configuring cluster permissions ...")
	if err := util.RetryAfter(100, elevateKubeSystemPrivileges, time.Millisecond*500); err != nil {
		return errors.Wrap(err, "timed out waiting to elevate kube-system RBAC privileges")
	}

	if err := k.adjustResourceLimits(); err != nil {
		glog.Warningf("unable to adjust resource limits: %v", err)
	}
	return nil
}

// adjustResourceLimits makes fine adjustments to pod resources that aren't possible via kubeadm config.
func (k *Bootstrapper) adjustResourceLimits() error {
	score, err := k.c.CombinedOutput("cat /proc/$(pgrep kube-apiserver)/oom_adj")
	if err != nil {
		return errors.Wrap(err, "oom_adj check")
	}
	glog.Infof("apiserver oom_adj: %s", score)
	// oom_adj is already a negative number
	if strings.HasPrefix(score, "-") {
		return nil
	}
	glog.Infof("adjusting apiserver oom_adj to -10")

	// Prevent the apiserver from OOM'ing before other pods, as it is our gateway into the cluster.
	// It'd be preferable to do this via Kubernetes, but kubeadm doesn't have a way to set pod QoS.
	if err := k.c.Run("echo -10 | sudo tee /proc/$(pgrep kube-apiserver)/oom_adj"); err != nil {
		return errors.Wrap(err, "oom_adj adjust")
	}
	return nil
}

func addAddons(files *[]assets.CopyableFile, data interface{}) error {
	// add addons to file list
	// custom addons
	if err := assets.AddMinikubeDirAssets(files); err != nil {
		return errors.Wrap(err, "adding minikube dir assets")
	}
	// bundled addons
	for _, addonBundle := range assets.Addons {
		if isEnabled, err := addonBundle.IsEnabled(); err == nil && isEnabled {
			for _, addon := range addonBundle.Assets {
				if addon.IsTemplate() {
					addonFile, err := addon.Evaluate(data)
					if err != nil {
						return errors.Wrapf(err, "evaluate bundled addon %s asset", addon.GetAssetName())
					}

					*files = append(*files, addonFile)
				} else {
					*files = append(*files, addon)
				}
			}
		} else if err != nil {
			return nil
		}
	}

	return nil
}

// WaitCluster blocks until Kubernetes appears to be healthy.
func (k *Bootstrapper) WaitCluster(k8s config.KubernetesConfig) error {
	// Do not wait for "k8s-app" pods in the case of CNI, as they are managed
	// by a CNI plugin which is usually started after minikube has been brought
	// up. Otherwise, minikube won't start, as "k8s-app" pods are not ready.
	componentsOnly := k8s.NetworkPlugin == "cni"
	console.OutStyle("waiting-pods", "Verifying:")
	client, err := util.GetClient()
	if err != nil {
		return errors.Wrap(err, "k8s client")
	}

	// Wait until the apiserver can answer queries properly. We don't care if the apiserver
	// pod shows up as registered, but need the webserver for all subsequent queries.
	console.Out(" apiserver")
	if err := k.waitForAPIServer(k8s); err != nil {
		return errors.Wrap(err, "waiting for apiserver")
	}

	for _, p := range PodsByLayer {
		if componentsOnly && p.key != "component" {
			continue
		}

		console.Out(" %s", p.name)
		selector := labels.SelectorFromSet(labels.Set(map[string]string{p.key: p.value}))
		if err := util.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
			return errors.Wrap(err, fmt.Sprintf("waiting for %s=%s", p.key, p.value))
		}
	}
	console.OutLn("")
	return nil
}

// RestartCluster restarts the Kubernetes cluster configured by kubeadm
func (k *Bootstrapper) RestartCluster(k8s config.KubernetesConfig) error {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	phase := "alpha"
	controlPlane := "controlplane"
	if version.GTE(semver.MustParse("1.13.0")) {
		phase = "init"
		controlPlane = "control-plane"
	}

	configPath := constants.KubeadmConfigFile
	baseCmd := fmt.Sprintf("sudo kubeadm %s", phase)
	cmds := []string{
		fmt.Sprintf("%s phase certs all --config %s", baseCmd, configPath),
		fmt.Sprintf("%s phase kubeconfig all --config %s", baseCmd, configPath),
		fmt.Sprintf("%s phase %s all --config %s", baseCmd, controlPlane, configPath),
		fmt.Sprintf("%s phase etcd local --config %s", baseCmd, configPath),
	}

	// Run commands one at a time so that it is easier to root cause failures.
	for _, cmd := range cmds {
		if err := k.c.Run(cmd); err != nil {
			return errors.Wrapf(err, "running cmd: %s", cmd)
		}
	}

	if err := k.waitForAPIServer(k8s); err != nil {
		return errors.Wrap(err, "waiting for apiserver")
	}
	// restart the proxy and coredns
	if err := k.c.Run(fmt.Sprintf("%s phase addon all --config %s", baseCmd, configPath)); err != nil {
		return errors.Wrapf(err, "addon phase")
	}

	if err := k.adjustResourceLimits(); err != nil {
		glog.Warningf("unable to adjust resource limits: %v", err)
	}
	return nil
}

// waitForAPIServer waits for the apiserver to start up
func (k *Bootstrapper) waitForAPIServer(k8s config.KubernetesConfig) error {
	glog.Infof("Waiting for apiserver ...")
	return wait.PollImmediate(time.Millisecond*200, time.Minute*1, func() (bool, error) {
		status, err := k.GetAPIServerStatus(net.ParseIP(k8s.NodeIP), k8s.NodePort)
		glog.Infof("status: %s, err: %v", status, err)
		if err != nil {
			return false, err
		}
		if status != "Running" {
			return false, nil
		}
		return true, nil
	})
}

// DeleteCluster removes the components that were started earlier
func (k *Bootstrapper) DeleteCluster(k8s config.KubernetesConfig) error {
	cmd := fmt.Sprintf("sudo kubeadm reset --force")
	out, err := k.c.CombinedOutput(cmd)
	if err != nil {
		return errors.Wrapf(err, "kubeadm reset: %s\n%s\n", cmd, out)
	}

	return nil
}

// PullImages downloads images that will be used by RestartCluster
func (k *Bootstrapper) PullImages(k8s config.KubernetesConfig) error {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}
	if version.LT(semver.MustParse("1.11.0")) {
		return fmt.Errorf("pull command is not supported by kubeadm v%s", version)
	}

	cmd := fmt.Sprintf("sudo kubeadm config images pull --config %s", constants.KubeadmConfigFile)
	if err := k.c.Run(cmd); err != nil {
		return errors.Wrapf(err, "running cmd: %s", cmd)
	}
	return nil
}

// SetupCerts sets up certificates within the cluster.
func (k *Bootstrapper) SetupCerts(k8s config.KubernetesConfig) error {
	return bootstrapper.SetupCerts(k.c, k8s)
}

// NewKubeletConfig generates a new systemd unit containing a configured kubelet
// based on the options present in the KubernetesConfig.
func NewKubeletConfig(k8s config.KubernetesConfig, r cruntime.Manager) (string, error) {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return "", errors.Wrap(err, "parsing kubernetes version")
	}

	extraOpts, err := ExtraConfigForComponent(Kubelet, k8s.ExtraOptions, version)
	if err != nil {
		return "", errors.Wrap(err, "generating extra configuration for kubelet")
	}

	for k, v := range r.KubeletOptions() {
		extraOpts[k] = v
	}
	if k8s.NetworkPlugin != "" {
		extraOpts["network-plugin"] = k8s.NetworkPlugin
	}

	podInfraContainerImage, _ := constants.GetKubeadmCachedImages(k8s.ImageRepository, k8s.KubernetesVersion)
	if _, ok := extraOpts["pod-infra-container-image"]; !ok && k8s.ImageRepository != "" && podInfraContainerImage != "" {
		extraOpts["pod-infra-container-image"] = podInfraContainerImage
	}

	// parses a map of the feature gates for kubelet
	_, kubeletFeatureArgs, err := ParseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return "", errors.Wrap(err, "parses feature gate config for kubelet")
	}

	if kubeletFeatureArgs != "" {
		extraOpts["feature-gates"] = kubeletFeatureArgs
	}

	extraFlags := convertToFlags(extraOpts)

	b := bytes.Buffer{}
	opts := struct {
		ExtraOptions     string
		ContainerRuntime string
	}{
		ExtraOptions:     extraFlags,
		ContainerRuntime: k8s.ContainerRuntime,
	}
	if err := kubeletSystemdTemplate.Execute(&b, opts); err != nil {
		return "", err
	}

	return b.String(), nil
}

// UpdateCluster updates the cluster
func (k *Bootstrapper) UpdateCluster(cfg config.KubernetesConfig) error {
	_, images := constants.GetKubeadmCachedImages(cfg.ImageRepository, cfg.KubernetesVersion)
	if cfg.ShouldLoadCachedImages {
		if err := machine.LoadImages(k.c, images, constants.ImageCacheDir); err != nil {
			console.Failure("Unable to load cached images: %v", err)
		}
	}
	r, err := cruntime.New(cruntime.Config{Type: cfg.ContainerRuntime, Socket: cfg.CRISocket})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}
	kubeadmCfg, err := generateConfig(cfg, r)
	if err != nil {
		return errors.Wrap(err, "generating kubeadm cfg")
	}

	kubeletCfg, err := NewKubeletConfig(cfg, r)
	if err != nil {
		return errors.Wrap(err, "generating kubelet config")
	}
	glog.Infof("kubelet %s config:\n%s", cfg.KubernetesVersion, kubeletCfg)

	var files []assets.CopyableFile
	files = copyConfig(cfg, files, kubeadmCfg, kubeletCfg)

	if err := downloadBinaries(cfg, k.c); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}

	if err := addAddons(&files, assets.GenerateTemplateData(cfg)); err != nil {
		return errors.Wrap(err, "adding addons")
	}

	for _, f := range files {
		if err := k.c.Copy(f); err != nil {
			return errors.Wrapf(err, "copy")
		}
	}
	err = k.c.Run(`
sudo systemctl daemon-reload &&
sudo systemctl start kubelet
`)
	if err != nil {
		return errors.Wrap(err, "starting kubelet")
	}

	return nil
}

// createExtraComponentConfig generates a map of component to extra args for all of the components except kubeadm
func createExtraComponentConfig(extraOptions util.ExtraOptionSlice, version semver.Version, componentFeatureArgs string) ([]ComponentExtraArgs, error) {
	extraArgsSlice, err := NewComponentExtraArgs(extraOptions, version, componentFeatureArgs)
	if err != nil {
		return nil, err
	}

	// kubeadm extra args should not be included in the kubeadm config in the extra args section (instead, they must
	// be inserted explicitly in the appropriate places or supplied from the command line); here we remove all of the
	// kubeadm extra args from the slice
	for i, extraArgs := range extraArgsSlice {
		if extraArgs.Component == Kubeadm {
			extraArgsSlice = append(extraArgsSlice[:i], extraArgsSlice[i+1:]...)
			break
		}
	}
	return extraArgsSlice, nil
}

func generateConfig(k8s config.KubernetesConfig, r cruntime.Manager) (string, error) {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return "", errors.Wrap(err, "parsing kubernetes version")
	}

	// parses a map of the feature gates for kubeadm and component
	kubeadmFeatureArgs, componentFeatureArgs, err := ParseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return "", errors.Wrap(err, "parses feature gate config for kubeadm and component")
	}

	extraComponentConfig, err := createExtraComponentConfig(k8s.ExtraOptions, version, componentFeatureArgs)
	if err != nil {
		return "", errors.Wrap(err, "generating extra component config for kubeadm")
	}

	// In case of no port assigned, use util.APIServerPort
	nodePort := k8s.NodePort
	if nodePort <= 0 {
		nodePort = util.APIServerPort
	}

	opts := struct {
		CertDir           string
		ServiceCIDR       string
		PodSubnet         string
		AdvertiseAddress  string
		APIServerPort     int
		KubernetesVersion string
		EtcdDataDir       string
		NodeName          string
		CRISocket         string
		ImageRepository   string
		ExtraArgs         []ComponentExtraArgs
		FeatureArgs       map[string]bool
		NoTaintMaster     bool
	}{
		CertDir:           util.DefaultCertPath,
		ServiceCIDR:       util.DefaultServiceCIDR,
		PodSubnet:         k8s.ExtraOptions.Get("pod-network-cidr", Kubeadm),
		AdvertiseAddress:  k8s.NodeIP,
		APIServerPort:     nodePort,
		KubernetesVersion: k8s.KubernetesVersion,
		EtcdDataDir:       "/data/minikube", //TODO(r2d4): change to something else persisted
		NodeName:          k8s.NodeName,
		CRISocket:         r.SocketPath(),
		ImageRepository:   k8s.ImageRepository,
		ExtraArgs:         extraComponentConfig,
		FeatureArgs:       kubeadmFeatureArgs,
		NoTaintMaster:     false, // That does not work with k8s 1.12+
	}

	if k8s.ServiceCIDR != "" {
		opts.ServiceCIDR = k8s.ServiceCIDR
	}

	if version.GTE(semver.MustParse("1.10.0-alpha.0")) {
		opts.NoTaintMaster = true
	}

	b := bytes.Buffer{}
	configTmpl := configTmplV1Alpha1
	if version.GTE(semver.MustParse("1.12.0")) {
		configTmpl = configTmplV1Alpha3
	}
	// v1beta1 works in v1.13, but isn't required until v1.14.
	if version.GTE(semver.MustParse("1.14.0-alpha.0")) {
		configTmpl = configTmplV1Beta1
	}
	if err := configTmpl.Execute(&b, opts); err != nil {
		return "", err
	}

	return b.String(), nil
}

func copyConfig(cfg config.KubernetesConfig, files []assets.CopyableFile, kubeadmCfg string, kubeletCfg string) []assets.CopyableFile {
	files = append(files,
		assets.NewMemoryAssetTarget([]byte(kubeletService), constants.KubeletServiceFile, "0640"),
		assets.NewMemoryAssetTarget([]byte(kubeletCfg), constants.KubeletSystemdConfFile, "0640"),
		assets.NewMemoryAssetTarget([]byte(kubeadmCfg), constants.KubeadmConfigFile, "0640"))

	// Copy the default CNI config (k8s.conf), so that kubelet can successfully
	// start a Pod in the case a user hasn't manually installed any CNI plugin
	// and minikube was started with "--extra-config=kubelet.network-plugin=cni".
	if cfg.EnableDefaultCNI {
		files = append(files,
			assets.NewMemoryAssetTarget([]byte(defaultCNIConfig), constants.DefaultCNIConfigPath, "0644"),
			assets.NewMemoryAssetTarget([]byte(defaultCNIConfig), constants.DefaultRktNetConfigPath, "0644"))
	}

	return files
}

func downloadBinaries(cfg config.KubernetesConfig, c bootstrapper.CommandRunner) error {
	var g errgroup.Group
	for _, bin := range constants.GetKubeadmCachedBinaries() {
		bin := bin
		g.Go(func() error {
			path, err := machine.CacheBinary(bin, cfg.KubernetesVersion, "linux", runtime.GOARCH)
			if err != nil {
				return errors.Wrapf(err, "downloading %s", bin)
			}
			err = machine.CopyBinary(c, bin, path)
			if err != nil {
				return errors.Wrapf(err, "copying %s", bin)
			}
			return nil
		})
	}
	return g.Wait()
}
