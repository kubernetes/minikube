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

func TestGetKuberneterVersion(t *testing.T) {
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

func TestGenerateCfgFromFlagsHTTPProxyHandling(t *testing.T) {
	viper.SetDefault(memory, defaultMemorySize)
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
