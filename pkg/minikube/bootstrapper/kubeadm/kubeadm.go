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
	"fmt"
	"io"
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
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/util"
)

type KubeadmBootstrapper struct {
	c bootstrapper.CommandRunner
}

func NewKubeadmBootstrapper(api libmachine.API) (*KubeadmBootstrapper, error) {
	h, err := api.Load(config.GetMachineName())
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	var cmd bootstrapper.CommandRunner
	// The none driver executes commands directly on the host
	if h.Driver.DriverName() == constants.DriverNone {
		cmd = &bootstrapper.ExecRunner{}
	} else {
		client, err := sshutil.NewSSHClient(h.Driver)
		if err != nil {
			return nil, errors.Wrap(err, "getting ssh client")
		}
		cmd = bootstrapper.NewSSHRunner(client)
	}
	return &KubeadmBootstrapper{
		c: cmd,
	}, nil
}

//TODO(r2d4): This should most likely check the health of the apiserver
func (k *KubeadmBootstrapper) GetClusterStatus() (string, error) {
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
	}
	return state.Error.String(), nil
}

// TODO(r2d4): Should this aggregate all the logs from the control plane?
// Maybe subcommands for each component? minikube logs apiserver?
func (k *KubeadmBootstrapper) GetClusterLogsTo(follow bool, out io.Writer) error {
	var flags []string
	if follow {
		flags = append(flags, "-f")
	}
	logsCommand := fmt.Sprintf("sudo journalctl %s -u kubelet", strings.Join(flags, " "))

	if follow {
		if err := k.c.CombinedOutputTo(logsCommand, out); err != nil {
			return errors.Wrap(err, "getting cluster logs")
		}
	} else {

		logs, err := k.c.CombinedOutput(logsCommand)
		if err != nil {
			return errors.Wrap(err, "getting cluster logs")
		}
		fmt.Fprint(out, logs)
	}
	return nil
}

func (k *KubeadmBootstrapper) StartCluster(k8s config.KubernetesConfig) error {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	b := bytes.Buffer{}
	templateContext := struct {
		KubeadmConfigFile   string
		SkipPreflightChecks bool
		Preflights          []string
	}{
		KubeadmConfigFile: constants.KubeadmConfigFile,
		SkipPreflightChecks: !VersionIsBetween(version,
			semver.MustParse("1.9.0-alpha.0"),
			semver.Version{}),
		Preflights: constants.Preflights,
	}
	if err := kubeadmInitTemplate.Execute(&b, templateContext); err != nil {
		return err
	}

	out, err := k.c.CombinedOutput(b.String())
	if err != nil {
		return errors.Wrapf(err, "kubeadm init error %s running command: %s", b.String(), out)
	}

	if version.LT(semver.MustParse("1.10.0-alpha.0")) {
		//TODO(r2d4): get rid of global here
		master = k8s.NodeName
		if err := util.RetryAfter(200, unmarkMaster, time.Second*1); err != nil {
			return errors.Wrap(err, "timed out waiting to unmark master")
		}
	}

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
	for addonName, addonBundle := range assets.Addons {
		// TODO(r2d4): Kubeadm ignores the kube-dns addon and uses its own.
		// expose this in a better way
		if addonName == "kube-dns" {
			continue
		}
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

func (k *KubeadmBootstrapper) RestartCluster(k8s config.KubernetesConfig) error {
	opts := struct {
		KubeadmConfigFile string
	}{
		KubeadmConfigFile: constants.KubeadmConfigFile,
	}

	b := bytes.Buffer{}
	if err := kubeadmRestoreTemplate.Execute(&b, opts); err != nil {
		return err
	}

	if err := k.c.Run(b.String()); err != nil {
		return errors.Wrapf(err, "running cmd: %s", b.String())
	}

	if err := restartKubeProxy(k8s); err != nil {
		return errors.Wrap(err, "restarting kube-proxy")
	}

	return nil
}

func (k *KubeadmBootstrapper) SetupCerts(k8s config.KubernetesConfig) error {
	return bootstrapper.SetupCerts(k.c, k8s)
}

// SetContainerRuntime possibly sets the container runtime, if it hasn't already
// been specified by the extra-config option.  It has a set of defaults known to
// work for a particular runtime.
func SetContainerRuntime(cfg map[string]string, runtime string) map[string]string {
	if _, ok := cfg["container-runtime"]; ok {
		glog.Infoln("Container runtime already set through extra options, ignoring --container-runtime flag.")
		return cfg
	}

	if runtime == "" {
		glog.Infoln("Container runtime flag provided with no value, using defaults.")
		return cfg
	}

	switch runtime {
	case "crio", "cri-o":
		cfg["container-runtime"] = "remote"
		cfg["container-runtime-endpoint"] = "/var/run/crio/crio.sock"
		cfg["image-service-endpoint"] = "/var/run/crio/crio.sock"
		cfg["runtime-request-timeout"] = "15m"
	case "containerd":
		cfg["container-runtime"] = "remote"
		cfg["container-runtime-endpoint"] = "unix:///run/containerd/containerd.sock"
		cfg["image-service-endpoint"] = "unix:///run/containerd/containerd.sock"
		cfg["runtime-request-timeout"] = "15m"
	default:
		cfg["container-runtime"] = runtime
	}

	return cfg
}

// NewKubeletConfig generates a new systemd unit containing a configured kubelet
// based on the options present in the KubernetesConfig.
func NewKubeletConfig(k8s config.KubernetesConfig) (string, error) {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return "", errors.Wrap(err, "parsing kubernetes version")
	}

	extraOpts, err := ExtraConfigForComponent(Kubelet, k8s.ExtraOptions, version)
	if err != nil {
		return "", errors.Wrap(err, "generating extra configuration for kubelet")
	}

	extraOpts = SetContainerRuntime(extraOpts, k8s.ContainerRuntime)
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
		err := machine.LoadImages(k.c, constants.GetKubeadmCachedImages(cfg.KubernetesVersion), constants.ImageCacheDir)
		if err != nil {
			return errors.Wrap(err, "loading cached images")
		}
	}

	kubeadmCfg, err := generateConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "generating kubeadm cfg")
	}

	kubeletCfg, err := NewKubeletConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "generating kubelet config")
	}

	files := []assets.CopyableFile{
		assets.NewMemoryAssetTarget([]byte(kubeletService), constants.KubeletServiceFile, "0640"),
		assets.NewMemoryAssetTarget([]byte(kubeletCfg), constants.KubeletSystemdConfFile, "0640"),
		assets.NewMemoryAssetTarget([]byte(kubeadmCfg), constants.KubeadmConfigFile, "0640"),
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
				return errors.Wrap(err, "making new file asset")
			}
			if err := k.c.Copy(f); err != nil {
				return errors.Wrapf(err, "transferring kubeadm file: %+v", f)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}

	if err := addAddons(&files); err != nil {
		return errors.Wrap(err, "adding addons to copyable files")
	}

	for _, f := range files {
		if err := k.c.Copy(f); err != nil {
			return errors.Wrapf(err, "transferring kubeadm file: %+v", f)
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

func generateConfig(k8s config.KubernetesConfig) (string, error) {
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
		ExtraArgs:         extraComponentConfig,
		FeatureArgs:       kubeadmFeatureArgs,
		NoTaintMaster:     false,
	}

	if version.GTE(semver.MustParse("1.10.0-alpha.0")) {
		opts.NoTaintMaster = true
	}

	b := bytes.Buffer{}
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

	fmt.Printf("Downloading %s %s\n", binary, version)
	if err := download.ToFile(url, targetFilepath, options); err != nil {
		return "", errors.Wrapf(err, "Error downloading %s %s", binary, version)
	}
	fmt.Printf("Finished Downloading %s %s\n", binary, version)

	return targetFilepath, nil
}
