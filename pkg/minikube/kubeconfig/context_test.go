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
	"testing"
)

func TestDeleteContext(t *testing.T) {
	configFilename := tempFile(t, fakeKubeCfg)
	if err := DeleteContext(configFilename, "la-croix"); err != nil {
		t.Fatal(err)
	}

	cfg, err := readOrNew(configFilename)
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
	contextName := "minikube"

	kubeConfigFile, err := ioutil.TempFile("/tmp", "kubeconfig")
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}
	defer os.Remove(kubeConfigFile.Name())

	cfg, err := readOrNew(kubeConfigFile.Name())
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != "" {
		t.Errorf("Expected empty context but got %v", cfg.CurrentContext)
	}

	err = SetCurrentContext(kubeConfigFile.Name(), contextName)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}
	defer func() {
		err := UnsetCurrentContext(kubeConfigFile.Name(), contextName)
		if err != nil {
			t.Fatalf("Error not expected but got %v", err)
		}
	}()

	cfg, err = readOrNew(kubeConfigFile.Name())
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != contextName {
		t.Errorf("Expected context name %s but got %s", contextName, cfg.CurrentContext)
	}
}

func TestUnsetCurrentContext(t *testing.T) {
	kubeConfigFile := "./testdata/kubeconfig/config1"
	contextName := "minikube"

	cfg, err := readOrNew(kubeConfigFile)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != contextName {
		t.Errorf("Expected context name %s but got %s", contextName, cfg.CurrentContext)
	}

	err = UnsetCurrentContext(kubeConfigFile, contextName)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}
	defer func() {
		err := SetCurrentContext(kubeConfigFile, contextName)
		if err != nil {
			t.Fatalf("Error not expected but got %v", err)
		}
	}()

	cfg, err = readOrNew(kubeConfigFile)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != "" {
		t.Errorf("Expected empty context but got %v", cfg.CurrentContext)
	}
}

func TestUnsetCurrentContextOnlyChangesIfProfileIsTheCurrentContext(t *testing.T) {
	contextName := "minikube"
	kubeConfigFile := "./testdata/kubeconfig/config2"

	cfg, err := readOrNew(kubeConfigFile)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != contextName {
		t.Errorf("Expected context name %s but got %s", contextName, cfg.CurrentContext)
	}

	err = UnsetCurrentContext(kubeConfigFile, "differentContextName")
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	cfg, err = readOrNew(kubeConfigFile)
	if err != nil {
		t.Fatalf("Error not expected but got %v", err)
	}

	if cfg.CurrentContext != contextName {
		t.Errorf("Expected context name %s but got %s", contextName, cfg.CurrentContext)
	}
}
