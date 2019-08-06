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

// PopulateKubeConfig populates an api.Config object.
func PopulateKubeConfig(cfg *Setup, kubecfg *api.Config) error {
	var err error
	clusterName := cfg.ClusterName
	cluster := api.NewCluster()
	cluster.Server = cfg.ClusterServerAddress
	if cfg.EmbedCerts {
		cluster.CertificateAuthorityData, err = ioutil.ReadFile(cfg.CertificateAuthority)
		if err != nil {
			return err
		}
	} else {
		cluster.CertificateAuthority = cfg.CertificateAuthority
	}
	kubecfg.Clusters[clusterName] = cluster

	// user
	userName := cfg.ClusterName
	user := api.NewAuthInfo()
	if cfg.EmbedCerts {
		user.ClientCertificateData, err = ioutil.ReadFile(cfg.ClientCertificate)
		if err != nil {
			return err
		}
		user.ClientKeyData, err = ioutil.ReadFile(cfg.ClientKey)
		if err != nil {
			return err
		}
	} else {
		user.ClientCertificate = cfg.ClientCertificate
		user.ClientKey = cfg.ClientKey
	}
	kubecfg.AuthInfos[userName] = user

	// context
	contextName := cfg.ClusterName
	context := api.NewContext()
	context.Cluster = cfg.ClusterName
	context.AuthInfo = userName
	kubecfg.Contexts[contextName] = context

	// Only set current context to minikube if the user has not used the keepContext flag
	if !cfg.KeepContext {
		kubecfg.CurrentContext = cfg.ClusterName
	}

	return nil
}

// SetupKubeConfig reads config from disk, adds the minikube settings, and writes it back.
// activeContext is true when minikube is the CurrentContext
// If no CurrentContext is set, the given name will be used.
func SetupKubeConfig(cfg *Setup) error {
	glog.Infoln("Using kubeconfig: ", cfg.GetKubeConfigFile())

	// read existing config or create new if does not exist
	config, err := readOrNew(cfg.GetKubeConfigFile())
	if err != nil {
		return err
	}

	err = PopulateKubeConfig(cfg, config)
	if err != nil {
		return err
	}

	// write back to disk
	if err := writeToFile(config, cfg.GetKubeConfigFile()); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}
	return nil
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

<<<<<<< HEAD
// WriteConfig encodes the configuration and writes it to the given file.
// If the file exists, its contents will be overwritten.
func WriteConfig(config *api.Config, filename string) error {
||||||| merged common ancestors
// WriteConfig encodes the configuration and writes it to the given file.
// If the file exists, it's contents will be overwritten.
func WriteConfig(config *api.Config, filename string) error {
=======
// writeToFile encodes the configuration and writes it to the given file.
// If the file exists, it's contents will be overwritten.
func writeToFile(config runtime.Object, filename string) error {
>>>>>>> statisfy lint interfacer
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

// GetKubeConfigStatus verifies the ip stored in kubeconfig.
func GetKubeConfigStatus(ip net.IP, filename string, machineName string) (bool, error) {
	if ip == nil {
		return false, fmt.Errorf("error, empty ip passed")
	}
	kip, err := getIPFromKubeConfig(filename, machineName)
	if err != nil {
		return false, err
	}
	if kip.Equal(ip) {
		return true, nil
	}
	// Kubeconfig IP misconfigured
	return false, nil

}

// UpdateKubeconfigIP overwrites the IP stored in kubeconfig with the provided IP.
func UpdateKubeconfigIP(ip net.IP, filename string, machineName string) (bool, error) {
	if ip == nil {
		return false, fmt.Errorf("error, empty ip passed")
	}
	kip, err := getIPFromKubeConfig(filename, machineName)
	if err != nil {
		return false, err
	}
	if kip.Equal(ip) {
		return false, nil
	}
	kport, err := GetPortFromKubeConfig(filename, machineName)
	if err != nil {
		return false, err
	}
	con, err := readOrNew(filename)
	if err != nil {
		return false, errors.Wrap(err, "Error getting kubeconfig status")
	}
	// Safe to lookup server because if field non-existent getIPFromKubeconfig would have given an error
	con.Clusters[machineName].Server = "https://" + ip.String() + ":" + strconv.Itoa(kport)
	err = writeToFile(con, filename)
	if err != nil {
		return false, err
	}
	// Kubeconfig IP reconfigured
	return true, nil
}

// getIPFromKubeConfig returns the IP address stored for minikube in the kubeconfig specified
func getIPFromKubeConfig(filename, machineName string) (net.IP, error) {
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

// GetPortFromKubeConfig returns the Port number stored for minikube in the kubeconfig specified
func GetPortFromKubeConfig(filename, machineName string) (int, error) {
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

// UnsetCurrentContext unsets the current-context from minikube to "" on minikube stop
func UnsetCurrentContext(filename, machineName string) error {
	confg, err := readOrNew(filename)
	if err != nil {
		return errors.Wrap(err, "Error getting kubeconfig status")
	}

	// Unset current-context only if profile is the current-context
	if confg.CurrentContext == machineName {
		confg.CurrentContext = ""
		if err := writeToFile(confg, filename); err != nil {
			return errors.Wrap(err, "writing kubeconfig")
		}
		return nil
	}

	return nil
}

// SetCurrentContext sets the kubectl's current-context
func SetCurrentContext(kubeCfgPath, name string) error {
	kcfg, err := readOrNew(kubeCfgPath)
	if err != nil {
		return errors.Wrap(err, "Error getting kubeconfig status")
	}
	kcfg.CurrentContext = name
	if err := writeToFile(kcfg, kubeCfgPath); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}
	return nil
}

// DeleteKubeConfigContext deletes the specified machine's kubeconfig context
func DeleteKubeConfigContext(kubeCfgPath, machineName string) error {
	kcfg, err := readOrNew(kubeCfgPath)
	if err != nil {
		return errors.Wrap(err, "Error getting kubeconfig status")
	}

	if kcfg == nil || api.IsConfigEmpty(kcfg) {
		glog.V(2).Info("kubeconfig is empty")
		return nil
	}

	delete(kcfg.Clusters, machineName)
	delete(kcfg.AuthInfos, machineName)
	delete(kcfg.Contexts, machineName)

	if kcfg.CurrentContext == machineName {
		kcfg.CurrentContext = ""
	}

	if err := writeToFile(kcfg, kubeCfgPath); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}
	return nil
}

// GetKubeConfigPath gets the path to the first kubeconfig
func GetKubeConfigPath() string {
	kubeConfigEnv := os.Getenv(constants.KubeconfigEnvVar)
	if kubeConfigEnv == "" {
		return constants.KubeconfigPath
	}
	return filepath.SplitList(kubeConfigEnv)[0]
}
