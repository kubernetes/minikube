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

package machine

import (
	"flag"
	"fmt"
	"testing"
	"time"

	// Driver used by testdata
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/virtualbox"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/tests"
)

func createMockDriverHost(c config.ClusterConfig, n config.Node) (interface{}, error) {
	return nil, nil
}

func RegisterMockDriver(t *testing.T) {
	// Debugging this test is a nightmare.
	if err := flag.Lookup("logtostderr").Value.Set("true"); err != nil {
		t.Logf("unable to set logtostderr: %v", err)
	}

	t.Helper()
	if !registry.Driver(driver.Mock).Empty() {
		return
	}
	err := registry.Register(registry.DriverDef{
		Name:   driver.Mock,
		Config: createMockDriverHost,
		Init:   func() drivers.Driver { return &tests.MockDriver{T: t} },
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
}

var defaultClusterConfig = config.ClusterConfig{
	Name:      viper.GetString("profile"),
	Driver:    driver.Mock,
	DockerEnv: []string{"MOCK_MAKE_IT_PROVISION=true"},
	Nodes:     []config.Node{{Name: "minikube"}},
}

func TestCreateHost(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	download.EnableMock(true)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	exists, _ := api.Exists(viper.GetString("profile"))
	if exists {
		t.Fatal("Machine already exists.")
	}

	_, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"})
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	exists, err = api.Exists(viper.GetString("profile"))
	if err != nil {
		t.Fatalf("exists failed for %q: %v", viper.GetString("profile"), err)
	}
	if !exists {
		t.Fatalf("%q does not exist, but should.", viper.GetString("profile"))
	}

	h, err := api.Load(viper.GetString("profile"))
	if err != nil {
		t.Fatalf("Error loading machine: %v", err)
	}

	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine is not running. State is: %s", s)
	}

	found := false
	for _, def := range registry.List() {
		if h.DriverName == def.Name {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Wrong driver name: %v. It should be among drivers %v", h.DriverName, registry.List())
	}
}

func TestStartHostExists(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	download.EnableMock(true)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	// Create an initial host.
	ih, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"})
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

	mc := defaultClusterConfig
	mc.Name = ih.Name

	n := config.Node{Name: ih.Name}

	// This should pass without calling Create because the host exists already.
	h, _, err := StartHost(api, &mc, &n)
	if err != nil {
		t.Fatalf("Error starting host: %v", err)
	}
	if h.Name != viper.GetString("profile") {
		t.Fatalf("GetMachineName()=%q, want %q", viper.GetString("profile"), h.Name)
	}
	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}
}

func TestStartHostErrMachineNotExist(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	download.EnableMock(true)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	// Create an incomplete host with machine does not exist error(i.e. User Interrupt Cancel)
	api.NotExistError = true
	h, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"})
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	mc := defaultClusterConfig
	mc.Name = h.Name

	n := config.Node{Name: h.Name}

	// This should pass with creating host, while machine does not exist.
	h, _, err = StartHost(api, &mc, &n)
	if err != nil {
		if err != constants.ErrMachineMissing {
			t.Fatalf("Error starting host: %v", err)
		}
	}

	mc.Name = h.Name
	n.Name = h.Name

	n.Name = h.Name

	// Second call. This should pass without calling Create because the host exists already.
	h, _, err = StartHost(api, &mc, &n)
	if err != nil {
		t.Fatalf("Error starting host: %v", err)
	}

	if h.Name != viper.GetString("profile") {
		t.Fatalf("GetMachineName()=%q, want %q", viper.GetString("profile"), h.Name)
	}
	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}
}

func TestStartStoppedHost(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	download.EnableMock(true)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	// Create an initial host.
	h, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"})
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	d := tests.MockDriver{T: t}
	h.Driver = &d
	d.CurrentState = state.Stopped

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)
	mc := defaultClusterConfig
	mc.Name = h.Name
	n := config.Node{Name: h.Name}
	h, _, err = StartHost(api, &mc, &n)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != viper.GetString("profile") {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}

	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}

	if !api.SaveCalled {
		t.Fatalf("Machine must be saved after starting.")
	}

}

func TestStartHost(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	download.EnableMock(true)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	h, _, err := StartHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"})
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != viper.GetString("profile") {
		t.Fatalf("GetMachineName()=%q, want %q", viper.GetString("profile"), h.Name)
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
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	download.EnableMock(true)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	cfg := config.ClusterConfig{
		Driver:    driver.Mock,
		DockerEnv: []string{"FOO=BAR"},
		DockerOpt: []string{"param=value"},
	}

	h, _, err := StartHost(api, &cfg, &config.Node{Name: "minikube"})
	if err != nil {
		t.Fatal("Error starting host.")
	}

	for i := range h.HostOptions.EngineOptions.Env {
		if h.HostOptions.EngineOptions.Env[i] != cfg.DockerEnv[i] {
			t.Fatalf("Docker env variables were not set! got %+v but want %+v", h.HostOptions.EngineOptions.Env, cfg.DockerEnv)
		}
	}

	for i := range h.HostOptions.EngineOptions.ArbitraryFlags {
		if h.HostOptions.EngineOptions.ArbitraryFlags[i] != cfg.DockerOpt[i] {
			t.Fatal("Docker flags were not set!")
		}
	}

}

func TestStopHostError(t *testing.T) {
	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	if err := StopHost(api, viper.GetString("profile")); err == nil {
		t.Fatal("An error should be thrown when stopping non-existing machine.")
	}
}

func TestStopHost(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	h, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"})
	if err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	cc := defaultClusterConfig
	cc.Name = viper.GetString("profile")
	m := driver.MachineName(cc, config.Node{Name: "minikube"})
	if err := StopHost(api, m); err != nil {
		t.Fatalf("Unexpected error stopping machine: %v", err)
	}
	if s, _ := h.Driver.GetState(); s != state.Stopped {
		t.Fatalf("Machine not stopped. Currently in state: %s", s)
	}
}

func TestDeleteHost(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	if _, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"}); err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	cc := defaultClusterConfig
	cc.Name = viper.GetString("profile")

	if err := DeleteHost(api, driver.MachineName(cc, config.Node{Name: "minikube"}), false); err != nil {
		t.Fatalf("Unexpected error deleting host: %v", err)
	}
}

func TestDeleteHostErrorDeletingVM(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	h, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"})
	if err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	d := &tests.MockDriver{RemoveError: true, T: t}
	h.Driver = d

	if err := DeleteHost(api, driver.MachineName(defaultClusterConfig, config.Node{Name: "minikube"}), false); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostErrorDeletingFiles(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	api.RemoveError = true
	if _, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"}); err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	if err := DeleteHost(api, driver.MachineName(defaultClusterConfig, config.Node{Name: "minikube"}), false); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostErrMachineNotExist(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	// Create an incomplete host with machine does not exist error(i.e. User Interrupt Cancel)
	api.NotExistError = true
	_, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"})
	if err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	if err := DeleteHost(api, driver.MachineName(defaultClusterConfig, config.Node{Name: "minikube"}), false); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestStatus(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	cc := defaultClusterConfig
	cc.Name = viper.GetString("profile")

	m := driver.MachineName(cc, config.Node{Name: "minikube"})

	checkState := func(expected string, machineName string) {
		s, err := Status(api, machineName)
		if err != nil {
			t.Fatalf("Unexpected error getting status: %v", err)
		}
		if s != expected {
			t.Fatalf("Expected status: %s, got: %s", expected, s)
		}
	}

	checkState(state.None.String(), m)

	if _, err := createHost(api, &cc, &config.Node{Name: "minikube"}); err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	cc.Name = viper.GetString("profile")

	m = driver.MachineName(cc, config.Node{Name: "minikube"})

	checkState(state.Running.String(), m)

	if err := StopHost(api, m); err != nil {
		t.Errorf("StopHost failed: %v", err)
	}
	checkState(state.Stopped.String(), m)
}

func TestGuestClockDelta(t *testing.T) {
	local := time.Now()
	h := tests.NewMockHost()
	// Truncate remote clock so that it is between 0 and 1 second behind
	h.CommandOutput["date +%s.%N"] = fmt.Sprintf("%d.0000", local.Unix())
	got, err := guestClockDelta(h, local)
	if err != nil {
		t.Fatalf("guestClock: %v", err)
	}
	if got > (0 * time.Second) {
		t.Errorf("unexpected positive delta (remote should be behind): %s", got)
	}
	if got < (-1 * time.Second) {
		t.Errorf("unexpectedly negative delta (remote too far behind): %s", got)
	}
}
