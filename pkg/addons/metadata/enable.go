package metadata

import (
	"path/filepath"

	"k8s.io/client-go/util/homedir"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
)

func EnableAddon(cfg *config.ClusterConfig, name string, val string) error {
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

	return r.Copy(f)

}

func PatchCABundle(cfg *config.ClusterConfig, name string, val string) error {

}
