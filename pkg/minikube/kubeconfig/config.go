package kubeconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api/latest"
	"k8s.io/kubernetes/pkg/runtime"
)

// ReadConfigOrNew retrieves Kubernetes client configuration from a file.
// If no files exists, an empty configuration is returned.
func ReadConfigOrNew(filename string) (*api.Config, error) {
	data, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return api.NewConfig(), nil
	} else if err != nil {
		return nil, err
	}

	// decode config, empty if no bytes
	config, err := decode(data)
	if err != nil {
		return nil, fmt.Errorf("could not read config: %v", err)
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

// WriteConfig encodes the configuration and writes it to the given file.
// If the file exists, it's contents will be overwritten.
func WriteConfig(config *api.Config, filename string) error {
	if config == nil {
		fmt.Errorf("could not write to '%s': config can't be nil", filename)
	}

	// encode config to YAML
	data, err := runtime.Encode(latest.Codec, config)
	if err != nil {
		return fmt.Errorf("could not write to '%s': failed to encode config: %v", filename, err)
	}

	// create parent dir if doesn't exist
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// write with restricted permissions
	if err := ioutil.WriteFile(filename, data, 0600); err != nil {
		return err
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
		return nil, err
	}

	return config.(*api.Config), nil
}
