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
	"log"
	"strings"
	"time"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/cli/constants"
)

// StartHost starts a host VM.
func StartHost(api libmachine.API) (*host.Host, error) {
	if exists, err := api.Exists(constants.MachineName); err != nil {
		return nil, fmt.Errorf("Error checking if host exists: %s", err)
	} else if exists {
		log.Println("Machine exists!")
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
		}
		return h, nil
	} else {
		return createHost(api)
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

// StartCluster starts as k8s cluster on the specified Host.
func StartCluster(h sshAble) error {
	for _, cmd := range []string{
		// Download and install weave, if it doesn't exist.
		`if [ ! -e /usr/local/bin/weave ]; then
		   sudo curl -L git.io/weave -o /usr/local/bin/weave
		   sudo chmod a+x /usr/local/bin/weave;
		 fi`,
		// Download and install localkube, if it doesn't exist yet.
		`if [ ! -e /usr/local/bin/localkube ];
		 then
		   sudo curl -L https://github.com/redspread/localkube/releases/download/v1.2.1-v1/localkube-linux -o /usr/local/bin/localkube
		   sudo chmod a+x /usr/local/bin/localkube;
		 fi`,
		// Start weave.
		"weave launch-router",
		"weave launch-proxy --without-dns --rewrite-inspect",
		"weave expose -h \"localkube.weave.local\"",
		// Localkube assumes containerized kubelet, which looks at /rootfs.
		"if [ ! -e /rootfs ]; then sudo ln -s / /rootfs; fi",
		// Run with nohup so it stays up. Redirect logs to useful places.
		"PATH=/usr/local/sbin:$PATH nohup sudo /usr/local/bin/localkube start > /var/log/localkube.out 2> /var/log/localkube.err < /dev/null &"} {
		output, err := h.RunSSHCommand(cmd)
		log.Println(output)
		if err != nil {
			return err
		}
	}

	return nil
}

func createHost(api libmachine.API) (*host.Host, error) {
	driver := virtualbox.NewDriver(constants.MachineName, constants.Minipath)
	driver.Boot2DockerURL = "https://storage.googleapis.com/tinykube/boot2docker.iso"
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
