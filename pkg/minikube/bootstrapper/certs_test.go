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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/util"
)

func TestSetupCerts(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	k8s := config.KubernetesConfig{
		APIServerName: constants.APIServerName,
		DNSDomain:     constants.ClusterDNSDomain,
		ServiceCIDR:   util.DefaultServiceCIDR,
	}

	if err := os.Mkdir(filepath.Join(tempDir, "certs"), 0777); err != nil {
		t.Fatalf("error create certificate directory: %v", err)
	}

	if err := util.GenerateCACert(
		filepath.Join(tempDir, "certs", "mycert.pem"),
		filepath.Join(tempDir, "certs", "mykey.pem"),
		"Test Certificate",
	); err != nil {
		t.Fatalf("error generating certificate: %v", err)
	}

	cmdMap := map[string]string{
		"sudo mkdir -p  /var/lib/minikube/certs": "",
	}
	certFilenames := map[string]string{"ca.crt": "minikubeCA.pem", "mycert.pem": "mycert.pem"}
	for _, dst := range certFilenames {
		certFile := path.Join(CACertificatesDir, dst)
		certStorePath := path.Join(SSLCertStoreDir, dst)
		certNameHash := "abcdef"
		remoteCertHashLink := path.Join(SSLCertStoreDir, fmt.Sprintf("%s.0", certNameHash))
		cmdMap[fmt.Sprintf("sudo ln -s %s %s", certFile, certStorePath)] = "1"
		cmdMap[fmt.Sprintf("openssl x509 -hash -noout -in %s", certFile)] = certNameHash
		cmdMap[fmt.Sprintf("sudo ln -s %s %s", certStorePath, remoteCertHashLink)] = "1"
	}
	f := command.NewFakeCommandRunner()
	f.SetCommandToOutput(cmdMap)

	var filesToBeTransferred []string
	for _, cert := range certs {
		filesToBeTransferred = append(filesToBeTransferred, filepath.Join(localpath.MiniPath(), cert))
	}
	filesToBeTransferred = append(filesToBeTransferred, filepath.Join(localpath.MiniPath(), "ca.crt"))
	filesToBeTransferred = append(filesToBeTransferred, filepath.Join(localpath.MiniPath(), "certs", "mycert.pem"))

	if err := SetupCerts(f, k8s); err != nil {
		t.Fatalf("Error starting cluster: %v", err)
	}
	for _, cert := range filesToBeTransferred {
		_, err := f.GetFileToContents(cert)
		if err != nil {
			t.Errorf("Cert not generated: %s", cert)
		}
	}
}
