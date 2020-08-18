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

package kubeconfig

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteContext(t *testing.T) {
	// See kubeconfig_test
	fn := tempFile(t, kubeConfigWithoutHTTPS)
	defer os.Remove(fn)
	if err := DeleteContext("la-croix", fn); err != nil {
		t.Fatal(err)
	}

	cfg, err := readOrNew(fn)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.AuthInfos) != 0 {
		t.Fail()
	}

	if len(cfg.Clusters) != 0 {
		t.Fail()
	}

	if len(cfg.Contexts) != 0 {
		t.Fail()
	}
}

func TestSetCurrentContext(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "kubeconfig")
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}
	defer os.Remove(f.Name())

	kcfg, err := readOrNew(f.Name())
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if kcfg.CurrentContext != "" {
		t.Errorf("Expected empty context but got %v", kcfg.CurrentContext)
	}

	contextName := "minikube"
	err = SetCurrentContext(contextName, f.Name())
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}
	defer func() {
		err := UnsetCurrentContext(contextName, f.Name())
		if err != nil {
			t.Fatalf("Error not expected but got %v", err)
		}
	}()

	kcfg, err = readOrNew(f.Name())
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}
	if kcfg.CurrentContext != contextName {
		t.Errorf("Expected context name %s but got %v : ", contextName, kcfg.CurrentContext)
	}
}

func TestUnsetCurrentContext(t *testing.T) {
	fn := filepath.Join("testdata", "kubeconfig", "config1")
	contextName := "minikube"

	cfg, err := readOrNew(fn)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != contextName {
		t.Errorf("Expected context name %s but got %s", contextName, cfg.CurrentContext)
	}

	err = UnsetCurrentContext(contextName, fn)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}
	defer func() {
		err := SetCurrentContext(contextName, fn)
		if err != nil {
			t.Fatalf("Error not expected but got %v", err)
		}
	}()

	cfg, err = readOrNew(fn)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != "" {
		t.Errorf("Expected empty context but got %v", cfg.CurrentContext)
	}
}

func TestUnsetCurrentContextOnlyChangesIfProfileIsTheCurrentContext(t *testing.T) {
	contextName := "minikube"

	fn := filepath.Join("testdata", "kubeconfig", "config2")
	cfg, err := readOrNew(fn)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != contextName {
		t.Errorf("Expected context name %s but got %s", contextName, cfg.CurrentContext)
	}

	err = UnsetCurrentContext("differentContextName", fn)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	cfg, err = readOrNew(fn)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != contextName {
		t.Errorf("Expected context name %s but got %s", contextName, cfg.CurrentContext)
	}
}
