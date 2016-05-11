/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package localkube

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"path"

	utilcrypto "k8s.io/kubernetes/pkg/util"
	utilnet "k8s.io/kubernetes/pkg/util/net"

	"k8s.io/minikube/pkg/util"
)

const serverInterval = 200

// LocalkubeServer provides a fully functional Kubernetes cluster running entirely through goroutines
type LocalkubeServer struct {
	// Inherits Servers
	Servers

	// Options
	Containerized            bool
	EnableDNS                bool
	DNSDomain                string
	DNSIP                    net.IP
	LocalkubeDirectory       string
	ServiceClusterIPRange    net.IPNet
	APIServerAddress         net.IP
	APIServerPort            int
	APIServerInsecureAddress net.IP
	APIServerInsecurePort    int
}

func (lk *LocalkubeServer) AddServer(server Server) {
	lk.Servers = append(lk.Servers, server)
}

func (lk LocalkubeServer) GetEtcdDataDirectory() string {
	return path.Join(lk.LocalkubeDirectory, "etcd")
}

func (lk LocalkubeServer) GetDNSDataDirectory() string {
	return path.Join(lk.LocalkubeDirectory, "dns")
}

func (lk LocalkubeServer) GetCertificateDirectory() string {
	return path.Join(lk.LocalkubeDirectory, "certs")
}
func (lk LocalkubeServer) GetPrivateKeyCertPath() string {
	return path.Join(lk.GetCertificateDirectory(), "apiserver.key")
}
func (lk LocalkubeServer) GetPublicKeyCertPath() string {
	return path.Join(lk.GetCertificateDirectory(), "apiserver.crt")
}

func (lk LocalkubeServer) GetAPIServerSecureURL() string {
	return fmt.Sprintf("https://%s:%d", lk.APIServerAddress.String(), lk.APIServerPort)
}

func (lk LocalkubeServer) GetAPIServerInsecureURL() string {
	return fmt.Sprintf("http://%s:%d", lk.APIServerInsecureAddress.String(), lk.APIServerInsecurePort)
}

// Get the host's public IP address
func (lk LocalkubeServer) GetHostIP() (net.IP, error) {
	return utilnet.ChooseBindAddress(net.ParseIP("0.0.0.0"))
}

func (lk LocalkubeServer) loadCert(path string) (*x509.Certificate, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoded, _ := pem.Decode(contents)
	if decoded == nil {
		return nil, fmt.Errorf("Unable to decode certificate.")
	}

	return x509.ParseCertificate(decoded.Bytes)
}

func (lk LocalkubeServer) shouldGenerateCerts(hostIP net.IP) bool {
	if !(util.CanReadFile(lk.GetPublicKeyCertPath()) &&
		util.CanReadFile(lk.GetPrivateKeyCertPath())) {
		fmt.Println("Regenerating certs because the files aren't readable.")
		return true
	}

	cert, err := lk.loadCert(lk.GetPublicKeyCertPath())
	if err != nil {
		fmt.Println("Regenerating certs because there was an error loading the certificate: ", err)
		return true
	}

	for _, certIP := range cert.IPAddresses {
		if certIP.Equal(hostIP) {
			return false
		}
	}
	fmt.Printf(
		"Regenerating certs because the IP didn't match. Got %s, expected %s",
		cert.IPAddresses, hostIP)
	return true
}

func (lk LocalkubeServer) GenerateCerts(hostIP net.IP) error {

	if !lk.shouldGenerateCerts(hostIP) {
		fmt.Println("Using these existing certs: ", lk.GetPublicKeyCertPath(), lk.GetPrivateKeyCertPath())
		return nil
	}

	alternateIPs := []net.IP{lk.ServiceClusterIPRange.IP}
	alternateDNS := []string{fmt.Sprintf("%s.%s", "kubernetes.default.svc", lk.DNSDomain), "kubernetes.default.svc", "kubernetes.default", "kubernetes"}

	if err := utilcrypto.GenerateSelfSignedCert(hostIP.String(), lk.GetPublicKeyCertPath(), lk.GetPrivateKeyCertPath(), alternateIPs, alternateDNS); err != nil {
		fmt.Println("Failed to create certs: ", err)
		return err
	}

	return nil
}
