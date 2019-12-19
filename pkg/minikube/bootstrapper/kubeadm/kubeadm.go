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
	"os/exec"

	"fmt"
	"net"
	"net/http"

	// WARNING: Do not use path/filepath in this package unless you want bizarre Windows paths
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
)

// enum to differentiate kubeadm command line parameters from kubeadm config file parameters (see the
// KubeadmExtraArgsWhitelist variable below for more info)
const (
	KubeadmCmdParam        = iota
	KubeadmConfigParam     = iota
	defaultCNIConfigPath   = "/etc/cni/net.d/k8s.conf"
	kubeletServiceFile     = "/lib/systemd/system/kubelet.service"
	kubeletSystemdConfFile = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
)

const (
	// Container runtimes
	remoteContainerRuntime = "remote"
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

// remote artifacts that must exist for minikube to function properly. The sign of a previously working installation.
// NOTE: /etc is not persistent across restarts, so don't bother checking there
var expectedArtifacts = []string{
	"/var/lib/kubelet/kubeadm-flags.env",
	"/var/lib/kubelet/config.yaml",
	etcdDataDir(),
}

// yamlConfigPath is the path to the kubeadm configuration
var yamlConfigPath = path.Join(vmpath.GuestEphemeralDir, "kubeadm.yaml")

// SkipAdditionalPreflights are additional preflights we skip depending on the runtime in use.
var SkipAdditionalPreflights = map[string][]string{}

// Bootstrapper is a bootstrapper using kubeadm
type Bootstrapper struct {
	c           command.Runner
	contextName string
}

// NewKubeadmBootstrapper creates a new kubeadm.Bootstrapper
func NewKubeadmBootstrapper(api libmachine.API) (*Bootstrapper, error) {
	name := viper.GetString(config.MachineProfile)
	h, err := api.Load(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return nil, errors.Wrap(err, "command runner")
	}
	return &Bootstrapper{c: runner, contextName: name}, nil
}

// GetKubeletStatus returns the kubelet status
func (k *Bootstrapper) GetKubeletStatus() (string, error) {
	rr, err := k.c.RunCmd(exec.Command("sudo", "systemctl", "is-active", "kubelet"))
	if err != nil {
		return "", errors.Wrapf(err, "getting kublet status. command: %q", rr.Command())
	}
	s := strings.TrimSpace(rr.Stdout.String())
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
	// Connection refused, usually.
	if err != nil {
		return state.Stopped.String(), nil
	}
	if resp.StatusCode != http.StatusOK {
		glog.Warningf("%s response: %v %+v", url, err, resp)
		return state.Error.String(), nil
	}
	return state.Running.String(), nil
}

// LogCommands returns a map of log type to a command which will display that log.
func (k *Bootstrapper) LogCommands(o bootstrapper.LogOptions) map[string]string {
	var kubelet strings.Builder
	kubelet.WriteString("sudo journalctl -u kubelet")
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

// createFlagsFromExtraArgs converts kubeadm extra args into flags to be supplied from the command linne
func createFlagsFromExtraArgs(extraOptions config.ExtraOptionSlice) string {
	kubeadmExtraOpts := extraOptions.AsMap().Get(Kubeadm)

	// kubeadm allows only a small set of parameters to be supplied from the command line when the --config param
	// is specified, here we remove those that are not allowed
	for opt := range kubeadmExtraOpts {
		if !config.ContainsParam(KubeadmExtraArgsWhitelist[KubeadmCmdParam], opt) {
			// kubeadmExtraOpts is a copy so safe to delete
			delete(kubeadmExtraOpts, opt)
		}
	}
	return convertToFlags(kubeadmExtraOpts)
}

// etcdDataDir is where etcd data is stored.
func etcdDataDir() string {
	return path.Join(vmpath.GuestPersistentDir, "etcd")
}

// createCompatSymlinks creates compatibility symlinks to transition running services to new directory structures
func (k *Bootstrapper) createCompatSymlinks() error {
	legacyEtcd := "/data/minikube"

	if _, err := k.c.RunCmd(exec.Command("sudo", "test", "-d", legacyEtcd)); err != nil {
		glog.Infof("%s skipping compat symlinks: %v", legacyEtcd, err)
		return nil
	}
	glog.Infof("Found %s, creating compatibility symlinks ...", legacyEtcd)

	c := exec.Command("sudo", "ln", "-s", legacyEtcd, etcdDataDir())
	if rr, err := k.c.RunCmd(c); err != nil {
		return errors.Wrapf(err, "create symlink failed: %s", rr.Command())
	}
	return nil
}

func (k *Bootstrapper) existingConfig() error {
	args := append([]string{"ls"}, expectedArtifacts...)
	_, err := k.c.RunCmd(exec.Command("sudo", args...))
	return err
}

// StartCluster starts the cluster
func (k *Bootstrapper) StartCluster(k8s config.KubernetesConfig) error {
	err := k.existingConfig()
	if err == nil {
		return k.restartCluster(k8s)
	}
	glog.Infof("existence check: %v", err)

	start := time.Now()
	glog.Infof("StartCluster: %+v", k8s)
	defer func() {
		glog.Infof("StartCluster complete in %s", time.Since(start))
	}()

	version, err := parseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	extraFlags := createFlagsFromExtraArgs(k8s.ExtraOptions)
	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime})
	if err != nil {
		return err
	}

	ignore := []string{
		fmt.Sprintf("DirAvailable-%s", strings.Replace(vmpath.GuestManifestsDir, "/", "-", -1)),
		fmt.Sprintf("DirAvailable-%s", strings.Replace(vmpath.GuestPersistentDir, "/", "-", -1)),
		fmt.Sprintf("DirAvailable-%s", strings.Replace(etcdDataDir(), "/", "-", -1)),
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

	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s init --config %s %s --ignore-preflight-errors=%s", invokeKubeadm(k8s.KubernetesVersion), yamlConfigPath, extraFlags, strings.Join(ignore, ",")))
	if rr, err := k.c.RunCmd(c); err != nil {
		return errors.Wrapf(err, "init failed. cmd: %q", rr.Command())
	}

	glog.Infof("Configuring cluster permissions ...")

	elevate := func() error {
		client, err := k.client(k8s)
		if err != nil {
			return err
		}
		return elevateKubeSystemPrivileges(client)
	}

	if err := retry.Expo(elevate, time.Millisecond*500, 120*time.Second); err != nil {
		return errors.Wrap(err, "timed out waiting to elevate kube-system RBAC privileges")
	}

	if err := k.adjustResourceLimits(); err != nil {
		glog.Warningf("unable to adjust resource limits: %v", err)
	}

	return nil
}

// adjustResourceLimits makes fine adjustments to pod resources that aren't possible via kubeadm config.
func (k *Bootstrapper) adjustResourceLimits() error {
	rr, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", "cat /proc/$(pgrep kube-apiserver)/oom_adj"))
	if err != nil {
		return errors.Wrapf(err, "oom_adj check cmd %s. ", rr.Command())
	}
	glog.Infof("apiserver oom_adj: %s", rr.Stdout.String())
	// oom_adj is already a negative number
	if strings.HasPrefix(rr.Stdout.String(), "-") {
		return nil
	}
	glog.Infof("adjusting apiserver oom_adj to -10")

	// Prevent the apiserver from OOM'ing before other pods, as it is our gateway into the cluster.
	// It'd be preferable to do this via Kubernetes, but kubeadm doesn't have a way to set pod QoS.
	if _, err = k.c.RunCmd(exec.Command("/bin/bash", "-c", "echo -10 | sudo tee /proc/$(pgrep kube-apiserver)/oom_adj")); err != nil {
		return errors.Wrap(err, fmt.Sprintf("oom_adj adjust"))
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

// client returns a Kubernetes client to use to speak to a kubeadm launched apiserver
func (k *Bootstrapper) client(k8s config.KubernetesConfig) (*kubernetes.Clientset, error) {
	config, err := kapi.ClientConfig(k.contextName)
	if err != nil {
		return nil, errors.Wrap(err, "client config")
	}

	endpoint := fmt.Sprintf("https://%s:%d", k8s.NodeIP, k8s.NodePort)
	if config.Host != endpoint {
		glog.Errorf("Overriding stale ClientConfig host %s with %s", config.Host, endpoint)
		config.Host = endpoint
	}

	return kubernetes.NewForConfig(config)
}

func (k *Bootstrapper) waitForAPIServerProcess(start time.Time, timeout time.Duration) error {
	glog.Infof("waiting for apiserver process to appear ...")
	err := wait.PollImmediate(time.Second*1, timeout, func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during process check")
		}
		rr, ierr := k.c.RunCmd(exec.Command("sudo", "pgrep", "kube-apiserver"))
		if ierr != nil {
			glog.Warningf("pgrep apiserver: %v cmd: %s", ierr, rr.Command())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("apiserver process never appeared")
	}
	glog.Infof("duration metric: took %s to wait for apiserver process to appear ...", time.Since(start))
	return nil
}

func (k *Bootstrapper) waitForAPIServerHealthz(start time.Time, k8s config.KubernetesConfig, timeout time.Duration) error {
	glog.Infof("waiting for apiserver healthz status ...")
	hStart := time.Now()
	healthz := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during healthz check")
		}

		status, err := k.GetAPIServerStatus(net.ParseIP(k8s.NodeIP), k8s.NodePort)
		if err != nil {
			glog.Warningf("status: %v", err)
			return false, nil
		}
		if status != "Running" {
			return false, nil
		}
		return true, nil
	}

	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, healthz); err != nil {
		return fmt.Errorf("apiserver healthz never reported healthy")
	}
	glog.Infof("duration metric: took %s to wait for apiserver healthz status ...", time.Since(hStart))
	return nil
}

func (k *Bootstrapper) waitForSystemPods(start time.Time, k8s config.KubernetesConfig, timeout time.Duration) error {
	glog.Infof("waiting for kube-system pods to appear ...")
	pStart := time.Now()
	client, err := k.client(k8s)
	if err != nil {
		return errors.Wrap(err, "client")
	}

	podStart := time.Time{}
	podList := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during pod check")
		}
		// Wait for any system pod, as waiting for apiserver may block until etcd
		pods, err := client.CoreV1().Pods("kube-system").List(meta.ListOptions{})
		if len(pods.Items) < 2 {
			podStart = time.Time{}
			return false, nil
		}
		if err != nil {
			podStart = time.Time{}
			return false, nil
		}
		if podStart.IsZero() {
			podStart = time.Now()
		}

		glog.Infof("%d kube-system pods found since %s", len(pods.Items), podStart)
		if time.Since(podStart) > 2*kconst.APICallRetryInterval {
			glog.Infof("stability requirement met, returning")
			return true, nil
		}
		return false, nil
	}
	if err = wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, podList); err != nil {
		return fmt.Errorf("apiserver never returned a pod list")
	}
	glog.Infof("duration metric: took %s to wait for pod list to return data ...", time.Since(pStart))
	return nil
}

// WaitForCluster blocks until the cluster appears to be healthy
func (k *Bootstrapper) WaitForCluster(k8s config.KubernetesConfig, timeout time.Duration) error {
	start := time.Now()
	out.T(out.Waiting, "Waiting for cluster to come online ...")
	if err := k.waitForAPIServerProcess(start, timeout); err != nil {
		return err
	}
	if err := k.waitForAPIServerHealthz(start, k8s, timeout); err != nil {
		return err
	}
	return k.waitForSystemPods(start, k8s, timeout)
}

// restartCluster restarts the Kubernetes cluster configured by kubeadm
func (k *Bootstrapper) restartCluster(k8s config.KubernetesConfig) error {
	glog.Infof("restartCluster start")

	start := time.Now()
	defer func() {
		glog.Infof("restartCluster took %s", time.Since(start))
	}()

	version, err := parseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	phase := "alpha"
	controlPlane := "controlplane"
	if version.GTE(semver.MustParse("1.13.0")) {
		phase = "init"
		controlPlane = "control-plane"
	}

	if err := k.createCompatSymlinks(); err != nil {
		glog.Errorf("failed to create compat symlinks: %v", err)
	}

	baseCmd := fmt.Sprintf("%s %s", invokeKubeadm(k8s.KubernetesVersion), phase)
	cmds := []string{
		fmt.Sprintf("%s phase certs all --config %s", baseCmd, yamlConfigPath),
		fmt.Sprintf("%s phase kubeconfig all --config %s", baseCmd, yamlConfigPath),
		fmt.Sprintf("%s phase %s all --config %s", baseCmd, controlPlane, yamlConfigPath),
		fmt.Sprintf("%s phase etcd local --config %s", baseCmd, yamlConfigPath),
	}

	// Run commands one at a time so that it is easier to root cause failures.
	for _, c := range cmds {
		rr, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", c))
		if err != nil {
			return errors.Wrapf(err, "running cmd: %s", rr.Command())
		}
	}

	// We must ensure that the apiserver is healthy before proceeding
	if err := k.waitForAPIServerHealthz(time.Now(), k8s, kconst.DefaultControlPlaneTimeout); err != nil {
		return errors.Wrap(err, "apiserver healthz")
	}
	if err := k.waitForSystemPods(time.Now(), k8s, kconst.DefaultControlPlaneTimeout); err != nil {
		return errors.Wrap(err, "system pods")
	}

	// Explicitly re-enable kubeadm addons (proxy, coredns) so that they will check for IP or configuration changes.
	if rr, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s phase addon all --config %s", baseCmd, yamlConfigPath))); err != nil {
		return errors.Wrapf(err, fmt.Sprintf("addon phase cmd:%q", rr.Command()))
	}

	if err := k.adjustResourceLimits(); err != nil {
		glog.Warningf("unable to adjust resource limits: %v", err)
	}
	return nil
}

// DeleteCluster removes the components that were started earlier
func (k *Bootstrapper) DeleteCluster(k8s config.KubernetesConfig) error {
	version, err := parseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	cmd := fmt.Sprintf("%s reset --force", invokeKubeadm(k8s.KubernetesVersion))
	if version.LT(semver.MustParse("1.11.0")) {
		cmd = fmt.Sprintf("%s reset", invokeKubeadm(k8s.KubernetesVersion))
	}

	if rr, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", cmd)); err != nil {
		return errors.Wrapf(err, "kubeadm reset: cmd: %q", rr.Command())
	}

	return nil
}

// PullImages downloads images that will be used by Kubernetes
func (k *Bootstrapper) PullImages(k8s config.KubernetesConfig) error {
	version, err := parseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}
	if version.LT(semver.MustParse("1.11.0")) {
		return fmt.Errorf("pull command is not supported by kubeadm v%s", version)
	}

	rr, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s config images pull --config %s", invokeKubeadm(k8s.KubernetesVersion), yamlConfigPath)))
	if err != nil {
		return errors.Wrapf(err, "running cmd: %q", rr.Command())
	}
	return nil
}

// SetupCerts sets up certificates within the cluster.
func (k *Bootstrapper) SetupCerts(k8s config.KubernetesConfig) error {
	return bootstrapper.SetupCerts(k.c, k8s)
}

// NewKubeletConfig generates a new systemd unit containing a configured kubelet
// based on the options present in the KubernetesConfig.
func NewKubeletConfig(k8s config.KubernetesConfig, r cruntime.Manager) ([]byte, error) {
	version, err := parseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "parsing kubernetes version")
	}

	extraOpts, err := ExtraConfigForComponent(Kubelet, k8s.ExtraOptions, version)
	if err != nil {
		return nil, errors.Wrap(err, "generating extra configuration for kubelet")
	}

	for k, v := range r.KubeletOptions() {
		extraOpts[k] = v
	}
	if k8s.NetworkPlugin != "" {
		extraOpts["network-plugin"] = k8s.NetworkPlugin
	}
	if _, ok := extraOpts["node-ip"]; !ok {
		extraOpts["node-ip"] = k8s.NodeIP
	}

	pauseImage := images.Pause(k8s.ImageRepository)
	if _, ok := extraOpts["pod-infra-container-image"]; !ok && k8s.ImageRepository != "" && pauseImage != "" && k8s.ContainerRuntime != remoteContainerRuntime {
		extraOpts["pod-infra-container-image"] = pauseImage
	}

	// parses a map of the feature gates for kubelet
	_, kubeletFeatureArgs, err := ParseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return nil, errors.Wrap(err, "parses feature gate config for kubelet")
	}

	if kubeletFeatureArgs != "" {
		extraOpts["feature-gates"] = kubeletFeatureArgs
	}

	b := bytes.Buffer{}
	opts := struct {
		ExtraOptions     string
		ContainerRuntime string
		KubeletPath      string
	}{
		ExtraOptions:     convertToFlags(extraOpts),
		ContainerRuntime: k8s.ContainerRuntime,
		KubeletPath:      path.Join(binRoot(k8s.KubernetesVersion), "kubelet"),
	}
	if err := kubeletSystemdTemplate.Execute(&b, opts); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// UpdateCluster updates the cluster
func (k *Bootstrapper) UpdateCluster(cfg config.MachineConfig) error {
	images, err := images.Kubeadm(cfg.KubernetesConfig.ImageRepository, cfg.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "kubeadm images")
	}

	if cfg.KubernetesConfig.ShouldLoadCachedImages {
		if err := machine.LoadImages(&cfg, k.c, images, constants.ImageCacheDir); err != nil {
			out.FailureT("Unable to load cached images: {{.error}}", out.V{"error": err})
		}
	}
	r, err := cruntime.New(cruntime.Config{Type: cfg.ContainerRuntime, Socket: cfg.KubernetesConfig.CRISocket})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}
	kubeadmCfg, err := generateConfig(cfg.KubernetesConfig, r)
	if err != nil {
		return errors.Wrap(err, "generating kubeadm cfg")
	}

	kubeletCfg, err := NewKubeletConfig(cfg.KubernetesConfig, r)
	if err != nil {
		return errors.Wrap(err, "generating kubelet config")
	}

	kubeletService, err := NewKubeletService(cfg.KubernetesConfig)
	if err != nil {
		return errors.Wrap(err, "generating kubelet service")
	}

	glog.Infof("kubelet %s config:\n%+v", kubeletCfg, cfg.KubernetesConfig)

	stopCmd := exec.Command("/bin/bash", "-c", "pgrep kubelet && sudo systemctl stop kubelet")
	// stop kubelet to avoid "Text File Busy" error
	if rr, err := k.c.RunCmd(stopCmd); err != nil {
		glog.Warningf("unable to stop kubelet: %s command: %q output: %q", err, rr.Command(), rr.Output())
	}

	if err := transferBinaries(cfg.KubernetesConfig, k.c); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}
	files := configFiles(cfg.KubernetesConfig, kubeadmCfg, kubeletCfg, kubeletService)
	if err := addAddons(&files, assets.GenerateTemplateData(cfg.KubernetesConfig)); err != nil {
		return errors.Wrap(err, "adding addons")
	}
	for _, f := range files {
		if err := k.c.Copy(f); err != nil {
			return errors.Wrapf(err, "copy")
		}
	}

	if _, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", "sudo systemctl daemon-reload && sudo systemctl start kubelet")); err != nil {
		return errors.Wrap(err, "starting kubelet")
	}
	return nil
}

// createExtraComponentConfig generates a map of component to extra args for all of the components except kubeadm
func createExtraComponentConfig(extraOptions config.ExtraOptionSlice, version semver.Version, componentFeatureArgs string) ([]ComponentExtraArgs, error) {
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

// generateConfig generates the kubeadm.yaml file
func generateConfig(k8s config.KubernetesConfig, r cruntime.Manager) ([]byte, error) {
	version, err := parseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "parsing kubernetes version")
	}

	// parses a map of the feature gates for kubeadm and component
	kubeadmFeatureArgs, componentFeatureArgs, err := ParseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return nil, errors.Wrap(err, "parses feature gate config for kubeadm and component")
	}

	extraComponentConfig, err := createExtraComponentConfig(k8s.ExtraOptions, version, componentFeatureArgs)
	if err != nil {
		return nil, errors.Wrap(err, "generating extra component config for kubeadm")
	}

	// In case of no port assigned, use util.APIServerPort
	nodePort := k8s.NodePort
	if nodePort <= 0 {
		nodePort = constants.APIServerPort
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
		DNSDomain         string
		CRISocket         string
		ImageRepository   string
		ExtraArgs         []ComponentExtraArgs
		FeatureArgs       map[string]bool
		NoTaintMaster     bool
	}{
		CertDir:           vmpath.GuestCertsDir,
		ServiceCIDR:       util.DefaultServiceCIDR,
		PodSubnet:         k8s.ExtraOptions.Get("pod-network-cidr", Kubeadm),
		AdvertiseAddress:  k8s.NodeIP,
		APIServerPort:     nodePort,
		KubernetesVersion: k8s.KubernetesVersion,
		EtcdDataDir:       etcdDataDir(),
		NodeName:          k8s.NodeName,
		CRISocket:         r.SocketPath(),
		ImageRepository:   k8s.ImageRepository,
		ExtraArgs:         extraComponentConfig,
		FeatureArgs:       kubeadmFeatureArgs,
		NoTaintMaster:     false, // That does not work with k8s 1.12+
		DNSDomain:         k8s.DNSDomain,
	}

	if k8s.ServiceCIDR != "" {
		opts.ServiceCIDR = k8s.ServiceCIDR
	}

	opts.NoTaintMaster = true
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
		return nil, err
	}

	return b.Bytes(), nil
}

// NewKubeletService returns a generated systemd unit file for the kubelet
func NewKubeletService(cfg config.KubernetesConfig) ([]byte, error) {
	var b bytes.Buffer
	opts := struct{ KubeletPath string }{KubeletPath: path.Join(binRoot(cfg.KubernetesVersion), "kubelet")}
	if err := kubeletServiceTemplate.Execute(&b, opts); err != nil {
		return nil, errors.Wrap(err, "template execute")
	}
	return b.Bytes(), nil
}

// configFiles returns configuration file assets
func configFiles(cfg config.KubernetesConfig, kubeadm []byte, kubelet []byte, kubeletSvc []byte) []assets.CopyableFile {
	fs := []assets.CopyableFile{
		assets.NewMemoryAssetTarget(kubeadm, yamlConfigPath, "0640"),
		assets.NewMemoryAssetTarget(kubelet, kubeletSystemdConfFile, "0644"),
		assets.NewMemoryAssetTarget(kubeletSvc, kubeletServiceFile, "0644"),
	}
	// Copy the default CNI config (k8s.conf), so that kubelet can successfully
	// start a Pod in the case a user hasn't manually installed any CNI plugin
	// and minikube was started with "--extra-config=kubelet.network-plugin=cni".
	if cfg.EnableDefaultCNI {
		fs = append(fs, assets.NewMemoryAssetTarget([]byte(defaultCNIConfig), defaultCNIConfigPath, "0644"))
	}
	return fs
}

// binDir returns the persistent path binaries are stored in
func binRoot(version string) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", version)
}

// invokeKubeadm returns the invocation command for Kubeadm
func invokeKubeadm(version string) string {
	return fmt.Sprintf("sudo env PATH=%s:$PATH kubeadm", binRoot(version))
}

// transferBinaries transfers all required Kubernetes binaries
func transferBinaries(cfg config.KubernetesConfig, c command.Runner) error {
	var g errgroup.Group
	for _, name := range constants.KubeadmBinaries {
		name := name
		g.Go(func() error {
			src, err := machine.CacheBinary(name, cfg.KubernetesVersion, "linux", runtime.GOARCH)
			if err != nil {
				return errors.Wrapf(err, "downloading %s", name)
			}

			dst := path.Join(binRoot(cfg.KubernetesVersion), name)
			if err := machine.CopyBinary(c, src, dst); err != nil {
				return errors.Wrapf(err, "copybinary %s -> %s", src, dst)
			}
			return nil
		})
	}
	return g.Wait()
}
