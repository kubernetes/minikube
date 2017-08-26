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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/provision"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/check"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	rpcdriver "github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/version"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type driverGetter func([]byte) (drivers.Driver, error)

func NewRPCClient(storePath, certsDir string) libmachine.API {
	c := libmachine.NewClient(storePath, certsDir)
	c.SSHClientType = ssh.Native
	return c
}

// NewAPIClient gets a new client.
func NewAPIClient() (*LocalClient, error) {
	storePath := constants.GetMinipath()
	certsDir := constants.MakeMiniPath("certs")

	return &LocalClient{
		certsDir:     certsDir,
		storePath:    storePath,
		EventCh:      make(chan string),
		ErrorCh:      make(chan error),
		Filestore:    persist.NewFilestore(storePath, certsDir, certsDir),
		legacyClient: NewRPCClient(storePath, certsDir),
	}, nil
}

func getDriver(driverName string, rawDriver []byte) (drivers.Driver, error) {
	driverGetter, ok := driverMap[driverName]
	if !ok {
		return nil, fmt.Errorf("Unknown driver %s for platform.", driverName)
	}
	driver, err := driverGetter(rawDriver)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting driver for %s", driverName)
	}

	return driver, nil
}

func getVirtualboxDriver(rawDriver []byte) (drivers.Driver, error) {
	var driver drivers.Driver
	driver = virtualbox.NewDriver("", "")
	err := json.Unmarshal(rawDriver, driver)
	if err != nil {
		return nil, errors.Wrapf(err, "Error unmarshalling virtualbox driver %s", string(rawDriver))
	}
	return driver, nil
}

func getDriverRPC(driverName string, rawDriver []byte) (drivers.Driver, error) {
	return rpcdriver.NewRPCClientDriverFactory().NewRPCClientDriver(driverName, rawDriver)
}

// LocalClient is a non-RPC implemenation
// of the libmachine API
type LocalClient struct {
	certsDir  string
	storePath string

	EventCh chan string
	ErrorCh chan error

	*persist.Filestore
	legacyClient libmachine.API
}

func (api *LocalClient) NewHost(driverName string, rawDriver []byte) (*host.Host, error) {
	// If not should get Driver, use legacy
	if _, ok := driverMap[driverName]; !ok {
		return api.legacyClient.NewHost(driverName, rawDriver)
	}

	driver, err := getDriver(driverName, rawDriver)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting driver")
	}
	return &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          driver.GetMachineName(),
		Driver:        driver,
		DriverName:    driver.DriverName(),
		HostOptions: &host.Options{
			AuthOptions: &auth.Options{
				CertDir:          api.certsDir,
				CaCertPath:       filepath.Join(api.certsDir, "ca.pem"),
				CaPrivateKeyPath: filepath.Join(api.certsDir, "ca-key.pem"),
				ClientCertPath:   filepath.Join(api.certsDir, "cert.pem"),
				ClientKeyPath:    filepath.Join(api.certsDir, "key.pem"),
				ServerCertPath:   filepath.Join(api.GetMachinesDir(), "server.pem"),
				ServerKeyPath:    filepath.Join(api.GetMachinesDir(), "server-key.pem"),
			},
			EngineOptions: &engine.Options{
				StorageDriver: "aufs",
				TLSVerify:     true,
			},
			SwarmOptions: &swarm.Options{},
		},
	}, nil
}

func (api *LocalClient) Load(name string) (*host.Host, error) {
	h, err := api.Filestore.Load(name)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading host from store")
	}

	// If not should get Driver, use legacy
	if _, ok := driverMap[h.DriverName]; !ok {
		return api.legacyClient.Load(name)
	}

	h.Driver, err = getDriver(h.DriverName, h.RawDriver)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading driver from host")
	}

	return h, nil
}

func GetCommandRunner(h *host.Host) (bootstrapper.CommandRunner, error) {
	if h.DriverName != constants.DriverNone {
		client, err := sshutil.NewSSHClient(h.Driver)
		if err != nil {
			return nil, errors.Wrap(err, "getting ssh client for bootstrapper")
		}
		return bootstrapper.NewSSHRunner(client), nil
	}

	return &bootstrapper.ExecRunner{}, nil
}

func (api *LocalClient) Close() error {
	if api.legacyClient != nil {
		return api.legacyClient.Close()
	}
	return nil
}

const (
	BootstrapCertsEvent         = "Bootstrapping certs."
	PreCreateCheckEvent         = "Running precreate checks."
	SaveConfigBeforeCreateEvent = "Saving driver before create."
	CreateVMEvent               = "Creating VM."
	ResumeEvent                 = "Starting VM."
	WaitForStartEvent           = "Waiting for VM to start."
	SaveConfigRunningVM         = "Saving driver after creating VM."
	ProvisionEvent              = "Provisioning VM."
)

type machineStep struct {
	name string
	f    func() error
}

func provisionStep(h *host.Host) machineStep {
	return machineStep{
		name: ProvisionEvent,
		f: func() error {
			if h.Driver.DriverName() == "none" {
				return nil
			}
			pv := provision.NewBuildrootProvisioner(h.Driver)
			return pv.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
		},
	}
}

// Create runs the steps necessary to create, start, and provision a VM.
//
// Create publishes events as they are completed on the api.EventCh,
// and errors on the api.ErrorCh
func (api *LocalClient) Create(h *host.Host) error {
	steps := []machineStep{
		{
			BootstrapCertsEvent,
			func() error { return cert.BootstrapCertificates(h.AuthOptions()) },
		},
		{
			PreCreateCheckEvent,
			h.Driver.PreCreateCheck,
		},
		{
			SaveConfigBeforeCreateEvent,
			func() error {
				return api.Save(h)
			},
		},
		{
			CreateVMEvent,
			h.Driver.Create,
		},
		{
			WaitForStartEvent,
			func() error {
				if h.Driver.DriverName() == "none" {
					return nil
				}
				return mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running))
			},
		},
		{
			SaveConfigRunningVM,
			func() error {
				return api.Save(h)
			},
		},
		provisionStep(h),
	}

	if err := api.runSteps(steps); err != nil {
		return errors.Wrap(err, "Error executing steps")
	}

	return nil
}

func (api *LocalClient) runSteps(steps []machineStep) error {
	for _, step := range steps {
		if err := step.f(); err != nil {
			api.ErrorCh <- err
			return errors.Wrap(err, fmt.Sprintf("step: %s\n", step.name))
		}
		api.EventCh <- step.name
	}
	return nil
}

// Start runs the steps necessary to start and provision a VM that has been created.
//
// Start publishes start events as they are completed on the api.EventCh,
// and errors on the api.ErrorCh
func (api *LocalClient) Start(h *host.Host) error {
	var steps []machineStep
	s, err := h.Driver.GetState()
	glog.Infoln("Machine state: ", s)
	if err != nil {
		return errors.Wrap(err, "Error getting state for host")
	}

	// If we are not in a running state, add a start event.
	if s != state.Running {
		steps = append(steps, machineStep{
			ResumeEvent,
			func() error { return h.Driver.Start() },
		})
	}
	steps = append(steps, machineStep{
		SaveConfigRunningVM,
		func() error {
			return api.Save(h)
		},
	})
	steps = append(steps, provisionStep(h))

	if err := api.runSteps(steps); err != nil {
		return errors.Wrap(err, "Error executing steps")
	}
	return nil
}

func StartDriver() {
	cert.SetCertGenerator(&CertGenerator{})
	check.DefaultConnChecker = &ConnChecker{}
	if os.Getenv(localbinary.PluginEnvKey) == localbinary.PluginEnvVal {
		registerDriver(os.Getenv(localbinary.PluginEnvDriverName))
	}

	localbinary.CurrentBinaryIsDockerMachine = true
}

type ConnChecker struct {
}

func (cc *ConnChecker) Check(h *host.Host, swarm bool) (string, *auth.Options, error) {
	authOptions := h.AuthOptions()
	dockerHost, err := h.Driver.GetURL()
	if err != nil {
		return "", &auth.Options{}, err
	}
	return dockerHost, authOptions, nil
}

// CertGenerator is used to override the default machine CertGenerator with a longer timeout.
type CertGenerator struct {
	cert.X509CertGenerator
}

// ValidateCertificate is a reimplementation of the default generator with a longer timeout.
func (cg *CertGenerator) ValidateCertificate(addr string, authOptions *auth.Options) (bool, error) {
	tlsConfig, err := cg.ReadTLSConfig(addr, authOptions)
	if err != nil {
		return false, err
	}

	dialer := &net.Dialer{
		Timeout: time.Second * 40,
	}

	_, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return false, err
	}

	return true, nil
}
