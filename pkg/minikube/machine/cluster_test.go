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
	"k8s.io/minikube/pkg/minikube/localpath"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/virtualbox"
	"k8s.io/minikube/pkg/minikube/run"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/tests"
)

func createMockDriverHost(cfg config.ClusterConfig, n config.Node) (interface{}, error) {
	return &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{
			MachineName: config.MachineName(cfg, n),
			StorePath:   localpath.MiniPath(),
			SSHUser:     "docker",
		},
	}, nil
}

func RegisterMockDriver(t *testing.T) {
	f := flag.Lookup("logtostderr")
	if f != nil {
		if err := f.Value.Set("true"); err != nil {
			t.Logf("unable to set logtostderr: %v", err)
		}
	}

	t.Helper()
	if !registry.Driver(driver.Mock).Empty() {
		return
	}
	err := registry.Register(registry.DriverDef{
		Name:   driver.Mock,
		Config: createMockDriverHost,
		Init:   func(_ *run.CommandOptions) drivers.Driver { return &tests.MockDriver{T: t} },
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
}

var defaultClusterConfig = config.ClusterConfig{
	Name:      constants.DefaultClusterName,
	Driver:    driver.Mock,
	DockerEnv: []string{"MOCK_MAKE_IT_PROVISION=true"},
	Nodes:     []config.Node{{Name: constants.DefaultClusterName}},
}

func TestCreateHost(t *testing.T) {
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	download.DownloadMock = download.CreateDstDownloadMock

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	exists, _ := api.Exists(options.ProfileName)
	if exists {
		t.Fatal("Machine already exists.")
	}

	_, err := createHost(api, &defaultClusterConfig, &config.Node{Name: options.ProfileName}, options)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	exists, err = api.Exists(options.ProfileName)
	if err != nil {
		t.Fatalf("exists failed for %q: %v", options.ProfileName, err)
	}
	if !exists {
		t.Fatalf("%q does not exist, but should.", options.ProfileName)
	}

	h, err := api.Load(options.ProfileName)
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
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	download.DownloadMock = download.CreateDstDownloadMock

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	// Create an initial host.
	ih, err := createHost(api, &defaultClusterConfig, &config.Node{Name: options.ProfileName}, options)
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

	// This should pass without calling Create because the host exists already.
	h, _, err := StartHost(api, &mc, &(mc.Nodes[0]), options)
	if err != nil {
		t.Fatalf("Error starting host: %v", err)
	}
	if h.Name != options.ProfileName {
		t.Fatalf("GetMachineName()=%q, want %q", options.ProfileName, h.Name)
	}
	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}
}

func TestStartHostErrMachineNotExist(t *testing.T) {
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	download.DownloadMock = download.CreateDstDownloadMock

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	// Create an incomplete host with machine does not exist error(i.e. User Interrupt Cancel)
	api.NotExistError = true
	h, err := createHost(api, &defaultClusterConfig, &config.Node{Name: "minikube"}, options)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	mc := defaultClusterConfig
	mc.Name = h.Name

	n := config.Node{Name: h.Name}

	// This should pass with creating host, while machine does not exist.
	h, _, err = StartHost(api, &mc, &n, options)
	if err != nil {
		if err != constants.ErrMachineMissing {
			t.Fatalf("Error starting host: %v", err)
		}
	}

	mc.Name = h.Name
	n.Name = h.Name

	n.Name = h.Name

	// Second call. This should pass without calling Create because the host exists already.
	h, _, err = StartHost(api, &mc, &n, options)
	if err != nil {
		t.Fatalf("Error starting host: %v", err)
	}

	if h.Name != options.ProfileName {
		t.Fatalf("GetMachineName()=%q, want %q", options.ProfileName, h.Name)
	}
	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine not started.")
	}
}

func TestStartStoppedHost(t *testing.T) {
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	download.DownloadMock = download.CreateDstDownloadMock

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	// Create an initial host.
	h, err := createHost(api, &defaultClusterConfig, &config.Node{Name: options.ProfileName}, options)
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
	h, _, err = StartHost(api, &mc, &n, options)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != options.ProfileName {
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
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	download.DownloadMock = download.CreateDstDownloadMock

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	h, _, err := StartHost(api, &defaultClusterConfig, &config.Node{Name: options.ProfileName}, options)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != options.ProfileName {
		t.Fatalf("GetMachineName()=%q, want %q", options.ProfileName, h.Name)
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
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	download.DownloadMock = download.CreateDstDownloadMock

	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	md := &tests.MockDetector{Provisioner: &tests.MockProvisioner{}}
	provision.SetDetector(md)

	cfg := config.ClusterConfig{
		Driver:    driver.Mock,
		DockerEnv: []string{"FOO=BAR"},
		DockerOpt: []string{"param=value"},
	}

	h, _, err := StartHost(api, &cfg, &config.Node{Name: options.ProfileName}, options)
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
	if err := StopHost(api, constants.DefaultClusterName); err == nil {
		t.Fatal("An error should be thrown when stopping non-existing machine.")
	}
}

func TestStopHost(t *testing.T) {
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	h, err := createHost(api, &defaultClusterConfig, &config.Node{Name: options.ProfileName}, options)
	if err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	cc := defaultClusterConfig
	cc.Name = options.ProfileName
	m := config.MachineName(cc, config.Node{Name: options.ProfileName})
	if err := StopHost(api, m); err != nil {
		t.Fatalf("Unexpected error stopping machine: %v", err)
	}
	if s, _ := h.Driver.GetState(); s != state.Stopped {
		t.Fatalf("Machine not stopped. Currently in state: %s", s)
	}
}

func TestDeleteHost(t *testing.T) {
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	if _, err := createHost(api, &defaultClusterConfig, &config.Node{Name: options.ProfileName}, options); err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	cc := defaultClusterConfig
	cc.Name = options.ProfileName

	if err := DeleteHost(api, config.MachineName(cc, config.Node{Name: options.ProfileName}), false); err != nil {
		t.Fatalf("Unexpected error deleting host: %v", err)
	}
}

func TestDeleteHostErrorDeletingVM(t *testing.T) {
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	h, err := createHost(api, &defaultClusterConfig, &config.Node{Name: options.ProfileName}, options)
	if err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	d := &tests.MockDriver{RemoveError: true, T: t}
	h.Driver = d

	if err := DeleteHost(api, config.MachineName(defaultClusterConfig, config.Node{Name: options.ProfileName}), false); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostErrorDeletingFiles(t *testing.T) {
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	api.RemoveError = true
	if _, err := createHost(api, &defaultClusterConfig, &config.Node{Name: options.ProfileName}, options); err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	if err := DeleteHost(api, config.MachineName(defaultClusterConfig, config.Node{Name: options.ProfileName}), false); err == nil {
		t.Fatal("Expected error deleting host.")
	}
}

func TestDeleteHostErrMachineNotExist(t *testing.T) {
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)
	// Create an incomplete host with machine does not exist error(i.e. User Interrupt Cancel)
	api.NotExistError = true
	_, err := createHost(api, &defaultClusterConfig, &config.Node{Name: options.ProfileName}, options)
	if err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	if err := DeleteHost(api, config.MachineName(defaultClusterConfig, config.Node{Name: options.ProfileName}), false); err != nil {
		t.Fatal("Deleting host with missing machine should not fail")
	}
}

func TestStatus(t *testing.T) {
	tests.MakeTempDir(t)

	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}
	RegisterMockDriver(t)
	api := tests.NewMockAPI(t)

	cc := defaultClusterConfig
	cc.Name = options.ProfileName

	m := config.MachineName(cc, config.Node{Name: options.ProfileName})

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

	if _, err := createHost(api, &cc, &config.Node{Name: options.ProfileName}, options); err != nil {
		t.Errorf("createHost failed: %v", err)
	}

	cc.Name = options.ProfileName

	m = config.MachineName(cc, config.Node{Name: options.ProfileName})

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
