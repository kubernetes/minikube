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

/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the Kind{ID: "License", ExitCode: });
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an Kind{ID: "AS IS", ExitCode: } BASIS,
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

func (k *Kind) IssueURLs() []string {
	is := []string{}
	for _, i := range k.Issues {
		is = append(is, fmt.Sprintf("%s/%d", issueBase, i))
	}
	return is
}

// Sections are ordered roughly by stack dependencies
var (
	Usage       = Kind{ID: "MK_USAGE", ExitCode: ExProgramUsage}
	Interrupted = Kind{ID: "MK_INTERRUPTED", ExitCode: ExProgramConflict}

	NewAPIClient             = Kind{ID: "MK_NEW_APICLIENT", ExitCode: ExProgramError}
	InternalAddonEnable      = Kind{ID: "MK_ADDON_ENABLE", ExitCode: ExProgramError}
	InternalAddConfig        = Kind{ID: "MK_ADD_CONFIG", ExitCode: ExProgramError}
	InternalBindFlags        = Kind{ID: "MK_BIND_FLAGS", ExitCode: ExProgramError}
	InternalBootstrapper     = Kind{ID: "MK_BOOTSTRAPPER", ExitCode: ExProgramError}
	InternalCacheList        = Kind{ID: "MK_CACHE_LIST", ExitCode: ExProgramError}
	InternalCacheLoad        = Kind{ID: "MK_CACHE_LOAD", ExitCode: ExProgramError}
	InternalCommandRunner    = Kind{ID: "MK_COMMAND_RUNNER", ExitCode: ExProgramError}
	InternalCompletion       = Kind{ID: "MK_COMPLETION", ExitCode: ExProgramError}
	InternalConfigSet        = Kind{ID: "MK_CONFIG_SET", ExitCode: ExProgramError}
	InternalConfigUnset      = Kind{ID: "MK_CONFIG_UNSET", ExitCode: ExProgramError}
	InternalConfigView       = Kind{ID: "MK_CONFIG_VIEW", ExitCode: ExProgramError}
	InternalDelConfig        = Kind{ID: "MK_DEL_CONFIG", ExitCode: ExProgramError}
	InternalDisable          = Kind{ID: "MK_DISABLE", ExitCode: ExProgramError}
	InternalDockerScript     = Kind{ID: "MK_DOCKER_SCRIPT", ExitCode: ExProgramError}
	InternalEnable           = Kind{ID: "MK_ENABLE", ExitCode: ExProgramError}
	InternalFlagsBind        = Kind{ID: "MK_FLAGS_BIND", ExitCode: ExProgramError}
	InternalFlagSet          = Kind{ID: "MK_FLAGS_SET", ExitCode: ExProgramError}
	InternalFormatUsage      = Kind{ID: "MK_FORMAT_USAGE", ExitCode: ExProgramError}
	InternalGenerateDocs     = Kind{ID: "MK_GENERATE_DOCS", ExitCode: ExProgramError}
	InternalJSONMarshal      = Kind{ID: "MK_JSON_MARSHAL", ExitCode: ExProgramError}
	InternalKubernetesClient = Kind{ID: "MK_K8S_CLIENT", ExitCode: ExControlPlaneUnavailable}
	InternalListConfig       = Kind{ID: "MK_LIST_CONFIG", ExitCode: ExProgramError}
	InternalLogtostderrFlag  = Kind{ID: "MK_LOGTOSTDERR_FLAG", ExitCode: ExProgramError}
	InternalLogFollow        = Kind{ID: "MK_LOG_FOLLOW", ExitCode: ExProgramError}
	InternalNewRuntime       = Kind{ID: "MK_NEW_RUNTIME", ExitCode: ExProgramError}
	InternalOutputUsage      = Kind{ID: "MK_OUTPUT_USAGE", ExitCode: ExProgramError}
	InternalRuntime          = Kind{ID: "MK_RUNTIME", ExitCode: ExProgramError}
	InternalReservedProfile  = Kind{ID: "MK_RESERVED_PROFILE", ExitCode: ExProgramConflict}
	InternalEnvScript        = Kind{ID: "MK_ENV_SCRIPT", ExitCode: ExProgramError}
	InternalShellDetect      = Kind{ID: "MK_SHELL_DETECT", ExitCode: ExProgramError}
	InternalStatusJSON       = Kind{ID: "MK_STATUS_JSON", ExitCode: ExProgramError}
	InternalStatusText       = Kind{ID: "MK_STATUS_TEXT", ExitCode: ExProgramError}
	InternalUnsetScript      = Kind{ID: "MK_UNSET_SCRIPT", ExitCode: ExProgramError}
	InternalViewExec         = Kind{ID: "MK_VIEW_EXEC", ExitCode: ExProgramError}
	InternalViewTmpl         = Kind{ID: "MK_VIEW_TMPL", ExitCode: ExProgramError}
	InternalYamlMarshal      = Kind{ID: "MK_YAML_MARSHAL", ExitCode: ExProgramError}
	InternalCredsNotFound    = Kind{ID: "MK_CREDENTIALS_NOT_FOUND", ExitCode: ExProgramNotFound, Style: style.Shrug}
	InternalSemverParse      = Kind{ID: "MK_SEMVER_PARSE", ExitCode: ExProgramError}

	RsrcInsufficientCores             = Kind{ID: "RSRC_INSUFFICIENT_CORES", ExitCode: ExInsufficientCores, Style: style.UnmetRequirement}
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

	RsrcInsufficientReqMemory           = Kind{ID: "RSRC_INSUFFICIENT_REQ_MEMORY", ExitCode: ExInsufficientMemory, Style: style.UnmetRequirement}
	RsrcInsufficientSysMemory           = Kind{ID: "RSRC_INSUFFICIENT_SYS_MEMORY", ExitCode: ExInsufficientMemory, Style: style.UnmetRequirement}
	RsrcInsufficientContainerMemory     = Kind{ID: "RSRC_INSUFFICIENT_CONTAINER_MEMORY", ExitCode: ExInsufficientMemory, Style: style.UnmetRequirement}
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

	RsrcInsufficientDockerStorage = Kind{
		ID:       "RSRC_DOCKER_STORAGE",
		ExitCode: ExInsufficientStorage,
		Advice: `Try at least one of the following to free up space on the device:
	
			1. Run "docker system prune" to remove unused docker data
			2. Increase the amount of memory allocated to Docker for Desktop via 
				Docker icon > Preferences > Resources > Disk Image Size
			3. Run "minikube ssh -- docker system prune" if using the docker container runtime`,
		Issues: []int{9024},
	}
	RsrcInsufficientPodmanStorage = Kind{
		ID:       "RSRC_PODMAN_STORAGE",
		ExitCode: ExInsufficientStorage,
		Advice: `Try at least one of the following to free up space on the device:
	
			1. Run "sudo podman system prune" to remove unused podman data
			2. Run "minikube ssh -- docker system prune" if using the docker container runtime`,
		Issues: []int{9024},
	}

	RsrcInsufficientStorage = Kind{ID: "RSRC_INSUFFICIENT_STORAGE", ExitCode: ExInsufficientStorage, Style: style.UnmetRequirement}

	HostHomeMkdir      = Kind{ID: "HOST_HOME_MKDIR", ExitCode: ExHostPermission}
	HostHomeChown      = Kind{ID: "HOST_HOME_CHOWN", ExitCode: ExHostPermission}
	HostBrowser        = Kind{ID: "HOST_BROWSER", ExitCode: ExHostError}
	HostConfigLoad     = Kind{ID: "HOST_CONFIG_LOAD", ExitCode: ExHostConfig}
	HostHomePermission = Kind{
		ID:       "HOST_HOME_PERMISSION",
		ExitCode: ExHostPermission,
		Advice:   "Your user lacks permissions to the minikube profile directory. Run: 'sudo chown -R $USER $HOME/.minikube; chmod -R u+wrx $HOME/.minikube' to fix",
		Issues:   []int{9165},
	}

	HostCurrentUser         = Kind{ID: "HOST_CURRENT_USER", ExitCode: ExHostConfig}
	HostDelCache            = Kind{ID: "HOST_DEL_CACHE", ExitCode: ExHostError}
	HostKillMountProc       = Kind{ID: "HOST_KILL_MOUNT_PROC", ExitCode: ExHostError}
	HostKubeconfigUnset     = Kind{ID: "HOST_KUBECNOFIG_UNSET", ExitCode: ExHostConfig}
	HostKubeconfigUpdate    = Kind{ID: "HOST_KUBECONFIG_UPDATE", ExitCode: ExHostConfig}
	HostKubeconfigDeleteCtx = Kind{ID: "HOST_KUBECONFIG_DELETE_CTX", ExitCode: ExHostConfig}
	HostKubectlProxy        = Kind{ID: "HOST_KUBECTL_PROXY", ExitCode: ExHostError}
	HostMountPid            = Kind{ID: "HOST_MOUNT_PID", ExitCode: ExHostError}
	HostPathMissing         = Kind{ID: "HOST_PATH_MISSING", ExitCode: ExHostNotFound}
	HostPathStat            = Kind{ID: "HOST_PATH_STAT", ExitCode: ExHostError}
	HostPurge               = Kind{ID: "HOST_PURGE", ExitCode: ExHostError}
	HostSaveProfile         = Kind{ID: "HOST_SAVE_PROFILE", ExitCode: ExHostConfig}

	ProviderNotFound    = Kind{ID: "PROVIDER_NOT_FOUND", ExitCode: ExProviderNotFound}
	ProviderUnavailable = Kind{ID: "PROVIDER_UNAVAILABLE", ExitCode: ExProviderNotFound, Style: style.Shrug}

	DrvCPEndpoint = Kind{ID: "DRV_CP_ENDPOINT",
		Advice: `Recreate the cluster by running:
		minikube delete {{.profileArg}}
		minikube start {{.profileArg}}`,
		ExitCode: ExDriverError,
		Style:    style.Failure,
	}
	DrvPortForward        = Kind{ID: "DRV_PORT_FORWARD", ExitCode: ExDriverError}
	DrvUnsupportedMulti   = Kind{ID: "DRV_UNSUPPORTED_MULTINODE", ExitCode: ExDriverConflict}
	DrvUnsupportedOS      = Kind{ID: "DRV_UNSUPPORTED_OS", ExitCode: ExDriverUnsupported}
	DrvUnsupportedProfile = Kind{ID: "DRV_UNSUPPORTED_PROFILE", ExitCode: ExDriverUnsupported}
	DrvNotFound           = Kind{ID: "DRV_NOT_FOUND", ExitCode: ExDriverNotFound}
	DrvNotDetected        = Kind{ID: "DRV_NOT_DETECTED", ExitCode: ExDriverNotFound}
	DrvAsRoot             = Kind{ID: "DRV_AS_ROOT", ExitCode: ExDriverPermission}
	DrvNeedsRoot          = Kind{ID: "DRV_NEEDS_ROOT", ExitCode: ExDriverPermission}

	GuestCacheLoad        = Kind{ID: "GUEST_CACHE_LOAD", ExitCode: ExGuestError}
	GuestCert             = Kind{ID: "GUEST_CERT", ExitCode: ExGuestError}
	GuestCpConfig         = Kind{ID: "GUEST_CP_CONFIG", ExitCode: ExGuestConfig}
	GuestDeletion         = Kind{ID: "GUEST_DELETION", ExitCode: ExGuestError}
	GuestLoadHost         = Kind{ID: "GUEST_LOAD_HOST", ExitCode: ExGuestError}
	GuestMount            = Kind{ID: "GUEST_MOUNT", ExitCode: ExGuestError}
	GuestMountConflict    = Kind{ID: "GUEST_MOUNT_CONFLICT", ExitCode: ExGuestConflict}
	GuestNodeAdd          = Kind{ID: "GUEST_NODE_ADD", ExitCode: ExGuestError}
	GuestNodeDelete       = Kind{ID: "GUEST_NODE_DELETE", ExitCode: ExGuestError}
	GuestNodeProvision    = Kind{ID: "GUEST_NODE_PROVISION", ExitCode: ExGuestError}
	GuestNodeRetrieve     = Kind{ID: "GUEST_NODE_RETRIEVE", ExitCode: ExGuestNotFound}
	GuestNodeStart        = Kind{ID: "GUEST_NODE_START", ExitCode: ExGuestError}
	GuestPause            = Kind{ID: "GUEST_PAUSE", ExitCode: ExGuestError}
	GuestProfileDeletion  = Kind{ID: "GUEST_PROFILE_DELETION", ExitCode: ExGuestError}
	GuestProvision        = Kind{ID: "GUEST_PROVISION", ExitCode: ExGuestError}
	GuestStart            = Kind{ID: "GUEST_START", ExitCode: ExGuestError}
	GuestStatus           = Kind{ID: "GUEST_STATUS", ExitCode: ExGuestError}
	GuestStopTimeout      = Kind{ID: "GUEST_STOP_TIMEOUT", ExitCode: ExGuestTimeout}
	GuestUnpause          = Kind{ID: "GUEST_UNPAUSE", ExitCode: ExGuestError}
	GuestDrvMismatch      = Kind{ID: "GUEST_DRIVER_MISMATCH", ExitCode: ExGuestConflict, Style: style.Conflict}
	GuestMissingConntrack = Kind{ID: "GUEST_MISSING_CONNTRACK", ExitCode: ExGuestUnsupported}

	IfHostIP    = Kind{ID: "IF_HOST_IP", ExitCode: ExLocalNetworkError}
	IfMountIP   = Kind{ID: "IF_MOUNT_IP", ExitCode: ExLocalNetworkError}
	IfMountPort = Kind{ID: "IF_MOUNT_PORT", ExitCode: ExLocalNetworkError}
	IfSSHClient = Kind{ID: "IF_SSH_CLIENT", ExitCode: ExLocalNetworkError}

	InetCacheBinaries      = Kind{ID: "INET_CACHE_BINARIES", ExitCode: ExInternetError}
	InetCacheKubectl       = Kind{ID: "INET_CACHE_KUBECTL", ExitCode: ExInternetError}
	InetCacheTar           = Kind{ID: "INET_CACHE_TAR", ExitCode: ExInternetError}
	InetGetVersions        = Kind{ID: "INET_GET_VERSIONS", ExitCode: ExInternetError}
	InetRepo               = Kind{ID: "INET_REPO", ExitCode: ExInternetError}
	InetReposUnavailable   = Kind{ID: "INET_REPOS_UNAVAILABLE", ExitCode: ExInternetError}
	InetVersionUnavailable = Kind{ID: "INET_VERSION_UNAVAILABLE", ExitCode: ExInternetUnavailable}
	InetVersionEmpty       = Kind{ID: "INET_VERSION_EMPTY", ExitCode: ExInternetConfig}

	RuntimeEnable  = Kind{ID: "RUNTIME_ENABLE", ExitCode: ExRuntimeError}
	RuntimeCache   = Kind{ID: "RUNTIME_CACHE", ExitCode: ExRuntimeError}
	RuntimeRestart = Kind{ID: "RUNTIME_RESTART", ExitCode: ExRuntimeError}

	SvcCheckTimeout = Kind{ID: "SVC_CHECK_TIMEOUT", ExitCode: ExSvcTimeout}
	SvcTimeout      = Kind{ID: "SVC_TIMEOUT", ExitCode: ExSvcTimeout}
	SvcList         = Kind{ID: "SVC_LIST", ExitCode: ExSvcError}
	SvcTunnelStart  = Kind{ID: "SVC_TUNNEL_START", ExitCode: ExSvcError}
	SvcTunnelStop   = Kind{ID: "SVC_TUNNEL_STOP", ExitCode: ExSvcError}
	SvcURLTimeout   = Kind{ID: "SVC_URL_TIMEOUT", ExitCode: ExSvcTimeout}
	SvcNotFound     = Kind{ID: "SVC_NOT_FOUND", ExitCode: ExSvcNotFound}

	EnvDriverConflict    = Kind{ID: "ENV_DRIVER_CONFLICT", ExitCode: ExDriverConflict}
	EnvMultiConflict     = Kind{ID: "ENV_MULTINODE_CONFLICT", ExitCode: ExGuestConflict}
	EnvDockerUnavailable = Kind{ID: "ENV_DOCKER_UNAVAILABLE", ExitCode: ExRuntimeUnavailable}
	EnvPodmanUnavailable = Kind{ID: "ENV_PODMAN_UNAVAILABLE", ExitCode: ExRuntimeUnavailable}

	AddonUnsupported = Kind{ID: "SVC_ADDON_UNSUPPORTED", ExitCode: ExSvcUnsupported}
	AddonNotEnabled  = Kind{ID: "SVC_ADDON_NOT_ENABLED", ExitCode: ExProgramConflict}

	KubernetesInstallFailed = Kind{ID: "K8S_INSTALL_FAILED", ExitCode: ExControlPlaneError}
	KubernetesTooOld        = Kind{ID: "K8S_OLD_UNSUPPORTED", ExitCode: ExControlPlaneUnsupported}
	KubernetesDowngrade     = Kind{
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
