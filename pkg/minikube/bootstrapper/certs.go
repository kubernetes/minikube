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
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/lock"

	"github.com/juju/mutex"
)

var (
	certs = []string{
		"ca.crt", "ca.key", "apiserver.crt", "apiserver.key", "proxy-client-ca.crt",
		"proxy-client-ca.key", "proxy-client.crt", "proxy-client.key",
	}
)

// SetupCerts gets the generated credentials required to talk to the APIServer.
func SetupCerts(cmd command.Runner, k8s config.KubernetesConfig, n config.Node) error {

	localPath := localpath.MiniPath()
	glog.Infof("Setting up %s for IP: %s\n", localPath, n.IP)

	// WARNING: This function was not designed for multiple profiles, so it is VERY racey:
	//
	// It updates a shared certificate file and uploads it to the apiserver before launch.
	//
	// If another process updates the shared certificate, it's invalid.
	// TODO: Instead of racey manipulation of a shared certificate, use per-profile certs
	spec := lock.PathMutexSpec(filepath.Join(localPath, "certs"))
	glog.Infof("acquiring lock: %+v", spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return errors.Wrapf(err, "unable to acquire lock for %+v", spec)
	}
	defer releaser.Release()

	if err := generateCerts(k8s, n); err != nil {
		return errors.Wrap(err, "Error generating certs")
	}
	copyableFiles := []assets.CopyableFile{}
	for _, cert := range certs {
		p := filepath.Join(localPath, cert)
		perms := "0644"
		if strings.HasSuffix(cert, ".key") {
			perms = "0600"
		}
		certFile, err := assets.NewFileAsset(p, vmpath.GuestKubernetesCertsDir, cert, perms)
		if err != nil {
			return err
		}
		copyableFiles = append(copyableFiles, certFile)
	}

	caCerts, err := collectCACerts()
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
		ClusterName:          n.Name,
		ClusterServerAddress: fmt.Sprintf("https://%s", net.JoinHostPort("localhost", fmt.Sprint(n.Port))),
		ClientCertificate:    path.Join(vmpath.GuestKubernetesCertsDir, "apiserver.crt"),
		ClientKey:            path.Join(vmpath.GuestKubernetesCertsDir, "apiserver.key"),
		CertificateAuthority: path.Join(vmpath.GuestKubernetesCertsDir, "ca.crt"),
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
		return errors.Wrapf(err, "Configuring CA certs")
	}
	return nil
}

func generateCerts(k8s config.KubernetesConfig, n config.Node) error {
	serviceIP, err := util.GetServiceClusterIP(k8s.ServiceCIDR)
	if err != nil {
		return errors.Wrap(err, "getting service cluster ip")
	}

	localPath := localpath.MiniPath()
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
		[]net.IP{net.ParseIP(n.IP), serviceIP, net.ParseIP(kic.DefaultBindIPV4), net.ParseIP("10.0.0.1")}...)
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

	for _, caCertSpec := range caCertSpecs {
		if !(canReadFile(caCertSpec.certPath) &&
			canReadFile(caCertSpec.keyPath)) {
			if err := util.GenerateCACert(
				caCertSpec.certPath, caCertSpec.keyPath, caCertSpec.subject,
			); err != nil {
				return errors.Wrap(err, "Error generating CA certificate")
			}
		}
	}

	for _, signedCertSpec := range signedCertSpecs {
		if err := util.GenerateSignedCert(
			signedCertSpec.certPath, signedCertSpec.keyPath, signedCertSpec.subject,
			signedCertSpec.ips, signedCertSpec.alternateNames,
			signedCertSpec.caCertPath, signedCertSpec.caKeyPath,
		); err != nil {
			return errors.Wrap(err, "Error generating signed apiserver serving cert")
		}
	}

	return nil
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

// collectCACerts looks up all PEM certificates with .crt or .pem extension in ~/.minikube/certs to copy to the host.
// minikube root CA is also included but libmachine certificates (ca.pem/cert.pem) are excluded.
func collectCACerts() (map[string]string, error) {
	localPath := localpath.MiniPath()
	certFiles := map[string]string{}

	certsDir := filepath.Join(localPath, "certs")
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
					certFiles[hostpath] = path.Join(vmpath.GuestCertAuthDir, dst)
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
	certFiles[filepath.Join(localPath, "ca.crt")] = path.Join(vmpath.GuestCertAuthDir, "minikubeCA.pem")

	filtered := map[string]string{}
	for k, v := range certFiles {
		if v != "" {
			filtered[k] = v
		}
	}
	return filtered, nil
}

// getSubjectHash calculates Certificate Subject Hash for creating certificate symlinks
func getSubjectHash(cr command.Runner, filePath string) (string, error) {
	rr, err := cr.RunCmd(exec.Command("openssl", "x509", "-hash", "-noout", "-in", filePath))
	if err != nil {
		return "", errors.Wrapf(err, rr.Command())
	}
	stringHash := strings.TrimSpace(rr.Stdout.String())
	return stringHash, nil
}

// configureCACerts looks up and installs all uploaded PEM certificates in /usr/share/ca-certificates to system-wide certificate store (/etc/ssl/certs).
// OpenSSL binary required in minikube ISO
func configureCACerts(cr command.Runner, caCerts map[string]string) error {
	hasSSLBinary := true
	_, err := cr.RunCmd(exec.Command("openssl", "version"))
	if err != nil {
		hasSSLBinary = false
	}

	if !hasSSLBinary && len(caCerts) > 0 {
		glog.Warning("OpenSSL not found. Please recreate the cluster with the latest minikube ISO.")
	}

	for _, caCertFile := range caCerts {
		dstFilename := path.Base(caCertFile)
		certStorePath := path.Join(vmpath.GuestCertStoreDir, dstFilename)
		cmd := fmt.Sprintf("test -f %s || ln -fs %s %s", caCertFile, certStorePath, caCertFile)
		if _, err := cr.RunCmd(exec.Command("sudo", "/bin/bash", "-c", cmd)); err != nil {
			return errors.Wrapf(err, "create symlink for %s", caCertFile)
		}
		if hasSSLBinary {
			subjectHash, err := getSubjectHash(cr, caCertFile)
			if err != nil {
				return errors.Wrapf(err, "calculate hash for cacert %s", caCertFile)
			}
			subjectHashLink := path.Join(vmpath.GuestCertStoreDir, fmt.Sprintf("%s.0", subjectHash))

			// NOTE: This symlink may exist, but point to a missing file
			cmd := fmt.Sprintf("test -L %s || ln -fs %s %s", subjectHashLink, certStorePath, subjectHashLink)
			if _, err := cr.RunCmd(exec.Command("sudo", "/bin/bash", "-c", cmd)); err != nil {
				return errors.Wrapf(err, "create symlink for %s", caCertFile)
			}
		}
	}
	return nil
}

// canReadFile returns true if the file represented
// by path exists and is readable, otherwise false.
func canReadFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	return true
}
