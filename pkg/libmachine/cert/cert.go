package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"time"

	"errors"

	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/log"
)

var defaultGenerator = NewX509CertGenerator()

type Options struct {
	Hosts                                     []string
	CertFile, KeyFile, CAFile, CAKeyFile, Org string
	Bits                                      int
	SwarmMaster                               bool
}

type Generator interface {
	GenerateCACertificate(certFile, keyFile, org string, bits int) error
	GenerateCert(opts *Options) error
	ReadTLSConfig(addr string, authOptions *auth.Options) (*tls.Config, error)
	ValidateCertificate(addr string, authOptions *auth.Options) (bool, error)
}

type X509CertGenerator struct{}

func NewX509CertGenerator() Generator {
	return &X509CertGenerator{}
}

func GenerateCACertificate(certFile, keyFile, org string, bits int) error {
	return defaultGenerator.GenerateCACertificate(certFile, keyFile, org, bits)
}

func GenerateCert(opts *Options) error {
	return defaultGenerator.GenerateCert(opts)
}

func ValidateCertificate(addr string, authOptions *auth.Options) (bool, error) {
	return defaultGenerator.ValidateCertificate(addr, authOptions)
}

func ReadTLSConfig(addr string, authOptions *auth.Options) (*tls.Config, error) {
	return defaultGenerator.ReadTLSConfig(addr, authOptions)
}

func SetCertGenerator(cg Generator) {
	defaultGenerator = cg
}

func (xcg *X509CertGenerator) getTLSConfig(caCert, cert, key []byte, allowInsecure bool) (*tls.Config, error) {
	// TLS config
	var tlsConfig tls.Config
	tlsConfig.InsecureSkipVerify = allowInsecure
	certPool := x509.NewCertPool()

	ok := certPool.AppendCertsFromPEM(caCert)
	if !ok {
		return &tlsConfig, errors.New("There was an error reading certificate")
	}

	tlsConfig.RootCAs = certPool
	keypair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{keypair}

	return &tlsConfig, nil
}

func (xcg *X509CertGenerator) newCertificate(org string) (*x509.Certificate, error) {
	now := time.Now()
	// need to set notBefore slightly in the past to account for time
	// skew in the VMs otherwise the certs sometimes are not yet valid
	notBefore := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()-5, 0, 0, time.Local)
	notAfter := notBefore.Add(time.Hour * 24 * 1080)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{org},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement,
		BasicConstraintsValid: true,
	}, nil

}

// GenerateCACertificate generates a new certificate authority from the specified org
// and bit size and stores the resulting certificate and key file
// in the arguments.
func (xcg *X509CertGenerator) GenerateCACertificate(certFile, keyFile, org string, bits int) error {
	template, err := xcg.newCertificate(org)
	if err != nil {
		return err
	}

	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign
	template.KeyUsage |= x509.KeyUsageKeyEncipherment
	template.KeyUsage |= x509.KeyUsageKeyAgreement

	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err

	}

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return nil
}

// GenerateCert generates a new certificate signed using the provided
// certificate authority files and stores the result in the certificate
// file and key provided.  The provided host names are set to the
// appropriate certificate fields.
func (xcg *X509CertGenerator) GenerateCert(opts *Options) error {
	template, err := xcg.newCertificate(opts.Org)
	if err != nil {
		return err
	}
	// client
	if len(opts.Hosts) == 1 && opts.Hosts[0] == "" {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		template.KeyUsage = x509.KeyUsageDigitalSignature
	} else { // server
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		if opts.SwarmMaster {
			// Extend the Swarm master's server certificate
			// permissions to also be able to connect to downstream
			// nodes as a client.
			template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
		}
		for _, h := range opts.Hosts {
			if ip := net.ParseIP(h); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, h)
			}
		}
	}

	tlsCert, err := tls.LoadX509KeyPair(opts.CAFile, opts.CAKeyFile)
	if err != nil {
		return err
	}

	priv, err := rsa.GenerateKey(rand.Reader, opts.Bits)
	if err != nil {
		return err
	}

	x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, x509Cert, &priv.PublicKey, tlsCert.PrivateKey)
	if err != nil {
		return err
	}

	certOut, err := os.Create(opts.CertFile)
	if err != nil {
		return err
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(opts.KeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return nil
}

// ReadTLSConfig reads the tls config for a machine.
func (xcg *X509CertGenerator) ReadTLSConfig(addr string, authOptions *auth.Options) (*tls.Config, error) {
	caCertPath := authOptions.CaCertPath
	clientCertPath := authOptions.ClientCertPath
	clientKeyPath := authOptions.ClientKeyPath

	log.Debugf("Reading CA certificate from %s", caCertPath)
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}

	log.Debugf("Reading client certificate from %s", clientCertPath)
	clientCert, err := ioutil.ReadFile(clientCertPath)
	if err != nil {
		return nil, err
	}

	log.Debugf("Reading client key from %s", clientKeyPath)
	clientKey, err := ioutil.ReadFile(clientKeyPath)
	if err != nil {
		return nil, err
	}

	return xcg.getTLSConfig(caCert, clientCert, clientKey, false)
}

// ValidateCertificate validate the certificate installed on the vm.
func (xcg *X509CertGenerator) ValidateCertificate(addr string, authOptions *auth.Options) (bool, error) {
	tlsConfig, err := xcg.ReadTLSConfig(addr, authOptions)
	if err != nil {
		return false, err
	}

	dialer := &net.Dialer{
		Timeout: time.Second * 20,
	}

	_, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return false, err
	}

	return true, nil
}

func CheckCertificateDate(certPath string) (bool, error) {
	log.Debugf("Reading certificate data from %s", certPath)
	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		return false, err
	}

	log.Debug("Decoding PEM data...")
	pemBlock, _ := pem.Decode(certBytes)
	if pemBlock == nil {
		return false, errors.New("Failed to decode PEM data")
	}

	log.Debug("Parsing certificate...")
	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return false, err
	}
	if time.Now().After(cert.NotAfter) {
		return false, nil
	}

	return true, nil
}
