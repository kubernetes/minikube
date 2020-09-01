/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"testing"
)

var (
	daemonResponseMock   string
	daemonInfoGetterMock = func() (string, error) {
		return daemonResponseMock, nil
	}
)

func TestDockerSystemInfo(t *testing.T) {
	testCases := []struct {
		Name          string // test case bane
		OciBin        string // Docker or Podman
		RawJSON       string // raw response from json
		ShouldError   bool
		CPUs          int
		Memory        int64
		OS            string
		Swarm         bool
		StorageDriver string
	}{
		{
			Name:          "linux_docker",
			OciBin:        "docker",
			RawJSON:       `{"ID":"7PYP:53DU:MLWX:EDQG:YG2Y:UJLB:J7SD:4SAI:XF2Y:N2MR:MU53:DR3N","Containers":3,"ContainersRunning":1,"ContainersPaused":0,"ContainersStopped":2,"Images":76,"Driver":"overlay2","DriverStatus":[["Backing Filesystem","extfs"],["Supports d_type","true"],["Native Overlay Diff","true"]],"SystemStatus":null,"Plugins":{"Volume":["local"],"Network":["bridge","host","macvlan","null","overlay"],"Authorization":null,"Log":["awslogs","fluentd","gcplogs","gelf","journald","json-file","local","logentries","splunk","syslog"]},"MemoryLimit":true,"SwapLimit":false,"KernelMemory":true,"KernelMemoryTCP":false,"CpuCfsPeriod":true,"CpuCfsQuota":true,"CPUShares":true,"CPUSet":true,"PidsLimit":false,"IPv4Forwarding":true,"BridgeNfIptables":true,"BridgeNfIp6tables":true,"Debug":false,"NFd":27,"OomKillDisable":true,"NGoroutines":48,"SystemTime":"2020-08-11T18:16:17.494440681Z","LoggingDriver":"json-file","CgroupDriver":"cgroupfs","NEventsListener":0,"KernelVersion":"4.9.0-8-amd64","OperatingSystem":"Debian GNU/Linux 9 (stretch)","OSType":"linux","Architecture":"x86_64","IndexServerAddress":"https://index.docker.io/v1/","RegistryConfig":{"AllowNondistributableArtifactsCIDRs":[],"AllowNondistributableArtifactsHostnames":[],"InsecureRegistryCIDRs":["127.0.0.0/8"],"IndexConfigs":{"docker.io":{"Name":"docker.io","Mirrors":[],"Secure":true,"Official":true}},"Mirrors":[]},"NCPU":16,"MemTotal":63336071168,"GenericResources":null,"DockerRootDir":"/var/lib/docker","HttpProxy":"","HttpsProxy":"","NoProxy":"","Name":"image-builder-cloud-shell-v20200811-102837","Labels":[],"ExperimentalBuild":false,"ServerVersion":"18.09.0","ClusterStore":"","ClusterAdvertise":"","Runtimes":{"runc":{"path":"runc"}},"DefaultRuntime":"runc","Swarm":{"NodeID":"","NodeAddr":"","LocalNodeState":"inactive","ControlAvailable":false,"Error":"","RemoteManagers":null},"LiveRestoreEnabled":false,"Isolation":"","InitBinary":"docker-init","ContainerdCommit":{"ID":"7ad184331fa3e55e52b890ea95e65ba581ae3429","Expected":"7ad184331fa3e55e52b890ea95e65ba581ae3429"},"RuncCommit":{"ID":"dc9208a3303feef5b3839f4323d9beb36df0a9dd","Expected":"dc9208a3303feef5b3839f4323d9beb36df0a9dd"},"InitCommit":{"ID":"fec3683","Expected":"fec3683"},"SecurityOptions":["name=seccomp,profile=default"],"ProductLicense":"Community Engine","Warnings":["WARNING: No swap limit support"],"ClientInfo":{"Debug":false,"Plugins":[],"Warnings":null}}`,
			ShouldError:   false,
			CPUs:          16,
			Memory:        63336071168,
			OS:            "linux",
			Swarm:         false,
			StorageDriver: "overlay2",
		},
		{
			Name:   "macos_docker",
			OciBin: "docker",
			RawJSON: `{"ID":"T54Z:I56K:XRG5:BTMK:BI72:IMI3:QBBF:H2PD:DGAF:EQLJ:7JFZ:PF54","Containers":5,"ContainersRunning":1,"ContainersPaused":0,"ContainersStopped":4,"Images":84,"Driver":"overlay2","DriverStatus":[["Backing Filesystem","extfs"],["Supports d_type","true"],["Native Overlay Diff","true"]],"SystemStatus":null,"Plugins":{"Volume":["local"],"Network":["bridge","host","ipvlan","macvlan","null","overlay"],"Authorization":null,"Log":["awslogs","fluentd","gcplogs","gelf","journald","json-file","local","logentries","splunk","syslog"]},"MemoryLimit":true,"SwapLimit":true,"KernelMemory":true,"KernelMemoryTCP":true,"CpuCfsPeriod":true,"CpuCfsQuota":true,"CPUShares":true,"CPUSet":true,"PidsLimit":true,"IPv4Forwarding":true,"BridgeNfIptables":true,"BridgeNfIp6tables":true,"Debug":true,"NFd":46,"OomKillDisable":true,"NGoroutines":56,"SystemTime":"2020-08-11T19:33:23.8936297Z","LoggingDriver":"json-file","CgroupDriver":"cgroupfs","NEventsListener":3,"KernelVersion":"4.19.76-linuxkit","OperatingSystem":"Docker Desktop","OSType":"linux","Architecture":"x86_64","IndexServerAddress":"https://index.docker.io/v1/","RegistryConfig":{"AllowNondistributableArtifactsCIDRs":[],"AllowNondistributableArtifactsHostnames":[],"InsecureRegistryCIDRs":["127.0.0.0/8"],"IndexConfigs":{"docker.io":{"Name":"docker.io","Mirrors":[],"Secure":true,"Official":true}},"Mirrors":[]},"NCPU":4,"MemTotal":3142250496,"GenericResources":null,"DockerRootDir":"/var/lib/docker","HttpProxy":"gateway.docker.internal:3128","HttpsProxy":"gateway.docker.internal:3129","NoProxy":"","Name":"docker-desktop","Labels":[],"ExperimentalBuild":false,"ServerVersion":"19.03.12","ClusterStore":"","ClusterAdvertise":"","Runtimes":{"runc":{"path":"runc"}},"DefaultRuntime":"runc","Swarm":{"NodeID":"","NodeAddr":"","LocalNodeState":"inactive","ControlAvailable":false,"Error":"","RemoteManagers":null},"LiveRestoreEnabled":false,"Isolation":"","InitBinary":"docker-init","ContainerdCommit":{"ID":"7ad184331fa3e55e52b890ea95e65ba581ae3429","Expected":"7ad184331fa3e55e52b890ea95e65ba581ae3429"},"RuncCommit":{"ID":"dc9208a3303feef5b3839f4323d9beb36df0a9dd","Expected":"dc9208a3303feef5b3839f4323d9beb36df0a9dd"},"InitCommit":{"ID":"fec3683","Expected":"fec3683"},"SecurityOptions":["name=seccomp,profile=default"],"ProductLicense":"Community Engine","Warnings":null,"ClientInfo":{"Debug":false,"Plugins":[],"Warnings":null}}
`,
			ShouldError:   false,
			CPUs:          4,
			Memory:        3142250496,
			OS:            "linux",
			Swarm:         false,
			StorageDriver: "overlay2",
		},
		{
			Name:   "windows_docker",
			OciBin: "docker",
			RawJSON: `{"ID":"CVVH:7ZIB:S5EO:L6VO:MGZ3:TRLS:JGIS:4ZI2:27Z7:MQAQ:YSLT:HEHB","Containers":0,"ContainersRunning":0,"ContainersPaused":0,"ContainersStopped":0,"Images":3,"Driver":"overlay2","DriverStatus":[["Backing Filesystem","extfs"],["Supports d_type","true"],["Native Overlay Diff","true"]],"SystemStatus":null,"Plugins":{"Volume":["local"],"Network":["bridge","host","ipvlan","macvlan","null","overlay"],"Authorization":null,"Log":["awslogs","fluentd","gcplogs","gelf","journald","json-file","local","logentries","splunk","syslog"]},"MemoryLimit":true,"SwapLimit":true,"KernelMemory":true,"KernelMemoryTCP":true,"CpuCfsPeriod":true,"CpuCfsQuota":true,"CPUShares":true,"CPUSet":true,"PidsLimit":true,"IPv4Forwarding":true,"BridgeNfIptables":true,"BridgeNfIp6tables":true,"Debug":true,"NFd":35,"OomKillDisable":true,"NGoroutines":45,"SystemTime":"2020-08-11T19:39:26.083212722Z","LoggingDriver":"json-file","CgroupDriver":"cgroupfs","NEventsListener":1,"KernelVersion":"4.19.76-linuxkit","OperatingSystem":"Docker Desktop","OSType":"linux","Architecture":"x86_64","IndexServerAddress":"https://index.docker.io/v1/","RegistryConfig":{"AllowNondistributableArtifactsCIDRs":[],"AllowNondistributableArtifactsHostnames":[],"InsecureRegistryCIDRs":["127.0.0.0/8"],"IndexConfigs":{"docker.io":{"Name":"docker.io","Mirrors":[],"Secure":true,"Official":true}},"Mirrors":[]},"NCPU":4,"MemTotal":10454695936,"GenericResources":null,"DockerRootDir":"/var/lib/docker","HttpProxy":"","HttpsProxy":"","NoProxy":"","Name":"docker-desktop","Labels":[],"ExperimentalBuild":false,"ServerVersion":"19.03.12","ClusterStore":"","ClusterAdvertise":"","Runtimes":{"runc":{"path":"runc"}},"DefaultRuntime":"runc","Swarm":{"NodeID":"","NodeAddr":"","LocalNodeState":"inactive","ControlAvailable":false,"Error":"","RemoteManagers":null},"LiveRestoreEnabled":false,"Isolation":"","InitBinary":"docker-init","ContainerdCommit":{"ID":"7ad184331fa3e55e52b890ea95e65ba581ae3429","Expected":"7ad184331fa3e55e52b890ea95e65ba581ae3429"},"RuncCommit":{"ID":"dc9208a3303feef5b3839f4323d9beb36df0a9dd","Expected":"dc9208a3303feef5b3839f4323d9beb36df0a9dd"},"InitCommit":{"ID":"fec3683","Expected":"fec3683"},"SecurityOptions":["name=seccomp,profile=default"],"ProductLicense":"Community Engine","Warnings":null,"ClientInfo":{"Debug":false,"Plugins":[],"Warnings":null}}
`,
			ShouldError:   false,
			CPUs:          4,
			Memory:        10454695936,
			OS:            "linux",
			Swarm:         false,
			StorageDriver: "overlay2",
		},
		{
			Name:   "podman_1.8_linux",
			OciBin: "podman",
			RawJSON: `{
				"host": {
					"BuildahVersion": "1.13.1",
					"CgroupVersion": "v1",
					"Conmon": {
						"package": "conmon: /usr/libexec/podman/conmon",
						"path": "/usr/libexec/podman/conmon",
						"version": "conmon version 2.0.10, commit: unknown"
					},
					"Distribution": {
						"distribution": "debian",
						"version": "10"
					},
					"MemFree": 4907147264,
					"MemTotal": 7839653888,
					"OCIRuntime": {
						"name": "runc",
						"package": "runc: /usr/sbin/runc",
						"path": "/usr/sbin/runc",
						"version": "runc version 1.0.0~rc6+dfsg1\ncommit: 1.0.0~rc6+dfsg1-3 spec: 1.0.1"
					},
					"SwapFree": 0,
					"SwapTotal": 0,
					"arch": "amd64",
					"cpus": 2,
					"eventlogger": "journald",
					"hostname": "podman-exp-temp",
					"kernel": "4.19.0-8-cloud-amd64",
					"os": "linux",
					"rootless": false,
					"uptime": "2690h 47m 23.31s (Approximately 112.08 days)"
				},
				"registries": {
					"search": [
						"docker.io",
						"quay.io"
					]
				},
				"store": {
					"ConfigFile": "/etc/containers/storage.conf",
					"ContainerStore": {
						"number": 1
					},
					"GraphDriverName": "overlay",
					"GraphOptions": {},
					"GraphRoot": "/var/lib/containers/storage",
					"GraphStatus": {
						"Backing Filesystem": "extfs",
					  "Native Overlay Diff": "true",
						"Supports d_type": "true",
						"Using metacopy": "false"
					},
					"ImageStore": {
						"number": 2
					},
					"RunRoot": "/var/run/containers/storage",
					"VolumePath": "/var/lib/containers/storage/volumes"
				}
			}
`, CPUs: 2,
			Memory:        7839653888,
			OS:            "linux",
			Swarm:         false,
			StorageDriver: "overlay",
		},
		{
			Name:          "mac_swarm_enabled",
			OciBin:        "docker",
			RawJSON:       `{"ID":"HR5B:K7CR:NGTA:ZSHO:4Z3Q:2K7I:PUQ6:H752:FA6K:YF5R:PPU4:OAIQ","Containers":7,"ContainersRunning":1,"ContainersPaused":0,"ContainersStopped":6,"Images":168,"Driver":"overlay2","DriverStatus":[["Backing Filesystem","extfs"],["Supports d_type","true"],["Native Overlay Diff","true"]],"SystemStatus":null,"Plugins":{"Volume":["local"],"Network":["bridge","host","ipvlan","macvlan","null","overlay"],"Authorization":null,"Log":["awslogs","fluentd","gcplogs","gelf","journald","json-file","local","logentries","splunk","syslog"]},"MemoryLimit":true,"SwapLimit":true,"KernelMemory":true,"KernelMemoryTCP":true,"CpuCfsPeriod":true,"CpuCfsQuota":true,"CPUShares":true,"CPUSet":true,"PidsLimit":true,"IPv4Forwarding":true,"BridgeNfIptables":true,"BridgeNfIp6tables":true,"Debug":true,"NFd":53,"OomKillDisable":true,"NGoroutines":162,"SystemTime":"2020-08-19T21:33:00.513075607Z","LoggingDriver":"json-file","CgroupDriver":"cgroupfs","NEventsListener":3,"KernelVersion":"4.19.76-linuxkit","OperatingSystem":"Docker Desktop","OSType":"linux","Architecture":"x86_64","IndexServerAddress":"https://index.docker.io/v1/","RegistryConfig":{"AllowNondistributableArtifactsCIDRs":[],"AllowNondistributableArtifactsHostnames":[],"InsecureRegistryCIDRs":["127.0.0.0/8"],"IndexConfigs":{"docker.io":{"Name":"docker.io","Mirrors":[],"Secure":true,"Official":true}},"Mirrors":[]},"NCPU":4,"MemTotal":3142250496,"GenericResources":null,"DockerRootDir":"/var/lib/docker","HttpProxy":"gateway.docker.internal:3128","HttpsProxy":"gateway.docker.internal:3129","NoProxy":"","Name":"docker-desktop","Labels":[],"ExperimentalBuild":false,"ServerVersion":"19.03.12","ClusterStore":"","ClusterAdvertise":"","Runtimes":{"runc":{"path":"runc"}},"DefaultRuntime":"runc","Swarm":{"NodeID":"v6tstvvbuc0pgyrp44ub6srey","NodeAddr":"192.168.65.3","LocalNodeState":"active","ControlAvailable":true,"Error":"","RemoteManagers":[{"NodeID":"v6tstvvbuc0pgyrp44ub6srey","Addr":"192.168.65.3:2377"}],"Nodes":1,"Managers":1,"Cluster":{"ID":"tp9f8rys388jhhscmcq4hfrer","Version":{"Index":10},"CreatedAt":"2020-08-19T21:30:10.678729711Z","UpdatedAt":"2020-08-19T21:30:11.199400544Z","Spec":{"Name":"default","Labels":{},"Orchestration":{"TaskHistoryRetentionLimit":5},"Raft":{"SnapshotInterval":10000,"KeepOldSnapshots":0,"LogEntriesForSlowFollowers":500,"ElectionTick":10,"HeartbeatTick":1},"Dispatcher":{"HeartbeatPeriod":5000000000},"CAConfig":{"NodeCertExpiry":7776000000000000},"TaskDefaults":{},"EncryptionConfig":{"AutoLockManagers":false}},"TLSInfo":{"TrustRoot":"-----BEGIN CERTIFICATE-----\nMIIBajCCARCgAwIBAgIUCUPy2vYpHAhU3a2rSNOkoMJQn3kwCgYIKoZIzj0EAwIw\nEzERMA8GA1UEAxMIc3dhcm0tY2EwHhcNMjAwODE5MjEyNTAwWhcNNDAwODE0MjEy\nNTAwWjATMREwDwYDVQQDEwhzd2FybS1jYTBZMBMGByqGSM49AgEGCCqGSM49AwEH\nA0IABM8lNy+NqoBOOuEcNrRvWHuFKdGWoxJpnm2zcAULJrG3YjjDpJ/BT2gMxCSa\nMLoKx1+EzYBUEs8rD2vMhB3QNh6jQjBAMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMB\nAf8EBTADAQH/MB0GA1UdDgQWBBRALAC6bczt2ItCs5gicm7SjWVTETAKBggqhkjO\nPQQDAgNIADBFAiEA/shMkzxdxJHbm9PHf5D/1Bqm1A3/DFYeJiWENcqbrT8CIGLC\n86V5u7leZGfP0F3ibGdXH++oc26mI5eUr2czrteY\n-----END CERTIFICATE-----\n","CertIssuerSubject":"MBMxETAPBgNVBAMTCHN3YXJtLWNh","CertIssuerPublicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEzyU3L42qgE464Rw2tG9Ye4Up0ZajEmmebbNwBQsmsbdiOMOkn8FPaAzEJJowugrHX4TNgFQSzysPa8yEHdA2Hg=="},"RootRotationInProgress":false,"DefaultAddrPool":["10.0.0.0/8"],"SubnetSize":24,"DataPathPort":4789}},"LiveRestoreEnabled":false,"Isolation":"","InitBinary":"docker-init","ContainerdCommit":{"ID":"7ad184331fa3e55e52b890ea95e65ba581ae3429","Expected":"7ad184331fa3e55e52b890ea95e65ba581ae3429"},"RuncCommit":{"ID":"dc9208a3303feef5b3839f4323d9beb36df0a9dd","Expected":"dc9208a3303feef5b3839f4323d9beb36df0a9dd"},"InitCommit":{"ID":"fec3683","Expected":"fec3683"},"SecurityOptions":["name=seccomp,profile=default"],"ProductLicense":"Community Engine","Warnings":null,"ClientInfo":{"Debug":false,"Plugins":[],"Warnings":null}}`,
			Memory:        3142250496,
			OS:            "linux",
			ShouldError:   false,
			CPUs:          4,
			Swarm:         true,
			StorageDriver: "overlay2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			daemonResponseMock = tc.RawJSON
			// setting up mock funcs
			dockerInfoGetter = daemonInfoGetterMock
			podmanInfoGetter = daemonInfoGetterMock
			s, err := DaemonInfo(tc.OciBin)

			if err != nil && !tc.ShouldError {
				t.Errorf("Expected not to have error but got %v", err)
			}
			if s.CPUs != tc.CPUs {
				t.Errorf("Expected CPUs to be %d but got %d", tc.CPUs, s.CPUs)
			}
			if s.TotalMemory != tc.Memory {
				t.Errorf("Expected Memory to be %d but got %d", tc.Memory, s.TotalMemory)
			}
			if s.OSType != tc.OS {
				t.Errorf("Expected OS type to be %q but got %q", tc.OS, s.OSType)
			}
			if s.StorageDriver != tc.StorageDriver {
				t.Errorf("Expected OS type to be %q but got %q", tc.StorageDriver, s.StorageDriver)
			}
			if s.Swarm != tc.Swarm {
				t.Errorf("Expected Swarm to be %t but got %t", tc.Swarm, s.Swarm)
			}
		})
	}
}
