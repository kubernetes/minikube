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
Copyright 2019 TheExSvc Authors All rights reserved.

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

const (
	// Reserved UNIX exit codes
	ExFailure     = 1 // Failure represents a general failure code
	ExInterrupted = 2 // Ctrl-C (SIGINT)

	// 3-7 are reserved for crazy legacy codes returned by "minikube status"

	// How to assign new minikube exit codes:
	//
	// * Each error source is indexed from 10 onward, in general, it follows the dependency stack
	// * For each error source, we roughly try to follow sysexits(3) for backwards compatibility
	//
	// errorOff       = 0 // (~EX_SOFTWARE)
	// conflictOff    = 1 // (~EX_OSERR)
	// timeoutOff     = 2 // (~EX_INTERRUPTED)
	// notRunningOff  = 3 // custom
	// usageOff       = 4 // (~EX_USAGE)
	// notFoundOff    = 5 // (~EX_DATAERR)
	// unsupportedOff = 6 // (~EX_PROTOCOL)
	// permissionOff  = 7 // (~EX_NOPERM)
	// configOff      = 8 // (~EX_CONFIG)
	// navailableOff = 9 // (~EX_UNAVAILABLE)

	// Error codes specific to the minikube program
	ExProgramError       = 10 // generic error
	ExProgramUsage       = 14 // bad command-line options
	ExProgramConflict    = 11 // can't do what you want because of existing data
	ExProgramNotFound    = 15 // something was not found
	ExProgramUnsupported = 16 // unsupported features
	ExProgramConfig      = 18 // bad configuration specified

	// Error codes specific to resource limits (exit code layout follows no rules)
	ExResourceError          = 20
	ExInsufficientMemory     = 23
	ExInsufficientStorage    = 26
	ExInsufficientPermission = 27
	ExInsufficientCores      = 29

	// Error codes specific to the host
	ExHostError       = 30
	ExHostConflict    = 31
	ExHostTimeout     = 32
	ExHostUsage       = 34
	ExHostNotFound    = 35
	ExHostUnsupported = 38
	ExHostPermission  = 37
	ExHostConfig      = 38

	// Error codes specific to remote networking
	ExInternetError       = 40
	ExInternetConflict    = 41
	ExInternetTimeout     = 42
	ExInternetNotFound    = 45
	ExInternetConfig      = 48
	ExInternetUnavailable = 49

	// Error codes specific to the libmachine driver
	ExDriverError       = 50
	ExDriverConflict    = 51
	ExDriverTimeout     = 52
	ExDriverUsage       = 54
	ExDriverNotFound    = 55
	ExDriverUnsupported = 56
	ExDriverPermission  = 57
	ExDriverConfig      = 58
	ExDriverUnavailable = 59

	// Error codes specific to the driver provider
	ExProviderError      = 60
	ExProviderConflict   = 61
	ExProviderTimeout    = 62
	ExProviderNotRunning = 63
	// Reserve 64 for the moment as it used to be usage
	ExProviderNotFound    = 65
	ExProviderUnsupported = 66
	ExProviderPermission  = 67
	ExProviderConfig      = 68
	ExProviderUnavailable = 69 // In common use

	// Error codes specific to local networking
	ExLocalNetworkError       = 70
	ExLocalNetworkConflict    = 71
	ExLocalNetworkTimeout     = 72
	ExLocalNetworkNotFound    = 75
	ExLocalNetworkPermission  = 77
	ExLocalNetworkConfig      = 78
	ExLocalNetworkUnavailable = 79

	// Error codes specific to the guest host
	ExGuestError       = 80
	ExGuestConflict    = 81
	ExGuestTimeout     = 82
	ExGuestNotRunning  = 83
	ExGuestNotFound    = 85
	ExGuestUnsupported = 86
	ExGuestPermission  = 87
	ExGuestConfig      = 88
	ExGuestUnavailable = 89

	// Error codes specific to the container runtime
	ExRuntimeError       = 90
	ExRuntimeNotRunning  = 93
	ExRuntimeNotFound    = 95
	ExRuntimeUnavailable = 99

	// Error codes specific to the Kubernetes control plane
	ExControlPlaneError       = 100
	ExControlPlaneConflict    = 101
	ExControlPlaneTimeout     = 102
	ExControlPlaneNotRunning  = 103
	ExControlPlaneNotFound    = 105
	ExControlPlaneUnsupported = 106
	ExControlPlaneConfig      = 108
	ExControlPlaneUnavailable = 109

	// Error codes specific to a Kubernetes service
	ExSvcError       = 110
	ExSvcConflict    = 111
	ExSvcTimeout     = 112
	ExSvcNotRunning  = 113
	ExSvcNotFound    = 115
	ExSvcUnsupported = 116
	ExSvcPermission  = 117
	ExSvcConfig      = 118
	ExSvcUnavailable = 119

	// Reserve 128+ for OS signal based exit codes
)
