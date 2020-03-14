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

package cluster

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
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
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/proxy"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
)

const (
	waitTimeout      = "wait-timeout"
	waitUntilHealthy = "wait"
	embedCerts       = "embed-certs"
	keepContext      = "keep-context"
	imageRepository  = "image-repository"
	containerRuntime = "container-runtime"
)

// InitialSetup performs all necessary operations on the initial control plane node when first spinning up a cluster
func InitialSetup(cc config.ClusterConfig, n config.Node, existingAddons map[string]bool) (*kubeconfig.Settings, error) {
	var kicGroup errgroup.Group
	if driver.IsKIC(cc.Driver) {
		BeginDownloadKicArtifacts(&kicGroup)
	}

	var cacheGroup errgroup.Group
	if !driver.BareMetal(cc.Driver) {
		BeginCacheKubernetesImages(&cacheGroup, cc.KubernetesConfig.ImageRepository, n.KubernetesVersion, cc.KubernetesConfig.ContainerRuntime)
	}

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot.
	// Hence, saveConfig must be called before startHost, and again afterwards when we know the IP.
	if err := config.SaveProfile(viper.GetString(config.ProfileName), &cc); err != nil {
		exit.WithError("Failed to save config", err)
	}

	HandleDownloadOnly(&cacheGroup, &kicGroup, n.KubernetesVersion)
	WaitDownloadKicArtifacts(&kicGroup)

	mRunner, preExists, machineAPI, host := StartMachine(&cc, &n)
	defer machineAPI.Close()

	// wait for preloaded tarball to finish downloading before configuring runtimes
	WaitCacheRequiredImages(&cacheGroup)

	sv, err := util.ParseKubernetesVersion(n.KubernetesVersion)
	if err != nil {
		return nil, err
	}

	// configure the runtime (docker, containerd, crio)
	cr := ConfigureRuntimes(mRunner, cc.Driver, cc.KubernetesConfig, sv)

	// Must be written before bootstrap, otherwise health checks may flake due to stale IP
	kubeconfig, err := setupKubeconfig(host, &cc, &n, cc.Name)
	if err != nil {
		exit.WithError("Failed to setup kubeconfig", err)
	}

	// setup kubeadm (must come after setupKubeconfig)
	bs := setupKubeAdm(machineAPI, cc, n)

	// pull images or restart cluster
	out.T(out.Launch, "Launching Kubernetes ... ")
	err = bs.StartCluster(cc)
	if err != nil {
		exit.WithLogEntries("Error starting cluster", err, logs.FindProblems(cr, bs, mRunner))
	}
	//configureMounts()

	if err := CacheAndLoadImagesInConfig(); err != nil {
		out.T(out.FailureType, "Unable to load cached images from config file.")
	}

	// enable addons, both old and new!
	if existingAddons != nil {
		addons.Start(viper.GetString(config.ProfileName), existingAddons, config.AddonList)
	}

	// special ops for none , like change minikube directory.
	// multinode super doesn't work on the none driver
	if cc.Driver == driver.None && len(cc.Nodes) == 1 {
		prepareNone()
	}

	// Skip pre-existing, because we already waited for health
	if viper.GetBool(waitUntilHealthy) && !preExists {
		if err := bs.WaitForNode(cc, n, viper.GetDuration(waitTimeout)); err != nil {
			exit.WithError("Wait failed", err)
		}
	}

	return kubeconfig, nil

}

// ConfigureRuntimes does what needs to happen to get a runtime going.
func ConfigureRuntimes(runner cruntime.CommandRunner, drvName string, k8s config.KubernetesConfig, kv semver.Version) cruntime.Manager {
	co := cruntime.Config{
		Type:   viper.GetString(containerRuntime),
		Runner: runner, ImageRepository: k8s.ImageRepository,
		KubernetesVersion: kv,
	}
	cr, err := cruntime.New(co)
	if err != nil {
		exit.WithError("Failed runtime", err)
	}

	disableOthers := true
	if driver.BareMetal(drvName) {
		disableOthers = false
	}

	// Preload is overly invasive for bare metal, and caching is not meaningful. KIC handled elsewhere.
	if driver.IsVM(drvName) {
		if err := cr.Preload(k8s); err != nil {
			switch err.(type) {
			case *cruntime.ErrISOFeature:
				out.T(out.Tip, "Existing disk is missing new features ({{.error}}). To upgrade, run 'minikube delete'", out.V{"error": err})
			default:
				glog.Warningf("%s preload failed: %v, falling back to caching images", cr.Name(), err)
			}

			if err := machine.CacheImagesForBootstrapper(k8s.ImageRepository, k8s.KubernetesVersion, viper.GetString(cmdcfg.Bootstrapper)); err != nil {
				exit.WithError("Failed to cache images", err)
			}
		}
	}

	err = cr.Enable(disableOthers)
	if err != nil {
		exit.WithError("Failed to enable container runtime", err)
	}

	return cr
}

// setupKubeAdm adds any requested files into the VM before Kubernetes is started
func setupKubeAdm(mAPI libmachine.API, cfg config.ClusterConfig, n config.Node) bootstrapper.Bootstrapper {
	bs, err := Bootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper), cfg, n)
	if err != nil {
		exit.WithError("Failed to get bootstrapper", err)
	}
	for _, eo := range config.ExtraOptions {
		out.T(out.Option, "{{.extra_option_component_name}}.{{.key}}={{.value}}", out.V{"extra_option_component_name": eo.Component, "key": eo.Key, "value": eo.Value})
	}
	// Loads cached images, generates config files, download binaries
	if err := bs.UpdateCluster(cfg); err != nil {
		exit.WithError("Failed to update cluster", err)
	}
	if err := bs.SetupCerts(cfg.KubernetesConfig, n); err != nil {
		exit.WithError("Failed to setup certs", err)
	}
	return bs
}

func setupKubeconfig(h *host.Host, c *config.ClusterConfig, n *config.Node, clusterName string) (*kubeconfig.Settings, error) {
	addr, err := h.Driver.GetURL()
	if err != nil {
		exit.WithError("Failed to get driver URL", err)
	}
	if !driver.IsKIC(h.DriverName) {
		addr = strings.Replace(addr, "tcp://", "https://", -1)
		addr = strings.Replace(addr, ":2376", ":"+strconv.Itoa(n.Port), -1)
	}

	if c.KubernetesConfig.APIServerName != constants.APIServerName {
		addr = strings.Replace(addr, n.IP, c.KubernetesConfig.APIServerName, -1)
	}
	kcs := &kubeconfig.Settings{
		ClusterName:          clusterName,
		ClusterServerAddress: addr,
		ClientCertificate:    localpath.MakeMiniPath("client.crt"),
		ClientKey:            localpath.MakeMiniPath("client.key"),
		CertificateAuthority: localpath.MakeMiniPath("ca.crt"),
		KeepContext:          viper.GetBool(keepContext),
		EmbedCerts:           viper.GetBool(embedCerts),
	}

	kcs.SetPath(kubeconfig.PathFromEnv())
	if err := kubeconfig.Update(kcs); err != nil {
		return kcs, err
	}
	return kcs, nil
}

// StartMachine starts a VM
func StartMachine(cfg *config.ClusterConfig, node *config.Node) (runner command.Runner, preExists bool, machineAPI libmachine.API, host *host.Host) {
	m, err := machine.NewAPIClient()
	if err != nil {
		exit.WithError("Failed to get machine client", err)
	}
	host, preExists = startHost(m, *cfg, *node)
	runner, err = machine.CommandRunner(host)
	if err != nil {
		exit.WithError("Failed to get command runner", err)
	}

	ip := validateNetwork(host, runner)

	// Bypass proxy for minikube's vm host ip
	err = proxy.ExcludeIP(ip)
	if err != nil {
		out.ErrT(out.FailureType, "Failed to set NO_PROXY Env. Please use `export NO_PROXY=$NO_PROXY,{{.ip}}`.", out.V{"ip": ip})
	}

	node.IP = ip
	err = config.SaveNode(cfg, node)
	if err != nil {
		exit.WithError("saving node", err)
	}

	return runner, preExists, m, host
}

// startHost starts a new minikube host using a VM or None
func startHost(api libmachine.API, mc config.ClusterConfig, n config.Node) (*host.Host, bool) {
	host, exists, err := machine.StartHost(api, mc, n)
	if err != nil {
		exit.WithError("Unable to start VM. Please investigate and run 'minikube delete' if possible", err)
	}
	return host, exists
}

// validateNetwork tries to catch network problems as soon as possible
func validateNetwork(h *host.Host, r command.Runner) string {
	ip, err := h.Driver.GetIP()
	if err != nil {
		exit.WithError("Unable to get VM IP address", err)
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
		trySSH(h, ip)
	}

	tryLookup(r)
	tryRegistry(r)
	return ip
}

func trySSH(h *host.Host, ip string) {
	if viper.GetBool("force") {
		return
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

	if err := retry.Expo(dial, time.Second, 13*time.Second); err != nil {
		exit.WithCodeT(exit.IO, `minikube is unable to connect to the VM: {{.error}}

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
}

func tryLookup(r command.Runner) {
	// DNS check
	if rr, err := r.RunCmd(exec.Command("nslookup", "kubernetes.io", "-type=ns")); err != nil {
		glog.Infof("%s failed: %v which might be okay will retry nslookup without query type", rr.Args, err)
		// will try with without query type for ISOs with different busybox versions.
		if _, err = r.RunCmd(exec.Command("nslookup", "kubernetes.io")); err != nil {
			glog.Warningf("nslookup failed: %v", err)
			out.WarningT("Node may be unable to resolve external DNS records")
		}
	}
}
func tryRegistry(r command.Runner) {
	// Try an HTTPS connection to the image repository
	proxy := os.Getenv("HTTPS_PROXY")
	opts := []string{"-sS"}
	if proxy != "" && !strings.HasPrefix(proxy, "localhost") && !strings.HasPrefix(proxy, "127.0") {
		opts = append([]string{"-x", proxy}, opts...)
	}

	repo := viper.GetString(imageRepository)
	if repo == "" {
		repo = images.DefaultKubernetesRepo
	}

	opts = append(opts, fmt.Sprintf("https://%s/", repo))
	if rr, err := r.RunCmd(exec.Command("curl", opts...)); err != nil {
		glog.Warningf("%s failed: %v", rr.Args, err)
		out.WarningT("VM is unable to access {{.repository}}, you may need to configure a proxy or set --image-repository", out.V{"repository": repo})
	}
}

// prepareNone prepares the user and host for the joy of the "none" driver
func prepareNone() {
	out.T(out.StartingNone, "Configuring local host environment ...")
	if viper.GetBool(config.WantNoneDriverWarning) {
		out.T(out.Empty, "")
		out.WarningT("The 'none' driver provides limited isolation and may reduce system security and reliability.")
		out.WarningT("For more information, see:")
		out.T(out.URL, "https://minikube.sigs.k8s.io/docs/reference/drivers/none/")
		out.T(out.Empty, "")
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		home := os.Getenv("HOME")
		out.WarningT("kubectl and minikube configuration will be stored in {{.home_folder}}", out.V{"home_folder": home})
		out.WarningT("To use kubectl or minikube commands as your own user, you may need to relocate them. For example, to overwrite your own settings, run:")

		out.T(out.Empty, "")
		out.T(out.Command, "sudo mv {{.home_folder}}/.kube {{.home_folder}}/.minikube $HOME", out.V{"home_folder": home})
		out.T(out.Command, "sudo chown -R $USER $HOME/.kube $HOME/.minikube")
		out.T(out.Empty, "")

		out.T(out.Tip, "This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true")
	}

	if err := util.MaybeChownDirRecursiveToMinikubeUser(localpath.MiniPath()); err != nil {
		exit.WithCodeT(exit.Permissions, "Failed to change permissions for {{.minikube_dir_path}}: {{.error}}", out.V{"minikube_dir_path": localpath.MiniPath(), "error": err})
	}
}
