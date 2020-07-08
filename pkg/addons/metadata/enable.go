package metadata

import (
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/homedir"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
)

// EnableOrDisable enables or disables the metadata addon depending on the val parameter
func EnableOrDisable(cfg *config.ClusterConfig, name string, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	if enable {
		return enableAddon(cfg)
	}
	return disableAddon(cfg)

}

func enableAddon(cfg *config.ClusterConfig) error {
	// This is the default location for GCP credentials to live, it's where they're stored when gcloud login is run
	credsPath := filepath.Join(homedir.HomeDir(), ".config", "gcloud", "application_default_credentials.json")
	f, err := assets.NewFileAsset(credsPath, "/tmp/", "google_application_credentials.json", "0444")
	if err != nil {
		return err
	}

	api, err := machine.NewAPIClient()
	if err != nil {
		return err
	}

	host, err := machine.LoadHost(api, driver.MachineName(*cfg, cfg.Nodes[0]))
	if err != nil {
		return err
	}

	r, err := machine.CommandRunner(host)
	if err != nil {
		return err
	}

	err = r.Copy(f)
	if err != nil {
		return err
	}

	// We're currently assuming gcloud is installed and in the user's path
	project, err := exec.Command("gcloud", "config", "get-value", "project").Output()
	if err == nil && len(project) > 0 {
		f := assets.NewMemoryAssetTarget(project, "/tmp/google_cloud_project", "0444")
		return r.Copy(f)
	}

	return nil
}

func disableAddon(cfg *config.ClusterConfig) error {
	api, err := machine.NewAPIClient()
	if err != nil {
		return err
	}

	host, err := machine.LoadHost(api, driver.MachineName(*cfg, cfg.Nodes[0]))
	if err != nil {
		return err
	}

	r, err := machine.CommandRunner(host)
	if err != nil {
		return err
	}

	// Clean up the files generated when enabling the addon
	creds := assets.NewMemoryAssetTarget([]byte{}, "/tmp/google_application_credentials.json", "0444")
	err = r.Remove(creds)
	if err != nil {
		return err
	}

	project := assets.NewMemoryAssetTarget([]byte{}, "/tmp/google_cloud_project", "0444")
	err = r.Remove(project)
	if err != nil {
		return err
	}

	return nil
}
