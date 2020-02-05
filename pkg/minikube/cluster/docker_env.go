/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"net"

	"github.com/docker/machine/libmachine"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// GetNodeDockerEnv gets the necessary docker env variables to allow the use of docker through minikube's vm
func GetNodeDockerEnv(api libmachine.API) (map[string]string, error) {
	pName := viper.GetString(config.MachineProfile)
	host, err := CheckIfHostExistsAndLoad(api, pName)
	if err != nil {
		return nil, errors.Wrap(err, "Error checking that api exists and loading it")
	}

	ip := kic.DefaultBindIPV4
	if !driver.IsKIC(host.Driver.DriverName()) { // kic externally accessible ip is different that node ip
		ip, err = host.Driver.GetIP()
		if err != nil {
			return nil, errors.Wrap(err, "Error getting ip from host")
		}

	}

	tcpPrefix := "tcp://"
	port := constants.DockerDaemonPort
	if driver.IsKIC(host.Driver.DriverName()) { // for kic we need to find out what port docker allocated during creation
		port, err = oci.HostPortBinding(host.Driver.DriverName(), pName, constants.DockerDaemonPort)
		if err != nil {
			return nil, errors.Wrapf(err, "get hostbind port for %d", constants.DockerDaemonPort)
		}
	}

	envMap := map[string]string{
		constants.DockerTLSVerifyEnv:       "1",
		constants.DockerHostEnv:            tcpPrefix + net.JoinHostPort(ip, fmt.Sprint(port)),
		constants.DockerCertPathEnv:        localpath.MakeMiniPath("certs"),
		constants.MinikubeActiveDockerdEnv: pName,
	}
	return envMap, nil
}
