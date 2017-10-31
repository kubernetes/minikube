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
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"strconv"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apiserver/pkg/util/flag"
	"k8s.io/minikube/pkg/util/kubeconfig"

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
	LocalkubeDirectory       string
	ServiceClusterIPRange    net.IPNet
	APIServerAddress         net.IP
	APIServerPort            int
	APIServerInsecureAddress net.IP
	APIServerInsecurePort    int
	APIServerName            string
	ShouldGenerateCerts      bool
	ShouldGenerateKubeconfig bool
	ShowVersion              bool
	ShowHostIP               bool
	RuntimeConfig            flag.ConfigurationMap
	NodeIP                   net.IP
	ContainerRuntime         string
	RemoteRuntimeEndpoint    string
	RemoteImageEndpoint      string
	NetworkPlugin            string
	FeatureGates             string
	ExtraConfig              util.ExtraOptionSlice
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
func (lk LocalkubeServer) GetCAPrivateKeyCertPath() string {
	return path.Join(lk.GetCertificateDirectory(), "ca.key")
}
func (lk LocalkubeServer) GetCAPublicKeyCertPath() string {
	return path.Join(lk.GetCertificateDirectory(), "ca.crt")
}

func (lk LocalkubeServer) GetProxyClientPrivateKeyCertPath() string {
	return path.Join(lk.GetCertificateDirectory(), "proxy-client.key")
}
func (lk LocalkubeServer) GetProxyClientPublicKeyCertPath() string {
	return path.Join(lk.GetCertificateDirectory(), "proxy-client.crt")
}
func (lk LocalkubeServer) GetProxyClientCAPublicKeyCertPath() string {
	return path.Join(lk.GetCertificateDirectory(), "proxy-client-ca.crt")
}
func (lk LocalkubeServer) GetProxyClientCAPrivateKeyCertPath() string {
	return path.Join(lk.GetCertificateDirectory(), "proxy-client-ca.key")
}

func (lk LocalkubeServer) GetAPIServerSecureURL() string {
	return fmt.Sprintf("https://%s:%d", lk.APIServerAddress.String(), lk.APIServerPort)
}

func (lk LocalkubeServer) GetAPIServerInsecureURL() string {
	if lk.APIServerInsecurePort != 0 {
		return fmt.Sprintf("http://%s:%d", lk.APIServerInsecureAddress.String(), lk.APIServerInsecurePort)
	}
	return ""
}

func (lk LocalkubeServer) GetAPIServerProtocol() string {
	if lk.APIServerInsecurePort != 0 {
		return "http://"
	}
	return "https://"
}

func (lk LocalkubeServer) GetTransport() (*http.Transport, error) {
	if lk.APIServerInsecurePort != 0 {
		return &http.Transport{}, nil
	}
	cert, err := tls.LoadX509KeyPair(lk.GetPublicKeyCertPath(), lk.GetPrivateKeyCertPath())
	if err != nil {
		glog.Error(err)
		return &http.Transport{}, err
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(lk.GetCAPublicKeyCertPath())
	if err != nil {
		glog.Warning(err)
		return &http.Transport{}, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	return &http.Transport{TLSClientConfig: tlsConfig}, nil
}

// Get the host's public IP address
func (lk LocalkubeServer) GetHostIP() (net.IP, error) {
	return utilnet.ChooseBindAddress(net.ParseIP("0.0.0.0"))
}

func (lk LocalkubeServer) getExtraConfigForComponent(component string) []util.ExtraOption {
	e := []util.ExtraOption{}
	for _, c := range lk.ExtraConfig {
		if c.Component == component {
			e = append(e, c)
		}
	}
	return e
}

func (lk LocalkubeServer) SetExtraConfigForComponent(component string, config interface{}) {
	extra := lk.getExtraConfigForComponent(component)
	for _, e := range extra {
		glog.Infof("Setting %s to %s on %s.\n", e.Key, e.Value, component)
		if err := util.FindAndSet(e.Key, config, e.Value); err != nil {
			glog.Warningf("Unable to set %s to %s. Error: %s", e.Key, e.Value, err)
		}
	}
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

func (lk LocalkubeServer) shouldGenerateCerts(ips []net.IP) bool {
	if !(util.CanReadFile(lk.GetPublicKeyCertPath()) &&
		util.CanReadFile(lk.GetPrivateKeyCertPath())) {
		fmt.Println("Regenerating certs because the files aren't readable")
		return true
	}

	cert, err := lk.loadCert(lk.GetPublicKeyCertPath())
	if err != nil {
		fmt.Println("Regenerating certs because there was an error loading the certificate: ", err)
		return true
	}

	certIPs := map[string]bool{}
	for _, certIP := range cert.IPAddresses {
		certIPs[certIP.String()] = true
	}

	for _, ip := range ips {
		if _, ok := certIPs[ip.String()]; !ok {
			fmt.Println("Regenerating certs becase an IP is missing: ", ip)
			return true
		}
	}
	return false
}

func (lk LocalkubeServer) shouldGenerateCACerts() bool {
	if !(util.CanReadFile(lk.GetCAPublicKeyCertPath()) &&
		util.CanReadFile(lk.GetCAPrivateKeyCertPath())) {
		fmt.Println("Regenerating CA certs because the files aren't readable")
		return true
	}

	_, err := lk.loadCert(lk.GetCAPublicKeyCertPath())
	if err != nil {
		fmt.Println("Regenerating CA certs because there was an error loading the certificate: ", err)
		return true
	}

	return false
}

func (lk LocalkubeServer) GenerateKubeconfig() error {
	if !lk.ShouldGenerateKubeconfig {
		return nil
	}

	// setup kubeconfig
	kubeConfigFile := util.DefaultKubeConfigPath
	glog.Infof("Setting up kubeconfig at: %s", kubeConfigFile)
	kubeHost := "http://127.0.0.1:" + strconv.Itoa(lk.APIServerInsecurePort)

	//TODO(aaron-prindle) configure this so that it can generate secure certs as well
	kubeCfgSetup := &kubeconfig.KubeConfigSetup{
		ClusterName:          lk.APIServerName,
		ClusterServerAddress: kubeHost,
		KeepContext:          false,
	}

	kubeCfgSetup.SetKubeConfigFile(kubeConfigFile)

	if err := kubeconfig.SetupKubeConfig(kubeCfgSetup); err != nil {
		glog.Errorln("Error setting up kubeconfig: ", err)
		return err
	}

	return nil
}

func (lk LocalkubeServer) getAllIPs() ([]net.IP, error) {
	serviceIP, err := util.GetServiceClusterIP(lk.ServiceClusterIPRange.String())
	if err != nil {
		return nil, errors.Wrap(err, "getting service cluster ip")
	}
	ips := []net.IP{serviceIP}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok {
			fmt.Println("Skipping: ", addr)
			continue
		}
		ips = append(ips, ipnet.IP)
	}
	return ips, nil
}

func (lk LocalkubeServer) GenerateCerts() error {
	if !lk.shouldGenerateCACerts() {
		fmt.Println(
			"Using these existing CA certs: ", lk.GetCAPublicKeyCertPath(),
			lk.GetCAPrivateKeyCertPath(), lk.GetProxyClientCAPublicKeyCertPath(),
			lk.GetProxyClientCAPrivateKeyCertPath(),
		)
	} else {
		fmt.Println("Creating CA cert")
		if err := util.GenerateCACert(
			lk.GetCAPublicKeyCertPath(), lk.GetCAPrivateKeyCertPath(),
			lk.APIServerName,
		); err != nil {
			fmt.Println("Failed to create CA cert: ", err)
			return err
		}
		fmt.Println("Creating proxy client CA cert")
		if err := util.GenerateCACert(
			lk.GetProxyClientCAPublicKeyCertPath(),
			lk.GetProxyClientCAPrivateKeyCertPath(), "proxyClientCA",
		); err != nil {
			fmt.Println("Failed to create proxy client CA cert: ", err)
			return err
		}
	}

	ips, err := lk.getAllIPs()
	if err != nil {
		return err
	}

	if !lk.shouldGenerateCerts(ips) {
		fmt.Println(
			"Using these existing certs: ", lk.GetPublicKeyCertPath(),
			lk.GetPrivateKeyCertPath(), lk.GetProxyClientPublicKeyCertPath(),
			lk.GetProxyClientPrivateKeyCertPath(),
		)
		return nil
	}
	fmt.Println("Creating cert with IPs: ", ips)

	if err := util.GenerateSignedCert(
		lk.GetPublicKeyCertPath(), lk.GetPrivateKeyCertPath(), "minikube", ips,
		util.GetAlternateDNS(lk.DNSDomain), lk.GetCAPublicKeyCertPath(),
		lk.GetCAPrivateKeyCertPath(),
	); err != nil {
		fmt.Println("Failed to create cert: ", err)
		return err
	}

	if err := util.GenerateSignedCert(
		lk.GetProxyClientPublicKeyCertPath(), lk.GetProxyClientPrivateKeyCertPath(),
		"aggregator", []net.IP{}, []string{},
		lk.GetProxyClientCAPublicKeyCertPath(),
		lk.GetProxyClientCAPrivateKeyCertPath(),
	); err != nil {
		fmt.Println("Failed to create proxy client cert: ", err)
		return err
	}

	return nil
}
