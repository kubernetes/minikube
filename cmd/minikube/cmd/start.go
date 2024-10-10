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
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/blang/semver/v4"
	"github.com/docker/go-connections/nat"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/cpu"
	gopshost "github.com/shirou/gopsutil/v3/host"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/firewall"
	netutil "k8s.io/minikube/pkg/network"

	"k8s.io/klog/v2"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/driver/auxdriver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/pause"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	pkgtrace "k8s.io/minikube/pkg/trace"

	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/translate"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/version"
)

type versionJSON struct {
	IsoVersion      string `json:"iso_version"`
	KicbaseVersion  string `json:"kicbase_version"`
	MinikubeVersion string `json:"minikube_version"`
	Commit          string `json:"commit"`
}

var (
	// ErrKubernetesPatchNotFound is when a patch was not found for the given <major>.<minor> version
	ErrKubernetesPatchNotFound = errors.New("Unable to detect the latest patch release for specified Kubernetes version")
	registryMirror             []string
	insecureRegistry           []string
	apiServerNames             []string
	apiServerIPs               []net.IP
	hostRe                     = regexp.MustCompile(`^[^-][\w\.-]+$`)
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
		s.WriteString(fmt.Sprintf("%s %s", cases.Title(language.Und).String(hi.Platform), hi.PlatformVersion))
		klog.Infof("hostinfo: %+v", hi)
	} else {
		klog.Warningf("gopshost.Info returned error: %v", err)
		s.WriteString(runtime.GOOS)
	}

	vsys, vrole, err := gopshost.Virtualization()
	if err != nil {
		klog.Warningf("gopshost.Virtualization returned error: %v", err)
	} else {
		klog.Infof("virtualization: %s %s", vsys, vrole)
	}

	// This environment is exotic, let's output a bit more.
	if vrole == "guest" || runtime.GOARCH != "amd64" {
		if vrole == "guest" && vsys != "" {
			s.WriteString(fmt.Sprintf(" (%s/%s)", vsys, runtime.GOARCH))
		} else {
			s.WriteString(fmt.Sprintf(" (%s)", runtime.GOARCH))
		}
	}
	return s.String()
}

// runStart handles the executes the flow of "minikube start"
func runStart(cmd *cobra.Command, _ []string) {
	register.SetEventLogPath(localpath.EventLog(ClusterFlagValue()))
	ctx := context.Background()
	out.SetJSON(outputFormat == "json")
	if err := pkgtrace.Initialize(viper.GetString(trace)); err != nil {
		exit.Message(reason.Usage, "error initializing tracing: {{.Error}}", out.V{"Error": err.Error()})
	}
	defer pkgtrace.Cleanup()

	displayVersion(version.GetVersion())
	go download.CleanUpOlderPreloads()

	// Avoid blocking execution on optional HTTP fetches
	go notify.MaybePrintUpdateTextFromGithub()

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
		registryMirror = viper.GetStringSlice("registry-mirror")
	}

	if !config.ProfileNameValid(ClusterFlagValue()) {
		out.WarningT("Profile name '{{.name}}' is not valid", out.V{"name": ClusterFlagValue()})
		exit.Message(reason.Usage, "Only alphanumeric and dashes '-' are permitted. Minimum 2 characters, starting with alphanumeric.")
	}
	existing, err := config.Load(ClusterFlagValue())
	if err != nil && !config.IsNotExist(err) {
		kind := reason.HostConfigLoad
		if config.IsPermissionDenied(err) {
			kind = reason.HostHomePermission
		}
		exit.Message(kind, "Unable to load config: {{.error}}", out.V{"error": err})
	}

	if existing != nil {
		upgradeExistingConfig(cmd, existing)
	} else {
		validateProfileName()
	}

	validateSpecifiedDriver(existing)
	validateKubernetesVersion(existing)
	validateContainerRuntime(existing)

	ds, alts, specified := selectDriver(existing)
	if cmd.Flag(kicBaseImage).Changed {
		if !isBaseImageApplicable(ds.Name) {
			exit.Message(reason.Usage,
				"flag --{{.imgFlag}} is not available for driver '{{.driver}}'. Did you mean to use '{{.docker}}' or '{{.podman}}' driver instead?\n"+
					"Please use --{{.isoFlag}} flag to configure VM based drivers",
				out.V{
					"imgFlag": kicBaseImage,
					"driver":  ds.Name,
					"docker":  registry.Docker,
					"podman":  registry.Podman,
					"isoFlag": isoURL,
				},
			)
		}
	}

	useForce := viper.GetBool(force)

	starter, err := provisionWithDriver(cmd, ds, existing)
	if err != nil {
		node.ExitIfFatal(err, useForce)
		machine.MaybeDisplayAdvice(err, ds.Name)
		if specified {
			// If the user specified a driver, don't fallback to anything else
			exitGuestProvision(err)
		} else {
			success := false
			// Walk down the rest of the options
			for _, alt := range alts {
				// Skip non-default drivers
				if !alt.Default {
					continue
				}
				out.WarningT("Startup with {{.old_driver}} driver failed, trying with alternate driver {{.new_driver}}: {{.error}}", out.V{"old_driver": ds.Name, "new_driver": alt.Name, "error": err})
				ds = alt
				// Delete the existing cluster and try again with the next driver on the list
				profile, err := config.LoadProfile(ClusterFlagValue())
				if err != nil {
					klog.Warningf("%s profile does not exist, trying anyways.", ClusterFlagValue())
				}

				err = deleteProfile(ctx, profile)
				if err != nil {
					out.WarningT("Failed to delete cluster {{.name}}, proceeding with retry anyway.", out.V{"name": ClusterFlagValue()})
				}
				starter, err = provisionWithDriver(cmd, ds, existing)
				if err != nil {
					continue
				}
				// Success!
				success = true
				break
			}
			if !success {
				exitGuestProvision(err)
			}
		}
	}

	validateBuiltImageVersion(starter.Runner, ds.Name)

	if existing != nil && driver.IsKIC(existing.Driver) && viper.GetBool(createMount) {
		old := ""
		if len(existing.ContainerVolumeMounts) > 0 {
			old = existing.ContainerVolumeMounts[0]
		}
		if mount := viper.GetString(mountString); old != mount {
			exit.Message(reason.GuestMountConflict, "Sorry, {{.driver}} does not allow mounts to be changed after container creation (previous mount: '{{.old}}', new mount: '{{.new}})'", out.V{
				"driver": existing.Driver,
				"new":    mount,
				"old":    old,
			})
		}
	}

	kubeconfig, err := startWithDriver(cmd, starter, existing)
	if err != nil {
		node.ExitIfFatal(err, useForce)
		exit.Error(reason.GuestStart, "failed to start node", err)
	}

	if err := showKubectlInfo(kubeconfig, starter.Node.KubernetesVersion, starter.Node.ContainerRuntime, starter.Cfg.Name); err != nil {
		klog.Errorf("kubectl info: %v", err)
	}
}

func provisionWithDriver(cmd *cobra.Command, ds registry.DriverState, existing *config.ClusterConfig) (node.Starter, error) {
	driverName := ds.Name
	klog.Infof("selected driver: %s", driverName)
	validateDriver(ds, existing)
	err := autoSetDriverOptions(cmd, driverName)
	if err != nil {
		klog.Errorf("Error autoSetOptions : %v", err)
	}

	virtualBoxMacOS13PlusWarning(driverName)
	validateFlags(cmd, driverName)
	validateUser(driverName)
	if driverName == oci.Docker {
		validateDockerStorageDriver(driverName)
	}

	k8sVersion, err := getKubernetesVersion(existing)
	if err != nil {
		klog.Warningf("failed getting Kubernetes version: %v", err)
	}

	// Disallow accepting addons flag without Kubernetes
	// It places here because cluster config is required to get the old version.
	if cmd.Flags().Changed(config.AddonListFlag) {
		if k8sVersion == constants.NoKubernetesVersion || viper.GetBool(noKubernetes) {
			exit.Message(reason.Usage, "You cannot enable addons on a cluster without Kubernetes, to enable Kubernetes on your cluster, run: minikube start --kubernetes-version=stable")
		}
	}

	// Download & update the driver, even in --download-only mode
	if !viper.GetBool(dryRun) {
		updateDriver(driverName)
	}

	// Check whether we may need to stop Kubernetes.
	var stopk8s bool
	if existing != nil && viper.GetBool(noKubernetes) {
		stopk8s = true
	}

	rtime := getContainerRuntime(existing)
	cc, n, err := generateClusterConfig(cmd, existing, k8sVersion, rtime, driverName)
	if err != nil {
		return node.Starter{}, errors.Wrap(err, "Failed to generate cluster config")
	}
	klog.Infof("cluster config:\n%+v", cc)

	if firewall.IsBootpdBlocked(cc) {
		if err := firewall.UnblockBootpd(); err != nil {
			klog.Warningf("failed unblocking bootpd from firewall: %v", err)
		}
	}

	if driver.IsVM(cc.Driver) && runtime.GOARCH == "arm64" && cc.KubernetesConfig.ContainerRuntime == "crio" {
		exit.Message(reason.Unimplemented, "arm64 VM drivers do not currently support the crio container runtime. See https://github.com/kubernetes/minikube/issues/14146 for details.")
	}

	// This is about as far as we can go without overwriting config files
	if viper.GetBool(dryRun) {
		out.Step(style.DryRun, `dry-run validation complete!`)
		os.Exit(0)
	}

	if driver.IsVM(driverName) && !driver.IsSSH(driverName) {
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

	mRunner, preExists, mAPI, host, err := node.Provision(&cc, &n, viper.GetBool(deleteOnFailure))
	if err != nil {
		return node.Starter{}, err
	}

	return node.Starter{
		Runner:         mRunner,
		PreExists:      preExists,
		StopK8s:        stopk8s,
		MachineAPI:     mAPI,
		Host:           host,
		ExistingAddons: existingAddons,
		Cfg:            &cc,
		Node:           &n,
	}, nil
}

func virtualBoxMacOS13PlusWarning(driverName string) {
	if !driver.IsVirtualBox(driverName) || !detect.MacOS13Plus() {
		return
	}
	suggestedDriver := driver.HyperKit
	if runtime.GOARCH == "arm64" {
		suggestedDriver = driver.QEMU
	}
	out.WarningT(`Due to changes in macOS 13+ minikube doesn't currently support VirtualBox. You can use alternative drivers such as docker or {{.driver}}.
    https://minikube.sigs.k8s.io/docs/drivers/docker/
    https://minikube.sigs.k8s.io/docs/drivers/{{.driver}}/

    For more details on the issue see: https://github.com/kubernetes/minikube/issues/15274
`, out.V{"driver": suggestedDriver})
}

func validateBuiltImageVersion(r command.Runner, driverName string) {
	if driver.IsNone(driverName) {
		return
	}
	res, err := r.RunCmd(exec.Command("cat", "/version.json"))
	if err != nil {
		klog.Warningf("Unable to open version.json: %s", err)
		return
	}

	var versionDetails versionJSON
	if err := json.Unmarshal(res.Stdout.Bytes(), &versionDetails); err != nil {
		out.WarningT("Unable to parse version.json: {{.error}}, json: {{.json}}", out.V{"error": err, "json": res.Stdout.String()})
		return
	}

	if !imageMatchesBinaryVersion(versionDetails.MinikubeVersion, version.GetVersion()) {
		out.WarningT("Image was not built for the current minikube version. To resolve this you can delete and recreate your minikube cluster using the latest images. Expected minikube version: {{.imageMinikubeVersion}} -> Actual minikube version: {{.minikubeVersion}}", out.V{"imageMinikubeVersion": versionDetails.MinikubeVersion, "minikubeVersion": version.GetVersion()})
	}
}

func imageMatchesBinaryVersion(imageVersion, binaryVersion string) bool {
	if binaryVersion == imageVersion {
		return true
	}

	// the map below is used to map the binary version to the version the image expects
	// this is usually done when a patch version is released but a new ISO/Kicbase is not needed
	// that way a version mismatch warning won't be thrown
	//
	// ex.
	// the v1.31.0 and v1.31.1 minikube binaries both use v1.31.0 ISO & Kicbase
	// to prevent the v1.31.1 binary from throwing a version mismatch warning we use the map to change the binary version used in the comparison

	mappedVersions := map[string]string{
		"v1.31.1": "v1.31.0",
		"v1.31.2": "v1.31.0",
	}
	binaryVersion, ok := mappedVersions[binaryVersion]

	return ok && binaryVersion == imageVersion
}

func startWithDriver(cmd *cobra.Command, starter node.Starter, existing *config.ClusterConfig) (*kubeconfig.Settings, error) {
	// start primary control-plane node
	kubeconfig, err := node.Start(starter)
	if err != nil {
		kubeconfig, err = maybeDeleteAndRetry(cmd, *starter.Cfg, *starter.Node, starter.ExistingAddons, err)
		if err != nil {
			return nil, err
		}
	}

	// target total and number of control-plane nodes
	numCPNodes := 1
	numNodes := viper.GetInt(nodes)
	if existing != nil {
		numCPNodes = 0
		for _, n := range existing.Nodes {
			if n.ControlPlane {
				numCPNodes++
			}
		}
		numNodes = len(existing.Nodes)
	} else if viper.GetBool(ha) {
		numCPNodes = 3
	}

	// apart from starter, add any additional existing or new nodes
	for i := 1; i < numNodes; i++ {
		var n config.Node
		if existing != nil {
			n = existing.Nodes[i]
		} else {
			nodeName := node.Name(i + 1)
			n = config.Node{
				Name:              nodeName,
				Port:              starter.Cfg.APIServerPort,
				KubernetesVersion: starter.Cfg.KubernetesConfig.KubernetesVersion,
				ContainerRuntime:  starter.Cfg.KubernetesConfig.ContainerRuntime,
				Worker:            true,
			}
			if i < numCPNodes { // starter node is also counted as (primary) cp node
				n.ControlPlane = true
			}
		}

		out.Ln("") // extra newline for clarity on the command line
		if err := node.Add(starter.Cfg, n, viper.GetBool(deleteOnFailure)); err != nil {
			return nil, errors.Wrap(err, "adding node")
		}
	}

	pause.RemovePausedFile(starter.Runner)

	return kubeconfig, nil
}

func warnAboutMultiNodeCNI() {
	out.WarningT("Cluster was created without any CNI, adding a node to it might cause broken networking.")
}

func updateDriver(driverName string) {
	v, err := version.GetSemverVersion()
	if err != nil {
		out.WarningT("Error parsing minikube version: {{.error}}", out.V{"error": err})
	} else if err := auxdriver.InstallOrUpdate(driverName, localpath.MakeMiniPath("bin"), v, viper.GetBool(interactive), viper.GetBool(autoUpdate)); err != nil {
		out.WarningT("Unable to update {{.driver}} driver: {{.error}}", out.V{"driver": driverName, "error": err})
	}
}

func displayVersion(version string) {
	prefix := ""
	if ClusterFlagValue() != constants.DefaultClusterName {
		prefix = fmt.Sprintf("[%s] ", ClusterFlagValue())
	}

	register.Reg.SetStep(register.InitialSetup)
	out.Step(style.Happy, "{{.prefix}}minikube {{.version}} on {{.platform}}", out.V{"prefix": prefix, "version": version, "platform": platform()})
}

// displayEnviron makes the user aware of environment variables that will affect how minikube operates
func displayEnviron(env []string) {
	for _, kv := range env {
		bits := strings.SplitN(kv, "=", 2)
		if len(bits) < 2 {
			continue
		}
		k := bits[0]
		v := bits[1]
		if strings.HasPrefix(k, "MINIKUBE_") || k == constants.KubeconfigEnvVar {
			out.Infof("{{.key}}={{.value}}", out.V{"key": k, "value": v})
		}
	}
}

func showKubectlInfo(kcs *kubeconfig.Settings, k8sVersion, rtime, machineName string) error {
	if k8sVersion == constants.NoKubernetesVersion {
		register.Reg.SetStep(register.Done)
		out.Step(style.Ready, "Done! minikube is ready without Kubernetes!")

		// Runtime message.
		boxConfig := box.Config{Py: 1, Px: 4, Type: "Round", Color: "Green"}
		switch rtime {
		case constants.Docker:
			out.BoxedWithConfig(boxConfig, style.Tip, "Things to try without Kubernetes ...", `- "minikube ssh" to SSH into minikube's node.
- "minikube docker-env" to point your docker-cli to the docker inside minikube.
- "minikube image" to build images without docker.`)
		case constants.Containerd:
			out.BoxedWithConfig(boxConfig, style.Tip, "Things to try without Kubernetes ...", `- "minikube ssh" to SSH into minikube's node.
- "minikube image" to build images without docker.`)
		case constants.CRIO:
			out.BoxedWithConfig(boxConfig, style.Tip, "Things to try without Kubernetes ...", `- "minikube ssh" to SSH into minikube's node.
- "minikube podman-env" to point your podman-cli to the podman inside minikube.
- "minikube image" to build images without docker.`)
		}
		return nil
	}

	// To be shown at the end, regardless of exit path
	defer func() {
		register.Reg.SetStep(register.Done)
		if kcs.KeepContext {
			out.Step(style.Kubectl, "To connect to this cluster, use:  --context={{.name}}", out.V{"name": kcs.ClusterName})
		} else {
			out.Step(style.Ready, `Done! kubectl is now configured to use "{{.name}}" cluster and "{{.ns}}" namespace by default`, out.V{"name": machineName, "ns": kcs.Namespace})
		}
	}()

	path, err := exec.LookPath("kubectl")
	if err != nil {
		out.Styled(style.Tip, "kubectl not found. If you need it, try: 'minikube kubectl -- get pods -A'")
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
	klog.Infof("kubectl: %s, cluster: %s (minor skew: %d)", client, cluster, minorSkew)

	if client.Major != cluster.Major || minorSkew > 1 {
		out.Ln("")
		out.WarningT("{{.path}} is version {{.client_version}}, which may have incompatibilities with Kubernetes {{.cluster_version}}.",
			out.V{"path": path, "client_version": client, "cluster_version": cluster})
		out.Infof("Want kubectl {{.version}}? Try 'minikube kubectl -- get pods -A'", out.V{"version": k8sVersion})
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

		err = deleteProfile(context.Background(), profile)
		if err != nil {
			out.WarningT("Failed to delete cluster {{.name}}, proceeding with retry anyway.", out.V{"name": existing.Name})
		}

		// Re-generate the cluster config, just in case the failure was related to an old config format
		cc := updateExistingConfigFromFlags(cmd, &existing)
		var kubeconfig *kubeconfig.Settings
		for _, n := range cc.Nodes {
			r, p, m, h, err := node.Provision(&cc, &n, false)
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

			k, err := node.Start(s)
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

// returns (current_driver, suggested_drivers, "true, if the driver is set by command line arg or in the config file")
func selectDriver(existing *config.ClusterConfig) (registry.DriverState, []registry.DriverState, bool) {
	// Technically unrelated, but important to perform before detection
	driver.SetLibvirtURI(viper.GetString(kvmQemuURI))
	register.Reg.SetStep(register.SelectingDriver)
	// By default, the driver is whatever we used last time
	if existing != nil {
		old := hostDriver(existing)
		ds := driver.Status(old)
		out.Step(style.Sparkle, `Using the {{.driver}} driver based on existing profile`, out.V{"driver": ds.String()})
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
			exit.Message(reason.DrvUnsupportedOS, "The driver '{{.driver}}' is not supported on {{.os}}/{{.arch}}", out.V{"driver": d, "os": runtime.GOOS, "arch": runtime.GOARCH})
		}
		out.Step(style.Sparkle, `Using the {{.driver}} driver based on user configuration`, out.V{"driver": ds.String()})
		return ds, nil, true
	}

	// Fallback to old driver parameter
	if d := viper.GetString("vm-driver"); d != "" {
		ds := driver.Status(viper.GetString("vm-driver"))
		if ds.Name == "" {
			exit.Message(reason.DrvUnsupportedOS, "The driver '{{.driver}}' is not supported on {{.os}}/{{.arch}}", out.V{"driver": d, "os": runtime.GOOS, "arch": runtime.GOARCH})
		}
		out.Step(style.Sparkle, `Using the {{.driver}} driver based on user configuration`, out.V{"driver": ds.String()})
		return ds, nil, true
	}

	choices := driver.Choices(viper.GetBool("vm"))
	pick, alts, rejects := driver.Suggest(choices)
	if pick.Name == "" {
		out.Step(style.ThumbsDown, "Unable to pick a default driver. Here is what was considered, in preference order:")
		sort.Slice(rejects, func(i, j int) bool {
			if rejects[i].Priority == rejects[j].Priority {
				return rejects[i].Preference > rejects[j].Preference
			}
			return rejects[i].Priority > rejects[j].Priority
		})

		// Display the issue for installed drivers
		for _, r := range rejects {
			if r.Default && r.State.Installed {
				out.Infof("{{ .name }}: {{ .rejection }}", out.V{"name": r.Name, "rejection": r.Rejection})
				if r.Suggestion != "" {
					out.Infof("{{ .name }}: Suggestion: {{ .suggestion}}", out.V{"name": r.Name, "suggestion": r.Suggestion})
				}
			}
		}

		// Display the other drivers users can install
		out.Step(style.Tip, "Alternatively you could install one of these drivers:")
		for _, r := range rejects {
			if !r.Default || r.State.Installed {
				continue
			}
			out.Infof("{{ .name }}: {{ .rejection }}", out.V{"name": r.Name, "rejection": r.Rejection})
			if r.Suggestion != "" {
				out.Infof("{{ .name }}: Suggestion: {{ .suggestion}}", out.V{"name": r.Name, "suggestion": r.Suggestion})
			}
		}
		foundStoppedDocker := false
		foundUnhealthy := false
		for _, reject := range rejects {
			if reject.Name == driver.Docker && reject.State.Installed && !reject.State.Running {
				foundStoppedDocker = true
				break
			} else if reject.State.Installed && !reject.State.Healthy {
				foundUnhealthy = true
				break
			}
		}
		if foundStoppedDocker {
			exit.Message(reason.DrvDockerNotRunning, "Found docker, but the docker service isn't running. Try restarting the docker service.")
		} else if foundUnhealthy {
			exit.Message(reason.DrvNotHealthy, "Found driver(s) but none were healthy. See above for suggestions how to fix installed drivers.")
		} else {
			exit.Message(reason.DrvNotDetected, "No possible driver was detected. Try specifying --driver, or see https://minikube.sigs.k8s.io/docs/start/")
		}
	}

	if len(alts) > 1 {
		altNames := []string{}
		for _, a := range alts {
			altNames = append(altNames, a.String())
		}
		out.Step(style.Sparkle, `Automatically selected the {{.driver}} driver. Other choices: {{.alternates}}`, out.V{"driver": pick.Name, "alternates": strings.Join(altNames, ", ")})
	} else {
		out.Step(style.Sparkle, `Automatically selected the {{.driver}} driver`, out.V{"driver": pick.String()})
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
		klog.Warningf("selectDriver NewAPIClient: %v", err)
		return existing.Driver
	}

	cp, err := config.ControlPlane(*existing)
	if err != nil {
		klog.Errorf("Unable to get primary control-plane node from existing config: %v", err)
		return existing.Driver
	}

	machineName := config.MachineName(*existing, cp)
	h, err := api.Load(machineName)
	if err != nil {
		klog.Errorf("api.Load failed for %s: %v", machineName, err)
		return existing.Driver
	}

	return h.Driver.DriverName()
}

// validateProfileName makes sure that new profile name not duplicated with any of machine names in existing multi-node clusters.
func validateProfileName() {
	profiles, err := config.ListValidProfiles()
	if err != nil {
		exit.Message(reason.InternalListConfig, "Unable to list profiles: {{.error}}", out.V{"error": err})
	}
	for _, p := range profiles {
		for _, n := range p.Config.Nodes {
			machineName := config.MachineName(*p.Config, n)
			if ClusterFlagValue() == machineName {
				out.WarningT("Profile name '{{.name}}' is duplicated with machine name '{{.machine}}' in profile '{{.profile}}'", out.V{"name": ClusterFlagValue(),
					"machine": machineName,
					"profile": p.Name})
				exit.Message(reason.Usage, "Profile name should be unique")
			}
		}
	}
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

	// hostDriver always returns original driver name even if an alias is used to start minikube.
	// For all next start with alias needs to be check against the host driver aliases.
	if driver.IsAlias(old, requested) {
		return
	}

	if viper.GetBool(deleteOnFailure) {
		out.WarningT("Deleting existing cluster {{.name}} with different driver {{.driver_name}} due to --delete-on-failure flag set by the user. ", out.V{"name": existing.Name, "driver_name": old})
		// Start failed, delete the cluster
		profile, err := config.LoadProfile(existing.Name)
		if err != nil {
			out.ErrT(style.Meh, `"{{.name}}" profile does not exist, trying anyways.`, out.V{"name": existing.Name})
		}

		err = deleteProfile(context.Background(), profile)
		if err != nil {
			out.WarningT("Failed to delete cluster {{.name}}.", out.V{"name": existing.Name})
		}
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
	os := detect.RuntimeOS()
	arch := detect.RuntimeArch()
	klog.Infof("validating driver %q against %+v", name, existing)
	if !driver.Supported(name) {
		exit.Message(reason.DrvUnsupportedOS, "The driver '{{.driver}}' is not supported on {{.os}}/{{.arch}}", out.V{"driver": name, "os": os, "arch": arch})
	}

	// if we are only downloading artifacts for a driver, we can stop validation here
	if viper.GetBool("download-only") {
		return
	}

	st := ds.State
	klog.Infof("status for %s: %+v", name, st)

	if st.NeedsImprovement {
		out.Styled(style.Improvement, `For improved {{.driver}} performance, {{.fix}}`, out.V{"driver": driver.FullName(ds.Name), "fix": translate.T(st.Fix)})
	}

	if ds.Priority == registry.Obsolete {
		exit.Message(reason.Kind{
			ID:       fmt.Sprintf("PROVIDER_%s_OBSOLETE", strings.ToUpper(name)),
			Advice:   translate.T(st.Fix),
			ExitCode: reason.ExProviderUnsupported,
			URL:      st.Doc,
			Style:    style.Shrug,
		}, st.Error.Error())
	}

	if st.Error == nil {
		return
	}

	r := reason.MatchKnownIssue(reason.Kind{}, st.Error, runtime.GOOS)
	if r != nil && r.ID != "" {
		exitIfNotForced(*r, st.Error.Error())
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

	id := st.Reason
	if id == "" {
		id = fmt.Sprintf("PROVIDER_%s_ERROR", strings.ToUpper(name))
	}

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
	klog.Infof("selecting image repository for country %s ...", mirrorCountry)

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

	for _, code := range tryCountries {
		localRepos := constants.ImageRepositories[code]
		for _, repo := range localRepos {
			err := checkRepository(v, repo)
			if err == nil {
				return true, repo, nil
			}
		}
	}

	return false, fallback, nil
}

var checkRepository = func(v semver.Version, repo string) error {
	pauseImage := images.Pause(v, repo)
	ref, err := name.ParseReference(pauseImage, name.WeakValidation)
	if err != nil {
		return err
	}

	_, err = remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	return err
}

// validateUser validates minikube is run by the recommended user (privileged or regular)
func validateUser(drvName string) {
	u, err := user.Current()
	if err != nil {
		klog.Errorf("Error getting the current user: %v", err)
		return
	}

	useForce := viper.GetBool(force)

	// None driver works with root and without root on Linux
	if runtime.GOOS == "linux" && drvName == driver.None {
		if !viper.GetBool(interactive) {
			test := exec.Command("sudo", "-n", "echo", "-n")
			if err := test.Run(); err != nil {
				exit.Message(reason.DrvNeedsRoot, `sudo requires a password, and --interactive=false`)
			}
		}
		return
	}

	// If we are not root, exit early
	if u.Uid != "0" {
		return
	}

	out.ErrT(style.Stopped, `The "{{.driver_name}}" driver should not be used with root privileges. If you wish to continue as root, use --force.`, out.V{"driver_name": drvName})
	out.ErrT(style.Tip, "If you are running minikube within a VM, consider using --driver=none:")
	out.ErrT(style.Documentation, "  {{.url}}", out.V{"url": "https://minikube.sigs.k8s.io/docs/reference/drivers/none/"})

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
	info, cpuErr, memErr, diskErr := machine.LocalHostInfo()
	if cpuErr != nil {
		klog.Warningf("could not get system cpu info while verifying memory limits, which might be okay: %v", cpuErr)
	}
	if diskErr != nil {
		klog.Warningf("could not get system disk info while verifying memory limits, which might be okay: %v", diskErr)
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
		containerLimit = util.ConvertBytesToMB(s.TotalMemory)
	}

	return sysLimit, containerLimit, nil
}

// suggestMemoryAllocation calculates the default memory footprint in MiB.
func suggestMemoryAllocation(sysLimit, containerLimit, nodes int) int {
	if mem := viper.GetInt(memory); mem != 0 && mem < sysLimit {
		return mem
	}

	const fallback = 2200
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
		klog.Warningf("Unable to query memory limits: %v", err)
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
		exitIfNotForced(reason.RsrcInsufficientSysMemory, "System only has {{.size}}MiB available, less than the required {{.req}}MiB for Kubernetes", out.V{"size": sysLimit, "req": minUsableMem})
	}

	// if --memory=no-limit, ignore remaining checks
	if req == 0 && driver.IsKIC(drvName) {
		return
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

	if driver.IsHyperV(drvName) && req%2 == 1 {
		exitIfNotForced(reason.RsrcInvalidHyperVMemory, "Hyper-V requires that memory MB be an even number, {{.memory}}MB was specified, try passing `--memory {{.suggestMemory}}`", out.V{"memory": req, "suggestMemory": req - 1})
	}
}

// validateCPUCount validates the cpu count matches the minimum recommended & not exceeding the available cpu count
func validateCPUCount(drvName string) {
	var availableCPUs int

	cpuCount := getCPUCount(drvName)
	isKIC := driver.IsKIC(drvName)

	if isKIC {
		si, err := oci.CachedDaemonInfo(drvName)
		if err != nil {
			si, err = oci.DaemonInfo(drvName)
			if err != nil {
				exit.Message(reason.Usage, "Ensure your {{.driver_name}} is running and is healthy.", out.V{"driver_name": driver.FullName(drvName)})
			}
		}
		availableCPUs = si.CPUs
	} else {
		ci, err := cpu.Counts(true)
		if err != nil {
			exit.Message(reason.Usage, "Unable to get CPU info: {{.err}}", out.V{"err": err})
		}
		availableCPUs = ci
	}

	if availableCPUs < 2 {
		if drvName == oci.Docker && runtime.GOOS == "darwin" {
			exitIfNotForced(reason.RsrcInsufficientDarwinDockerCores, "Docker Desktop has less than 2 CPUs configured, but Kubernetes requires at least 2 to be available")
		} else if drvName == oci.Docker && runtime.GOOS == "windows" {
			exitIfNotForced(reason.RsrcInsufficientWindowsDockerCores, "Docker Desktop has less than 2 CPUs configured, but Kubernetes requires at least 2 to be available")
		} else {
			exitIfNotForced(reason.RsrcInsufficientCores, "{{.driver_name}} has less than 2 CPUs available, but Kubernetes requires at least 2 to be available", out.V{"driver_name": driver.FullName(viper.GetString("driver"))})
		}
	}

	// if --cpus=no-limit, ignore remaining checks
	if cpuCount == 0 && driver.IsKIC(drvName) {
		return
	}

	if cpuCount < minimumCPUS {
		exitIfNotForced(reason.RsrcInsufficientCores, "Requested cpu count {{.requested_cpus}} is less than the minimum allowed of {{.minimum_cpus}}", out.V{"requested_cpus": cpuCount, "minimum_cpus": minimumCPUS})
	}

	if availableCPUs < cpuCount {
		if driver.IsDockerDesktop(drvName) {
			out.Styled(style.Empty, `- Ensure your {{.driver_name}} daemon has access to enough CPU/memory resources.`, out.V{"driver_name": drvName})
			if runtime.GOOS == "darwin" {
				out.Styled(style.Empty, `- Docs https://docs.docker.com/docker-for-mac/#resources`)
			}
			if runtime.GOOS == "windows" {
				out.String("\n\t")
				out.Styled(style.Empty, `- Docs https://docs.docker.com/docker-for-windows/#resources`)
			}
		}

		exitIfNotForced(reason.RsrcInsufficientCores, "Requested cpu count {{.requested_cpus}} is greater than the available cpus of {{.avail_cpus}}", out.V{"requested_cpus": cpuCount, "avail_cpus": availableCPUs})
	}
}

// validateFlags validates the supplied flags against known bad combinations
func validateFlags(cmd *cobra.Command, drvName string) { //nolint:gocyclo
	if cmd.Flags().Changed(humanReadableDiskSize) {
		err := validateDiskSize(viper.GetString(humanReadableDiskSize))
		if err != nil {
			exitIfNotForced(reason.Usage, "{{.err}}", out.V{"err": err})
		}
	}

	if cmd.Flags().Changed(cpus) {
		if !driver.HasResourceLimits(drvName) {
			out.WarningT("The '{{.name}}' driver does not respect the --cpus flag", out.V{"name": drvName})
		}
	}

	validateCPUCount(drvName)

	if drvName == driver.None && viper.GetBool(noKubernetes) {
		exit.Message(reason.Usage, "Cannot use the option --no-kubernetes on the {{.name}} driver", out.V{"name": drvName})
	}

	if cmd.Flags().Changed(memory) {
		validateChangedMemoryFlags(drvName)
	}

	if cmd.Flags().Changed(listenAddress) {
		validateListenAddress(viper.GetString(listenAddress))
	}

	if cmd.Flags().Changed(imageRepository) {
		viper.Set(imageRepository, validateImageRepository(viper.GetString(imageRepository)))
	}

	if cmd.Flags().Changed(ports) {
		err := validatePorts(viper.GetStringSlice(ports))
		if err != nil {
			exit.Message(reason.Usage, "{{.err}}", out.V{"err": err})
		}

	}

	if cmd.Flags().Changed(subnet) {
		err := validateSubnet(viper.GetString(subnet))
		if err != nil {
			exit.Message(reason.Usage, "{{.err}}", out.V{"err": err})
		}
	}

	if cmd.Flags().Changed(containerRuntime) {
		err := validateRuntime(viper.GetString(containerRuntime))
		if err != nil {
			exit.Message(reason.Usage, "{{.err}}", out.V{"err": err})
		}
		validateCNI(cmd, viper.GetString(containerRuntime))
	}

	if cmd.Flags().Changed(staticIP) {
		if err := validateStaticIP(viper.GetString(staticIP), drvName, viper.GetString(subnet)); err != nil {
			exit.Message(reason.Usage, "{{.err}}", out.V{"err": err})
		}
	}

	if cmd.Flags().Changed(gpus) {
		if err := validateGPUs(viper.GetString(gpus), drvName, viper.GetString(containerRuntime)); err != nil {
			exit.Message(reason.Usage, "{{.err}}", out.V{"err": err})
		}
	}

	if cmd.Flags().Changed(autoPauseInterval) {
		if err := validateAutoPauseInterval(viper.GetDuration(autoPauseInterval)); err != nil {
			exit.Message(reason.Usage, "{{.err}}", out.V{"err": err})
		}
	}

	if driver.IsSSH(drvName) {
		sshIPAddress := viper.GetString(sshIPAddress)
		if sshIPAddress == "" {
			exit.Message(reason.Usage, "No IP address provided. Try specifying --ssh-ip-address, or see https://minikube.sigs.k8s.io/docs/drivers/ssh/")
		}

		if net.ParseIP(sshIPAddress) == nil {
			_, err := net.LookupIP(sshIPAddress)
			if err != nil {
				exit.Error(reason.Usage, "Could not resolve IP address", err)
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

	if outputFormat != "text" && outputFormat != "json" {
		exit.Message(reason.Usage, "Sorry, please set the --output flag to one of the following valid options: [text,json]")
	}

	validateBareMetal(drvName)
	validateRegistryMirror()
	validateInsecureRegistry()
}

// validatePorts validates that the --ports are not outside range
func validatePorts(ports []string) error {
	var exposedPorts, hostPorts, portSpecs []string
	for _, p := range ports {
		if strings.Contains(p, ":") {
			portSpecs = append(portSpecs, p)
		} else {
			exposedPorts = append(exposedPorts, p)
		}
	}
	_, portBindingsMap, err := nat.ParsePortSpecs(portSpecs)
	if err != nil {
		return errors.Errorf("Sorry, one of the ports provided with --ports flag is not valid %s (%v)", ports, err)
	}
	for exposedPort, portBindings := range portBindingsMap {
		exposedPorts = append(exposedPorts, exposedPort.Port())
		for _, portBinding := range portBindings {
			hostPorts = append(hostPorts, portBinding.HostPort)
		}
	}
	for _, p := range exposedPorts {
		if err := validatePort(p); err != nil {
			return err
		}
	}
	for _, p := range hostPorts {
		if err := validatePort(p); err != nil {
			return err
		}
	}
	return nil
}

func validatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return errors.Errorf("Sorry, one of the ports provided with --ports flag is not valid: %s", port)
	}
	if p > 65535 || p < 1 {
		return errors.Errorf("Sorry, one of the ports provided with --ports flag is outside range: %s", port)
	}
	return nil
}

// validateDiskSize validates the supplied disk size
func validateDiskSize(diskSize string) error {
	diskSizeMB, err := util.CalculateSizeInMB(diskSize)
	if err != nil {
		return errors.Errorf("Validation unable to parse disk size %v: %v", diskSize, err)
	}
	if diskSizeMB < minimumDiskSize {
		return errors.Errorf("Requested disk size %v is less than minimum of %v", diskSizeMB, minimumDiskSize)
	}
	return nil
}

// validateRuntime validates the supplied runtime
func validateRuntime(rtime string) error {
	validOptions := cruntime.ValidRuntimes()
	// `crio` is accepted as an alternative spelling to `cri-o`
	validOptions = append(validOptions, constants.CRIO)

	if rtime == constants.DefaultContainerRuntime {
		return nil
	}

	var validRuntime bool
	for _, option := range validOptions {
		if rtime == option {
			validRuntime = true
		}

		// Convert `cri-o` to `crio` as the K8s config uses the `crio` spelling
		if rtime == "cri-o" {
			viper.Set(containerRuntime, constants.CRIO)
		}

	}

	if (rtime == "crio" || rtime == "cri-o") && (strings.HasPrefix(runtime.GOARCH, "ppc64") || detect.RuntimeArch() == "arm" || strings.HasPrefix(detect.RuntimeArch(), "arm/")) {
		return errors.Errorf("The %s runtime is not compatible with the %s architecture. See https://github.com/cri-o/cri-o/issues/2467 for more details", rtime, runtime.GOARCH)
	}

	if !validRuntime {
		return errors.Errorf("Invalid Container Runtime: %s. Valid runtimes are: %s", rtime, cruntime.ValidRuntimes())
	}
	return nil
}

// validateGPUs validates that a valid option was given, and if so, can it be used with the given configuration
func validateGPUs(value, drvName, rtime string) error {
	if value == "" {
		return nil
	}
	if err := validateGPUsArch(); err != nil {
		return err
	}
	if value != "nvidia" && value != "all" && value != "amd" {
		return errors.Errorf(`The gpus flag must be passed a value of "nvidia", "amd" or "all"`)
	}
	if drvName == constants.Docker && (rtime == constants.Docker || rtime == constants.DefaultContainerRuntime) {
		return nil
	}
	return errors.Errorf("The gpus flag can only be used with the docker driver and docker container-runtime")
}

func validateGPUsArch() error {
	switch runtime.GOARCH {
	case "amd64", "arm64", "ppc64le":
		return nil
	}
	return errors.Errorf("The GPUs flag is only supported on amd64, arm64 & ppc64le, currently using %s", runtime.GOARCH)
}

func validateAutoPauseInterval(interval time.Duration) error {
	if interval != interval.Abs() || interval.String() == "0s" {
		return errors.New("auto-pause-interval must be greater than 0s")
	}
	return nil
}

func getContainerRuntime(old *config.ClusterConfig) string {
	paramRuntime := viper.GetString(containerRuntime)

	// try to load the old version first if the user didn't specify anything
	if paramRuntime == constants.DefaultContainerRuntime && old != nil {
		paramRuntime = old.KubernetesConfig.ContainerRuntime
	}

	if paramRuntime == constants.DefaultContainerRuntime {
		paramRuntime = defaultRuntime()
	}

	return paramRuntime
}

// defaultRuntime returns the default container runtime
func defaultRuntime() string {
	// minikube default
	return constants.Docker
}

// if container runtime is not docker, check that cni is not disabled
func validateCNI(cmd *cobra.Command, runtime string) {
	if runtime == constants.Docker {
		return
	}
	if cmd.Flags().Changed(cniFlag) && strings.ToLower(viper.GetString(cniFlag)) == "false" {
		if viper.GetBool(force) {
			out.WarnReason(reason.Usage, "You have chosen to disable the CNI but the \"{{.name}}\" container runtime requires CNI", out.V{"name": runtime})
		} else {
			exit.Message(reason.Usage, "The \"{{.name}}\" container runtime requires CNI", out.V{"name": runtime})
		}
	}
}

// validateChangedMemoryFlags validates memory related flags.
func validateChangedMemoryFlags(drvName string) {
	if driver.IsKIC(drvName) && !oci.HasMemoryCgroup() {
		out.WarningT("Your cgroup does not allow setting memory.")
		out.Infof("More information: https://docs.docker.com/engine/install/linux-postinstall/#your-kernel-does-not-support-cgroup-swap-limit-capabilities")
	}
	if !driver.HasResourceLimits(drvName) {
		out.WarningT("The '{{.name}}' driver does not respect the --memory flag", out.V{"name": drvName})
	}
	var req int
	var err error
	memString := viper.GetString(memory)
	if memString == constants.NoLimit && driver.IsKIC(drvName) {
		req = 0
	} else if memString == constants.MaxResources {
		sysLimit, containerLimit, err := memoryLimits(drvName)
		if err != nil {
			klog.Warningf("Unable to query memory limits: %+v", err)
		}
		req = noLimitMemory(sysLimit, containerLimit, drvName)
	} else {
		if memString == constants.NoLimit {
			exit.Message(reason.Usage, "The '{{.name}}' driver does not support --memory=no-limit", out.V{"name": drvName})
		}
		req, err = util.CalculateSizeInMB(memString)
		if err != nil {
			exitIfNotForced(reason.Usage, "Unable to parse memory '{{.memory}}': {{.error}}", out.V{"memory": memString, "error": err})
		}
	}
	validateRequestedMemorySize(req, drvName)
}

func noLimitMemory(sysLimit, containerLimit int, drvName string) int {
	if containerLimit != 0 {
		return containerLimit
	}
	// Recommend 1GB to handle OS/VM overhead
	sysOverhead := 1024
	if driver.IsVirtualBox(drvName) {
		// VirtualBox fully allocates all requested memory on start, it doesn't dynamically allocate when needed like other drivers
		// Because of this allow more system overhead to prevent out of memory issues
		sysOverhead = 1536
	}
	mem := sysLimit - sysOverhead
	// Hyper-V requires an even number of MB, so if odd remove one MB
	if driver.IsHyperV(drvName) && mem%2 == 1 {
		mem--
	}
	return mem
}

// This function validates if the --registry-mirror
// args match the format of http://localhost
func validateRegistryMirror() {
	if len(registryMirror) > 0 {
		for _, loc := range registryMirror {
			URL, err := url.Parse(loc)
			if err != nil {
				klog.Errorln("Error Parsing URL: ", err)
			}
			if (URL.Scheme != "http" && URL.Scheme != "https") || URL.Path != "" {
				exit.Message(reason.Usage, "Sorry, the url provided with the --registry-mirror flag is invalid: {{.url}}", out.V{"url": loc})
			}

		}
	}
}

// This function validates if the --image-repository
// args match the format of registry.cn-hangzhou.aliyuncs.com/google_containers
// also "<hostname>[:<port>]"
func validateImageRepository(imageRepo string) (validImageRepo string) {
	expression := regexp.MustCompile(`^(?:(\w+)\:\/\/)?([-a-zA-Z0-9]{1,}(?:\.[-a-zA-Z0-9]{1,}){0,})(?:\:(\d+))?(\/.*)?$`)

	if strings.ToLower(imageRepo) == "auto" {
		imageRepo = "auto"
	}

	if !expression.MatchString(imageRepo) {
		klog.Errorln("Provided repository is not a valid URL. Defaulting to \"auto\"")
		imageRepo = "auto"
	}

	var imageRepoPort string
	groups := expression.FindStringSubmatch(imageRepo)

	scheme := groups[1]
	hostname := groups[2]
	port := groups[3]
	path := groups[4]

	if port != "" && strings.Contains(imageRepo, ":"+port) {
		imageRepoPort = ":" + port
	}

	// tips when imageRepo ended with a trailing /.
	if strings.HasSuffix(imageRepo, "/") {
		out.Infof("The --image-repository flag your provided ended with a trailing / that could cause conflict in kubernetes, removed automatically")
	}
	// tips when imageRepo started with scheme such as http(s).
	if scheme != "" {
		out.Infof("The --image-repository flag you provided contains Scheme: {{.scheme}}, which will be removed automatically", out.V{"scheme": scheme})
	}

	validImageRepo = hostname + imageRepoPort + strings.TrimSuffix(path, "/")

	return validImageRepo
}

// This function validates if the --listen-address
// match the format 0.0.0.0
func validateListenAddress(listenAddr string) {
	if len(listenAddr) > 0 && net.ParseIP(listenAddr) == nil {
		exit.Message(reason.Usage, "Sorry, the IP provided with the --listen-address flag is invalid: {{.listenAddr}}.", out.V{"listenAddr": listenAddr})
	}
}

// This function validates that the --insecure-registry follows one of the following formats:
// "<ip>[:<port>]" "<hostname>[:<port>]" "<network>/<netmask>"
func validateInsecureRegistry() {
	if len(insecureRegistry) > 0 {
		for _, addr := range insecureRegistry {
			// Remove http or https from registryMirror
			if strings.HasPrefix(strings.ToLower(addr), "http://") || strings.HasPrefix(strings.ToLower(addr), "https://") {
				i := strings.Index(addr, "//")
				addr = addr[i+2:]
			} else if strings.Contains(addr, "://") || strings.HasSuffix(addr, ":") {
				exit.Message(reason.Usage, "Sorry, the address provided with the --insecure-registry flag is invalid: {{.addr}}. Expected formats are: <ip>[:<port>], <hostname>[:<port>] or <network>/<netmask>", out.V{"addr": addr})
			}
			hostnameOrIP, port, err := net.SplitHostPort(addr)
			if err != nil {
				_, _, err := net.ParseCIDR(addr)
				if err == nil {
					continue
				}
				hostnameOrIP = addr
			}
			if !hostRe.MatchString(hostnameOrIP) && net.ParseIP(hostnameOrIP) == nil {
				//		fmt.Printf("This is not hostname or ip %s", hostnameOrIP)
				exit.Message(reason.Usage, "Sorry, the address provided with the --insecure-registry flag is invalid: {{.addr}}. Expected formats are: <ip>[:<port>], <hostname>[:<port>] or <network>/<netmask>", out.V{"addr": addr})
			}
			if port != "" {
				v, err := strconv.Atoi(port)
				if err != nil {
					exit.Message(reason.Usage, "Sorry, the address provided with the --insecure-registry flag is invalid: {{.addr}}. Expected formats are: <ip>[:<port>], <hostname>[:<port>] or <network>/<netmask>", out.V{"addr": addr})
				}
				if v < 0 || v > 65535 {
					exit.Message(reason.Usage, "Sorry, the address provided with the --insecure-registry flag is invalid: {{.addr}}. Expected formats are: <ip>[:<port>], <hostname>[:<port>] or <network>/<netmask>", out.V{"addr": addr})
				}
			}
		}
	}
}

// configureNodes creates primary control-plane node config on first cluster start or updates existing cluster nodes configs on restart.
// It will return updated cluster config and primary control-plane node or any error occurred.
func configureNodes(cc config.ClusterConfig, existing *config.ClusterConfig) (config.ClusterConfig, config.Node, error) {
	kv, err := getKubernetesVersion(&cc)
	if err != nil {
		return cc, config.Node{}, errors.Wrapf(err, "failed getting kubernetes version")
	}
	cr := getContainerRuntime(&cc)

	// create the initial node, which will necessarily be primary control-plane node
	if existing == nil {
		pcp := config.Node{
			Port:              cc.APIServerPort,
			KubernetesVersion: kv,
			ContainerRuntime:  cr,
			ControlPlane:      true,
			Worker:            true,
		}
		cc.Nodes = []config.Node{pcp}
		return cc, pcp, nil
	}

	// Make sure that existing nodes honor if KubernetesVersion gets specified on restart
	// KubernetesVersion is the only attribute that the user can override in the Node object
	nodes := []config.Node{}
	for _, n := range existing.Nodes {
		n.KubernetesVersion = kv
		n.ContainerRuntime = cr
		nodes = append(nodes, n)
	}
	cc.Nodes = nodes

	pcp, err := config.ControlPlane(*existing)
	if err != nil {
		return cc, config.Node{}, errors.Wrapf(err, "failed getting control-plane node")
	}
	pcp.KubernetesVersion = kv
	pcp.ContainerRuntime = cr

	return cc, pcp, nil
}

// autoSetDriverOptions sets the options needed for specific driver automatically.
func autoSetDriverOptions(cmd *cobra.Command, drvName string) (err error) {
	err = nil
	hints := driver.FlagDefaults(drvName)
	if len(hints.ExtraOptions) > 0 {
		for _, eo := range hints.ExtraOptions {
			if config.ExtraOptions.Exists(eo) {
				klog.Infof("skipping extra-config %q.", eo)
				continue
			}
			klog.Infof("auto setting extra-config to %q.", eo)
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
		klog.Infof("auto set %s to %q.", containerRuntime, hints.ContainerRuntime)
	}

	if !cmd.Flags().Changed(cmdcfg.Bootstrapper) && hints.Bootstrapper != "" {
		viper.Set(cmdcfg.Bootstrapper, hints.Bootstrapper)
		klog.Infof("auto set %s to %q.", cmdcfg.Bootstrapper, hints.Bootstrapper)

	}

	return err
}

// validateKubernetesVersion ensures that the requested version is reasonable
func validateKubernetesVersion(old *config.ClusterConfig) {
	paramVersion := viper.GetString(kubernetesVersion)
	paramVersion = strings.TrimPrefix(strings.ToLower(paramVersion), version.VersionPrefix)
	kubernetesVer, err := getKubernetesVersion(old)
	if err != nil {
		if errors.Is(err, ErrKubernetesPatchNotFound) {
			exit.Message(reason.PatchNotFound, "Unable to detect the latest patch release for specified major.minor version v{{.majorminor}}",
				out.V{"majorminor": paramVersion})
		}
		exit.Message(reason.Usage, `Unable to parse "{{.kubernetes_version}}": {{.error}}`, out.V{"kubernetes_version": paramVersion, "error": err})

	}

	nvs, _ := semver.Make(strings.TrimPrefix(kubernetesVer, version.VersionPrefix))
	oldestVersion := semver.MustParse(strings.TrimPrefix(constants.OldestKubernetesVersion, version.VersionPrefix))
	defaultVersion := semver.MustParse(strings.TrimPrefix(constants.DefaultKubernetesVersion, version.VersionPrefix))
	newestVersion := semver.MustParse(strings.TrimPrefix(constants.NewestKubernetesVersion, version.VersionPrefix))
	zeroVersion := semver.MustParse(strings.TrimPrefix(constants.NoKubernetesVersion, version.VersionPrefix))

	if isTwoDigitSemver(paramVersion) && getLatestPatch(paramVersion) != "" {
		out.Styled(style.Workaround, `Using Kubernetes {{.version}} since patch version was unspecified`, out.V{"version": nvs})
	}
	if nvs.Equals(zeroVersion) {
		klog.Infof("No Kubernetes version set for minikube, setting Kubernetes version to %s", constants.NoKubernetesVersion)
		return
	}
	if nvs.Major > newestVersion.Major {
		out.WarningT("Specified Major version of Kubernetes {{.specifiedMajor}} is newer than the newest supported Major version: {{.newestMajor}}", out.V{"specifiedMajor": nvs.Major, "newestMajor": newestVersion.Major})
		if !viper.GetBool(force) {
			out.WarningT("You can force an unsupported Kubernetes version via the --force flag")
		}
		exitIfNotForced(reason.KubernetesTooNew, "Kubernetes {{.version}} is not supported by this release of minikube", out.V{"version": nvs})
	}
	if nvs.GT(newestVersion) {
		out.WarningT("Specified Kubernetes version {{.specified}} is newer than the newest supported version: {{.newest}}. Use `minikube config defaults kubernetes-version` for details.", out.V{"specified": nvs, "newest": constants.NewestKubernetesVersion})
		if contains(constants.ValidKubernetesVersions, kubernetesVer) {
			out.Styled(style.Check, "Kubernetes version {{.specified}} found in version list", out.V{"specified": nvs})
		} else {
			out.WarningT("Specified Kubernetes version {{.specified}} not found in Kubernetes version list", out.V{"specified": nvs})
			out.Styled(style.Verifying, "Searching the internet for Kubernetes version...")
			found, err := cmdcfg.IsInGitHubKubernetesVersions(kubernetesVer)
			if err != nil && !viper.GetBool(force) {
				exit.Error(reason.KubernetesNotConnect, "error fetching Kubernetes version list from GitHub", err)
			}
			if found {
				out.Styled(style.Check, "Kubernetes version {{.specified}} found in GitHub version list", out.V{"specified": nvs})
			} else if !viper.GetBool(force) {
				out.WarningT("Kubernetes version not found in GitHub version list. You can force a Kubernetes version via the --force flag")
				exitIfNotForced(reason.KubernetesTooNew, "Kubernetes version {{.version}} is not supported by this release of minikube", out.V{"version": nvs})
			}
		}
	}
	if nvs.LT(oldestVersion) {
		out.WarningT("Specified Kubernetes version {{.specified}} is less than the oldest supported version: {{.oldest}}. Use `minikube config defaults kubernetes-version` for details.", out.V{"specified": nvs, "oldest": constants.OldestKubernetesVersion})
		if !viper.GetBool(force) {
			out.WarningT("You can force an unsupported Kubernetes version via the --force flag")
		}
		exitIfNotForced(reason.KubernetesTooOld, "Kubernetes {{.version}} is not supported by this release of minikube", out.V{"version": nvs})
	}

	// If the version of Kubernetes has a known issue, print a warning out to the screen
	if issue := reason.ProblematicK8sVersion(nvs); issue != nil {
		out.WarningT(issue.Description, out.V{"version": nvs.String()})
		if issue.URL != "" {
			out.WarningT("For more information, see: {{.url}}", out.V{"url": issue.URL})
		}
	}

	if old == nil || old.KubernetesConfig.KubernetesVersion == "" {
		return
	}

	ovs, err := semver.Make(strings.TrimPrefix(old.KubernetesConfig.KubernetesVersion, version.VersionPrefix))
	if err != nil {
		klog.Errorf("Error parsing old version %q: %v", old.KubernetesConfig.KubernetesVersion, err)
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
		out.Styled(style.New, "Kubernetes {{.new}} is now available. If you would like to upgrade, specify: --kubernetes-version={{.prefix}}{{.new}}", out.V{"prefix": version.VersionPrefix, "new": defaultVersion})
	}
}

// validateContainerRuntime ensures that the container runtime is reasonable
func validateContainerRuntime(old *config.ClusterConfig) {
	if old == nil || old.KubernetesConfig.ContainerRuntime == "" {
		return
	}

	if err := validateRuntime(old.KubernetesConfig.ContainerRuntime); err != nil {
		klog.Errorf("Error parsing old runtime %q: %v", old.KubernetesConfig.ContainerRuntime, err)
	}
}

func isBaseImageApplicable(drv string) bool {
	return registry.IsKIC(drv)
}

func getKubernetesVersion(old *config.ClusterConfig) (string, error) {
	if viper.GetBool(noKubernetes) {
		// Exit if --kubernetes-version is specified.
		if viper.GetString(kubernetesVersion) != "" {
			exit.Message(reason.Usage, `cannot specify --kubernetes-version with --no-kubernetes,
to unset a global config run:

$ minikube config unset kubernetes-version`)
		}

		klog.Infof("No Kubernetes flag is set, setting Kubernetes version to %s", constants.NoKubernetesVersion)
		if old != nil {
			old.KubernetesConfig.KubernetesVersion = constants.NoKubernetesVersion
		}
	}

	paramVersion := viper.GetString(kubernetesVersion)

	// try to load the old version first if the user didn't specify anything
	if paramVersion == "" && old != nil {
		paramVersion = old.KubernetesConfig.KubernetesVersion
	}

	if paramVersion == "" || strings.EqualFold(paramVersion, "stable") {
		paramVersion = constants.DefaultKubernetesVersion
	} else if strings.EqualFold(strings.ToLower(paramVersion), "latest") || strings.EqualFold(strings.ToLower(paramVersion), "newest") {
		paramVersion = constants.NewestKubernetesVersion
	}

	kubernetesSemver := strings.TrimPrefix(strings.ToLower(paramVersion), version.VersionPrefix)
	if isTwoDigitSemver(kubernetesSemver) {
		potentialPatch := getLatestPatch(kubernetesSemver)
		if potentialPatch == "" {
			return "", ErrKubernetesPatchNotFound
		}
		kubernetesSemver = potentialPatch
	}
	nvs, err := semver.Make(kubernetesSemver)
	if err != nil {
		exit.Message(reason.Usage, `Unable to parse "{{.kubernetes_version}}": {{.error}}`, out.V{"kubernetes_version": paramVersion, "error": err})
	}

	return version.VersionPrefix + nvs.String(), nil
}

// validateDockerStorageDriver checks that docker is using overlay2
// if not, set preload=false (see #7626)
func validateDockerStorageDriver(drvName string) {
	if !driver.IsKIC(drvName) {
		return
	}
	if _, err := exec.LookPath(drvName); err != nil {
		exit.Error(reason.DrvNotFound, fmt.Sprintf("%s not found on PATH", drvName), err)
	}
	si, err := oci.DaemonInfo(drvName)
	if err != nil {
		klog.Warningf("Unable to confirm that %s is using overlay2 storage driver; setting preload=false", drvName)
		viper.Set(preload, false)
		return
	}
	if si.StorageDriver == "overlay2" || si.StorageDriver == "overlayfs" {
		return
	}
	out.WarningT("{{.Driver}} is currently using the {{.StorageDriver}} storage driver, setting preload=false", out.V{"StorageDriver": si.StorageDriver, "Driver": drvName})
	viper.Set(preload, false)
}

// validateSubnet checks that the subnet provided has a private IP
// and does not have a mask of more that /30
func validateSubnet(subnet string) error {
	ip, cidr, err := netutil.ParseAddr(subnet)
	if err != nil {
		return errors.Errorf("Sorry, unable to parse subnet: %v", err)
	}
	if !ip.IsPrivate() {
		return errors.Errorf("Sorry, the subnet %s is not a private IP", ip)
	}

	if cidr != nil {
		mask, _ := cidr.Mask.Size()
		if mask > 30 {
			return errors.Errorf("Sorry, the subnet provided does not have a mask less than or equal to /30")
		}
	}
	return nil
}

func validateStaticIP(staticIP, drvName, subnet string) error {
	if !driver.IsKIC(drvName) {
		if staticIP != "" {
			out.WarningT("--static-ip is only implemented on Docker and Podman drivers, flag will be ignored")
		}
		return nil
	}
	if subnet != "" {
		out.WarningT("--static-ip overrides --subnet, --subnet will be ignored")
	}
	ip := net.ParseIP(staticIP)
	if !ip.IsPrivate() {
		return fmt.Errorf("static IP must be private")
	}
	if ip.To4() == nil {
		return fmt.Errorf("static IP must be IPv4")
	}
	lastOctet, _ := strconv.Atoi(strings.Split(ip.String(), ".")[3])
	if lastOctet < 2 || lastOctet > 254 {
		return fmt.Errorf("static IPs last octet must be between 2 and 254 (X.X.X.2 - X.X.X.254), for example 192.168.200.200")
	}
	return nil
}

func validateBareMetal(drvName string) {
	if !driver.BareMetal(drvName) {
		return
	}

	if viper.GetInt(nodes) > 1 || viper.GetBool(ha) {
		exit.Message(reason.DrvUnsupportedMulti, "The none driver is not compatible with multi-node clusters.")
	}

	if ClusterFlagValue() != constants.DefaultClusterName {
		exit.Message(reason.DrvUnsupportedProfile, "The '{{.name}} driver does not support multiple profiles: https://minikube.sigs.k8s.io/docs/reference/drivers/none/", out.V{"name": drvName})
	}

	// default container runtime varies, starting with Kubernetes 1.24 - assume that only the default container runtime has been tested
	rtime := viper.GetString(containerRuntime)
	if rtime != constants.DefaultContainerRuntime && rtime != defaultRuntime() {
		out.WarningT("Using the '{{.runtime}}' runtime with the 'none' driver is an untested configuration!", out.V{"runtime": rtime})
	}

	// conntrack is required starting with Kubernetes 1.18, include the release candidates for completion
	kubeVer, err := getKubernetesVersion(nil)
	if err != nil {
		klog.Warningf("failed getting Kubernetes version: %v", err)
	}
	version, _ := util.ParseKubernetesVersion(kubeVer)
	if version.GTE(semver.MustParse("1.18.0-beta.1")) {
		if _, err := exec.LookPath("conntrack"); err != nil {
			exit.Message(reason.GuestMissingConntrack, "Sorry, Kubernetes {{.k8sVersion}} requires conntrack to be installed in root's path", out.V{"k8sVersion": version.String()})
		}
	}
	// crictl is required starting with Kubernetes 1.24, for all runtimes since the removal of dockershim
	if version.GTE(semver.MustParse("1.24.0-alpha.0")) {
		if _, err := exec.LookPath("crictl"); err != nil {
			exit.Message(reason.GuestMissingConntrack, "Sorry, Kubernetes {{.k8sVersion}} requires crictl to be installed in root's path", out.V{"k8sVersion": version.String()})
		}
	}
}

func exitIfNotForced(r reason.Kind, message string, v ...out.V) {
	if !viper.GetBool(force) {
		exit.Message(r, message, v...)
	}
	out.Error(r, message, v...)
}

func exitGuestProvision(err error) {
	if errors.Cause(err) == oci.ErrInsufficientDockerStorage {
		exit.Message(reason.RsrcInsufficientDockerStorage, "preload extraction failed: \"No space left on device\"")
	}
	if errors.Cause(err) == oci.ErrGetSSHPortContainerNotRunning {
		exit.Message(reason.GuestProvisionContainerExited, "Docker container exited prematurely after it was created, consider investigating Docker's performance/health.")
	}
	exit.Error(reason.GuestProvision, "error provisioning guest", err)
}

// Example input = 1.26 => output = "1.26.5"
// Example input = 1.26.5 => output = "1.26.5"
// Example input = 1.26.999 => output = ""
func getLatestPatch(majorMinorVer string) string {
	for _, k := range constants.ValidKubernetesVersions {
		if strings.HasPrefix(k, fmt.Sprintf("v%s.", majorMinorVer)) {
			return strings.TrimPrefix(k, version.VersionPrefix)
		}

	}
	return ""
}

func isTwoDigitSemver(ver string) bool {
	majorMinorOnly := regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)$`)
	return majorMinorOnly.MatchString(ver)
}

func startNerdctld() {
	// for containerd runtime using ssh, we have installed nerdctld and nerdctl into kicbase
	// These things will be included in the ISO/Base image in the future versions

	// copy these binaries to the path of the containerd node
	co := mustload.Running(ClusterFlagValue())
	runner := co.CP.Runner

	// and set 777 to these files
	if out, err := runner.RunCmd(exec.Command("sudo", "chmod", "777", "/usr/local/bin/nerdctl", "/usr/local/bin/nerdctld")); err != nil {
		exit.Error(reason.StartNerdctld, fmt.Sprintf("Failed setting permission for nerdctl: %s", out.Output()), err)
	}

	// sudo systemctl start nerdctld.socket
	if out, err := runner.RunCmd(exec.Command("sudo", "systemctl", "start", "nerdctld.socket")); err != nil {
		exit.Error(reason.StartNerdctld, fmt.Sprintf("Failed to enable nerdctld.socket: %s", out.Output()), err)
	}
	// sudo systemctl start nerdctld.service
	if out, err := runner.RunCmd(exec.Command("sudo", "systemctl", "start", "nerdctld.service")); err != nil {
		exit.Error(reason.StartNerdctld, fmt.Sprintf("Failed to enable nerdctld.service: %s", out.Output()), err)
	}

	// set up environment variable on remote machine. docker client uses 'non-login & non-interactive shell' therefore the only way is to modify .bashrc file of user 'docker'
	// insert this at 4th line
	envSetupCommand := exec.Command("/bin/bash", "-c", "sed -i '4i export DOCKER_HOST=unix:///run/nerdctld.sock' .bashrc")
	if out, err := runner.RunCmd(envSetupCommand); err != nil {
		exit.Error(reason.StartNerdctld, fmt.Sprintf("Failed to set up DOCKER_HOST: %s", out.Output()), err)
	}
}

// contains checks whether the parameter slice contains the parameter string
func contains(sl []string, s string) bool {
	for _, k := range sl {
		if s == k {
			return true
		}

	}
	return false
}
