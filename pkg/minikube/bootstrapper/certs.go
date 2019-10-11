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

package bootstrapper

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"

	"github.com/juju/clock"
	"github.com/juju/mutex"
)

const (
	// guestCACertsDir contains CA certificates
	guestCACertsDir = "/usr/share/ca-certificates"
	// guestSSLCertsDir contains SSL certificates
	guestSSLCertsDir = "/etc/ssl/certs"
)

// SetupCerts gets the generated credentials required to talk to the APIServer.
func SetupCerts(profile string, cmd command.Runner, k8s config.KubernetesConfig) error {
	spec := mutex.Spec{
		Name:  "setupCerts",
		Clock: clock.WallClock,
		Delay: 15 * time.Second,
	}
	glog.Infof("acquiring lock: %+v", spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return errors.Wrapf(err, "unable to acquire lock for %+v", spec)
	}
	defer releaser.Release()

	glog.Infof("Setting up %s for IP: %s\n", profile, k8s.NodeIP)

	certs, err := generateCerts(profile, k8s)
	if err != nil {
		return errors.Wrap(err, "Error generating certs")
	}
	copyableFiles := []assets.CopyableFile{}
	for _, cert := range certs {
		perms := "0644"
		if strings.HasSuffix(cert, ".key") {
			perms = "0600"
		}
		certFile, err := assets.NewFileAsset(cert, vmpath.GuestCertsDir, filepath.Base(cert), perms)
		if err != nil {
			return err
		}
		copyableFiles = append(copyableFiles, certFile)
	}
	caCerts, err := collectCACerts(profile)
	if err != nil {
		return err
	}
	for src, dst := range caCerts {
		certFile, err := assets.NewFileAsset(src, path.Dir(dst), path.Base(dst), "0644")
		if err != nil {
			return err
		}

		copyableFiles = append(copyableFiles, certFile)
	}
	kcs := &kubeconfig.Settings{
		ClusterName:          k8s.NodeName,
		ClusterServerAddress: fmt.Sprintf("https://localhost:%d", k8s.NodePort),
		ClientCertificate:    path.Join(vmpath.GuestCertsDir, "apiserver.crt"),
		ClientKey:            path.Join(vmpath.GuestCertsDir, "apiserver.key"),
		CertificateAuthority: path.Join(vmpath.GuestCertsDir, "ca.crt"),
		KeepContext:          false,
	}

	kubeCfg := api.NewConfig()
	err = kubeconfig.PopulateFromSettings(kcs, kubeCfg)
	if err != nil {
		return errors.Wrap(err, "populating kubeconfig")
	}
	data, err := runtime.Encode(latest.Codec, kubeCfg)
	if err != nil {
		return errors.Wrap(err, "encoding kubeconfig")
	}

	kubeCfgFile := assets.NewMemoryAsset(data, vmpath.GuestPersistentDir, "kubeconfig", "0644")
	copyableFiles = append(copyableFiles, kubeCfgFile)

	for _, f := range copyableFiles {
		if err := cmd.Copy(f); err != nil {
			return errors.Wrapf(err, "Copy %s", f.GetAssetName())
		}
	}

	// configure CA certificates
	if err := configureCACerts(cmd, caCerts); err != nil {
		return errors.Wrapf(err, "error configuring CA certificates during provisioning %v", err)
	}
	return nil
}

func generateCerts(profile string, k8s config.KubernetesConfig) ([]string, error) {
	serviceIP, err := util.GetServiceClusterIP(k8s.ServiceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "getting service cluster ip")
	}

	localPath := localpath.KubernetesCerts(profile)
	caCertPath := filepath.Join(localPath, "ca.crt")
	caKeyPath := filepath.Join(localPath, "ca.key")

	proxyClientCACertPath := filepath.Join(localPath, "proxy-client-ca.crt")
	proxyClientCAKeyPath := filepath.Join(localPath, "proxy-client-ca.key")

	caCertSpecs := []struct {
		certPath string
		keyPath  string
		subject  string
	}{
		{ // client / apiserver CA
			certPath: caCertPath,
			keyPath:  caKeyPath,
			subject:  "minikubeCA",
		},
		{ // proxy-client CA
			certPath: proxyClientCACertPath,
			keyPath:  proxyClientCAKeyPath,
			subject:  "proxyClientCA",
		},
	}

	apiServerIPs := append(
		k8s.APIServerIPs,
		[]net.IP{net.ParseIP(k8s.NodeIP), serviceIP, net.ParseIP("10.0.0.1")}...)
	apiServerNames := append(k8s.APIServerNames, k8s.APIServerName)
	apiServerAlternateNames := append(
		apiServerNames,
		util.GetAlternateDNS(k8s.DNSDomain)...)

	signedCertSpecs := []struct {
		certPath       string
		keyPath        string
		subject        string
		ips            []net.IP
		alternateNames []string
		caCertPath     string
		caKeyPath      string
	}{
		{ // Client cert
			certPath:       filepath.Join(localPath, "client.crt"),
			keyPath:        filepath.Join(localPath, "client.key"),
			subject:        "minikube-user",
			ips:            []net.IP{},
			alternateNames: []string{},
			caCertPath:     caCertPath,
			caKeyPath:      caKeyPath,
		},
		{ // apiserver serving cert
			certPath:       filepath.Join(localPath, "apiserver.crt"),
			keyPath:        filepath.Join(localPath, "apiserver.key"),
			subject:        "minikube",
			ips:            apiServerIPs,
			alternateNames: apiServerAlternateNames,
			caCertPath:     caCertPath,
			caKeyPath:      caKeyPath,
		},
		{ // aggregator proxy-client cert
			certPath:       filepath.Join(localPath, "proxy-client.crt"),
			keyPath:        filepath.Join(localPath, "proxy-client.key"),
			subject:        "aggregator",
			ips:            []net.IP{},
			alternateNames: []string{},
			caCertPath:     proxyClientCACertPath,
			caKeyPath:      proxyClientCAKeyPath,
		},
	}

	generated := []string{}
	for _, ca := range caCertSpecs {
		if !(util.CanReadFile(ca.certPath) && util.CanReadFile(ca.keyPath)) {
			if err := util.GenerateCACert(ca.certPath, ca.keyPath, ca.subject); err != nil {
				return generated, errors.Wrap(err, "Error generating CA certificate")
			}
			generated = append(generated, ca.certPath, ca.keyPath)
		}
	}

	for _, sc := range signedCertSpecs {
		if err := util.GenerateSignedCert(sc.certPath, sc.keyPath, sc.subject, sc.ips, sc.alternateNames, sc.caCertPath, sc.caKeyPath); err != nil {
			return generated, errors.Wrap(err, "Error generating signed apiserver serving cert")
		}
		generated = append(generated, sc.certPath, sc.keyPath)
	}

	return generated, nil
}

// isValidPEMCertificate checks whether the input file is a valid PEM certificate (with at least one CERTIFICATE block)
func isValidPEMCertificate(filePath string) (bool, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	for {
		block, rest := pem.Decode(fileBytes)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			// certificate found
			return true, nil
		}
		fileBytes = rest
	}

	return false, nil
}

// collectStoreCerts copies .crt and .pem certificates from the libmachine store
func collectCACerts(profile string) (map[string]string, error) {
	certFiles := map[string]string{}
	certsDir := localpath.MachineCerts(profile)
	err := filepath.Walk(certsDir, func(hostpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info != nil && !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(hostpath))
			if ext == ".crt" || ext == ".pem" {
				validPem, err := isValidPEMCertificate(hostpath)
				if err != nil {
					return err
				}
				if validPem {
					filename := filepath.Base(hostpath)
					dst := fmt.Sprintf("%s.%s", strings.TrimSuffix(filename, ext), "pem")
					certFiles[hostpath] = path.Join(guestCACertsDir, dst)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "provisioning: traversal certificates dir %s", certsDir)
	}

	for _, excluded := range []string{"ca.pem", "cert.pem"} {
		certFiles[filepath.Join(certsDir, excluded)] = ""
	}

	// populates minikube CA
	certFiles[filepath.Join(certsDir, "ca.crt")] = path.Join(guestCACertsDir, "minikubeCA.pem")

	filtered := map[string]string{}
	for k, v := range certFiles {
		if v != "" {
			filtered[k] = v
		}
	}
	return filtered, nil
}

// getSubjectHash calculates Certificate Subject Hash for creating certificate symlinks
func getSubjectHash(cmd command.Runner, filePath string) (string, error) {
	out, err := cmd.CombinedOutput(fmt.Sprintf("openssl x509 -hash -noout -in '%s'", filePath))
	if err != nil {
		return "", err
	}

	stringHash := strings.TrimSpace(out)
	return stringHash, nil
}

// configureCACerts looks up and installs all uploaded PEM certificates in /usr/share/ca-certificates to system-wide certificate store (/etc/ssl/certs).
// OpenSSL binary required in minikube ISO
func configureCACerts(cmd command.Runner, caCerts map[string]string) error {
	hasSSLBinary := true
	if err := cmd.Run("which openssl"); err != nil {
		hasSSLBinary = false
	}

	if !hasSSLBinary && len(caCerts) > 0 {
		glog.Warning("OpenSSL not found. Please recreate the cluster with the latest minikube ISO.")
	}

	for _, caCertFile := range caCerts {
		dstFilename := path.Base(caCertFile)
		certStorePath := path.Join(guestSSLCertsDir, dstFilename)
		if err := cmd.Run(fmt.Sprintf("sudo test -f '%s'", certStorePath)); err != nil {
			if err := cmd.Run(fmt.Sprintf("sudo ln -s '%s' '%s'", caCertFile, certStorePath)); err != nil {
				return errors.Wrapf(err, "error making symbol link for certificate %s", caCertFile)
			}
		}
		if hasSSLBinary {
			subjectHash, err := getSubjectHash(cmd, caCertFile)
			if err != nil {
				return errors.Wrapf(err, "error calculating subject hash for certificate %s", caCertFile)
			}
			subjectHashLink := path.Join(guestSSLCertsDir, fmt.Sprintf("%s.0", subjectHash))
			if err := cmd.Run(fmt.Sprintf("sudo test -f '%s'", subjectHashLink)); err != nil {
				if err := cmd.Run(fmt.Sprintf("sudo ln -s '%s' '%s'", certStorePath, subjectHashLink)); err != nil {
					return errors.Wrapf(err, "error making subject hash symbol link for certificate %s", caCertFile)
				}
			}
		}
	}

	return nil
}
