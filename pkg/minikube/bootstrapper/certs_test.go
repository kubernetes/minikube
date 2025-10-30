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
	"os"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/util"
)

func TestSetupCerts(t *testing.T) {
	tempDir := tests.MakeTempDir(t)

	k8s := config.ClusterConfig{
		CertExpiration: constants.DefaultCertExpiration,
		KubernetesConfig: config.KubernetesConfig{
			APIServerName: constants.APIServerName,
			DNSDomain:     constants.DefaultDNSDomain,
			ServiceCIDR:   constants.DefaultServiceCIDR,
		},
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

	expected := map[string]string{
		`sudo /bin/bash -c "test -s /usr/share/ca-certificates/mycert.pem && ln -fs /usr/share/ca-certificates/mycert.pem /etc/ssl/certs/mycert.pem"`:             "-",
		`sudo /bin/bash -c "test -s /usr/share/ca-certificates/minikubeCA.pem && ln -fs /usr/share/ca-certificates/minikubeCA.pem /etc/ssl/certs/minikubeCA.pem"`: "-",
		`date -u +%d-%m-%y-%T`: time.Now().Format("02-01-06-15:04:05"),
	}
	f := command.NewFakeCommandRunner()
	f.SetCommandToOutput(expected)

	p := command.NewFakeCommandRunner()
	p.SetCommandToOutput(map[string]string{})

	if err := SetupCerts(k8s, config.Node{}, p, f); err != nil {
		t.Fatalf("Error starting cluster: %v", err)
	}
}
