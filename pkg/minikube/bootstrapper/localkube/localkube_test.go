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

package localkube

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/tests"
)

func TestIsLocalkubeCached(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	inputArr := [...]string{
		"v1.3.3",
		"1.3.0",
		"http://test-url.localkube.com/localkube-binary",
		"file:///test/dir/to/localkube-binary",
	}

	localkubeCacher := localkubeCacher{
		k8sConf: bootstrapper.KubernetesConfig{},
	}

	inner := func(input string) {
		localkubeCacher.k8sConf = bootstrapper.KubernetesConfig{
			KubernetesVersion: input,
		}
		if localkubeCacher.isLocalkubeCached() {
			t.Errorf("IsLocalKubeCached returned true even though %s was not cached",
				localkubeCacher.getLocalkubeCacheFilepath())
			return
		}

		f, err := os.Create(localkubeCacher.getLocalkubeCacheFilepath())
		if err != nil {
			t.Errorf("failed to create dummy cache file: %v", err)
			return
		}
		f.Close()
		defer os.Remove(f.Name())
		if !localkubeCacher.isLocalkubeCached() {
			t.Errorf("IsLocalKubeCached returned false even though %s was cached",
				localkubeCacher.getLocalkubeCacheFilepath())
		}

	}
	for _, input := range inputArr {
		inner(input)
	}
}

func TestHostGetLogs(t *testing.T) {
	api, s := setupSSHServerAndAPI(t)
	b, err := NewLocalkubeBootstrapper(api)
	if err != nil {
		t.Fatalf("error getting bootstrapper: %s", err)
	}

	tests := []struct {
		description string
		follow      bool
	}{
		{
			description: "logs -f",
			follow:      true,
		},
		// TODO(r2d4): fix this test
		// {
		// 	description: "logs",
		// 	follow:      false,
		// },
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cmd, err := GetLogsCommand(test.follow)
			if err != nil {
				t.Errorf("Error getting the logs command: %s", err)
			}

			if _, err = b.GetClusterLogs(test.follow); err != nil {
				t.Errorf("Error getting host logs: %s", err)
			}
			if _, ok := s.Commands[cmd]; !ok {
				t.Errorf("Expected command to run but did not: %s", cmd)
			}
		})
	}
}

func TestGetLocalkubeStatus(t *testing.T) {
	api, s := setupSSHServerAndAPI(t)

	s.SetCommandToOutput(map[string]string{
		localkubeStatusCommand: state.Running.String(),
	})

	b, err := NewLocalkubeBootstrapper(api)
	if err != nil {
		t.Fatalf("error getting bootstrapper: %s", err)
	}

	if _, err := b.GetClusterStatus(); err != nil {
		t.Fatalf("Error getting localkube status: %s", err)
	}

	s.SetCommandToOutput(map[string]string{
		localkubeStatusCommand: state.Stopped.String(),
	})
	if _, err := b.GetClusterStatus(); err != nil {
		t.Fatalf("Error getting localkube status: %s", err)
	}

	s.SetCommandToOutput(map[string]string{
		localkubeStatusCommand: "Bad Output",
	})
	if _, err := b.GetClusterStatus(); err == nil {
		t.Fatalf("Expected error in getting localkube status as ssh returned bad output")
	}
}

func TestStartCluster(t *testing.T) {
	api, s := setupSSHServerAndAPI(t)
	lk, err := NewLocalkubeBootstrapper(api)
	if err != nil {
		t.Fatalf("Error getting bootstrapper: %s", err)
	}

	err = lk.StartCluster(bootstrapper.KubernetesConfig{})
	if err != nil {
		t.Fatalf("Error starting cluster: %s", err)
	}

	startCommand, err := GetStartCommand(bootstrapper.KubernetesConfig{})
	if err != nil {
		t.Fatalf("Error getting start command: %s", err)
	}
	for _, cmd := range []string{startCommand} {
		if _, ok := s.Commands[cmd]; !ok {
			t.Fatalf("Expected command not run: %s. Commands run: %v", cmd, s.Commands)
		}
	}
}

func setupSSHServerAndAPI(t *testing.T) (*tests.MockAPI, *tests.SSHServer) {
	api := tests.NewMockAPI()
	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}

	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	api.Hosts[config.GetMachineName()] = &host.Host{Driver: d}

	return api, s
}

func TestUpdateDefault(t *testing.T) {
	api, s := setupSSHServerAndAPI(t)

	kubernetesConfig := bootstrapper.KubernetesConfig{
		KubernetesVersion: constants.DefaultKubernetesVersion,
	}

	lk, err := NewLocalkubeBootstrapper(api)
	if err != nil {
		t.Fatalf("Error getting bootstrapper: %s", err)
	}

	if err := lk.UpdateCluster(kubernetesConfig); err != nil {
		t.Fatalf("Error updating cluster: %s", err)
	}
	transferred := s.Transfers.Bytes()

	for _, addonBundle := range assets.Addons {
		if isEnabled, err := addonBundle.IsEnabled(); err == nil && isEnabled {
			for _, addon := range addonBundle.Assets {
				contents, _ := assets.Asset(addon.GetAssetName())
				if !bytes.Contains(transferred, contents) {
					t.Fatalf("File not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
				}
			}
		} else if err != nil {
			t.Fatalf("File not copied. Unexpected error while attempting to check transferred addons: %s", err)
		}
	}

	//test that localkube is transferred properly
	contents, _ := assets.Asset("out/localkube")
	if !bytes.Contains(transferred, contents) {
		t.Fatalf("File not copied. Expected transfers to contain: %s. It was: %s", contents, transferred)
	}
}

var testLocalkubeBin = "hello"

type K8sVersionHandlerCorrect struct{}

func (h *K8sVersionHandlerCorrect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, testLocalkubeBin)
}

func TestUpdateCustomAddons(t *testing.T) {
	tempDir := tests.MakeTempDir()
	os.Mkdir(constants.MakeMiniPath("addons", "subdir"), 0777)
	defer os.RemoveAll(tempDir)

	api, s := setupSSHServerAndAPI(t)

	//write a file into ~/.minikube/addons
	path := filepath.Join(constants.MakeMiniPath("addons"), "dir-addon.yaml")
	testContent1 := []byte("CUSTOM ADDON TEST STRING#1, In Addons Dir")
	err := ioutil.WriteFile(path, testContent1, 0644)
	if err != nil {
		t.Fatalf("Error writing custom addon: %s", err)
	}

	path = filepath.Join(constants.MakeMiniPath("addons", "subdir"), "subdir-addon.yaml")
	testContent2 := []byte("CUSTOM ADDON TEST STRING#2, In Addons SubDir")
	err = ioutil.WriteFile(path, testContent2, 0644)
	if err != nil {
		t.Fatalf("Error writing custom addon: %s", err)
	}

	//run update
	kubernetesConfig := bootstrapper.KubernetesConfig{
		KubernetesVersion: constants.DefaultKubernetesVersion,
	}
	lk, err := NewLocalkubeBootstrapper(api)
	if err != nil {
		t.Fatalf("Error getting bootstrapper: %s", err)
	}
	if err := lk.UpdateCluster(kubernetesConfig); err != nil {
		t.Fatalf("Error updating cluster: %s", err)
	}
	transferred := s.Transfers.Bytes()

	//test that custom addons are transferred properly
	if !bytes.Contains(transferred, testContent1) {
		t.Fatalf("Custom addon not copied. Expected transfers to contain custom addon with content: %s. It was: %s", testContent1, transferred)
	}

	if !bytes.Contains(transferred, testContent2) {
		t.Fatalf("Custom addon not copied. Expected transfers to contain custom addon with content: %s. It was: %s", testContent2, transferred)
	}
}
