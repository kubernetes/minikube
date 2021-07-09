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

package reason

import (
	"fmt"

	"k8s.io/minikube/pkg/minikube/style"
)

const issueBase = "https://github.com/kubernetes/minikube/issues"

// Kind describes reason metadata
type Kind struct {
	// ID is an unique and stable string describing a reason
	ID string
	// ExitCode to be used (defaults to 1)
	ExitCode int
	// Style is what emoji prefix to use for this reason
	Style style.Enum

	// Advice is actionable text that the user should follow
	Advice string
	// URL is a reference URL for more information
	URL string
	// Issues are a list of related issues to this issue
	Issues []int
	// Show the new issue link
	NewIssueLink bool
	// Do not attempt to match this reason to a specific known issue
	NoMatch bool
}

// IssueURLs returns URLs for issues
func (k *Kind) IssueURLs() []string {
	is := []string{}
	for _, i := range k.Issues {
		is = append(is, fmt.Sprintf("%s/%d", issueBase, i))
	}
	return is
}

// Sections are ordered roughly by stack dependencies
var (
	// minikube has been passed an incorrect parameter
	Usage = Kind{ID: "MK_USAGE", ExitCode: ExProgramUsage}
	// minikube has no current cluster running
	UsageNoProfileRunning = Kind{ID: "MK_USAGE_NO_PROFILE", ExitCode: ExProgramUsage,
		Advice: `You can create one using 'minikube start'.
		`,
		Style: style.Caching,
	}
	// minikube was interrupted by an OS signal
	Interrupted = Kind{ID: "MK_INTERRUPTED", ExitCode: ExProgramConflict}

	// user attempted to run a Windows executable (.exe) inside of WSL rather than using the Linux binary
	WrongBinaryWSL = Kind{ID: "MK_WRONG_BINARY_WSL", ExitCode: ExProgramUnsupported}

	// minikube failed to create a new Docker Machine api client
	NewAPIClient = Kind{ID: "MK_NEW_APICLIENT", ExitCode: ExProgramError}
	// minikube could not disable an addon, e.g. dashboard addon
	InternalAddonDisable = Kind{ID: "MK_ADDON_DISABLE", ExitCode: ExProgramError}
	// minikube could not enable an addon, e.g. dashboard addon
	InternalAddonEnable = Kind{ID: "MK_ADDON_ENABLE", ExitCode: ExProgramError}
	// minikube failed to update internal configuration, such as the cached images config map
	InternalAddConfig = Kind{ID: "MK_ADD_CONFIG", ExitCode: ExProgramError}
	// minikube failed to create a cluster bootstrapper
	InternalBootstrapper = Kind{ID: "MK_BOOTSTRAPPER", ExitCode: ExProgramError}
	// minikube failed to list cached images
	InternalCacheList = Kind{ID: "MK_CACHE_LIST", ExitCode: ExProgramError}
	// minkube failed to cache and load cached images
	InternalCacheLoad = Kind{ID: "MK_CACHE_LOAD", ExitCode: ExProgramError}
	// minikube failed to load a Docker Machine CommandRunner
	InternalCommandRunner = Kind{ID: "MK_COMMAND_RUNNER", ExitCode: ExProgramError}
	// minikube failed to generate shell command completion for a supported shell
	InternalCompletion = Kind{ID: "MK_COMPLETION", ExitCode: ExProgramError}
	// minikube failed to set an internal config value
	InternalConfigSet = Kind{ID: "MK_CONFIG_SET", ExitCode: ExProgramError}
	// minikube failed to unset an internal config value
	InternalConfigUnset = Kind{ID: "MK_CONFIG_UNSET", ExitCode: ExProgramError}
	// minikube failed to view current config values
	InternalConfigView = Kind{ID: "MK_CONFIG_VIEW", ExitCode: ExProgramError}
	// minikybe failed to delete an internal configuration, such as a cached image
	InternalDelConfig = Kind{ID: "MK_DEL_CONFIG", ExitCode: ExProgramError}
	// minikube failed to generate script to activate minikube docker-env
	InternalDockerScript = Kind{ID: "MK_DOCKER_SCRIPT", ExitCode: ExProgramError}
	// an error occurred when viper attempted to bind flags to configuration
	InternalBindFlags = Kind{ID: "MK_BIND_FLAGS", ExitCode: ExProgramError}
	// minkube was passed an invalid format string in the --format flag
	InternalFormatUsage = Kind{ID: "MK_FORMAT_USAGE", ExitCode: ExProgramError}
	// minikube failed to auto-generate markdown-based documentation in the specified folder
	InternalGenerateDocs = Kind{ID: "MK_GENERATE_DOCS", ExitCode: ExProgramError}
	// minikube failed to marshal a JSON object
	InternalJSONMarshal = Kind{ID: "MK_JSON_MARSHAL", ExitCode: ExProgramError}
	// minikube failed to create a Kubernetes client set which is necessary for querying the Kubernetes API
	InternalKubernetesClient = Kind{ID: "MK_K8S_CLIENT", ExitCode: ExControlPlaneUnavailable}
	// minikube failed to list some configuration data
	InternalListConfig = Kind{ID: "MK_LIST_CONFIG", ExitCode: ExProgramError}
	// minikube failed to follow or watch minikube logs
	InternalLogFollow = Kind{ID: "MK_LOG_FOLLOW", ExitCode: ExProgramError}
	// minikube failed to create an appropriate new runtime based on the driver in use
	InternalNewRuntime = Kind{ID: "MK_NEW_RUNTIME", ExitCode: ExProgramError}
	// minikube was passed an invalid value for the --output command line flag
	InternalOutputUsage = Kind{ID: "MK_OUTPUT_USAGE", ExitCode: ExProgramError}
	// minikube could not configure the runtime in use, or the runtime failed
	InternalRuntime = Kind{ID: "MK_RUNTIME", ExitCode: ExProgramError}
	// minikube was passed a reserved keyword as a profile name, which is not allowed
	InternalReservedProfile = Kind{ID: "MK_RESERVED_PROFILE", ExitCode: ExProgramConflict}
	// minkube failed to generate script to set or unset minikube-env
	InternalEnvScript = Kind{ID: "MK_ENV_SCRIPT", ExitCode: ExProgramError}
	// minikube failed to detect the shell in use
	InternalShellDetect = Kind{ID: "MK_SHELL_DETECT", ExitCode: ExProgramError}
	// minikube failed to output JSON-formatted minikube status
	InternalStatusJSON = Kind{ID: "MK_STATUS_JSON", ExitCode: ExProgramError}
	// minikube failed to output minikube status text
	InternalStatusText = Kind{ID: "MK_STATUS_TEXT", ExitCode: ExProgramError}
	// minikube failed to execute (i.e. fill in values for) a view template for displaying current config
	InternalViewExec = Kind{ID: "MK_VIEW_EXEC", ExitCode: ExProgramError}
	// minikube failed to create view template for displaying current config
	InternalViewTmpl = Kind{ID: "MK_VIEW_TMPL", ExitCode: ExProgramError}
	// minikube failed to marshal a YAML object
	InternalYamlMarshal = Kind{ID: "MK_YAML_MARSHAL", ExitCode: ExProgramError}
	// minikube could not locate credentials needed to utilize an appropriate service, e.g. GCP
	InternalCredsNotFound = Kind{ID: "MK_CREDENTIALS_NOT_FOUND", ExitCode: ExProgramNotFound, Style: style.Shrug}
	// minikube was passed service credentials when they were not needed, such as when using the GCP Auth addon when running in GCE
	InternalCredsNotNeeded = Kind{ID: "MK_CREDENTIALS_NOT_NEEDED", ExitCode: ExProgramNotFound, Style: style.Shrug}
	// minikube found an invalid semver string for kubernetes in the minikube constants
	InternalSemverParse = Kind{ID: "MK_SEMVER_PARSE", ExitCode: ExProgramError}
	// minikube was unable to daemonize the minikube process
	DaemonizeError = Kind{ID: "MK_DAEMONIZE", ExitCode: ExProgramError}

	// insufficient cores available for use by minikube and kubernetes
	RsrcInsufficientCores = Kind{ID: "RSRC_INSUFFICIENT_CORES", ExitCode: ExInsufficientCores, Style: style.UnmetRequirement}
	// insufficient cores available for use by Docker Desktop on Mac
	RsrcInsufficientDarwinDockerCores = Kind{
		ID:       "RSRC_DOCKER_CORES",
		ExitCode: ExInsufficientCores,
		Advice: `1. Click on "Docker for Desktop" menu icon
			2. Click "Preferences"
			3. Click "Resources"
			4. Increase "CPUs" slider bar to 2 or higher
			5. Click "Apply & Restart"`,
		Style: style.UnmetRequirement,
		URL:   "https://docs.docker.com/docker-for-mac/#resources",
	}

	// insufficient cores available for use by Docker Desktop on Windows
	RsrcInsufficientWindowsDockerCores = Kind{
		ID:       "RSRC_DOCKER_CORES",
		ExitCode: ExInsufficientCores,
		Advice: `1. Open the "Docker Desktop" menu by clicking the Docker icon in the system tray
		2. Click "Settings"
		3. Click "Resources"
		4. Increase "CPUs" slider bar to 2 or higher
		5. Click "Apply & Restart"`,
		URL:   "https://docs.docker.com/docker-for-windows/#resources",
		Style: style.UnmetRequirement,
	}

	// insufficient memory (less than the recommended minimum) allocated to minikube
	RsrcInsufficientReqMemory = Kind{ID: "RSRC_INSUFFICIENT_REQ_MEMORY", ExitCode: ExInsufficientMemory, Style: style.UnmetRequirement}
	// insufficient memory (less than the recommended minimum) available on the system running minikube
	RsrcInsufficientSysMemory = Kind{ID: "RSRC_INSUFFICIENT_SYS_MEMORY", ExitCode: ExInsufficientMemory, Style: style.UnmetRequirement}
	// insufficient memory available for the driver in use by minikube
	RsrcInsufficientContainerMemory = Kind{ID: "RSRC_INSUFFICIENT_CONTAINER_MEMORY", ExitCode: ExInsufficientMemory, Style: style.UnmetRequirement}
	// insufficient memory available to Docker Desktop on Windows
	RsrcInsufficientWindowsDockerMemory = Kind{
		ID:       "RSRC_DOCKER_MEMORY",
		ExitCode: ExInsufficientMemory,
		Advice: `1. Open the "Docker Desktop" menu by clicking the Docker icon in the system tray
		2. Click "Settings"
		3. Click "Resources"
		4. Increase "Memory" slider bar to {{.recommend}} or higher
		5. Click "Apply & Restart"`,
		URL:   "https://docs.docker.com/docker-for-windows/#resources",
		Style: style.UnmetRequirement,
	}
	// insufficient memory available to Docker Desktop on Mac
	RsrcInsufficientDarwinDockerMemory = Kind{
		ID:       "RSRC_DOCKER_MEMORY",
		ExitCode: ExInsufficientMemory,
		Advice: `1. Click on "Docker for Desktop" menu icon
			2. Click "Preferences"
			3. Click "Resources"
			4. Increase "Memory" slider bar to {{.recommend}} or higher
			5. Click "Apply & Restart"`,
		Style: style.UnmetRequirement,
		URL:   "https://docs.docker.com/docker-for-mac/#resources",
	}

	// insufficient disk storage available to the docker driver
	RsrcInsufficientDockerStorage = Kind{
		ID:       "RSRC_DOCKER_STORAGE",
		ExitCode: ExInsufficientStorage,
		Advice: `Try one or more of the following to free up space on the device:
	
			1. Run "docker system prune" to remove unused Docker data (optionally with "-a")
			2. Increase the storage allocated to Docker for Desktop by clicking on:
				Docker icon > Preferences > Resources > Disk Image Size
			3. Run "minikube ssh -- docker system prune" if using the Docker container runtime`,
		Issues: []int{9024},
	}
	// insufficient disk storage available to the podman driver
	RsrcInsufficientPodmanStorage = Kind{
		ID:       "RSRC_PODMAN_STORAGE",
		ExitCode: ExInsufficientStorage,
		Advice: `Try one or more of the following to free up space on the device:
	
			1. Run "sudo podman system prune" to remove unused podman data
			2. Run "minikube ssh -- docker system prune" if using the Docker container runtime`,
		Issues: []int{9024},
	}

	// insufficient disk storage available for running minikube and kubernetes
	RsrcInsufficientStorage = Kind{ID: "RSRC_INSUFFICIENT_STORAGE", ExitCode: ExInsufficientStorage, Style: style.UnmetRequirement}

	// minikube could not create the minikube directory
	HostHomeMkdir = Kind{ID: "HOST_HOME_MKDIR", ExitCode: ExHostPermission}
	// minikube could not change permissions for the minikube directory
	HostHomeChown = Kind{ID: "HOST_HOME_CHOWN", ExitCode: ExHostPermission}
	// minikube failed to open the host browser, such as when running minikube dashboard
	HostBrowser = Kind{ID: "HOST_BROWSER", ExitCode: ExHostError}
	// minikube failed to load cluster config from the host for the profile in use
	HostConfigLoad = Kind{ID: "HOST_CONFIG_LOAD", ExitCode: ExHostConfig}
	// the current user has insufficient permissions to create the minikube profile directory
	HostHomePermission = Kind{
		ID:       "HOST_HOME_PERMISSION",
		ExitCode: ExHostPermission,
		Advice:   "Your user lacks permissions to the minikube profile directory. Run: 'sudo chown -R $USER $HOME/.minikube; chmod -R u+wrx $HOME/.minikube' to fix",
		Issues:   []int{9165},
	}

	// minikube failed to determine current user
	HostCurrentUser = Kind{ID: "HOST_CURRENT_USER", ExitCode: ExHostConfig}
	// minikube failed to delete cached images from host
	HostDelCache = Kind{ID: "HOST_DEL_CACHE", ExitCode: ExHostError}
	// minikube failed to kill a mount process
	HostKillMountProc = Kind{ID: "HOST_KILL_MOUNT_PROC", ExitCode: ExHostError}
	// minikube failed to update host Kubernetes resources config
	HostKubeconfigUpdate = Kind{ID: "HOST_KUBECONFIG_UPDATE", ExitCode: ExHostConfig}
	// minikube failed to delete Kubernetes config from context for a given profile
	HostKubeconfigDeleteCtx = Kind{ID: "HOST_KUBECONFIG_DELETE_CTX", ExitCode: ExHostConfig}
	// minikube failed to launch a kubectl proxy
	HostKubectlProxy = Kind{ID: "HOST_KUBECTL_PROXY", ExitCode: ExHostError}
	// minikube failed to write mount pid
	HostMountPid = Kind{ID: "HOST_MOUNT_PID", ExitCode: ExHostError}
	// minikube was passed a path to a host directory that does not exist
	HostPathMissing = Kind{ID: "HOST_PATH_MISSING", ExitCode: ExHostNotFound}
	// minikube failed to access info for a directory path
	HostPathStat = Kind{ID: "HOST_PATH_STAT", ExitCode: ExHostError}
	// minikube failed to purge minikube config directories
	HostPurge = Kind{ID: "HOST_PURGE", ExitCode: ExHostError}
	// minikube failed to persist profile config
	HostSaveProfile = Kind{ID: "HOST_SAVE_PROFILE", ExitCode: ExHostConfig}

	// minikube could not find a provider for the selected driver
	ProviderNotFound = Kind{ID: "PROVIDER_NOT_FOUND", ExitCode: ExProviderNotFound}
	// the host does not support or is improperly configured to support a provider for the selected driver
	ProviderUnavailable = Kind{ID: "PROVIDER_UNAVAILABLE", ExitCode: ExProviderNotFound, Style: style.Shrug}

	// minikube failed to access the driver control plane or API endpoint
	DrvCPEndpoint = Kind{ID: "DRV_CP_ENDPOINT",
		Advice: `Recreate the cluster by running:
		minikube delete {{.profileArg}}
		minikube start {{.profileArg}}`,
		ExitCode: ExDriverError,
		Style:    style.Failure,
	}
	// minikube failed to bind container ports to host ports
	DrvPortForward = Kind{ID: "DRV_PORT_FORWARD", ExitCode: ExDriverError}
	// the driver in use does not support multi-node clusters
	DrvUnsupportedMulti = Kind{ID: "DRV_UNSUPPORTED_MULTINODE", ExitCode: ExDriverConflict}
	// the specified driver is not supported on the host OS
	DrvUnsupportedOS = Kind{ID: "DRV_UNSUPPORTED_OS", ExitCode: ExDriverUnsupported}
	// the driver in use does not support the selected profile or multiple profiles
	DrvUnsupportedProfile = Kind{ID: "DRV_UNSUPPORTED_PROFILE", ExitCode: ExDriverUnsupported}
	// minikube failed to locate specified driver
	DrvNotFound = Kind{ID: "DRV_NOT_FOUND", ExitCode: ExDriverNotFound}
	// minikube could not find a valid driver
	DrvNotDetected = Kind{ID: "DRV_NOT_DETECTED", ExitCode: ExDriverNotFound}
	// minikube found drivers but none were ready to use
	DrvNotHealthy = Kind{ID: "DRV_NOT_HEALTHY", ExitCode: ExDriverNotFound}
	// minikube found the docker driver but the docker service was not running
	DrvDockerNotRunning = Kind{ID: "DRV_DOCKER_NOT_RUNNING", ExitCode: ExDriverNotFound}
	// the driver in use is being run as root
	DrvAsRoot = Kind{ID: "DRV_AS_ROOT", ExitCode: ExDriverPermission}
	// the specified driver needs to be run as root
	DrvNeedsRoot = Kind{ID: "DRV_NEEDS_ROOT", ExitCode: ExDriverPermission}

	// minikube failed to load cached images
	GuestCacheLoad = Kind{ID: "GUEST_CACHE_LOAD", ExitCode: ExGuestError}
	// minikube failed to setup certificates
	GuestCert = Kind{ID: "GUEST_CERT", ExitCode: ExGuestError}
	// minikube failed to access the control plane
	GuestCpConfig = Kind{ID: "GUEST_CP_CONFIG", ExitCode: ExGuestConfig}
	// minikube failed to properly delete a resource, such as a profile
	GuestDeletion = Kind{ID: "GUEST_DELETION", ExitCode: ExGuestError}
	// minikube failed to list images on the machine
	GuestImageList = Kind{ID: "GUEST_IMAGE_LIST", ExitCode: ExGuestError}
	// minikube failed to pull or load an image
	GuestImageLoad = Kind{ID: "GUEST_IMAGE_LOAD", ExitCode: ExGuestError}
	// minikube failed to remove an image
	GuestImageRemove = Kind{ID: "GUEST_IMAGE_REMOVE", ExitCode: ExGuestError}
	// minikube failed to build an image
	GuestImageBuild = Kind{ID: "GUEST_IMAGE_BUILD", ExitCode: ExGuestError}
	// minikube failed to load host
	GuestLoadHost = Kind{ID: "GUEST_LOAD_HOST", ExitCode: ExGuestError}
	// minkube failed to create a mount
	GuestMount = Kind{ID: "GUEST_MOUNT", ExitCode: ExGuestError}
	// minkube failed to update a mount
	GuestMountConflict = Kind{ID: "GUEST_MOUNT_CONFLICT", ExitCode: ExGuestConflict}
	// minikube failed to add a node to the cluster
	GuestNodeAdd = Kind{ID: "GUEST_NODE_ADD", ExitCode: ExGuestError}
	// minikube failed to remove a node from the cluster
	GuestNodeDelete = Kind{ID: "GUEST_NODE_DELETE", ExitCode: ExGuestError}
	// minikube failed to provision a node
	GuestNodeProvision = Kind{ID: "GUEST_NODE_PROVISION", ExitCode: ExGuestError}
	// minikube failed to retrieve information for a cluster node
	GuestNodeRetrieve = Kind{ID: "GUEST_NODE_RETRIEVE", ExitCode: ExGuestNotFound}
	// minikube failed to startup a cluster node
	GuestNodeStart = Kind{ID: "GUEST_NODE_START", ExitCode: ExGuestError}
	// minikube failed to pause the cluster process
	GuestPause = Kind{ID: "GUEST_PAUSE", ExitCode: ExGuestError}
	// minikube failed to delete a machine profile directory
	GuestProfileDeletion = Kind{ID: "GUEST_PROFILE_DELETION", ExitCode: ExGuestError}
	// minikube failed while attempting to provision the guest
	GuestProvision = Kind{ID: "GUEST_PROVISION", ExitCode: ExGuestError}
	// docker container exited prematurely during provisioning
	GuestProvisionContainerExited = Kind{ID: "GUEST_PROVISION_CONTAINER_EXITED", ExitCode: ExGuestError}
	// minikube failed to start a node with current driver
	GuestStart = Kind{ID: "GUEST_START", ExitCode: ExGuestError}
	// minikube failed to get docker machine status
	GuestStatus = Kind{ID: "GUEST_STATUS", ExitCode: ExGuestError}
	// stopping the cluster process timed out
	GuestStopTimeout = Kind{ID: "GUEST_STOP_TIMEOUT", ExitCode: ExGuestTimeout}
	// minikube failed to unpause the cluster process
	GuestUnpause = Kind{ID: "GUEST_UNPAUSE", ExitCode: ExGuestError}
	// minikube failed to check if Kubernetes containers are paused
	GuestCheckPaused = Kind{ID: "GUEST_CHECK_PAUSED", ExitCode: ExGuestError}
	// minikube cluster was created used a driver that is incompatible with the driver being requested
	GuestDrvMismatch = Kind{ID: "GUEST_DRIVER_MISMATCH", ExitCode: ExGuestConflict, Style: style.Conflict}
	// minikube could not find conntrack on the host, which is required from Kubernetes 1.18 onwards
	GuestMissingConntrack = Kind{ID: "GUEST_MISSING_CONNTRACK", ExitCode: ExGuestUnsupported}

	// minikube failed to get the host IP to use from within the VM
	IfHostIP = Kind{ID: "IF_HOST_IP", ExitCode: ExLocalNetworkError}
	// minikube failed to parse the input IP address for mount
	IfMountIP = Kind{ID: "IF_MOUNT_IP", ExitCode: ExLocalNetworkError}
	// minikube failed to parse or find port for mount
	IfMountPort = Kind{ID: "IF_MOUNT_PORT", ExitCode: ExLocalNetworkError}
	// minikube failed to access an ssh client on the host machine
	IfSSHClient = Kind{ID: "IF_SSH_CLIENT", ExitCode: ExLocalNetworkError}

	// minikube failed to cache kubernetes binaries for the current runtime
	InetCacheBinaries = Kind{ID: "INET_CACHE_BINARIES", ExitCode: ExInternetError}
	// minikube failed to cache the kubectl binary
	InetCacheKubectl = Kind{ID: "INET_CACHE_KUBECTL", ExitCode: ExInternetError}
	// minikube failed to cache required images to tar files
	InetCacheTar = Kind{ID: "INET_CACHE_TAR", ExitCode: ExInternetError}
	// minikube was unable to access main repository and mirrors for images
	InetRepo = Kind{ID: "INET_REPO", ExitCode: ExInternetError}
	// minikube was unable to access any known image repositories
	InetReposUnavailable = Kind{ID: "INET_REPOS_UNAVAILABLE", ExitCode: ExInternetError}
	// minikube was unable to fetch latest release/version info for minkikube
	InetVersionUnavailable = Kind{ID: "INET_VERSION_UNAVAILABLE", ExitCode: ExInternetUnavailable}
	// minikube received invalid empty data for latest release/version info from the server
	InetVersionEmpty = Kind{ID: "INET_VERSION_EMPTY", ExitCode: ExInternetConfig}

	// minikube failed to enable the current container runtime
	RuntimeEnable = Kind{ID: "RUNTIME_ENABLE", ExitCode: ExRuntimeError}
	// minikube failed to cache images for the current container runtime
	RuntimeCache = Kind{ID: "RUNTIME_CACHE", ExitCode: ExRuntimeError}

	// service check timed out while starting minikube dashboard
	SvcCheckTimeout = Kind{ID: "SVC_CHECK_TIMEOUT", ExitCode: ExSvcTimeout}
	// minikube was unable to access a service
	SvcTimeout = Kind{ID: "SVC_TIMEOUT", ExitCode: ExSvcTimeout}
	// minikube failed to list services for the specified namespace
	SvcList = Kind{ID: "SVC_LIST", ExitCode: ExSvcError}
	// minikube failed to start a tunnel
	SvcTunnelStart = Kind{ID: "SVC_TUNNEL_START", ExitCode: ExSvcError}
	// minikube could not stop an active tunnel
	SvcTunnelStop = Kind{ID: "SVC_TUNNEL_STOP", ExitCode: ExSvcError}
	// minikube was unable to access the service url
	SvcURLTimeout = Kind{ID: "SVC_URL_TIMEOUT", ExitCode: ExSvcTimeout}
	// minikube couldn't find the specified service in the specified namespace
	SvcNotFound = Kind{ID: "SVC_NOT_FOUND", ExitCode: ExSvcNotFound}

	// user attempted to use a command that is not supported by the driver currently in use
	EnvDriverConflict = Kind{ID: "ENV_DRIVER_CONFLICT", ExitCode: ExDriverConflict}
	// user attempted to run a command that is not supported on multi-node setup without some additional configuration
	EnvMultiConflict = Kind{ID: "ENV_MULTINODE_CONFLICT", ExitCode: ExGuestConflict}
	// the podman service was unavailable to the cluster
	EnvPodmanUnavailable = Kind{ID: "ENV_PODMAN_UNAVAILABLE", ExitCode: ExRuntimeUnavailable}

	// user attempted to use an addon that is not supported
	AddonUnsupported = Kind{ID: "SVC_ADDON_UNSUPPORTED", ExitCode: ExSvcUnsupported}
	// user attempted to use an addon that is currently not enabled
	AddonNotEnabled = Kind{ID: "SVC_ADDON_NOT_ENABLED", ExitCode: ExProgramConflict}

	// minikube failed to update the Kubernetes cluster
	KubernetesInstallFailed = Kind{ID: "K8S_INSTALL_FAILED", ExitCode: ExControlPlaneError}
	// minikube failed to update the Kubernetes cluster because the container runtime was unavailable
	KubernetesInstallFailedRuntimeNotRunning = Kind{ID: "K8S_INSTALL_FAILED_CONTAINER_RUNTIME_NOT_RUNNING", ExitCode: ExRuntimeNotRunning}
	// an outdated Kubernetes version was specified for minikube to use
	KubernetesTooOld = Kind{ID: "K8S_OLD_UNSUPPORTED", ExitCode: ExControlPlaneUnsupported}
	// minikube was unable to safely downgrade installed Kubernetes version
	KubernetesDowngrade = Kind{
		ID:       "K8S_DOWNGRADE_UNSUPPORTED",
		ExitCode: ExControlPlaneUnsupported,
		Advice: `1) Recreate the cluster with Kubernetes {{.new}}, by running:
	  
		  minikube delete{{.profile}}
		  minikube start{{.profile}} --kubernetes-version={{.prefix}}{{.new}}
	  
		2) Create a second cluster with Kubernetes {{.new}}, by running:
	  
		  minikube start -p {{.suggestedName}} --kubernetes-version={{.prefix}}{{.new}}
	  
		3) Use the existing cluster at version Kubernetes {{.old}}, by running:
	  
		  minikube start{{.profile}} --kubernetes-version={{.prefix}}{{.old}}
		`,
		Style: style.SeeNoEvil,
	}
)
