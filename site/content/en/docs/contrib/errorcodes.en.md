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

"MK_WRONG_BINARY_WSL" (Exit code ExProgramUnsupported)  

"MK_WRONG_BINARY_M1" (Exit code ExProgramUnsupported)  

"MK_NEW_APICLIENT" (Exit code ExProgramError)  

"MK_ADDON_ENABLE" (Exit code ExProgramError)  

"MK_ADD_CONFIG" (Exit code ExProgramError)  

"MK_BIND_FLAGS" (Exit code ExProgramError)  

"MK_BOOTSTRAPPER" (Exit code ExProgramError)  

"MK_CACHE_LIST" (Exit code ExProgramError)  

"MK_CACHE_LOAD" (Exit code ExProgramError)  

"MK_COMMAND_RUNNER" (Exit code ExProgramError)  

"MK_COMPLETION" (Exit code ExProgramError)  

"MK_CONFIG_SET" (Exit code ExProgramError)  

"MK_CONFIG_UNSET" (Exit code ExProgramError)  

"MK_CONFIG_VIEW" (Exit code ExProgramError)  

"MK_DEL_CONFIG" (Exit code ExProgramError)  

"MK_DISABLE" (Exit code ExProgramError)  

"MK_DOCKER_SCRIPT" (Exit code ExProgramError)  

"MK_ENABLE" (Exit code ExProgramError)  

"MK_FLAGS_BIND" (Exit code ExProgramError)  

"MK_FLAGS_SET" (Exit code ExProgramError)  

"MK_FORMAT_USAGE" (Exit code ExProgramError)  

"MK_GENERATE_DOCS" (Exit code ExProgramError)  

"MK_JSON_MARSHAL" (Exit code ExProgramError)  

"MK_K8S_CLIENT" (Exit code ExControlPlaneUnavailable)  

"MK_LIST_CONFIG" (Exit code ExProgramError)  

"MK_LOGTOSTDERR_FLAG" (Exit code ExProgramError)  

"MK_LOG_FOLLOW" (Exit code ExProgramError)  

"MK_NEW_RUNTIME" (Exit code ExProgramError)  

"MK_OUTPUT_USAGE" (Exit code ExProgramError)  

"MK_RUNTIME" (Exit code ExProgramError)  

"MK_RESERVED_PROFILE" (Exit code ExProgramConflict)  

"MK_ENV_SCRIPT" (Exit code ExProgramError)  

"MK_SHELL_DETECT" (Exit code ExProgramError)  

"MK_STATUS_JSON" (Exit code ExProgramError)  

"MK_STATUS_TEXT" (Exit code ExProgramError)  

"MK_UNSET_SCRIPT" (Exit code ExProgramError)  

"MK_VIEW_EXEC" (Exit code ExProgramError)  

"MK_VIEW_TMPL" (Exit code ExProgramError)  

"MK_YAML_MARSHAL" (Exit code ExProgramError)  

"MK_CREDENTIALS_NOT_FOUND" (Exit code ExProgramNotFound)  

"MK_CREDENTIALS_NOT_NEEDED" (Exit code ExProgramNotFound)  

"MK_SEMVER_PARSE" (Exit code ExProgramError)  

"MK_DAEMONIZE" (Exit code ExProgramError)  

"RSRC_INSUFFICIENT_CORES" (Exit code ExInsufficientCores)  

"RSRC_DOCKER_CORES" (Exit code ExInsufficientCores)  

"RSRC_DOCKER_CORES" (Exit code ExInsufficientCores)  

"RSRC_INSUFFICIENT_REQ_MEMORY" (Exit code ExInsufficientMemory)  

"RSRC_INSUFFICIENT_SYS_MEMORY" (Exit code ExInsufficientMemory)  

"RSRC_INSUFFICIENT_CONTAINER_MEMORY" (Exit code ExInsufficientMemory)  

"RSRC_DOCKER_MEMORY" (Exit code ExInsufficientMemory)  

"RSRC_DOCKER_MEMORY" (Exit code ExInsufficientMemory)  

"RSRC_DOCKER_STORAGE" (Exit code ExInsufficientStorage)  

"RSRC_PODMAN_STORAGE" (Exit code ExInsufficientStorage)  

"RSRC_INSUFFICIENT_STORAGE" (Exit code ExInsufficientStorage)  

"HOST_HOME_MKDIR" (Exit code ExHostPermission)  

"HOST_HOME_CHOWN" (Exit code ExHostPermission)  

"HOST_BROWSER" (Exit code ExHostError)  

"HOST_CONFIG_LOAD" (Exit code ExHostConfig)  

"HOST_HOME_PERMISSION" (Exit code ExHostPermission)  

"HOST_CURRENT_USER" (Exit code ExHostConfig)  

"HOST_DEL_CACHE" (Exit code ExHostError)  

"HOST_KILL_MOUNT_PROC" (Exit code ExHostError)  

"HOST_KUBECNOFIG_UNSET" (Exit code ExHostConfig)  

"HOST_KUBECONFIG_UPDATE" (Exit code ExHostConfig)  

"HOST_KUBECONFIG_DELETE_CTX" (Exit code ExHostConfig)  

"HOST_KUBECTL_PROXY" (Exit code ExHostError)  

"HOST_MOUNT_PID" (Exit code ExHostError)  

"HOST_PATH_MISSING" (Exit code ExHostNotFound)  

"HOST_PATH_STAT" (Exit code ExHostError)  

"HOST_PURGE" (Exit code ExHostError)  

"HOST_SAVE_PROFILE" (Exit code ExHostConfig)  

"PROVIDER_NOT_FOUND" (Exit code ExProviderNotFound)  

"PROVIDER_UNAVAILABLE" (Exit code ExProviderNotFound)  

"DRV_CP_ENDPOINT" (Exit code ExDriverError)  

"DRV_PORT_FORWARD" (Exit code ExDriverError)  

"DRV_UNSUPPORTED_MULTINODE" (Exit code ExDriverConflict)  

"DRV_UNSUPPORTED_OS" (Exit code ExDriverUnsupported)  

"DRV_UNSUPPORTED_PROFILE" (Exit code ExDriverUnsupported)  

"DRV_NOT_FOUND" (Exit code ExDriverNotFound)  

"DRV_NOT_DETECTED" (Exit code ExDriverNotFound)  

"DRV_NOT_HEALTHY" (Exit code ExDriverNotFound)  

"DRV_DOCKER_NOT_RUNNING" (Exit code ExDriverNotFound)  

"DRV_AS_ROOT" (Exit code ExDriverPermission)  

"DRV_NEEDS_ROOT" (Exit code ExDriverPermission)  

"DRV_NEEDS_ADMINISTRATOR" (Exit code ExDriverPermission)  

"GUEST_CACHE_LOAD" (Exit code ExGuestError)  

"GUEST_CERT" (Exit code ExGuestError)  

"GUEST_CP_CONFIG" (Exit code ExGuestConfig)  

"GUEST_DELETION" (Exit code ExGuestError)  

"GUEST_IMAGE_LIST" (Exit code ExGuestError)  

"GUEST_IMAGE_LOAD" (Exit code ExGuestError)  

"GUEST_IMAGE_REMOVE" (Exit code ExGuestError)  

"GUEST_IMAGE_BUILD" (Exit code ExGuestError)  

"GUEST_LOAD_HOST" (Exit code ExGuestError)  

"GUEST_MOUNT" (Exit code ExGuestError)  

"GUEST_MOUNT_CONFLICT" (Exit code ExGuestConflict)  

"GUEST_NODE_ADD" (Exit code ExGuestError)  

"GUEST_NODE_DELETE" (Exit code ExGuestError)  

"GUEST_NODE_PROVISION" (Exit code ExGuestError)  

"GUEST_NODE_RETRIEVE" (Exit code ExGuestNotFound)  

"GUEST_NODE_START" (Exit code ExGuestError)  

"GUEST_PAUSE" (Exit code ExGuestError)  

"GUEST_PROFILE_DELETION" (Exit code ExGuestError)  

"GUEST_PROVISION" (Exit code ExGuestError)  

"GUEST_PROVISION_CONTAINER_EXITED" (Exit code ExGuestError)  

"GUEST_START" (Exit code ExGuestError)  

"GUEST_STATUS" (Exit code ExGuestError)  

"GUEST_STOP_TIMEOUT" (Exit code ExGuestTimeout)  

"GUEST_UNPAUSE" (Exit code ExGuestError)  

"GUEST_CHECK_PAUSED" (Exit code ExGuestError)  

"GUEST_DRIVER_MISMATCH" (Exit code ExGuestConflict)  

"GUEST_MISSING_CONNTRACK" (Exit code ExGuestUnsupported)  

"IF_HOST_IP" (Exit code ExLocalNetworkError)  

"IF_MOUNT_IP" (Exit code ExLocalNetworkError)  

"IF_MOUNT_PORT" (Exit code ExLocalNetworkError)  

"IF_SSH_CLIENT" (Exit code ExLocalNetworkError)  

"INET_CACHE_BINARIES" (Exit code ExInternetError)  

"INET_CACHE_KUBECTL" (Exit code ExInternetError)  

"INET_CACHE_TAR" (Exit code ExInternetError)  

"INET_GET_VERSIONS" (Exit code ExInternetError)  

"INET_REPO" (Exit code ExInternetError)  

"INET_REPOS_UNAVAILABLE" (Exit code ExInternetError)  

"INET_VERSION_UNAVAILABLE" (Exit code ExInternetUnavailable)  

"INET_VERSION_EMPTY" (Exit code ExInternetConfig)  

"RUNTIME_ENABLE" (Exit code ExRuntimeError)  

"RUNTIME_CACHE" (Exit code ExRuntimeError)  

"RUNTIME_RESTART" (Exit code ExRuntimeError)  

"SVC_CHECK_TIMEOUT" (Exit code ExSvcTimeout)  

"SVC_TIMEOUT" (Exit code ExSvcTimeout)  

"SVC_LIST" (Exit code ExSvcError)  

"SVC_TUNNEL_START" (Exit code ExSvcError)  

"SVC_TUNNEL_STOP" (Exit code ExSvcError)  

"SVC_URL_TIMEOUT" (Exit code ExSvcTimeout)  

"SVC_NOT_FOUND" (Exit code ExSvcNotFound)  

"ENV_DRIVER_CONFLICT" (Exit code ExDriverConflict)  

"ENV_MULTINODE_CONFLICT" (Exit code ExGuestConflict)  

"ENV_DOCKER_UNAVAILABLE" (Exit code ExRuntimeUnavailable)  

"ENV_PODMAN_UNAVAILABLE" (Exit code ExRuntimeUnavailable)  

"SVC_ADDON_UNSUPPORTED" (Exit code ExSvcUnsupported)  

"SVC_ADDON_NOT_ENABLED" (Exit code ExProgramConflict)  

"K8S_INSTALL_FAILED" (Exit code ExControlPlaneError)  

"K8S_INSTALL_FAILED_CONTAINER_RUNTIME_NOT_RUNNING" (Exit code ExRuntimeNotRunning)  

"K8S_OLD_UNSUPPORTED" (Exit code ExControlPlaneUnsupported)  

"K8S_DOWNGRADE_UNSUPPORTED" (Exit code ExControlPlaneUnsupported)

"K8S_RELEASE_FETCH_FAILED"  (Exit code ExControlPlaneUnsupported)

"K8S_VERSION_NOT_FOUND"     (Exit code ExControlPlaneUnsupported)