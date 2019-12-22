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

// Package kicbs is a kubeadm-flavor bootstrapper for kic
package kicbs

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/verify"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// Bootstrapper is a bootstrapper using kicbs
type Bootstrapper struct {
	c           command.Runner
	k8sClient   *kubernetes.Clientset // kubernetes client used to verify pods inside cluster
	contextName string
}

// NewBootstrapper creates a new kicbs.Bootstrapper
func NewBootstrapper(api libmachine.API) (*Bootstrapper, error) {
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

// UpdateCluster updates the cluster
func (k *Bootstrapper) UpdateCluster(cfg config.MachineConfig) error {
	images, err := images.KIC(cfg.KubernetesConfig.ImageRepository, cfg.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "kic images")
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
	kubeadmCfg, err := bsutil.GenerateKubeadmYAML(cfg.KubernetesConfig, r)
	if err != nil {
		return errors.Wrap(err, "generating kubeadm cfg")
	}

	kubeletCfg, err := bsutil.NewKubeletConfig(cfg.KubernetesConfig, r)
	if err != nil {
		return errors.Wrap(err, "generating kubelet config")
	}

	kubeletService, err := bsutil.NewKubeletService(cfg.KubernetesConfig)
	if err != nil {
		return errors.Wrap(err, "generating kubelet service")
	}

	// TODO:medyagh remove this one
	delStuff := exec.Command("rm", "-f", "/kind/systemd/")
	if rr, err := k.c.RunCmd(delStuff); err != nil {
		glog.Warningf("unable to del crap kubelet: %s command: %q output: %q", err, rr.Command(), rr.Output())
	}

	glog.Infof("kubelet %s config:\n%+v", kubeletCfg, cfg.KubernetesConfig)

	stopCmd := exec.Command("/bin/bash", "-c", "pgrep kubelet && sudo systemctl stop kubelet")
	// stop kubelet to avoid "Text File Busy" error
	if rr, err := k.c.RunCmd(stopCmd); err != nil {
		glog.Warningf("unable to stop kubelet: %s command: %q output: %q", err, rr.Command(), rr.Output())
	}

	if err := bsutil.TransferBinaries(cfg.KubernetesConfig, k.c); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}

	var cniFile []byte = nil
	cniFile = []byte(defaultCNIManifest)

	files := bsutil.ConfigFileAssets(cfg.KubernetesConfig, kubeadmCfg, kubeletCfg, kubeletService, cniFile)

	// if err := bsutil.AddAddons(&files, assets.GenerateTemplateData(cfg.KubernetesConfig)); err != nil {
	// 	return errors.Wrap(err, "adding addons")
	// }
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

// SetupCerts generates the certs the cluster
func (k *Bootstrapper) SetupCerts(cfg config.KubernetesConfig) error {
	return bootstrapper.SetupCerts(k.c, cfg)
}

// PullImages downloads images that will be used by Kubernetes
func (k *Bootstrapper) PullImages(k8s config.KubernetesConfig) error {
	version, err := bsutil.ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}
	if version.LT(semver.MustParse("1.11.0")) {
		return fmt.Errorf("pull command is not supported by kubeadm v%s", version)
	}

	rr, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s config images pull --config %s", bsutil.InvokeKubeadm(k8s.KubernetesVersion), bsutil.KubeadmYamlPath)))
	if err != nil {
		return errors.Wrapf(err, "running cmd: %q", rr.Command())
	}
	return nil
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

	version, err := bsutil.ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	extraFlags := bsutil.CreateFlagsFromExtraArgs(k8s.ExtraOptions)
	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime})
	if err != nil {
		return err
	}

	ignore := []string{
		fmt.Sprintf("DirAvailable-%s", strings.Replace(vmpath.GuestManifestsDir, "/", "-", -1)),
		fmt.Sprintf("DirAvailable-%s", strings.Replace(vmpath.GuestPersistentDir, "/", "-", -1)),
		fmt.Sprintf("DirAvailable-%s", strings.Replace(bsutil.EtcdDataDir(), "/", "-", -1)),
		"FileAvailable--etc-kubernetes-manifests-kube-scheduler.yaml",
		"FileAvailable--etc-kubernetes-manifests-kube-apiserver.yaml",
		"FileAvailable--etc-kubernetes-manifests-kube-controller-manager.yaml",
		"FileAvailable--etc-kubernetes-manifests-etcd.yaml",
		"FileContent--proc-sys-net-bridge-bridge-nf-call-iptables", // for kic only
		"Port-10250", // For "none" users who already have a kubelet online
		"Swap",       // For "none" users who have swap configured
	}
	ignore = append(ignore, bsutil.SkipAdditionalPreflights[r.Name()]...)

	// Allow older kubeadm versions to function with newer Docker releases.
	if version.LT(semver.MustParse("1.13.0")) {
		glog.Infof("Older Kubernetes release detected (%s), disabling SystemVerification check.", version)
		ignore = append(ignore, "SystemVerification")
	}

	// TODO:medyagh delete this temp work arround
	rr, err := k.c.RunCmd(exec.Command("rm", "-f", "/usr/bin/kubeadm"))
	fmt.Printf("Deleting kics kubeadm %s %v \n ", rr.Output(), err)
	rr, err = k.c.RunCmd(exec.Command("rm", "-f", "/usr/bin/kubelet"))
	fmt.Printf("Deleting kics kubelet %s %v \n", rr.Output(), err)

	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s init --config %s %s --ignore-preflight-errors=%s", bsutil.InvokeKubeadm(k8s.KubernetesVersion), bsutil.KubeadmYamlPath, extraFlags, strings.Join(ignore, ",")))
	if rr, err := k.c.RunCmd(c); err != nil {
		return errors.Wrapf(err, "init failed. cmd: %q output: %q", rr.Command(), rr.Output())
	}

	glog.Infof("applying kic overlay network")
	if err := k.applyOverlayNetwork(); err != nil {
		return errors.Wrap(err, "applying kic overlay network")
	}

	// glog.Infof("removing master taint")
	// if err := k.removeMasterTaint(); err != nil {
	// 	return errors.Wrap(err, "remove master taint")
	// }

	glog.Infof("Skipping Configuring cluster permissions for kic...")

	// elevate := func() error {
	// 	client, err := k.client(k8s)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return bsutil.ElevateKubeSystemPrivileges(client)
	// }

	// if err := retry.Expo(elevate, time.Millisecond*500, 120*time.Second); err != nil {
	// 	return errors.Wrap(err, "timed out waiting to elevate kube-system RBAC privileges")
	// }

	if err := k.adjustResourceLimits(); err != nil {
		glog.Warningf("unable to adjust resource limits: %v", err)
	}

	return nil
}

func (k *Bootstrapper) existingConfig() error {
	args := append([]string{"ls"}, bsutil.ExpectedRemoteArtifacts...)
	_, err := k.c.RunCmd(exec.Command("sudo", args...))
	return err
}

// restartCluster restarts the Kubernetes cluster configured by kubeadm
func (k *Bootstrapper) restartCluster(k8s config.KubernetesConfig) error {
	glog.Infof("restartCluster start")

	start := time.Now()
	defer func() {
		glog.Infof("restartCluster took %s", time.Since(start))
	}()

	version, err := bsutil.ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	phase := "alpha"
	controlPlane := "controlplane"
	if version.GTE(semver.MustParse("1.13.0")) {
		phase = "init"
		controlPlane = "control-plane"
	}

	baseCmd := fmt.Sprintf("%s %s", bsutil.InvokeKubeadm(k8s.KubernetesVersion), phase)
	cmds := []string{
		fmt.Sprintf("%s phase certs all --config %s", baseCmd, bsutil.KubeadmYamlPath),
		fmt.Sprintf("%s phase kubeconfig all --config %s", baseCmd, bsutil.KubeadmYamlPath),
		fmt.Sprintf("%s phase %s all --config %s", baseCmd, controlPlane, bsutil.KubeadmYamlPath),
		fmt.Sprintf("%s phase etcd local --config %s", baseCmd, bsutil.KubeadmYamlPath),
	}

	// Run commands one at a time so that it is easier to root cause failures.
	for _, c := range cmds {
		rr, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", c))
		if err != nil {
			return errors.Wrapf(err, "running cmd: %s", rr.Command())
		}
	}

	// We must ensure that the apiserver is healthy before proceeding
	if err := verify.APIServerProcess(k.c, time.Now(), kconst.DefaultControlPlaneTimeout); err != nil {
		return errors.Wrap(err, "apiserver healthz")
	}

	client, err := k.client(k8s)
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}

	if err := verify.SystemPods(client, time.Now(), k8s, kconst.DefaultControlPlaneTimeout); err != nil {
		return errors.Wrap(err, "system pods")
	}

	// Explicitly re-enable kubeadm addons (proxy, coredns) so that they will check for IP or configuration changes.
	if rr, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s phase addon all --config %s", baseCmd, bsutil.KubeadmYamlPath))); err != nil {
		return errors.Wrapf(err, fmt.Sprintf("addon phase cmd:%q", rr.Command()))
	}

	if err := k.adjustResourceLimits(); err != nil {
		glog.Warningf("unable to adjust resource limits: %v", err)
	}
	return nil
}

// WaitForCluster blocks until the cluster appears to be healthy
func (k *Bootstrapper) WaitForCluster(k8s config.KubernetesConfig, timeout time.Duration) error {
	start := time.Now()
	out.T(out.Waiting, "Waiting for cluster to come online ...")
	if err := verify.APIServerProcess(k.c, start, timeout); err != nil {
		return errors.Wrap(err, "wait for api proc")
	}
	c, err := k.client(k8s)
	if err != nil {
		return errors.Wrap(err, "get k8s client")
	}

	if err := verify.SystemPods(c, start, k8s, timeout); err != nil {
		return errors.Wrap(err, "wait for system pods")
	}

	if err := verify.APIServerIsRunning(start, k8s.NodeIP, k8s.NodePort, timeout); err != nil {
		return err
	}

	return nil
}

func (k *Bootstrapper) DeleteCluster(config.KubernetesConfig) error {
	return fmt.Errorf("the DeleteCluster is not implemented in kicbs yet")
}

func (k *Bootstrapper) LogCommands(bootstrapper.LogOptions) map[string]string {
	return map[string]string{}
}

func (k *Bootstrapper) GetKubeletStatus() (string, error) {
	return "", fmt.Errorf("the GetKubeletStatus is not implemented in kicbs yet")
}
func (k *Bootstrapper) GetAPIServerStatus(net.IP, int) (string, error) {
	return "", fmt.Errorf("the GetAPIServerStatus is not implemented in kicbs yet")
}

// TODO:medyagh use the kapi package for this
// client sets and returns a Kubernetes client to use to speak to a kubeadm launched apiserver
func (k *Bootstrapper) client(k8s config.KubernetesConfig) (*kubernetes.Clientset, error) {
	if k.k8sClient != nil {
		return k.k8sClient, nil
	}

	config, err := kapi.ClientConfig(k.contextName)
	if err != nil {
		return nil, errors.Wrap(err, "client config")
	}

	// TODO:medyagh maybe for liux machines we could use container ip
	endpoint := fmt.Sprintf("https://%s", net.JoinHostPort("127.0.0.1", fmt.Sprint(k8s.HostBindPort)))
	if config.Host != endpoint {
		glog.Errorf("Overriding stale ClientConfig host %s with %s", config.Host, endpoint)
		config.Host = endpoint
	}
	c, err := kubernetes.NewForConfig(config)
	if err == nil {
		k.k8sClient = c
	}
	return c, err
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

// applyOverlayNetwork applies the CNI plugin needed to make kic work
func (k *Bootstrapper) applyOverlayNetwork() error {
	cmd := exec.Command(
		"kubectl", "create", "--kubeconfig=/etc/kubernetes/admin.conf",
		"-f", bsutil.DefaultCNIConfigPath,
	)
	if rr, err := k.c.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "cmd: %s output: %s", rr.Command(), rr.Output())
	}
	return nil
}

// removeMasterTaint so pods can be scheduled on the master node
func (k *Bootstrapper) removeMasterTaint() error {
	// if we are only provisioning one node, remove the master taint
	// https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/#master-isolation
	cmd := exec.Command(
		"kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"taint", "nodes", "--all", "node-role.kubernetes.io/master-",
	)

	if rr, err := k.c.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "output: %s", rr.Output())
	}
	return nil
}
