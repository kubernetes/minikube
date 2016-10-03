package cert

import (
	"errors"
	"fmt"
	"os"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
)

func BootstrapCertificates(authOptions *auth.Options) error {
	certDir := authOptions.CertDir
	caCertPath := authOptions.CaCertPath
	caPrivateKeyPath := authOptions.CaPrivateKeyPath
	clientCertPath := authOptions.ClientCertPath
	clientKeyPath := authOptions.ClientKeyPath

	// TODO: I'm not super happy about this use of "org", the user should
	// have to specify it explicitly instead of implicitly basing it on
	// $USER.
	caOrg := mcnutils.GetUsername()
	org := caOrg + ".<bootstrap>"

	bits := 2048

	if _, err := os.Stat(certDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(certDir, 0700); err != nil {
				return fmt.Errorf("Creating machine certificate dir failed: %s", err)
			}
		} else {
			return err
		}
	}

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		log.Infof("Creating CA: %s", caCertPath)

		// check if the key path exists; if so, error
		if _, err := os.Stat(caPrivateKeyPath); err == nil {
			return errors.New("The CA key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := GenerateCACertificate(caCertPath, caPrivateKeyPath, caOrg, bits); err != nil {
			return fmt.Errorf("Generating CA certificate failed: %s", err)
		}
	}

	if _, err := os.Stat(clientCertPath); os.IsNotExist(err) {
		log.Infof("Creating client certificate: %s", clientCertPath)

		if _, err := os.Stat(certDir); err != nil {
			if os.IsNotExist(err) {
				if err := os.Mkdir(certDir, 0700); err != nil {
					return fmt.Errorf("Creating machine client cert dir failed: %s", err)
				}
			} else {
				return err
			}
		}

		// check if the key path exists; if so, error
		if _, err := os.Stat(clientKeyPath); err == nil {
			return errors.New("The client key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := GenerateCert([]string{""}, clientCertPath, clientKeyPath, caCertPath, caPrivateKeyPath, org, bits); err != nil {
			return fmt.Errorf("Generating client certificate failed: %s", err)
		}
	}

	return nil
}
