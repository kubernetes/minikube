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

package config

import (
	"testing"

	"gotest.tools/assert"
	pkgConfig "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestEnableUnknownAddon(t *testing.T) {
	if err := Set("InvalidAddon", "false"); err == nil {
		t.Fatalf("Enable did not return error for unknown addon")
	}
}

func TestEnableAddon(t *testing.T) {
	if err := Set("ingress", "true"); err != nil {
		t.Fatalf("Enable returned unexpected error: " + err.Error())
	}
	config, _ := pkgConfig.ReadConfig(localpath.ConfigFile)
	assert.Equal(t, config["ingress"], true)
}
