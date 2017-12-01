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

package localkube

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/constants"
)

func TestStartCluster(t *testing.T) {
	expectedStartCmd, err := GetStartCommand(bootstrapper.KubernetesConfig{})
	if err != nil {
		t.Fatalf("generating start command: %s", err)
	}

	cases := []struct {
		description string
		startCmd    string
	}{
		{
			description: "start cluster success",
			startCmd:    expectedStartCmd,
		},
		{
			description: "start cluster failure",
			startCmd:    "something else",
		},
	}

	for _, test := range cases {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			f := bootstrapper.NewFakeCommandRunner()
			f.SetCommandToOutput(map[string]string{test.startCmd: "ok"})
			l := LocalkubeBootstrapper{f}
			err := l.StartCluster(bootstrapper.KubernetesConfig{})
			if err != nil && test.startCmd == expectedStartCmd {
				t.Errorf("Error starting cluster: %s", err)
			}
		})
	}
}

func TestUpdateCluster(t *testing.T) {
	defaultCfg := bootstrapper.KubernetesConfig{
		KubernetesVersion: constants.DefaultKubernetesVersion,
	}
	defaultAddons := []string{
		"deploy/addons/kube-dns/kube-dns-cm.yaml",
		"deploy/addons/kube-dns/kube-dns-svc.yaml",
		"deploy/addons/addon-manager.yaml",
		"deploy/addons/dashboard/dashboard-rc.yaml",
		"deploy/addons/dashboard/dashboard-svc.yaml",
		"deploy/addons/storageclass/storageclass.yaml",
		"deploy/addons/kube-dns/kube-dns-controller.yaml",
	}
	cases := []struct {
		description   string
		k8s           bootstrapper.KubernetesConfig
		expectedFiles []string
		shouldErr     bool
	}{
		{
			description:   "transfer localkube correct",
			k8s:           defaultCfg,
			expectedFiles: []string{"out/localkube"},
		},
		{
			description:   "addons are transferred",
			k8s:           defaultCfg,
			expectedFiles: defaultAddons,
		},
		{
			description: "no localkube version",
			k8s:         bootstrapper.KubernetesConfig{},
			shouldErr:   true,
		},
	}

	for _, test := range cases {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			f := bootstrapper.NewFakeCommandRunner()
			l := LocalkubeBootstrapper{f}
			err := l.UpdateCluster(test.k8s)
			if err != nil && !test.shouldErr {
				t.Errorf("Error updating cluster: %s", err)
				return
			}
			if err == nil && test.shouldErr {
				t.Error("Didn't get error, but expected to")
				return
			}
			for _, expectedFile := range test.expectedFiles {
				_, err := f.GetFileToContents(expectedFile)
				if err != nil {
					t.Errorf("Expected file %s, but was not present", expectedFile)
				}
			}
		})
	}
}

func TestGetLocalkubeStatus(t *testing.T) {
	cases := []struct {
		description    string
		statusCmdMap   map[string]string
		expectedStatus string
		shouldErr      bool
	}{
		{
			description:    "get status running",
			statusCmdMap:   map[string]string{localkubeStatusCommand: "Running"},
			expectedStatus: "Running",
		},
		{
			description:    "get status stopped",
			statusCmdMap:   map[string]string{localkubeStatusCommand: "Stopped"},
			expectedStatus: "Stopped",
		},
		{
			description:  "get status unknown status",
			statusCmdMap: map[string]string{localkubeStatusCommand: "Recalculating..."},
			shouldErr:    true,
		},
		{
			description:  "get status error",
			statusCmdMap: map[string]string{"a": "b"},
			shouldErr:    true,
		},
	}

	for _, test := range cases {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			f := bootstrapper.NewFakeCommandRunner()
			f.SetCommandToOutput(test.statusCmdMap)
			l := LocalkubeBootstrapper{f}
			actualStatus, err := l.GetClusterStatus()
			if err != nil && !test.shouldErr {
				t.Errorf("Error getting localkube status: %s", err)
				return
			}
			if err == nil && test.shouldErr {
				t.Error("Didn't get error, but expected to")
				return
			}
			if test.expectedStatus != actualStatus {
				t.Errorf("Expected status: %s, Actual status: %s", test.expectedStatus, actualStatus)
			}
		})
	}
}

func TestGetHostLogs(t *testing.T) {
	logs, err := GetLogsCommand(false)
	if err != nil {
		t.Fatalf("Error getting logs command: %s", err)
	}
	logsf, err := GetLogsCommand(true)
	if err != nil {
		t.Fatalf("Error gettings logs -f command: %s", err)
	}

	cases := []struct {
		description string
		logsCmdMap  map[string]string
		follow      bool
		shouldErr   bool
	}{
		{
			description: "get logs correct",
			logsCmdMap:  map[string]string{logs: "fee"},
		},
		{
			description: "follow logs correct",
			logsCmdMap:  map[string]string{logsf: "fi"},
			follow:      true,
		},
		{
			description: "get logs incorrect",
			logsCmdMap:  map[string]string{"fo": "fum"},
			shouldErr:   true,
		},
	}

	for _, test := range cases {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			f := bootstrapper.NewFakeCommandRunner()
			f.SetCommandToOutput(test.logsCmdMap)
			l := LocalkubeBootstrapper{f}
			_, err := l.GetClusterLogs(test.follow)
			if err != nil && !test.shouldErr {
				t.Errorf("Error getting localkube logs: %s", err)
				return
			}
			if err == nil && test.shouldErr {
				t.Error("Didn't get error, but expected to")
				return
			}
		})
	}
}
