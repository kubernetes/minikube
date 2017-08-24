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

package cluster

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/tests"
)

type MockDownloader struct{}

func (d MockDownloader) GetISOFileURI(isoURL string) string          { return "" }
func (d MockDownloader) CacheMinikubeISOFromURL(isoURL string) error { return nil }

var defaultMachineConfig = MachineConfig{
	VMDriver:    constants.DefaultVMDriver,
	MinikubeISO: constants.DefaultIsoUrl,
	Downloader:  MockDownloader{},
}

func TestCreateHost(t *testing.T) {
	api := tests.NewMockAPI()

	exists, _ := api.Exists(config.GetMachineName())
	if exists {
		t.Fatal("Machine already exists.")
	}
	_, err := createHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	exists, _ = api.Exists(config.GetMachineName())
	if !exists {
		t.Fatal("Machine does not exist, but should.")
	}

	h, err := api.Load(config.GetMachineName())
	if err != nil {
		t.Fatalf("Error loading machine: %v", err)
	}

	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine is not running. State is: %s", s)
	}

	found := false
	for _, driver := range constants.SupportedVMDrivers {
		if h.DriverName == driver {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Wrong driver name: %v. Should be virtualbox, vmwarefusion, kvm or xhyve.", h.DriverName)
	}
}

func TestStartCluster(t *testing.T) {
	expectedStartCmd, err := GetStartCommand(KubernetesConfig{})
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
			err := StartCluster(f, KubernetesConfig{})
			if err != nil && test.startCmd == expectedStartCmd {
				t.Errorf("Error starting cluster: %s", err)
			}
		})
	}
}

func TestUpdateCluster(t *testing.T) {
	defaultCfg := KubernetesConfig{
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
		k8s           KubernetesConfig
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
			k8s:         KubernetesConfig{},
			shouldErr:   true,
		},
	}

	for _, test := range cases {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			f := bootstrapper.NewFakeCommandRunner()
			err := UpdateCluster(f, test.k8s)
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
			actualStatus, err := GetLocalkubeStatus(f)
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
			_, err := GetHostLogs(f, test.follow)
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

func TestSetupCerts(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	f := bootstrapper.NewFakeCommandRunner()
	k8s := KubernetesConfig{
		APIServerName: constants.APIServerName,
		DNSDomain:     constants.ClusterDNSDomain,
	}

	var filesToBeTransferred []string
	for _, cert := range certs {
		filesToBeTransferred = append(filesToBeTransferred, filepath.Join(constants.GetMinipath(), cert))
	}

	if err := SetupCerts(f, k8s); err != nil {
		t.Fatalf("Error starting cluster: %s", err)
	}
	for _, cert := range filesToBeTransferred {
		_, err := f.GetFileToContents(cert)
		if err != nil {
			t.Errorf("Cert not generated: %s", cert)
		}
	}
}

func TestStartHostExists(t *testing.T) {
	api := tests.NewMockAPI()
	// Create an initial host.
	_, err := createHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}

	// Make sure the next call to Create will fail, to assert it doesn't get called again.
	api.CreateError = true
	if err := api.Create(&host.Host{}); err == nil {
		t.Fatal("api.Create did not fail, but should have.")
	}

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	// This should pass without calling Create because the host exists already.
	h, err := StartHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != config.GetMachineName() {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}
	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}
	if !md.Provisioner.Provisioned {
		t.Fatalf("Expected provision to be called")
	}
}

func TestStartStoppedHost(t *testing.T) {
	api := tests.NewMockAPI()
	// Create an initial host.
	h, err := createHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	d := tests.MockDriver{}
	h.Driver = &d
	d.CurrentState = state.Stopped

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)
	h, err = StartHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != config.GetMachineName() {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}

	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}

	if !api.SaveCalled {
		t.Fatalf("Machine must be saved after starting.")
	}

	if !md.Provisioner.Provisioned {
		t.Fatalf("Expected provision to be called")
	}
}

func TestStartHost(t *testing.T) {
	api := tests.NewMockAPI()

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	h, err := StartHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != config.GetMachineName() {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}
	if exists, _ := api.Exists(h.Name); !exists {
		t.Fatal("Machine not saved.")
	}
	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}

	// Provision regenerates Docker certs. This happens automatically during create,
	// so we should only call it again if the host already exists.
	if md.Provisioner.Provisioned {
		t.Fatalf("Did not expect Provision to be called")
	}
}

func TestStartHostConfig(t *testing.T) {
	api := tests.NewMockAPI()

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	config := MachineConfig{
		VMDriver:   constants.DefaultVMDriver,
		DockerEnv:  []string{"FOO=BAR"},
		DockerOpt:  []string{"param=value"},
		Downloader: MockDownloader{},
	}

	h, err := StartHost(api, config)
	if err != nil {
		t.Fatal("Error starting host.")
	}

	for i := range h.HostOptions.EngineOptions.Env {
		if h.HostOptions.EngineOptions.Env[i] != config.DockerEnv[i] {
			t.Fatal("Docker env variables were not set!")
		}
	}

	for i := range h.HostOptions.EngineOptions.ArbitraryFlags {
		if h.HostOptions.EngineOptions.ArbitraryFlags[i] != config.DockerOpt[i] {
			t.Fatal("Docker flags were not set!")
		}
	}

}

func TestStopHostError(t *testing.T) {
	api := tests.NewMockAPI()
	if err := StopHost(api); err == nil {
		t.Fatal("An error should be thrown when stopping non-existing machine.")
	}
}

func TestStopHost(t *testing.T) {
	api := tests.NewMockAPI()
	h, _ := createHost(api, defaultMachineConfig)
	if err := StopHost(api); err != nil {
		t.Fatal("An error should be thrown when stopping non-existing machine.")
	}
	if s, _ := h.Driver.GetState(); s != state.Stopped {
		t.Fatalf("Machine not stopped. Currently in state: %s", s)
	}
}

func TestDeleteHost(t *testing.T) {
	api := tests.NewMockAPI()
	createHost(api, defaultMachineConfig)

	if err := DeleteHost(api); err != nil {
		t.Fatalf("Unexpected error deleting host: %s", err)
	}
}

func TestDeleteHostErrorDeletingVM(t *testing.T) {
	api := tests.NewMockAPI()
	h, _ := createHost(api, defaultMachineConfig)

	d := &tests.MockDriver{RemoveError: true}

	h.Driver = d

	if err := DeleteHost(api); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostErrorDeletingFiles(t *testing.T) {
	api := tests.NewMockAPI()
	api.RemoveError = true
	createHost(api, defaultMachineConfig)

	if err := DeleteHost(api); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostMultipleErrors(t *testing.T) {
	api := tests.NewMockAPI()
	api.RemoveError = true
	h, _ := createHost(api, defaultMachineConfig)

	d := &tests.MockDriver{RemoveError: true}

	h.Driver = d

	err := DeleteHost(api)

	if err == nil {
		t.Fatal("Expected error deleting host, didn't get one.")
	}

	expectedErrors := []string{"Error removing " + config.GetMachineName(), "Error deleting machine"}
	for _, expectedError := range expectedErrors {
		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error %s expected to contain: %s.", err, expectedError)
		}
	}
}

func TestGetHostStatus(t *testing.T) {
	api := tests.NewMockAPI()

	checkState := func(expected string) {
		s, err := GetHostStatus(api)
		if err != nil {
			t.Fatalf("Unexpected error getting status: %s", err)
		}
		if s != expected {
			t.Fatalf("Expected status: %s, got %s", s, expected)
		}
	}

	checkState(state.None.String())

	createHost(api, defaultMachineConfig)
	checkState(state.Running.String())

	StopHost(api)
	checkState(state.Stopped.String())
}

func TestGetHostDockerEnv(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	api := tests.NewMockAPI()
	h, err := createHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{
			IPAddress: "127.0.0.1",
		},
	}
	h.Driver = d

	envMap, err := GetHostDockerEnv(api)
	if err != nil {
		t.Fatalf("Unexpected error getting env: %s", err)
	}

	dockerEnvKeys := [...]string{
		"DOCKER_TLS_VERIFY",
		"DOCKER_HOST",
		"DOCKER_CERT_PATH",
	}
	for _, dockerEnvKey := range dockerEnvKeys {
		if _, hasKey := envMap[dockerEnvKey]; !hasKey {
			t.Fatalf("Expected envMap[\"%s\"] key to be defined", dockerEnvKey)
		}
	}
}

func TestGetHostDockerEnvIPv6(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	api := tests.NewMockAPI()
	h, err := createHost(api, defaultMachineConfig)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{
			IPAddress: "fe80::215:5dff:fe00:a903",
		},
	}
	h.Driver = d

	envMap, err := GetHostDockerEnv(api)
	if err != nil {
		t.Fatalf("Unexpected error getting env: %s", err)
	}

	expected := "tcp://[fe80::215:5dff:fe00:a903]:2376"
	v := envMap["DOCKER_HOST"]
	if v != expected {
		t.Fatalf("Expected DOCKER_HOST to be defined as %s but was %s", expected, v)
	}
}

func TestCreateSSHShell(t *testing.T) {
	api := tests.NewMockAPI()

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port:         port,
		CurrentState: state.Running,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	api.Hosts[config.GetMachineName()] = &host.Host{Driver: d}

	cliArgs := []string{"exit"}
	if err := CreateSSHShell(api, cliArgs); err != nil {
		t.Fatalf("Error running ssh command: %s", err)
	}

	if !s.IsSessionRequested() {
		t.Fatalf("Expected ssh session to be run")
	}
}

func TestIsLocalkubeCached(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	inputArr := [...]string{
		"v1.3.3",
		"1.3.0",
		"http://test-url.localkube.com/localkube-binary",
		"file:///test/dir/to/localkube-binary",
	}

	localkubeCacher := localkubeCacher{
		k8sConf: KubernetesConfig{},
	}

	inner := func(input string) {
		localkubeCacher.k8sConf = KubernetesConfig{
			KubernetesVersion: input,
		}
		if localkubeCacher.isLocalkubeCached() {
			t.Errorf("IsLocalKubeCached returned true even though %s was not cached",
				localkubeCacher.getLocalkubeCacheFilepath())
			return
		}

		f, err := os.Create(localkubeCacher.getLocalkubeCacheFilepath())
		if err != nil {
			t.Errorf("failed to create dummy cache file: %v", err)
			return
		}
		f.Close()
		defer os.Remove(f.Name())
		if !localkubeCacher.isLocalkubeCached() {
			t.Errorf("IsLocalKubeCached returned false even though %s was cached",
				localkubeCacher.getLocalkubeCacheFilepath())
		}

	}
	for _, input := range inputArr {
		inner(input)
	}
}
