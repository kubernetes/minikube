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

package machine

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestListMachines(t *testing.T) {
	const (
		numberOfValidMachines   = 2
		numberOfInValidMachines = 3
		totalNumberOfMachines   = numberOfValidMachines + numberOfInValidMachines
	)

	viper.Set(config.MachineProfile, "")

	testMinikubeDir := "./testdata/list-machines/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	validMachines, inValidMachines, err := List()

	if err != nil {
		t.Error(err)
	}

	if numberOfValidMachines != len(validMachines) {
		t.Errorf("expected %d valid machines, got %d", numberOfValidMachines, len(validMachines))
	}

	if numberOfInValidMachines != len(inValidMachines) {
		t.Errorf("expected %d invalid machines, got %d", numberOfInValidMachines, len(inValidMachines))
	}

	if totalNumberOfMachines != len(validMachines)+len(inValidMachines) {
		t.Errorf("expected %d total machines, got %d", totalNumberOfMachines, len(validMachines)+len(inValidMachines))
	}

	if numberOfMachineDirs != len(validMachines)+len(inValidMachines) {
		t.Error("expected number of machine directories to be equal to the number of total machines")
	}
}
