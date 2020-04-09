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
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/cluster"
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
	"k8s.io/minikube/pkg/minikube/proxy"
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

// Start spins up a guest and starts the kubernetes node.
func Start(starter Starter, apiServer bool) (*kubeconfig.Settings, error) {
	// wait for preloaded tarball to finish downloading before configuring runtimes
	waitCacheRequiredImages(&cacheGroup)

	sv, err := util.ParseKubernetesVersion(starter.Node.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse kubernetes version")
	}

	// configure the runtime (docker, containerd, crio)
	cr := configureRuntimes(starter.Runner, *starter.Cfg, sv)
	showVersionInfo(starter.Node.KubernetesVersion, cr)

	// ssh should be set up by now
	// switch to using ssh runner since it is faster
	if driver.IsKIC(starter.Cfg.Driver) {
		sshRunner, err := machine.SSHRunner(starter.Host)
		if err != nil {
			glog.Infof("error getting ssh runner: %v", err)
		} else {
			glog.Infof("Using ssh runner for kic...")
			starter.Runner = sshRunner
		}
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
		bs = setupKubeAdm(starter.MachineAPI, *starter.Cfg, *starter.Node)
		err = bs.StartCluster(*starter.Cfg)

		if err != nil {
			out.LogEntries("Error starting cluster", err, logs.FindProblems(cr, bs, *starter.Cfg, starter.Runner))
			return nil, err
		}

		// write the kubeconfig to the file system after everything required (like certs) are created by the bootstrapper
		if err := kubeconfig.Update(kcs); err != nil {
			return nil, errors.Wrap(err, "Failed to update kubeconfig file.")
		}
	} else {
		bs, err = cluster.Bootstrapper(starter.MachineAPI, viper.GetString(cmdcfg.Bootstrapper), *starter.Cfg, *starter.Node)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get bootstrapper")
		}

		if err = bs.SetupCerts(starter.Cfg.KubernetesConfig, *starter.Node); err != nil {
			return nil, errors.Wrap(err, "setting up certs")
		}
	}

	var wg sync.WaitGroup
	go configureMounts(&wg)

	wg.Add(1)
	go func() {
		if err := CacheAndLoadImagesInConfig(); err != nil {
			out.FailureT("Unable to load cached images from config file: {{error}}", out.V{"error": err})
		}
		wg.Done()
	}()

	// enable addons, both old and new!
	if starter.ExistingAddons != nil {
		go addons.Start(&wg, starter.Cfg, starter.ExistingAddons, config.AddonList)
	}

	if apiServer {
		// special ops for none , like change minikube directory.
		// multinode super doesn't work on the none driver
		if starter.Cfg.Driver == driver.None && len(starter.Cfg.Nodes) == 1 {
			prepareNone()
		}

		// Skip pre-existing, because we already waited for health
		if kverify.ShouldWait(starter.Cfg.VerifyComponents) && !starter.PreExists {
			if err := bs.WaitForNode(*starter.Cfg, *starter.Node, viper.GetDuration(waitTimeout)); err != nil {
				return nil, errors.Wrap(err, "Wait failed")
			}
		}
	} else {
		if err := bs.UpdateNode(*starter.Cfg, *starter.Node, cr); err != nil {
			return nil, errors.Wrap(err, "Updating node")
		}

		cp, err := config.PrimaryControlPlane(starter.Cfg)
		if err != nil {
			return nil, errors.Wrap(err, "Getting primary control plane")
		}
		cpBs, err := cluster.Bootstrapper(starter.MachineAPI, viper.GetString(cmdcfg.Bootstrapper), *starter.Cfg, cp)
		if err != nil {
			return nil, errors.Wrap(err, "Getting bootstrapper")
		}

		joinCmd, err := cpBs.GenerateToken(*starter.Cfg)
		if err != nil {
			return nil, errors.Wrap(err, "generating join token")
		}

		if err = bs.JoinCluster(*starter.Cfg, *starter.Node, joinCmd); err != nil {
			return nil, errors.Wrap(err, "joining cluster")
		}
	}

	wg.Wait()

	// Write enabled addons to the config before completion
	return kcs, config.Write(viper.GetString(config.ProfileName), starter.Cfg)
}

// Provision provisions the machine/container for the node
func Provision(cc *config.ClusterConfig, n *config.Node, apiServer bool) (command.Runner, bool, libmachine.API, *host.Host, error) {

	name := driver.MachineName(*cc, *n)
	if apiServer {
		out.T(out.ThumbsUp, "Starting control plane node {{.name}} in cluster {{.cluster}}", out.V{"name": name, "cluster": cc.Name})
	} else {
		out.T(out.ThumbsUp, "Starting node {{.name}} in cluster {{.cluster}}", out.V{"name": name, "cluster": cc.Name})
	}

	if driver.IsKIC(cc.Driver) {
		beginDownloadKicArtifacts(&kicGroup)
	}

	if !driver.BareMetal(cc.Driver) {
		beginCacheKubernetesImages(&cacheGroup, cc.KubernetesConfig.ImageRepository, n.KubernetesVersion, cc.KubernetesConfig.ContainerRuntime)
	}

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot.
	// Hence, saveConfig must be called before startHost, and again afterwards when we know the IP.
	if err := config.SaveProfile(viper.GetString(config.ProfileName), cc); err != nil {
		return nil, false, nil, nil, errors.Wrap(err, "Failed to save config")
	}

	handleDownloadOnly(&cacheGroup, &kicGroup, n.KubernetesVersion)
	waitDownloadKicArtifacts(&kicGroup)

	return startMachine(cc, n)

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
		exit.WithError("Failed runtime", err)
	}

	disableOthers := true
	if driver.BareMetal(cc.Driver) {
		disableOthers = false
	}

	// Preload is overly invasive for bare metal, and caching is not meaningful. KIC handled elsewhere.
	if driver.IsVM(cc.Driver) {
		if err := cr.Preload(cc.KubernetesConfig); err != nil {
			switch err.(type) {
			case *cruntime.ErrISOFeature:
				out.ErrT(out.Tip, "Existing disk is missing new features ({{.error}}). To upgrade, run 'minikube delete'", out.V{"error": err})
			default:
				glog.Warningf("%s preload failed: %v, falling back to caching images", cr.Name(), err)
			}

			if err := machine.CacheImagesForBootstrapper(cc.KubernetesConfig.ImageRepository, cc.KubernetesConfig.KubernetesVersion, viper.GetString(cmdcfg.Bootstrapper)); err != nil {
				exit.WithError("Failed to cache images", err)
			}
		}
	}

	err = cr.Enable(disableOthers)
	if err != nil {
		debug.PrintStack()
		exit.WithError("Failed to enable container runtime", err)
	}

	return cr
}

// setupKubeAdm adds any requested files into the VM before Kubernetes is started
func setupKubeAdm(mAPI libmachine.API, cfg config.ClusterConfig, n config.Node) bootstrapper.Bootstrapper {
	bs, err := cluster.Bootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper), cfg, n)
	if err != nil {
		exit.WithError("Failed to get bootstrapper", err)
	}
	for _, eo := range config.ExtraOptions {
		out.T(out.Option, "{{.extra_option_component_name}}.{{.key}}={{.value}}", out.V{"extra_option_component_name": eo.Component, "key": eo.Key, "value": eo.Value})
	}
	// Loads cached images, generates config files, download binaries
	// update cluster and set up certs in parallel
	var parallel sync.WaitGroup
	parallel.Add(2)
	go func() {
		if err := bs.UpdateCluster(cfg); err != nil {
			exit.WithError("Failed to update cluster", err)
		}
		parallel.Done()
	}()

	go func() {
		if err := bs.SetupCerts(cfg.KubernetesConfig, n); err != nil {
			exit.WithError("Failed to setup certs", err)
		}
		parallel.Done()
	}()

	parallel.Wait()
	return bs
}

func setupKubeconfig(h *host.Host, cc *config.ClusterConfig, n *config.Node, clusterName string) *kubeconfig.Settings {
	addr, err := apiServerURL(*h, *cc, *n)
	if err != nil {
		exit.WithError("Failed to get API Server URL", err)
	}

	if cc.KubernetesConfig.APIServerName != constants.APIServerName {
		addr = strings.Replace(addr, n.IP, cc.KubernetesConfig.APIServerName, -1)
	}
	kcs := &kubeconfig.Settings{
		ClusterName:          clusterName,
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
	hostname, _, port, err := driver.ControlPaneEndpoint(&cc, &n, h.DriverName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://" + net.JoinHostPort(hostname, strconv.Itoa(port))), nil
}

// StartMachine starts a VM
func startMachine(cfg *config.ClusterConfig, node *config.Node) (runner command.Runner, preExists bool, machineAPI libmachine.API, host *host.Host, err error) {
	m, err := machine.NewAPIClient()
	if err != nil {
		return runner, preExists, m, host, errors.Wrap(err, "Failed to get machine client")
	}
	host, preExists, err = startHost(m, *cfg, *node)
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

	// Save IP to config file for subsequent use
	node.IP = ip
	err = config.SaveNode(cfg, node)
	if err != nil {
		return runner, preExists, m, host, errors.Wrap(err, "saving node")
	}

	return runner, preExists, m, host, err
}

// startHost starts a new minikube host using a VM or None
func startHost(api libmachine.API, cc config.ClusterConfig, n config.Node) (*host.Host, bool, error) {
	host, exists, err := machine.StartHost(api, cc, n)
	if err == nil {
		return host, exists, nil
	}
	out.ErrT(out.Embarrassed, "StartHost failed, but will try again: {{.error}}", out.V{"error": err})

	// NOTE: People get very cranky if you delete their prexisting VM. Only delete new ones.
	if !exists {
		err := machine.DeleteHost(api, driver.MachineName(cc, n))
		if err != nil {
			glog.Warningf("delete host: %v", err)
		}
	}

	// Try again, but just once to avoid making the logs overly confusing
	time.Sleep(5 * time.Second)

	host, exists, err = machine.StartHost(api, cc, n)
	if err == nil {
		return host, exists, nil
	}

	// Don't use host.Driver to avoid nil pointer deref
	drv := cc.Driver
	out.ErrT(out.Sad, `Failed to start {{.driver}} {{.driver_type}}. "{{.cmd}}" may fix it: {{.error}}`, out.V{"driver": drv, "driver_type": driver.MachineType(drv), "cmd": mustload.ExampleCmd(cc.Name, "start"), "error": err})
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
				out.T(out.Internet, "Found network options:")
				optSeen = true
			}
			out.T(out.Option, "{{.key}}={{.value}}", out.V{"key": k, "value": v})
			ipExcluded := proxy.IsIPExcluded(ip) // Skip warning if minikube ip is already in NO_PROXY
			k = strings.ToUpper(k)               // for http_proxy & https_proxy
			if (k == "HTTP_PROXY" || k == "HTTPS_PROXY") && !ipExcluded && !warnedOnce {
				out.WarningT("You appear to be using a proxy, but your NO_PROXY environment does not include the minikube IP ({{.ip_address}}). Please see {{.documentation_url}} for more details", out.V{"ip_address": ip, "documentation_url": "https://minikube.sigs.k8s.io/docs/reference/networking/proxy/"})
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
			out.WarningT("Unable to verify SSH connectivity: {{.error}}. Will retry...", out.V{"error": err})
			return err
		}
		_ = conn.Close()
		return nil
	}

	err := retry.Expo(dial, time.Second, 13*time.Second)
	if err != nil {
		out.ErrT(out.FailureType, `minikube is unable to connect to the VM: {{.error}}

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
		glog.Warningf("%s failed: %v", rr.Args, err)
		out.WarningT("This {{.type}} is having trouble accessing https://{{.repository}}", out.V{"repository": imageRepository, "type": driver.MachineType(driverName)})
		out.ErrT(out.Tip, "To pull new external images, you may need to configure a proxy: https://minikube.sigs.k8s.io/docs/reference/networking/proxy/")
	}
}

// prepareNone prepares the user and host for the joy of the "none" driver
func prepareNone() {
	out.T(out.StartingNone, "Configuring local host environment ...")
	if viper.GetBool(config.WantNoneDriverWarning) {
		out.ErrT(out.Empty, "")
		out.WarningT("The 'none' driver is designed for experts who need to integrate with an existing VM")
		out.ErrT(out.Tip, "Most users should use the newer 'docker' driver instead, which does not require root!")
		out.ErrT(out.Documentation, "For more information, see: https://minikube.sigs.k8s.io/docs/reference/drivers/none/")
		out.ErrT(out.Empty, "")
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		home := os.Getenv("HOME")
		out.WarningT("kubectl and minikube configuration will be stored in {{.home_folder}}", out.V{"home_folder": home})
		out.WarningT("To use kubectl or minikube commands as your own user, you may need to relocate them. For example, to overwrite your own settings, run:")

		out.ErrT(out.Empty, "")
		out.ErrT(out.Command, "sudo mv {{.home_folder}}/.kube {{.home_folder}}/.minikube $HOME", out.V{"home_folder": home})
		out.ErrT(out.Command, "sudo chown -R $USER $HOME/.kube $HOME/.minikube")
		out.ErrT(out.Empty, "")

		out.ErrT(out.Tip, "This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true")
	}

	if err := util.MaybeChownDirRecursiveToMinikubeUser(localpath.MiniPath()); err != nil {
		exit.WithCodeT(exit.Permissions, "Failed to change permissions for {{.minikube_dir_path}}: {{.error}}", out.V{"minikube_dir_path": localpath.MiniPath(), "error": err})
	}
}
