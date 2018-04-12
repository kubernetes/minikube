package cert

import (
	"errors"
	"fmt"
	"os"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
)

func createCACert(authOptions *auth.Options, caOrg string, bits int) error {
	caCertPath := authOptions.CaCertPath
	caPrivateKeyPath := authOptions.CaPrivateKeyPath

	log.Infof("Creating CA: %s", caCertPath)

	// check if the key path exists; if so, error
	if _, err := os.Stat(caPrivateKeyPath); err == nil {
		return errors.New("certificate authority key already exists")
	}

	if err := GenerateCACertificate(caCertPath, caPrivateKeyPath, caOrg, bits); err != nil {
		return fmt.Errorf("generating CA certificate failed: %s", err)
	}

	return nil
}

func createCert(authOptions *auth.Options, org string, bits int) error {
	certDir := authOptions.CertDir
	caCertPath := authOptions.CaCertPath
	caPrivateKeyPath := authOptions.CaPrivateKeyPath
	clientCertPath := authOptions.ClientCertPath
	clientKeyPath := authOptions.ClientKeyPath

	log.Infof("Creating client certificate: %s", clientCertPath)

	if _, err := os.Stat(certDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(certDir, 0700); err != nil {
				return fmt.Errorf("failure creating machine client cert dir: %s", err)
			}
		} else {
			return err
		}
	}

	// check if the key path exists; if so, error
	if _, err := os.Stat(clientKeyPath); err == nil {
		return errors.New("client key already exists")
	}

	// Used to generate the client certificate.
	certOptions := &Options{
		Hosts:       []string{""},
		CertFile:    clientCertPath,
		KeyFile:     clientKeyPath,
		CAFile:      caCertPath,
		CAKeyFile:   caPrivateKeyPath,
		Org:         org,
		Bits:        bits,
		SwarmMaster: false,
	}

	if err := GenerateCert(certOptions); err != nil {
		return fmt.Errorf("failure generating client certificate: %s", err)
	}

	return nil
}

func BootstrapCertificates(authOptions *auth.Options) error {
	certDir := authOptions.CertDir
	caCertPath := authOptions.CaCertPath
	clientCertPath := authOptions.ClientCertPath
	clientKeyPath := authOptions.ClientKeyPath
	caPrivateKeyPath := authOptions.CaPrivateKeyPath

	// TODO: I'm not super happy about this use of "org", the user should
	// have to specify it explicitly instead of implicitly basing it on
	// $USER.
	caOrg := mcnutils.GetUsername()
	org := caOrg + ".<bootstrap>"

	bits := 2048

	if _, err := os.Stat(certDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(certDir, 0700); err != nil {
				return fmt.Errorf("creating machine certificate dir failed: %s", err)
			}
		} else {
			return err
		}
	}

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		if err := createCACert(authOptions, caOrg, bits); err != nil {
			return err
		}
	} else {
		current, err := CheckCertificateDate(caCertPath)
		if err != nil {
			return err
		}
		if !current {
			log.Info("CA certificate is outdated and needs to be regenerated")
			os.Remove(caPrivateKeyPath)
			if err := createCACert(authOptions, caOrg, bits); err != nil {
				return err
			}
		}
	}

	if _, err := os.Stat(clientCertPath); os.IsNotExist(err) {
		if err := createCert(authOptions, org, bits); err != nil {
			return err
		}
	} else {
		current, err := CheckCertificateDate(clientCertPath)
		if err != nil {
			return err
		}
		if !current {
			log.Info("Client certificate is outdated and needs to be regenerated")
			os.Remove(clientKeyPath)
			if err := createCert(authOptions, org, bits); err != nil {
				return err
			}
		}
	}

	return nil
}
