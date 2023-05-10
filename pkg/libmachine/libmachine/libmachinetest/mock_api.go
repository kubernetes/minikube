package libmachinetest

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/klog"
	"k8s.io/minikube/pkg/libmachine/drivers/mockdriver"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/host"
	"k8s.io/minikube/pkg/libmachine/libmachine/persist/persisttest"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
)

// MockAPI is a struct used to mock out libmachine.API
type MockAPI struct {
	persisttest.FakeStore
	CreateError   bool
	RemoveError   bool
	NotExistError bool
	SaveCalled    bool
	t             *testing.T
}

// NewMockAPI returns a new MockAPI
func NewMockAPI(t *testing.T) *MockAPI {
	t.Helper()
	m := MockAPI{
		FakeStore: persisttest.FakeStore{
			MiniHosts: make(map[string]*host.Host),
			T:         t,
		},
		t: t,
	}
	return &m
}

// Logf logs mock interactions
func (api *MockAPI) Logf(format string, args ...interface{}) {
	if api.t == nil {
		klog.Infof(format, args...)
		return
	}
	api.t.Logf(format, args...)
}

// Close closes the API.
func (api *MockAPI) Close() error {
	return nil
}

// x7TODO: fix this test -- provisionerName parameter added and not used
// NewHost creates a new host.Host instance.
func (api *MockAPI) NewHost(drvName, _ string, rawDriver []byte) (*host.Host, error) {
	var driver mockdriver.MockDriver
	if err := json.Unmarshal(rawDriver, &driver); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling json")
	}

	h := &host.Host{
		DriverName: drvName,
		RawDriver:  rawDriver,
		Driver:     &mockdriver.MockDriver{},
		Name:       fmt.Sprintf("mock-machine-%.8f", rand.Float64()),
		HostOptions: &host.Options{
			AuthOptions:  &auth.Options{},
			SwarmOptions: &swarm.Options{},
		},
	}

	api.Logf("MockAPI.NewHost: Setting profile=%q", h.Name)
	viper.Set("profile", h.Name)

	api.Logf("MockAPI.NewHost: %+v", h)
	return h, nil
}

// Load a created mock
func (api *MockAPI) Load(name string) (*host.Host, error) {
	h, err := api.FakeStore.Load(name)
	api.Logf("MockAPI.Load: %+v - %v", h, err)
	return h, err
}

// Create creates the actual host.
func (api *MockAPI) Create(h *host.Host) error {
	api.Logf("MockAPI.Create: %+v", h)
	if api.CreateError {
		return errors.New("error creating host")
	}
	// Propagate test info messages
	drv, ok := h.Driver.(*mockdriver.MockDriver)
	if ok {
		drv.T = api.t
	}
	if api.NotExistError {
		// initialize api.NotExistError
		api.NotExistError = false
		// reproduce ErrMachineNotExist
		drv.NotExistError = true
	}
	return h.Driver.CreateMachine()
}

// List the existing hosts.
func (api *MockAPI) List() ([]string, error) {
	return []string{}, nil
}

// Remove a host.
func (api *MockAPI) Remove(name string) error {
	api.Logf("MockAPI.Delete: %s", name)
	if api.RemoveError {
		return fmt.Errorf("error removing %s", name)
	}

	delete(api.MiniHosts, name)
	return nil
}

// Save saves a host to disk.
func (api *MockAPI) Save(host *host.Host) error {
	api.SaveCalled = true
	api.Logf("MockAPI.Save: %+v", host)
	return api.FakeStore.Save(host)
}

// GetMachinesDir returns the directory to store machines in.
func (api MockAPI) GetMachinesDir() string {
	return ""
}
