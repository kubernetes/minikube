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
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	// WARNING: Do not use path/filepath in this package unless you want bizarre Windows paths

	"github.com/blang/semver/v4"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	kubevip "k8s.io/minikube/pkg/minikube/cluster/ha/kube-vip"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/network"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
	"k8s.io/minikube/pkg/version"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

// Bootstrapper is a bootstrapper using kubeadm
type Bootstrapper struct {
	c           command.Runner
	k8sClient   *kubernetes.Clientset // Kubernetes client used to verify pods inside cluster
	contextName string
}

// NewBootstrapper creates a new kubeadm.Bootstrapper
func NewBootstrapper(_ libmachine.API, cc config.ClusterConfig, r command.Runner) (*Bootstrapper, error) {
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

// clearStaleConfigs tries to clear configurations which may have stale IP addresses.
func (k *Bootstrapper) clearStaleConfigs(cfg config.ClusterConfig) {
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
	}
	klog.Infof("found existing configuration files:\n%s\n", rr.Stdout.String())

	endpoint := fmt.Sprintf("https://%s", net.JoinHostPort(constants.ControlPlaneAlias, strconv.Itoa(cfg.APIServerPort)))
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
}

// init initialises primary control-plane using kubeadm.
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
		fmt.Sprintf("DirAvailable-%s", strings.ReplaceAll(vmpath.GuestManifestsDir, "/", "-")),
		fmt.Sprintf("DirAvailable-%s", strings.ReplaceAll(vmpath.GuestPersistentDir, "/", "-")),
		fmt.Sprintf("DirAvailable-%s", strings.ReplaceAll(bsutil.EtcdDataDir(), "/", "-")),
		"FileAvailable--etc-kubernetes-manifests-kube-scheduler.yaml",
		"FileAvailable--etc-kubernetes-manifests-kube-apiserver.yaml",
		"FileAvailable--etc-kubernetes-manifests-kube-controller-manager.yaml",
		"FileAvailable--etc-kubernetes-manifests-etcd.yaml",
		"Port-10250", // For "none" users who already have a kubelet online
		"Swap",       // For "none" users who have swap configured
		"NumCPU",     // For "none" users who have too few CPUs
	}
	if version.GE(semver.MustParse("1.20.0")) {
		ignore = append(ignore,
			"Mem", // For "none" users who have too little memory
		)
	}
	ignore = append(ignore, bsutil.SkipAdditionalPreflights[r.Name()]...)

	skipSystemVerification := false
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

	k.clearStaleConfigs(cfg)

	conf := constants.KubeadmYamlPath
	ctx, cancel := context.WithTimeout(context.Background(), initTimeoutMinutes*time.Minute)
	defer cancel()
	kr, kw := io.Pipe()
	c := exec.CommandContext(ctx, "/bin/bash", "-c", fmt.Sprintf("%s init --config %s %s --ignore-preflight-errors=%s",
		bsutil.InvokeKubeadm(cfg.KubernetesConfig.KubernetesVersion), conf, extraFlags, strings.Join(ignore, ",")))
	c.Stdout = kw
	c.Stderr = kw
	var wg sync.WaitGroup
	wg.Add(1)
	sc, err := k.c.StartCmd(c)
	if err != nil {
		return errors.Wrap(err, "start")
	}
	go outputKubeadmInitSteps(kr, &wg)
	if _, err := k.c.WaitCmd(sc); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrInitTimedout
		}

		if strings.Contains(err.Error(), "'kubeadm': Permission denied") {
			return ErrNoExecLinux
		}
		return errors.Wrap(err, "wait")
	}
	kw.Close()
	wg.Wait()

	if err := k.applyCNI(cfg, true); err != nil {
		return errors.Wrap(err, "apply cni")
	}

	wg.Add(3)

	go func() {
		defer wg.Done()
		// we need to have cluster role binding before applying overlay to avoid #7428
		if err := k.elevateKubeSystemPrivileges(cfg); err != nil {
			klog.Errorf("unable to create cluster role binding for primary control-plane node, some addons might not work: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := k.LabelAndUntaintNode(cfg, config.ControlPlanes(cfg)[0]); err != nil {
			klog.Warningf("unable to apply primary control-plane node labels and taints: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := bsutil.AdjustResourceLimits(k.c); err != nil {
			klog.Warningf("unable to adjust resource limits for primary control-plane node: %v", err)
		}
	}()

	wg.Wait()

	// tunnel apiserver to guest
	if err := k.tunnelToAPIServer(cfg); err != nil {
		klog.Warningf("apiserver tunnel failed: %v", err)
	}

	return nil
}

// outputKubeadmInitSteps streams the pipe and outputs the current step
func outputKubeadmInitSteps(logs io.Reader, wg *sync.WaitGroup) {
	type step struct {
		logTag       string
		registerStep register.RegStep
	}

	steps := []step{
		{logTag: "certs", registerStep: register.PreparingKubernetesCerts},
		{logTag: "control-plane", registerStep: register.PreparingKubernetesControlPlane},
		{logTag: "bootstrap-token", registerStep: register.PreparingKubernetesBootstrapToken},
	}
	nextStepIndex := 0

	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		line := scanner.Text()
		klog.Info(line)
		if nextStepIndex >= len(steps) {
			continue
		}
		nextStep := steps[nextStepIndex]
		if !strings.Contains(line, fmt.Sprintf("[%s]", nextStep.logTag)) {
			continue
		}
		register.Reg.SetStep(nextStep.registerStep)
		// because the translation extract (make extract) needs simple strings to be included in translations we have to pass simple strings
		if nextStepIndex == 0 {
			out.Step(style.SubStep, "Generating certificates and keys ...")
		}
		if nextStepIndex == 1 {
			out.Step(style.SubStep, "Booting up control plane ...")
		}
		if nextStepIndex == 2 {
			out.Step(style.SubStep, "Configuring RBAC rules ...")
		}

		nextStepIndex++
	}
	if err := scanner.Err(); err != nil {
		klog.Warningf("failed to read logs: %v", err)
	}
	wg.Done()
}

// applyCNI applies CNI to a cluster. Needs to be done every time a VM is powered up.
func (k *Bootstrapper) applyCNI(cfg config.ClusterConfig, registerStep ...bool) error {
	regStep := false
	if len(registerStep) > 0 {
		regStep = registerStep[0]
	}

	cnm, err := cni.New(&cfg)
	if err != nil {
		return errors.Wrap(err, "cni config")
	}

	if _, ok := cnm.(cni.Disabled); ok {
		return nil
	}

	// when not on init, can run in parallel and break step output order
	if regStep {
		register.Reg.SetStep(register.ConfiguringCNI)
		out.Step(style.CNI, "Configuring {{.name}} (Container Networking Interface) ...", out.V{"name": cnm.String()})
	} else {
		out.Styled(style.CNI, "Configuring {{.name}} (Container Networking Interface) ...", out.V{"name": cnm.String()})
	}

	if err := cnm.Apply(k.c); err != nil {
		return errors.Wrap(err, "cni apply")
	}

	return nil
}

// unpause unpauses any Kubernetes backplane components
func (k *Bootstrapper) unpause(cfg config.ClusterConfig) error {
	cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: k.c})
	if err != nil {
		return err
	}

	ids, err := cr.ListContainers(cruntime.ListContainersOptions{State: cruntime.Paused, Namespaces: []string{"kube-system"}})
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
		klog.Infof("duration metric: took %s to StartCluster", time.Since(start))
	}()

	// Before we start, ensure that no paused components are lurking around
	if err := k.unpause(cfg); err != nil {
		klog.Warningf("unpause failed: %v", err)
	}

	if err := bsutil.ExistingConfig(k.c); err == nil {
		// if the guest already exists and was stopped, re-establish the apiserver tunnel so checks pass
		if err := k.tunnelToAPIServer(cfg); err != nil {
			klog.Warningf("apiserver tunnel failed: %v", err)
		}

		klog.Infof("found existing configuration files, will attempt cluster restart")

		var rerr error
		if rerr := k.restartPrimaryControlPlane(cfg); rerr == nil {
			return nil
		}
		out.ErrT(style.Embarrassed, "Unable to restart control-plane node(s), will reset cluster: {{.error}}", out.V{"error": rerr})
		if err := k.DeleteCluster(cfg.KubernetesConfig); err != nil {
			klog.Warningf("delete failed: %v", err)
		}
		// Fall-through to init
	}

	conf := constants.KubeadmYamlPath
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

// tunnelToAPIServer creates ssh tunnel between apiserver:port inside control-plane node and host on port 8443.
func (k *Bootstrapper) tunnelToAPIServer(cfg config.ClusterConfig) error {
	if cfg.APIServerPort == 0 {
		return fmt.Errorf("apiserver port not set")
	}
	// An API server tunnel is only needed for QEMU w/ builtin network, for
	// everything else return
	if !driver.IsQEMU(cfg.Driver) || !network.IsBuiltinQEMU(cfg.Network) {
		return nil
	}

	m, err := machine.NewAPIClient()
	if err != nil {
		return errors.Wrapf(err, "create libmachine api client")
	}

	cp, err := config.ControlPlane(cfg)
	if err != nil {
		return errors.Wrapf(err, "get control-plane node")
	}

	args := []string{"-f", "-NTL", fmt.Sprintf("%d:localhost:8443", cfg.APIServerPort)}
	if err = machine.CreateSSHShell(m, cfg, cp, args, false); err != nil {
		return errors.Wrapf(err, "ssh command")
	}
	return nil
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
	out.Step(style.HealthCheck, "Verifying Kubernetes components...")
	// regardless if waiting is set or not, we will make sure kubelet is not stopped
	// to solve corner cases when a container is hibernated and once coming back kubelet not running.
	if err := sysinit.New(k.c).Start("kubelet"); err != nil {
		klog.Warningf("Couldn't ensure kubelet is started this might cause issues: %v", err)
	}
	// TODO: #7706: for better performance we could use k.client inside minikube to avoid asking for external IP:PORT
	cp, err := config.ControlPlane(cfg)
	if err != nil {
		return errors.Wrap(err, "get control-plane node")
	}
	hostname, _, port, err := driver.ControlPlaneEndpoint(&cfg, &cp, cfg.Driver)
	if err != nil {
		return errors.Wrap(err, "get control-plane endpoint")
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

	if cfg.VerifyComponents[kverify.NodeReadyKey] {
		name := bsutil.KubeNodeName(cfg, n)
		if err := kverify.WaitNodeCondition(client, name, core.NodeReady, timeout); err != nil {
			return errors.Wrap(err, "waiting for node to be ready")
		}
	}

	if cfg.VerifyComponents[kverify.ExtraKey] {
		if err := kverify.WaitExtra(client, kverify.CorePodsLabels, timeout); err != nil {
			return errors.Wrap(err, "extra waiting")
		}
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

	klog.Infof("duration metric: took %s to wait for: %+v", time.Since(start), cfg.VerifyComponents)

	if err := kverify.NodePressure(client); err != nil {
		adviseNodePressure(err, cfg.Name, cfg.Driver)
		return errors.Wrap(err, "node pressure")
	}
	return nil
}

// restartPrimaryControlPlane restarts the kubernetes cluster configured by kubeadm.
func (k *Bootstrapper) restartPrimaryControlPlane(cfg config.ClusterConfig) error {
	klog.Infof("restartPrimaryControlPlane start ...")

	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s to restartPrimaryControlPlane", time.Since(start))
	}()

	if err := k.createCompatSymlinks(); err != nil {
		klog.Errorf("failed to create compat symlinks: %v", err)
	}

	pcp, err := config.ControlPlane(cfg)
	if err != nil || !config.IsPrimaryControlPlane(cfg, pcp) {
		return errors.Wrap(err, "get primary control-plane node")
	}

	host, _, port, err := driver.ControlPlaneEndpoint(&cfg, &pcp, cfg.Driver)
	if err != nil {
		return errors.Wrap(err, "get primary control-plane endpoint")
	}

	// Save the costly tax of reinstalling Kubernetes if the only issue is a missing kube context
	if _, err := kubeconfig.UpdateEndpoint(cfg.Name, host, port, kubeconfig.PathFromEnv(), kubeconfig.NewExtension()); err != nil {
		klog.Warningf("unable to update kubeconfig (cluster will likely require a reset): %v", err)
	}

	client, err := k.client(host, port)
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}

	// If the cluster is running, check if we have any work to do.
	conf := constants.KubeadmYamlPath

	// check whether or not the cluster needs to be reconfigured
	if rr, err := k.c.RunCmd(exec.Command("sudo", "diff", "-u", conf, conf+".new")); err == nil {
		// DANGER: This log message is hard-coded in an integration test!
		klog.Infof("The running cluster does not require reconfiguration: %s", host)
		// taking a shortcut, as the cluster seems to be properly configured
		// except for vm driver in non-ha (non-multi-control plane) cluster - fallback to old behaviour
		// here we're making a tradeoff to avoid significant (10sec) waiting on restarting stopped non-ha (non-multi-control plane) cluster with vm driver
		// where such cluster needs to be reconfigured b/c of (currently) ephemeral config, but then also,
		// starting already started such cluster (hard to know w/o investing that time) will fallthrough the same path and reconfigure cluster
		if config.IsHA(cfg) || !driver.IsVM(cfg.Driver) {
			return nil
		}
	} else {
		klog.Infof("detected kubeadm config drift (will reconfigure cluster from new %s):\n%s", conf, rr.Output())
	}

	if err := k.stopKubeSystem(cfg); err != nil {
		klog.Warningf("Failed to stop kube-system containers, port conflicts may arise: %v", err)
	}

	if err := sysinit.New(k.c).Stop("kubelet"); err != nil {
		klog.Warningf("Failed to stop kubelet, this might cause upgrade errors: %v", err)
	}

	k.clearStaleConfigs(cfg)

	if _, err := k.c.RunCmd(exec.Command("sudo", "cp", conf+".new", conf)); err != nil {
		return errors.Wrap(err, "cp")
	}

	baseCmd := fmt.Sprintf("%s init", bsutil.InvokeKubeadm(cfg.KubernetesConfig.KubernetesVersion))
	cmds := []string{
		fmt.Sprintf("%s phase certs all --config %s", baseCmd, conf),
		fmt.Sprintf("%s phase kubeconfig all --config %s", baseCmd, conf),
		fmt.Sprintf("%s phase kubelet-start --config %s", baseCmd, conf),
		fmt.Sprintf("%s phase control-plane all --config %s", baseCmd, conf),
		fmt.Sprintf("%s phase etcd local --config %s", baseCmd, conf),
	}

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

	if err := kverify.WaitForHealthyAPIServer(cr, k, cfg, k.c, client, time.Now(), host, port, kconst.DefaultControlPlaneTimeout); err != nil {
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
		addons := "all"
		if cfg.KubernetesConfig.ExtraOptions.Exists("kubeadm.skip-phases=addon/kube-proxy") {
			addons = "coredns"
		}
		_, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s phase addon %s --config %s", baseCmd, addons, conf)))
		return err
	}
	if err = retry.Expo(addonPhase, 100*time.Microsecond, 30*time.Second); err != nil {
		klog.Warningf("addon install failed, wil retry: %v", err)
		return errors.Wrap(err, "addons")
	}

	// must be called after applyCNI and `kubeadm phase addon all` (ie, coredns redeploy)
	if cfg.VerifyComponents[kverify.ExtraKey] {
		// after kubelet is restarted (with 'kubeadm init phase kubelet-start' above),
		// it appears as to be immediately Ready as well as all kube-system pods (last observed state),
		// then (after ~10sec) it realises it has some changes to apply, implying also pods restarts,
		// and by that time we would exit completely, so we wait until kubelet begins restarting pods
		klog.Info("waiting for restarted kubelet to initialise ...")
		start := time.Now()
		wait := func() error {
			pods, err := client.CoreV1().Pods(meta.NamespaceSystem).List(context.Background(), meta.ListOptions{LabelSelector: "tier=control-plane"})
			if err != nil {
				return err
			}
			for _, pod := range pods.Items {
				if ready, _ := kverify.IsPodReady(&pod); !ready {
					return nil
				}
			}
			return fmt.Errorf("kubelet not initialised")
		}
		_ = retry.Expo(wait, 250*time.Millisecond, 1*time.Minute)
		klog.Infof("kubelet initialised")
		klog.Infof("duration metric: took %s waiting for restarted kubelet to initialise ...", time.Since(start))

		if err := kverify.WaitExtra(client, kverify.CorePodsLabels, kconst.DefaultControlPlaneTimeout); err != nil {
			return errors.Wrap(err, "extra")
		}
	}

	if err := bsutil.AdjustResourceLimits(k.c); err != nil {
		klog.Warningf("unable to adjust resource limits: %v", err)
	}

	return nil
}

// JoinCluster adds new node to an existing cluster.
func (k *Bootstrapper) JoinCluster(cc config.ClusterConfig, n config.Node, joinCmd string) error {
	// Join the control plane by specifying its token
	joinCmd = fmt.Sprintf("%s --node-name=%s", joinCmd, config.MachineName(cc, n))

	if n.ControlPlane {
		joinCmd += " --control-plane"
		// fix kvm driver where ip address is automatically taken from the "default" network instead from the dedicated network
		// avoid error: "error execution phase control-plane-prepare/certs: error creating PKI assets: failed to write or validate certificate "apiserver": certificate apiserver is invalid: x509: certificate is valid for 192.168.39.147, 10.96.0.1, 127.0.0.1, 10.0.0.1, 192.168.39.58, not 192.168.122.21"
		// ref: https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-join/#options
		// "If the node should host a new control plane instance, the IP address the API Server will advertise it's listening on. If not set the default network interface will be used."
		// "If the node should host a new control plane instance, the port for the API Server to bind to."
		joinCmd += " --apiserver-advertise-address=" + n.IP +
			" --apiserver-bind-port=" + strconv.Itoa(n.Port)
	}

	if _, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", joinCmd)); err != nil {
		return errors.Wrapf(err, "kubeadm join")
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

	// avoid "Found multiple CRI endpoints on the host. Please define which one do you wish to use by setting the 'criSocket' field in the kubeadm configuration file: unix:///var/run/containerd/containerd.sock, unix:///var/run/cri-dockerd.sock" error
	version, err := util.ParseKubernetesVersion(cc.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return "", errors.Wrap(err, "parsing Kubernetes version")
	}
	cr, err := cruntime.New(cruntime.Config{Type: cc.KubernetesConfig.ContainerRuntime, Runner: k.c, Socket: cc.KubernetesConfig.CRISocket, KubernetesVersion: version})
	if err != nil {
		klog.Errorf("cruntime: %v", err)
	}

	sp := cr.SocketPath()
	// avoid warning/error:
	// 'Usage of CRI endpoints without URL scheme is deprecated and can cause kubelet errors in the future.
	//  Automatically prepending scheme "unix" to the "criSocket" with value "/var/run/cri-dockerd.sock".
	//  Please update your configuration!'
	if !strings.HasPrefix(sp, "unix://") {
		sp = "unix://" + sp
	}
	joinCmd = fmt.Sprintf("%s --cri-socket %s", joinCmd, sp)

	return joinCmd, nil
}

// StopKubernetes attempts to stop existing kubernetes.
func StopKubernetes(runner command.Runner, cr cruntime.Manager) {
	// Verify that Kubernetes is still running.
	stk := kverify.ServiceStatus(runner, "kubelet")
	if stk.String() != "Running" {
		return
	}

	out.Infof("Kubernetes: Stopping ...")

	// Force stop "Kubelet".
	if err := sysinit.New(runner).ForceStop("kubelet"); err != nil {
		klog.Warningf("stop kubelet: %v", err)
	}

	// Stop each Kubernetes container.
	containers, err := cr.ListContainers(cruntime.ListContainersOptions{Namespaces: []string{"kube-system"}})
	if err != nil {
		klog.Warningf("unable to list kube-system containers: %v", err)
	}
	if len(containers) > 0 {
		klog.Warningf("found %d kube-system containers to stop", len(containers))
		if err := cr.StopContainers(containers); err != nil {
			klog.Warningf("error stopping containers: %v", err)
		}
	}

	// Verify that Kubernetes has stopped.
	stk = kverify.ServiceStatus(runner, "kubelet")
	out.Infof("Kubernetes: {{.status}}", out.V{"status": stk.String()})
}

// DeleteCluster removes the components that were started earlier
func (k *Bootstrapper) DeleteCluster(k8s config.KubernetesConfig) error {
	version, err := util.ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kubernetes version")
	}
	cr, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: k.c, Socket: k8s.CRISocket, KubernetesVersion: version})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	ka := bsutil.InvokeKubeadm(k8s.KubernetesVersion)
	sp := cr.SocketPath()
	cmd := fmt.Sprintf("%s reset --cri-socket %s --force", ka, sp)
	if version.LT(semver.MustParse("1.11.0")) {
		cmd = fmt.Sprintf("%s reset --cri-socket %s", ka, sp)
	}

	rr, derr := k.c.RunCmd(exec.Command("/bin/bash", "-c", cmd))
	if derr != nil {
		klog.Warningf("%s: %v", rr.Command(), err)
	}

	StopKubernetes(k.c, cr)
	return derr
}

// SetupCerts sets up certificates within the cluster.
func (k *Bootstrapper) SetupCerts(k8s config.ClusterConfig, n config.Node, pcpCmd cruntime.CommandRunner) error {
	return bootstrapper.SetupCerts(k8s, n, pcpCmd, k.c)
}

// UpdateCluster updates the control plane with cluster-level info.
func (k *Bootstrapper) UpdateCluster(cfg config.ClusterConfig) error {
	klog.Infof("updating cluster %+v ...", cfg)

	images, err := images.Kubeadm(cfg.KubernetesConfig.ImageRepository, cfg.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "kubeadm images")
	}

	version, err := util.ParseKubernetesVersion(cfg.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kubernetes version")
	}
	r, err := cruntime.New(cruntime.Config{
		Type:              cfg.KubernetesConfig.ContainerRuntime,
		Runner:            k.c,
		Socket:            cfg.KubernetesConfig.CRISocket,
		KubernetesVersion: version,
	})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	if err := r.Preload(cfg); err != nil {
		switch err.(type) {
		case *cruntime.ErrISOFeature:
			out.ErrT(style.Tip, "Existing disk is missing new features ({{.error}}). To upgrade, run 'minikube delete'", out.V{"error": err})
		default:
			klog.Infof("preload failed, will try to load cached images: %v", err)
		}
	}

	if cfg.KubernetesConfig.ShouldLoadCachedImages {
		if err := machine.LoadCachedImages(&cfg, k.c, images, detect.ImageCacheDir(), false); err != nil {
			out.FailureT("Unable to load cached images: {{.error}}", out.V{"error": err})
		}
	}

	pcp, err := config.ControlPlane(cfg)
	if err != nil || !config.IsPrimaryControlPlane(cfg, pcp) {
		return errors.Wrap(err, "get primary control-plane node")
	}

	err = k.UpdateNode(cfg, pcp, r)
	if err != nil {
		return errors.Wrap(err, "update primary control-plane node")
	}

	return nil
}

// UpdateNode updates new or existing node.
func (k *Bootstrapper) UpdateNode(cfg config.ClusterConfig, n config.Node, r cruntime.Manager) error {
	klog.Infof("updating node %v ...", n)

	kubeletCfg, err := bsutil.NewKubeletConfig(cfg, n, r)
	if err != nil {
		return errors.Wrap(err, "generating kubelet config")
	}

	kubeletService, err := bsutil.NewKubeletService(cfg.KubernetesConfig)
	if err != nil {
		return errors.Wrap(err, "generating kubelet service")
	}

	klog.Infof("kubelet %s config:\n%+v", kubeletCfg, cfg.KubernetesConfig)

	files := []assets.CopyableFile{
		assets.NewMemoryAssetTarget(kubeletCfg, bsutil.KubeletSystemdConfFile, "0644"),
		assets.NewMemoryAssetTarget(kubeletService, bsutil.KubeletServiceFile, "0644"),
	}

	if n.ControlPlane {
		// for primary control-plane node only, generate kubeadm config based on current params
		// on node restart, it will be checked against later if anything needs changing
		var kubeadmCfg []byte
		if config.IsPrimaryControlPlane(cfg, n) {
			kubeadmCfg, err = bsutil.GenerateKubeadmYAML(cfg, n, r)
			if err != nil {
				return errors.Wrap(err, "generating kubeadm cfg")
			}
			files = append(files, assets.NewMemoryAssetTarget(kubeadmCfg, constants.KubeadmYamlPath+".new", "0640"))
		}
		// deploy kube-vip for ha (multi-control plane) cluster
		if config.IsHA(cfg) {
			// workaround for kube-vip
			// only applicable for k8s v1.29+ during primary control-plane node's kubeadm init (ie, first boot)
			// TODO (prezha): remove when fixed upstream - ref: https://github.com/kube-vip/kube-vip/issues/684#issuecomment-1864855405
			kv, err := semver.ParseTolerant(cfg.KubernetesConfig.KubernetesVersion)
			if err != nil {
				return errors.Wrapf(err, "parsing kubernetes version %q", cfg.KubernetesConfig.KubernetesVersion)
			}
			workaround := kv.GTE(semver.Version{Major: 1, Minor: 29}) && config.IsPrimaryControlPlane(cfg, n) && len(config.ControlPlanes(cfg)) == 1
			kubevipCfg, err := kubevip.Configure(cfg, k.c, kubeadmCfg, workaround)
			if err != nil {
				klog.Errorf("couldn't generate kube-vip config, this might cause issues (will continue): %v", err)
			} else {
				files = append(files, assets.NewMemoryAssetTarget(kubevipCfg, path.Join(vmpath.GuestManifestsDir, kubevip.Manifest), "0600"))
			}
		}
	}

	sm := sysinit.New(k.c)

	if err := bsutil.TransferBinaries(cfg.KubernetesConfig, k.c, sm, cfg.BinaryMirror); err != nil {
		return errors.Wrap(err, "downloading binaries")
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

	if err := k.copyResolvConf(cfg); err != nil {
		return errors.Wrap(err, "resolv.conf")
	}

	// add "control-plane.minikube.internal" dns alias
	// note: needs to be called after APIServerHAVIP is set (in startPrimaryControlPlane()) and before kubeadm kicks off
	cpIP := cfg.KubernetesConfig.APIServerHAVIP
	if !config.IsHA(cfg) {
		cp, err := config.ControlPlane(cfg)
		if err != nil {
			return errors.Wrap(err, "get control-plane node")
		}
		cpIP = cp.IP
	}
	if err := machine.AddHostAlias(k.c, constants.ControlPlaneAlias, net.ParseIP(cpIP)); err != nil {
		return errors.Wrap(err, "add control-plane alias")
	}

	// "ensure" kubelet is started, intentionally non-fatal in case of an error
	if err := sysinit.New(k.c).Start("kubelet"); err != nil {
		klog.Errorf("Couldn't ensure kubelet is started this might cause issues (will continue): %v", err)
	}

	return nil
}

// copyResolvConf is a workaround for a regression introduced with https://github.com/kubernetes/kubernetes/pull/109441
// The regression is resolved by making a copy of /etc/resolv.conf, removing the line "search ." from the copy, and setting kubelet to use the copy
// Only Kubernetes v1.25.0 is affected by this regression
func (k *Bootstrapper) copyResolvConf(cfg config.ClusterConfig) error {
	if !bsutil.HasResolvConfSearchRegression(cfg.KubernetesConfig.KubernetesVersion) {
		return nil
	}
	if _, err := k.c.RunCmd(exec.Command("sudo", "cp", "/etc/resolv.conf", "/etc/kubelet-resolv.conf")); err != nil {
		return errors.Wrap(err, "copy")
	}
	if _, err := k.c.RunCmd(exec.Command("sudo", "sed", "-i", "-e", "s/^search .$//", "/etc/kubelet-resolv.conf")); err != nil {
		return errors.Wrap(err, "sed")
	}

	return nil
}

// kubectlPath returns the path to the kubelet
func kubectlPath(cfg config.ClusterConfig) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", cfg.KubernetesConfig.KubernetesVersion, "kubectl")
}

func (k *Bootstrapper) LabelAndUntaintNode(cfg config.ClusterConfig, n config.Node) error {
	return k.labelAndUntaintNode(cfg, n)
}

// labelAndUntaintNode applies minikube labels to node and removes NoSchedule taints that might be set to secondary control-plane nodes by default in ha (multi-control plane) cluster.
func (k *Bootstrapper) labelAndUntaintNode(cfg config.ClusterConfig, n config.Node) error {
	// time node was created. time format is based on ISO 8601 (RFC 3339)
	// converting - and : to _ because of Kubernetes label restriction
	createdAtLbl := "minikube.k8s.io/updated_at=" + time.Now().Format("2006_01_02T15_04_05_0700")

	verLbl := "minikube.k8s.io/version=" + version.GetVersion()
	commitLbl := "minikube.k8s.io/commit=" + version.GetGitCommitID()
	profileNameLbl := "minikube.k8s.io/name=" + cfg.Name

	// ensure that "primary" label is applied only to the 1st node in the cluster (used eg for placing ingress there)
	// this is used to uniquely distinguish that from other nodes in multi-master/multi-control-plane cluster config
	primaryLbl := "minikube.k8s.io/primary=false"
	if config.IsPrimaryControlPlane(cfg, n) {
		primaryLbl = "minikube.k8s.io/primary=true"
	}

	ctx, cancel := context.WithTimeout(context.Background(), applyTimeoutSeconds*time.Second)
	defer cancel()

	// node name is usually based on profile/cluster name, except for "none" driver where it assumes hostname
	nodeName := config.MachineName(cfg, n)
	if driver.IsNone(cfg.Driver) {
		if n, err := os.Hostname(); err == nil {
			nodeName = n
		}
	}

	// example:
	// sudo /var/lib/minikube/binaries/<version>/kubectl --kubeconfig=/var/lib/minikube/kubeconfig label --overwrite nodes test-357 minikube.k8s.io/version=<version> minikube.k8s.io/commit=aa91f39ffbcf27dcbb93c4ff3f457c54e585cf4a-dirty minikube.k8s.io/name=p1 minikube.k8s.io/updated_at=2020_02_20T12_05_35_0700
	cmd := exec.CommandContext(ctx, "sudo", kubectlPath(cfg), fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
		"label", "--overwrite", "nodes", nodeName, createdAtLbl, verLbl, commitLbl, profileNameLbl, primaryLbl)
	if _, err := k.c.RunCmd(cmd); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return errors.Wrapf(err, "timeout apply node labels")
		}
		return errors.Wrapf(err, "apply node labels")
	}

	// primary control-plane and worker nodes should be untainted by default
	if n.ControlPlane && !config.IsPrimaryControlPlane(cfg, n) {
		// example:
		// sudo /var/lib/minikube/binaries/<version>/kubectl --kubeconfig=/var/lib/minikube/kubeconfig taint nodes test-357 node-role.kubernetes.io/control-plane:NoSchedule-
		cmd := exec.CommandContext(ctx, "sudo", kubectlPath(cfg), fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
			"taint", "nodes", config.MachineName(cfg, n), "node-role.kubernetes.io/control-plane:NoSchedule-")
		if _, err := k.c.RunCmd(cmd); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return errors.Wrapf(err, "timeout remove node taints")
			}
			return errors.Wrapf(err, "remove node taints")
		}
	}

	return nil
}

// elevateKubeSystemPrivileges gives the kube-system service account cluster admin privileges to work with RBAC.
func (k *Bootstrapper) elevateKubeSystemPrivileges(cfg config.ClusterConfig) error {
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s to wait for elevateKubeSystemPrivileges", time.Since(start))
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
		// double checking default sa was created.
		// good for ensuring using minikube in CI is robust.
		checkSA := func(_ context.Context) (bool, error) {
			cmd = exec.Command("sudo", kubectlPath(cfg),
				"get", "sa", "default", fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")))
			rr, err = k.c.RunCmd(cmd)
			if err != nil {
				return false, nil
			}
			return true, nil
		}

		// retry up to make sure SA is created
		if err := wait.PollUntilContextTimeout(context.Background(), kconst.APICallRetryInterval, time.Minute, true, checkSA); err != nil {
			return errors.Wrap(err, "ensure sa was created")
		}
	}
	return nil
}

// stopKubeSystem stops all the containers in the kube-system to prevent #8740 when doing hot upgrade
func (k *Bootstrapper) stopKubeSystem(cfg config.ClusterConfig) error {
	klog.Info("stopping kube-system containers ...")
	cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Socket: cfg.KubernetesConfig.CRISocket, Runner: k.c})
	if err != nil {
		return errors.Wrap(err, "new cruntime")
	}

	ids, err := cr.ListContainers(cruntime.ListContainersOptions{Namespaces: []string{"kube-system"}})
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
		out.Styled(style.Tip, "Please free up disk or prune images.")
		if driver.IsVM(drv) {
			out.Styled(style.Stopped, "Please create a cluster with bigger disk size: `minikube start --disk SIZE_MB` ")
		} else if drv == oci.Docker && runtime.GOOS != "linux" {
			out.Styled(style.Stopped, "Please increase Desktop's disk size.")
			if runtime.GOOS == "darwin" {
				out.Styled(style.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
			}
			if runtime.GOOS == "windows" {
				out.Styled(style.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
			}
		}
		out.ErrLn("")
		return
	}

	if memErr, ok := err.(*kverify.ErrMemoryPressure); ok {
		out.ErrLn("")
		klog.Warning(memErr)
		out.WarningT("The node {{.name}} has ran out of memory.", out.V{"name": name})
		out.Styled(style.Tip, "Check if you have unnecessary pods running by running 'kubectl get po -A")
		if driver.IsVM(drv) {
			out.Styled(style.Stopped, "Consider creating a cluster with larger memory size using `minikube start --memory SIZE_MB` ")
		} else if drv == oci.Docker && runtime.GOOS != "linux" {
			out.Styled(style.Stopped, "Consider increasing Docker Desktop's memory size.")
			if runtime.GOOS == "darwin" {
				out.Styled(style.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
			}
			if runtime.GOOS == "windows" {
				out.Styled(style.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
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
