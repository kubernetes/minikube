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
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/minikube/pkg/minikube/constants"
	pkgutil "k8s.io/minikube/pkg/util"
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

// GetKubeConfigPath gets the path to the first kubeconfig
func Path() string {
	kubeConfigEnv := os.Getenv(constants.KubeconfigEnvVar)
	if kubeConfigEnv == "" {
		return constants.KubeconfigPath
	}
	return filepath.SplitList(kubeConfigEnv)[0]
}

// readOrNew retrieves Kubernetes client configuration from a file.
// If no files exists, an empty configuration is returned.
func readOrNew(filename string) (*api.Config, error) {
	data, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return api.NewConfig(), nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "Error reading file %q", filename)
	}

	// decode config, empty if no bytes
	config, err := decode(data)
	if err != nil {
		return nil, errors.Errorf("could not read config: %v", err)
	}

	// initialize nil maps
	if config.AuthInfos == nil {
		config.AuthInfos = map[string]*api.AuthInfo{}
	}
	if config.Clusters == nil {
		config.Clusters = map[string]*api.Cluster{}
	}
	if config.Contexts == nil {
		config.Contexts = map[string]*api.Context{}
	}

	return config, nil
}

// writeToFile encodes the configuration and writes it to the given file.
// If the file exists, it's contents will be overwritten.
func writeToFile(config runtime.Object, filename string) error {
	if config == nil {
		glog.Errorf("could not write to '%s': config can't be nil", filename)
	}

	// encode config to YAML
	data, err := runtime.Encode(latest.Codec, config)
	if err != nil {
		return errors.Errorf("could not write to '%s': failed to encode config: %v", filename, err)
	}

	// create parent dir if doesn't exist
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "Error creating directory: %s", dir)
		}
	}

	// write with restricted permissions
	if err := ioutil.WriteFile(filename, data, 0600); err != nil {
		return errors.Wrapf(err, "Error writing file %s", filename)
	}
	if err := pkgutil.MaybeChownDirRecursiveToMinikubeUser(dir); err != nil {
		return errors.Wrapf(err, "Error recursively changing ownership for dir: %s", dir)
	}

	return nil
}

// extractIP returns the IP address stored for minikube in the kubeconfig specified
func extractIP(filename, machineName string) (net.IP, error) {
	con, err := readOrNew(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting kubeconfig status")
	}
	cluster, ok := con.Clusters[machineName]
	if !ok {
		return nil, errors.Errorf("Kubeconfig does not have a record of the machine cluster")
	}
	kurl, err := url.Parse(cluster.Server)
	if err != nil {
		return net.ParseIP(cluster.Server), nil
	}
	kip, _, err := net.SplitHostPort(kurl.Host)
	if err != nil {
		return net.ParseIP(kurl.Host), nil
	}
	ip := net.ParseIP(kip)
	return ip, nil
}

// decode reads a Config object from bytes.
// Returns empty config if no bytes.
func decode(data []byte) (*api.Config, error) {
	// if no data, return empty config
	if len(data) == 0 {
		return api.NewConfig(), nil
	}

	config, _, err := latest.Codec.Decode(data, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Error decoding config from data: %s", string(data))
	}

	return config.(*api.Config), nil
}
