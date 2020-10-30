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
	"crypto/sha1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
)

// SetupCerts gets the generated credentials required to talk to the APIServer.
func SetupCerts(cmd command.Runner, k8s config.KubernetesConfig, n config.Node) ([]assets.CopyableFile, error) {
	localPath := localpath.Profile(k8s.ClusterName)
	klog.Infof("Setting up %s for IP: %s\n", localPath, n.IP)

	ccs, err := generateSharedCACerts()
	if err != nil {
		return nil, errors.Wrap(err, "shared CA certs")
	}

	xfer, err := generateProfileCerts(k8s, n, ccs)
	if err != nil {
		return nil, errors.Wrap(err, "profile certs")
	}

	xfer = append(xfer, ccs.caCert)
	xfer = append(xfer, ccs.caKey)
	xfer = append(xfer, ccs.proxyCert)
	xfer = append(xfer, ccs.proxyKey)

	copyableFiles := []assets.CopyableFile{}
	for _, p := range xfer {
		cert := filepath.Base(p)
		perms := "0644"
		if strings.HasSuffix(cert, ".key") {
			perms = "0600"
		}
		certFile, err := assets.NewFileAsset(p, vmpath.GuestKubernetesCertsDir, cert, perms)
		if err != nil {
			return nil, errors.Wrapf(err, "key asset %s", cert)
		}
		copyableFiles = append(copyableFiles, certFile)
	}

	caCerts, err := collectCACerts()
	if err != nil {
		return nil, err
	}
	for src, dst := range caCerts {
		certFile, err := assets.NewFileAsset(src, path.Dir(dst), path.Base(dst), "0644")
		if err != nil {
			return nil, errors.Wrapf(err, "ca asset %s", src)
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
		return nil, errors.Wrap(err, "populating kubeconfig")
	}
	data, err := runtime.Encode(latest.Codec, kubeCfg)
	if err != nil {
		return nil, errors.Wrap(err, "encoding kubeconfig")
	}

	if n.ControlPlane {
		kubeCfgFile := assets.NewMemoryAsset(data, vmpath.GuestPersistentDir, "kubeconfig", "0644")
		copyableFiles = append(copyableFiles, kubeCfgFile)
	}

	for _, f := range copyableFiles {
		if err := cmd.Copy(f); err != nil {
			return nil, errors.Wrapf(err, "Copy %s", f.GetSourcePath())
		}
	}

	if err := installCertSymlinks(cmd, caCerts); err != nil {
		return nil, errors.Wrapf(err, "certificate symlinks")
	}
	return copyableFiles, nil
}

// CACerts has cert and key for CA (and Proxy)
type CACerts struct {
	caCert    string
	caKey     string
	proxyCert string
	proxyKey  string
}

// generateSharedCACerts generates CA certs shared among profiles, but only if missing
func generateSharedCACerts() (CACerts, error) {
	globalPath := localpath.MiniPath()
	cc := CACerts{
		caCert:    localpath.CACert(),
		caKey:     filepath.Join(globalPath, "ca.key"),
		proxyCert: filepath.Join(globalPath, "proxy-client-ca.crt"),
		proxyKey:  filepath.Join(globalPath, "proxy-client-ca.key"),
	}

	caCertSpecs := []struct {
		certPath string
		keyPath  string
		subject  string
	}{
		{ // client / apiserver CA
			certPath: cc.caCert,
			keyPath:  cc.caKey,
			subject:  "minikubeCA",
		},
		{ // proxy-client CA
			certPath: cc.proxyCert,
			keyPath:  cc.proxyKey,
			subject:  "proxyClientCA",
		},
	}

	for _, ca := range caCertSpecs {
		if canRead(ca.certPath) && canRead(ca.keyPath) {
			klog.Infof("skipping %s CA generation: %s", ca.subject, ca.keyPath)
			continue
		}

		klog.Infof("generating %s CA: %s", ca.subject, ca.keyPath)
		if err := util.GenerateCACert(ca.certPath, ca.keyPath, ca.subject); err != nil {
			return cc, errors.Wrap(err, "generate ca cert")
		}
	}

	return cc, nil
}

// generateProfileCerts generates profile certs for a profile
func generateProfileCerts(k8s config.KubernetesConfig, n config.Node, ccs CACerts) ([]string, error) {

	// Only generate these certs for the api server
	if !n.ControlPlane {
		return []string{}, nil
	}

	profilePath := localpath.Profile(k8s.ClusterName)

	serviceIP, err := util.GetServiceClusterIP(k8s.ServiceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "getting service cluster ip")
	}

	apiServerIPs := append(
		k8s.APIServerIPs,
		[]net.IP{net.ParseIP(n.IP), serviceIP, net.ParseIP(oci.DaemonHost(k8s.ContainerRuntime)), net.ParseIP("10.0.0.1")}...)
	apiServerNames := append(k8s.APIServerNames, k8s.APIServerName, constants.ControlPlaneAlias)
	apiServerAlternateNames := append(
		apiServerNames,
		util.GetAlternateDNS(k8s.DNSDomain)...)

	// Generate a hash input for certs that depend on ip/name combinations
	hi := []string{}
	hi = append(hi, apiServerAlternateNames...)
	for _, ip := range apiServerIPs {
		hi = append(hi, ip.String())
	}
	sort.Strings(hi)

	specs := []struct {
		certPath string
		keyPath  string
		hash     string

		subject        string
		ips            []net.IP
		alternateNames []string
		caCertPath     string
		caKeyPath      string
	}{
		{ // Client cert
			certPath:       localpath.ClientCert(k8s.ClusterName),
			keyPath:        localpath.ClientKey(k8s.ClusterName),
			subject:        "minikube-user",
			ips:            []net.IP{},
			alternateNames: []string{},
			caCertPath:     ccs.caCert,
			caKeyPath:      ccs.caKey,
		},
		{ // apiserver serving cert
			hash:           fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(hi, "/"))))[0:8],
			certPath:       filepath.Join(profilePath, "apiserver.crt"),
			keyPath:        filepath.Join(profilePath, "apiserver.key"),
			subject:        "minikube",
			ips:            apiServerIPs,
			alternateNames: apiServerAlternateNames,
			caCertPath:     ccs.caCert,
			caKeyPath:      ccs.caKey,
		},
		{ // aggregator proxy-client cert
			certPath:       filepath.Join(profilePath, "proxy-client.crt"),
			keyPath:        filepath.Join(profilePath, "proxy-client.key"),
			subject:        "aggregator",
			ips:            []net.IP{},
			alternateNames: []string{},
			caCertPath:     ccs.proxyCert,
			caKeyPath:      ccs.proxyKey,
		},
	}

	xfer := []string{}
	for _, spec := range specs {
		if spec.subject != "minikube-user" {
			xfer = append(xfer, spec.certPath)
			xfer = append(xfer, spec.keyPath)
		}

		cp := spec.certPath
		kp := spec.keyPath
		if spec.hash != "" {
			cp = cp + "." + spec.hash
			kp = kp + "." + spec.hash
		}

		if canRead(cp) && canRead(kp) {
			klog.Infof("skipping %s signed cert generation: %s", spec.subject, kp)
			continue
		}

		klog.Infof("generating %s signed cert: %s", spec.subject, kp)
		err := util.GenerateSignedCert(
			cp, kp, spec.subject,
			spec.ips, spec.alternateNames,
			spec.caCertPath, spec.caKeyPath,
		)
		if err != nil {
			return xfer, errors.Wrapf(err, "generate signed cert for %q", spec.subject)
		}

		if spec.hash != "" {
			klog.Infof("copying %s -> %s", cp, spec.certPath)
			if err := copy.Copy(cp, spec.certPath); err != nil {
				return xfer, errors.Wrap(err, "copy cert")
			}
			klog.Infof("copying %s -> %s", kp, spec.keyPath)
			if err := copy.Copy(kp, spec.keyPath); err != nil {
				return xfer, errors.Wrap(err, "copy key")
			}
		}
	}

	return xfer, nil
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
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		fullPath := filepath.Join(certsDir, hostpath)
		ext := strings.ToLower(filepath.Ext(hostpath))

		if ext == ".crt" || ext == ".pem" {
			if info.Size() < 32 {
				klog.Warningf("ignoring %s, impossibly tiny %d bytes", fullPath, info.Size())
				return nil
			}

			klog.Infof("found cert: %s (%d bytes)", fullPath, info.Size())

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
	lrr, err := cr.RunCmd(exec.Command("ls", "-la", filePath))
	if err != nil {
		return "", err
	}
	klog.Infof("hashing: %s", lrr.Stdout.String())

	rr, err := cr.RunCmd(exec.Command("openssl", "x509", "-hash", "-noout", "-in", filePath))
	if err != nil {
		crr, _ := cr.RunCmd(exec.Command("cat", filePath))
		return "", errors.Wrapf(err, "cert:\n%s\n---\n%s", lrr.Output(), crr.Stdout.String())
	}
	stringHash := strings.TrimSpace(rr.Stdout.String())
	return stringHash, nil
}

// installCertSymlinks installs certs in /usr/share/ca-certificates into system-wide certificate store (/etc/ssl/certs).
// OpenSSL binary required in minikube ISO
func installCertSymlinks(cr command.Runner, caCerts map[string]string) error {
	hasSSLBinary := true
	_, err := cr.RunCmd(exec.Command("openssl", "version"))
	if err != nil {
		hasSSLBinary = false
	}

	if !hasSSLBinary && len(caCerts) > 0 {
		klog.Warning("OpenSSL not found. Please recreate the cluster with the latest minikube ISO.")
	}

	for _, caCertFile := range caCerts {
		dstFilename := path.Base(caCertFile)
		certStorePath := path.Join(vmpath.GuestCertStoreDir, dstFilename)

		cmd := fmt.Sprintf("test -s %s && ln -fs %s %s", caCertFile, caCertFile, certStorePath)
		if _, err := cr.RunCmd(exec.Command("sudo", "/bin/bash", "-c", cmd)); err != nil {
			return errors.Wrapf(err, "create symlink for %s", caCertFile)
		}

		if !hasSSLBinary {
			continue
		}

		subjectHash, err := getSubjectHash(cr, caCertFile)
		if err != nil {
			return errors.Wrapf(err, "calculate hash for cacert %s", caCertFile)
		}
		subjectHashLink := path.Join(vmpath.GuestCertStoreDir, fmt.Sprintf("%s.0", subjectHash))

		// NOTE: This symlink may exist, but point to a missing file
		cmd = fmt.Sprintf("test -L %s || ln -fs %s %s", subjectHashLink, certStorePath, subjectHashLink)
		if _, err := cr.RunCmd(exec.Command("sudo", "/bin/bash", "-c", cmd)); err != nil {
			return errors.Wrapf(err, "create symlink for %s", caCertFile)
		}
	}
	return nil
}

// canRead returns true if the file represented
// by path exists and is readable, otherwise false.
func canRead(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	return true
}
