/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"reflect"
	"testing"

	"github.com/docker/machine/libmachine/host"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/tests"
)

type FakeShellDetector struct {
	Shell string
}

func (f FakeShellDetector) GetShell(_ string) (string, error) {
	return f.Shell, nil
}

type FakeNoProxyGetter struct {
	NoProxyVar   string
	NoProxyValue string
}

func (f FakeNoProxyGetter) GetNoProxyVar() (string, string) {
	return f.NoProxyVar, f.NoProxyValue
}

var defaultAPI = &tests.MockAPI{
	FakeStore: tests.FakeStore{
		Hosts: map[string]*host.Host{
			constants.DefaultMachineName: {
				Name:   constants.DefaultMachineName,
				Driver: &tests.MockDriver{},
			},
		},
	},
}

// Most of the shell cfg isn't configurable
func newShellCfg(shell, prefix, suffix, delim string) *ShellConfig {
	return &ShellConfig{
		DockerCertPath:  localpath.MakeMiniPath("certs"),
		DockerTLSVerify: "1",
		DockerHost:      "tcp://127.0.0.1:2376",
		UsageHint:       generateUsageHint(shell),
		Prefix:          prefix,
		Suffix:          suffix,
		Delimiter:       delim,
	}
}

func TestShellCfgSet(t *testing.T) {
	var tests = []struct {
		description      string
		api              *tests.MockAPI
		shell            string
		noProxyVar       string
		noProxyValue     string
		expectedShellCfg *ShellConfig
		shouldErr        bool
		noProxyFlag      bool
	}{
		{
			description: "no host specified",
			api: &tests.MockAPI{
				FakeStore: tests.FakeStore{
					Hosts: make(map[string]*host.Host),
				},
			},
			shell:            "bash",
			expectedShellCfg: nil,
			shouldErr:        true,
		},
		{
			description:      "default",
			api:              defaultAPI,
			shell:            "bash",
			expectedShellCfg: newShellCfg("", bashSetPfx, bashSetSfx, bashSetDelim),
			shouldErr:        false,
		},
		{
			description:      "bash",
			api:              defaultAPI,
			shell:            "bash",
			expectedShellCfg: newShellCfg("bash", bashSetPfx, bashSetSfx, bashSetDelim),
			shouldErr:        false,
		},
		{
			description:      "fish",
			api:              defaultAPI,
			shell:            "fish",
			expectedShellCfg: newShellCfg("fish", fishSetPfx, fishSetSfx, fishSetDelim),
			shouldErr:        false,
		},
		{
			description:      "powershell",
			api:              defaultAPI,
			shell:            "powershell",
			expectedShellCfg: newShellCfg("powershell", psSetPfx, psSetSfx, psSetDelim),
			shouldErr:        false,
		},
		{
			description:      "cmd",
			api:              defaultAPI,
			shell:            "cmd",
			expectedShellCfg: newShellCfg("cmd", cmdSetPfx, cmdSetSfx, cmdSetDelim),
			shouldErr:        false,
		},
		{
			description:      "emacs",
			api:              defaultAPI,
			shell:            "emacs",
			expectedShellCfg: newShellCfg("emacs", emacsSetPfx, emacsSetSfx, emacsSetDelim),
			shouldErr:        false,
		},
		{
			description:  "no proxy add uppercase",
			api:          defaultAPI,
			shell:        "bash",
			noProxyVar:   "NO_PROXY",
			noProxyValue: "",
			noProxyFlag:  true,
			expectedShellCfg: &ShellConfig{
				DockerCertPath:  localpath.MakeMiniPath("certs"),
				DockerTLSVerify: "1",
				DockerHost:      "tcp://127.0.0.1:2376",
				UsageHint:       usageHintMap["bash"],
				Prefix:          bashSetPfx,
				Suffix:          bashSetSfx,
				Delimiter:       bashSetDelim,
				NoProxyVar:      "NO_PROXY",
				NoProxyValue:    "127.0.0.1",
			},
		},
		{
			description:  "no proxy add lowercase",
			api:          defaultAPI,
			shell:        "bash",
			noProxyVar:   "no_proxy",
			noProxyValue: "",
			noProxyFlag:  true,
			expectedShellCfg: &ShellConfig{
				DockerCertPath:  localpath.MakeMiniPath("certs"),
				DockerTLSVerify: "1",
				DockerHost:      "tcp://127.0.0.1:2376",
				UsageHint:       usageHintMap["bash"],
				Prefix:          bashSetPfx,
				Suffix:          bashSetSfx,
				Delimiter:       bashSetDelim,
				NoProxyVar:      "no_proxy",
				NoProxyValue:    "127.0.0.1",
			},
		},
		{
			description:  "no proxy idempotent",
			api:          defaultAPI,
			shell:        "bash",
			noProxyVar:   "no_proxy",
			noProxyValue: "127.0.0.1",
			noProxyFlag:  true,
			expectedShellCfg: &ShellConfig{
				DockerCertPath:  localpath.MakeMiniPath("certs"),
				DockerTLSVerify: "1",
				DockerHost:      "tcp://127.0.0.1:2376",
				UsageHint:       usageHintMap["bash"],
				Prefix:          bashSetPfx,
				Suffix:          bashSetSfx,
				Delimiter:       bashSetDelim,
				NoProxyVar:      "no_proxy",
				NoProxyValue:    "127.0.0.1",
			},
		},
		{
			description:  "no proxy list add",
			api:          defaultAPI,
			shell:        "bash",
			noProxyVar:   "no_proxy",
			noProxyValue: "0.0.0.0",
			noProxyFlag:  true,
			expectedShellCfg: &ShellConfig{
				DockerCertPath:  localpath.MakeMiniPath("certs"),
				DockerTLSVerify: "1",
				DockerHost:      "tcp://127.0.0.1:2376",
				UsageHint:       usageHintMap["bash"],
				Prefix:          bashSetPfx,
				Suffix:          bashSetSfx,
				Delimiter:       bashSetDelim,
				NoProxyVar:      "no_proxy",
				NoProxyValue:    "0.0.0.0,127.0.0.1",
			},
		},
		{
			description:  "no proxy list already present",
			api:          defaultAPI,
			shell:        "bash",
			noProxyVar:   "no_proxy",
			noProxyValue: "0.0.0.0,127.0.0.1",
			noProxyFlag:  true,
			expectedShellCfg: &ShellConfig{
				DockerCertPath:  localpath.MakeMiniPath("certs"),
				DockerTLSVerify: "1",
				DockerHost:      "tcp://127.0.0.1:2376",
				UsageHint:       usageHintMap["bash"],
				Prefix:          bashSetPfx,
				Suffix:          bashSetSfx,
				Delimiter:       bashSetDelim,
				NoProxyVar:      "no_proxy",
				NoProxyValue:    "0.0.0.0,127.0.0.1",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {

			viper.Set(config.MachineProfile, constants.DefaultMachineName)
			defaultShellDetector = &FakeShellDetector{test.shell}
			defaultNoProxyGetter = &FakeNoProxyGetter{test.noProxyVar, test.noProxyValue}
			noProxy = test.noProxyFlag
			test.api.T = t
			shellCfg, err := shellCfgSet(test.api)
			if !reflect.DeepEqual(shellCfg, test.expectedShellCfg) {
				t.Errorf("Shell cfgs differ: expected %+v, \n\n got %+v", test.expectedShellCfg, shellCfg)
			}
			if err != nil && !test.shouldErr {
				t.Errorf("Unexpected error occurred: %s, error: %v", test.description, err)
			}
			if err == nil && test.shouldErr {
				t.Errorf("Test didn't return error but should have: %s", test.description)
			}
		})
	}
}

func TestShellCfgUnset(t *testing.T) {
	var tests = []struct {
		description      string
		shell            string
		expectedShellCfg *ShellConfig
	}{
		{
			description: "unset default",
			shell:       "bash",
			expectedShellCfg: &ShellConfig{
				Prefix:    bashUnsetPfx,
				Suffix:    bashUnsetSfx,
				Delimiter: bashUnsetDelim,
				UsageHint: usageHintMap["bash"],
			},
		},
		{
			description: "unset bash",
			shell:       "bash",
			expectedShellCfg: &ShellConfig{
				Prefix:    bashUnsetPfx,
				Suffix:    bashUnsetSfx,
				Delimiter: bashUnsetDelim,
				UsageHint: usageHintMap["bash"],
			},
		},
		{
			description: "unset fish",
			shell:       "fish",
			expectedShellCfg: &ShellConfig{
				Prefix:    fishUnsetPfx,
				Suffix:    fishUnsetSfx,
				Delimiter: fishUnsetDelim,
				UsageHint: usageHintMap["fish"],
			},
		},
		{
			description: "unset powershell",
			shell:       "powershell",
			expectedShellCfg: &ShellConfig{
				Prefix:    psUnsetPfx,
				Suffix:    psUnsetSfx,
				Delimiter: psUnsetDelim,
				UsageHint: usageHintMap["powershell"],
			},
		},
		{
			description: "unset cmd",
			shell:       "cmd",
			expectedShellCfg: &ShellConfig{
				Prefix:    cmdUnsetPfx,
				Suffix:    cmdUnsetSfx,
				Delimiter: cmdUnsetDelim,
				UsageHint: usageHintMap["cmd"],
			},
		},
		{
			description: "unset emacs",
			shell:       "emacs",
			expectedShellCfg: &ShellConfig{
				Prefix:    emacsUnsetPfx,
				Suffix:    emacsUnsetSfx,
				Delimiter: emacsUnsetDelim,
				UsageHint: usageHintMap["emacs"],
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			defaultShellDetector = &FakeShellDetector{test.shell}
			defaultNoProxyGetter = &FakeNoProxyGetter{}
			actual, _ := shellCfgUnset()
			if !reflect.DeepEqual(actual, test.expectedShellCfg) {
				t.Errorf("Actual shell config did not match expected: \n\n actual: \n%+v \n\n expected: \n%+v \n\n", actual, test.expectedShellCfg)
			}
		})
	}
}
