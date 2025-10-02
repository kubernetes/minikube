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
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/lock"
)

// UpdateEndpoint overwrites the IP stored in kubeconfig with the provided IP.
// It will also fix missing cluster or context in kubeconfig, if needed.
// Returns if the change was made and any error occurred.
func UpdateEndpoint(contextName string, host string, port int, configPath string, ext *Extension) (bool, error) {
	if host == "" {
		return false, fmt.Errorf("empty host")
	}

	if err := VerifyEndpoint(contextName, host, port, configPath); err != nil {
		klog.Infof("verify endpoint returned: %v", err)
	}

	cfg, err := readOrNew(configPath)
	if err != nil {
		return false, errors.Wrap(err, "get kubeconfig")
	}

        // Bracket IPv6 literals for a valid URL (e.g. https://[::1]:8443)
        hostForURL := host
        if strings.Contains(hostForURL, ":") && !strings.HasPrefix(hostForURL, "[") {
                hostForURL = "[" + hostForURL + "]"
        }

        address := "https://" + hostForURL + ":" + strconv.Itoa(port)

	// check & fix kubeconfig if the cluster or context setting is missing, or server address needs updating
	errs := configIssues(cfg, contextName, address)
	if errs == nil {
		return false, nil
	}
	klog.Infof("%s needs updating (will repair): %v", configPath, errs)

	kcs := &Settings{
		ClusterName:          contextName,
		ClusterServerAddress: address,
		KeepContext:          false,
	}

	populateCerts(kcs, *cfg, contextName)

	if ext != nil {
		kcs.ExtensionCluster = ext
	}
	if err = PopulateFromSettings(kcs, cfg); err != nil {
		return false, errors.Wrap(err, "populate kubeconfig")
	}

	err = writeToFile(cfg, configPath)
	if err != nil {
		return false, errors.Wrap(err, "write kubeconfig")
	}

	return true, nil
}

// VerifyEndpoint verifies the host:port stored in kubeconfig.
func VerifyEndpoint(contextName string, host string, port int, configPath string) error {
	if host == "" {
		return fmt.Errorf("empty host")
	}

	if configPath == "" {
		configPath = PathFromEnv()
	}

	gotHost, gotPort, err := Endpoint(contextName, configPath)
	if err != nil {
		return errors.Wrap(err, "get endpoint")
	}

	if host != gotHost || port != gotPort {
		return fmt.Errorf("got: %s:%d, want: %s:%d", gotHost, gotPort, host, port)
	}

	return nil
}

// Endpoint returns the IP:port address stored for minikube in the kubeconfig specified.
func Endpoint(contextName string, configPath string) (string, int, error) {
	if configPath == "" {
		configPath = PathFromEnv()
	}

	apiCfg, err := readOrNew(configPath)
	if err != nil {
		return "", 0, errors.Wrap(err, "read kubeconfig")
	}

	cluster, ok := apiCfg.Clusters[contextName]
	if !ok {
		return "", 0, errors.Errorf("%q does not appear in %s", contextName, configPath)
	}

	klog.Infof("found %q server: %q", contextName, cluster.Server)
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

// configIssues returns list of errors found in kubeconfig for given contextName and server address.
func configIssues(cfg *api.Config, contextName string, address string) []error {
	errs := []error{}
	if _, ok := cfg.Clusters[contextName]; !ok {
		errs = append(errs, errors.Errorf("kubeconfig missing %q cluster setting", contextName))
	} else if cfg.Clusters[contextName].Server != address {
		errs = append(errs, errors.Errorf("kubeconfig needs server address update"))
	}

	if _, ok := cfg.Contexts[contextName]; !ok {
		errs = append(errs, errors.Errorf("kubeconfig missing %q context setting", contextName))
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// populateCerts retains certs already defined in kubeconfig or sets default ones for those missing.
func populateCerts(kcs *Settings, cfg api.Config, contextName string) {
	lp := localpath.Profile(contextName)
	gp := localpath.MiniPath()

	kcs.CertificateAuthority = path.Join(gp, "ca.crt")
	if cluster, ok := cfg.Clusters[contextName]; ok {
		kcs.CertificateAuthority = cluster.CertificateAuthority
	}

	kcs.ClientCertificate = path.Join(lp, "client.crt")
	kcs.ClientKey = path.Join(lp, "client.key")
	if context, ok := cfg.Contexts[contextName]; ok {
		if user, ok := cfg.AuthInfos[context.AuthInfo]; ok {
			kcs.ClientCertificate = user.ClientCertificate
			kcs.ClientKey = user.ClientKey
		}
	}
}

// readOrNew retrieves Kubernetes client configuration from a file.
// If no files exists, an empty configuration is returned.
func readOrNew(configPath string) (*api.Config, error) {
	if configPath == "" {
		configPath = PathFromEnv()
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return api.NewConfig(), nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "read kubeconfig from %q", configPath)
	}

	// decode config, empty if no bytes
	kcfg, err := decode(data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode kubeconfig from %q", configPath)
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
		return nil, errors.Wrapf(err, "decode data: %s", string(data))
	}

	return kcfg.(*api.Config), nil
}

// writeToFile encodes the configuration and writes it to the given file.
// If the file exists, it's contents will be overwritten.
func writeToFile(config runtime.Object, configPath string) error {
	if configPath == "" {
		configPath = PathFromEnv()
	}

	if config == nil {
		klog.Errorf("could not write to '%s': config can't be nil", configPath)
	}

	// encode config to YAML
	data, err := runtime.Encode(latest.Codec, config)
	if err != nil {
		return errors.Errorf("could not write to '%s': failed to encode config: %v", configPath, err)
	}

	// create parent dir if doesn't exist
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "Error creating directory: %s", dir)
		}
	}

	// write with restricted permissions
	if err := lock.WriteFile(configPath, data, 0600); err != nil {
		return errors.Wrapf(err, "Error writing file %s", configPath)
	}

	if err := pkgutil.MaybeChownDirRecursiveToMinikubeUser(dir); err != nil {
		return errors.Wrapf(err, "Error recursively changing ownership for dir: %s", dir)
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
		klog.Infof("Ignoring empty entry in %s env var", constants.KubeconfigEnvVar)
	}
	return constants.KubeconfigPath
}
