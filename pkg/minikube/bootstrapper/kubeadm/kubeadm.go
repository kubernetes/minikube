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
	"runtime"
	"sync"

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
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
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
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
	"k8s.io/minikube/pkg/version"
)

// Bootstrapper is a bootstrapper using kubeadm
type Bootstrapper struct {
	c           command.Runner
	k8sClient   *kubernetes.Clientset // Kubernetes client used to verify pods inside cluster
	contextName string
}

// NewBootstrapper creates a new kubeadm.Bootstrapper
func NewBootstrapper(api libmachine.API, cc config.ClusterConfig, r command.Runner) (*Bootstrapper, error) {
	return &Bootstrapper{c: r, contextName: cc.Name, k8sClient: nil}, nil
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

	describeNodes := fmt.Sprintf("sudo %s describe nodes --kubeconfig=%s", kubectlPath(cfg),
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
	// These are the files that kubeadm will reject stale versions of
	paths := []string{
		"/etc/kubernetes/admin.conf",
		"/etc/kubernetes/kubelet.conf",
		"/etc/kubernetes/controller-manager.conf",
		"/etc/kubernetes/scheduler.conf",
	}

	args := append([]string{"ls", "-la"}, paths...)
	rr, err := k.c.RunCmd(exec.Command("sudo", args...))
	if err != nil {
		glog.Infof("config check failed, skipping stale config cleanup: %v", err)
		return nil
	}
	glog.Infof("found existing configuration files:\n%s\n", rr.Stdout.String())

	cp, err := config.PrimaryControlPlane(&cfg)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("https://%s", net.JoinHostPort(constants.ControlPlaneAlias, strconv.Itoa(cp.Port)))
	for _, path := range paths {
		_, err := k.c.RunCmd(exec.Command("sudo", "grep", endpoint, path))
		if err != nil {
			glog.Infof("%q may not be in %s - will remove: %v", endpoint, path, err)

			_, err := k.c.RunCmd(exec.Command("sudo", "rm", "-f", path))
			if err != nil {
				glog.Errorf("rm failed: %v", err)
			}
		}
	}
	return nil
}

func (k *Bootstrapper) init(cfg config.ClusterConfig) error {
	version, err := util.ParseKubernetesVersion(cfg.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kubernetes version")
	}

	extraFlags := bsutil.CreateFlagsFromExtraArgs(cfg.KubernetesConfig.ExtraOptions)
	r, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: k.c})
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
	}
	ignore = append(ignore, bsutil.SkipAdditionalPreflights[r.Name()]...)

	skipSystemVerification := false
	// Allow older kubeadm versions to function with newer Docker releases.
	if version.LT(semver.MustParse("1.13.0")) {
		glog.Infof("ignoring SystemVerification for kubeadm because of old Kubernetes version %v", version)
		skipSystemVerification = true
	}
	if driver.BareMetal(cfg.Driver) && r.Name() == "Docker" {
		if v, err := r.Version(); err == nil && strings.Contains(v, "azure") {
			glog.Infof("ignoring SystemVerification for kubeadm because of unknown docker version %s", v)
			skipSystemVerification = true
		}
	}
	// For kic on linux example error: "modprobe: FATAL: Module configs not found in directory /lib/modules/5.2.17-1rodete3-amd64"
	if driver.IsKIC(cfg.Driver) {
		glog.Infof("ignoring SystemVerification for kubeadm because of %s driver", cfg.Driver)
		skipSystemVerification = true
	}
	if skipSystemVerification {
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

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		// we need to have cluster role binding before applying overlay to avoid #7428
		if err := k.elevateKubeSystemPrivileges(cfg); err != nil {
			glog.Errorf("unable to create cluster role binding, some addons might not work: %v", err)
		}
		// the overlay is required for containerd and cri-o runtime: see #7428
		if config.MultiNode(cfg) || (driver.IsKIC(cfg.Driver) && cfg.KubernetesConfig.ContainerRuntime != "docker") {
			if err := k.applyKICOverlay(cfg); err != nil {
				glog.Errorf("failed to apply kic overlay: %v", err)
			}
		}
		wg.Done()
	}()

	go func() {
		if err := k.applyNodeLabels(cfg); err != nil {
			glog.Warningf("unable to apply node labels: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := bsutil.AdjustResourceLimits(k.c); err != nil {
			glog.Warningf("unable to adjust resource limits: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()
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
		glog.Warningf("Overriding stale ClientConfig host %s with %s", cc.Host, endpoint)
		cc.Host = endpoint
	}
	c, err := kubernetes.NewForConfig(cc)
	if err == nil {
		k.k8sClient = c
	}
	return c, err
}

// WaitForNode blocks until the node appears to be healthy
func (k *Bootstrapper) WaitForNode(cfg config.ClusterConfig, n config.Node, timeout time.Duration) (waitErr error) {
	start := time.Now()

	if !n.ControlPlane {
		glog.Infof("%s is not a control plane, nothing to wait for", n.Name)
		return nil
	}

	out.T(out.HealthCheck, "Verifying Kubernetes components...")

	// TODO: #7706: for better performance we could use k.client inside minikube to avoid asking for external IP:PORT
	hostname, _, port, err := driver.ControlPaneEndpoint(&cfg, &n, cfg.Driver)
	if err != nil {
		return errors.Wrap(err, "get control plane endpoint")
	}

	defer func() { // run pressure verification after all other checks, so there be an api server to talk to.
		client, err := k.client(hostname, port)
		if err != nil {
			waitErr = errors.Wrap(err, "get k8s client")
		}
		if err := kverify.NodePressure(client); err != nil {
			adviseNodePressure(err, cfg.Name, cfg.Driver)
			waitErr = errors.Wrap(err, "node pressure")
		}
	}()

	if !kverify.ShouldWait(cfg.VerifyComponents) {
		glog.Infof("skip waiting for components based on config.")
		return waitErr
	}

	cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: k.c})
	if err != nil {
		return errors.Wrapf(err, "create runtme-manager %s", cfg.KubernetesConfig.ContainerRuntime)
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

	if cfg.VerifyComponents[kverify.AppsRunningKey] {
		client, err := k.client(hostname, port)
		if err != nil {
			return errors.Wrap(err, "get k8s client")
		}
		if err := kverify.WaitForAppsRunning(client, kverify.AppsRunningList, timeout); err != nil {
			return errors.Wrap(err, "waiting for apps_running")
		}
	}

	if cfg.VerifyComponents[kverify.NodeReadyKey] {
		client, err := k.client(hostname, port)
		if err != nil {
			return errors.Wrap(err, "get k8s client")
		}
		if err := kverify.WaitForNodeReady(client, timeout); err != nil {
			return errors.Wrap(err, "waiting for node to be ready")
		}
	}

	glog.Infof("duration metric: took %s to wait for : %+v ...", time.Since(start), cfg.VerifyComponents)
	return waitErr
}

// needsReconfigure returns whether or not the cluster needs to be reconfigured
func (k *Bootstrapper) needsReconfigure(conf string, hostname string, port int, client *kubernetes.Clientset, version string) bool {
	if rr, err := k.c.RunCmd(exec.Command("sudo", "diff", "-u", conf, conf+".new")); err != nil {
		glog.Infof("needs reconfigure: configs differ:\n%s", rr.Output())
		return true
	}

	st, err := kverify.APIServerStatus(k.c, hostname, port)
	if err != nil {
		glog.Infof("needs reconfigure: apiserver error: %v", err)
		return true
	}

	if st != state.Running {
		glog.Infof("needs reconfigure: apiserver in state %s", st)
		return true
	}

	if err := kverify.ExpectAppsRunning(client, kverify.AppsRunningList); err != nil {
		glog.Infof("needs reconfigure: %v", err)
		return true
	}

	if err := kverify.APIServerVersionMatch(client, version); err != nil {
		glog.Infof("needs reconfigure: %v", err)
		return true
	}

	// DANGER: This log message is hard-coded in an integration test!
	glog.Infof("The running cluster does not require reconfiguration: %s", hostname)
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
		return errors.Wrap(err, "parsing Kubernetes version")
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
	if !k.needsReconfigure(conf, hostname, port, client, cfg.KubernetesConfig.KubernetesVersion) {
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

	glog.Infof("reconfiguring cluster from %s", conf)
	// Run commands one at a time so that it is easier to root cause failures.
	for _, c := range cmds {
		if _, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", c)); err != nil {
			glog.Errorf("%s failed - will try once more: %v", c, err)

			if _, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", c)); err != nil {
				return errors.Wrap(err, "run")
			}
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

	if err := kverify.NodePressure(client); err != nil {
		adviseNodePressure(err, cfg.Name, cfg.Driver)
	}

	// This can fail during upgrades if the old pods have not shut down yet
	addonPhase := func() error {
		_, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s phase addon all --config %s", baseCmd, conf)))
		return err
	}
	if err = retry.Expo(addonPhase, 100*time.Microsecond, 30*time.Second); err != nil {
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
	cr, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: k.c, Socket: k8s.CRISocket})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	version, err := util.ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kubernetes version")
	}

	ka := bsutil.InvokeKubeadm(k8s.KubernetesVersion)
	sp := cr.SocketPath()
	if sp == "" {
		sp = kconst.DefaultDockerCRISocket
	}
	cmd := fmt.Sprintf("%s reset --cri-socket %s --force", ka, sp)
	if version.LT(semver.MustParse("1.11.0")) {
		cmd = fmt.Sprintf("%s reset --cri-socket %s", ka, sp)
	}

	rr, derr := k.c.RunCmd(exec.Command("/bin/bash", "-c", cmd))
	if derr != nil {
		glog.Warningf("%s: %v", rr.Command(), err)
	}

	if err := sysinit.New(k.c).ForceStop("kubelet"); err != nil {
		glog.Warningf("stop kubelet: %v", err)
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

	sm := sysinit.New(k.c)

	if err := bsutil.TransferBinaries(cfg.KubernetesConfig, k.c, sm); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}

	files := []assets.CopyableFile{
		assets.NewMemoryAssetTarget(kubeadmCfg, bsutil.KubeadmYamlPath+".new", "0640"),
		assets.NewMemoryAssetTarget(kubeletCfg, bsutil.KubeletSystemdConfFile+".new", "0644"),
		assets.NewMemoryAssetTarget(kubeletService, bsutil.KubeletServiceFile+".new", "0644"),
	}
	// Copy the default CNI config (k8s.conf), so that kubelet can successfully
	// start a Pod in the case a user hasn't manually installed any CNI plugin
	// and minikube was started with "--extra-config=kubelet.network-plugin=cni".
	if cfg.KubernetesConfig.EnableDefaultCNI {
		files = append(files, assets.NewMemoryAssetTarget([]byte(defaultCNIConfig), bsutil.DefaultCNIConfigPath, "0644"))
	}

	// Installs compatibility shims for non-systemd environments
	kubeletPath := path.Join(vmpath.GuestPersistentDir, "binaries", cfg.KubernetesConfig.KubernetesVersion, "kubelet")
	shims, err := sm.GenerateInitShim("kubelet", kubeletPath, bsutil.KubeletSystemdConfFile)
	if err != nil {
		return errors.Wrap(err, "shim")
	}
	files = append(files, shims...)

	if err := copyFiles(k.c, files); err != nil {
		return errors.Wrap(err, "copy")
	}

	cp, err := config.PrimaryControlPlane(&cfg)
	if err != nil {
		return errors.Wrap(err, "control plane")
	}

	if err := machine.AddHostAlias(k.c, constants.ControlPlaneAlias, net.ParseIP(cp.IP)); err != nil {
		return errors.Wrap(err, "host alias")
	}

	return sm.Start("kubelet")
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

// kubectlPath returns the path to the kubelet
func kubectlPath(cfg config.ClusterConfig) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", cfg.KubernetesConfig.KubernetesVersion, "kubectl")
}

// applyKICOverlay applies the CNI plugin needed to make kic work
func (k *Bootstrapper) applyKICOverlay(cfg config.ClusterConfig) error {
	b := bytes.Buffer{}
	if err := kicCNIConfig.Execute(&b, struct{ ImageName string }{ImageName: kic.OverlayImage}); err != nil {
		return err
	}

	ko := path.Join(vmpath.GuestEphemeralDir, "kic_overlay.yaml")
	f := assets.NewMemoryAssetTarget(b.Bytes(), ko, "0644")

	if err := k.c.Copy(f); err != nil {
		return errors.Wrapf(err, "copy")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", kubectlPath(cfg), "apply",
		fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
		"-f", ko)

	if rr, err := k.c.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "cmd: %s output: %s", rr.Command(), rr.Output())
	}

	// Inform cri-o that the CNI has changed
	if cfg.KubernetesConfig.ContainerRuntime == "crio" {
		if err := sysinit.New(k.c).Restart("crio"); err != nil {
			return errors.Wrap(err, "restart crio")
		}
	}

	return nil
}

// applyNodeLabels applies minikube labels to all the nodes
func (k *Bootstrapper) applyNodeLabels(cfg config.ClusterConfig) error {
	// time cluster was created. time format is based on ISO 8601 (RFC 3339)
	// converting - and : to _ because of Kubernetes label restriction
	createdAtLbl := "minikube.k8s.io/updated_at=" + time.Now().Format("2006_01_02T15_04_05_0700")
	verLbl := "minikube.k8s.io/version=" + version.GetVersion()
	commitLbl := "minikube.k8s.io/commit=" + version.GetGitCommitID()
	nameLbl := "minikube.k8s.io/name=" + cfg.Name

	// Allow no more than 5 seconds for applying labels
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// example:
	// sudo /var/lib/minikube/binaries/<version>/kubectl label nodes minikube.k8s.io/version=<version> minikube.k8s.io/commit=aa91f39ffbcf27dcbb93c4ff3f457c54e585cf4a-dirty minikube.k8s.io/name=p1 minikube.k8s.io/updated_at=2020_02_20T12_05_35_0700 --all --overwrite --kubeconfig=/var/lib/minikube/kubeconfig
	cmd := exec.CommandContext(ctx, "sudo", kubectlPath(cfg),
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
	defer func() {
		glog.Infof("duration metric: took %s to wait for elevateKubeSystemPrivileges.", time.Since(start))
	}()

	// Allow no more than 5 seconds for creating cluster role bindings
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rbacName := "minikube-rbac"
	// kubectl create clusterrolebinding minikube-rbac --clusterrole=cluster-admin --serviceaccount=kube-system:default
	cmd := exec.CommandContext(ctx, "sudo", kubectlPath(cfg),
		"create", "clusterrolebinding", rbacName, "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default",
		fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")))
	rr, err := k.c.RunCmd(cmd)
	if err != nil {
		// Error from server (AlreadyExists): clusterrolebindings.rbac.authorization.k8s.io "minikube-rbac" already exists
		if strings.Contains(rr.Output(), "Error from server (AlreadyExists)") {
			glog.Infof("rbac %q already exists not need to re-create.", rbacName)
			return nil
		}
	}

	if cfg.VerifyComponents[kverify.DefaultSAWaitKey] {
		// double checking defalut sa was created.
		// good for ensuring using minikube in CI is robust.
		checkSA := func() (bool, error) {
			cmd = exec.Command("sudo", kubectlPath(cfg),
				"get", "sa", "default", fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")))
			rr, err = k.c.RunCmd(cmd)
			if err != nil {
				return false, nil
			}
			return true, nil
		}

		// retry up to make sure SA is created
		if err := wait.PollImmediate(kconst.APICallRetryInterval, time.Minute, checkSA); err != nil {
			return errors.Wrap(err, "ensure sa was created")
		}
	}
	return nil
}

// adviseNodePressure will advise the user what to do with difference pressure errors based on their environment
func adviseNodePressure(err error, name string, drv string) {
	if diskErr, ok := err.(*kverify.ErrDiskPressure); ok {
		out.ErrLn("")
		glog.Warning(diskErr)
		out.WarningT("The node {{.name}} has ran out of disk space.", out.V{"name": name})
		// generic advice for all drivers
		out.T(out.Tip, "Please free up disk or prune images.")
		if driver.IsVM(drv) {
			out.T(out.Stopped, "Please create a cluster with bigger disk size: `minikube start --disk SIZE_MB` ")
		} else if drv == oci.Docker && runtime.GOOS != "linux" {
			out.T(out.Stopped, "Please increse Desktop's disk size.")
			if runtime.GOOS == "darwin" {
				out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
			}
			if runtime.GOOS == "windows" {
				out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
			}
		}
		out.ErrLn("")
		return
	}

	if memErr, ok := err.(*kverify.ErrMemoryPressure); ok {
		out.ErrLn("")
		glog.Warning(memErr)
		out.WarningT("The node {{.name}} has ran out of memory.", out.V{"name": name})
		out.T(out.Tip, "Check if you have unnecessary pods running by running 'kubectl get po -A")
		if driver.IsVM(drv) {
			out.T(out.Stopped, "Consider creating a cluster with larger memory size using `minikube start --memory SIZE_MB` ")
		} else if drv == oci.Docker && runtime.GOOS != "linux" {
			out.T(out.Stopped, "Consider increasing Docker Desktop's memory size.")
			if runtime.GOOS == "darwin" {
				out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
			}
			if runtime.GOOS == "windows" {
				out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
			}
		}
		out.ErrLn("")
		return
	}

	if pidErr, ok := err.(*kverify.ErrPIDPressure); ok {
		glog.Warning(pidErr)
		out.ErrLn("")
		out.WarningT("The node {{.name}} has ran out of available PIDs.", out.V{"name": name})
		out.ErrLn("")
		return
	}

	if netErr, ok := err.(*kverify.ErrNetworkNotReady); ok {
		glog.Warning(netErr)
		out.ErrLn("")
		out.WarningT("The node {{.name}} network is not available. Please verify network settings.", out.V{"name": name})
		out.ErrLn("")
		return
	}
}
