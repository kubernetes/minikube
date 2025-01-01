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

package node

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blang/semver/v4"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/bootstrapper/kubeadm"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/proxy"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/network"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

const waitTimeout = "wait-timeout"

var (
	kicGroup   errgroup.Group
	cacheGroup errgroup.Group
)

// Starter is a struct with all the necessary information to start a node
type Starter struct {
	Runner         command.Runner
	PreExists      bool
	StopK8s        bool
	MachineAPI     libmachine.API
	Host           *host.Host
	Cfg            *config.ClusterConfig
	Node           *config.Node
	ExistingAddons map[string]bool
}

// Start spins up a guest and starts the Kubernetes node.
func Start(starter Starter) (*kubeconfig.Settings, error) { // nolint:gocyclo
	var wg sync.WaitGroup
	stopk8s, err := handleNoKubernetes(starter)
	if err != nil {
		return nil, err
	}
	if stopk8s {
		nv := semver.Version{Major: 0, Minor: 0, Patch: 0}
		cr := configureRuntimes(starter.Runner, *starter.Cfg, nv)

		showNoK8sVersionInfo(cr)

		configureMounts(&wg, *starter.Cfg)
		return nil, config.Write(viper.GetString(config.ProfileName), starter.Cfg)
	}

	// wait for preloaded tarball to finish downloading before configuring runtimes
	waitCacheRequiredImages(&cacheGroup)

	sv, err := util.ParseKubernetesVersion(starter.Node.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse Kubernetes version")
	}

	// configure the runtime (docker, containerd, crio)
	cr := configureRuntimes(starter.Runner, *starter.Cfg, sv)

	// check if installed runtime is compatible with current minikube code
	if err = cruntime.CheckCompatibility(cr); err != nil {
		return nil, err
	}

	showVersionInfo(starter.Node.KubernetesVersion, cr)

	// add "host.minikube.internal" dns alias (intentionally non-fatal)
	hostIP, err := cluster.HostIP(starter.Host, starter.Cfg.Name)
	if err != nil {
		klog.Errorf("Unable to get host IP: %v", err)
	} else if err := machine.AddHostAlias(starter.Runner, constants.HostAlias, hostIP); err != nil {
		klog.Errorf("Unable to add minikube host alias: %v", err)
	}

	var kcs *kubeconfig.Settings
	var bs bootstrapper.Bootstrapper
	if config.IsPrimaryControlPlane(*starter.Cfg, *starter.Node) {
		// [re]start primary control-plane node
		kcs, bs, err = startPrimaryControlPlane(starter, cr)
		if err != nil {
			return nil, err
		}
		// configure CoreDNS concurently from primary control-plane node only and only on first node start
		if !starter.PreExists {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// inject {"host.minikube.internal": hostIP} record into coredns for primary control-plane node host ip
				if hostIP != nil {
					if err := addCoreDNSEntry(starter.Runner, constants.HostAlias, hostIP.String(), *starter.Cfg); err != nil {
						klog.Warningf("Unable to inject {%q: %s} record into CoreDNS: %v", constants.HostAlias, hostIP.String(), err)
						out.Err("Failed to inject host.minikube.internal into CoreDNS, this will limit the pods access to the host IP")
					}
				}
				// scale down CoreDNS from default 2 to 1 replica only for non-ha (non-multi-control plane) cluster and if optimisation is not disabled
				if !starter.Cfg.DisableOptimizations && !config.IsHA(*starter.Cfg) {
					if err := kapi.ScaleDeployment(starter.Cfg.Name, meta.NamespaceSystem, kconst.CoreDNSDeploymentName, 1); err != nil {
						klog.Errorf("Unable to scale down deployment %q in namespace %q to 1 replica: %v", kconst.CoreDNSDeploymentName, meta.NamespaceSystem, err)
					}
				}
			}()
		}
	} else {
		bs, err = cluster.Bootstrapper(starter.MachineAPI, viper.GetString(cmdcfg.Bootstrapper), *starter.Cfg, starter.Runner)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get bootstrapper")
		}

		// for ha (multi-control plane) cluster, use already running control-plane node to copy over certs to this secondary control-plane node
		cpr := mustload.Running(starter.Cfg.Name).CP.Runner
		if err = bs.SetupCerts(*starter.Cfg, *starter.Node, cpr); err != nil {
			return nil, errors.Wrap(err, "setting up certs")
		}

		if err := bs.UpdateNode(*starter.Cfg, *starter.Node, cr); err != nil {
			return nil, errors.Wrap(err, "update node")
		}

		// join cluster only on first node start
		// except for vm driver in non-ha (non-multi-control plane) cluster - fallback to old behaviour
		if !starter.PreExists || (driver.IsVM(starter.Cfg.Driver) && !config.IsHA(*starter.Cfg)) {
			// make sure to use the command runner for the primary control plane to generate the join token
			pcpBs, err := cluster.ControlPlaneBootstrapper(starter.MachineAPI, starter.Cfg, viper.GetString(cmdcfg.Bootstrapper))
			if err != nil {
				return nil, errors.Wrap(err, "get primary control-plane bootstrapper")
			}
			if err := joinCluster(starter, pcpBs, bs); err != nil {
				return nil, errors.Wrap(err, "join node to cluster")
			}
		}
	}

	go configureMounts(&wg, *starter.Cfg)

	wg.Add(1)
	go func() {
		defer wg.Done()
		profile, err := config.LoadProfile(starter.Cfg.Name)
		if err != nil {
			out.FailureT("Unable to load profile: {{.error}}", out.V{"error": err})
		}
		if err := CacheAndLoadImagesInConfig([]*config.Profile{profile}); err != nil {
			out.FailureT("Unable to push cached images: {{.error}}", out.V{"error": err})
		}
	}()

	// enable addons, both old and new!
	addonList := viper.GetStringSlice(config.AddonListFlag)
	enabledAddons := make(chan []string, 1)
	if starter.ExistingAddons != nil {
		if viper.GetBool("force") {
			addons.Force = true
		}
		list := addons.ToEnable(starter.Cfg, starter.ExistingAddons, addonList)
		wg.Add(1)
		go addons.Enable(&wg, starter.Cfg, list, enabledAddons)
	}

	// discourage use of the virtualbox driver
	if starter.Cfg.Driver == driver.VirtualBox && viper.GetBool(config.WantVirtualBoxDriverWarning) {
		warnVirtualBox()
	}

	// special ops for "none" driver on control-plane node, like change minikube directory
	if starter.Node.ControlPlane && driver.IsNone(starter.Cfg.Driver) {
		prepareNone()
	}

	// for ha (multi-control plane) cluster, primary control-plane node will not come up alone until secondary joins
	if config.IsHA(*starter.Cfg) && config.IsPrimaryControlPlane(*starter.Cfg, *starter.Node) {
		klog.Infof("HA (multi-control plane) cluster: will skip waiting for primary control-plane node %+v", starter.Node)
	} else {
		klog.Infof("Will wait %s for node %+v", viper.GetDuration(waitTimeout), starter.Node)
		if err := bs.WaitForNode(*starter.Cfg, *starter.Node, viper.GetDuration(waitTimeout)); err != nil {
			return nil, errors.Wrapf(err, "wait %s for node", viper.GetDuration(waitTimeout))
		}
	}

	klog.Infof("waiting for startup goroutines ...")
	wg.Wait()

	// update config with enabled addons
	if starter.ExistingAddons != nil {
		klog.Infof("waiting for cluster config update ...")
		if ea, ok := <-enabledAddons; ok {
			addons.UpdateConfigToEnable(starter.Cfg, ea)
		}
	} else {
		addons.UpdateConfigToDisable(starter.Cfg)
	}

	// Write enabled addons to the config before completion
	klog.Infof("writing updated cluster config ...")
	return kcs, config.Write(viper.GetString(config.ProfileName), starter.Cfg)
}

// handleNoKubernetes handles starting minikube without Kubernetes.
func handleNoKubernetes(starter Starter) (bool, error) {
	// Do not bootstrap cluster if --no-kubernetes.
	if starter.Node.KubernetesVersion == constants.NoKubernetesVersion {
		// Stop existing Kubernetes node if applicable.
		if starter.StopK8s {
			cr, err := cruntime.New(cruntime.Config{Type: starter.Cfg.KubernetesConfig.ContainerRuntime, Runner: starter.Runner, Socket: starter.Cfg.KubernetesConfig.CRISocket})
			if err != nil {
				return false, err
			}
			kubeadm.StopKubernetes(starter.Runner, cr)
		}
		return true, config.Write(viper.GetString(config.ProfileName), starter.Cfg)
	}
	return false, nil
}

// startPrimaryControlPlane starts control-plane node.
func startPrimaryControlPlane(starter Starter, cr cruntime.Manager) (*kubeconfig.Settings, bootstrapper.Bootstrapper, error) {
	if !config.IsPrimaryControlPlane(*starter.Cfg, *starter.Node) {
		return nil, nil, fmt.Errorf("node not marked as primary control-plane")
	}

	if config.IsHA(*starter.Cfg) {
		n, err := network.Inspect(starter.Node.IP)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "inspect network")
		}
		// update cluster config
		starter.Cfg.KubernetesConfig.APIServerHAVIP = n.ClientMax // last available ip from node's subnet, should've been reserved already
	}

	// must be written before bootstrap, otherwise health checks may flake due to stale IP
	kcs := setupKubeconfig(*starter.Host, *starter.Cfg, *starter.Node, starter.Cfg.Name)

	// setup kubeadm (must come after setupKubeconfig)
	bs, err := setupKubeadm(starter.MachineAPI, *starter.Cfg, *starter.Node, starter.Runner)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to setup kubeadm")
	}

	if err := bs.StartCluster(*starter.Cfg); err != nil {
		ExitIfFatal(err, false)
		out.LogEntries("Error starting cluster", err, logs.FindProblems(cr, bs, *starter.Cfg, starter.Runner))
		return nil, bs, err
	}

	// Write the kubeconfig to the file system after everything required (like certs) are created by the bootstrapper.
	if err := kubeconfig.Update(kcs); err != nil {
		return nil, bs, errors.Wrap(err, "Failed kubeconfig update")
	}

	return kcs, bs, nil
}

// joinCluster adds new or prepares and then adds existing node to the cluster.
func joinCluster(starter Starter, cpBs bootstrapper.Bootstrapper, bs bootstrapper.Bootstrapper) error {
	start := time.Now()
	klog.Infof("joinCluster: %+v", starter.Cfg)
	defer func() {
		klog.Infof("duration metric: took %s to joinCluster", time.Since(start))
	}()

	role := "worker"
	if starter.Node.ControlPlane {
		role = "control-plane"
	}

	// avoid "error execution phase kubelet-start: a Node with name "<name>" and status "Ready" already exists in the cluster.
	// You must delete the existing Node or change the name of this new joining Node"
	if starter.PreExists {
		klog.Infof("removing existing %s node %q before attempting to rejoin cluster: %+v", role, starter.Node.Name, starter.Node)
		if _, err := teardown(*starter.Cfg, starter.Node.Name); err != nil {
			klog.Errorf("error removing existing %s node %q before rejoining cluster, will continue anyway: %v", role, starter.Node.Name, err)
		}
		klog.Infof("successfully removed existing %s node %q from cluster: %+v", role, starter.Node.Name, starter.Node)
	}

	joinCmd, err := cpBs.GenerateToken(*starter.Cfg)
	if err != nil {
		return fmt.Errorf("error generating join token: %w", err)
	}

	join := func() error {
		klog.Infof("trying to join %s node %q to cluster: %+v", role, starter.Node.Name, starter.Node)
		if err := bs.JoinCluster(*starter.Cfg, *starter.Node, joinCmd); err != nil {
			klog.Errorf("%s node failed to join cluster, will retry: %v", role, err)

			// reset node to revert any changes made by previous kubeadm init/join
			klog.Infof("resetting %s node %q before attempting to rejoin cluster...", role, starter.Node.Name)
			if _, err := starter.Runner.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s reset --force", bsutil.InvokeKubeadm(starter.Cfg.KubernetesConfig.KubernetesVersion)))); err != nil {
				klog.Infof("kubeadm reset failed, continuing anyway: %v", err)
			} else {
				klog.Infof("successfully reset %s node %q", role, starter.Node.Name)
			}

			return err
		}
		return nil
	}
	if err := retry.Expo(join, 10*time.Second, 3*time.Minute); err != nil {
		return fmt.Errorf("error joining %s node %q to cluster: %w", role, starter.Node.Name, err)
	}

	if err := cpBs.LabelAndUntaintNode(*starter.Cfg, *starter.Node); err != nil {
		return fmt.Errorf("error applying %s node %q label: %w", role, starter.Node.Name, err)
	}
	return nil
}

// Provision provisions the machine/container for the node
func Provision(cc *config.ClusterConfig, n *config.Node, delOnFail bool) (command.Runner, bool, libmachine.API, *host.Host, error) {
	register.Reg.SetStep(register.StartingNode)
	name := config.MachineName(*cc, *n)

	// Be explicit with each case for the sake of translations
	if cc.KubernetesConfig.KubernetesVersion == constants.NoKubernetesVersion {
		out.Step(style.ThumbsUp, "Starting minikube without Kubernetes in cluster {{.cluster}}", out.V{"cluster": cc.Name})
	} else {
		role := "worker"
		if n.ControlPlane {
			role = "control-plane"
		}
		if config.IsPrimaryControlPlane(*cc, *n) {
			role = "primary control-plane"
		}
		out.Step(style.ThumbsUp, "Starting \"{{.node}}\" {{.role}} node in \"{{.cluster}}\" cluster", out.V{"node": name, "role": role, "cluster": cc.Name})
	}

	if driver.IsKIC(cc.Driver) {
		beginDownloadKicBaseImage(&kicGroup, cc, viper.GetBool("download-only"))
	}

	if !driver.BareMetal(cc.Driver) {
		beginCacheKubernetesImages(&cacheGroup, cc.KubernetesConfig.ImageRepository, n.KubernetesVersion, cc.KubernetesConfig.ContainerRuntime, cc.Driver)
	}

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot.
	// Hence, SaveProfile must be called before startHost, and again afterwards when we know the IP.
	if err := config.SaveProfile(viper.GetString(config.ProfileName), cc); err != nil {
		return nil, false, nil, nil, errors.Wrap(err, "Failed to save config")
	}

	handleDownloadOnly(&cacheGroup, &kicGroup, n.KubernetesVersion, cc.KubernetesConfig.ContainerRuntime, cc.Driver)
	if driver.IsKIC(cc.Driver) {
		waitDownloadKicBaseImage(&kicGroup)
	}

	return startMachine(cc, n, delOnFail)
}

// ConfigureRuntimes does what needs to happen to get a runtime going.
func configureRuntimes(runner cruntime.CommandRunner, cc config.ClusterConfig, kv semver.Version) cruntime.Manager {
	co := cruntime.Config{
		Type:              cc.KubernetesConfig.ContainerRuntime,
		Socket:            cc.KubernetesConfig.CRISocket,
		Runner:            runner,
		NetworkPlugin:     cc.KubernetesConfig.NetworkPlugin,
		ImageRepository:   cc.KubernetesConfig.ImageRepository,
		KubernetesVersion: kv,
		InsecureRegistry:  cc.InsecureRegistry,
	}
	if cc.GPUs != "" {
		co.GPUs = cc.GPUs
	}
	cr, err := cruntime.New(co)
	if err != nil {
		exit.Error(reason.InternalRuntime, "Failed runtime", err)
	}

	// 87-podman.conflist cni conf potentially conflicts with others and is created by podman on its first invocation,
	// so we "provoke" it here to ensure it's generated and that we can disable it
	// note: using 'help' or '--help' would be cheaper, but does not trigger that; 'version' seems to be next best option
	if co.Type == constants.CRIO {
		_, _ = runner.RunCmd(exec.Command("sudo", "sh", "-c", `podman version >/dev/null`))
	}
	// ensure loopback is properly configured
	// make sure container runtime is restarted afterwards for these changes to take effect
	disableLoopback := co.Type == constants.CRIO
	if err := cni.ConfigureLoopbackCNI(runner, disableLoopback); err != nil {
		klog.Warningf("unable to name loopback interface in configureRuntimes: %v", err)
	}
	// ensure all default CNI(s) are properly configured on each and every node (re)start
	// make sure container runtime is restarted afterwards for these changes to take effect
	if err := cni.ConfigureDefaultBridgeCNIs(runner, cc.KubernetesConfig.NetworkPlugin); err != nil {
		klog.Errorf("unable to disable preinstalled bridge CNI(s): %v", err)
	}

	inUserNamespace := strings.Contains(cc.KubernetesConfig.FeatureGates, "KubeletInUserNamespace=true")
	// for docker container runtime: ensure containerd is properly configured by calling Enable(), as docker could be bound to containerd
	// it will also "soft" start containerd, but it will not disable others; docker will disable containerd if not used in the next step
	if co.Type == constants.Docker {
		containerd, err := cruntime.New(cruntime.Config{
			Type:              constants.Containerd,
			Socket:            "", // use default
			Runner:            co.Runner,
			ImageRepository:   co.ImageRepository,
			KubernetesVersion: co.KubernetesVersion,
			InsecureRegistry:  co.InsecureRegistry})
		if err == nil {
			err = containerd.Enable(false, cgroupDriver(cc), inUserNamespace) // do not disableOthers, as it's not primary cr
		}
		if err != nil {
			klog.Warningf("cannot ensure containerd is configured properly and reloaded for docker - cluster might be unstable: %v", err)
		}
	}

	disableOthers := !driver.BareMetal(cc.Driver)
	if err = cr.Enable(disableOthers, cgroupDriver(cc), inUserNamespace); err != nil {
		exit.Error(reason.RuntimeEnable, "Failed to enable container runtime", err)
	}

	// Wait for the CRI to be "live", before returning it
	if err = waitForCRISocket(runner, cr.SocketPath(), 60, 1); err != nil {
		exit.Error(reason.RuntimeEnable, "Failed to start container runtime", err)
	}

	// Wait for the CRI to actually work, before returning
	if err = waitForCRIVersion(runner, cr.SocketPath(), 60, 10); err != nil {
		exit.Error(reason.RuntimeEnable, "Failed to start container runtime", err)
	}

	return cr
}

// cgroupDriver returns cgroup driver that should be used to further configure container runtime, node(s) and cluster.
// It is based on:
// - (forced) user preference (set via flags or env), if present, or
// - default settings for vm or ssh driver, if user, or
// - host os config detection, if possible.
// Possible mappings are: "v1" (legacy) cgroups => "cgroupfs", "v2" (unified) cgroups => "systemd" and "" (unknown) cgroups => constants.DefaultCgroupDriver.
// Note: starting from k8s v1.22, "kubeadm clusters should be using the systemd driver":
// ref: https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.22.md#no-really-you-must-read-this-before-you-upgrade
// ref: https://kubernetes.io/docs/setup/production-environment/container-runtimes/#cgroup-drivers
// ref: https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/configure-cgroup-driver/
func cgroupDriver(cc config.ClusterConfig) string {
	klog.Info("detecting cgroup driver to use...")

	// check flags for user preference
	if viper.GetBool("force-systemd") {
		klog.Infof("using %q cgroup driver as enforced via flags", constants.SystemdCgroupDriver)
		return constants.SystemdCgroupDriver
	}

	// check env for user preference
	env := os.Getenv(constants.MinikubeForceSystemdEnv)
	if force, err := strconv.ParseBool(env); env != "" && err == nil && force {
		klog.Infof("using %q cgroup driver as enforced via env", constants.SystemdCgroupDriver)
		return constants.SystemdCgroupDriver
	}

	// vm driver uses iso that boots with cgroupfs cgroup driver by default atm (keep in sync!)
	if driver.IsVM(cc.Driver) {
		return constants.CgroupfsCgroupDriver
	}

	// for "remote baremetal", we assume cgroupfs and user can "force-systemd" with flag to override
	// potential improvement: use systemd as default (in line with k8s) and allow user to override it with new flag (eg, "cgroup-driver", that would replace "force-systemd")
	if driver.IsSSH(cc.Driver) {
		return constants.CgroupfsCgroupDriver
	}

	// in all other cases - try to detect and use what's on user's machine
	return detect.CgroupDriver()
}

func pathExists(runner cruntime.CommandRunner, path string) (bool, error) {
	_, err := runner.RunCmd(exec.Command("stat", path))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func waitForCRISocket(runner cruntime.CommandRunner, socket string, wait int, interval int) error {

	if socket == "" || socket == "/var/run/dockershim.sock" {
		return nil
	}

	klog.Infof("Will wait %ds for socket path %s", wait, socket)

	chkPath := func() error {
		e, err := pathExists(runner, socket)
		if err != nil {
			return err
		}
		if !e {
			return &retry.RetriableError{Err: err}
		}
		return nil
	}
	return retry.Expo(chkPath, time.Duration(interval)*time.Second, time.Duration(wait)*time.Second)
}

func waitForCRIVersion(runner cruntime.CommandRunner, socket string, wait int, interval int) error {

	if socket == "" || socket == "/var/run/dockershim.sock" {
		return nil
	}

	klog.Infof("Will wait %ds for crictl version", wait)

	cmd := exec.Command("which", "crictl")
	rr, err := runner.RunCmd(cmd)
	if err != nil {
		return err
	}
	crictl := strings.TrimSuffix(rr.Stdout.String(), "\n")

	chkInfo := func() error {
		args := []string{crictl, "version"}
		cmd := exec.Command("sudo", args...)
		rr, err := runner.RunCmd(cmd)
		if err != nil && !os.IsNotExist(err) {
			return &retry.RetriableError{Err: err}
		}
		klog.Info(rr.Stdout.String())
		return nil
	}
	return retry.Expo(chkInfo, time.Duration(interval)*time.Second, time.Duration(wait)*time.Second)
}

// setupKubeadm adds any requested files into the VM before Kubernetes is started.
func setupKubeadm(mAPI libmachine.API, cfg config.ClusterConfig, n config.Node, r command.Runner) (bootstrapper.Bootstrapper, error) {
	deleteOnFailure := viper.GetBool("delete-on-failure")
	bs, err := cluster.Bootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper), cfg, r)
	if err != nil {
		klog.Errorf("Failed to get bootstrapper: %v", err)
		if !deleteOnFailure {
			exit.Error(reason.InternalBootstrapper, "Failed to get bootstrapper", err)
		}
		return nil, err
	}
	for _, eo := range cfg.KubernetesConfig.ExtraOptions {
		out.Infof("{{.extra_option_component_name}}.{{.key}}={{.value}}", out.V{"extra_option_component_name": eo.Component, "key": eo.Key, "value": eo.Value})
	}

	// Loads cached images, generates config files, download binaries
	// update cluster and set up certs

	if err := bs.UpdateCluster(cfg); err != nil {
		if !deleteOnFailure {
			if errors.Is(err, cruntime.ErrContainerRuntimeNotRunning) {
				exit.Error(reason.KubernetesInstallFailedRuntimeNotRunning, "Failed to update cluster", err)
			}
			exit.Error(reason.KubernetesInstallFailed, "Failed to update cluster", err)
		}
		klog.Errorf("Failed to update cluster: %v", err)
		return nil, err
	}

	if err := bs.SetupCerts(cfg, n, r); err != nil {
		if !deleteOnFailure {
			exit.Error(reason.GuestCert, "Failed to setup certs", err)
		}
		klog.Errorf("Failed to setup certs: %v", err)
		return nil, err
	}

	return bs, nil
}

// setupKubeconfig generates kubeconfig.
func setupKubeconfig(h host.Host, cc config.ClusterConfig, n config.Node, clusterName string) *kubeconfig.Settings {
	host := cc.KubernetesConfig.APIServerHAVIP
	port := cc.APIServerPort
	if !config.IsHA(cc) || driver.NeedsPortForward(cc.Driver) {
		var err error
		if host, _, port, err = driver.ControlPlaneEndpoint(&cc, &n, h.DriverName); err != nil {
			exit.Message(reason.DrvCPEndpoint, fmt.Sprintf("failed to construct cluster server address: %v", err), out.V{"profileArg": fmt.Sprintf("--profile=%s", clusterName)})
		}
	}
	addr := fmt.Sprintf("https://%s", net.JoinHostPort(host, strconv.Itoa(port)))

	if cc.KubernetesConfig.APIServerName != constants.APIServerName {
		addr = strings.ReplaceAll(addr, host, cc.KubernetesConfig.APIServerName)
	}

	kcs := &kubeconfig.Settings{
		ClusterName:          clusterName,
		Namespace:            cc.KubernetesConfig.Namespace,
		ClusterServerAddress: addr,
		ClientCertificate:    localpath.ClientCert(cc.Name),
		ClientKey:            localpath.ClientKey(cc.Name),
		CertificateAuthority: localpath.CACert(),
		KeepContext:          cc.KeepContext,
		EmbedCerts:           cc.EmbedCerts,
	}

	kcs.SetPath(kubeconfig.PathFromEnv())
	return kcs
}

// StartMachine starts a VM
func startMachine(cfg *config.ClusterConfig, node *config.Node, delOnFail bool) (runner command.Runner, preExists bool, machineAPI libmachine.API, host *host.Host, err error) {
	m, err := machine.NewAPIClient()
	if err != nil {
		return runner, preExists, m, host, errors.Wrap(err, "Failed to get machine client")
	}
	host, preExists, err = startHostInternal(m, cfg, node, delOnFail)
	if err != nil {
		return runner, preExists, m, host, errors.Wrap(err, "Failed to start host")
	}
	runner, err = machine.CommandRunner(host)
	if err != nil {
		return runner, preExists, m, host, errors.Wrap(err, "Failed to get command runner")
	}

	ip, err := validateNetwork(host, runner, cfg.KubernetesConfig.ImageRepository)
	if err != nil {
		return runner, preExists, m, host, errors.Wrap(err, "Failed to validate network")
	}

	if driver.IsQEMU(host.Driver.DriverName()) && network.IsBuiltinQEMU(cfg.Network) {
		apiServerPort, err := getPort()
		if err != nil {
			return runner, preExists, m, host, errors.Wrap(err, "Failed to find apiserver port")
		}
		cfg.APIServerPort = apiServerPort
	}

	// Bypass proxy for minikube's vm host ip
	err = proxy.ExcludeIP(ip)
	if err != nil {
		out.FailureT("Failed to set NO_PROXY Env. Please use `export NO_PROXY=$NO_PROXY,{{.ip}}`.", out.V{"ip": ip})
	}

	return runner, preExists, m, host, err
}

// getPort asks the kernel for a free open port that is ready to use
func getPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return -1, errors.Errorf("Error accessing port %d", addr.Port)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// startHostInternal starts a new minikube host using a VM or None
func startHostInternal(api libmachine.API, cc *config.ClusterConfig, n *config.Node, delOnFail bool) (*host.Host, bool, error) {
	host, exists, err := machine.StartHost(api, cc, n)
	if err == nil {
		return host, exists, nil
	}
	klog.Warningf("error starting host: %v", err)
	// NOTE: People get very cranky if you delete their preexisting VM. Only delete new ones.
	if !exists {
		err := machine.DeleteHost(api, config.MachineName(*cc, *n))
		if err != nil {
			klog.Warningf("delete host: %v", err)
		}
	}

	if err, ff := errors.Cause(err).(*oci.FailFastError); ff {
		klog.Infof("will skip retrying to create machine because error is not retriable: %v", err)
		return host, exists, err
	}

	out.ErrT(style.Embarrassed, "StartHost failed, but will try again: {{.error}}", out.V{"error": err})
	klog.Info("Will try again in 5 seconds ...")
	// Try again, but just once to avoid making the logs overly confusing
	time.Sleep(5 * time.Second)

	if delOnFail {
		klog.Info("Deleting existing host since delete-on-failure was set.")
		// Delete the failed existing host
		err := machine.DeleteHost(api, config.MachineName(*cc, *n))
		if err != nil {
			klog.Warningf("delete host: %v", err)
		}
	}

	host, exists, err = machine.StartHost(api, cc, n)
	if err == nil {
		return host, exists, nil
	}

	// Don't use host.Driver to avoid nil pointer deref
	drv := cc.Driver
	out.ErrT(style.Sad, `Failed to start {{.driver}} {{.driver_type}}. Running "{{.cmd}}" may fix it: {{.error}}`, out.V{"driver": drv, "driver_type": driver.MachineType(drv), "cmd": mustload.ExampleCmd(cc.Name, "delete"), "error": err})
	return host, exists, err
}

// validateNetwork tries to catch network problems as soon as possible
func validateNetwork(h *host.Host, r command.Runner, imageRepository string) (string, error) {
	ip, err := h.Driver.GetIP()
	if err != nil {
		return ip, err
	}

	optSeen := false
	warnedOnce := false
	for _, k := range proxy.EnvVars {
		if v := os.Getenv(k); v != "" {
			if !optSeen {
				out.Styled(style.Internet, "Found network options:")
				optSeen = true
			}
			k = strings.ToUpper(k) // let's get the key right away to mask password from output
			// If http(s)_proxy contains password, let's not splatter on the screen
			if k == "HTTP_PROXY" || k == "HTTPS_PROXY" {
				v = util.MaskProxyPassword(v)
			}
			out.Infof("{{.key}}={{.value}}", out.V{"key": k, "value": v})
			ipExcluded := proxy.IsIPExcluded(ip) // Skip warning if minikube ip is already in NO_PROXY
			if (k == "HTTP_PROXY" || k == "HTTPS_PROXY") && !ipExcluded && !warnedOnce {
				out.WarningT("You appear to be using a proxy, but your NO_PROXY environment does not include the minikube IP ({{.ip_address}}).", out.V{"ip_address": ip})
				out.Styled(style.Documentation, "Please see {{.documentation_url}} for more details", out.V{"documentation_url": "https://minikube.sigs.k8s.io/docs/handbook/vpn_and_proxy/"})
				warnedOnce = true
			}
		}
	}

	if shouldTrySSH(h.Driver.DriverName(), ip) {
		if err := trySSH(h, ip); err != nil {
			return ip, err
		}
	}

	// Non-blocking
	go tryRegistry(r, h.Driver.DriverName(), imageRepository, ip)
	return ip, nil
}

func shouldTrySSH(driverName, ip string) bool {
	if driver.BareMetal(driverName) || driver.IsKIC(driverName) {
		return false
	}
	// QEMU with user network
	if driver.IsQEMU(driverName) && ip == "127.0.0.1" {
		return false
	}
	return true
}

func trySSH(h *host.Host, ip string) error {
	if viper.GetBool("force") {
		return nil
	}

	sshAddr := net.JoinHostPort(ip, "22")

	dial := func() (err error) {
		d := net.Dialer{Timeout: 3 * time.Second}
		conn, err := d.Dial("tcp", sshAddr)
		if err != nil {
			klog.Warningf("dial failed (will retry): %v", err)
			return err
		}
		_ = conn.Close()
		return nil
	}

	err := retry.Expo(dial, time.Second, 13*time.Second)
	if err != nil {
		out.ErrT(style.Failure, `minikube is unable to connect to the VM: {{.error}}

	This is likely due to one of two reasons:

	- VPN or firewall interference
	- {{.hypervisor}} network configuration issue

	Suggested workarounds:

	- Disable your local VPN or firewall software
	- Configure your local VPN or firewall to allow access to {{.ip}}
	- Restart or reinstall {{.hypervisor}}
	- Use an alternative --vm-driver
	- Use --force to override this connectivity check
	`, out.V{"error": err, "hypervisor": h.Driver.DriverName(), "ip": ip})
	}

	return err
}

// tryRegistry tries to connect to the image repository
func tryRegistry(r command.Runner, driverName, imageRepository, ip string) {
	// 2 second timeout. For best results, call tryRegistry in a non-blocking manner.
	opts := []string{"-sS", "-m", "2"}

	proxy := os.Getenv("HTTPS_PROXY")
	if proxy != "" && !strings.HasPrefix(proxy, "localhost") && !strings.HasPrefix(proxy, "127.0") {
		opts = append([]string{"-x", proxy}, opts...)
	}

	if imageRepository == "" {
		imageRepository = images.DefaultKubernetesRepo
	}

	curlTarget := fmt.Sprintf("https://%s/", imageRepository)
	opts = append(opts, curlTarget)
	exe := "curl"
	if runtime.GOOS == "windows" {
		exe = "curl.exe"
	}
	cmd := exec.Command(exe, opts...)
	if rr, err := r.RunCmd(cmd); err != nil {
		klog.Warningf("%s failed: %v", rr.Args, err)

		// using QEMU with the user network
		if driver.IsQEMU(driverName) && ip == "127.0.0.1" {
			out.WarningT("Due to DNS issues your cluster may have problems starting and you may not be able to pull images\nMore details available at: https://minikube.sigs.k8s.io/docs/drivers/qemu/#known-issues")
		}
		// now we shall also try whether this registry is reachable
		// outside the machine so that we can tell in the logs that if
		// the user's computer had any network issue or could it be
		// related to a network module config change in minikube ISO

		// We should skip the second check if the user is using the none
		// or ssh driver since there is no difference between an "inside"
		// and "outside" check on the none driver, and checking the host
		// on the ssh driver is not helpful.
		warning := "Failing to connect to {{.curlTarget}} from inside the minikube {{.type}}"
		if !driver.IsNone(driverName) && !driver.IsSSH(driverName) {
			if err := cmd.Run(); err != nil {
				// both inside and outside failed
				warning = "Failing to connect to {{.curlTarget}} from both inside the minikube {{.type}} and host machine"
			}
		}
		out.WarningT(warning, out.V{"curlTarget": curlTarget, "type": driver.MachineType(driverName)})

		out.ErrT(style.Tip, "To pull new external images, you may need to configure a proxy: https://minikube.sigs.k8s.io/docs/reference/networking/proxy/")
	}
}

// prepareNone prepares the user and host for the joy of the "none" driver
func prepareNone() {
	register.Reg.SetStep(register.ConfiguringLHEnv)
	out.Step(style.StartingNone, "Configuring local host environment ...")
	if viper.GetBool(config.WantNoneDriverWarning) {
		out.ErrT(style.Empty, "")
		out.WarningT("The 'none' driver is designed for experts who need to integrate with an existing VM")
		out.ErrT(style.Tip, "Most users should use the newer 'docker' driver instead, which does not require root!")
		out.ErrT(style.Documentation, "For more information, see: https://minikube.sigs.k8s.io/docs/reference/drivers/none/")
		out.ErrT(style.Empty, "")
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		home := os.Getenv("HOME")
		out.WarningT("kubectl and minikube configuration will be stored in {{.home_folder}}", out.V{"home_folder": home})
		out.WarningT("To use kubectl or minikube commands as your own user, you may need to relocate them. For example, to overwrite your own settings, run:")

		out.ErrT(style.Empty, "")
		out.ErrT(style.Command, "sudo mv {{.home_folder}}/.kube {{.home_folder}}/.minikube $HOME", out.V{"home_folder": home})
		out.ErrT(style.Command, "sudo chown -R $USER $HOME/.kube $HOME/.minikube")
		out.ErrT(style.Empty, "")

		out.ErrT(style.Tip, "This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true")
	}

	if err := util.MaybeChownDirRecursiveToMinikubeUser(localpath.MiniPath()); err != nil {
		exit.Message(reason.HostHomeChown, "Failed to change permissions for {{.minikube_dir_path}}: {{.error}}", out.V{"minikube_dir_path": localpath.MiniPath(), "error": err})
	}
}

// addCoreDNSEntry adds host name and IP record to the DNS by updating CoreDNS's ConfigMap.
// ref: https://coredns.io/plugins/hosts/
// note: there can be only one 'hosts' block in CoreDNS's ConfigMap (avoid "plugin/hosts: this plugin can only be used once per Server Block" error)
func addCoreDNSEntry(runner command.Runner, name, ip string, cc config.ClusterConfig) error {
	kubectl := kapi.KubectlBinaryPath(cc.KubernetesConfig.KubernetesVersion)
	kubecfg := path.Join(vmpath.GuestPersistentDir, "kubeconfig")

	// get current coredns configmap via kubectl
	get := fmt.Sprintf("sudo %s --kubeconfig=%s -n kube-system get configmap coredns -o yaml", kubectl, kubecfg)
	out, err := runner.RunCmd(exec.Command("/bin/bash", "-c", get))
	if err != nil {
		klog.Errorf("failed to get current CoreDNS ConfigMap: %v", err)
		return err
	}
	cm := strings.TrimSpace(out.Stdout.String())

	// check if this specific host entry already exists in coredns configmap, so not to duplicate/override it
	host := regexp.MustCompile(fmt.Sprintf(`(?smU)^ *hosts {.*%s.*}`, name))
	if host.MatchString(cm) {
		klog.Infof("CoreDNS already contains %q host record, skipping...", name)
		return nil
	}

	// inject hosts block with host record into coredns configmap
	sed := fmt.Sprintf("sed -e '/^        forward . \\/etc\\/resolv.conf.*/i \\        hosts {\\n           %s %s\\n           fallthrough\\n        }'", ip, name)
	// check if hosts block already exists in coredns configmap
	hosts := regexp.MustCompile(`(?smU)^ *hosts {.*}`)
	if hosts.MatchString(cm) {
		// inject host record into existing coredns configmap hosts block instead
		klog.Info("CoreDNS already contains hosts block, will inject host record there...")
		sed = fmt.Sprintf("sed -e '/^        hosts {.*/a \\           %s %s'", ip, name)
	}

	// check if logging is already enabled (via log plugin) in coredns configmap, so not to duplicate it
	logs := regexp.MustCompile(`(?smU)^ *log *$`)
	if !logs.MatchString(cm) {
		// inject log plugin into coredns configmap
		sed = fmt.Sprintf("%s -e '/^        errors *$/i \\        log'", sed)
	}

	// replace coredns configmap via kubectl
	replace := fmt.Sprintf("sudo %s --kubeconfig=%s replace -f -", kubectl, kubecfg)
	if _, err := runner.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("%s | %s | %s", get, sed, replace))); err != nil {
		klog.Errorf("failed to inject {%q: %s} host record into CoreDNS", name, ip)
		return err
	}
	klog.Infof("{%q: %s} host record injected into CoreDNS's ConfigMap", name, ip)

	return nil
}

// prints a warning to the console against the use of the 'virtualbox' driver, if alternatives are available and healthy
func warnVirtualBox() {
	var altDriverList strings.Builder
	for _, choice := range driver.Choices(true) {
		if !driver.IsVirtualBox(choice.Name) && choice.Priority != registry.Discouraged && choice.State.Installed && choice.State.Healthy {
			altDriverList.WriteString(fmt.Sprintf("\n\t- %s", choice.Name))
		}
	}

	if altDriverList.Len() != 0 {
		out.Boxed(`You have selected "virtualbox" driver, but there are better options !
For better performance and support consider using a different driver: {{.drivers}}

To turn off this warning run:

	$ minikube config set WantVirtualBoxDriverWarning false


To learn more about on minikube drivers checkout https://minikube.sigs.k8s.io/docs/drivers/
To see benchmarks checkout https://minikube.sigs.k8s.io/docs/benchmarks/cpuusage/

`, out.V{"drivers": altDriverList.String()})
	}
}
