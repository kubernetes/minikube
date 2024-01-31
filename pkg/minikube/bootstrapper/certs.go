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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	// WARNING: use path for kic/iso and path/filepath for user os
	"path"
	"path/filepath"

	"github.com/juju/mutex/v2"
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
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/lock"
)

// sharedCACerts represents minikube Root CA and Proxy Client CA certs and keys shared among profiles.
type sharedCACerts struct {
	caCert    string
	caKey     string
	proxyCert string
	proxyKey  string
}

// SetupCerts gets the generated credentials required to talk to the APIServer.
func SetupCerts(k8s config.ClusterConfig, n config.Node, pcpCmd command.Runner, cmd command.Runner) error {
	localPath := localpath.Profile(k8s.KubernetesConfig.ClusterName)
	klog.Infof("Setting up %s for IP: %s", localPath, n.IP)

	sharedCerts, regen, err := generateSharedCACerts()
	if err != nil {
		return errors.Wrap(err, "generate shared ca certs")
	}

	xfer := []string{
		sharedCerts.caCert,
		sharedCerts.caKey,
		sharedCerts.proxyCert,
		sharedCerts.proxyKey,
	}

	// only generate/renew certs for control-plane nodes or if needs regenating
	if n.ControlPlane || regen {
		profileCerts, err := generateProfileCerts(k8s, n, sharedCerts, regen)
		if err != nil {
			return errors.Wrap(err, "generate profile certs")
		}
		xfer = append(xfer, profileCerts...)
	}

	copyableFiles := []assets.CopyableFile{}
	defer func() {
		for _, f := range copyableFiles {
			if err := f.Close(); err != nil {
				klog.Warningf("error closing the file %s: %v", f.GetSourcePath(), err)
			}
		}
	}()

	for _, c := range xfer {
		// note: src(c) is user os' path, dst is kic/iso (linux) path
		certFile, err := assets.NewFileAsset(c, vmpath.GuestKubernetesCertsDir, filepath.Base(c), properPerms(c))
		if err != nil {
			return errors.Wrapf(err, "create cert file asset for %s", c)
		}
		copyableFiles = append(copyableFiles, certFile)
	}

	caCerts, err := collectCACerts()
	if err != nil {
		return errors.Wrap(err, "collect ca certs")
	}

	for src, dst := range caCerts {
		// note: these are all public certs, so should be world-readeable
		// note: src is user os' path, dst is kic/iso (linux) path
		certFile, err := assets.NewFileAsset(src, path.Dir(dst), path.Base(dst), "0644")
		if err != nil {
			return errors.Wrapf(err, "create ca cert file asset for %s", src)
		}
		copyableFiles = append(copyableFiles, certFile)
	}

	if n.ControlPlane {
		// copy essential certs from primary control-plane node to secondaries
		// ref: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/high-availability/#manual-certs
		if !config.IsPrimaryControlPlane(k8s, n) {
			pcpCerts := []struct {
				srcDir  string
				srcFile string
				dstFile string
			}{
				{vmpath.GuestKubernetesCertsDir, "sa.pub", "sa.pub"},
				{vmpath.GuestKubernetesCertsDir, "sa.key", "sa.key"},
				{vmpath.GuestKubernetesCertsDir, "front-proxy-ca.crt", "front-proxy-ca.crt"},
				{vmpath.GuestKubernetesCertsDir, "front-proxy-ca.key", "front-proxy-ca.key"},
				{vmpath.GuestKubernetesCertsDir + "/etcd", "ca.crt", "etcd-ca.crt"},
				{vmpath.GuestKubernetesCertsDir + "/etcd", "ca.key", "etcd-ca.key"},
			}
			for _, c := range pcpCerts {
				// get cert from primary control-plane node
				f := assets.NewMemoryAsset(nil, c.srcDir, c.srcFile, properPerms(c.dstFile))
				if err := pcpCmd.CopyFrom(f); err != nil {
					klog.Errorf("unable to copy %s/%s from primary control-plane to %s in node %q: %v", c.srcDir, c.srcFile, c.dstFile, n.Name, err)
				}
				// put cert to secondary control-plane node
				copyableFiles = append(copyableFiles, f)
			}
		}

		// generate kubeconfig for control-plane node
		kcs := &kubeconfig.Settings{
			ClusterName:          n.Name,
			ClusterServerAddress: fmt.Sprintf("https://%s", net.JoinHostPort("localhost", fmt.Sprint(n.Port))),
			ClientCertificate:    path.Join(vmpath.GuestKubernetesCertsDir, "apiserver.crt"),
			ClientKey:            path.Join(vmpath.GuestKubernetesCertsDir, "apiserver.key"),
			CertificateAuthority: path.Join(vmpath.GuestKubernetesCertsDir, "ca.crt"),
			ExtensionContext:     kubeconfig.NewExtension(),
			ExtensionCluster:     kubeconfig.NewExtension(),
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
	}

	for _, f := range copyableFiles {
		if err := cmd.Copy(f); err != nil {
			return errors.Wrapf(err, "Copy %s", f.GetSourcePath())
		}
	}

	if err := installCertSymlinks(cmd, caCerts); err != nil {
		return errors.Wrap(err, "install cert symlinks")
	}

	if err := renewExpiredKubeadmCerts(cmd, k8s); err != nil {
		return errors.Wrap(err, "renew expired kubeadm certs")
	}

	return nil
}

// generateSharedCACerts generates minikube Root CA and Proxy Client CA certs, but only if missing or expired.
func generateSharedCACerts() (sharedCACerts, bool, error) {
	klog.Info("generating shared ca certs ...")

	regenProfileCerts := false
	globalPath := localpath.MiniPath()
	cc := sharedCACerts{
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

	// create a lock for "ca-certs" to avoid race condition over multiple minikube instances rewriting ca certs
	hold := filepath.Join(globalPath, "ca-certs")
	spec := lock.PathMutexSpec(hold)
	spec.Timeout = 1 * time.Minute
	klog.Infof("acquiring lock for ca certs: %+v", spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return cc, false, errors.Wrapf(err, "acquire lock for ca certs %+v", spec)
	}
	defer releaser.Release()

	for _, ca := range caCertSpecs {
		if isValid(ca.certPath, ca.keyPath) {
			klog.Infof("skipping valid %q ca cert: %s", ca.subject, ca.keyPath)
			continue
		}

		regenProfileCerts = true
		klog.Infof("generating %q ca cert: %s", ca.subject, ca.keyPath)
		if err := util.GenerateCACert(ca.certPath, ca.keyPath, ca.subject); err != nil {
			return cc, false, errors.Wrapf(err, "generate %q ca cert: %s", ca.subject, ca.keyPath)
		}
	}

	return cc, regenProfileCerts, nil
}

// generateProfileCerts generates certs for a profile, but only if missing, expired or needs regenerating.
func generateProfileCerts(cfg config.ClusterConfig, n config.Node, shared sharedCACerts, regen bool) ([]string, error) {
	// Only generate these certs for the api server
	if !n.ControlPlane {
		return []string{}, nil
	}

	klog.Info("generating profile certs ...")

	k8s := cfg.KubernetesConfig

	serviceIP, err := util.ServiceClusterIP(k8s.ServiceCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "get service cluster ip")
	}

	apiServerIPs := append([]net.IP{}, k8s.APIServerIPs...)
	apiServerIPs = append(apiServerIPs, serviceIP, net.ParseIP(oci.DefaultBindIPV4), net.ParseIP("10.0.0.1"))
	// append ip addresses of all control-plane nodes
	for _, n := range config.ControlPlanes(cfg) {
		apiServerIPs = append(apiServerIPs, net.ParseIP(n.IP))
	}
	if config.IsHA(cfg) {
		apiServerIPs = append(apiServerIPs, net.ParseIP(cfg.KubernetesConfig.APIServerHAVIP))
	}

	apiServerNames := append([]string{}, k8s.APIServerNames...)
	apiServerNames = append(apiServerNames, k8s.APIServerName, constants.ControlPlaneAlias, config.MachineName(cfg, n))

	apiServerAlternateNames := append([]string{}, apiServerNames...)
	apiServerAlternateNames = append(apiServerAlternateNames, util.AlternateDNS(k8s.DNSDomain)...)

	daemonHost := oci.DaemonHost(k8s.ContainerRuntime)
	if daemonHost != oci.DefaultBindIPV4 {
		daemonHostIP := net.ParseIP(daemonHost)
		// if daemonHost is an IP we add it to the certificate's IPs, otherwise we assume it's an hostname and add it to the alternate names
		if daemonHostIP != nil {
			apiServerIPs = append(apiServerIPs, daemonHostIP)
		} else {
			apiServerAlternateNames = append(apiServerAlternateNames, daemonHost)
		}
	}

	// Generate a hash input for certs that depend on ip/name combinations
	hi := append([]string{}, apiServerAlternateNames...)
	for _, ip := range apiServerIPs {
		hi = append(hi, ip.String())
	}
	// eliminate duplicates in 'hi'
	slices.Sort(hi)
	hi = slices.Compact(hi)

	profilePath := localpath.Profile(k8s.ClusterName)

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
		{ // client cert
			certPath:       localpath.ClientCert(k8s.ClusterName),
			keyPath:        localpath.ClientKey(k8s.ClusterName),
			subject:        "minikube-user",
			ips:            []net.IP{},
			alternateNames: []string{},
			caCertPath:     shared.caCert,
			caKeyPath:      shared.caKey,
		},
		{ // apiserver serving cert
			hash:           fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(hi, "/"))))[0:8],
			certPath:       filepath.Join(profilePath, "apiserver.crt"),
			keyPath:        filepath.Join(profilePath, "apiserver.key"),
			subject:        "minikube",
			ips:            apiServerIPs,
			alternateNames: apiServerAlternateNames,
			caCertPath:     shared.caCert,
			caKeyPath:      shared.caKey,
		},
		{ // aggregator proxy-client cert
			certPath:       filepath.Join(profilePath, "proxy-client.crt"),
			keyPath:        filepath.Join(profilePath, "proxy-client.key"),
			subject:        "aggregator",
			ips:            []net.IP{},
			alternateNames: []string{},
			caCertPath:     shared.proxyCert,
			caKeyPath:      shared.proxyKey,
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

		if !regen && isValid(cp, kp) {
			klog.Infof("skipping valid signed profile cert regeneration for %q: %s", spec.subject, kp)
			continue
		}

		klog.Infof("generating signed profile cert for %q: %s", spec.subject, kp)
		if canRead(cp) {
			os.Remove(cp)
		}
		if canRead(kp) {
			os.Remove(kp)
		}
		err := util.GenerateSignedCert(
			cp, kp, spec.subject,
			spec.ips, spec.alternateNames,
			spec.caCertPath, spec.caKeyPath,
			cfg.CertExpiration,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "generate signed profile cert for %q", spec.subject)
		}

		if spec.hash != "" {
			klog.Infof("copying %s -> %s", cp, spec.certPath)
			if err := copy.Copy(cp, spec.certPath); err != nil {
				return nil, errors.Wrap(err, "copy profile cert")
			}
			klog.Infof("copying %s -> %s", kp, spec.keyPath)
			if err := copy.Copy(kp, spec.keyPath); err != nil {
				return nil, errors.Wrap(err, "copy profile cert key")
			}
		}
	}

	return xfer, nil
}

// renewExpiredKubeadmCerts checks if kubeadm certs already exists and are still valid, then renews them if needed.
// if certs don't exist already (eg, kubeadm hasn't run yet), then checks are skipped.
func renewExpiredKubeadmCerts(cmd command.Runner, cc config.ClusterConfig) error {
	if _, err := cmd.RunCmd(exec.Command("stat", path.Join(vmpath.GuestPersistentDir, "certs", "apiserver-kubelet-client.crt"))); err != nil {
		klog.Infof("'apiserver-kubelet-client' cert doesn't exist, likely first start: %v", err)
		return nil
	}

	expiredCerts := false
	certs := []string{"apiserver-etcd-client", "apiserver-kubelet-client", "etcd-server", "etcd-healthcheck-client", "etcd-peer", "front-proxy-client"}
	for _, cert := range certs {
		certPath := []string{vmpath.GuestPersistentDir, "certs"}
		// certs starting with "etcd-" are in the "etcd" dir
		// ex: etcd-server => etcd/server
		if strings.HasPrefix(cert, "etcd-") {
			certPath = append(certPath, "etcd")
		}
		certPath = append(certPath, strings.TrimPrefix(cert, "etcd-")+".crt")
		if !isKubeadmCertValid(cmd, path.Join(certPath...)) {
			expiredCerts = true
		}
	}
	if !expiredCerts {
		return nil
	}
	out.WarningT("kubeadm certificates have expired. Generating new ones...")
	kubeadmPath := path.Join(vmpath.GuestPersistentDir, "binaries", cc.KubernetesConfig.KubernetesVersion)
	bashCmd := fmt.Sprintf("sudo env PATH=\"%s:$PATH\" kubeadm certs renew all --config %s", kubeadmPath, constants.KubeadmYamlPath)
	if _, err := cmd.RunCmd(exec.Command("/bin/bash", "-c", bashCmd)); err != nil {
		return errors.Wrap(err, "kubeadm certs renew")
	}
	return nil
}

// isValidPEMCertificate checks whether the input file is a valid PEM certificate (with at least one CERTIFICATE block)
func isValidPEMCertificate(filePath string) (bool, error) {
	fileBytes, err := os.ReadFile(filePath)
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

// collectCACerts looks up all public pem certificates with .crt or .pem extension
// in ~/.minikube/certs or ~/.minikube/files/etc/ssl/certs
// to copy them to the vmpath.GuestCertAuthDir ("/usr/share/ca-certificates") in host.
// minikube root CA is also included but libmachine certificates (ca.pem/cert.pem) are excluded.
func collectCACerts() (map[string]string, error) {
	localPath := localpath.MiniPath()
	// note: certFiles map's key is user os' path, whereas map's value is kic/iso (linux) path
	certFiles := map[string]string{}

	dirs := []string{filepath.Join(localPath, "certs"), filepath.Join(localPath, "files", "etc", "ssl", "certs")}
	for _, certsDir := range dirs {
		err := filepath.Walk(certsDir, func(hostpath string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			if info == nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}

			ext := filepath.Ext(hostpath)
			if strings.ToLower(ext) == ".crt" || strings.ToLower(ext) == ".pem" {
				if info.Size() < 32 {
					klog.Warningf("ignoring %s, impossibly tiny %d bytes", hostpath, info.Size())
					return nil
				}

				klog.Infof("found cert: %s (%d bytes)", hostpath, info.Size())

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
			return nil, errors.Wrapf(err, "collecting CA certs from %s", certsDir)
		}

		excluded := []string{"ca.pem", "cert.pem"}
		for _, e := range excluded {
			certFiles[filepath.Join(certsDir, e)] = ""
		}
	}

	// include minikube CA
	certFiles[localpath.CACert()] = path.Join(vmpath.GuestCertAuthDir, "minikubeCA.pem")

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

// isValid checks a cert & key paths exist and are still valid.
// If a cert is expired or otherwise invalid, it will be deleted.
func isValid(certPath, keyPath string) bool {
	if !canRead(keyPath) {
		return false
	}

	certFile, err := os.ReadFile(certPath)
	if err != nil {
		klog.Infof("failed to read cert file %s: %v", certPath, err)
		os.Remove(certPath)
		os.Remove(keyPath)
		return false
	}

	certData, _ := pem.Decode(certFile)
	if certData == nil {
		klog.Infof("failed to decode cert file %s", certPath)
		os.Remove(certPath)
		os.Remove(keyPath)
		return false
	}

	cert, err := x509.ParseCertificate(certData.Bytes)
	if err != nil {
		klog.Infof("failed to parse cert file %s: %v\n", certPath, err)
		os.Remove(certPath)
		os.Remove(keyPath)
		return false
	}

	if cert.NotAfter.Before(time.Now()) {
		out.WarningT("Certificate {{.certPath}} has expired. Generating a new one...", out.V{"certPath": filepath.Base(certPath)})
		klog.Infof("cert expired %s: expiration: %s, now: %s", certPath, cert.NotAfter, time.Now())
		os.Remove(certPath)
		os.Remove(keyPath)
		return false
	}

	return true
}

func isKubeadmCertValid(cmd command.Runner, certPath string) bool {
	_, err := cmd.RunCmd(exec.Command("openssl", "x509", "-noout", "-in", certPath, "-checkend", "86400"))
	if err != nil {
		klog.Infof("%v", err)
	}
	return err == nil
}

// properPerms returns proper permissions for given cert file, based on its extension.
func properPerms(cert string) string {
	perms := "0644"

	ext := strings.ToLower(filepath.Ext(cert))
	if ext == ".key" || ext == ".pem" {
		perms = "0600"
	}

	return perms
}
