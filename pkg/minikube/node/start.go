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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
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
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
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
	MachineAPI     libmachine.API
	Host           *host.Host
	Cfg            *config.ClusterConfig
	Node           *config.Node
	ExistingAddons map[string]bool
}

// Start spins up a guest and starts the Kubernetes node.
func Start(starter Starter, apiServer bool) (*kubeconfig.Settings, error) {
	// wait for preloaded tarball to finish downloading before configuring runtimes
	waitCacheRequiredImages(&cacheGroup)

	sv, err := util.ParseKubernetesVersion(starter.Node.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse Kubernetes version")
	}

	// configure the runtime (docker, containerd, crio)
	cr := configureRuntimes(starter.Runner, *starter.Cfg, sv)
	showVersionInfo(starter.Node.KubernetesVersion, cr)

	// Add "host.minikube.internal" DNS alias (intentionally non-fatal)
	hostIP, err := cluster.HostIP(starter.Host, starter.Cfg.Name)
	if err != nil {
		klog.Errorf("Unable to get host IP: %v", err)
	} else if err := machine.AddHostAlias(starter.Runner, constants.HostAlias, hostIP); err != nil {
		klog.Errorf("Unable to add host alias: %v", err)
	}

	var bs bootstrapper.Bootstrapper
	var kcs *kubeconfig.Settings
	if apiServer {
		// Must be written before bootstrap, otherwise health checks may flake due to stale IP
		kcs = setupKubeconfig(starter.Host, starter.Cfg, starter.Node, starter.Cfg.Name)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to setup kubeconfig")
		}

		// setup kubeadm (must come after setupKubeconfig)
		bs = setupKubeAdm(starter.MachineAPI, *starter.Cfg, *starter.Node, starter.Runner)
		err = bs.StartCluster(*starter.Cfg)
		if err != nil {
			ExitIfFatal(err)
			out.LogEntries("Error starting cluster", err, logs.FindProblems(cr, bs, *starter.Cfg, starter.Runner))
			return nil, err
		}

		// write the kubeconfig to the file system after everything required (like certs) are created by the bootstrapper
		if err := kubeconfig.Update(kcs); err != nil {
			return nil, errors.Wrap(err, "Failed kubeconfig update")
		}
	} else {
		bs, err = cluster.Bootstrapper(starter.MachineAPI, viper.GetString(cmdcfg.Bootstrapper), *starter.Cfg, starter.Runner)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get bootstrapper")
		}

		if err = bs.SetupCerts(starter.Cfg.KubernetesConfig, *starter.Node); err != nil {
			return nil, errors.Wrap(err, "setting up certs")
		}

		if err := bs.UpdateNode(*starter.Cfg, *starter.Node, cr); err != nil {
			return nil, errors.Wrap(err, "update node")
		}
	}

	var wg sync.WaitGroup
	if !driver.IsKIC(starter.Cfg.Driver) {
		go configureMounts(&wg)
	}

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
	if starter.ExistingAddons != nil {
		wg.Add(1)
		go addons.Start(&wg, starter.Cfg, starter.ExistingAddons, config.AddonList)
	}

	wg.Add(1)
	go func() {
		rescaleCoreDNS(starter.Cfg, starter.Runner)
		wg.Done()
	}()

	if apiServer {
		// special ops for none , like change minikube directory.
		// multinode super doesn't work on the none driver
		if starter.Cfg.Driver == driver.None && len(starter.Cfg.Nodes) == 1 {
			prepareNone()
		}
	} else {
		// Make sure to use the command runner for the control plane to generate the join token
		cpBs, cpr, err := cluster.ControlPlaneBootstrapper(starter.MachineAPI, starter.Cfg, viper.GetString(cmdcfg.Bootstrapper))
		if err != nil {
			return nil, errors.Wrap(err, "getting control plane bootstrapper")
		}

		joinCmd, err := cpBs.GenerateToken(*starter.Cfg)
		if err != nil {
			return nil, errors.Wrap(err, "generating join token")
		}

		if err = bs.JoinCluster(*starter.Cfg, *starter.Node, joinCmd); err != nil {
			return nil, errors.Wrap(err, "joining cluster")
		}

		cnm, err := cni.New(*starter.Cfg)
		if err != nil {
			return nil, errors.Wrap(err, "cni")
		}

		if err := cnm.Apply(cpr); err != nil {
			return nil, errors.Wrap(err, "cni apply")
		}
	}

	klog.Infof("Will wait %s for node up to ", viper.GetDuration(waitTimeout))
	if err := bs.WaitForNode(*starter.Cfg, *starter.Node, viper.GetDuration(waitTimeout)); err != nil {
		return nil, errors.Wrapf(err, "wait %s for node", viper.GetDuration(waitTimeout))
	}

	klog.Infof("waiting for startup goroutines ...")
	wg.Wait()

	// Write enabled addons to the config before completion
	return kcs, config.Write(viper.GetString(config.ProfileName), starter.Cfg)
}

// Provision provisions the machine/container for the node
func Provision(cc *config.ClusterConfig, n *config.Node, apiServer bool, delOnFail bool) (command.Runner, bool, libmachine.API, *host.Host, error) {
	register.Reg.SetStep(register.StartingNode)
	name := config.MachineName(*cc, *n)
	if apiServer {
		out.Step(style.ThumbsUp, "Starting control plane node {{.name}} in cluster {{.cluster}}", out.V{"name": name, "cluster": cc.Name})
	} else {
		out.Step(style.ThumbsUp, "Starting node {{.name}} in cluster {{.cluster}}", out.V{"name": name, "cluster": cc.Name})
	}

	if driver.IsKIC(cc.Driver) {
		beginDownloadKicBaseImage(&kicGroup, cc, viper.GetBool("download-only"))
	}

	if !driver.BareMetal(cc.Driver) {
		beginCacheKubernetesImages(&cacheGroup, cc.KubernetesConfig.ImageRepository, n.KubernetesVersion, cc.KubernetesConfig.ContainerRuntime)
	}

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot.
	// Hence, SaveProfile must be called before startHost, and again afterwards when we know the IP.
	if err := config.SaveProfile(viper.GetString(config.ProfileName), cc); err != nil {
		return nil, false, nil, nil, errors.Wrap(err, "Failed to save config")
	}

	handleDownloadOnly(&cacheGroup, &kicGroup, n.KubernetesVersion)
	waitDownloadKicBaseImage(&kicGroup)

	return startMachine(cc, n, delOnFail)
}

// ConfigureRuntimes does what needs to happen to get a runtime going.
func configureRuntimes(runner cruntime.CommandRunner, cc config.ClusterConfig, kv semver.Version) cruntime.Manager {
	co := cruntime.Config{
		Type:              cc.KubernetesConfig.ContainerRuntime,
		Runner:            runner,
		ImageRepository:   cc.KubernetesConfig.ImageRepository,
		KubernetesVersion: kv,
	}
	cr, err := cruntime.New(co)
	if err != nil {
		exit.Error(reason.InternalRuntime, "Failed runtime", err)
	}

	disableOthers := true
	if driver.BareMetal(cc.Driver) {
		disableOthers = false
	}

	// Preload is overly invasive for bare metal, and caching is not meaningful.
	// KIC handles preload elsewhere.
	if driver.IsVM(cc.Driver) {
		if err := cr.Preload(cc.KubernetesConfig); err != nil {
			switch err.(type) {
			case *cruntime.ErrISOFeature:
				out.ErrT(style.Tip, "Existing disk is missing new features ({{.error}}). To upgrade, run 'minikube delete'", out.V{"error": err})
			default:
				klog.Warningf("%s preload failed: %v, falling back to caching images", cr.Name(), err)
			}

			if err := machine.CacheImagesForBootstrapper(cc.KubernetesConfig.ImageRepository, cc.KubernetesConfig.KubernetesVersion, viper.GetString(cmdcfg.Bootstrapper)); err != nil {
				exit.Error(reason.RuntimeCache, "Failed to cache images", err)
			}
		}
	}

	err = cr.Enable(disableOthers, forceSystemd())
	if err != nil {
		exit.Error(reason.RuntimeEnable, "Failed to enable container runtime", err)
	}

	return cr
}

func forceSystemd() bool {
	return viper.GetBool("force-systemd") || os.Getenv(constants.MinikubeForceSystemdEnv) == "true"
}

// setupKubeAdm adds any requested files into the VM before Kubernetes is started
func setupKubeAdm(mAPI libmachine.API, cfg config.ClusterConfig, n config.Node, r command.Runner) bootstrapper.Bootstrapper {
	bs, err := cluster.Bootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper), cfg, r)
	if err != nil {
		exit.Error(reason.InternalBootstrapper, "Failed to get bootstrapper", err)
	}
	for _, eo := range config.ExtraOptions {
		out.Infof("{{.extra_option_component_name}}.{{.key}}={{.value}}", out.V{"extra_option_component_name": eo.Component, "key": eo.Key, "value": eo.Value})
	}
	// Loads cached images, generates config files, download binaries
	// update cluster and set up certs

	if err := bs.UpdateCluster(cfg); err != nil {
		exit.Error(reason.KubernetesInstallFailed, "Failed to update cluster", err)
	}

	if err := bs.SetupCerts(cfg.KubernetesConfig, n); err != nil {
		exit.Error(reason.GuestCert, "Failed to setup certs", err)
	}

	return bs
}

func setupKubeconfig(h *host.Host, cc *config.ClusterConfig, n *config.Node, clusterName string) *kubeconfig.Settings {
	addr, err := apiServerURL(*h, *cc, *n)
	if err != nil {
		exit.Message(reason.DrvCPEndpoint, fmt.Sprintf("failed to get API Server URL: %v", err), out.V{"profileArg": fmt.Sprintf("--profile=%s", clusterName)})
	}

	if cc.KubernetesConfig.APIServerName != constants.APIServerName {
		addr = strings.Replace(addr, n.IP, cc.KubernetesConfig.APIServerName, -1)
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

func apiServerURL(h host.Host, cc config.ClusterConfig, n config.Node) (string, error) {
	hostname, _, port, err := driver.ControlPlaneEndpoint(&cc, &n, h.DriverName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://" + net.JoinHostPort(hostname, strconv.Itoa(port))), nil
}

// StartMachine starts a VM
func startMachine(cfg *config.ClusterConfig, node *config.Node, delOnFail bool) (runner command.Runner, preExists bool, machineAPI libmachine.API, host *host.Host, err error) {
	m, err := machine.NewAPIClient()
	if err != nil {
		return runner, preExists, m, host, errors.Wrap(err, "Failed to get machine client")
	}
	host, preExists, err = startHost(m, cfg, node, delOnFail)
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

	// Bypass proxy for minikube's vm host ip
	err = proxy.ExcludeIP(ip)
	if err != nil {
		out.FailureT("Failed to set NO_PROXY Env. Please use `export NO_PROXY=$NO_PROXY,{{.ip}}`.", out.V{"ip": ip})
	}

	return runner, preExists, m, host, err
}

// startHost starts a new minikube host using a VM or None
func startHost(api libmachine.API, cc *config.ClusterConfig, n *config.Node, delOnFail bool) (*host.Host, bool, error) {
	host, exists, err := machine.StartHost(api, cc, n)
	if err == nil {
		return host, exists, nil
	}
	klog.Warningf("error starting host: %v", err)
	// NOTE: People get very cranky if you delete their prexisting VM. Only delete new ones.
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
				out.Step(style.Internet, "Found network options:")
				optSeen = true
			}
			out.Infof("{{.key}}={{.value}}", out.V{"key": k, "value": v})
			ipExcluded := proxy.IsIPExcluded(ip) // Skip warning if minikube ip is already in NO_PROXY
			k = strings.ToUpper(k)               // for http_proxy & https_proxy
			if (k == "HTTP_PROXY" || k == "HTTPS_PROXY") && !ipExcluded && !warnedOnce {
				out.WarningT("You appear to be using a proxy, but your NO_PROXY environment does not include the minikube IP ({{.ip_address}}).", out.V{"ip_address": ip})
				out.Step(style.Documentation, "Please see {{.documentation_url}} for more details", out.V{"documentation_url": "https://minikube.sigs.k8s.io/docs/handbook/vpn_and_proxy/"})
				warnedOnce = true
			}
		}
	}

	if !driver.BareMetal(h.Driver.DriverName()) && !driver.IsKIC(h.Driver.DriverName()) {
		if err := trySSH(h, ip); err != nil {
			return ip, err
		}
	}

	// Non-blocking
	go tryRegistry(r, h.Driver.DriverName(), imageRepository)
	return ip, nil
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
func tryRegistry(r command.Runner, driverName string, imageRepository string) {
	// 2 second timeout. For best results, call tryRegistry in a non-blocking manner.
	opts := []string{"-sS", "-m", "2"}

	proxy := os.Getenv("HTTPS_PROXY")
	if proxy != "" && !strings.HasPrefix(proxy, "localhost") && !strings.HasPrefix(proxy, "127.0") {
		opts = append([]string{"-x", proxy}, opts...)
	}

	if imageRepository == "" {
		imageRepository = images.DefaultKubernetesRepo
	}

	opts = append(opts, fmt.Sprintf("https://%s/", imageRepository))
	if rr, err := r.RunCmd(exec.Command("curl", opts...)); err != nil {
		klog.Warningf("%s failed: %v", rr.Args, err)
		out.WarningT("This {{.type}} is having trouble accessing https://{{.repository}}", out.V{"repository": imageRepository, "type": driver.MachineType(driverName)})
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

// rescaleCoreDNS attempts to reduce coredns replicas from 2 to 1 to improve CPU overhead
// no worries if this doesn't work
func rescaleCoreDNS(cc *config.ClusterConfig, runner command.Runner) {
	kubectl := kapi.KubectlBinaryPath(cc.KubernetesConfig.KubernetesVersion)
	cmd := exec.Command("sudo", "KUBECONFIG=/var/lib/minikube/kubeconfig", kubectl, "scale", "deployment", "--replicas=1", "coredns", "-n=kube-system")
	if _, err := runner.RunCmd(cmd); err != nil {
		klog.Warningf("unable to scale coredns replicas to 1: %v", err)
	} else {
		klog.Infof("successfully scaled coredns replicas to 1")
	}
}
