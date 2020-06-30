package metadata

import (
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/homedir"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
)

//EnableOrDisable enables or disables the metadata addon based on val
func EnableOrDisable(cfg *config.ClusterConfig, name, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	if enable {
		return enableAddon(cfg)
	}
	return disableAddon()

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

	return r.Copy(f)

	/*secretCmd := exec.Command("kubectl", "create", "secret", "generic", "metadata-certs", "--from-file", "key.pem=server-key.pem", "--from-file", "cert.pem=server-cert.pem", "--dry-run", "-o", "yaml")
	secretYaml, err := secretCmd.Output()
	if err != nil {
		return err
	}

	applyCmd := exec.Command("kubectl", "-n", "metadata", "apply", "-f", "-")
	reader := bytes.NewReader(secretYaml)
	applyCmd.Stdin = reader
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stdout

	return applyCmd.Run()*/
}

func disableAddon() error {
	return nil
}
