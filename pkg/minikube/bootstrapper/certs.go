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
	"net"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/kubeconfig"
)

var (
	certs = []string{"ca.crt", "ca.key", "apiserver.crt", "apiserver.key"}
	// This is the internalIP , the API server and other components communicate on.
	internalIP = net.ParseIP(util.DefaultServiceClusterIP)
)

// SetupCerts gets the generated credentials required to talk to the APIServer.
func SetupCerts(cmd CommandRunner, k8s KubernetesConfig) error {
	localPath := constants.GetMinipath()
	glog.Infoln("Setting up certificates for IP: %s", k8s.NodeIP)

	ip := net.ParseIP(k8s.NodeIP)
	caCert := filepath.Join(localPath, "ca.crt")
	caKey := filepath.Join(localPath, "ca.key")
	publicPath := filepath.Join(localPath, "apiserver.crt")
	privatePath := filepath.Join(localPath, "apiserver.key")
	if err := generateCerts(caCert, caKey, publicPath, privatePath, ip, k8s.APIServerName, k8s.DNSDomain); err != nil {
		return errors.Wrap(err, "Error generating certs")
	}

	copyableFiles := []assets.CopyableFile{}

	for _, cert := range certs {
		p := filepath.Join(localPath, cert)
		perms := "0644"
		if strings.HasSuffix(cert, ".key") {
			perms = "0600"
		}
		certFile, err := assets.NewFileAsset(p, util.DefaultCertPath, cert, perms)
		if err != nil {
			return err
		}
		copyableFiles = append(copyableFiles, certFile)
	}

	kubeCfgSetup := &kubeconfig.KubeConfigSetup{
		ClusterName:          k8s.NodeName,
		ClusterServerAddress: "https://localhost:8443",
		ClientCertificate:    filepath.Join(util.DefaultCertPath, "apiserver.crt"),
		ClientKey:            filepath.Join(util.DefaultCertPath, "apiserver.key"),
		CertificateAuthority: filepath.Join(util.DefaultCertPath, "ca.crt"),
		KeepContext:          false,
	}

	kubeCfg := api.NewConfig()
	kubeconfig.PopulateKubeConfig(kubeCfgSetup, kubeCfg)
	data, err := runtime.Encode(latest.Codec, kubeCfg)
	if err != nil {
		return errors.Wrap(err, "encoding kubeconfig")
	}

	kubeCfgFile := assets.NewMemoryAsset(data,
		util.DefaultLocalkubeDirectory, "kubeconfig", "0644")
	copyableFiles = append(copyableFiles, kubeCfgFile)

	for _, f := range copyableFiles {
		if err := cmd.Copy(f); err != nil {
			return err
		}
	}
	return nil
}

func generateCerts(caCert, caKey, pub, priv string, ip net.IP, name string, dnsDomain string) error {
	if !(util.CanReadFile(caCert) && util.CanReadFile(caKey)) {
		if err := util.GenerateCACert(caCert, caKey, name); err != nil {
			return errors.Wrap(err, "Error generating certificate")
		}
	}

	ips := []net.IP{ip, internalIP}
	if err := util.GenerateSignedCert(pub, priv, ips, util.GetAlternateDNS(dnsDomain), caCert, caKey); err != nil {
		return errors.Wrap(err, "Error generating signed cert")
	}
	return nil
}
