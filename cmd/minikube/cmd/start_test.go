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

package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/klog/v2"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/proxy"
)

func TestGetKubernetesVersion(t *testing.T) {
	var tests = []struct {
		description     string
		expectedVersion string
		paramVersion    string
		cfg             *cfg.ClusterConfig
	}{
		{
			description:     "kubernetes-version not given, no config",
			expectedVersion: constants.DefaultKubernetesVersion,
			paramVersion:    "",
		},
		{
			description:     "kubernetes-version not given, config available",
			expectedVersion: "v1.15.0",
			paramVersion:    "",
			cfg:             &cfg.ClusterConfig{KubernetesConfig: cfg.KubernetesConfig{KubernetesVersion: "v1.15.0"}},
		},
		{
			description:     "kubernetes-version given, no config",
			expectedVersion: "v1.15.0",
			paramVersion:    "v1.15.0",
		},
		{
			description:     "kubernetes-version given, config available",
			expectedVersion: "v1.16.0",
			paramVersion:    "v1.16.0",
			cfg:             &cfg.ClusterConfig{KubernetesConfig: cfg.KubernetesConfig{KubernetesVersion: "v1.15.0"}},
		},
		{
			description:     "kubernetes-version without patch version",
			expectedVersion: "v1.16.15",
			paramVersion:    "v1.16",
		},
		{
			description:     "kubernetes-version without patch version",
			expectedVersion: "v1.16.15",
			paramVersion:    "1.16",
		},
		{
			description:     "kubernetes-version given as 'stable', no config",
			expectedVersion: constants.DefaultKubernetesVersion,
			paramVersion:    "stable",
		},
		{
			description:     "kubernetes-version given as 'latest', no config",
			expectedVersion: constants.NewestKubernetesVersion,
			paramVersion:    "latest",
		},
		{
			description:     "kubernetes-version given as 'LATEST', no config",
			expectedVersion: constants.NewestKubernetesVersion,
			paramVersion:    "LATEST",
		},
		{
			description:     "kubernetes-version given as 'newest', no config",
			expectedVersion: constants.NewestKubernetesVersion,
			paramVersion:    "newest",
		},
		{
			description:     "kubernetes-version given as 'NEWEST', no config",
			expectedVersion: constants.NewestKubernetesVersion,
			paramVersion:    "NEWEST",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			viper.SetDefault(kubernetesVersion, test.paramVersion)
			version, err := getKubernetesVersion(test.cfg)
			if err != nil {
				klog.Warningf("failed getting Kubernetes version: %v", err)
			}

			// check whether we are getting the expected version
			if version != test.expectedVersion {
				t.Fatalf("test failed because the expected version %s is not returned", test.expectedVersion)
			}
		})
	}
}

var checkRepoMock = func(_ semver.Version, _ string) error { return nil }

func TestMirrorCountry(t *testing.T) {
	// Set default disk size value in lieu of flag init
	viper.SetDefault(humanReadableDiskSize, defaultDiskSize)
	checkRepository = checkRepoMock
	k8sVersion := constants.DefaultKubernetesVersion
	rtime := constants.DefaultContainerRuntime
	var tests = []struct {
		description     string
		k8sVersion      string
		imageRepository string
		mirrorCountry   string
		cfg             *cfg.ClusterConfig
	}{
		{
			description:     "repository-none_mirror-none",
			imageRepository: "",
			mirrorCountry:   "",
		},
		{
			description:     "repository-none_mirror-cn",
			imageRepository: "",
			mirrorCountry:   "cn",
		},
		{
			description:     "repository-auto_mirror-none",
			imageRepository: "auto",
			mirrorCountry:   "",
		},
		{
			description:     "repository-auto_mirror-cn",
			imageRepository: "auto",
			mirrorCountry:   "cn",
		},
		{
			description:     "repository-registry.test.com_mirror-none",
			imageRepository: "registry.test.com",
			mirrorCountry:   "",
		},
		{
			description:     "repository-registry.test.com_mirror-cn",
			imageRepository: "registry.test.com",
			mirrorCountry:   "cn",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cmd := &cobra.Command{}
			viper.SetDefault(imageRepository, test.imageRepository)
			viper.SetDefault(imageMirrorCountry, test.mirrorCountry)
			viper.SetDefault(kvmNUMACount, 1)
			config, _, err := generateClusterConfig(cmd, nil, k8sVersion, rtime, driver.Mock)
			if err != nil {
				t.Fatalf("Got unexpected error %v during config generation", err)
			}
			// the result can still be "", but anyway
			_ = config.KubernetesConfig.ImageRepository
		})
	}
}

func TestGenerateCfgFromFlagsHTTPProxyHandling(t *testing.T) {
	// Set default disk size value in lieu of flag init
	viper.SetDefault(humanReadableDiskSize, defaultDiskSize)

	k8sVersion := constants.NewestKubernetesVersion
	rtime := constants.DefaultContainerRuntime
	var tests = []struct {
		description  string
		proxy        string
		proxyIgnored bool
	}{

		{
			description:  "http_proxy=127.0.0.1:3128",
			proxy:        "127.0.0.1:3128",
			proxyIgnored: true,
		},
		{
			description:  "http_proxy=localhost:3128",
			proxy:        "localhost:3128",
			proxyIgnored: true,
		},
		{
			description:  "http_proxy=http://localhost:3128",
			proxy:        "http://localhost:3128",
			proxyIgnored: true,
		},
		{
			description:  "http_proxy=http://127.0.0.1:3128",
			proxy:        "http://127.0.0.1:3128",
			proxyIgnored: true,
		},
		{
			description: "http_proxy=http://1.2.127.0:3128",
			proxy:       "http://1.2.127.0:3128",
		},
		{
			description: "http_proxy=1.2.3.4:3128",
			proxy:       "1.2.3.4:3128",
		},
		{
			description: "no http_proxy",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cmd := &cobra.Command{}
			t.Setenv("HTTP_PROXY", test.proxy)

			cfg.DockerEnv = []string{} // clear docker env to avoid pollution
			proxy.SetDockerEnv()
			config, _, err := generateClusterConfig(cmd, nil, k8sVersion, rtime, "none")
			if err != nil {
				t.Fatalf("Got unexpected error %v during config generation", err)
			}
			envPrefix := "HTTP_PROXY="
			proxyEnv := envPrefix + test.proxy
			if test.proxy == "" {
				// If test.proxy is not set, ensure HTTP_PROXY is empty
				for _, v := range config.DockerEnv {
					if strings.HasPrefix(v, envPrefix) && len(v) > len(envPrefix) {
						t.Fatalf("HTTP_PROXY should be empty but got %s", v)
					}
				}
			} else {
				if test.proxyIgnored {
					// ignored proxy should not in config
					for _, v := range config.DockerEnv {
						if v == proxyEnv {
							t.Fatalf("Value %v not expected in dockerEnv but occurred", test.proxy)
						}
					}
				} else {
					// proxy must in config
					found := false
					for _, v := range config.DockerEnv {
						if v == proxyEnv {
							found = true
							break
						}
					}
					if !found {
						t.Fatalf("Value %s expected in dockerEnv but not occurred", test.proxy)
					}
				}
			}
		})
	}
}

func TestSuggestMemoryAllocation(t *testing.T) {
	var tests = []struct {
		description    string
		sysLimit       int
		containerLimit int
		nodes          int
		want           int
	}{
		{"128GB sys", 128000, 0, 1, 6000},
		{"64GB sys", 64000, 0, 1, 6000},
		{"32GB sys", 32768, 0, 1, 6000},
		{"16GB sys", 16384, 0, 1, 4000},
		{"odd sys", 14567, 0, 1, 3600},
		{"4GB sys", 4096, 0, 1, 2200},
		{"2GB sys", 2048, 0, 1, 2048},
		{"Unable to poll sys", 0, 0, 1, 2200},
		{"128GB sys, 16GB container", 128000, 16384, 1, 16336},
		{"64GB sys, 16GB container", 64000, 16384, 1, 16000},
		{"16GB sys, 4GB container", 16384, 4096, 1, 4000},
		{"4GB sys, 3.5GB container", 16384, 3500, 1, 3452},
		{"16GB sys, 2GB container", 16384, 2048, 1, 2048},
		{"16GB sys, unable to poll container", 16384, 0, 1, 4000},
		{"128GB sys 2 nodes", 128000, 0, 2, 6000},
		{"8GB sys 3 nodes", 8192, 0, 3, 2200},
		{"16GB sys 2 nodes", 16384, 0, 2, 2200},
		{"32GB sys 2 nodes", 32768, 0, 2, 4050},
		{"odd sys 2 nodes", 14567, 0, 2, 2200},
		{"4GB sys 2 nodes", 4096, 0, 2, 2200},
		{"2GB sys 3 nodes", 2048, 0, 3, 2048},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := suggestMemoryAllocation(test.sysLimit, test.containerLimit, test.nodes)
			if got != test.want {
				t.Errorf("defaultMemorySize(sys=%d, container=%d) = %d, want: %d", test.sysLimit, test.containerLimit, got, test.want)
			}
		})
	}
}

func TestBaseImageFlagDriverCombo(t *testing.T) {
	tests := []struct {
		driver        string
		canUseBaseImg bool
	}{
		{driver.Docker, true},
		{driver.Podman, true},
		{driver.None, false},
		{driver.KVM2, false},
		{driver.VirtualBox, false},
		{driver.HyperKit, false},
		{driver.VMware, false},
		{driver.HyperV, false},
		{driver.Parallels, false},
		{"something_invalid", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(test.driver, func(t *testing.T) {
			got := isBaseImageApplicable(test.driver)
			if got != test.canUseBaseImg {
				t.Errorf("isBaseImageApplicable(driver=%v): got %v, expected %v",
					test.driver, got, test.canUseBaseImg)
			}
		})
	}
}

func TestValidateImageRepository(t *testing.T) {
	var tests = []struct {
		imageRepository      string
		validImageRepository string
	}{
		{
			imageRepository:      "auto",
			validImageRepository: "auto",
		},
		{
			imageRepository:      "$$$$invalid",
			validImageRepository: "auto",
		},
		{
			imageRepository:      "",
			validImageRepository: "auto",
		},
		{
			imageRepository:      "http://registry.test.com/google_containers/",
			validImageRepository: "registry.test.com/google_containers",
		},
		{
			imageRepository:      "https://registry.test.com/google_containers/",
			validImageRepository: "registry.test.com/google_containers",
		},
		{
			imageRepository:      "registry.test.com/google_containers/",
			validImageRepository: "registry.test.com/google_containers",
		},
		{
			imageRepository:      "http://registry.test.com/google_containers",
			validImageRepository: "registry.test.com/google_containers",
		},
		{
			imageRepository:      "https://registry.test.com/google_containers",
			validImageRepository: "registry.test.com/google_containers",
		},
		{
			imageRepository:      "https://registry.test.com:6666/google_containers",
			validImageRepository: "registry.test.com:6666/google_containers",
		},
		{
			imageRepository:      "registry.test.com:6666/google_containers",
			validImageRepository: "registry.test.com:6666/google_containers",
		},
		{
			imageRepository:      "registry.1test.com:6666/google_containers",
			validImageRepository: "registry.1test.com:6666/google_containers",
		},
		{
			imageRepository:      "registry.t1est.com:6666/google_containers",
			validImageRepository: "registry.t1est.com:6666/google_containers",
		},
		{
			imageRepository:      "registry.test1.com:6666/google_containers",
			validImageRepository: "registry.test1.com:6666/google_containers",
		},
		{
			imageRepository:      "abc.xyz1.example.com",
			validImageRepository: "abc.xyz1.example.com",
		},
	}

	for _, test := range tests {
		t.Run(test.imageRepository, func(t *testing.T) {
			validImageRepository := validateImageRepository(test.imageRepository)
			if validImageRepository != test.validImageRepository {
				t.Errorf("validateImageRepository(imageRepo=%v): got %v, expected %v",
					test.imageRepository, validImageRepository, test.validImageRepository)
			}
		})
	}
}

func TestValidateDiskSize(t *testing.T) {
	var tests = []struct {
		diskSize string
		errorMsg string
	}{
		{
			diskSize: "2G",
			errorMsg: "",
		},
		{
			diskSize: "test",
			errorMsg: "Validation unable to parse disk size test: FromHumanSize: invalid size: 'test'",
		},
		{
			diskSize: "6M",
			errorMsg: fmt.Sprintf("Requested disk size 6 is less than minimum of %v", minimumDiskSize),
		},
	}
	for _, test := range tests {
		t.Run(test.diskSize, func(t *testing.T) {
			got := validateDiskSize(test.diskSize)
			gotError := ""
			if got != nil {
				gotError = got.Error()
			}
			if gotError != test.errorMsg {
				t.Errorf("validateDiskSize(diskSize=%v): got %v, expected %v", test.diskSize, got, test.errorMsg)
			}
		})
	}
}

func TestValidateRuntime(t *testing.T) {
	var tests = []struct {
		runtime  string
		errorMsg string
	}{
		{
			runtime:  "cri-o",
			errorMsg: "",
		},
		{
			runtime:  "docker",
			errorMsg: "",
		},
		{
			runtime:  "test",
			errorMsg: fmt.Sprintf("Invalid Container Runtime: test. Valid runtimes are: %v", cruntime.ValidRuntimes()),
		},
	}
	for _, test := range tests {
		t.Run(test.runtime, func(t *testing.T) {
			got := validateRuntime(test.runtime)
			gotError := ""
			if got != nil {
				gotError = got.Error()
			}
			if gotError != test.errorMsg {
				t.Errorf("ValidateRuntime(runtime=%v): got %v, expected %v", test.runtime, got, test.errorMsg)
			}
		})
	}
}

func TestIsTwoDigitSemver(t *testing.T) {
	var tcs = []struct {
		desc     string
		version  string
		expected bool
	}{
		{
			desc:     "a valid three digit version",
			version:  "1.26.5",
			expected: false,
		},
		{
			desc:     "a valid two digit version",
			version:  "1.26",
			expected: true,
		},
		{
			desc:     "a valid two digit version with a period",
			version:  "1.26.",
			expected: false,
		},
		{
			desc:     "an invalid major version",
			version:  "2",
			expected: false,
		},
		{
			desc:     "a valid major version",
			version:  "1",
			expected: false,
		},
		{
			desc:     "a two digit version with a 0 as the major/minor components",
			version:  "0.0",
			expected: true,
		},
		{
			desc:     "a two digit version with negative major version",
			version:  "-1.0",
			expected: false,
		},
		{
			desc:     "a two digit version with negative minor version",
			version:  "1.-1",
			expected: false,
		},
		{
			desc:     "a missing minor version",
			version:  "1.",
			expected: false,
		},
		{
			desc:     "a missing major version",
			version:  ".2",
			expected: false,
		},
		{
			desc:     "a valid two digit version with whitespace between components",
			version:  "1. 1",
			expected: false,
		},
		{
			desc:     "a two digit version with a non-digit major component",
			version:  "a.12",
			expected: false,
		},
		{
			desc:     "a two digit version with a non-digit minor component",
			version:  "1.a",
			expected: false,
		},
		{
			desc:     "a two digit version with extraneous non-digits in minor component",
			version:  "1.2a",
			expected: false,
		},
		{
			desc:     "a two digit version larger major/minor components",
			version:  "123456789.987654321",
			expected: true,
		},
		{
			desc:     "a valid two digit version with a version prefix",
			version:  "v1.26",
			expected: false,
		},
		{
			desc:     "a valid three digit version with a version prefix",
			version:  "v1.26.5",
			expected: false,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			actual := isTwoDigitSemver(tc.version)
			// check whether the function correctly verifies if it is a 2 digit semver
			if actual != tc.expected {
				t.Fatalf("test failed. Expected version %s to return %t", tc.version, tc.expected)
			}
		})
	}
}

func TestValidatePorts(t *testing.T) {
	type portTest struct {
		ports    []string
		errorMsg string
	}
	var tests = []portTest{
		{
			ports:    []string{"8080:80"},
			errorMsg: "",
		},
		{
			ports:    []string{"8080:80/tcp", "8080:80/udp"},
			errorMsg: "",
		},
		{
			ports:    []string{"test:8080"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is not valid [test:8080] (Invalid hostPort: test)",
		},
		{
			ports:    []string{"0:80"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is outside range: 0",
		},
		{
			ports:    []string{"0:80/tcp"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is outside range: 0",
		},
		{
			ports:    []string{"65536:80/udp"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is not valid [65536:80/udp] (Invalid hostPort: 65536)",
		},
		{
			ports:    []string{"0-1:80-81/tcp"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is outside range: 0",
		},
		{
			ports:    []string{"0-1:80-81/udp"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is outside range: 0",
		},
		{
			ports:    []string{"80:80", "1023-1025:8023-8025", "1023-1025:8023-8025/tcp", "1023-1025:8023-8025/udp"},
			errorMsg: "",
		},
		{
			ports:    []string{"127.0.0.1:8080:80", "127.0.0.1:8081:80/tcp", "127.0.0.1:8081:80/udp", "127.0.0.1:8082-8083:8082-8083/tcp"},
			errorMsg: "",
		},
		{
			ports:    []string{"1000.0.0.1:80:80"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is not valid [1000.0.0.1:80:80] (Invalid ip address: 1000.0.0.1)",
		},
		{
			ports:    []string{"127.0.0.1:80:80", "127.0.0.1:81:81/tcp", "127.0.0.1:81:81/udp", "127.0.0.1:82-83:82-83/tcp", "127.0.0.1:82-83:82-83/udp"},
			errorMsg: "",
		},
		{
			ports:    []string{"80"},
			errorMsg: "",
		},
		{
			ports:    []string{"80", "65535", "65536"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is outside range: 65536",
		},
		{
			ports:    []string{"0", "80", "65535"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is outside range: 0",
		},
		{
			ports:    []string{"cats"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is not valid: cats",
		},
		{
			ports:    []string{"127.0.0.1:81:0/tcp"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is outside range: 0",
		},
		{
			ports:    []string{"127.0.0.1:81:65536/tcp"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is not valid [127.0.0.1:81:65536/tcp] (Invalid containerPort: 65536)",
		},
		{
			ports:    []string{"1-65536:80-81/tcp"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is not valid [1-65536:80-81/tcp] (Invalid hostPort: 1-65536)",
		},
		{
			ports:    []string{"1-80:0-81/tcp"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is not valid [1-80:0-81/tcp] (Invalid ranges specified for container and host Ports: 0-81 and 1-80)",
		},
		{
			ports:    []string{"1-80:1-65536/tcp"},
			errorMsg: "Sorry, one of the ports provided with --ports flag is not valid [1-80:1-65536/tcp] (Invalid containerPort: 1-65536)",
		},
	}
	for _, test := range tests {
		t.Run(strings.Join(test.ports, ","), func(t *testing.T) {
			gotError := ""
			got := validatePorts(test.ports)
			if got != nil {
				gotError = got.Error()
			}
			if !strings.EqualFold(gotError, test.errorMsg) {
				t.Errorf("validatePorts(ports=%v): got %v, expected %v", test.ports, got, test.errorMsg)
			}
		})
	}
}

func TestValidateSubnet(t *testing.T) {
	type subnetTest struct {
		subnet   string
		errorMsg string
	}
	var tests = []subnetTest{
		{
			subnet:   "192.168.1.1",
			errorMsg: "",
		},
		{
			subnet:   "193.169.1.1",
			errorMsg: "Sorry, the subnet 193.169.1.1 is not a private IP",
		},
		{
			subnet:   "192.168.1.1/24",
			errorMsg: "",
		},
		{
			subnet:   "192.168.1.1/32",
			errorMsg: "Sorry, the subnet provided does not have a mask less than or equal to /30",
		},
	}
	for _, test := range tests {
		t.Run(test.subnet, func(t *testing.T) {
			gotError := ""
			got := validateSubnet(test.subnet)
			if got != nil {
				gotError = got.Error()
			}
			if gotError != test.errorMsg {
				t.Errorf("validateSubnet(subnet=%v): got %v, expected %v", test.subnet, got, test.errorMsg)
			}
		})
	}
}

func TestValidateStaticIP(t *testing.T) {
	tests := []struct {
		staticIP string
		drvName  string
		errorMsg string
	}{
		{
			staticIP: "8.8.8.8",
			drvName:  "docker",
			errorMsg: "static IP must be private",
		},
		{
			staticIP: "8.8.8.8",
			drvName:  "hyperkit",
			errorMsg: "",
		},
		{
			staticIP: "fdfc:a4c0:e99e:7ad3::",
			drvName:  "docker",
			errorMsg: "static IP must be IPv4",
		},
		{
			staticIP: "192.168.49.0",
			drvName:  "docker",
			errorMsg: "static IPs last octet must be between 2 and 254 (X.X.X.2 - X.X.X.254), for example 192.168.200.200",
		},
		{
			staticIP: "192.168.49.1",
			drvName:  "docker",
			errorMsg: "static IPs last octet must be between 2 and 254 (X.X.X.2 - X.X.X.254), for example 192.168.200.200",
		},
		{
			staticIP: "192.168.49.255",
			drvName:  "docker",
			errorMsg: "static IPs last octet must be between 2 and 254 (X.X.X.2 - X.X.X.254), for example 192.168.200.200",
		},
		{
			staticIP: "192.168.49.2",
			drvName:  "docker",
			errorMsg: "",
		},
		{
			staticIP: "192.168.49.254",
			drvName:  "docker",
			errorMsg: "",
		},
	}
	for _, tt := range tests {
		gotError := ""
		got := validateStaticIP(tt.staticIP, tt.drvName, "")
		if got != nil {
			gotError = got.Error()
		}
		if gotError != tt.errorMsg {
			t.Errorf("validateStaticIP(%s, %s): got %v, expected %v", tt.staticIP, tt.drvName, got, tt.errorMsg)
		}
	}
}

func TestImageMatchesBinaryVersion(t *testing.T) {
	tests := []struct {
		imageVersion  string
		binaryVersion string
		versionMatch  bool
	}{
		{"v1.17.0", "v1.17.0", true},
		{"v1.17.0", "v1.20.0", false},
		{"v1.31.0", "v1.31.1", true},
		{"v1.31.1", "v1.31.0", false},
	}

	for _, tc := range tests {
		got := imageMatchesBinaryVersion(tc.imageVersion, tc.binaryVersion)
		if got != tc.versionMatch {
			t.Errorf("imageMatchesBinaryVersion(%s, %s) = %t; want = %t", tc.imageVersion, tc.binaryVersion, got, tc.versionMatch)
		}
	}
}

func TestValidateGPUs(t *testing.T) {
	tests := []struct {
		gpus     string
		drvName  string
		runtime  string
		errorMsg string
	}{
		{"", "kvm", "containerd", ""},
		{"all", "docker", "docker", ""},
		{"nvidia", "docker", "docker", ""},
		{"all", "docker", "", ""},
		{"nvidia", "docker", "", ""},
		{"all", "kvm", "docker", "The gpus flag can only be used with the docker driver and docker container-runtime"},
		{"nvidia", "docker", "containerd", "The gpus flag can only be used with the docker driver and docker container-runtime"},
		{"cat", "docker", "docker", `The gpus flag must be passed a value of "nvidia", "amd" or "all"`},
		{"amd", "docker", "docker", ""},
		{"amd", "docker", "", ""},
		{"amd", "docker", "containerd", "The gpus flag can only be used with the docker driver and docker container-runtime"},
	}

	for _, tc := range tests {
		gotError := ""
		got := validateGPUs(tc.gpus, tc.drvName, tc.runtime)
		if got != nil {
			gotError = got.Error()
		}
		if gotError != tc.errorMsg {
			t.Errorf("validateGPUs(%s, %s, %s) = %q; want = %q", tc.gpus, tc.drvName, tc.runtime, got, tc.errorMsg)
		}
	}
}

func TestValidateAutoPause(t *testing.T) {
	tests := []struct {
		interval    string
		shouldError bool
	}{
		{"1m0s", false},
		{"5m", false},
		{"1s", false},
		{"0s", true},
		{"-2m", true},
	}
	for _, tc := range tests {
		input, err := time.ParseDuration(tc.interval)
		if err != nil {
			t.Fatalf("test has an invalid input duration of %q", tc.interval)
		}
		err = validateAutoPauseInterval(input)
		if err != nil && !tc.shouldError {
			t.Errorf("interval of %q failed validation; expected it to pass: %v", input, err)
		}
		if err == nil && tc.shouldError {
			t.Errorf("interval of %q passed validataion; expected it to fail: %v", input, err)
		}
	}
}
