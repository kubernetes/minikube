/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"context"
	"os/exec"
	"path"

	"fmt"
	"net"

	// WARNING: Do not use path/filepath in this package unless you want bizarre Windows paths

	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/kubelet"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
	"k8s.io/minikube/pkg/version"
)

// Bootstrapper is a bootstrapper using kubeadm
type Bootstrapper struct {
	c           command.Runner
	k8sClient   *kubernetes.Clientset // kubernetes client used to verify pods inside cluster
	contextName string
}

// NewBootstrapper creates a new kubeadm.Bootstrapper
// TODO(#6891): Remove node as an argument
func NewBootstrapper(api libmachine.API, cc config.ClusterConfig, n config.Node) (*Bootstrapper, error) {
	name := driver.MachineName(cc, n)
	h, err := api.Load(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return nil, errors.Wrap(err, "command runner")
	}
	return &Bootstrapper{c: runner, contextName: cc.Name, k8sClient: nil}, nil
}

// GetKubeletStatus returns the kubelet status
func (k *Bootstrapper) GetKubeletStatus() (string, error) {
	rr, err := k.c.RunCmd(exec.Command("sudo", "systemctl", "is-active", "kubelet"))
	if err != nil {
		// Do not return now, as we still have parsing to do!
		glog.Warningf("%s returned error: %v", rr.Command(), err)
	}
	s := strings.TrimSpace(rr.Stdout.String())
	glog.Infof("kubelet is-active: %s", s)
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
func (k *Bootstrapper) GetAPIServerStatus(hostname string, port int) (string, error) {
	s, err := kverify.APIServerStatus(k.c, hostname, port)
	if err != nil {
		return state.Error.String(), err
	}
	return s.String(), nil
}

// LogCommands returns a map of log type to a command which will display that log.
func (k *Bootstrapper) LogCommands(cfg config.ClusterConfig, o bootstrapper.LogOptions) map[string]string {
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

	describeNodes := fmt.Sprintf("sudo %s describe nodes --kubeconfig=%s",
		path.Join(vmpath.GuestPersistentDir, "binaries", cfg.KubernetesConfig.KubernetesVersion, "kubectl"),
		path.Join(vmpath.GuestPersistentDir, "kubeconfig"))

	return map[string]string{
		"kubelet":        kubelet.String(),
		"dmesg":          dmesg.String(),
		"describe nodes": describeNodes,
	}
}

// createCompatSymlinks creates compatibility symlinks to transition running services to new directory structures
func (k *Bootstrapper) createCompatSymlinks() error {
	legacyEtcd := "/data/minikube"

	if _, err := k.c.RunCmd(exec.Command("sudo", "test", "-d", legacyEtcd)); err != nil {
		glog.Infof("%s skipping compat symlinks: %v", legacyEtcd, err)
		return nil
	}
	glog.Infof("Found %s, creating compatibility symlinks ...", legacyEtcd)

	c := exec.Command("sudo", "ln", "-s", legacyEtcd, bsutil.EtcdDataDir())
	if rr, err := k.c.RunCmd(c); err != nil {
		return errors.Wrapf(err, "create symlink failed: %s", rr.Command())
	}
	return nil
}

// clearStaleConfigs clears configurations which may have stale IP addresses
func (k *Bootstrapper) clearStaleConfigs(cfg config.ClusterConfig) error {
	cp, err := config.PrimaryControlPlane(&cfg)
	if err != nil {
		return err
	}

	paths := []string{
		"/etc/kubernetes/admin.conf",
		"/etc/kubernetes/kubelet.conf",
		"/etc/kubernetes/controller-manager.conf",
		"/etc/kubernetes/scheduler.conf",
	}

	endpoint := fmt.Sprintf("https://%s", net.JoinHostPort(cp.IP, strconv.Itoa(cp.Port)))
	for _, path := range paths {
		_, err := k.c.RunCmd(exec.Command("sudo", "/bin/bash", "-c", fmt.Sprintf("grep %s %s || sudo rm -f %s", endpoint, path, path)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *Bootstrapper) init(cfg config.ClusterConfig) error {
	version, err := util.ParseKubernetesVersion(cfg.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	extraFlags := bsutil.CreateFlagsFromExtraArgs(cfg.KubernetesConfig.ExtraOptions)
	r, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime})
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
		"Port-10250", // For "none" users who already have a kubelet online
		"Swap",       // For "none" users who have swap configured
		"SystemVerification",
	}
	ignore = append(ignore, bsutil.SkipAdditionalPreflights[r.Name()]...)

	// Allow older kubeadm versions to function with newer Docker releases.
	// For kic on linux example error: "modprobe: FATAL: Module configs not found in directory /lib/modules/5.2.17-1rodete3-amd64"
	if version.LT(semver.MustParse("1.13.0")) || driver.IsKIC(cfg.Driver) {
		glog.Info("ignoring SystemVerification for kubeadm because of either driver or kubernetes version")
		ignore = append(ignore, "SystemVerification")
	}

	if driver.IsKIC(cfg.Driver) { // to bypass this error: /proc/sys/net/bridge/bridge-nf-call-iptables does not exist
		ignore = append(ignore, "FileContent--proc-sys-net-bridge-bridge-nf-call-iptables")

	}

	if err := k.clearStaleConfigs(cfg); err != nil {
		return errors.Wrap(err, "clearing stale configs")
	}

	conf := bsutil.KubeadmYamlPath
	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s init --config %s %s --ignore-preflight-errors=%s",
		bsutil.InvokeKubeadm(cfg.KubernetesConfig.KubernetesVersion), conf, extraFlags, strings.Join(ignore, ",")))
	if _, err := k.c.RunCmd(c); err != nil {
		return errors.Wrap(err, "run")
	}

	// this is required for containerd and cri-o runtime. till we close https://github.com/kubernetes/minikube/issues/7428
	if driver.IsKIC(cfg.Driver) && cfg.KubernetesConfig.ContainerRuntime != "docker" {
		if err := k.applyKicOverlay(cfg); err != nil {
			return errors.Wrap(err, "apply kic overlay")
		}
	}

	if err := k.applyNodeLabels(cfg); err != nil {
		glog.Warningf("unable to apply node labels: %v", err)
	}

	if err := bsutil.AdjustResourceLimits(k.c); err != nil {
		glog.Warningf("unable to adjust resource limits: %v", err)
	}

	if err := k.elevateKubeSystemPrivileges(cfg); err != nil {
		glog.Warningf("unable to create cluster role binding, some addons might not work: %v", err)
	}
	return nil
}

// unpause unpauses any Kubernetes backplane components
func (k *Bootstrapper) unpause(cfg config.ClusterConfig) error {

	cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: k.c})
	if err != nil {
		return err
	}

	ids, err := cr.ListContainers(cruntime.ListOptions{State: cruntime.Paused, Namespaces: []string{"kube-system"}})
	if err != nil {
		return errors.Wrap(err, "list paused")
	}

	if len(ids) > 0 {
		if err := cr.UnpauseContainers(ids); err != nil {
			return err
		}
	}
	return nil
}

// StartCluster starts the cluster
func (k *Bootstrapper) StartCluster(cfg config.ClusterConfig) error {
	start := time.Now()
	glog.Infof("StartCluster: %+v", cfg)
	defer func() {
		glog.Infof("StartCluster complete in %s", time.Since(start))
	}()

	// Before we start, ensure that no paused components are lurking around
	if err := k.unpause(cfg); err != nil {
		glog.Warningf("unpause failed: %v", err)
	}

	if err := bsutil.ExistingConfig(k.c); err == nil {
		glog.Infof("found existing configuration files, will attempt cluster restart")
		rerr := k.restartCluster(cfg)
		if rerr == nil {
			return nil
		}
		out.ErrT(out.Embarrassed, "Unable to restart cluster, will reset it: {{.error}}", out.V{"error": rerr})
		if err := k.DeleteCluster(cfg.KubernetesConfig); err != nil {
			glog.Warningf("delete failed: %v", err)
		}
		// Fall-through to init
	}

	conf := bsutil.KubeadmYamlPath
	if _, err := k.c.RunCmd(exec.Command("sudo", "cp", conf+".new", conf)); err != nil {
		return errors.Wrap(err, "cp")
	}

	err := k.init(cfg)
	if err == nil {
		return nil
	}

	out.ErrT(out.Conflict, "initialization failed, will try again: {{.error}}", out.V{"error": err})
	if err := k.DeleteCluster(cfg.KubernetesConfig); err != nil {
		glog.Warningf("delete failed: %v", err)
	}
	return k.init(cfg)
}

// client sets and returns a Kubernetes client to use to speak to a kubeadm launched apiserver
func (k *Bootstrapper) client(ip string, port int) (*kubernetes.Clientset, error) {
	if k.k8sClient != nil {
		return k.k8sClient, nil
	}

	cc, err := kapi.ClientConfig(k.contextName)
	if err != nil {
		return nil, errors.Wrap(err, "client config")
	}

	endpoint := fmt.Sprintf("https://%s", net.JoinHostPort(ip, strconv.Itoa(port)))
	if cc.Host != endpoint {
		glog.Errorf("Overriding stale ClientConfig host %s with %s", cc.Host, endpoint)
		cc.Host = endpoint
	}
	c, err := kubernetes.NewForConfig(cc)
	if err == nil {
		k.k8sClient = c
	}
	return c, err
}

// WaitForNode blocks until the node appears to be healthy
func (k *Bootstrapper) WaitForNode(cfg config.ClusterConfig, n config.Node, timeout time.Duration) error {
	start := time.Now()

	if !n.ControlPlane {
		glog.Infof("%s is not a control plane, nothing to wait for", n.Name)
		return nil
	}
	if !kverify.ShouldWait(cfg.VerifyComponents) {
		glog.Infof("skip waiting for components based on config.")
		return nil
	}

	cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: k.c})
	if err != nil {
		return errors.Wrapf(err, "create runtme-manager %s", cfg.KubernetesConfig.ContainerRuntime)
	}

	hostname, _, port, err := driver.ControlPaneEndpoint(&cfg, &n, cfg.Driver)
	if err != nil {
		return errors.Wrap(err, "get control plane endpoint")
	}

	if cfg.VerifyComponents[kverify.APIServerWaitKey] {
		client, err := k.client(hostname, port)
		if err != nil {
			return errors.Wrap(err, "get k8s client")
		}
		if err := kverify.WaitForAPIServerProcess(cr, k, cfg, k.c, start, timeout); err != nil {
			return errors.Wrap(err, "wait for apiserver proc")
		}

		if err := kverify.WaitForHealthyAPIServer(cr, k, cfg, k.c, client, start, hostname, port, timeout); err != nil {
			return errors.Wrap(err, "wait for healthy API server")
		}
	}

	if cfg.VerifyComponents[kverify.SystemPodsWaitKey] {
		client, err := k.client(hostname, port)
		if err != nil {
			return errors.Wrap(err, "get k8s client")
		}
		if err := kverify.WaitForSystemPods(cr, k, cfg, k.c, client, start, timeout); err != nil {
			return errors.Wrap(err, "waiting for system pods")
		}
	}

	if cfg.VerifyComponents[kverify.DefaultSAWaitKey] {
		client, err := k.client(hostname, port)
		if err != nil {
			return errors.Wrap(err, "get k8s client")
		}
		if err := kverify.WaitForDefaultSA(client, timeout); err != nil {
			return errors.Wrap(err, "waiting for default service account")
		}
	}

	if cfg.VerifyComponents[kverify.AppsRunning] {
		client, err := k.client(hostname, port)
		if err != nil {
			return errors.Wrap(err, "get k8s client")
		}
		if err := kverify.WaitForAppsRunning(client, kverify.AppsRunningList, timeout); err != nil {
			return errors.Wrap(err, "waiting for apps_running")
		}
	}

	glog.Infof("duration metric: took %s to wait for : %+v ...", time.Since(start), cfg.VerifyComponents)
	return nil
}

// needsReset returns whether or not the cluster needs to be reconfigured
func (k *Bootstrapper) needsReset(conf string, hostname string, port int, client *kubernetes.Clientset, version string) bool {
	if rr, err := k.c.RunCmd(exec.Command("sudo", "diff", "-u", conf, conf+".new")); err != nil {
		glog.Infof("needs reset: configs differ:\n%s", rr.Output())
		return true
	}

	st, err := kverify.APIServerStatus(k.c, hostname, port)
	if err != nil {
		glog.Infof("needs reset: apiserver error: %v", err)
		return true
	}

	if st != state.Running {
		glog.Infof("needs reset: apiserver in state %s", st)
		return true
	}

	if err := kverify.ExpectAppsRunning(client, kverify.AppsRunningList); err != nil {
		glog.Infof("needs reset: %v", err)
		return true
	}

	if err := kverify.APIServerVersionMatch(client, version); err != nil {
		glog.Infof("needs reset: %v", err)
		return true
	}

	return false
}

// restartCluster restarts the Kubernetes cluster configured by kubeadm
func (k *Bootstrapper) restartCluster(cfg config.ClusterConfig) error {
	glog.Infof("restartCluster start")

	start := time.Now()
	defer func() {
		glog.Infof("restartCluster took %s", time.Since(start))
	}()

	version, err := util.ParseKubernetesVersion(cfg.KubernetesConfig.KubernetesVersion)
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

	cp, err := config.PrimaryControlPlane(&cfg)
	if err != nil {
		return errors.Wrap(err, "primary control plane")
	}

	hostname, _, port, err := driver.ControlPaneEndpoint(&cfg, &cp, cfg.Driver)
	if err != nil {
		return errors.Wrap(err, "control plane")
	}

	client, err := k.client(hostname, port)
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}

	// If the cluster is running, check if we have any work to do.
	conf := bsutil.KubeadmYamlPath
	if !k.needsReset(conf, hostname, port, client, cfg.KubernetesConfig.KubernetesVersion) {
		glog.Infof("Taking a shortcut, as the cluster seems to be properly configured")
		return nil
	}

	if err := k.clearStaleConfigs(cfg); err != nil {
		return errors.Wrap(err, "clearing stale configs")
	}

	if _, err := k.c.RunCmd(exec.Command("sudo", "cp", conf+".new", conf)); err != nil {
		return errors.Wrap(err, "cp")
	}

	baseCmd := fmt.Sprintf("%s %s", bsutil.InvokeKubeadm(cfg.KubernetesConfig.KubernetesVersion), phase)
	cmds := []string{
		fmt.Sprintf("%s phase certs all --config %s", baseCmd, conf),
		fmt.Sprintf("%s phase kubeconfig all --config %s", baseCmd, conf),
		fmt.Sprintf("%s phase %s all --config %s", baseCmd, controlPlane, conf),
		fmt.Sprintf("%s phase etcd local --config %s", baseCmd, conf),
	}

	glog.Infof("resetting cluster from %s", conf)
	// Run commands one at a time so that it is easier to root cause failures.
	for _, c := range cmds {
		_, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", c))
		if err != nil {
			return errors.Wrap(err, "run")
		}
	}

	cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: k.c})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	// We must ensure that the apiserver is healthy before proceeding
	if err := kverify.WaitForAPIServerProcess(cr, k, cfg, k.c, time.Now(), kconst.DefaultControlPlaneTimeout); err != nil {
		return errors.Wrap(err, "apiserver healthz")
	}

	if err := kverify.WaitForHealthyAPIServer(cr, k, cfg, k.c, client, time.Now(), hostname, port, kconst.DefaultControlPlaneTimeout); err != nil {
		return errors.Wrap(err, "apiserver health")
	}

	if err := kverify.WaitForSystemPods(cr, k, cfg, k.c, client, time.Now(), kconst.DefaultControlPlaneTimeout); err != nil {
		return errors.Wrap(err, "system pods")
	}

	// This can fail during upgrades if the old pods have not shut down yet
	addonPhase := func() error {
		_, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s phase addon all --config %s", baseCmd, conf)))
		return err
	}
	if err = retry.Expo(addonPhase, 1*time.Second, 30*time.Second); err != nil {
		glog.Warningf("addon install failed, wil retry: %v", err)
		return errors.Wrap(err, "addons")
	}

	if err := bsutil.AdjustResourceLimits(k.c); err != nil {
		glog.Warningf("unable to adjust resource limits: %v", err)
	}
	return nil
}

// JoinCluster adds a node to an existing cluster
func (k *Bootstrapper) JoinCluster(cc config.ClusterConfig, n config.Node, joinCmd string) error {
	start := time.Now()
	glog.Infof("JoinCluster: %+v", cc)
	defer func() {
		glog.Infof("JoinCluster complete in %s", time.Since(start))
	}()

	// Join the master by specifying its token
	joinCmd = fmt.Sprintf("%s --v=10 --node-name=%s", joinCmd, driver.MachineName(cc, n))
	out, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", joinCmd))
	if err != nil {
		return errors.Wrapf(err, "cmd failed: %s\n%+v\n", joinCmd, out)
	}

	if _, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", "sudo systemctl daemon-reload && sudo systemctl enable kubelet && sudo systemctl start kubelet")); err != nil {
		return errors.Wrap(err, "starting kubelet")
	}

	return nil
}

// GenerateToken creates a token and returns the appropriate kubeadm join command to run
func (k *Bootstrapper) GenerateToken(cc config.ClusterConfig) (string, error) {
	tokenCmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s token create --print-join-command --ttl=0", bsutil.InvokeKubeadm(cc.KubernetesConfig.KubernetesVersion)))
	r, err := k.c.RunCmd(tokenCmd)
	if err != nil {
		return "", errors.Wrap(err, "generating bootstrap token")
	}

	joinCmd := r.Stdout.String()
	joinCmd = strings.Replace(joinCmd, "kubeadm", bsutil.InvokeKubeadm(cc.KubernetesConfig.KubernetesVersion), 1)
	joinCmd = fmt.Sprintf("%s --ignore-preflight-errors=all", strings.TrimSpace(joinCmd))

	return joinCmd, nil
}

// DeleteCluster removes the components that were started earlier
func (k *Bootstrapper) DeleteCluster(k8s config.KubernetesConfig) error {
	version, err := util.ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing kubernetes version")
	}

	cmd := fmt.Sprintf("%s reset --force", bsutil.InvokeKubeadm(k8s.KubernetesVersion))
	if version.LT(semver.MustParse("1.11.0")) {
		cmd = fmt.Sprintf("%s reset", bsutil.InvokeKubeadm(k8s.KubernetesVersion))
	}

	rr, derr := k.c.RunCmd(exec.Command("/bin/bash", "-c", cmd))
	if derr != nil {
		glog.Warningf("%s: %v", rr.Command(), err)
	}

	if err := kubelet.ForceStop(k.c); err != nil {
		glog.Warningf("stop kubelet: %v", err)
	}

	cr, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: k.c, Socket: k8s.CRISocket})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	containers, err := cr.ListContainers(cruntime.ListOptions{Namespaces: []string{"kube-system"}})
	if err != nil {
		glog.Warningf("unable to list kube-system containers: %v", err)
	}
	if len(containers) > 0 {
		glog.Warningf("found %d kube-system containers to stop", len(containers))
		if err := cr.StopContainers(containers); err != nil {
			glog.Warningf("error stopping containers: %v", err)
		}
	}

	return derr
}

// SetupCerts sets up certificates within the cluster.
func (k *Bootstrapper) SetupCerts(k8s config.KubernetesConfig, n config.Node) error {
	_, err := bootstrapper.SetupCerts(k.c, k8s, n)
	return err
}

// UpdateCluster updates the cluster.
func (k *Bootstrapper) UpdateCluster(cfg config.ClusterConfig) error {
	images, err := images.Kubeadm(cfg.KubernetesConfig.ImageRepository, cfg.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "kubeadm images")
	}

	r, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime,
		Runner: k.c, Socket: cfg.KubernetesConfig.CRISocket})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	if err := r.Preload(cfg.KubernetesConfig); err != nil {
		glog.Infof("prelaoding failed, will try to load cached images: %v", err)
	}

	if cfg.KubernetesConfig.ShouldLoadCachedImages {
		if err := machine.LoadImages(&cfg, k.c, images, constants.ImageCacheDir); err != nil {
			out.FailureT("Unable to load cached images: {{.error}}", out.V{"error": err})
		}
	}

	for _, n := range cfg.Nodes {
		err := k.UpdateNode(cfg, n, r)
		if err != nil {
			return errors.Wrap(err, "updating node")
		}
	}

	return nil
}

// UpdateNode updates a node.
func (k *Bootstrapper) UpdateNode(cfg config.ClusterConfig, n config.Node, r cruntime.Manager) error {
	kubeadmCfg, err := bsutil.GenerateKubeadmYAML(cfg, n, r)
	if err != nil {
		return errors.Wrap(err, "generating kubeadm cfg")
	}

	kubeletCfg, err := bsutil.NewKubeletConfig(cfg, n, r)
	if err != nil {
		return errors.Wrap(err, "generating kubelet config")
	}

	kubeletService, err := bsutil.NewKubeletService(cfg.KubernetesConfig)
	if err != nil {
		return errors.Wrap(err, "generating kubelet service")
	}

	glog.Infof("kubelet %s config:\n%+v", kubeletCfg, cfg.KubernetesConfig)

	if err := bsutil.TransferBinaries(cfg.KubernetesConfig, k.c); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}

	var cniFile []byte
	if cfg.KubernetesConfig.EnableDefaultCNI {
		cniFile = []byte(defaultCNIConfig)
	}

	// Install assets into temporary files
	files := bsutil.ConfigFileAssets(cfg.KubernetesConfig, kubeadmCfg, kubeletCfg, kubeletService, cniFile)
	if err := copyFiles(k.c, files); err != nil {
		return err
	}

	if err := reloadKubelet(k.c); err != nil {
		return err
	}
	return nil
}

func copyFiles(runner command.Runner, files []assets.CopyableFile) error {
	// Combine mkdir request into a single call to reduce load
	dirs := []string{}
	for _, f := range files {
		dirs = append(dirs, f.GetTargetDir())
	}
	args := append([]string{"mkdir", "-p"}, dirs...)
	if _, err := runner.RunCmd(exec.Command("sudo", args...)); err != nil {
		return errors.Wrap(err, "mkdir")
	}

	for _, f := range files {
		if err := runner.Copy(f); err != nil {
			return errors.Wrapf(err, "copy")
		}
	}
	return nil
}

func reloadKubelet(runner command.Runner) error {
	svc := bsutil.KubeletServiceFile
	conf := bsutil.KubeletSystemdConfFile

	checkCmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("pgrep kubelet && diff -u %s %s.new && diff -u %s %s.new", svc, svc, conf, conf))
	if _, err := runner.RunCmd(checkCmd); err == nil {
		glog.Infof("kubelet is already running with the right configs")
		return nil
	}

	startCmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("sudo cp %s.new %s && sudo cp %s.new %s && sudo systemctl daemon-reload && sudo systemctl restart kubelet", svc, svc, conf, conf))
	if _, err := runner.RunCmd(startCmd); err != nil {
		return errors.Wrap(err, "starting kubelet")
	}
	return nil
}

// applyKicOverlay applies the CNI plugin needed to make kic work
func (k *Bootstrapper) applyKicOverlay(cfg config.ClusterConfig) error {
	// Allow no more than 5 seconds for apply kic overlay
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sudo",
		path.Join(vmpath.GuestPersistentDir, "binaries", cfg.KubernetesConfig.KubernetesVersion, "kubectl"), "create", fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
		"-f", "-")
	b := bytes.Buffer{}
	if err := kicCNIConfig.Execute(&b, struct{ ImageName string }{ImageName: kic.OverlayImage}); err != nil {
		return err
	}
	cmd.Stdin = bytes.NewReader(b.Bytes())
	if rr, err := k.c.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "cmd: %s output: %s", rr.Command(), rr.Output())
	}
	return nil
}

// applyNodeLabels applies minikube labels to all the nodes
func (k *Bootstrapper) applyNodeLabels(cfg config.ClusterConfig) error {
	// time cluster was created. time format is based on ISO 8601 (RFC 3339)
	// converting - and : to _ because of kubernetes label restriction
	createdAtLbl := "minikube.k8s.io/updated_at=" + time.Now().Format("2006_01_02T15_04_05_0700")
	verLbl := "minikube.k8s.io/version=" + version.GetVersion()
	commitLbl := "minikube.k8s.io/commit=" + version.GetGitCommitID()
	nameLbl := "minikube.k8s.io/name=" + cfg.Name

	// Allow no more than 5 seconds for applying labels
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// example:
	// sudo /var/lib/minikube/binaries/<version>/kubectl label nodes minikube.k8s.io/version=<version> minikube.k8s.io/commit=aa91f39ffbcf27dcbb93c4ff3f457c54e585cf4a-dirty minikube.k8s.io/name=p1 minikube.k8s.io/updated_at=2020_02_20T12_05_35_0700 --all --overwrite --kubeconfig=/var/lib/minikube/kubeconfig
	cmd := exec.CommandContext(ctx, "sudo",
		path.Join(vmpath.GuestPersistentDir, "binaries", cfg.KubernetesConfig.KubernetesVersion, "kubectl"),
		"label", "nodes", verLbl, commitLbl, nameLbl, createdAtLbl, "--all", "--overwrite",
		fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")))

	if _, err := k.c.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "applying node labels")
	}
	return nil
}

// elevateKubeSystemPrivileges gives the kube-system service account cluster admin privileges to work with RBAC.
func (k *Bootstrapper) elevateKubeSystemPrivileges(cfg config.ClusterConfig) error {
	start := time.Now()
	// Allow no more than 5 seconds for creating cluster role bindings
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rbacName := "minikube-rbac"
	// kubectl create clusterrolebinding minikube-rbac --clusterrole=cluster-admin --serviceaccount=kube-system:default
	cmd := exec.CommandContext(ctx, "sudo",
		path.Join(vmpath.GuestPersistentDir, "binaries", cfg.KubernetesConfig.KubernetesVersion, "kubectl"),
		"create", "clusterrolebinding", rbacName, "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default",
		fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")))
	rr, err := k.c.RunCmd(cmd)
	if err != nil {
		// Error from server (AlreadyExists): clusterrolebindings.rbac.authorization.k8s.io "minikube-rbac" already exists
		if strings.Contains(rr.Output(), fmt.Sprintf("Error from server (AlreadyExists)")) {
			glog.Infof("rbac %q already exists not need to re-create.", rbacName)
			return nil
		}
	}
	glog.Infof("duration metric: took %s to wait for elevateKubeSystemPrivileges.", time.Since(start))
	return err
}
