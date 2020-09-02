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
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/lock"
)

// VerifyEndpoint verifies the IP:port stored in kubeconfig.
func VerifyEndpoint(contextName string, hostname string, port int, configPath ...string) error {
	path := PathFromEnv()
	if configPath != nil {
		path = configPath[0]
	}

	if hostname == "" {
		return fmt.Errorf("empty IP")
	}

	gotHostname, gotPort, err := Endpoint(contextName, path)
	if err != nil {
		return errors.Wrap(err, "extract IP")
	}

	if hostname != gotHostname || port != gotPort {
		return fmt.Errorf("got: %s:%d, want: %s:%d", gotHostname, gotPort, hostname, port)
	}

	return nil
}

// PathFromEnv gets the path to the first kubeconfig
func PathFromEnv() string {
	kubeConfigEnv := os.Getenv(constants.KubeconfigEnvVar)
	if kubeConfigEnv == "" {
		return constants.KubeconfigPath
	}
	kubeConfigFiles := filepath.SplitList(kubeConfigEnv)
	for _, kubeConfigFile := range kubeConfigFiles {
		if kubeConfigFile != "" {
			return kubeConfigFile
		}
		glog.Infof("Ignoring empty entry in %s env var", constants.KubeconfigEnvVar)
	}
	return constants.KubeconfigPath
}

// Endpoint returns the IP:port address stored for minikube in the kubeconfig specified
func Endpoint(contextName string, configPath ...string) (string, int, error) {
	path := PathFromEnv()
	if configPath != nil {
		path = configPath[0]
	}
	apiCfg, err := readOrNew(path)
	if err != nil {
		return "", 0, errors.Wrap(err, "read")
	}
	cluster, ok := apiCfg.Clusters[contextName]
	if !ok {
		return "", 0, errors.Errorf("%q does not appear in %s", contextName, path)
	}

	glog.Infof("found %q server: %q", contextName, cluster.Server)
	u, err := url.Parse(cluster.Server)
	if err != nil {
		return "", 0, errors.Wrap(err, "url parse")
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return "", 0, errors.Wrap(err, "atoi")
	}

	return u.Hostname(), port, nil
}

// UpdateEndpoint overwrites the IP stored in kubeconfig with the provided IP.
func UpdateEndpoint(contextName string, hostname string, port int, confpath string) (bool, error) {
	if hostname == "" {
		return false, fmt.Errorf("empty ip")
	}

	err := VerifyEndpoint(contextName, hostname, port, confpath)
	if err == nil {
		return false, nil
	}
	glog.Infof("verify returned: %v", err)

	cfg, err := readOrNew(confpath)
	if err != nil {
		return false, errors.Wrap(err, "read")
	}

	address := "https://" + hostname + ":" + strconv.Itoa(port)

	// if the cluster setting is missed in the kubeconfig, create new one
	if _, ok := cfg.Clusters[contextName]; !ok {
		lp := localpath.Profile(contextName)
		gp := localpath.MiniPath()
		kcs := &Settings{
			ClusterName:          contextName,
			ClusterServerAddress: address,
			ClientCertificate:    path.Join(lp, "client.crt"),
			ClientKey:            path.Join(lp, "client.key"),
			CertificateAuthority: path.Join(gp, "ca.crt"),
			KeepContext:          false,
		}
		err = PopulateFromSettings(kcs, cfg)
		if err != nil {
			return false, errors.Wrap(err, "populating kubeconfig")
		}
	} else {
		cfg.Clusters[contextName].Server = address
	}

	err = writeToFile(cfg, confpath)
	if err != nil {
		return false, errors.Wrap(err, "write")
	}

	return true, nil
}

// writeToFile encodes the configuration and writes it to the given file.
// If the file exists, it's contents will be overwritten.
func writeToFile(config runtime.Object, configPath ...string) error {
	fPath := PathFromEnv()
	if configPath != nil {
		fPath = configPath[0]
	}

	if config == nil {
		glog.Errorf("could not write to '%s': config can't be nil", fPath)
	}

	// encode config to YAML
	data, err := runtime.Encode(latest.Codec, config)
	if err != nil {
		return errors.Errorf("could not write to '%s': failed to encode config: %v", fPath, err)
	}

	// create parent dir if doesn't exist
	dir := filepath.Dir(fPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "Error creating directory: %s", dir)
		}
	}

	// write with restricted permissions
	if err := lock.WriteFile(fPath, data, 0600); err != nil {
		return errors.Wrapf(err, "Error writing file %s", fPath)
	}

	if err := pkgutil.MaybeChownDirRecursiveToMinikubeUser(dir); err != nil {
		return errors.Wrapf(err, "Error recursively changing ownership for dir: %s", dir)
	}

	return nil
}

// readOrNew retrieves Kubernetes client configuration from a file.
// If no files exists, an empty configuration is returned.
func readOrNew(configPath ...string) (*api.Config, error) {
	fPath := PathFromEnv()
	if configPath != nil {
		fPath = configPath[0]
	}

	data, err := ioutil.ReadFile(fPath)
	if os.IsNotExist(err) {
		return api.NewConfig(), nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "Error reading file %q", fPath)
	}

	// decode config, empty if no bytes
	kcfg, err := decode(data)
	if err != nil {
		return nil, errors.Errorf("could not read config: %v", err)
	}

	// initialize nil maps
	if kcfg.AuthInfos == nil {
		kcfg.AuthInfos = map[string]*api.AuthInfo{}
	}
	if kcfg.Clusters == nil {
		kcfg.Clusters = map[string]*api.Cluster{}
	}
	if kcfg.Contexts == nil {
		kcfg.Contexts = map[string]*api.Context{}
	}

	return kcfg, nil
}

// decode reads a Config object from bytes.
// Returns empty config if no bytes.
func decode(data []byte) (*api.Config, error) {
	// if no data, return empty config
	if len(data) == 0 {
		return api.NewConfig(), nil
	}

	kcfg, _, err := latest.Codec.Decode(data, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Error decoding config from data: %s", string(data))
	}

	return kcfg.(*api.Config), nil
}
