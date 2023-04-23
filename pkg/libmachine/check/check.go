package check

import (
	"errors"
	"fmt"
	"net/url"

	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cert"
	"k8s.io/minikube/pkg/libmachine/machine"
)

var (
	DefaultConnChecker ConnChecker
	ErrSwarmNotStarted = errors.New("Connection to Swarm cannot be checked but the certs are valid. Maybe swarm is not started")
)

func init() {
	DefaultConnChecker = &MachineConnChecker{}
}

// ErrCertInvalid for when the cert is computed to be invalid.
type ErrCertInvalid struct {
	wrappedErr error
	hostURL    string
}

func (e ErrCertInvalid) Error() string {
	return fmt.Sprintf(`There was an error validating certificates for host %q: %s
You can attempt to regenerate them using 'docker-machine regenerate-certs [name]'.
Be advised that this will trigger a Docker daemon restart which might stop running containers.
`, e.hostURL, e.wrappedErr)
}

type ConnChecker interface {
	Check(*machine.Machine, bool) (dockerHost string, authOptions *auth.Options, err error)
}

type MachineConnChecker struct{}

func (mcc *MachineConnChecker) Check(h *machine.Machine, swarm bool) (string, *auth.Options, error) {
	dockerHost, err := h.Driver.GetURL()
	if err != nil {
		return "", &auth.Options{}, err
	}

	dockerURL := dockerHost

	u, err := url.Parse(dockerURL)
	if err != nil {
		return "", &auth.Options{}, fmt.Errorf("Error parsing URL: %s", err)
	}

	authOptions := h.AuthOptions()

	if err := checkCert(u.Host, authOptions); err != nil {
		if swarm {
			// Connection to the swarm port cannot be checked. Maybe it's just the swarm containers that are down
			// TODO: check the containers and restart them
			// Let's check the non-swarm connection to give a better error message to the user.
			if _, _, err := mcc.Check(h, false); err == nil {
				return "", &auth.Options{}, ErrSwarmNotStarted
			}
		}

		return "", &auth.Options{}, fmt.Errorf("Error checking and/or regenerating the certs: %s", err)
	}

	return dockerURL, authOptions, nil
}

func checkCert(hostURL string, authOptions *auth.Options) error {
	valid, err := cert.ValidateCertificate(hostURL, authOptions)
	if !valid || err != nil {
		return ErrCertInvalid{
			wrappedErr: err,
			hostURL:    hostURL,
		}
	}

	return nil
}
