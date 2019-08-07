/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package kubeconfig

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
)

// VeryifyMachineIP verifies the ip stored in kubeconfig.
func VeryifyMachineIP(ip net.IP, filename string, machineName string) (bool, error) {
	if ip == nil {
		return false, fmt.Errorf("error, empty ip passed")
	}
	kip, err := extractIP(filename, machineName)
	if err != nil {
		return false, err
	}
	if kip.Equal(ip) {
		return true, nil
	}
	// Kubeconfig IP misconfigured
	return false, nil

}

// GetPortFromKubeConfig returns the Port number stored for minikube in the kubeconfig specified
func Port(filename, machineName string) (int, error) {
	con, err := readOrNew(filename)
	if err != nil {
		return 0, errors.Wrap(err, "Error getting kubeconfig status")
	}
	cluster, ok := con.Clusters[machineName]
	if !ok {
		return 0, errors.Errorf("Kubeconfig does not have a record of the machine cluster")
	}
	kurl, err := url.Parse(cluster.Server)
	if err != nil {
		return constants.APIServerPort, nil
	}
	_, kport, err := net.SplitHostPort(kurl.Host)
	if err != nil {
		return constants.APIServerPort, nil
	}
	port, err := strconv.Atoi(kport)
	return port, err
}

// Path() gets the path to the first kubeconfig
func Path() string {
	kubeConfigEnv := os.Getenv(constants.KubeconfigEnvVar)
	if kubeConfigEnv == "" {
		return constants.KubeconfigPath
	}
	return filepath.SplitList(kubeConfigEnv)[0]
}
