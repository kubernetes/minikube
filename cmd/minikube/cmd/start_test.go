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
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
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
			description:     "kubernetes-version given as 'stable', no config",
			expectedVersion: constants.DefaultKubernetesVersion,
			paramVersion:    "stable",
		},
		{
			description:     "kubernetes-version given as 'latest', no config",
			expectedVersion: constants.NewestKubernetesVersion,
			paramVersion:    "latest",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			viper.SetDefault(kubernetesVersion, test.paramVersion)
			version := getKubernetesVersion(test.cfg)

			// check whether we are getting the expected version
			if version != test.expectedVersion {
				t.Fatalf("test failed because the expected version %s is not returned", test.expectedVersion)
			}
		})
	}
}

func TestMirrorCountry(t *testing.T) {
	// Set default disk size value in lieu of flag init
	viper.SetDefault(humanReadableDiskSize, defaultDiskSize)

	k8sVersion := constants.DefaultKubernetesVersion
	var tests = []struct {
		description     string
		k8sVersion      string
		imageRepository string
		mirrorCountry   string
		cfg             *cfg.ClusterConfig
	}{
		{
			description:     "image-repository none, image-mirror-country none",
			imageRepository: "",
			mirrorCountry:   "",
		},
		{
			description:     "image-repository none, image-mirror-country china",
			imageRepository: "",
			mirrorCountry:   "cn",
		},
		{
			description:     "image-repository auto, image-mirror-country none",
			imageRepository: "auto",
			mirrorCountry:   "",
		},
		{
			description:     "image-repository auto, image-mirror-country china",
			imageRepository: "auto",
			mirrorCountry:   "cn",
		},
		{
			description:     "image-repository registry.test.com, image-mirror-country none",
			imageRepository: "registry.test.com",
			mirrorCountry:   "",
		},
		{
			description:     "image-repository registry.test.com, image-mirror-country china",
			imageRepository: "registry.test.com",
			mirrorCountry:   "cn",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cmd := &cobra.Command{}
			viper.SetDefault(imageRepository, test.imageRepository)
			viper.SetDefault(imageMirrorCountry, test.mirrorCountry)
			config, _, err := generateClusterConfig(cmd, nil, k8sVersion, "none")
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

	originalEnv := os.Getenv("HTTP_PROXY")
	defer func() {
		err := os.Setenv("HTTP_PROXY", originalEnv)
		if err != nil {
			t.Fatalf("Error reverting env HTTP_PROXY to it's original value. Got err: %s", err)
		}
	}()
	k8sVersion := constants.NewestKubernetesVersion
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
			if err := os.Setenv("HTTP_PROXY", test.proxy); err != nil {
				t.Fatalf("Unexpected error setting HTTP_PROXY: %v", err)
			}

			cfg.DockerEnv = []string{} // clear docker env to avoid pollution
			proxy.SetDockerEnv()
			config, _, err := generateClusterConfig(cmd, nil, k8sVersion, "none")
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
		{driver.VMwareFusion, false},
		{driver.HyperV, false},
		{driver.Parallels, false},
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
