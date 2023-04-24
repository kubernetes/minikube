package detector

import (
	"fmt"
	"os/exec"

	"github.com/docker/machine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/provision"
)

var (
	provisioners          = make(map[string]*RegisteredProvisioner)
	detector     Detector = &StandardDetector{}
)

const (
	LastReleaseBeforeCEVersioning = "1.13.1"
)

// RegisteredProvisioner creates a new provisioner
type RegisteredProvisioner struct {
	New func(d drivers.Driver) provision.Provisioner
}

type Detector interface {
	DetectProvisioner(d drivers.Driver) (provision.Provisioner, error)
}

type StandardDetector struct{}

func DetectProvisioner(d drivers.Driver) (provision.Provisioner, error) {
	return detector.DetectProvisioner(d)
}

func (detector StandardDetector) DetectProvisioner(d drivers.Driver) (provision.Provisioner, error) {
	log.Info("Waiting for prompt to be available...")
	if err := drivers.WaitForPrompt(d); err != nil {
		return nil, err
	}

	log.Info("Detecting the provisioner...")

	rnr, err := d.GetRunner()
	if err != nil {
		return nil, fmt.Errorf("Error getting cmdRunner: %s", err)
	}

	osReleaseOut, err := rnr.RunCmd(exec.Command("cat /etc/os-release"))
	if err != nil {
		return nil, fmt.Errorf("Error getting SSH command: %s", err)
	}

	osReleaseInfo, err := provision.NewOsRelease(osReleaseOut.Stdout.Bytes())
	if err != nil {
		return nil, fmt.Errorf("Error parsing /etc/os-release file: %s", err)
	}

	// TODO:
	// this is pretty ugly...
	for _, p := range provisioners {
		provisioner := p.New(d)
		provisioner.SetOsReleaseInfo(osReleaseInfo)

		if provisioner.CompatibleWithMachine() {
			log.Debugf("found compatible host: %s", osReleaseInfo.ID)
			return provisioner, nil
		}
	}

	return nil, provision.ErrDetectionFailed
}
