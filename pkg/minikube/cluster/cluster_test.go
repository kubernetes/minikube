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
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/assets"
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
	api := tests.NewMockAPI()

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
		CurrentState: state.Running,
	}
	api.Hosts[config.GetMachineName()] = &host.Host{Driver: d}

	kubernetesConfig := KubernetesConfig{
		NodeIP: "",
	}

	err = StartCluster(api, kubernetesConfig)

	if err != nil {
		t.Fatalf("Error starting cluster: %s", err)
	}

	startCommand, err := GetStartCommand(kubernetesConfig)
	if err != nil {
		t.Fatalf("Error getting start command: %s", err)
	}
	for _, cmd := range []string{startCommand} {
		if _, ok := s.Commands[cmd]; !ok {
			t.Fatalf("Expected command not run: %s. Commands run: %v", cmd, s.Commands)
		}
	}
}

func TestStartClusterError(t *testing.T) {
	api := tests.NewMockAPI()

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
		CurrentState: state.Running,
		HostError:    true,
	}
	api.Hosts[config.GetMachineName()] = &host.Host{Driver: d}

	kubernetesConfig := KubernetesConfig{
		NodeIP: "192",
	}

	err = StartCluster(api, kubernetesConfig)

	if err == nil {
		t.Fatal("Error not thrown starting cluster.")
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

func TestGetLocalkubeStatus(t *testing.T) {
	api := tests.NewMockAPI()

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	api.Hosts[config.GetMachineName()] = &host.Host{Driver: d}

	s.SetCommandToOutput(map[string]string{
		localkubeStatusCommand: state.Running.String(),
	})
	if _, err := GetLocalkubeStatus(api); err != nil {
		t.Fatalf("Error getting localkube status: %s", err)
	}

	s.SetCommandToOutput(map[string]string{
		localkubeStatusCommand: state.Stopped.String(),
	})
	if _, err := GetLocalkubeStatus(api); err != nil {
		t.Fatalf("Error getting localkube status: %s", err)
	}

	s.SetCommandToOutput(map[string]string{
		localkubeStatusCommand: "Bad Output",
	})
	if _, err := GetLocalkubeStatus(api); err == nil {
		t.Fatalf("Expected error in getting localkube status as ssh returned bad output")
	}
}

func TestSetupCerts(t *testing.T) {
	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}

	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	if err := SetupCerts(d, constants.APIServerName, constants.ClusterDNSDomain); err != nil {
		t.Fatalf("Error starting cluster: %s", err)
	}

	for _, cert := range certs {
		contents, _ := ioutil.ReadFile(cert)
		transferred := s.Transfers.Bytes()
		if !bytes.Contains(transferred, contents) {
			t.Fatalf("Certificate not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
		}
	}
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

func TestHostGetLogs(t *testing.T) {
	api := tests.NewMockAPI()

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	api.Hosts[config.GetMachineName()] = &host.Host{Driver: d}

	tests := []struct {
		description string
		follow      bool
	}{
		{
			description: "logs",
			follow:      false,
		},
		{
			description: "logs -f",
			follow:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cmd, err := GetLogsCommand(test.follow)
			if err != nil {
				t.Errorf("Error getting the logs command: %s", err)
			}
			if _, err = GetHostLogs(api, test.follow); err != nil {
				t.Errorf("Error getting host logs: %s", err)
			}
			if _, ok := s.Commands[cmd]; !ok {
				t.Errorf("Expected command to run but did not: %s", cmd)
			}
		})
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

func TestUpdateDefault(t *testing.T) {
	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}

	kubernetesConfig := KubernetesConfig{
		KubernetesVersion: constants.DefaultKubernetesVersion,
	}

	if err := UpdateCluster(d, kubernetesConfig); err != nil {
		t.Fatalf("Error updating cluster: %s", err)
	}
	transferred := s.Transfers.Bytes()

	for _, addonBundle := range assets.Addons {
		if isEnabled, err := addonBundle.IsEnabled(); err == nil && isEnabled {
			for _, addon := range addonBundle.Assets {
				contents, _ := assets.Asset(addon.GetAssetName())
				if !bytes.Contains(transferred, contents) {
					t.Fatalf("File not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
				}
			}
		} else if err != nil {
			t.Fatalf("File not copied. Unexpected error while attempting to check transferred addons: %s", err)
		}
	}

	//test that localkube is transferred properly
	contents, _ := assets.Asset("out/localkube")
	if !bytes.Contains(transferred, contents) {
		t.Fatalf("File not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
	}
}

var testLocalkubeBin = "hello"

type K8sVersionHandlerCorrect struct{}

func (h *K8sVersionHandlerCorrect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, testLocalkubeBin)
}

func TestUpdateKubernetesVersion(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	handler := &K8sVersionHandlerCorrect{}
	server := httptest.NewServer(handler)

	kubernetesConfig := KubernetesConfig{
		KubernetesVersion: server.URL,
	}
	if err := UpdateCluster(d, kubernetesConfig); err != nil {
		t.Fatalf("Error updating cluster: %s", err)
	}
	transferred := s.Transfers.Bytes()

	//test that localkube is transferred properly
	contents := []byte(testLocalkubeBin)
	if !bytes.Contains(transferred, contents) {
		t.Fatalf("File not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
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

func TestUpdateCustomAddons(t *testing.T) {
	tempDir := tests.MakeTempDir()
	os.Mkdir(constants.MakeMiniPath("addons", "subdir"), 0777)
	defer os.RemoveAll(tempDir)

	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}

	//write a file into ~/.minikube/addons
	path := filepath.Join(constants.MakeMiniPath("addons"), "dir-addon.yaml")
	testContent1 := []byte("CUSTOM ADDON TEST STRING#1, In Addons Dir")
	err = ioutil.WriteFile(path, testContent1, 0644)
	if err != nil {
		t.Fatalf("Error writing custom addon: %s", err)
	}

	path = filepath.Join(constants.MakeMiniPath("addons", "subdir"), "subdir-addon.yaml")
	testContent2 := []byte("CUSTOM ADDON TEST STRING#2, In Addons SubDir")
	err = ioutil.WriteFile(path, testContent2, 0644)
	if err != nil {
		t.Fatalf("Error writing custom addon: %s", err)
	}

	//run update
	kubernetesConfig := KubernetesConfig{
		KubernetesVersion: constants.DefaultKubernetesVersion,
	}
	if err := UpdateCluster(d, kubernetesConfig); err != nil {
		t.Fatalf("Error updating cluster: %s", err)
	}
	transferred := s.Transfers.Bytes()

	//test that custom addons are transferred properly
	if !bytes.Contains(transferred, testContent1) {
		t.Fatalf("Custom addon not copied. Expected transfers to contain custom addon with content: %s. It was: %s", testContent1, transferred)
	}

	if !bytes.Contains(transferred, testContent2) {
		t.Fatalf("Custom addon not copied. Expected transfers to contain custom addon with content: %s. It was: %s", testContent2, transferred)
	}
}
