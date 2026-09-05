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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/minikube/tests"
)

func createTestProfile(t *testing.T) string {
	t.Helper()
	td := t.TempDir()

	t.Setenv(localpath.MinikubeHome, td)

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
		Nodes:            []config.Node{{ControlPlane: true}},
	}

	if err := config.DefaultLoader.WriteConfigToFile(name, cc); err != nil {
		t.Fatalf("error creating temporary profile config: %v", err)
	}
	return name
}

func TestIsAddonAlreadySet(t *testing.T) {
	cc := &config.ClusterConfig{
		Name:  "test",
		Nodes: []config.Node{{ControlPlane: true}},
	}

	if err := Set(cc, "registry", "true", &run.CommandOptions{}); err != nil {
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
	cc := &config.ClusterConfig{
		Name:  "test",
		Nodes: []config.Node{{ControlPlane: true}},
	}

	if err := Set(cc, "InvalidAddon", "false", &run.CommandOptions{}); err == nil {
		t.Fatalf("Disable did not return error for unknown addon")
	}
}

func TestEnableUnknownAddon(t *testing.T) {
	cc := &config.ClusterConfig{
		Name:  "test",
		Nodes: []config.Node{{ControlPlane: true}},
	}

	if err := Set(cc, "InvalidAddon", "true", &run.CommandOptions{}); err == nil {
		t.Fatalf("Enable did not return error for unknown addon")
	}
}

func TestSetAndSave(t *testing.T) {
	profile := createTestProfile(t)

	// enable
	if err := SetAndSave(profile, "dashboard", "true", &run.CommandOptions{}); err != nil {
		t.Errorf("Disable returned unexpected error: %v", err)
	}

	c, err := config.DefaultLoader.LoadConfigFromFile(profile)
	if err != nil {
		t.Errorf("unable to load profile: %v", err)
	}
	if c.Addons["dashboard"] != true {
		t.Errorf("expected dashboard to be enabled")
	}

	// disable
	if err := SetAndSave(profile, "dashboard", "false", &run.CommandOptions{}); err != nil {
		t.Errorf("Disable returned unexpected error: %v", err)
	}

	c, err = config.DefaultLoader.LoadConfigFromFile(profile)
	if err != nil {
		t.Errorf("unable to load profile: %v", err)
	}
	if c.Addons["dashboard"] != false {
		t.Errorf("expected dashboard to be enabled")
	}
}

func TestStartWithAddonsEnabled(t *testing.T) {
	// this test will write a config.json into MinikubeHome, create a temp dir for it
	tests.MakeTempDir(t)

	cc := &config.ClusterConfig{
		Name:             "start",
		CPUs:             2,
		Memory:           2500,
		KubernetesConfig: config.KubernetesConfig{},
		Nodes:            []config.Node{{ControlPlane: true}},
	}
	options := &run.CommandOptions{}
	toEnable := ToEnable(cc, map[string]bool{}, []string{"dashboard"})
	enabled := make(chan []string, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go Enable(&wg, cc, toEnable, enabled, options)
	wg.Wait()
	if ea, ok := <-enabled; ok {
		UpdateConfigToEnable(cc, ea, options)
	}

	if !assets.Addons["dashboard"].IsEnabled(cc) {
		t.Errorf("expected dashboard to be enabled")
	}
}

func TestStartWithAllAddonsDisabled(t *testing.T) {
	// this test will write a config.json into MinikubeHome, create a temp dir for it
	tests.MakeTempDir(t)

	cc := &config.ClusterConfig{
		Name:             "start",
		CPUs:             2,
		Memory:           2500,
		KubernetesConfig: config.KubernetesConfig{},
		Nodes:            []config.Node{{ControlPlane: true}},
	}

	UpdateConfigToDisable(cc, &run.CommandOptions{})

	for name := range assets.Addons {
		if assets.Addons[name].IsEnabled(cc) {
			t.Errorf("expected %s to be disabled", name)
		}
	}
}

func TestInvokeFailFast(t *testing.T) {
	secondCalled := false
	expectedErr := fmt.Errorf("first callback failed")

	firstFn := func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
		return expectedErr
	}
	secondFn := func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
		secondCalled = true
		return nil
	}

	fns := []setFn{firstFn, secondFn}
	err := invoke(nil, "test", "true", fns, nil)

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if secondCalled {
		t.Errorf("expected second callback NOT to be called after first callback failed")
	}
}

func TestInvokeAll(t *testing.T) {
	secondCalled := false
	err1 := fmt.Errorf("first callback failed")
	err2 := fmt.Errorf("second callback failed")

	firstFn := func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
		return err1
	}
	secondFn := func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
		secondCalled = true
		return err2
	}

	fns := []setFn{firstFn, secondFn}
	err := invokeAll(nil, "test", "false", fns, nil)

	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !secondCalled {
		t.Errorf("expected second callback to be called despite first callback failure")
	}
}

func TestInvokeAllSkipAddon(t *testing.T) {
	secondCalled := false

	firstFn := func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
		return ErrSkipThisAddon
	}
	secondFn := func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
		secondCalled = true
		return nil
	}

	fns := []setFn{firstFn, secondFn}
	err := invokeAll(nil, "test", "false", fns, nil)

	if !errors.Is(err, ErrSkipThisAddon) {
		t.Errorf("expected ErrSkipThisAddon, got %v", err)
	}
	if secondCalled {
		t.Errorf("expected second callback NOT to be called after ErrSkipThisAddon")
	}
}

func TestRunCallbacksMultiCallback(t *testing.T) {
	firstErr := fmt.Errorf("first callback failed")
	secondErr := fmt.Errorf("second callback failed")

	var firstCalled, secondCalled bool
	testAddon := &Addon{
		name: "test-multicallback-addon",
		callbacks: []setFn{
			func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
				firstCalled = true
				return firstErr
			},
			func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
				secondCalled = true
				return secondErr
			},
		},
	}

	Addons = append(Addons, testAddon)
	defer func() {
		Addons = Addons[:len(Addons)-1]
	}()

	cc := &config.ClusterConfig{Name: "test-profile"}

	t.Run("enable fails fast on first callback error", func(t *testing.T) {
		firstCalled = false
		secondCalled = false

		err := RunCallbacks(cc, testAddon.name, "true", nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !firstCalled {
			t.Errorf("expected first callback to be called")
		}
		if secondCalled {
			t.Errorf("expected second callback NOT to be called on enable failure")
		}
		if !errors.Is(err, firstErr) {
			t.Errorf("expected error %v to wrap %v", err, firstErr)
		}
	})

	t.Run("disable runs all callbacks despite errors", func(t *testing.T) {
		firstCalled = false
		secondCalled = false

		err := RunCallbacks(cc, testAddon.name, "false", nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !firstCalled {
			t.Errorf("expected first callback to be called")
		}
		if !secondCalled {
			t.Errorf("expected second callback to be called on disable failure")
		}
	})

	t.Run("enable with ErrSkipThisAddon returns sentinel unwrapped", func(t *testing.T) {
		skipAddon := &Addon{
			name: "test-skip-addon",
			callbacks: []setFn{
				func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
					return ErrSkipThisAddon
				},
				func(cc *config.ClusterConfig, name string, value string, options *run.CommandOptions) error {
					t.Errorf("expected second callback NOT to be called after ErrSkipThisAddon")
					return nil
				},
			},
		}
		Addons = append(Addons, skipAddon)
		defer func() {
			Addons = Addons[:len(Addons)-1]
		}()

		err := RunCallbacks(cc, skipAddon.name, "true", nil)
		if !errors.Is(err, ErrSkipThisAddon) {
			t.Errorf("expected ErrSkipThisAddon, got %v", err)
		}
	})
}

