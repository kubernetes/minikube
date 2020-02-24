/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package addons

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func createTestProfile(t *testing.T) string {
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
	if err := config.DefaultLoader.WriteConfigToFile(name, &config.ClusterConfig{}); err != nil {
		t.Fatalf("error creating temporary profile config: %v", err)
	}
	return name
}

func TestIsAddonAlreadySet(t *testing.T) {
	profile := createTestProfile(t)
	if err := Set("registry", "true", profile); err != nil {
		t.Errorf("unable to set registry true: %v", err)
	}
	enabled, err := assets.Addons["registry"].IsEnabled(profile)
	if err != nil {
		t.Errorf("registry: %v", err)
	}
	if !enabled {
		t.Errorf("expected registry to be enabled")
	}

	enabled, err = assets.Addons["ingress"].IsEnabled(profile)
	if err != nil {
		t.Errorf("ingress: %v", err)
	}
	if enabled {
		t.Errorf("expected ingress to not be enabled")
	}

}

func TestDisableUnknownAddon(t *testing.T) {
	profile := createTestProfile(t)
	if err := Set("InvalidAddon", "false", profile); err == nil {
		t.Fatalf("Disable did not return error for unknown addon")
	}
}

func TestEnableUnknownAddon(t *testing.T) {
	profile := createTestProfile(t)
	if err := Set("InvalidAddon", "true", profile); err == nil {
		t.Fatalf("Enable did not return error for unknown addon")
	}
}

func TestEnableAndDisableAddon(t *testing.T) {
	profile := createTestProfile(t)

	// enable
	if err := Set("dashboard", "true", profile); err != nil {
		t.Errorf("Disable returned unexpected error: " + err.Error())
	}

	c, err := config.DefaultLoader.LoadConfigFromFile(profile)
	if err != nil {
		t.Errorf("unable to load profile: %v", err)
	}
	if c.Addons["dashboard"] != true {
		t.Errorf("expected dashboard to be enabled")
	}

	// disable
	if err := Set("dashboard", "false", profile); err != nil {
		t.Errorf("Disable returned unexpected error: " + err.Error())
	}

	c, err = config.DefaultLoader.LoadConfigFromFile(profile)
	if err != nil {
		t.Errorf("unable to load profile: %v", err)
	}
	if c.Addons["dashboard"] != false {
		t.Errorf("expected dashboard to be enabled")
	}
}

func TestStart(t *testing.T) {
	profile := createTestProfile(t)
	Start(profile, map[string]bool{}, []string{"dashboard"})

	enabled, err := assets.Addons["dashboard"].IsEnabled(profile)
	if err != nil {
		t.Errorf("dashboard: %v", err)
	}
	if !enabled {
		t.Errorf("expected dashboard to be enabled")
	}
}
