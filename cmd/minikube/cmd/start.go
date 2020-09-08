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

package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/golang/glog"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	gopshost "github.com/shirou/gopsutil/host"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"

	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/translate"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/version"
)

var (
	registryMirror   []string
	insecureRegistry []string
	apiServerNames   []string
	apiServerIPs     []net.IP
)

func init() {
	initMinikubeFlags()
	initKubernetesFlags()
	initDriverFlags()
	initNetworkingFlags()
	if err := viper.BindPFlags(startCmd.Flags()); err != nil {
		exit.Error(reason.InternalBindFlags, "unable to bind flags", err)
	}
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a local Kubernetes cluster",
	Long:  "Starts a local Kubernetes cluster",
	Run:   runStart,
}

// platform generates a user-readable platform message
func platform() string {
	var s strings.Builder

	// Show the distro version if possible
	hi, err := gopshost.Info()
	if err == nil {
		s.WriteString(fmt.Sprintf("%s %s", strings.Title(hi.Platform), hi.PlatformVersion))
		glog.Infof("hostinfo: %+v", hi)
	} else {
		glog.Warningf("gopshost.Info returned error: %v", err)
		s.WriteString(runtime.GOOS)
	}

	vsys, vrole, err := gopshost.Virtualization()
	if err != nil {
		glog.Warningf("gopshost.Virtualization returned error: %v", err)
	} else {
		glog.Infof("virtualization: %s %s", vsys, vrole)
	}

	// This environment is exotic, let's output a bit more.
	if vrole == "guest" || runtime.GOARCH != "amd64" {
		if vsys != "" {
			s.WriteString(fmt.Sprintf(" (%s/%s)", vsys, runtime.GOARCH))
		} else {
			s.WriteString(fmt.Sprintf(" (%s)", runtime.GOARCH))
		}
	}
	return s.String()
}

// runStart handles the executes the flow of "minikube start"
func runStart(cmd *cobra.Command, args []string) {
	register.SetEventLogPath(localpath.EventLog(ClusterFlagValue()))

	out.SetJSON(viper.GetString(startOutput) == "json")
	displayVersion(version.GetVersion())

	// No need to do the update check if no one is going to see it
	if !viper.GetBool(interactive) || !viper.GetBool(dryRun) {
		// Avoid blocking execution on optional HTTP fetches
		go notify.MaybePrintUpdateTextFromGithub()
	}

	displayEnviron(os.Environ())
	if viper.GetBool(force) {
		out.WarningT("minikube skips various validations when --force is supplied; this may lead to unexpected behavior")
	}

	// if --registry-mirror specified when run minikube start,
	// take arg precedence over MINIKUBE_REGISTRY_MIRROR
	// actually this is a hack, because viper 1.0.0 can assign env to variable if StringSliceVar
	// and i can't update it to 1.4.0, it affects too much code
	// other types (like String, Bool) of flag works, so imageRepository, imageMirrorCountry
	// can be configured as MINIKUBE_IMAGE_REPOSITORY and IMAGE_MIRROR_COUNTRY
	// this should be updated to documentation
	if len(registryMirror) == 0 {
		registryMirror = viper.GetStringSlice("registry_mirror")
	}

	if !config.ProfileNameValid(ClusterFlagValue()) {
		out.WarningT("Profile name '{{.name}}' is not valid", out.V{"name": ClusterFlagValue()})
		exit.Message(reason.Usage, "Only alphanumeric and dashes '-' are permitted. Minimum 1 character, starting with alphanumeric.")
	}
	existing, err := config.Load(ClusterFlagValue())
	if err != nil && !config.IsNotExist(err) {
		exit.Message(reason.HostConfigLoad, "Unable to load config: {{.error}}", out.V{"error": err})
	}

	if existing != nil {
		upgradeExistingConfig(existing)
	}

	validateSpecifiedDriver(existing)
	validateKubernetesVersion(existing)
	ds, alts, specified := selectDriver(existing)
	starter, err := provisionWithDriver(cmd, ds, existing)
	if err != nil {
		node.ExitIfFatal(err)
		machine.MaybeDisplayAdvice(err, ds.Name)
		if specified {
			// If the user specified a driver, don't fallback to anything else
			exit.Error(reason.GuestProvision, "error provisioning host", err)
		} else {
			success := false
			// Walk down the rest of the options
			for _, alt := range alts {
				out.WarningT("Startup with {{.old_driver}} driver failed, trying with alternate driver {{.new_driver}}: {{.error}}", out.V{"old_driver": ds.Name, "new_driver": alt.Name, "error": err})
				ds = alt
				// Delete the existing cluster and try again with the next driver on the list
				profile, err := config.LoadProfile(ClusterFlagValue())
				if err != nil {
					glog.Warningf("%s profile does not exist, trying anyways.", ClusterFlagValue())
				}

				err = deleteProfile(profile)
				if err != nil {
					out.WarningT("Failed to delete cluster {{.name}}, proceeding with retry anyway.", out.V{"name": ClusterFlagValue()})
				}
				starter, err = provisionWithDriver(cmd, ds, existing)
				if err != nil {
					continue
				} else {
					// Success!
					success = true
					break
				}
			}
			if !success {
				exit.Error(reason.GuestProvision, "error provisioning host", err)
			}
		}
	}

	if existing != nil && driver.IsKIC(existing.Driver) {
		if viper.GetBool(createMount) {
			mount := viper.GetString(mountString)
			if len(existing.ContainerVolumeMounts) != 1 || existing.ContainerVolumeMounts[0] != mount {
				exit.Message(reason.GuestMountConflict, "Sorry, {{.driver}} does not allow mounts to be changed after container creation (previous mount: '{{.old}}', new mount: '{{.new}})'", out.V{
					"driver": existing.Driver,
					"new":    mount,
					"old":    existing.ContainerVolumeMounts[0],
				})
			}
		}

		if existing.KubernetesConfig.ContainerRuntime == "crio" {
			// Stop and start again if it's crio because it's broken above v1.17.3
			out.WarningT("Due to issues with CRI-O post v1.17.3, we need to restart your cluster.")
			out.WarningT("See details at https://github.com/kubernetes/minikube/issues/8861")
			stopProfile(existing.Name)
			starter, err = provisionWithDriver(cmd, ds, existing)
			if err != nil {
				exit.Error(reason.GuestProvision, "error provisioning host", err)
			}
		}
	}

	kubeconfig, err := startWithDriver(cmd, starter, existing)
	if err != nil {
		node.ExitIfFatal(err)
		exit.Error(reason.GuestStart, "failed to start node", err)
	}

	if err := showKubectlInfo(kubeconfig, starter.Node.KubernetesVersion, starter.Cfg.Name); err != nil {
		glog.Errorf("kubectl info: %v", err)
	}
}

func provisionWithDriver(cmd *cobra.Command, ds registry.DriverState, existing *config.ClusterConfig) (node.Starter, error) {
	driverName := ds.Name
	glog.Infof("selected driver: %s", driverName)
	validateDriver(ds, existing)
	err := autoSetDriverOptions(cmd, driverName)
	if err != nil {
		glog.Errorf("Error autoSetOptions : %v", err)
	}

	validateFlags(cmd, driverName)
	validateUser(driverName)

	// Download & update the driver, even in --download-only mode
	if !viper.GetBool(dryRun) {
		updateDriver(driverName)
	}

	k8sVersion := getKubernetesVersion(existing)
	cc, n, err := generateClusterConfig(cmd, existing, k8sVersion, driverName)
	if err != nil {
		return node.Starter{}, errors.Wrap(err, "Failed to generate config")
	}

	// This is about as far as we can go without overwriting config files
	if viper.GetBool(dryRun) {
		out.T(style.DryRun, `dry-run validation complete!`)
		os.Exit(0)
	}

	if driver.IsVM(driverName) {
		url, err := download.ISO(viper.GetStringSlice(isoURL), cmd.Flags().Changed(isoURL))
		if err != nil {
			return node.Starter{}, errors.Wrap(err, "Failed to cache ISO")
		}
		cc.MinikubeISO = url
	}

	var existingAddons map[string]bool
	if viper.GetBool(installAddons) {
		existingAddons = map[string]bool{}
		if existing != nil && existing.Addons != nil {
			existingAddons = existing.Addons
		}
	}

	if viper.GetBool(nativeSSH) {
		ssh.SetDefaultClient(ssh.Native)
	} else {
		ssh.SetDefaultClient(ssh.External)
	}

	mRunner, preExists, mAPI, host, err := node.Provision(&cc, &n, true, viper.GetBool(deleteOnFailure))
	if err != nil {
		return node.Starter{}, err
	}

	return node.Starter{
		Runner:         mRunner,
		PreExists:      preExists,
		MachineAPI:     mAPI,
		Host:           host,
		ExistingAddons: existingAddons,
		Cfg:            &cc,
		Node:           &n,
	}, nil
}

func startWithDriver(cmd *cobra.Command, starter node.Starter, existing *config.ClusterConfig) (*kubeconfig.Settings, error) {
	kubeconfig, err := node.Start(starter, true)
	if err != nil {
		kubeconfig, err = maybeDeleteAndRetry(cmd, *starter.Cfg, *starter.Node, starter.ExistingAddons, err)
		if err != nil {
			return nil, err
		}
	}

	numNodes := viper.GetInt(nodes)
	if existing != nil {
		if numNodes > 1 {
			// We ignore the --nodes parameter if we're restarting an existing cluster
			out.WarningT(`The cluster {{.cluster}} already exists which means the --nodes parameter will be ignored. Use "minikube node add" to add nodes to an existing cluster.`, out.V{"cluster": existing.Name})
		}
		numNodes = len(existing.Nodes)
	}
	if numNodes > 1 {
		if driver.BareMetal(starter.Cfg.Driver) {
			exit.Message(reason.DrvUnsupportedMulti, "The none driver is not compatible with multi-node clusters.")
		} else {
			// Only warn users on first start.
			if existing == nil {
				out.Ln("")
				warnAboutMultiNode()

				for i := 1; i < numNodes; i++ {
					nodeName := node.Name(i + 1)
					n := config.Node{
						Name:              nodeName,
						Worker:            true,
						ControlPlane:      false,
						KubernetesVersion: starter.Cfg.KubernetesConfig.KubernetesVersion,
					}
					out.Ln("") // extra newline for clarity on the command line
					err := node.Add(starter.Cfg, n, viper.GetBool(deleteOnFailure))
					if err != nil {
						return nil, errors.Wrap(err, "adding node")
					}
				}
			} else {
				for _, n := range existing.Nodes {
					if !n.ControlPlane {
						err := node.Add(starter.Cfg, n, viper.GetBool(deleteOnFailure))
						if err != nil {
							return nil, errors.Wrap(err, "adding node")
						}
					}
				}
			}
		}
	}

	return kubeconfig, nil
}

func warnAboutMultiNode() {
	out.WarningT("Multi-node clusters are currently experimental and might exhibit unintended behavior.")
	out.T(style.Documentation, "To track progress on multi-node clusters, see https://github.com/kubernetes/minikube/issues/7538.")
}

func updateDriver(driverName string) {
	v, err := version.GetSemverVersion()
	if err != nil {
		out.WarningT("Error parsing minikube version: {{.error}}", out.V{"error": err})
	} else if err := driver.InstallOrUpdate(driverName, localpath.MakeMiniPath("bin"), v, viper.GetBool(interactive), viper.GetBool(autoUpdate)); err != nil {
		out.WarningT("Unable to update {{.driver}} driver: {{.error}}", out.V{"driver": driverName, "error": err})
	}
}

func displayVersion(version string) {
	prefix := ""
	if ClusterFlagValue() != constants.DefaultClusterName {
		prefix = fmt.Sprintf("[%s] ", ClusterFlagValue())
	}

	register.Reg.SetStep(register.InitialSetup)
	out.T(style.Happy, "{{.prefix}}minikube {{.version}} on {{.platform}}", out.V{"prefix": prefix, "version": version, "platform": platform()})
}

// displayEnviron makes the user aware of environment variables that will affect how minikube operates
func displayEnviron(env []string) {
	for _, kv := range env {
		bits := strings.SplitN(kv, "=", 2)
		k := bits[0]
		v := bits[1]
		if strings.HasPrefix(k, "MINIKUBE_") || k == constants.KubeconfigEnvVar {
			out.Infof("{{.key}}={{.value}}", out.V{"key": k, "value": v})
		}
	}
}

func showKubectlInfo(kcs *kubeconfig.Settings, k8sVersion string, machineName string) error {
	// To be shown at the end, regardless of exit path
	defer func() {
		register.Reg.SetStep(register.Done)
		if kcs.KeepContext {
			out.T(style.Kubectl, "To connect to this cluster, use:  --context={{.name}}", out.V{"name": kcs.ClusterName})
		} else {
			out.T(style.Ready, `Done! kubectl is now configured to use "{{.name}}" by default`, out.V{"name": machineName})
		}
	}()

	path, err := exec.LookPath("kubectl")
	if err != nil {
		out.T(style.Tip, "kubectl not found. If you need it, try: 'minikube kubectl -- get pods -A'")
		return nil
	}

	gitVersion, err := kubectlVersion(path)
	if err != nil {
		return err
	}

	client, err := semver.Make(strings.TrimPrefix(gitVersion, version.VersionPrefix))
	if err != nil {
		return errors.Wrap(err, "client semver")
	}

	cluster := semver.MustParse(strings.TrimPrefix(k8sVersion, version.VersionPrefix))
	minorSkew := int(math.Abs(float64(int(client.Minor) - int(cluster.Minor))))
	glog.Infof("kubectl: %s, cluster: %s (minor skew: %d)", client, cluster, minorSkew)

	if client.Major != cluster.Major || minorSkew > 1 {
		out.Ln("")
		out.WarningT("{{.path}} is version {{.client_version}}, which may have incompatibilites with Kubernetes {{.cluster_version}}.",
			out.V{"path": path, "client_version": client, "cluster_version": cluster})
		out.T(style.Tip, "Want kubectl {{.version}}? Try 'minikube kubectl -- get pods -A'", out.V{"version": k8sVersion})
	}
	return nil
}

func maybeDeleteAndRetry(cmd *cobra.Command, existing config.ClusterConfig, n config.Node, existingAddons map[string]bool, originalErr error) (*kubeconfig.Settings, error) {
	if viper.GetBool(deleteOnFailure) {
		out.WarningT("Node {{.name}} failed to start, deleting and trying again.", out.V{"name": n.Name})
		// Start failed, delete the cluster and try again
		profile, err := config.LoadProfile(existing.Name)
		if err != nil {
			out.ErrT(style.Meh, `"{{.name}}" profile does not exist, trying anyways.`, out.V{"name": existing.Name})
		}

		err = deleteProfile(profile)
		if err != nil {
			out.WarningT("Failed to delete cluster {{.name}}, proceeding with retry anyway.", out.V{"name": existing.Name})
		}

		// Re-generate the cluster config, just in case the failure was related to an old config format
		cc := updateExistingConfigFromFlags(cmd, &existing)
		var kubeconfig *kubeconfig.Settings
		for _, n := range cc.Nodes {
			r, p, m, h, err := node.Provision(&cc, &n, n.ControlPlane, false)
			s := node.Starter{
				Runner:         r,
				PreExists:      p,
				MachineAPI:     m,
				Host:           h,
				Cfg:            &cc,
				Node:           &n,
				ExistingAddons: existingAddons,
			}
			if err != nil {
				// Ok we failed again, let's bail
				return nil, err
			}

			k, err := node.Start(s, n.ControlPlane)
			if n.ControlPlane {
				kubeconfig = k
			}
			if err != nil {
				// Ok we failed again, let's bail
				return nil, err
			}
		}
		return kubeconfig, nil
	}
	// Don't delete the cluster unless they ask
	return nil, originalErr
}

func kubectlVersion(path string) (string, error) {
	j, err := exec.Command(path, "version", "--client", "--output=json").Output()
	if err != nil {
		// really old Kubernetes clients did not have the --output parameter
		b, err := exec.Command(path, "version", "--client", "--short").Output()
		if err != nil {
			return "", errors.Wrap(err, "exec")
		}
		s := strings.TrimSpace(string(b))
		return strings.Replace(s, "Client Version: ", "", 1), nil
	}

	cv := struct {
		ClientVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"clientVersion"`
	}{}
	err = json.Unmarshal(j, &cv)
	if err != nil {
		return "", errors.Wrap(err, "unmarshal")
	}

	return cv.ClientVersion.GitVersion, nil
}

func selectDriver(existing *config.ClusterConfig) (registry.DriverState, []registry.DriverState, bool) {
	// Technically unrelated, but important to perform before detection
	driver.SetLibvirtURI(viper.GetString(kvmQemuURI))
	register.Reg.SetStep(register.SelectingDriver)
	// By default, the driver is whatever we used last time
	if existing != nil {
		old := hostDriver(existing)
		ds := driver.Status(old)
		out.T(style.Sparkle, `Using the {{.driver}} driver based on existing profile`, out.V{"driver": ds.String()})
		return ds, nil, true
	}

	// Default to looking at the new driver parameter
	if d := viper.GetString("driver"); d != "" {
		if vmd := viper.GetString("vm-driver"); vmd != "" {
			// Output a warning
			warning := `Both driver={{.driver}} and vm-driver={{.vmd}} have been set.

    Since vm-driver is deprecated, minikube will default to driver={{.driver}}.

    If vm-driver is set in the global config, please run "minikube config unset vm-driver" to resolve this warning.
			`
			out.WarningT(warning, out.V{"driver": d, "vmd": vmd})
		}
		ds := driver.Status(d)
		if ds.Name == "" {
			exit.Message(reason.DrvUnsupportedOS, "The driver '{{.driver}}' is not supported on {{.os}}", out.V{"driver": d, "os": runtime.GOOS})
		}
		out.T(style.Sparkle, `Using the {{.driver}} driver based on user configuration`, out.V{"driver": ds.String()})
		return ds, nil, true
	}

	// Fallback to old driver parameter
	if d := viper.GetString("vm-driver"); d != "" {
		ds := driver.Status(viper.GetString("vm-driver"))
		if ds.Name == "" {
			exit.Message(reason.DrvUnsupportedOS, "The driver '{{.driver}}' is not supported on {{.os}}", out.V{"driver": d, "os": runtime.GOOS})
		}
		out.T(style.Sparkle, `Using the {{.driver}} driver based on user configuration`, out.V{"driver": ds.String()})
		return ds, nil, true
	}

	choices := driver.Choices(viper.GetBool("vm"))
	pick, alts, rejects := driver.Suggest(choices)
	if pick.Name == "" {
		out.T(style.ThumbsDown, "Unable to pick a default driver. Here is what was considered, in preference order:")
		for _, r := range rejects {
			out.Infof("{{ .name }}: {{ .rejection }}", out.V{"name": r.Name, "rejection": r.Rejection})
		}
		exit.Message(reason.DrvNotDetected, "No possible driver was detected. Try specifying --driver, or see https://minikube.sigs.k8s.io/docs/start/")
	}

	if len(alts) > 1 {
		altNames := []string{}
		for _, a := range alts {
			altNames = append(altNames, a.String())
		}
		out.T(style.Sparkle, `Automatically selected the {{.driver}} driver. Other choices: {{.alternates}}`, out.V{"driver": pick.Name, "alternates": strings.Join(altNames, ", ")})
	} else {
		out.T(style.Sparkle, `Automatically selected the {{.driver}} driver`, out.V{"driver": pick.String()})
	}
	return pick, alts, false
}

// hostDriver returns the actual driver used by a libmachine host, which can differ from our config
func hostDriver(existing *config.ClusterConfig) string {
	if existing == nil {
		return ""
	}
	api, err := machine.NewAPIClient()
	if err != nil {
		glog.Warningf("selectDriver NewAPIClient: %v", err)
		return existing.Driver
	}

	cp, err := config.PrimaryControlPlane(existing)
	if err != nil {
		glog.Warningf("Unable to get control plane from existing config: %v", err)
		return existing.Driver
	}
	machineName := driver.MachineName(*existing, cp)
	h, err := api.Load(machineName)
	if err != nil {
		glog.Warningf("api.Load failed for %s: %v", machineName, err)
		if existing.VMDriver != "" {
			return existing.VMDriver
		}
		return existing.Driver
	}

	return h.Driver.DriverName()
}

// validateSpecifiedDriver makes sure that if a user has passed in a driver
// it matches the existing cluster if there is one
func validateSpecifiedDriver(existing *config.ClusterConfig) {
	if existing == nil {
		return
	}

	var requested string
	if d := viper.GetString("driver"); d != "" {
		requested = d
	} else if d := viper.GetString("vm-driver"); d != "" {
		requested = d
	}

	// Neither --vm-driver or --driver was specified
	if requested == "" {
		return
	}

	old := hostDriver(existing)
	if requested == old {
		return
	}

	exit.Advice(
		reason.GuestDrvMismatch,
		`The existing "{{.name}}" cluster was created using the "{{.old}}" driver, which is incompatible with requested "{{.new}}" driver.`,
		"Delete the existing '{{.name}}' cluster using: '{{.delcommand}}', or start the existing '{{.name}}' cluster using: '{{.command}} --driver={{.old}}'",
		out.V{
			"name":       existing.Name,
			"new":        requested,
			"old":        old,
			"command":    mustload.ExampleCmd(existing.Name, "start"),
			"delcommand": mustload.ExampleCmd(existing.Name, "delete"),
		},
	)
}

// validateDriver validates that the selected driver appears sane, exits if not
func validateDriver(ds registry.DriverState, existing *config.ClusterConfig) {
	name := ds.Name
	glog.Infof("validating driver %q against %+v", name, existing)
	if !driver.Supported(name) {
		exit.Message(reason.DrvUnsupportedOS, "The driver '{{.driver}}' is not supported on {{.os}}", out.V{"driver": name, "os": runtime.GOOS})
	}

	// if we are only downloading artifacts for a driver, we can stop validation here
	if viper.GetBool("download-only") {
		return
	}

	st := ds.State
	glog.Infof("status for %s: %+v", name, st)

	if st.NeedsImprovement {
		out.T(style.Improvement, `For improved {{.driver}} performance, {{.fix}}`, out.V{"driver": driver.FullName(ds.Name), "fix": translate.T(st.Fix)})
	}

	if st.Error == nil {
		return
	}

	if !st.Installed {
		exit.Message(reason.Kind{
			ID:       fmt.Sprintf("PROVIDER_%s_NOT_FOUND", strings.ToUpper(name)),
			Advice:   translate.T(st.Fix),
			ExitCode: reason.ExProviderNotFound,
			URL:      st.Doc,
			Style:    style.Shrug,
		}, `The '{{.driver}}' provider was not found: {{.error}}`, out.V{"driver": name, "error": st.Error})
	}

	id := fmt.Sprintf("PROVIDER_%s_ERROR", strings.ToUpper(name))
	code := reason.ExProviderUnavailable

	if !st.Running {
		id = fmt.Sprintf("PROVIDER_%s_NOT_RUNNING", strings.ToUpper(name))
		code = reason.ExProviderNotRunning
	}

	exitIfNotForced(reason.Kind{
		ID:       id,
		Advice:   translate.T(st.Fix),
		ExitCode: code,
		URL:      st.Doc,
		Style:    style.Fatal,
	}, st.Error.Error())
}

func selectImageRepository(mirrorCountry string, v semver.Version) (bool, string, error) {
	var tryCountries []string
	var fallback string
	glog.Infof("selecting image repository for country %s ...", mirrorCountry)

	if mirrorCountry != "" {
		localRepos, ok := constants.ImageRepositories[mirrorCountry]
		if !ok || len(localRepos) == 0 {
			return false, "", fmt.Errorf("invalid image mirror country code: %s", mirrorCountry)
		}

		tryCountries = append(tryCountries, mirrorCountry)

		// we'll use the first repository as fallback
		// when none of the mirrors in the given location is available
		fallback = localRepos[0]

	} else {
		// always make sure global is preferred
		tryCountries = append(tryCountries, "global")
		for k := range constants.ImageRepositories {
			if strings.ToLower(k) != "global" {
				tryCountries = append(tryCountries, k)
			}
		}
	}

	checkRepository := func(repo string) error {
		pauseImage := images.Pause(v, repo)
		ref, err := name.ParseReference(pauseImage, name.WeakValidation)
		if err != nil {
			return err
		}

		_, err = remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
		return err
	}

	for _, code := range tryCountries {
		localRepos := constants.ImageRepositories[code]
		for _, repo := range localRepos {
			err := checkRepository(repo)
			if err == nil {
				return true, repo, nil
			}
		}
	}

	return false, fallback, nil
}

// validateUser validates minikube is run by the recommended user (privileged or regular)
func validateUser(drvName string) {
	u, err := user.Current()
	if err != nil {
		glog.Errorf("Error getting the current user: %v", err)
		return
	}

	useForce := viper.GetBool(force)

	if driver.NeedsRoot(drvName) && u.Uid != "0" && !useForce {
		exit.Message(reason.DrvNeedsRoot, `The "{{.driver_name}}" driver requires root privileges. Please run minikube using 'sudo -E minikube start --driver={{.driver_name}}'.`, out.V{"driver_name": drvName})
	}

	// If root is required, or we are not root, exit early
	if driver.NeedsRoot(drvName) || u.Uid != "0" {
		return
	}

	out.ErrT(style.Stopped, `The "{{.driver_name}}" driver should not be used with root privileges.`, out.V{"driver_name": drvName})
	out.ErrT(style.Tip, "If you are running minikube within a VM, consider using --driver=none:")
	out.ErrT(style.Documentation, "  https://minikube.sigs.k8s.io/docs/reference/drivers/none/")

	cname := ClusterFlagValue()
	_, err = config.Load(cname)
	if err == nil || !config.IsNotExist(err) {
		out.ErrT(style.Tip, "Tip: To remove this root owned cluster, run: sudo {{.cmd}}", out.V{"cmd": mustload.ExampleCmd(cname, "delete")})
	}

	if !useForce {
		exit.Message(reason.DrvAsRoot, `The "{{.driver_name}}" driver should not be used with root privileges.`, out.V{"driver_name": drvName})
	}
}

// memoryLimits returns the amount of memory allocated to the system and hypervisor, the return value is in MiB
func memoryLimits(drvName string) (int, int, error) {
	info, cpuErr, memErr, diskErr := machine.CachedHostInfo()
	if cpuErr != nil {
		glog.Warningf("could not get system cpu info while verifying memory limits, which might be okay: %v", cpuErr)
	}
	if diskErr != nil {
		glog.Warningf("could not get system disk info while verifying memory limits, which might be okay: %v", diskErr)
	}

	if memErr != nil {
		return -1, -1, memErr
	}

	sysLimit := int(info.Memory)
	containerLimit := 0

	if driver.IsKIC(drvName) {
		s, err := oci.CachedDaemonInfo(drvName)
		if err != nil {
			return -1, -1, err
		}
		containerLimit = int(s.TotalMemory / 1024 / 1024)
	}

	return sysLimit, containerLimit, nil
}

// suggestMemoryAllocation calculates the default memory footprint in MiB
func suggestMemoryAllocation(sysLimit int, containerLimit int, nodes int) int {
	if mem := viper.GetInt(memory); mem != 0 {
		return mem
	}
	fallback := 2200
	maximum := 6000

	if sysLimit > 0 && fallback > sysLimit {
		return sysLimit
	}

	// If there are container limits, add tiny bit of slack for non-minikube components
	if containerLimit > 0 {
		if fallback > containerLimit {
			return containerLimit
		}
		maximum = containerLimit - 48
	}

	// Suggest 25% of RAM, rounded to nearest 100MB. Hyper-V requires an even number!
	suggested := int(float32(sysLimit)/400.0) * 100

	if nodes > 1 {
		suggested /= nodes
	}

	if suggested > maximum {
		return maximum
	}

	if suggested < fallback {
		return fallback
	}

	return suggested
}

// validateRequestedMemorySize validates the memory size matches the minimum recommended
func validateRequestedMemorySize(req int, drvName string) {
	// TODO: Fix MB vs MiB confusion
	sysLimit, containerLimit, err := memoryLimits(drvName)
	if err != nil {
		glog.Warningf("Unable to query memory limits: %v", err)
	}

	// Detect if their system doesn't have enough memory to work with.
	if driver.IsKIC(drvName) && containerLimit < minUsableMem {
		if driver.IsDockerDesktop(drvName) {
			if runtime.GOOS == "darwin" {
				exitIfNotForced(reason.RsrcInsufficientDarwinDockerMemory, "Docker Desktop only has {{.size}}MiB available, less than the required {{.req}}MiB for Kubernetes", out.V{"size": containerLimit, "req": minUsableMem, "recommend": "2.25 GB"})
			} else {
				exitIfNotForced(reason.RsrcInsufficientWindowsDockerMemory, "Docker Desktop only has {{.size}}MiB available, less than the required {{.req}}MiB for Kubernetes", out.V{"size": containerLimit, "req": minUsableMem, "recommend": "2.25 GB"})
			}
		}
		exitIfNotForced(reason.RsrcInsufficientContainerMemory, "{{.driver}} only has {{.size}}MiB available, less than the required {{.req}}MiB for Kubernetes", out.V{"size": containerLimit, "driver": drvName, "req": minUsableMem})
	}

	if sysLimit < minUsableMem {
		exitIfNotForced(reason.RsrcInsufficientSysMemory, "System only has {{.size}}MiB available, less than the required {{.req}}MiB for Kubernetes", out.V{"size": containerLimit, "driver": drvName, "req": minUsableMem})
	}

	if req < minUsableMem {
		exitIfNotForced(reason.RsrcInsufficientReqMemory, "Requested memory allocation {{.requested}}MiB is less than the usable minimum of {{.minimum_memory}}MB", out.V{"requested": req, "minimum_memory": minUsableMem})
	}
	if req < minRecommendedMem {
		if driver.IsDockerDesktop(drvName) {
			if runtime.GOOS == "darwin" {
				out.WarnReason(reason.RsrcInsufficientDarwinDockerMemory, "Docker Desktop only has {{.size}}MiB available, you may encounter application deployment failures.", out.V{"size": containerLimit, "req": minUsableMem, "recommend": "2.25 GB"})
			} else {
				out.WarnReason(reason.RsrcInsufficientWindowsDockerMemory, "Docker Desktop only has {{.size}}MiB available, you may encounter application deployment failures.", out.V{"size": containerLimit, "req": minUsableMem, "recommend": "2.25 GB"})
			}
		} else {
			out.WarnReason(reason.RsrcInsufficientReqMemory, "Requested memory allocation ({{.requested}}MB) is less than the recommended minimum {{.recommend}}MB. Deployments may fail.", out.V{"requested": req, "recommend": minRecommendedMem})
		}
	}

	advised := suggestMemoryAllocation(sysLimit, containerLimit, viper.GetInt(nodes))
	if req > sysLimit {
		exitIfNotForced(reason.Kind{ID: "RSRC_OVER_ALLOC_MEM", Advice: "Start minikube with less memory allocated: 'minikube start --memory={{.advised}}mb'"},
			`Requested memory allocation {{.requested}}MB is more than your system limit {{.system_limit}}MB.`,
			out.V{"requested": req, "system_limit": sysLimit, "advised": advised})
	}

	// Recommend 1GB to handle OS/VM overhead
	maxAdvised := sysLimit - 1024
	if req > maxAdvised {
		out.WarnReason(reason.Kind{ID: "RSRC_OVER_ALLOC_MEM", Advice: "Start minikube with less memory allocated: 'minikube start --memory={{.advised}}mb'"},
			`The requested memory allocation of {{.requested}}MiB does not leave room for system overhead (total system memory: {{.system_limit}}MiB). You may face stability issues.`,
			out.V{"requested": req, "system_limit": sysLimit, "advised": advised})
	}
}

// validateCPUCount validates the cpu count matches the minimum recommended
func validateCPUCount(drvName string) {
	var cpuCount int
	if driver.BareMetal(drvName) {
		// Uses the gopsutil cpu package to count the number of physical cpu cores
		ci, err := cpu.Counts(false)
		if err != nil {
			glog.Warningf("Unable to get CPU info: %v", err)
		} else {
			cpuCount = ci
		}
	} else {
		cpuCount = viper.GetInt(cpus)
	}

	if cpuCount < minimumCPUS {
		exitIfNotForced(reason.RsrcInsufficientCores, "Requested cpu count {{.requested_cpus}} is less than the minimum allowed of {{.minimum_cpus}}", out.V{"requested_cpus": cpuCount, "minimum_cpus": minimumCPUS})
	}

	if !driver.IsKIC((drvName)) {
		return
	}

	si, err := oci.CachedDaemonInfo(drvName)
	if err != nil {
		out.T(style.Confused, "Failed to verify '{{.driver_name}} info' will try again ...", out.V{"driver_name": drvName})
		si, err = oci.DaemonInfo(drvName)
		if err != nil {
			exit.Message(reason.Usage, "Ensure your {{.driver_name}} is running and is healthy.", out.V{"driver_name": driver.FullName(drvName)})
		}

	}

	// looks good
	if si.CPUs >= 2 {
		return
	}

	if drvName == oci.Docker && runtime.GOOS == "darwin" {
		exitIfNotForced(reason.RsrcInsufficientDarwinDockerCores, "Docker Desktop has less than 2 CPUs configured, but Kubernetes requires at least 2 to be available")
	} else if drvName == oci.Docker && runtime.GOOS == "windows" {
		exitIfNotForced(reason.RsrcInsufficientWindowsDockerCores, "Docker Desktop has less than 2 CPUs configured, but Kubernetes requires at least 2 to be available")
	} else {
		exitIfNotForced(reason.RsrcInsufficientCores, "{{.driver_name}} has less than 2 CPUs available, but Kubernetes requires at least 2 to be available", out.V{"driver_name": driver.FullName(viper.GetString("driver"))})
	}
}

// validateFlags validates the supplied flags against known bad combinations
func validateFlags(cmd *cobra.Command, drvName string) {
	if cmd.Flags().Changed(humanReadableDiskSize) {
		diskSizeMB, err := util.CalculateSizeInMB(viper.GetString(humanReadableDiskSize))
		if err != nil {
			exitIfNotForced(reason.Usage, "Validation unable to parse disk size '{{.diskSize}}': {{.error}}", out.V{"diskSize": viper.GetString(humanReadableDiskSize), "error": err})
		}

		if diskSizeMB < minimumDiskSize {
			exitIfNotForced(reason.RsrcInsufficientStorage, "Requested disk size {{.requested_size}} is less than minimum of {{.minimum_size}}", out.V{"requested_size": diskSizeMB, "minimum_size": minimumDiskSize})
		}
	}

	if cmd.Flags().Changed(cpus) {
		if !driver.HasResourceLimits(drvName) {
			out.WarningT("The '{{.name}}' driver does not respect the --cpus flag", out.V{"name": drvName})
		}
	}

	validateCPUCount(drvName)

	if cmd.Flags().Changed(memory) {
		if !driver.HasResourceLimits(drvName) {
			out.WarningT("The '{{.name}}' driver does not respect the --memory flag", out.V{"name": drvName})
		}
		req, err := util.CalculateSizeInMB(viper.GetString(memory))
		if err != nil {
			exitIfNotForced(reason.Usage, "Unable to parse memory '{{.memory}}': {{.error}}", out.V{"memory": viper.GetString(memory), "error": err})
		}
		validateRequestedMemorySize(req, drvName)
	}

	if cmd.Flags().Changed(containerRuntime) {
		runtime := strings.ToLower(viper.GetString(containerRuntime))

		validOptions := cruntime.ValidRuntimes()
		// `crio` is accepted as an alternative spelling to `cri-o`
		validOptions = append(validOptions, constants.CRIO)

		var validRuntime bool
		for _, option := range validOptions {
			if runtime == option {
				validRuntime = true
			}

			// Convert `cri-o` to `crio` as the K8s config uses the `crio` spelling
			if runtime == "cri-o" {
				viper.Set(containerRuntime, constants.CRIO)
			}
		}

		if !validRuntime {
			exit.Message(reason.Usage, `Invalid Container Runtime: "{{.runtime}}". Valid runtimes are: {{.validOptions}}`, out.V{"runtime": runtime, "validOptions": strings.Join(cruntime.ValidRuntimes(), ", ")})
		}
	}

	if driver.BareMetal(drvName) {
		if ClusterFlagValue() != constants.DefaultClusterName {
			exit.Message(reason.DrvUnsupportedProfile, "The '{{.name}} driver does not support multiple profiles: https://minikube.sigs.k8s.io/docs/reference/drivers/none/", out.V{"name": drvName})
		}

		runtime := viper.GetString(containerRuntime)
		if runtime != "docker" {
			out.WarningT("Using the '{{.runtime}}' runtime with the 'none' driver is an untested configuration!", out.V{"runtime": runtime})
		}

		// conntrack is required starting with Kubernetes 1.18, include the release candidates for completion
		version, _ := util.ParseKubernetesVersion(getKubernetesVersion(nil))
		if version.GTE(semver.MustParse("1.18.0-beta.1")) {
			if _, err := exec.LookPath("conntrack"); err != nil {
				exit.Message(reason.GuestMissingConntrack, "Sorry, Kubernetes {{.k8sVersion}} requires conntrack to be installed in root's path", out.V{"k8sVersion": version.String()})
			}
		}
	}

	// validate kubeadm extra args
	if invalidOpts := bsutil.FindInvalidExtraConfigFlags(config.ExtraOptions); len(invalidOpts) > 0 {
		out.WarningT(
			"These --extra-config parameters are invalid: {{.invalid_extra_opts}}",
			out.V{"invalid_extra_opts": invalidOpts},
		)
		exit.Message(
			reason.Usage,
			"Valid components are: {{.valid_extra_opts}}",
			out.V{"valid_extra_opts": bsutil.KubeadmExtraConfigOpts},
		)
	}

	// check that kubeadm extra args contain only allowed parameters
	for param := range config.ExtraOptions.AsMap().Get(bsutil.Kubeadm) {
		if !config.ContainsParam(bsutil.KubeadmExtraArgsAllowed[bsutil.KubeadmCmdParam], param) &&
			!config.ContainsParam(bsutil.KubeadmExtraArgsAllowed[bsutil.KubeadmConfigParam], param) {
			exit.Message(reason.Usage, "Sorry, the kubeadm.{{.parameter_name}} parameter is currently not supported by --extra-config", out.V{"parameter_name": param})
		}
	}

	if s := viper.GetString(startOutput); s != "text" && s != "json" {
		exit.Message(reason.Usage, "Sorry, please set the --output flag to one of the following valid options: [text,json]")
	}

	validateRegistryMirror()
}

// This function validates if the --registry-mirror
// args match the format of http://localhost
func validateRegistryMirror() {
	if len(registryMirror) > 0 {
		for _, loc := range registryMirror {
			URL, err := url.Parse(loc)
			if err != nil {
				glog.Errorln("Error Parsing URL: ", err)
			}
			if (URL.Scheme != "http" && URL.Scheme != "https") || URL.Path != "" {
				exit.Message(reason.Usage, "Sorry, the url provided with the --registry-mirror flag is invalid: {{.url}}", out.V{"url": loc})
			}

		}
	}
}

func createNode(cc config.ClusterConfig, kubeNodeName string, existing *config.ClusterConfig) (config.ClusterConfig, config.Node, error) {
	// Create the initial node, which will necessarily be a control plane
	if existing != nil {
		cp, err := config.PrimaryControlPlane(existing)
		cp.KubernetesVersion = getKubernetesVersion(&cc)
		if err != nil {
			return cc, config.Node{}, err
		}

		// Make sure that existing nodes honor if KubernetesVersion gets specified on restart
		// KubernetesVersion is the only attribute that the user can override in the Node object
		nodes := []config.Node{}
		for _, n := range existing.Nodes {
			n.KubernetesVersion = getKubernetesVersion(&cc)
			nodes = append(nodes, n)
		}
		cc.Nodes = nodes

		return cc, cp, nil
	}

	cp := config.Node{
		Port:              cc.KubernetesConfig.NodePort,
		KubernetesVersion: getKubernetesVersion(&cc),
		Name:              kubeNodeName,
		ControlPlane:      true,
		Worker:            true,
	}
	cc.Nodes = []config.Node{cp}
	return cc, cp, nil
}

// autoSetDriverOptions sets the options needed for specific driver automatically.
func autoSetDriverOptions(cmd *cobra.Command, drvName string) (err error) {
	err = nil
	hints := driver.FlagDefaults(drvName)
	if len(hints.ExtraOptions) > 0 {
		for _, eo := range hints.ExtraOptions {
			if config.ExtraOptions.Exists(eo) {
				glog.Infof("skipping extra-config %q.", eo)
				continue
			}
			glog.Infof("auto setting extra-config to %q.", eo)
			err = config.ExtraOptions.Set(eo)
			if err != nil {
				err = errors.Wrapf(err, "setting extra option %s", eo)
			}
		}
	}

	if !cmd.Flags().Changed(cacheImages) {
		viper.Set(cacheImages, hints.CacheImages)
	}

	if !cmd.Flags().Changed(containerRuntime) && hints.ContainerRuntime != "" {
		viper.Set(containerRuntime, hints.ContainerRuntime)
		glog.Infof("auto set %s to %q.", containerRuntime, hints.ContainerRuntime)
	}

	if !cmd.Flags().Changed(cmdcfg.Bootstrapper) && hints.Bootstrapper != "" {
		viper.Set(cmdcfg.Bootstrapper, hints.Bootstrapper)
		glog.Infof("auto set %s to %q.", cmdcfg.Bootstrapper, hints.Bootstrapper)

	}

	return err
}

// validateKubernetesVersion ensures that the requested version is reasonable
func validateKubernetesVersion(old *config.ClusterConfig) {
	nvs, _ := semver.Make(strings.TrimPrefix(getKubernetesVersion(old), version.VersionPrefix))

	oldestVersion, err := semver.Make(strings.TrimPrefix(constants.OldestKubernetesVersion, version.VersionPrefix))
	if err != nil {
		exit.Message(reason.InternalSemverParse, "Unable to parse oldest Kubernetes version from constants: {{.error}}", out.V{"error": err})
	}
	defaultVersion, err := semver.Make(strings.TrimPrefix(constants.DefaultKubernetesVersion, version.VersionPrefix))
	if err != nil {
		exit.Message(reason.InternalSemverParse, "Unable to parse default Kubernetes version from constants: {{.error}}", out.V{"error": err})
	}

	if nvs.LT(oldestVersion) {
		out.WarningT("Specified Kubernetes version {{.specified}} is less than the oldest supported version: {{.oldest}}", out.V{"specified": nvs, "oldest": constants.OldestKubernetesVersion})
		if !viper.GetBool(force) {
			out.WarningT("You can force an unsupported Kubernetes version via the --force flag")
		}
		exitIfNotForced(reason.KubernetesTooOld, "Kubernetes {{.version}} is not supported by this release of minikube", out.V{"version": nvs})
	}

	if old == nil || old.KubernetesConfig.KubernetesVersion == "" {
		return
	}

	ovs, err := semver.Make(strings.TrimPrefix(old.KubernetesConfig.KubernetesVersion, version.VersionPrefix))
	if err != nil {
		glog.Errorf("Error parsing old version %q: %v", old.KubernetesConfig.KubernetesVersion, err)
	}

	if nvs.LT(ovs) {
		profileArg := ""
		if old.Name != constants.DefaultClusterName {
			profileArg = fmt.Sprintf(" -p %s", old.Name)
		}

		suggestedName := old.Name + "2"
		exit.Message(reason.KubernetesDowngrade, "Unable to safely downgrade existing Kubernetes v{{.old}} cluster to v{{.new}}",
			out.V{"prefix": version.VersionPrefix, "new": nvs, "old": ovs, "profile": profileArg, "suggestedName": suggestedName})

	}
	if defaultVersion.GT(nvs) {
		out.T(style.New, "Kubernetes {{.new}} is now available. If you would like to upgrade, specify: --kubernetes-version={{.prefix}}{{.new}}", out.V{"prefix": version.VersionPrefix, "new": defaultVersion})
	}
}

func getKubernetesVersion(old *config.ClusterConfig) string {
	paramVersion := viper.GetString(kubernetesVersion)

	// try to load the old version first if the user didn't specify anything
	if paramVersion == "" && old != nil {
		paramVersion = old.KubernetesConfig.KubernetesVersion
	}

	if paramVersion == "" || strings.EqualFold(paramVersion, "stable") {
		paramVersion = constants.DefaultKubernetesVersion
	} else if strings.EqualFold(paramVersion, "latest") {
		paramVersion = constants.NewestKubernetesVersion
	}

	nvs, err := semver.Make(strings.TrimPrefix(paramVersion, version.VersionPrefix))
	if err != nil {
		exit.Message(reason.Usage, `Unable to parse "{{.kubernetes_version}}": {{.error}}`, out.V{"kubernetes_version": paramVersion, "error": err})
	}

	return version.VersionPrefix + nvs.String()
}

func exitIfNotForced(r reason.Kind, message string, v ...out.V) {
	if !viper.GetBool(force) {
		exit.Message(r, message, v...)
	}
	out.Error(r, message, v...)
}
