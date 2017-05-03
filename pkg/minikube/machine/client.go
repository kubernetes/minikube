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

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/provision"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/check"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/version"
	"github.com/pkg/errors"
)

type driverGetter func([]byte) (drivers.Driver, error)

type ClientType int
type clientFactory interface {
	NewClient(string, string) libmachine.API
}

type localClientFactory struct{}

func (*localClientFactory) NewClient(storePath, certsDir string) libmachine.API {
	return &LocalClient{
		certsDir:  certsDir,
		storePath: storePath,
		Filestore: persist.NewFilestore(storePath, certsDir, certsDir),
	}
}

type rpcClientFactory struct{}

func (*rpcClientFactory) NewClient(storePath, certsDir string) libmachine.API {
	c := libmachine.NewClient(storePath, certsDir)
	c.SSHClientType = ssh.Native
	return c
}

var clientFactories = map[ClientType]clientFactory{
	ClientTypeLocal: &localClientFactory{},
	ClientTypeRPC:   &rpcClientFactory{},
}

const (
	ClientTypeLocal ClientType = iota
	ClientTypeRPC
)

// Gets a new client depending on the clientType specified
// defaults to the libmachine client
func NewAPIClient(clientType ClientType) (libmachine.API, error) {
	storePath := constants.GetMinipath()
	certsDir := constants.MakeMiniPath("certs")
	newClientFactory, ok := clientFactories[clientType]
	if !ok {
		return nil, fmt.Errorf("No implementation for API client type %d", clientType)
	}

	return newClientFactory.NewClient(storePath, certsDir), nil
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
	*persist.Filestore
}

func (api *LocalClient) NewHost(driverName string, rawDriver []byte) (*host.Host, error) {
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

	h.Driver, err = getDriver(h.DriverName, h.RawDriver)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading driver from host")
	}

	return h, nil
}

func (api *LocalClient) Close() error { return nil }

func (api *LocalClient) Create(h *host.Host) error {
	steps := []struct {
		name string
		f    func() error
	}{
		{
			"Bootstrapping certs.",
			func() error { return cert.BootstrapCertificates(h.AuthOptions()) },
		},
		{
			"Running precreate checks.",
			h.Driver.PreCreateCheck,
		},
		{
			"Saving driver.",
			func() error {
				return api.Save(h)
			},
		},
		{
			"Creating VM.",
			h.Driver.Create,
		},
		{
			"Waiting for VM to start.",
			func() error {
				return mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running))
			},
		},
		{
			"Provisioning VM.",
			func() error {
				pv := provision.NewBuildrootProvisioner(h.Driver)
				return pv.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
			},
		},
	}

	for _, step := range steps {
		if err := step.f(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error executing step: %s\n", step.name))
		}
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
