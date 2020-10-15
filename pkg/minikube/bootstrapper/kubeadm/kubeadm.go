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
	"context"
	"fmt"
	"net"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	// WARNING: Do not use path/filepath in this package unless you want bizarre Windows paths

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/style"
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
		klog.Infof("%s skipping compat symlinks: %v", legacyEtcd, err)
		return nil
	}
	klog.Infof("Found %s, creating compatibility symlinks ...", legacyEtcd)

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
		klog.Infof("config check failed, skipping stale config cleanup: %v", err)
		return nil
	}
	klog.Infof("found existing configuration files:\n%s\n", rr.Stdout.String())

	cp, err := config.PrimaryControlPlane(&cfg)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("https://%s", net.JoinHostPort(constants.ControlPlaneAlias, strconv.Itoa(cp.Port)))
	for _, path := range paths {
		_, err := k.c.RunCmd(exec.Command("sudo", "grep", endpoint, path))
		if err != nil {
			klog.Infof("%q may not be in %s - will remove: %v", endpoint, path, err)

			_, err := k.c.RunCmd(exec.Command("sudo", "rm", "-f", path))
			if err != nil {
				klog.Errorf("rm failed: %v", err)
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
		klog.Infof("ignoring SystemVerification for kubeadm because of old Kubernetes version %v", version)
		skipSystemVerification = true
	}
	if driver.BareMetal(cfg.Driver) && r.Name() == "Docker" {
		if v, err := r.Version(); err == nil && strings.Contains(v, "azure") {
			klog.Infof("ignoring SystemVerification for kubeadm because of unknown docker version %s", v)
			skipSystemVerification = true
		}
	}
	// For kic on linux example error: "modprobe: FATAL: Module configs not found in directory /lib/modules/5.2.17-1rodete3-amd64"
	if driver.IsKIC(cfg.Driver) {
		klog.Infof("ignoring SystemVerification for kubeadm because of %s driver", cfg.Driver)
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
	ctx, cancel := context.WithTimeout(context.Background(), initTimeoutMinutes*time.Minute)
	defer cancel()
	c := exec.CommandContext(ctx, "/bin/bash", "-c", fmt.Sprintf("%s init --config %s %s --ignore-preflight-errors=%s",
		bsutil.InvokeKubeadm(cfg.KubernetesConfig.KubernetesVersion), conf, extraFlags, strings.Join(ignore, ",")))
	if _, err := k.c.RunCmd(c); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrInitTimedout
		}

		if strings.Contains(err.Error(), "'kubeadm': Permission denied") {
			return ErrNoExecLinux
		}
		return errors.Wrap(err, "run")
	}

	if err := k.applyCNI(cfg); err != nil {
		return errors.Wrap(err, "apply cni")
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		// we need to have cluster role binding before applying overlay to avoid #7428
		if err := k.elevateKubeSystemPrivileges(cfg); err != nil {
			klog.Errorf("unable to create cluster role binding, some addons might not work: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := k.applyNodeLabels(cfg); err != nil {
			klog.Warningf("unable to apply node labels: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := bsutil.AdjustResourceLimits(k.c); err != nil {
			klog.Warningf("unable to adjust resource limits: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()
	return nil
}

// applyCNI applies CNI to a cluster. Needs to be done every time a VM is powered up.
func (k *Bootstrapper) applyCNI(cfg config.ClusterConfig) error {
	cnm, err := cni.New(cfg)
	if err != nil {
		return errors.Wrap(err, "cni config")
	}

	if _, ok := cnm.(cni.Disabled); ok {
		return nil
	}

	out.T(style.CNI, "Configuring {{.name}} (Container Networking Interface) ...", out.V{"name": cnm.String()})

	if err := cnm.Apply(k.c); err != nil {
		return errors.Wrap(err, "cni apply")
	}

	if cfg.KubernetesConfig.ContainerRuntime == constants.CRIO {
		if err := cruntime.UpdateCRIONet(k.c, cnm.CIDR()); err != nil {
			return errors.Wrap(err, "update crio")
		}
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
	klog.Infof("StartCluster: %+v", cfg)
	defer func() {
		klog.Infof("StartCluster complete in %s", time.Since(start))
	}()

	// Before we start, ensure that no paused components are lurking around
	if err := k.unpause(cfg); err != nil {
		klog.Warningf("unpause failed: %v", err)
	}

	if err := bsutil.ExistingConfig(k.c); err == nil {
		klog.Infof("found existing configuration files, will attempt cluster restart")
		rerr := k.restartControlPlane(cfg)
		if rerr == nil {
			return nil
		}

		out.ErrT(style.Embarrassed, "Unable to restart cluster, will reset it: {{.error}}", out.V{"error": rerr})
		if err := k.DeleteCluster(cfg.KubernetesConfig); err != nil {
			klog.Warningf("delete failed: %v", err)
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

	// retry again if it is not a fail fast error
	if _, ff := err.(*FailFastError); !ff {
		out.ErrT(style.Conflict, "initialization failed, will try again: {{.error}}", out.V{"error": err})
		if err := k.DeleteCluster(cfg.KubernetesConfig); err != nil {
			klog.Warningf("delete failed: %v", err)
		}
		return k.init(cfg)
	}
	return err
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
		klog.Warningf("Overriding stale ClientConfig host %s with %s", cc.Host, endpoint)
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
	register.Reg.SetStep(register.VerifyingKubernetes)
	out.T(style.HealthCheck, "Verifying Kubernetes components...")
	// regardless if waiting is set or not, we will make sure kubelet is not stopped
	// to solve corner cases when a container is hibernated and once coming back kubelet not running.
	if err := k.ensureServiceStarted("kubelet"); err != nil {
		klog.Warningf("Couldn't ensure kubelet is started this might cause issues: %v", err)
	}
	// TODO: #7706: for better performance we could use k.client inside minikube to avoid asking for external IP:PORT
	cp, err := config.PrimaryControlPlane(&cfg)
	if err != nil {
		return errors.Wrap(err, "get primary control plane")
	}
	hostname, _, port, err := driver.ControlPlaneEndpoint(&cfg, &cp, cfg.Driver)
	if err != nil {
		return errors.Wrap(err, "get control plane endpoint")
	}

	client, err := k.client(hostname, port)
	if err != nil {
		return errors.Wrap(err, "kubernetes client")
	}

	if !kverify.ShouldWait(cfg.VerifyComponents) {
		klog.Infof("skip waiting for components based on config.")

		if err := kverify.NodePressure(client); err != nil {
			adviseNodePressure(err, cfg.Name, cfg.Driver)
			return errors.Wrap(err, "node pressure")
		}
		return nil
	}

	cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: k.c})
	if err != nil {
		return errors.Wrapf(err, "create runtme-manager %s", cfg.KubernetesConfig.ContainerRuntime)
	}

	if n.ControlPlane {
		if cfg.VerifyComponents[kverify.APIServerWaitKey] {
			if err := kverify.WaitForAPIServerProcess(cr, k, cfg, k.c, start, timeout); err != nil {
				return errors.Wrap(err, "wait for apiserver proc")
			}

			if err := kverify.WaitForHealthyAPIServer(cr, k, cfg, k.c, client, start, hostname, port, timeout); err != nil {
				return errors.Wrap(err, "wait for healthy API server")
			}
		}

		if cfg.VerifyComponents[kverify.SystemPodsWaitKey] {
			if err := kverify.WaitForSystemPods(cr, k, cfg, k.c, client, start, timeout); err != nil {
				return errors.Wrap(err, "waiting for system pods")
			}
		}

		if cfg.VerifyComponents[kverify.DefaultSAWaitKey] {
			if err := kverify.WaitForDefaultSA(client, timeout); err != nil {
				return errors.Wrap(err, "waiting for default service account")
			}
		}

		if cfg.VerifyComponents[kverify.AppsRunningKey] {
			if err := kverify.WaitForAppsRunning(client, kverify.AppsRunningList, timeout); err != nil {
				return errors.Wrap(err, "waiting for apps_running")
			}
		}
	}
	if cfg.VerifyComponents[kverify.KubeletKey] {
		if err := kverify.WaitForService(k.c, "kubelet", timeout); err != nil {
			return errors.Wrap(err, "waiting for kubelet")
		}

	}

	if cfg.VerifyComponents[kverify.NodeReadyKey] {
		if err := kverify.WaitForNodeReady(client, timeout); err != nil {
			return errors.Wrap(err, "waiting for node to be ready")
		}
	}

	klog.Infof("duration metric: took %s to wait for : %+v ...", time.Since(start), cfg.VerifyComponents)

	if err := kverify.NodePressure(client); err != nil {
		adviseNodePressure(err, cfg.Name, cfg.Driver)
		return errors.Wrap(err, "node pressure")
	}
	return nil
}

// ensureKubeletStarted will start a systemd or init.d service if it is not running.
func (k *Bootstrapper) ensureServiceStarted(svc string) error {
	if st := kverify.ServiceStatus(k.c, svc); st != state.Running {
		klog.Warningf("surprisingly %q service status was %s!. will try to start it, could be related to this issue https://github.com/kubernetes/minikube/issues/9458", svc, st)
		return sysinit.New(k.c).Start(svc)
	}
	return nil
}

// needsReconfigure returns whether or not the cluster needs to be reconfigured
func (k *Bootstrapper) needsReconfigure(conf string, hostname string, port int, client *kubernetes.Clientset, version string) bool {
	if rr, err := k.c.RunCmd(exec.Command("sudo", "diff", "-u", conf, conf+".new")); err != nil {
		klog.Infof("needs reconfigure: configs differ:\n%s", rr.Output())
		return true
	}

	st, err := kverify.APIServerStatus(k.c, hostname, port)
	if err != nil {
		klog.Infof("needs reconfigure: apiserver error: %v", err)
		return true
	}

	if st != state.Running {
		klog.Infof("needs reconfigure: apiserver in state %s", st)
		return true
	}

	if err := kverify.ExpectAppsRunning(client, kverify.AppsRunningList); err != nil {
		klog.Infof("needs reconfigure: %v", err)
		return true
	}

	if err := kverify.APIServerVersionMatch(client, version); err != nil {
		klog.Infof("needs reconfigure: %v", err)
		return true
	}

	// DANGER: This log message is hard-coded in an integration test!
	klog.Infof("The running cluster does not require reconfiguration: %s", hostname)
	return false
}

// restartCluster restarts the Kubernetes cluster configured by kubeadm
func (k *Bootstrapper) restartControlPlane(cfg config.ClusterConfig) error {
	klog.Infof("restartCluster start")

	start := time.Now()
	defer func() {
		klog.Infof("restartCluster took %s", time.Since(start))
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
		klog.Errorf("failed to create compat symlinks: %v", err)
	}

	cp, err := config.PrimaryControlPlane(&cfg)
	if err != nil {
		return errors.Wrap(err, "primary control plane")
	}

	hostname, _, port, err := driver.ControlPlaneEndpoint(&cfg, &cp, cfg.Driver)
	if err != nil {
		return errors.Wrap(err, "control plane")
	}

	// Save the costly tax of reinstalling Kubernetes if the only issue is a missing kube context
	_, err = kubeconfig.UpdateEndpoint(cfg.Name, hostname, port, kubeconfig.PathFromEnv())
	if err != nil {
		klog.Warningf("unable to update kubeconfig (cluster will likely require a reset): %v", err)
	}

	client, err := k.client(hostname, port)
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}

	// If the cluster is running, check if we have any work to do.
	conf := bsutil.KubeadmYamlPath
	if !k.needsReconfigure(conf, hostname, port, client, cfg.KubernetesConfig.KubernetesVersion) {
		klog.Infof("Taking a shortcut, as the cluster seems to be properly configured")
		return nil
	}

	if err := k.stopKubeSystem(cfg); err != nil {
		klog.Warningf("Failed to stop kube-system containers: port conflicts may arise: %v", err)
	}

	if err := sysinit.New(k.c).Stop("kubelet"); err != nil {
		klog.Warningf("Failed to stop kubelet, this might cause upgrade errors: %v", err)
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
		fmt.Sprintf("%s phase kubelet-start --config %s", baseCmd, conf),
		fmt.Sprintf("%s phase %s all --config %s", baseCmd, controlPlane, conf),
		fmt.Sprintf("%s phase etcd local --config %s", baseCmd, conf),
	}

	klog.Infof("reconfiguring cluster from %s", conf)
	// Run commands one at a time so that it is easier to root cause failures.
	for _, c := range cmds {
		if _, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", c)); err != nil {
			klog.Errorf("%s failed - will try once more: %v", c, err)

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

	// because reboots clear /etc/cni
	if err := k.applyCNI(cfg); err != nil {
		return errors.Wrap(err, "apply cni")
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
		klog.Warningf("addon install failed, wil retry: %v", err)
		return errors.Wrap(err, "addons")
	}

	if err := bsutil.AdjustResourceLimits(k.c); err != nil {
		klog.Warningf("unable to adjust resource limits: %v", err)
	}
	return nil
}

// JoinCluster adds a node to an existing cluster
func (k *Bootstrapper) JoinCluster(cc config.ClusterConfig, n config.Node, joinCmd string) error {
	start := time.Now()
	klog.Infof("JoinCluster: %+v", cc)
	defer func() {
		klog.Infof("JoinCluster complete in %s", time.Since(start))
	}()

	// Join the master by specifying its token
	joinCmd = fmt.Sprintf("%s --node-name=%s", joinCmd, driver.MachineName(cc, n))

	join := func() error {
		// reset first to clear any possibly existing state
		_, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s reset -f", bsutil.InvokeKubeadm(cc.KubernetesConfig.KubernetesVersion))))
		if err != nil {
			klog.Infof("kubeadm reset failed, continuing anyway: %v", err)
		}

		out, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", joinCmd))
		if err != nil {
			return errors.Wrapf(err, "cmd failed: %s\n%+v\n", joinCmd, out.Output())
		}
		return nil
	}

	if err := retry.Expo(join, 10*time.Second, 3*time.Minute); err != nil {
		return errors.Wrap(err, "joining cp")
	}

	if _, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", "sudo systemctl daemon-reload && sudo systemctl enable kubelet && sudo systemctl start kubelet")); err != nil {
		return errors.Wrap(err, "starting kubelet")
	}

	return nil
}

// GenerateToken creates a token and returns the appropriate kubeadm join command to run, or the already existing token
func (k *Bootstrapper) GenerateToken(cc config.ClusterConfig) (string, error) {
	// Take that generated token and use it to get a kubeadm join command
	tokenCmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s token create --print-join-command --ttl=0", bsutil.InvokeKubeadm(cc.KubernetesConfig.KubernetesVersion)))
	r, err := k.c.RunCmd(tokenCmd)
	if err != nil {
		return "", errors.Wrap(err, "generating join command")
	}

	joinCmd := r.Stdout.String()
	joinCmd = strings.Replace(joinCmd, "kubeadm", bsutil.InvokeKubeadm(cc.KubernetesConfig.KubernetesVersion), 1)
	joinCmd = fmt.Sprintf("%s --ignore-preflight-errors=all", strings.TrimSpace(joinCmd))
	if cc.KubernetesConfig.CRISocket != "" {
		joinCmd = fmt.Sprintf("%s --cri-socket %s", joinCmd, cc.KubernetesConfig.CRISocket)
	}

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
		klog.Warningf("%s: %v", rr.Command(), err)
	}

	if err := sysinit.New(k.c).ForceStop("kubelet"); err != nil {
		klog.Warningf("stop kubelet: %v", err)
	}

	containers, err := cr.ListContainers(cruntime.ListOptions{Namespaces: []string{"kube-system"}})
	if err != nil {
		klog.Warningf("unable to list kube-system containers: %v", err)
	}
	if len(containers) > 0 {
		klog.Warningf("found %d kube-system containers to stop", len(containers))
		if err := cr.StopContainers(containers); err != nil {
			klog.Warningf("error stopping containers: %v", err)
		}
	}

	return derr
}

// SetupCerts sets up certificates within the cluster.
func (k *Bootstrapper) SetupCerts(k8s config.KubernetesConfig, n config.Node) error {
	_, err := bootstrapper.SetupCerts(k.c, k8s, n)
	return err
}

// UpdateCluster updates the control plane with cluster-level info.
func (k *Bootstrapper) UpdateCluster(cfg config.ClusterConfig) error {
	images, err := images.Kubeadm(cfg.KubernetesConfig.ImageRepository, cfg.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "kubeadm images")
	}

	r, err := cruntime.New(cruntime.Config{
		Type:   cfg.KubernetesConfig.ContainerRuntime,
		Runner: k.c, Socket: cfg.KubernetesConfig.CRISocket,
	})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	if err := r.Preload(cfg.KubernetesConfig); err != nil {
		klog.Infof("preload failed, will try to load cached images: %v", err)
	}

	if cfg.KubernetesConfig.ShouldLoadCachedImages {
		if err := machine.LoadImages(&cfg, k.c, images, constants.ImageCacheDir); err != nil {
			out.FailureT("Unable to load cached images: {{.error}}", out.V{"error": err})
		}
	}

	cp, err := config.PrimaryControlPlane(&cfg)
	if err != nil {
		return errors.Wrap(err, "getting control plane")
	}

	err = k.UpdateNode(cfg, cp, r)
	if err != nil {
		return errors.Wrap(err, "updating control plane")
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

	klog.Infof("kubelet %s config:\n%+v", kubeletCfg, cfg.KubernetesConfig)

	sm := sysinit.New(k.c)

	if err := bsutil.TransferBinaries(cfg.KubernetesConfig, k.c, sm); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}

	files := []assets.CopyableFile{
		assets.NewMemoryAssetTarget(kubeletCfg, bsutil.KubeletSystemdConfFile, "0644"),
		assets.NewMemoryAssetTarget(kubeletService, bsutil.KubeletServiceFile, "0644"),
	}

	if n.ControlPlane {
		files = append(files, assets.NewMemoryAssetTarget(kubeadmCfg, bsutil.KubeadmYamlPath+".new", "0640"))
	}

	// Installs compatibility shims for non-systemd environments
	kubeletPath := path.Join(vmpath.GuestPersistentDir, "binaries", cfg.KubernetesConfig.KubernetesVersion, "kubelet")
	shims, err := sm.GenerateInitShim("kubelet", kubeletPath, bsutil.KubeletSystemdConfFile)
	if err != nil {
		return errors.Wrap(err, "shim")
	}
	files = append(files, shims...)

	if err := bsutil.CopyFiles(k.c, files); err != nil {
		return errors.Wrap(err, "copy")
	}

	cp, err := config.PrimaryControlPlane(&cfg)
	if err != nil {
		return errors.Wrap(err, "control plane")
	}

	if err := machine.AddHostAlias(k.c, constants.ControlPlaneAlias, net.ParseIP(cp.IP)); err != nil {
		return errors.Wrap(err, "host alias")
	}

	return nil
}

// kubectlPath returns the path to the kubelet
func kubectlPath(cfg config.ClusterConfig) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", cfg.KubernetesConfig.KubernetesVersion, "kubectl")
}

// applyNodeLabels applies minikube labels to all the nodes
func (k *Bootstrapper) applyNodeLabels(cfg config.ClusterConfig) error {
	// time cluster was created. time format is based on ISO 8601 (RFC 3339)
	// converting - and : to _ because of Kubernetes label restriction
	createdAtLbl := "minikube.k8s.io/updated_at=" + time.Now().Format("2006_01_02T15_04_05_0700")
	verLbl := "minikube.k8s.io/version=" + version.GetVersion()
	commitLbl := "minikube.k8s.io/commit=" + version.GetGitCommitID()
	nameLbl := "minikube.k8s.io/name=" + cfg.Name

	ctx, cancel := context.WithTimeout(context.Background(), applyTimeoutSeconds*time.Second)
	defer cancel()
	// example:
	// sudo /var/lib/minikube/binaries/<version>/kubectl label nodes minikube.k8s.io/version=<version> minikube.k8s.io/commit=aa91f39ffbcf27dcbb93c4ff3f457c54e585cf4a-dirty minikube.k8s.io/name=p1 minikube.k8s.io/updated_at=2020_02_20T12_05_35_0700 --all --overwrite --kubeconfig=/var/lib/minikube/kubeconfig
	cmd := exec.CommandContext(ctx, "sudo", kubectlPath(cfg),
		"label", "nodes", verLbl, commitLbl, nameLbl, createdAtLbl, "--all", "--overwrite",
		fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")))

	if _, err := k.c.RunCmd(cmd); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return errors.Wrapf(err, "timeout apply labels")
		}
		return errors.Wrapf(err, "applying node labels")
	}
	return nil
}

// elevateKubeSystemPrivileges gives the kube-system service account cluster admin privileges to work with RBAC.
func (k *Bootstrapper) elevateKubeSystemPrivileges(cfg config.ClusterConfig) error {
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s to wait for elevateKubeSystemPrivileges.", time.Since(start))
	}()

	// Allow no more than 5 seconds for creating cluster role bindings
	ctx, cancel := context.WithTimeout(context.Background(), applyTimeoutSeconds*time.Second)
	defer cancel()
	rbacName := "minikube-rbac"
	// kubectl create clusterrolebinding minikube-rbac --clusterrole=cluster-admin --serviceaccount=kube-system:default
	cmd := exec.CommandContext(ctx, "sudo", kubectlPath(cfg),
		"create", "clusterrolebinding", rbacName, "--clusterrole=cluster-admin", "--serviceaccount=kube-system:default",
		fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")))
	rr, err := k.c.RunCmd(cmd)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return errors.Wrapf(err, "timeout apply sa")
		}
		// Error from server (AlreadyExists): clusterrolebindings.rbac.authorization.k8s.io "minikube-rbac" already exists
		if strings.Contains(rr.Output(), "Error from server (AlreadyExists)") {
			klog.Infof("rbac %q already exists not need to re-create.", rbacName)
		} else {
			return errors.Wrapf(err, "apply sa")
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

// stopKubeSystem stops all the containers in the kube-system to prevent #8740 when doing hot upgrade
func (k *Bootstrapper) stopKubeSystem(cfg config.ClusterConfig) error {
	klog.Info("stopping kube-system containers ...")
	cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: k.c})
	if err != nil {
		return errors.Wrap(err, "new cruntime")
	}

	ids, err := cr.ListContainers(cruntime.ListOptions{Namespaces: []string{"kube-system"}})
	if err != nil {
		return errors.Wrap(err, "list")
	}

	if len(ids) > 0 {
		if err := cr.StopContainers(ids); err != nil {
			return errors.Wrap(err, "stop")
		}
	}
	return nil
}

// adviseNodePressure will advise the user what to do with difference pressure errors based on their environment
func adviseNodePressure(err error, name string, drv string) {
	if diskErr, ok := err.(*kverify.ErrDiskPressure); ok {
		out.ErrLn("")
		klog.Warning(diskErr)
		out.WarningT("The node {{.name}} has ran out of disk space.", out.V{"name": name})
		// generic advice for all drivers
		out.T(style.Tip, "Please free up disk or prune images.")
		if driver.IsVM(drv) {
			out.T(style.Stopped, "Please create a cluster with bigger disk size: `minikube start --disk SIZE_MB` ")
		} else if drv == oci.Docker && runtime.GOOS != "linux" {
			out.T(style.Stopped, "Please increse Desktop's disk size.")
			if runtime.GOOS == "darwin" {
				out.T(style.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
			}
			if runtime.GOOS == "windows" {
				out.T(style.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
			}
		}
		out.ErrLn("")
		return
	}

	if memErr, ok := err.(*kverify.ErrMemoryPressure); ok {
		out.ErrLn("")
		klog.Warning(memErr)
		out.WarningT("The node {{.name}} has ran out of memory.", out.V{"name": name})
		out.T(style.Tip, "Check if you have unnecessary pods running by running 'kubectl get po -A")
		if driver.IsVM(drv) {
			out.T(style.Stopped, "Consider creating a cluster with larger memory size using `minikube start --memory SIZE_MB` ")
		} else if drv == oci.Docker && runtime.GOOS != "linux" {
			out.T(style.Stopped, "Consider increasing Docker Desktop's memory size.")
			if runtime.GOOS == "darwin" {
				out.T(style.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
			}
			if runtime.GOOS == "windows" {
				out.T(style.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
			}
		}
		out.ErrLn("")
		return
	}

	if pidErr, ok := err.(*kverify.ErrPIDPressure); ok {
		klog.Warning(pidErr)
		out.ErrLn("")
		out.WarningT("The node {{.name}} has ran out of available PIDs.", out.V{"name": name})
		out.ErrLn("")
		return
	}

	if netErr, ok := err.(*kverify.ErrNetworkNotReady); ok {
		klog.Warning(netErr)
		out.ErrLn("")
		out.WarningT("The node {{.name}} network is not available. Please verify network settings.", out.V{"name": name})
		out.ErrLn("")
		return
	}
}
