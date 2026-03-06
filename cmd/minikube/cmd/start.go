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
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"errors"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/blang/semver/v4"
	"github.com/docker/go-connections/nat"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/shirou/gopsutil/v4/cpu"
	gopshost "github.com/shirou/gopsutil/v4/host"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/libmachine/ssh"

	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/cmd/minikube/cmd/flags"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/driver/auxdriver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/firewall"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/pause"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/util"
	"k8s.io/minikube/pkg/minikube/version"
	"k8s.io/minikube/pkg/util/kubeconfig"
	"k8s.io/minikube/pkg/util/lock"
	kconst "k8s.io/minikube/pkg/util/lock"
	netutil "k8s.io/minikube/pkg/util/net"
	pkgtrace "k8s.io/minikube/pkg/util/trace"
)

const (
	noKubernetes                = "no-kubernetes"
	kubernetesVersion           = "kubernetes-version"
	containerRuntime            = "container-runtime"
	nodes                       = "nodes"
	ha                          = "ha"
	deleteOnFailure             = "delete-on-failure"
	force                       = "force"
	dryRun                      = "dry-run"
	interactive                 = "interactive"
	autoUpdate                  = "auto-update"
	cpus                        = "cpus"
	memory                      = "memory"
	humanReadableDiskSize       = "disk-size"
	defaultDiskSize             = "20000mb"
	kvmQemuURI                  = "kvm-qemu-uri"
	installAddons               = "install-addons"
	nativeSSH                   = "native-ssh"
	minimumCPUS                 = 2
	minimumMemory               = 1800
	isoURL                      = "iso-url"
	kicBaseImage                = "base-image"
	preload                     = "preload"
	imageRepository             = "image-repository"
	imageMirrorCountry          = "image-mirror-country"
	insecureRegistry            = "insecure-registry"
	registryMirror              = "registry-mirror"
)

var (
	registryMirror []string
)

type versionJSON struct {
	MinikubeVersion string `json:"minikube_version"`
}

func runStart(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	displayVersion(version.GetVersion())

	options := &run.CommandOptions{
		Cwd: localpath.MakeMiniPath(""),
	}

	if viper.GetBool(force) {
		out.WarningT("minikube skips various validations when --force is supplied; this may lead to unexpected behavior")
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

	validateSpecifiedDriver(existing, options)
	validateKubernetesVersion(existing)
	validateContainerRuntime(existing)

	ds, alts, specified := selectDriver(existing, options)
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

	starter, err := provisionWithDriver(cmd, ds, existing, options)
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

				err = deleteProfile(ctx, profile, options)
				if err != nil {
					out.WarningT("Failed to delete cluster {{.name}}, proceeding with retry anyway.", out.V{"name": ClusterFlagValue()})
				}
				starter, err = provisionWithDriver(cmd, ds, existing, options)
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

	if existing != nil && driver.IsKIC(existing.Driver) && viper.GetString(mountString) != "" {
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

	configInfo, err := startWithDriver(cmd, starter, existing, options)
	if err != nil {
		node.ExitIfFatal(err, useForce)
		exit.Error(reason.GuestStart, "failed to start node", err)
	}

	if starter.Cfg.VerifyComponents[kverify.ExtraKey] {
		if err := kverify.WaitExtra(ClusterFlagValue(), kverify.CorePodsLabels, kconst.DefaultControlPlaneTimeout); err != nil {
			exit.Message(reason.GuestStart, "extra waiting: {{.error}}", out.V{"error": err})
		}
	}

	if err := showKubectlInfo(configInfo, starter.Node.KubernetesVersion, starter.Node.ContainerRuntime, starter.Cfg.Name); err != nil {
		klog.Errorf("kubectl info: %v", err)
	}
}

func provisionWithDriver(cmd *cobra.Command, ds registry.DriverState, existing *config.ClusterConfig, options *run.CommandOptions) (node.Starter, error) {
	driverName := ds.Name
	klog.Infof("selected driver: %s", driverName)
	validateDriver(ds, existing)
	err := autoSetDriverOptions(cmd, driverName)
	if err != nil {
		klog.Errorf("Error autoSetOptions : %v", err)
	}

	virtualBoxMacOS13PlusWarning(driverName)
	hyperkitDeprecationWarning(driverName)
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
	if rtime == constants.Docker && (existing == nil || viper.IsSet(containerRuntime)) {
		// TODO: remove this warning in minikube v1.40
		out.WarningT(constants.DefaultContainerRuntimeChangeWarning)
	}
	cc, n, err := generateClusterConfig(cmd, existing, k8sVersion, rtime, driverName, options)
	if err != nil {
		return node.Starter{}, fmt.Errorf("Failed to generate cluster config: %w", err)
	}
	klog.Infof("cluster config:\n%+v", cc)

	if firewall.IsBootpdBlocked(cc) {
		if err := firewall.UnblockBootpd(options); err != nil {
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
		urlString, err := download.ISO(viper.GetStringSlice(isoURL), cmd.Flags().Changed(isoURL))
		if err != nil {
			return node.Starter{}, fmt.Errorf("Failed to cache ISO: %w", err)
		}
		cc.MinikubeISO = urlString
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

	mRunner, preExists, mAPI, host, err := node.Provision(&cc, &n, viper.GetBool(deleteOnFailure), options)
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
	out.WarningT(`Due to changes in macOS 13+ minikube doesn't currently support VirtualBox. You can use alternative drivers such as 'vfkit', 'qemu', or 'docker'.
    https://minikube.sigs.k8s.io/docs/drivers/vfkit/
    https://minikube.sigs.k8s.io/docs/drivers/qemu/
    https://minikube.sigs.k8s.io/docs/drivers/docker/
    For more details on the issue see: https://github.com/kubernetes/minikube/issues/15274
`)
}

// hyperkitDeprecationWarning prints a deprecation warning for the hyperkit driver
func hyperkitDeprecationWarning(driverName string) {
	if !driver.IsHyperKit(driverName) {
		return
	}
	out.WarningT(`The 'hyperkit' driver is deprecated and will be removed in a future release.
    You can use alternative drivers such as 'vfkit', 'qemu', or 'docker'.
    https://minikube.sigs.k8s.io/docs/drivers/vfkit/
    https://minikube.sigs.k8s.io/docs/drivers/qemu/
    https://minikube.sigs.k8s.io/docs/drivers/docker/
	`)
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
		// v1.38.1 minikube binary uses v1.38.0 ISO
		"v1.38.1": "v1.38.0",
	}
	binaryVersion, ok := mappedVersions[binaryVersion]

	return ok && binaryVersion == imageVersion
}

func startWithDriver(cmd *cobra.Command, starter node.Starter, existing *config.ClusterConfig, options *run.CommandOptions) (*kubeconfig.Settings, error) {
	// start primary control-plane node
	configInfo, err := node.Start(starter, options)
	if err != nil {
		configInfo, err = maybeDeleteAndRetry(cmd, *starter.Cfg, *starter.Node, starter.ExistingAddons, err, options)
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
		if err := node.Add(starter.Cfg, n, viper.GetBool(deleteOnFailure), options); err != nil {
			return nil, fmt.Errorf("adding node: %w", err)
		}
	}

	pause.RemovePausedFile(starter.Runner)

	return configInfo, nil
}

func warnAboutMultiNodeCNI() {
	out.WarningT("Cluster was created without any CNI, adding a node to it might cause broken networking.")
}

func updateDriver(driverName string) {
	if err := auxdriver.InstallOrUpdate(driverName, localpath.MakeMiniPath("bin"), viper.GetBool(flags.Interactive), viper.GetBool(autoUpdate)); err != nil {
		if errors.Is(err, auxdriver.ErrAuxDriverVersionCommandFailed) {
			exit.Error(reason.DrvAuxNotHealthy, "Aux driver "+driverName, err)
		}
		if errors.Is(err, auxdriver.ErrAuxDriverVersionNotinPath) {
			exit.Error(reason.DrvAuxNotHealthy, "Aux driver"+driverName, err)
		} // if failed to update but not a fatal error, log it and continue (old version might still work)
		out.WarningT("Unable to update {{.driver}} driver: {{.error}}", out.V{"driver": driverName, "error": err})
	}
}

func displayVersion(ver string) {
	prefix := ""
	if ClusterFlagValue() != constants.DefaultClusterName {
		prefix = fmt.Sprintf("[%s] ", ClusterFlagValue())
	}

	register.Reg.SetStep(register.InitialSetup)
	out.Step(style.Happy, "{{.prefix}}minikube {{.version}} on {{.platform}}", out.V{"prefix": prefix, "version": ver, "platform": platform()})
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
		return fmt.Errorf("client semver: %w", err)
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

func maybeDeleteAndRetry(cmd *cobra.Command, existing config.ClusterConfig, n config.Node, existingAddons map[string]bool, originalErr error, options *run.CommandOptions) (*kubeconfig.Settings, error) {
	if viper.GetBool(deleteOnFailure) {
		out.WarningT("Node {{.name}} failed to start, deleting and trying again.", out.V{"name": n.Name})
		// Start failed, delete the cluster and try again
		profile, err := config.LoadProfile(existing.Name)
		if err != nil {
			out.ErrT(style.Meh, `"{{.name}}" profile does not exist, trying anyways.`, out.V{"name": existing.Name})
		}

		err = deleteProfile(context.Background(), profile, options)
		if err != nil {
			out.WarningT("Failed to delete cluster {{.name}}, proceeding with retry anyway.", out.V{"name": existing.Name})
		}

		// Re-generate the cluster config, just in case the failure was related to an old config format
		cc := updateExistingConfigFromFlags(cmd, &existing)
		var configInfo *kubeconfig.Settings

		for _, n := range cc.Nodes {
			r, p, m, h, err := node.Provision(&cc, &n, false, options)
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

			k, err := node.Start(s, options)
			if n.ControlPlane {
				configInfo = k
			}
			if err != nil {
				// Ok we failed again, let's bail
				return nil, err
			}
		}
		return configInfo, nil
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
			return "", fmt.Errorf("exec: %w", err)
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
		return "", fmt.Errorf("unmarshal: %w", err)
	}

	return cv.ClientVersion.GitVersion, nil
}

// returns (current_driver, suggested_drivers, "true, if the driver is set by command line arg or in the config file")
func selectDriver(existing *config.ClusterConfig, options *run.CommandOptions) (registry.DriverState, []registry.DriverState, bool) {
	// Technically unrelated, but important to perform before detection
	driver.SetLibvirtURI(viper.GetString(kvmQemuURI))
	register.Reg.SetStep(register.SelectingDriver)
	// By default, the driver is whatever we used last time
	if existing != nil {
		old := hostDriver(existing, options)
		ds := driver.Status(old, options)
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
		ds := driver.Status(d, options)
		if ds.Name == "" {
			exit.Message(reason.DrvUnsupportedOS, "The driver '{{.driver}}' is not supported on {{.os}}/{{.arch}}", out.V{"driver": d, "os": runtime.GOOS, "arch": runtime.GOARCH})
		}
		out.Step(style.Sparkle, `Using the {{.driver}} driver based on user configuration`, out.V{"driver": ds.String()})
		return ds, nil, true
	}

	// Fallback to old driver parameter
	if d := viper.GetString("vm-driver"); d != "" {
		ds := driver.Status(viper.GetString("vm-driver"), options)
		if ds.Name == "" {
			exit.Message(reason.DrvUnsupportedOS, "The driver '{{.driver}}' is not supported on {{.os}}/{{.arch}}", out.V{"driver": d, "os": runtime.GOOS, "arch": runtime.GOARCH})
		}
		out.Step(style.Sparkle, `Using the {{.driver}} driver based on user configuration`, out.V{"driver": ds.String()})
		return ds, nil, true
	}

	choices := driver.Choices(viper.GetBool("vm"), options)
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
		switch {
		case foundStoppedDocker:
			exit.Message(reason.DrvDockerNotRunning, "Found docker, but the docker service isn't running. Try restarting the docker service.")
		case foundUnhealthy:
			exit.Message(reason.DrvNotHealthy, "Found driver(s) but none were healthy. See above for suggestions how to fix installed drivers.")
		default:
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
func hostDriver(existing *config.ClusterConfig, options *run.CommandOptions) string {
	if existing == nil {
		return ""
	}

	api, err := machine.NewAPIClient(options)
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
				out.WarningT("Profile name '{{.name}}' is duplicated with machine name '{{.machine}}' in profile '{{.profile}}'", out.V{"name": ClusterFlagValue(), "machine": machineName, "profile": p.Name})
				exit.Message(reason.Usage, "Please choose a different profile name.")
			}
		}
	}
}

func validateSpecifiedDriver(existing *config.ClusterConfig, options *run.CommandOptions) {
	if existing == nil {
		return
	}
	if d := viper.GetString("driver"); d != "" && d != existing.Driver {
		exit.Message(reason.DrvUnsupportedProfile, "The '{{.profile}}' profile was created with the '{{.old}}' driver, but you've specified the '{{.new}}' driver. Please use the same driver or create a new profile.", out.V{"profile": ClusterFlagValue(), "old": existing.Driver, "new": d})
	}
}

func validateKubernetesVersion(existing *config.ClusterConfig) {
	if existing == nil {
		return
	}
	if v := viper.GetString(kubernetesVersion); v != "" && v != existing.KubernetesConfig.KubernetesVersion {
		// allow upgrading kubernetes version
		out.Step(style.Sparkle, "Upgrading Kubernetes to {{.version}}", out.V{"version": v})
	}
}

func validateContainerRuntime(existing *config.ClusterConfig) {
	if existing == nil {
		return
	}
	if r := viper.GetString(containerRuntime); r != "" && r != existing.KubernetesConfig.ContainerRuntime {
		exit.Message(reason.Usage, "Sorry, you cannot change the container runtime of an existing cluster. Please create a new profile.")
	}
}

// validateCPUCount validates the cpu count matches the minimum recommended & not exceeding the available cpu count
func validateCPUCount(drvName string) {
	var availableCPUs int

	cpuCount := getCPUCount(drvName)
	isKIC := driver.IsKIC(drvName)
	isNoK8s := viper.GetBool(noKubernetes) // Track if we are running without k8s

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

	switch {
	case availableCPUs < 2 && !isNoK8s: // Only enforce minimum if k8s is required
		switch {
		case drvName == oci.Docker && runtime.GOOS == "darwin":
			exitIfNotForced(reason.RsrcInsufficientDarwinDockerCores, "Docker Desktop has less than 2 CPUs configured, but Kubernetes requires at least 2 to be available")
		case drvName == oci.Docker && runtime.GOOS == "windows":
			exitIfNotForced(reason.RsrcInsufficientWindowsDockerCores, "Docker Desktop has less than 2 CPUs configured, but Kubernetes requires at least 2 to be available")
		default:
			exitIfNotForced(reason.RsrcInsufficientCores, "{{.driver_name}} has less than 2 CPUs available, but Kubernetes requires at least 2 to be available", out.V{"driver_name": driver.FullName(viper.GetString("driver"))})
		}
	}

	// if --cpus=no-limit, ignore remaining checks
	if cpuCount == 0 && driver.IsKIC(drvName) {
		return
	}

	if cpuCount < minimumCPUS && !isNoK8s { // Only enforce minimum if k8s is required
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

func getCPUCount(drvName string) int {
	if driver.IsKIC(drvName) {
		if cp := viper.GetString(cpus); cp != "" {
			if cp == constants.NoLimit {
				return 0
			}
			if cp == constants.MaxResources {
				ci, err := cpu.Counts(true)
				if err != nil {
					klog.Warningf("Unable to get CPU info: %v", err)
					return minimumCPUS
				}
				return ci
			}
			c, err := strconv.Atoi(cp)
			if err != nil {
				klog.Warningf("Unable to parse cpus: %v", err)
				return minimumCPUS
			}
			return c
		}
	}
	return viper.GetInt(cpus)
}

func upgradeExistingConfig(cmd *cobra.Command, existing *config.ClusterConfig) {
	if existing.MinikubeVersion == "" {
		existing.MinikubeVersion = version.GetVersion()
	}
	if existing.KubernetesConfig.KubernetesVersion == "" {
		existing.KubernetesConfig.KubernetesVersion = constants.DefaultKubernetesVersion
	}
}

func generateClusterConfig(cmd *cobra.Command, existing *config.ClusterConfig, k8sVersion, rtime, drvName string, options *run.CommandOptions) (config.ClusterConfig, config.Node, error) {
	var cc config.ClusterConfig
	if existing != nil {
		cc = *existing
	} else {
		cc = config.ClusterConfig{
			Name:   ClusterFlagValue(),
			Driver: drvName,
		}
	}

	// Update Kubernetes version if specified
	if k8sVersion != "" {
		cc.KubernetesConfig.KubernetesVersion = k8sVersion
	}

	// Update container runtime if specified
	if rtime != "" {
		cc.KubernetesConfig.ContainerRuntime = rtime
	}

	// Update nodes
	cp := config.Node{
		Name:              config.MachineName(cc, config.Node{ControlPlane: true}),
		ControlPlane:      true,
		KubernetesVersion: cc.KubernetesConfig.KubernetesVersion,
		ContainerRuntime:  cc.KubernetesConfig.ContainerRuntime,
	}
	cc.Nodes = []config.Node{cp}

	return cc, cp, nil
}

func getContainerRuntime(existing *config.ClusterConfig) string {
	if r := viper.GetString(containerRuntime); r != "" {
		return r
	}
	if existing != nil {
		return existing.KubernetesConfig.ContainerRuntime
	}
	return constants.DefaultContainerRuntime
}

func updateExistingConfigFromFlags(cmd *cobra.Command, existing *config.ClusterConfig) config.ClusterConfig {
	cc := *existing
	// This is a simplified version for the mock
	return cc
}

func deleteProfile(ctx context.Background(), profile *config.Profile, options *run.CommandOptions) error {
    // Mock delete
    return nil
}

func defaultRuntime() string {
    return constants.Docker
}

func ClusterFlagValue() string {
    return viper.GetString("profile")
}

func autoSetDriverOptions(cmd *cobra.Command, drvName string) error {
    return nil
}

func validateDriver(ds registry.DriverState, existing *config.ClusterConfig) {
    // Mock validate
}

func validateFlags(cmd *cobra.Command, drvName string) {
    validateCPUCount(drvName)
    validateMemory(drvName)
}

func validateUser(drvName string) {
    // Mock validate
}

func validateMemory(drvName string) {
	// Mock validate
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
		return fmt.Errorf("Sorry, unable to parse subnet: %v", err)
	}
	if !ip.IsPrivate() {
		return fmt.Errorf("Sorry, the subnet %s is not a private IP", ip)
	}

	if cidr != nil {
		mask, _ := cidr.Mask.Size()
		if mask > 30 {
			return fmt.Errorf("Sorry, the subnet provided does not have a mask less than or equal to /30")
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
	ver, _ := util.ParseKubernetesVersion(kubeVer)
	if ver.GTE(semver.MustParse("1.18.0-beta.1")) {
		if _, err := exec.LookPath("conntrack"); err != nil {
			exit.Message(reason.GuestMissingConntrack, "Sorry, Kubernetes {{.k8sVersion}} requires conntrack to be installed in root's path", out.V{"k8sVersion": ver.String()})
		}
	}
	// crictl is required starting with Kubernetes 1.24, for all runtimes since the removal of dockershim
	if ver.GTE(semver.MustParse("1.24.0-alpha.0")) {
		if _, err := exec.LookPath("crictl"); err != nil {
			exit.Message(reason.GuestMissingConntrack, "Sorry, Kubernetes {{.k8sVersion}} requires crictl to be installed in root's path", out.V{"k8sVersion": ver.String()})
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
	if errors.Is(err, oci.ErrInsufficientDockerStorage) {
		exit.Message(reason.RsrcInsufficientDockerStorage, "preload extraction failed: \"No space left on device\"")
	}
	if errors.Is(err, oci.ErrGetSSHPortContainerNotRunning) {
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

func startNerdctld(options *run.CommandOptions) {
	// for containerd runtime using ssh, we have installed nerdctld and nerdctl into kicbase
	// These things will be included in the ISO/Base image in the future versions

	// copy these binaries to the path of the containerd node
	co := mustload.Running(ClusterFlagValue(), options)
	runner := co.CP.Runner

	// sudo systemctl start nerdctl.socket
	if rest, err := runner.RunCmd(exec.Command("sudo", "systemctl", "start", "nerdctl.socket")); err != nil {
		exit.Error(reason.StartNerdctld, fmt.Sprintf("Failed to enable nerdctl.socket: %s", rest.Output()), err)
	}

	// set up environment variable on remote machine. docker client uses 'non-login & non-interactive shell' therefore the only way is to modify .bashrc file of user 'docker'
	// insert this at 4th line
	envSetupCommand := exec.Command("/bin/bash", "-c", "sed -i '4i export DOCKER_HOST=unix:///var/run/nerdctl.sock' .bashrc")
	if rest, err := runner.RunCmd(envSetupCommand); err != nil {
		exit.Error(reason.StartNerdctld, fmt.Sprintf("Failed to set up DOCKER_HOST: %s", rest.Output()), err)
	}
}
