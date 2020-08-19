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
	"sync"
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

	t.Cleanup(func() {
		err := os.RemoveAll(td)
		t.Logf("remove path %q", td)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", td)
		}
	})
	err = os.Setenv(localpath.MinikubeHome, td)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	// Not necessary, but it is a handy random alphanumeric
	name := filepath.Base(td)
	if err := os.MkdirAll(config.ProfileFolderPath(name), 0777); err != nil {
		t.Fatalf("error creating temporary directory")
	}

	cc := &config.ClusterConfig{
		Name:             name,
		CPUs:             2,
		Memory:           2500,
		KubernetesConfig: config.KubernetesConfig{},
	}

	if err := config.DefaultLoader.WriteConfigToFile(name, cc); err != nil {
		t.Fatalf("error creating temporary profile config: %v", err)
	}
	return name
}

func TestIsAddonAlreadySet(t *testing.T) {
	cc := &config.ClusterConfig{Name: "test"}

	if err := Set(cc, "registry", "true"); err != nil {
		t.Errorf("unable to set registry true: %v", err)
	}
	if !assets.Addons["registry"].IsEnabled(cc) {
		t.Errorf("expected registry to be enabled")
	}

	if assets.Addons["ingress"].IsEnabled(cc) {
		t.Errorf("expected ingress to not be enabled")
	}

}

func TestDisableUnknownAddon(t *testing.T) {
	cc := &config.ClusterConfig{Name: "test"}

	if err := Set(cc, "InvalidAddon", "false"); err == nil {
		t.Fatalf("Disable did not return error for unknown addon")
	}
}

func TestEnableUnknownAddon(t *testing.T) {
	cc := &config.ClusterConfig{Name: "test"}

	if err := Set(cc, "InvalidAddon", "true"); err == nil {
		t.Fatalf("Enable did not return error for unknown addon")
	}
}

func TestSetAndSave(t *testing.T) {
	profile := createTestProfile(t)

	// enable
	if err := SetAndSave(profile, "dashboard", "true"); err != nil {
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
	if err := SetAndSave(profile, "dashboard", "false"); err != nil {
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
	cc := &config.ClusterConfig{
		Name:             "start",
		CPUs:             2,
		Memory:           2500,
		KubernetesConfig: config.KubernetesConfig{},
	}

	var wg sync.WaitGroup
	Start(&wg, cc, map[string]bool{}, []string{"dashboard"})
	wg.Wait()

	if !assets.Addons["dashboard"].IsEnabled(cc) {
		t.Errorf("expected dashboard to be enabled")
	}
}
