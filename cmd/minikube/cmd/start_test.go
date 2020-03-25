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
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
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
			description:     "image-repository auto, image-mirror-country none",
			imageRepository: "auto",
			mirrorCountry:   "",
		},
		{
			description:     "image-repository auto, image-mirror-country china",
			imageRepository: "auto",
			mirrorCountry:   "cn",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cmd := &cobra.Command{}
			viper.SetDefault(imageRepository, test.imageRepository)
			viper.SetDefault(imageMirrorCountry, test.mirrorCountry)
			config, _, err := generateCfgFromFlags(cmd, k8sVersion, "none")
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
			config, _, err := generateCfgFromFlags(cmd, k8sVersion, "none")
			if err != nil {
				t.Fatalf("Got unexpected error %v during config generation", err)
			}
			// ignored proxy should not be in config
			for _, v := range config.DockerEnv {
				if v == test.proxy && test.proxyIgnored {
					t.Fatalf("Value %v not expected in dockerEnv but occurred", v)
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
		want           int
	}{
		{"128GB sys", 128000, 0, 6000},
		{"64GB sys", 64000, 0, 6000},
		{"16GB sys", 16384, 0, 4000},
		{"odd sys", 14567, 0, 3600},
		{"4GB sys", 4096, 0, 2200},
		{"2GB sys", 2048, 0, 2048},
		{"Unable to poll sys", 0, 0, 2200},
		{"128GB sys, 16GB container", 128000, 16384, 16336},
		{"64GB sys, 16GB container", 64000, 16384, 16000},
		{"16GB sys, 4GB container", 16384, 4096, 4000},
		{"4GB sys, 3.5GB container", 16384, 3500, 3452},
		{"2GB sys, 2GB container", 16384, 2048, 2048},
		{"2GB sys, unable to poll container", 16384, 0, 4000},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := suggestMemoryAllocation(test.sysLimit, test.containerLimit)
			if got != test.want {
				t.Errorf("defaultMemorySize(sys=%d, container=%d) = %d, want: %d", test.sysLimit, test.containerLimit, got, test.want)
			}
		})
	}
}
