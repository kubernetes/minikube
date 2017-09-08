// +build integration
// +build versioned

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

package integration

import (
	"fmt"
	"testing"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/kubernetes_versions"
	"k8s.io/minikube/test/integration/util"
)

func TestVersionedFunctional(t *testing.T) {
	k8sVersions, err := kubernetes_versions.GetK8sVersionsFromURL(constants.KubernetesVersionGCSURL)
	if err != nil {
		t.Fatalf(err.Error())
	}
	var minikubeRunner util.MinikubeRunner
	for _, version := range k8sVersions {
		vArgs := fmt.Sprintf("%s --kubernetes-version %s", *args, version.Version)
		minikubeRunner = NewMinikubeRunner(t)
		minikubeRunner.EnsureRunning()

		t.Run("Status", testClusterStatus)
		t.Run("DNS", testClusterDNS)
		t.Run("Addons", testAddons)
	}
}
