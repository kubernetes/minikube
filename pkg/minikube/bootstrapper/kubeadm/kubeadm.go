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
	"crypto"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	download "github.com/jimmidyson/go-download"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/util"
)

// SkipPreflights are preflight checks we always skip.
var SkipPreflights = []string{
	// We use --ignore-preflight-errors=DirAvailable since we have our own custom addons
	// that we also stick in /etc/kubernetes/manifests
	"DirAvailable--etc-kubernetes-manifests",
	"DirAvailable--data-minikube",
	"Port-10250",
	"FileAvailable--etc-kubernetes-manifests-kube-scheduler.yaml",
	"FileAvailable--etc-kubernetes-manifests-kube-apiserver.yaml",
	"FileAvailable--etc-kubernetes-manifests-kube-controller-manager.yaml",
	"FileAvailable--etc-kubernetes-manifests-etcd.yaml",
	// We use --ignore-preflight-errors=Swap since minikube.iso allocates a swap partition.
	// (it should probably stop doing this, though...)
	"Swap",
	// We use --ignore-preflight-errors=CRI since /var/run/dockershim.sock is not present.
	// (because we start kubelet with an invalid config)
	"CRI",
}

// SkipAdditionalPreflights are additional preflights we skip depending on the runtime in use.
var SkipAdditionalPreflights = map[string][]string{}

type KubeadmBootstrapper struct {
	c bootstrapper.CommandRunner
}

func NewKubeadmBootstrapper(api libmachine.API) (*KubeadmBootstrapper, error) {
	h, err := api.Load(config.GetMachineName())
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return nil, errors.Wrap(err, "command runner")
	}
	return &KubeadmBootstrapper{c: runner}, nil
}

func (k *KubeadmBootstrapper) GetKubeletStatus() (string, error) {
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

func (k *KubeadmBootstrapper) GetApiServerStatus(ip net.IP) (string, error) {
	url := fmt.Sprintf("https://%s:%d/healthz", ip, util.APIServerPort)
	// To avoid: x509: certificate signed by unknown authority
	tr := &http.Transport{
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
func (k *KubeadmBootstrapper) LogCommands(o bootstrapper.LogOptions) map[string]string {
	var kcmd strings.Builder
	kcmd.WriteString("journalctl -u kubelet")
	if o.Lines > 0 {
		kcmd.WriteString(fmt.Sprintf(" -n %d", o.Lines))
	}
	if o.Follow {
		kcmd.WriteString(" -f")
	}
	return map[string]string{"kubelet": kcmd.String()}
}

func (k *KubeadmBootstrapper) StartCluster(k8s config.KubernetesConfig) error {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime})
	if err != nil {
		return err
	}
	b := bytes.Buffer{}
	preflights := SkipPreflights
	preflights = append(preflights, SkipAdditionalPreflights[r.Name()]...)

	templateContext := struct {
		KubeadmConfigFile   string
		SkipPreflightChecks bool
		Preflights          []string
	}{
		KubeadmConfigFile: constants.KubeadmConfigFile,
		SkipPreflightChecks: !VersionIsBetween(version,
			semver.MustParse("1.9.0-alpha.0"),
			semver.Version{}),
		Preflights: preflights,
	}
	if err := kubeadmInitTemplate.Execute(&b, templateContext); err != nil {
		return err
	}

	out, err := k.c.CombinedOutput(b.String())
	if err != nil {
		return errors.Wrapf(err, "kubeadm init: %s\n%s\n", b.String(), out)
	}

	if version.LT(semver.MustParse("1.10.0-alpha.0")) {
		//TODO(r2d4): get rid of global here
		master = k8s.NodeName
		if err := util.RetryAfter(200, unmarkMaster, time.Second*1); err != nil {
			return errors.Wrap(err, "timed out waiting to unmark master")
		}
	}

	// NOTE: We have not yet asserted that we can access the apiserver. Now would be a great time to do so.
	console.OutStyle("permissions", "Configuring cluster permissions ...")
	if err := util.RetryAfter(100, elevateKubeSystemPrivileges, time.Millisecond*500); err != nil {
		return errors.Wrap(err, "timed out waiting to elevate kube-system RBAC privileges")
	}

	return nil
}

func addAddons(files *[]assets.CopyableFile) error {
	// add addons to file list
	// custom addons
	if err := assets.AddMinikubeDirAssets(files); err != nil {
		return errors.Wrap(err, "adding minikube dir assets")
	}
	// bundled addons
	for _, addonBundle := range assets.Addons {
		if isEnabled, err := addonBundle.IsEnabled(); err == nil && isEnabled {
			for _, addon := range addonBundle.Assets {
				*files = append(*files, addon)
			}
		} else if err != nil {
			return nil
		}
	}

	return nil
}

// RestartCluster restarts the Kubernetes cluster configured by kubeadm
func (k *KubeadmBootstrapper) RestartCluster(k8s config.KubernetesConfig) error {
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

	cmds := []string{
		fmt.Sprintf("sudo kubeadm %s phase certs all --config %s", phase, constants.KubeadmConfigFile),
		fmt.Sprintf("sudo kubeadm %s phase kubeconfig all --config %s", phase, constants.KubeadmConfigFile),
		fmt.Sprintf("sudo kubeadm %s phase %s all --config %s", phase, controlPlane, constants.KubeadmConfigFile),
		fmt.Sprintf("sudo kubeadm %s phase etcd local --config %s", phase, constants.KubeadmConfigFile),
	}

	// Run commands one at a time so that it is easier to root cause failures.
	for _, cmd := range cmds {
		if err := k.c.Run(cmd); err != nil {
			return errors.Wrapf(err, "running cmd: %s", cmd)
		}
	}

	// NOTE: Perhaps now would be a good time to check apiserver health?
	console.OutStyle("waiting", "Waiting for kube-proxy to come back up ...")
	if err := restartKubeProxy(k8s); err != nil {
		return errors.Wrap(err, "restarting kube-proxy")
	}

	return nil
}

// DeleteCluster removes the components that were started earlier
func (k *KubeadmBootstrapper) DeleteCluster(k8s config.KubernetesConfig) error {
	cmd := fmt.Sprintf("sudo kubeadm reset --force")
	out, err := k.c.CombinedOutput(cmd)
	if err != nil {
		return errors.Wrapf(err, "kubeadm reset: %s\n%s\n", cmd, out)
	}

	return nil
}

// PullImages downloads images that will be used by RestartCluster
func (k *KubeadmBootstrapper) PullImages(k8s config.KubernetesConfig) error {
	cmd := fmt.Sprintf("sudo kubeadm config images pull --config %s", constants.KubeadmConfigFile)
	if err := k.c.Run(cmd); err != nil {
		return errors.Wrapf(err, "running cmd: %s", cmd)
	}
	return nil
}

// SetupCerts sets up certificates within the cluster.
func (k *KubeadmBootstrapper) SetupCerts(k8s config.KubernetesConfig) error {
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

	extraFlags := convertToFlags(extraOpts)

	// parses a map of the feature gates for kubelet
	_, kubeletFeatureArgs, err := ParseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return "", errors.Wrap(err, "parses feature gate config for kubelet")
	}

	b := bytes.Buffer{}
	opts := struct {
		ExtraOptions     string
		FeatureGates     string
		ContainerRuntime string
	}{
		ExtraOptions:     extraFlags,
		FeatureGates:     kubeletFeatureArgs,
		ContainerRuntime: k8s.ContainerRuntime,
	}
	if err := kubeletSystemdTemplate.Execute(&b, opts); err != nil {
		return "", err
	}

	return b.String(), nil
}

func (k *KubeadmBootstrapper) UpdateCluster(cfg config.KubernetesConfig) error {
	if cfg.ShouldLoadCachedImages {
		if err := machine.LoadImages(k.c, constants.GetKubeadmCachedImages(cfg.KubernetesVersion), constants.ImageCacheDir); err != nil {
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

	files := []assets.CopyableFile{
		assets.NewMemoryAssetTarget([]byte(kubeletService), constants.KubeletServiceFile, "0640"),
		assets.NewMemoryAssetTarget([]byte(kubeletCfg), constants.KubeletSystemdConfFile, "0640"),
		assets.NewMemoryAssetTarget([]byte(kubeadmCfg), constants.KubeadmConfigFile, "0640"),
	}

	// Copy the default CNI config (k8s.conf), so that kubelet can successfully
	// start a Pod in the case a user hasn't manually installed any CNI plugin
	// and minikube was started with "--extra-config=kubelet.network-plugin=cni".
	if cfg.EnableDefaultCNI {
		files = append(files,
			assets.NewMemoryAssetTarget([]byte(defaultCNIConfig), constants.DefaultCNIConfigPath, "0644"),
			assets.NewMemoryAssetTarget([]byte(defaultCNIConfig), constants.DefaultRktNetConfigPath, "0644"))
	}

	var g errgroup.Group
	for _, bin := range []string{"kubelet", "kubeadm"} {
		bin := bin
		g.Go(func() error {
			path, err := maybeDownloadAndCache(bin, cfg.KubernetesVersion)
			if err != nil {
				return errors.Wrapf(err, "downloading %s", bin)
			}
			f, err := assets.NewFileAsset(path, "/usr/bin", bin, "0641")
			if err != nil {
				return errors.Wrap(err, "new file asset")
			}
			if err := k.c.Copy(f); err != nil {
				return errors.Wrapf(err, "copy")
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}

	if err := addAddons(&files); err != nil {
		return errors.Wrap(err, "adding addons")
	}

	for _, f := range files {
		if err := k.c.Copy(f); err != nil {
			return errors.Wrapf(err, "copy")
		}
	}
	err = k.c.Run(`
sudo systemctl daemon-reload &&
sudo systemctl enable kubelet &&
sudo systemctl start kubelet
`)
	if err != nil {
		return errors.Wrap(err, "starting kubelet")
	}

	return nil
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

	// generates a map of component to extra args for apiserver, controller-manager, and scheduler
	extraComponentConfig, err := NewComponentExtraArgs(k8s.ExtraOptions, version, componentFeatureArgs)
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
		AdvertiseAddress  string
		APIServerPort     int
		KubernetesVersion string
		EtcdDataDir       string
		NodeName          string
		CRISocket         string
		ExtraArgs         []ComponentExtraArgs
		FeatureArgs       map[string]bool
		NoTaintMaster     bool
	}{
		CertDir:           util.DefaultCertPath,
		ServiceCIDR:       util.DefaultServiceCIDR,
		AdvertiseAddress:  k8s.NodeIP,
		APIServerPort:     nodePort,
		KubernetesVersion: k8s.KubernetesVersion,
		EtcdDataDir:       "/data/minikube", //TODO(r2d4): change to something else persisted
		NodeName:          k8s.NodeName,
		CRISocket:         r.SocketPath(),
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
	kubeadmConfigTemplate := kubeadmConfigTemplateV1Alpha1
	if version.GTE(semver.MustParse("1.12.0")) {
		kubeadmConfigTemplate = kubeadmConfigTemplateV1Alpha3
	}
	if err := kubeadmConfigTemplate.Execute(&b, opts); err != nil {
		return "", err
	}

	return b.String(), nil
}

func maybeDownloadAndCache(binary, version string) (string, error) {
	targetDir := constants.MakeMiniPath("cache", version)
	targetFilepath := path.Join(targetDir, binary)

	_, err := os.Stat(targetFilepath)
	// If it exists, do no verification and continue
	if err == nil {
		return targetFilepath, nil
	}
	if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "stat %s version %s at %s", binary, version, targetDir)
	}

	if err = os.MkdirAll(targetDir, 0777); err != nil {
		return "", errors.Wrapf(err, "mkdir %s", targetDir)
	}

	url := constants.GetKubernetesReleaseURL(binary, version)
	options := download.FileOptions{
		Mkdirs: download.MkdirAll,
	}

	options.Checksum = constants.GetKubernetesReleaseURLSha1(binary, version)
	options.ChecksumHash = crypto.SHA1

	console.OutStyle("file-download", "Downloading %s %s", binary, version)
	if err := download.ToFile(url, targetFilepath, options); err != nil {
		return "", errors.Wrapf(err, "Error downloading %s %s", binary, version)
	}
	return targetFilepath, nil
}
