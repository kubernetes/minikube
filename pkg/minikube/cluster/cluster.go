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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/util"
)

var (
	certs = []string{"apiserver.crt", "apiserver.key"}
)

// StartHost starts a host VM.
func StartHost(api libmachine.API, config MachineConfig) (*host.Host, error) {
	if exists, err := api.Exists(constants.MachineName); err != nil {
		return nil, fmt.Errorf("Error checking if host exists: %s", err)
	} else if exists {
		glog.Infoln("Machine exists!")
		h, err := api.Load(constants.MachineName)
		if err != nil {
			return nil, fmt.Errorf("Error loading existing host: %s", err)
		}
		s, err := h.Driver.GetState()
		if err != nil {
			return nil, fmt.Errorf("Error getting state for host: %s", err)
		}
		if s != state.Running {
			if err := h.Driver.Start(); err != nil {
				return nil, fmt.Errorf("Error starting stopped host: %s", err)
			}
			if err := api.Save(h); err != nil {
				return nil, fmt.Errorf("Error saving started host: %s", err)
			}
		}
		return h, nil
	} else {
		return createHost(api, config)
	}
}

// StopHost stops the host VM.
func StopHost(api libmachine.API) error {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		return err
	}
	if err := host.Stop(); err != nil {
		return err
	}
	return nil
}

type multiError struct {
	Errors []error
}

func (m *multiError) Collect(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

func (m multiError) ToError() error {
	if len(m.Errors) == 0 {
		return nil
	}

	errStrings := []string{}
	for _, err := range m.Errors {
		errStrings = append(errStrings, err.Error())
	}
	return fmt.Errorf(strings.Join(errStrings, "\n"))
}

// DeleteHost deletes the host VM.
func DeleteHost(api libmachine.API) error {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		return err
	}
	m := multiError{}
	m.Collect(host.Driver.Remove())
	m.Collect(api.Remove(constants.MachineName))
	return m.ToError()
}

// GetHostStatus gets the status of the host VM.
func GetHostStatus(api libmachine.API) (string, error) {
	dne := "Does Not Exist"
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return "", err
	}
	if !exists {
		return dne, nil
	}

	host, err := api.Load(constants.MachineName)
	if err != nil {
		return "", err
	}

	s, err := host.Driver.GetState()
	if s.String() == "" {
		return dne, err
	}
	return s.String(), err
}

type sshAble interface {
	RunSSHCommand(string) (string, error)
}

// MachineConfig contains the parameters used to start a cluster.
type MachineConfig struct {
	MinikubeISO string
}

// StartCluster starts a k8s cluster on the specified Host.
func StartCluster(h sshAble) error {
	commands := []string{stopCommand, GetStartCommand()}

	for _, cmd := range commands {
		output, err := h.RunSSHCommand(cmd)
		glog.Infoln(output)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateCluster(d drivers.Driver) error {
	localkube, err := Asset("out/localkube")
	if err != nil {
		glog.Infoln("Error loading localkube: ", err)
		return err
	}

	client, err := sshutil.NewSSHClient(d)
	if err != nil {
		return err
	}
	if err := sshutil.Transfer(localkube, "/usr/local/bin/", "localkube", "0777", client); err != nil {
		return err
	}

	return nil
}

// SetupCerts gets the generated credentials required to talk to the APIServer.
func SetupCerts(d drivers.Driver) error {
	localPath := constants.Minipath
	ipStr, err := d.GetIP()
	if err != nil {
		return err
	}

	ip := net.ParseIP(ipStr)
	publicPath := filepath.Join(localPath, "apiserver.crt")
	privatePath := filepath.Join(localPath, "apiserver.key")
	if err := GenerateCerts(publicPath, privatePath, ip); err != nil {
		return err
	}

	client, err := sshutil.NewSSHClient(d)
	if err != nil {
		return err
	}

	for _, cert := range certs {
		p := filepath.Join(localPath, cert)
		data, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		if err := sshutil.Transfer(data, util.DefaultCertPath, cert, "0644", client); err != nil {
			return err
		}
	}
	return nil
}

func createHost(api libmachine.API, config MachineConfig) (*host.Host, error) {
	driver := virtualbox.NewDriver(constants.MachineName, constants.Minipath)
	driver.Boot2DockerURL = config.MinikubeISO
	data, err := json.Marshal(driver)
	if err != nil {
		return nil, err
	}

	driverName := "virtualbox"
	h, err := api.NewHost(driverName, data)
	if err != nil {
		return nil, fmt.Errorf("Error creating new host: %s", err)
	}

	h.HostOptions.AuthOptions.CertDir = constants.Minipath
	h.HostOptions.AuthOptions.StorePath = constants.Minipath

	if err := api.Create(h); err != nil {
		// Wait for all the logs to reach the client
		time.Sleep(2 * time.Second)
		return nil, fmt.Errorf("Error creating. %s", err)
	}

	if err := api.Save(h); err != nil {
		return nil, fmt.Errorf("Error attempting to save store: %s", err)
	}
	return h, nil
}

// GetHostDockerEnv gets the necessary docker env variables to allow the use of docker through minikube's vm
func GetHostDockerEnv(api libmachine.API) (map[string]string, error) {
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return nil, err
	}
	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, err
	}

	tcpPrefix := "tcp://"
	portDelimiter := ":"
	port := "2376"

	envMap := map[string]string{
		"DOCKER_TLS_VERIFY": "1",
		"DOCKER_HOST":       tcpPrefix + ip + portDelimiter + port,
		"DOCKER_CERT_PATH":  constants.MakeMiniPath("certs"),
	}
	return envMap, nil
}

// GetHostLogs gets the localkube logs of the host VM.
func GetHostLogs(api libmachine.API) (string, error) {
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}
	s, err := host.RunSSHCommand(logsCommand)
	if err != nil {
		return "", nil
	}
	return s, err
}

func checkIfApiExistsAndLoad(api libmachine.API) (*host.Host, error) {
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("Machine does not exist for api.Exists(%s)", constants.MachineName)
	}

	host, err := api.Load(constants.MachineName)
	if err != nil {
		return nil, err
	}
	return host, nil
}
