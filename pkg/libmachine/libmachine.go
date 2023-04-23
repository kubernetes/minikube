package libmachine

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/juju/fslock"
	"github.com/pkg/errors"

	"k8s.io/klog"
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cert"
	"k8s.io/minikube/pkg/libmachine/cruntime"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/machine"
	"k8s.io/minikube/pkg/libmachine/persist"
	"k8s.io/minikube/pkg/libmachine/registry"
	"k8s.io/minikube/pkg/libmachine/ssh"
	"k8s.io/minikube/pkg/libmachine/state"
	"k8s.io/minikube/pkg/libmachine/version"
	"k8s.io/minikube/pkg/minikube/driver"
)

type API interface {
	io.Closer
	// Close() error
	persist.Store
	// Exists(name string) (bool, error)
	// List() ([]string, error)
	// Load(name string) (*machine.Machine, error)
	// Remove(name string) error
	// Save(host *machine.Machine) error

	// butta fuori una definizione dell'host
	NewHost(driverName string, RawDriver []byte) (*machine.Machine, error)
	// Crea l'host partendo dalla definizione
	Create(h *machine.Machine) error
	GetMachinesDir() string
}

type Client struct {
	*persist.Filestore
	certsDir       string
	IsDebug        bool
	SSHClientType  ssh.ClientType
	GithubAPIToken string
	storePath      string
	flock          *fslock.Lock
}

func NewClient(storePath, certsDir string) *Client {
	return &Client{
		certsDir:      certsDir,
		IsDebug:       false,
		SSHClientType: ssh.External,
		Filestore:     persist.NewFilestore(storePath, certsDir, certsDir),
	}
}

// NewHost takes a driver name and its configuration in the form of a RawDriver,
// and from those it generates a machine
func (api *Client) NewHost(drvName string, rawDriver []byte) (*machine.Machine, error) {
	def := registry.Driver(drvName)
	if def.Empty() {
		return nil, fmt.Errorf("driver %q does not exist", drvName)
	}

	d := def.Init()
	err := json.Unmarshal(rawDriver, d)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting driver %s", string(rawDriver))
	}

	return &machine.Machine{
		ConfigVersion: version.ConfigVersion,
		Name:          d.GetMachineName(),
		Driver:        d,
		MachineOptions: &machine.Options{
			Driver: d.DriverName(),
			AuthOptions: &auth.Options{
				CertDir:          api.certsDir,
				CaCertPath:       filepath.Join(api.certsDir, "ca.pem"),
				CaPrivateKeyPath: filepath.Join(api.certsDir, "ca-key.pem"),
				ClientCertPath:   filepath.Join(api.certsDir, "cert.pem"),
				ClientKeyPath:    filepath.Join(api.certsDir, "key.pem"),
				ServerCertPath:   filepath.Join(api.GetMachinesDir(), "server.pem"),
				ServerKeyPath:    filepath.Join(api.GetMachinesDir(), "server-key.pem"),
			},
			RuntimeOptions: &cruntime.Options{
				StorageDriver: "overlay2",
				TLSVerify:     true,
			},
		},
	}, nil
}

// Create takes a configured machine.Machine struct and performs the actual
// creation and provisioning of the machine
func (api *Client) Create(m *machine.Machine) error {
	klog.Infof("LocalClient.Create starting")
	start := time.Now()
	defer func() {
		klog.Infof("LocalClient.Create took %s", time.Since(start))
	}()

	def := registry.Driver(m.MachineOptions.Driver)
	if def.Empty() {
		return fmt.Errorf("driver %q does not exist", m.MachineOptions.Driver)
	}

	steps := []struct {
		name string
		f    func() error
	}{
		{
			"bootstrapping certificates",
			func() error {
				// Lock is needed to avoid race condition in parallel Docker-Env test because issue #10107.
				// CA cert and client cert should be generated atomically, otherwise might cause bad certificate error.
				lockErr := api.flock.LockWithTimeout(time.Second * 5)
				if lockErr != nil {
					return fmt.Errorf("failed to acquire bootstrap client lock: %v " + lockErr.Error())
				}
				defer func() {
					lockErr = api.flock.Unlock()
					if lockErr != nil {
						klog.Errorf("failed to release bootstrap cert client lock: %v", lockErr.Error())
					}
				}()
				certErr := cert.BootstrapCertificates(m.AuthOptions())
				return certErr
			},
		},
		{
			"precreate",
			m.Driver.PreCreateCheck,
		},
		{
			"saving",
			func() error {
				return api.Save(m)
			},
		},
		{
			"creating",
			m.Driver.Create,
		},
		{
			"waiting",
			func() error {
				if driver.BareMetal(m.Driver.DriverName()) {
					return nil
				}
				return mcnutils.WaitFor(drivers.MachineInState(m.Driver, state.Running))
			},
		},
		{
			"provisioning",
			func() error {
				// Skippable because we don't reconfigure Docker?
				// TODO: absolutely no
				if driver.BareMetal(m.Driver.DriverName()) {
					return nil
				}
				return m.Provision()
			},
		},
	}

	for _, step := range steps {
		if err := step.f(); err != nil {
			return errors.Wrap(err, step.name)
		}
	}

	return nil
}

// Close ... I'm not sure what purpose it has...
func (api *Client) Close() error { return nil }

// Load reads the machine values from the config.json inside the minikube folder,
// it checks the existing machine's version and performs a migration if needed,
// ultimately it initializes the driver and write the RawDriver into it
func (api *Client) Load(name string) (*machine.Machine, error) {
	// here the machine gets exhumed, and migration is performed if needed
	// at this point the driver in use is libmachine/none
	h, err := api.Filestore.Load(name)
	if err != nil {
		return nil, errors.Wrapf(err, "filestore %q", name)
	}

	// At this point we're initializing the right driver inside the machine struct
	def := registry.Driver(h.MachineOptions.Driver)
	if def.Empty() {
		return nil, fmt.Errorf("driver %q does not exist", h.MachineOptions.Driver)
	}
	h.Driver = def.Init()
	return h, json.Unmarshal(h.RawDriver, h.Driver)
}
