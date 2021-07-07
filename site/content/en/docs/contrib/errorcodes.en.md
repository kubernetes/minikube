---
title: "Error Codes"
description: >
  minikube error codes and strings
---


## Error Codes


### Generic Errors
1: ExFailure  
2: ExInterrupted  

### Error codes specific to the minikube program
10: ExProgramError  
14: ExProgramUsage  
11: ExProgramConflict  
15: ExProgramNotFound  
16: ExProgramUnsupported  
18: ExProgramConfig  

### Error codes specific to resource limits (exit code layout follows no rules)
20: ExResourceError  
23: ExInsufficientMemory  
26: ExInsufficientStorage  
27: ExInsufficientPermission  
29: ExInsufficientCores  

### Error codes specific to the host
30: ExHostError  
31: ExHostConflict  
32: ExHostTimeout  
34: ExHostUsage  
35: ExHostNotFound  
38: ExHostUnsupported  
37: ExHostPermission  
38: ExHostConfig  

### Error codes specific to remote networking
40: ExInternetError  
41: ExInternetConflict  
42: ExInternetTimeout  
45: ExInternetNotFound  
48: ExInternetConfig  
49: ExInternetUnavailable  

### Error codes specific to the libmachine driver
50: ExDriverError  
51: ExDriverConflict  
52: ExDriverTimeout  
54: ExDriverUsage  
55: ExDriverNotFound  
56: ExDriverUnsupported  
57: ExDriverPermission  
58: ExDriverConfig  
59: ExDriverUnavailable  

### Error codes specific to the driver provider
60: ExProviderError  
61: ExProviderConflict  
62: ExProviderTimeout  
63: ExProviderNotRunning  
65: ExProviderNotFound  
66: ExProviderUnsupported  
67: ExProviderPermission  
68: ExProviderConfig  
69: ExProviderUnavailable  

### Error codes specific to local networking
70: ExLocalNetworkError  
71: ExLocalNetworkConflict  
72: ExLocalNetworkTimeout  
75: ExLocalNetworkNotFound  
77: ExLocalNetworkPermission  
78: ExLocalNetworkConfig  
79: ExLocalNetworkUnavailable  

### Error codes specific to the guest host
80: ExGuestError  
81: ExGuestConflict  
82: ExGuestTimeout  
83: ExGuestNotRunning  
85: ExGuestNotFound  
86: ExGuestUnsupported  
87: ExGuestPermission  
88: ExGuestConfig  
89: ExGuestUnavailable  

### Error codes specific to the container runtime
90: ExRuntimeError  
93: ExRuntimeNotRunning  
95: ExRuntimeNotFound  
99: ExRuntimeUnavailable  

### Error codes specific to the Kubernetes control plane
100: ExControlPlaneError  
101: ExControlPlaneConflict  
102: ExControlPlaneTimeout  
103: ExControlPlaneNotRunning  
105: ExControlPlaneNotFound  
106: ExControlPlaneUnsupported  
108: ExControlPlaneConfig  
109: ExControlPlaneUnavailable  

### Error codes specific to a Kubernetes service
110: ExSvcError  
111: ExSvcConflict  
112: ExSvcTimeout  
113: ExSvcNotRunning  
115: ExSvcNotFound  
116: ExSvcUnsupported  
117: ExSvcPermission  
118: ExSvcConfig  
119: ExSvcUnavailable  


## Error Strings

"MK_USAGE" (Exit code ExProgramUsage)  
minikube has been passed an incorrect parameter  

"MK_USAGE_NO_PROFILE" (Exit code ExProgramUsage)  
minikube has no current cluster running  

"MK_INTERRUPTED" (Exit code ExProgramConflict)  
minikube was interrupted by an OS signal  

"MK_WRONG_BINARY_WSL" (Exit code ExProgramUnsupported)  
user attempted to run a Windows executable (.exe) inside of WSL rather than using the Linux binary  

"MK_NEW_APICLIENT" (Exit code ExProgramError)  
minikube failed to create a new Docker Machine api client  

"MK_ADDON_DISABLE" (Exit code ExProgramError)  
minikube could not disable an addon, e.g. dashboard addon  

"MK_ADDON_ENABLE" (Exit code ExProgramError)  
minikube could not enable an addon, e.g. dashboard addon  

"MK_ADD_CONFIG" (Exit code ExProgramError)  
minikube failed to update internal configuration, such as the cached images config map  

"MK_BOOTSTRAPPER" (Exit code ExProgramError)  
minikube failed to create a cluster bootstrapper  

"MK_CACHE_LIST" (Exit code ExProgramError)  
minikube failed to list cached images  

"MK_CACHE_LOAD" (Exit code ExProgramError)  
minkube failed to cache and load cached images  

"MK_COMMAND_RUNNER" (Exit code ExProgramError)  
minikube failed to load a Docker Machine CommandRunner  

"MK_COMPLETION" (Exit code ExProgramError)  
minikube failed to generate shell command completion for a supported shell  

"MK_CONFIG_SET" (Exit code ExProgramError)  
minikube failed to set an internal config value  

"MK_CONFIG_UNSET" (Exit code ExProgramError)  
minikube failed to unset an internal config value  

"MK_CONFIG_VIEW" (Exit code ExProgramError)  
minikube failed to view current config values  

"MK_DEL_CONFIG" (Exit code ExProgramError)  
minikybe failed to delete an internal configuration, such as a cached image  

"MK_DOCKER_SCRIPT" (Exit code ExProgramError)  
minikube failed to generate script to activate minikube docker-env  

"MK_BIND_FLAGS" (Exit code ExProgramError)  
an error occurred when viper attempted to bind flags to configuration  

"MK_FORMAT_USAGE" (Exit code ExProgramError)  
minkube was passed an invalid format string in the --format flag  

"MK_GENERATE_DOCS" (Exit code ExProgramError)  
minikube failed to auto-generate markdown-based documentation in the specified folder  

"MK_JSON_MARSHAL" (Exit code ExProgramError)  
minikube failed to marshal a JSON object  

"MK_K8S_CLIENT" (Exit code ExControlPlaneUnavailable)  
minikube failed to create a Kubernetes client set which is necessary for querying the Kubernetes API  

"MK_LIST_CONFIG" (Exit code ExProgramError)  
minikube failed to list some configuration data  

"MK_LOG_FOLLOW" (Exit code ExProgramError)  
minikube failed to follow or watch minikube logs  

"MK_NEW_RUNTIME" (Exit code ExProgramError)  
minikube failed to create an appropriate new runtime based on the driver in use  

"MK_OUTPUT_USAGE" (Exit code ExProgramError)  
minikube was passed an invalid value for the --output command line flag  

"MK_RUNTIME" (Exit code ExProgramError)  
minikube could not configure the runtime in use, or the runtime failed  

"MK_RESERVED_PROFILE" (Exit code ExProgramConflict)  
minikube was passed a reserved keyword as a profile name, which is not allowed  

"MK_ENV_SCRIPT" (Exit code ExProgramError)  
minkube failed to generate script to set or unset minikube-env  

"MK_SHELL_DETECT" (Exit code ExProgramError)  
minikube failed to detect the shell in use  

"MK_STATUS_JSON" (Exit code ExProgramError)  
minikube failed to output JSON-formatted minikube status  

"MK_STATUS_TEXT" (Exit code ExProgramError)  
minikube failed to output minikube status text  

"MK_VIEW_EXEC" (Exit code ExProgramError)  
minikube failed to execute (i.e. fill in values for) a view template for displaying current config  

"MK_VIEW_TMPL" (Exit code ExProgramError)  
minikube failed to create view template for displaying current config  

"MK_YAML_MARSHAL" (Exit code ExProgramError)  
minikube failed to marshal a YAML object  

"MK_CREDENTIALS_NOT_FOUND" (Exit code ExProgramNotFound)  
minikube could not locate credentials needed to utilize an appropriate service, e.g. GCP  

"MK_CREDENTIALS_NOT_NEEDED" (Exit code ExProgramNotFound)  
minikube was passed service credentials when they were not needed, such as when using the GCP Auth addon when running in GCE  

"MK_SEMVER_PARSE" (Exit code ExProgramError)  
minikube found an invalid semver string for kubernetes in the minikube constants  

"MK_DAEMONIZE" (Exit code ExProgramError)  
minikube was unable to daemonize the minikube process  

"RSRC_INSUFFICIENT_CORES" (Exit code ExInsufficientCores)  
insufficient cores available for use by minikube and kubernetes  

"RSRC_DOCKER_CORES" (Exit code ExInsufficientCores)  
insufficient cores available for use by Docker Desktop on Mac  

"RSRC_DOCKER_CORES" (Exit code ExInsufficientCores)  
insufficient cores available for use by Docker Desktop on Windows  

"RSRC_INSUFFICIENT_REQ_MEMORY" (Exit code ExInsufficientMemory)  
insufficient memory (less than the recommended minimum) allocated to minikube  

"RSRC_INSUFFICIENT_SYS_MEMORY" (Exit code ExInsufficientMemory)  
insufficient memory (less than the recommended minimum) available on the system running minikube  

"RSRC_INSUFFICIENT_CONTAINER_MEMORY" (Exit code ExInsufficientMemory)  
insufficient memory available for the driver in use by minikube  

"RSRC_DOCKER_MEMORY" (Exit code ExInsufficientMemory)  
insufficient memory available to Docker Desktop on Windows  

"RSRC_DOCKER_MEMORY" (Exit code ExInsufficientMemory)  
insufficient memory available to Docker Desktop on Mac  

"RSRC_DOCKER_STORAGE" (Exit code ExInsufficientStorage)  
insufficient disk storage available to the docker driver  

"RSRC_PODMAN_STORAGE" (Exit code ExInsufficientStorage)  
insufficient disk storage available to the podman driver  

"RSRC_INSUFFICIENT_STORAGE" (Exit code ExInsufficientStorage)  
insufficient disk storage available for running minikube and kubernetes  

"HOST_HOME_MKDIR" (Exit code ExHostPermission)  
minikube could not create the minikube directory  

"HOST_HOME_CHOWN" (Exit code ExHostPermission)  
minikube could not change permissions for the minikube directory  

"HOST_BROWSER" (Exit code ExHostError)  
minikube failed to open the host browser, such as when running minikube dashboard  

"HOST_CONFIG_LOAD" (Exit code ExHostConfig)  
minikube failed to load cluster config from the host for the profile in use  

"HOST_HOME_PERMISSION" (Exit code ExHostPermission)  
the current user has insufficient permissions to create the minikube profile directory  

"HOST_CURRENT_USER" (Exit code ExHostConfig)  
minikube failed to determine current user  

"HOST_DEL_CACHE" (Exit code ExHostError)  
minikube failed to delete cached images from host  

"HOST_KILL_MOUNT_PROC" (Exit code ExHostError)  
minikube failed to kill a mount process  

"HOST_KUBECONFIG_UPDATE" (Exit code ExHostConfig)  
minikube failed to update host Kubernetes resources config  

"HOST_KUBECONFIG_DELETE_CTX" (Exit code ExHostConfig)  
minikube failed to delete Kubernetes config from context for a given profile  

"HOST_KUBECTL_PROXY" (Exit code ExHostError)  
minikube failed to launch a kubectl proxy  

"HOST_MOUNT_PID" (Exit code ExHostError)  
minikube failed to write mount pid  

"HOST_PATH_MISSING" (Exit code ExHostNotFound)  
minikube was passed a path to a host directory that does not exist  

"HOST_PATH_STAT" (Exit code ExHostError)  
minikube failed to access info for a directory path  

"HOST_PURGE" (Exit code ExHostError)  
minikube failed to purge minikube config directories  

"HOST_SAVE_PROFILE" (Exit code ExHostConfig)  
minikube failed to persist profile config  

"PROVIDER_NOT_FOUND" (Exit code ExProviderNotFound)  
minikube could not find a provider for the selected driver  

"PROVIDER_UNAVAILABLE" (Exit code ExProviderNotFound)  
the host does not support or is improperly configured to support a provider for the selected driver  

"DRV_CP_ENDPOINT" (Exit code ExDriverError)  
minikube failed to access the driver control plane or API endpoint  

"DRV_PORT_FORWARD" (Exit code ExDriverError)  
minikube failed to bind container ports to host ports  

"DRV_UNSUPPORTED_MULTINODE" (Exit code ExDriverConflict)  
the driver in use does not support multi-node clusters  

"DRV_UNSUPPORTED_OS" (Exit code ExDriverUnsupported)  
the specified driver is not supported on the host OS  

"DRV_UNSUPPORTED_PROFILE" (Exit code ExDriverUnsupported)  
the driver in use does not support the selected profile or multiple profiles  

"DRV_NOT_FOUND" (Exit code ExDriverNotFound)  
minikube failed to locate specified driver  

"DRV_NOT_DETECTED" (Exit code ExDriverNotFound)  
minikube could not find a valid driver  

"DRV_NOT_HEALTHY" (Exit code ExDriverNotFound)  
minikube found drivers but none were ready to use  

"DRV_DOCKER_NOT_RUNNING" (Exit code ExDriverNotFound)  
minikube found the docker driver but the docker service was not running  

"DRV_AS_ROOT" (Exit code ExDriverPermission)  
the driver in use is being run as root  

"DRV_NEEDS_ROOT" (Exit code ExDriverPermission)  
the specified driver needs to be run as root  

"GUEST_CACHE_LOAD" (Exit code ExGuestError)  
minikube failed to load cached images  

"GUEST_CERT" (Exit code ExGuestError)  
minikube failed to setup certificates  

"GUEST_CP_CONFIG" (Exit code ExGuestConfig)  
minikube failed to access the control plane  

"GUEST_DELETION" (Exit code ExGuestError)  
minikube failed to properly delete a resource, such as a profile  

"GUEST_IMAGE_LIST" (Exit code ExGuestError)  
minikube failed to list images on the machine  

"GUEST_IMAGE_LOAD" (Exit code ExGuestError)  
minikube failed to pull or load an image  

"GUEST_IMAGE_REMOVE" (Exit code ExGuestError)  
minikube failed to remove an image  

"GUEST_IMAGE_BUILD" (Exit code ExGuestError)  
minikube failed to build an image  

"GUEST_LOAD_HOST" (Exit code ExGuestError)  
minikube failed to load host  

"GUEST_MOUNT" (Exit code ExGuestError)  
minkube failed to create a mount  

"GUEST_MOUNT_CONFLICT" (Exit code ExGuestConflict)  
minkube failed to update a mount  

"GUEST_NODE_ADD" (Exit code ExGuestError)  
minikube failed to add a node to the cluster  

"GUEST_NODE_DELETE" (Exit code ExGuestError)  
minikube failed to remove a node from the cluster  

"GUEST_NODE_PROVISION" (Exit code ExGuestError)  
minikube failed to provision a node  

"GUEST_NODE_RETRIEVE" (Exit code ExGuestNotFound)  
minikube failed to retrieve information for a cluster node  

"GUEST_NODE_START" (Exit code ExGuestError)  
minikube failed to startup a cluster node  

"GUEST_PAUSE" (Exit code ExGuestError)  
minikube failed to pause the cluster process  

"GUEST_PROFILE_DELETION" (Exit code ExGuestError)  
minikube failed to delete a machine profile directory  

"GUEST_PROVISION" (Exit code ExGuestError)  
minikube failed while attempting to provision the guest  

"GUEST_PROVISION_CONTAINER_EXITED" (Exit code ExGuestError)  
docker container exited prematurely during provisioning  

"GUEST_START" (Exit code ExGuestError)  
minikube failed to start a node with current driver  

"GUEST_STATUS" (Exit code ExGuestError)  
minikube failed to get docker machine status  

"GUEST_STOP_TIMEOUT" (Exit code ExGuestTimeout)  
stopping the cluster process timed out  

"GUEST_UNPAUSE" (Exit code ExGuestError)  
minikube failed to unpause the cluster process  

"GUEST_CHECK_PAUSED" (Exit code ExGuestError)  
minikube failed to check if Kubernetes containers are paused  

"GUEST_DRIVER_MISMATCH" (Exit code ExGuestConflict)  
minikube cluster was created used a driver that is incompatible with the driver being requested  

"GUEST_MISSING_CONNTRACK" (Exit code ExGuestUnsupported)  
minikube could not find conntrack on the host, which is required from Kubernetes 1.18 onwards  

"IF_HOST_IP" (Exit code ExLocalNetworkError)  
minikube failed to get the host IP to use from within the VM  

"IF_MOUNT_IP" (Exit code ExLocalNetworkError)  
minikube failed to parse the input IP address for mount  

"IF_MOUNT_PORT" (Exit code ExLocalNetworkError)  
minikube failed to parse or find port for mount  

"IF_SSH_CLIENT" (Exit code ExLocalNetworkError)  
minikube failed to access an ssh client on the host machine  

"INET_CACHE_BINARIES" (Exit code ExInternetError)  
minikube failed to cache kubernetes binaries for the current runtime  

"INET_CACHE_KUBECTL" (Exit code ExInternetError)  
minikube failed to cache the kubectl binary  

"INET_CACHE_TAR" (Exit code ExInternetError)  
minikube failed to cache required images to tar files  

"INET_REPO" (Exit code ExInternetError)  
minikube was unable to access main repository and mirrors for images  

"INET_REPOS_UNAVAILABLE" (Exit code ExInternetError)  
minikube was unable to access any known image repositories  

"INET_VERSION_UNAVAILABLE" (Exit code ExInternetUnavailable)  
minikube was unable to fetch latest release/version info for minkikube  

"INET_VERSION_EMPTY" (Exit code ExInternetConfig)  
minikube received invalid empty data for latest release/version info from the server  

"RUNTIME_ENABLE" (Exit code ExRuntimeError)  
minikube failed to enable the current container runtime  

"RUNTIME_CACHE" (Exit code ExRuntimeError)  
minikube failed to cache images for the current container runtime  

"SVC_CHECK_TIMEOUT" (Exit code ExSvcTimeout)  
service check timed out while starting minikube dashboard  

"SVC_TIMEOUT" (Exit code ExSvcTimeout)  
minikube was unable to access a service  

"SVC_LIST" (Exit code ExSvcError)  
minikube failed to list services for the specified namespace  

"SVC_TUNNEL_START" (Exit code ExSvcError)  
minikube failed to start a tunnel  

"SVC_TUNNEL_STOP" (Exit code ExSvcError)  
minikube could not stop an active tunnel  

"SVC_URL_TIMEOUT" (Exit code ExSvcTimeout)  
minikube was unable to access the service url  

"SVC_NOT_FOUND" (Exit code ExSvcNotFound)  
minikube couldn't find the specified service in the specified namespace  

"ENV_DRIVER_CONFLICT" (Exit code ExDriverConflict)  
user attempted to use a command that is not supported by the driver currently in use  

"ENV_MULTINODE_CONFLICT" (Exit code ExGuestConflict)  
user attempted to run a command that is not supported on multi-node setup without some additional configuration  

"ENV_PODMAN_UNAVAILABLE" (Exit code ExRuntimeUnavailable)  
the podman service was unavailable to the cluster  

"SVC_ADDON_UNSUPPORTED" (Exit code ExSvcUnsupported)  
user attempted to use an addon that is not supported  

"SVC_ADDON_NOT_ENABLED" (Exit code ExProgramConflict)  
user attempted to use an addon that is currently not enabled  

"K8S_INSTALL_FAILED" (Exit code ExControlPlaneError)  
minikube failed to update the Kubernetes cluster  

"K8S_INSTALL_FAILED_CONTAINER_RUNTIME_NOT_RUNNING" (Exit code ExRuntimeNotRunning)  
minikube failed to update the Kubernetes cluster because the container runtime was unavailable  

"K8S_OLD_UNSUPPORTED" (Exit code ExControlPlaneUnsupported)  
an outdated Kubernetes version was specified for minikube to use  

"K8S_DOWNGRADE_UNSUPPORTED" (Exit code ExControlPlaneUnsupported)  
minikube was unable to safely downgrade installed Kubernetes version  

