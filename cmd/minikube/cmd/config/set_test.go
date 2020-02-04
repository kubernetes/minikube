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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestNotFound(t *testing.T) {
	createTestProfile(t)
	err := Set("nonexistent", "10")
	if err == nil {
		t.Fatalf("Set did not return error for unknown property")
	}
}

func TestSetNotAllowed(t *testing.T) {
	createTestProfile(t)
	err := Set("vm-driver", "123456")
	if err == nil || err.Error() != "[driver \"123456\" is not supported]" {
		t.Fatalf("Set did not return error for unallowed value")
	}
}

func TestSetOK(t *testing.T) {
	createTestProfile(t)
	err := Set("vm-driver", "virtualbox")
	defer func() {
		err = Unset("vm-driver")
		if err != nil {
			t.Errorf("failed to unset vm-driver")
		}
	}()
	if err != nil {
		t.Fatalf("Set returned error for valid property value")
	}
	val, err := Get("vm-driver")
	if err != nil {
		t.Fatalf("Get returned error for valid property")
	}
	if val != "virtualbox" {
		t.Fatalf("Get returned %s, expected \"virtualbox\"", val)
	}
}

func createTestProfile(t *testing.T) {
	t.Helper()
	td, err := ioutil.TempDir("", "profile")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}

	err = os.Setenv(localpath.MinikubeHome, td)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	// Not necessary, but it is a handy random alphanumeric
	name := filepath.Base(td)
	if err := os.MkdirAll(config.ProfileFolderPath(name), 0777); err != nil {
		t.Fatalf("error creating temporary directory")
	}
	if err := config.DefaultLoader.WriteConfigToFile(name, &config.MachineConfig{}); err != nil {
		t.Fatalf("error creating temporary profile config: %v", err)
	}
}
