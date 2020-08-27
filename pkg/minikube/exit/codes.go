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
Copyright 2019 The Service Authors All rights reserved.

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

// exit contains exit codes used by minikube
package exit

const (
	// Reserved UNIX exit codes
	Failure     = 1 // Failure represents a general failure code
	Interrupted = 2 // Ctrl-C (SIGINT)
	// 3-7 are reserved for crazy legacy codes returned by "minikube status"

	// How to assign new minikube exit codes:
	//
	// * Each error source is indexed from 10 onward, in general, it follows the dependency stack
	// * For each error source, we roughly try to follow sysexits(3) for backwards compatibility
	//
	// <source> + Error			0	(~EX_SOFTWARE)
	// <source> + Conflict		1	(~EX_OSERR)
	// <source> + Timeout		2	(~EX_INTERRUPTED)
	// <source> + NotRunning    3
	// <source> + Usage			4	(~EX_USAGE)
	// <source> + NotFound		5	(~EX_DATAERR)
	// <source> + Unsupported   6   (~EX_PROTOCOL)
	// <source> + Permission	7	(~EX_NOPERM)
	// <source> + Config		8	(~EX_CONFIG)
	// <source> + Unavailable	9	(~EX_UNAVAILABLE)
	//
	// NOTE: "3" and "6" are available for your own use

	// Error codes specific to the minikube program
	ProgramError       = 10 // generic error
	ProgramUsage       = 14 // bad command-line options
	ProgramConflict    = 11 // can't do what you want because of existing data
	ProgramNotFound    = 15 // something was not found
	ProgramUnsupported = 16 // unsupported features
	ProgramConfig      = 18 // bad configuration specified

	// Error codes specific to resource limits (exit code layout follows no rules)
	InsufficientMemory     = 23
	InsufficientStorage    = 26
	InsufficientPermission = 27
	InsufficientCores      = 29

	// Error codes specific to the host
	HostError       = 30
	HostConflict    = 31
	HostTimeout     = 32
	HostUsage       = 34
	HostNotFound    = 35
	HostUnsupported = 38
	HostPermission  = 37
	HostConfig      = 38

	// Error codes specific to remote networking
	InternetError       = 40
	InternetConflict    = 41
	InternetTimeout     = 42
	InternetNotFound    = 45
	InternetConfig      = 48
	InternetUnavailable = 49

	// Error codes specific to the libmachine driver
	DriverError       = 50
	DriverConflict    = 51
	DriverTimeout     = 52
	DriverUsage       = 54
	DriverNotFound    = 55
	DriverUnsupported = 56
	DriverPermission  = 57
	DriverConfig      = 58
	DriverUnavailable = 59

	// Error codes specific to the driver provider
	ProviderError      = 60
	ProviderConflict   = 61
	ProviderTimeout    = 62
	ProviderNotRunning = 63
	// Reserve 64 for the moment as it used to be usage
	ProviderNotFound    = 65
	ProviderUnsupported = 66
	ProviderPermission  = 67
	ProviderConfig      = 68
	ProviderUnavailable = 69 // In common use

	// Error codes specific to local networking
	LocalNetworkError       = 70
	LocalNetworkConflict    = 71
	LocalNetworkTimeout     = 72
	LocalNetworkNotFound    = 75
	LocalNetworkPermission  = 77
	LocalNetworkConfig      = 78
	LocalNetworkUnavailable = 79

	// Error codes specific to the guest host
	GuestError       = 80
	GuestConflict    = 81
	GuestNotRunning  = 83
	GuestNotFound    = 85
	GuestPermission  = 87
	GuestConfig      = 88
	GuestUnavailable = 89

	// Error codes specific to the container runtime
	RuntimeError       = 90
	RuntimeNotRunning  = 93
	RuntimeNotFound    = 95
	RuntimeUnavailable = 99

	// Error codes specific to the Kubernetes control plane
	ControlPlaneError       = 100
	ControlPlaneConflict    = 101
	ControlPlaneTimeout     = 102
	ControlPlaneNotRunning  = 103
	ControlPlaneNotFound    = 105
	ControlPlaneConfig      = 108
	ControlPlaneUnavailable = 109

	// Error codes specific to a Kubernetes service
	ServiceError       = 110
	ServiceConflict    = 111
	ServiceTimeout     = 112
	ServiceNotRunning  = 113
	ServiceNotFound    = 115
	ServiceUnsupported = 116
	ServicePermission  = 117
	ServiceConfig      = 118
	ServiceUnavailable = 119

	// Reserve 128+ for OS signal based exit codes
)
