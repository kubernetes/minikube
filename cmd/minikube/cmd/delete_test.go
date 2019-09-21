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

package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestDeleteProfileWithValidConfig(t *testing.T) {
	testMinikubeDir := "../../../pkg/minikube/config/testdata/delete-single/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs := len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	profileToDelete := "p1"
	profile, _ := config.LoadProfile(profileToDelete)

	errs := DeleteProfiles([]*config.Profile{profile})

	if len(errs) > 0 {
		HandleDeletionErrors(errs)
		t.Fatal("Errors while deleting profiles")
	}

	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	pathToMachine := cluster.MachinePath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles")); len(files) != (numberOfProfileDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines")); len(files) != (numberOfMachineDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	viper.Set(config.MachineProfile, "")
}

func TestDeleteProfileWithEmptyProfileConfig(t *testing.T) {
	testMinikubeDir := "../../../pkg/minikube/config/testdata/delete-single/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs := len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	profileToDelete := "p2_empty_profile_config"
	profile, _ := config.LoadProfile(profileToDelete)

	errs := DeleteProfiles([]*config.Profile{profile})

	if len(errs) > 0 {
		HandleDeletionErrors(errs)
		t.Fatal("Errors while deleting profiles")
	}

	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	pathToMachine := cluster.MachinePath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles")); len(files) != (numberOfProfileDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines")); len(files) != (numberOfMachineDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	viper.Set(config.MachineProfile, "")
}

func TestDeleteProfileWithInvalidProfileConfig(t *testing.T) {
	testMinikubeDir := "../../../pkg/minikube/config/testdata/delete-single/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs := len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	profileToDelete := "p3_invalid_profile_config"
	profile, _ := config.LoadProfile(profileToDelete)

	errs := DeleteProfiles([]*config.Profile{profile})

	if len(errs) > 0 {
		HandleDeletionErrors(errs)
		t.Fatal("Errors while deleting profiles")
	}

	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	pathToMachine := cluster.MachinePath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles")); len(files) != (numberOfProfileDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines")); len(files) != (numberOfMachineDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	viper.Set(config.MachineProfile, "")
}

func TestDeleteProfileWithPartialProfileConfig(t *testing.T) {
	testMinikubeDir := "../../../pkg/minikube/config/testdata/delete-single/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs := len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	profileToDelete := "p4_partial_profile_config"
	profile, _ := config.LoadProfile(profileToDelete)

	errs := DeleteProfiles([]*config.Profile{profile})

	if len(errs) > 0 {
		HandleDeletionErrors(errs)
		t.Fatal("Errors while deleting profiles")
	}

	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	pathToMachine := cluster.MachinePath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles")); len(files) != (numberOfProfileDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines")); len(files) != (numberOfMachineDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	viper.Set(config.MachineProfile, "")
}

func TestDeleteProfileWithMissingMachineConfig(t *testing.T) {
	testMinikubeDir := "../../../pkg/minikube/config/testdata/delete-single/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs := len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	profileToDelete := "p5_missing_machine_config"
	profile, _ := config.LoadProfile(profileToDelete)

	errs := DeleteProfiles([]*config.Profile{profile})

	if len(errs) > 0 {
		HandleDeletionErrors(errs)
		t.Fatal("Errors while deleting profiles")
	}

	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	pathToMachine := cluster.MachinePath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles")); len(files) != (numberOfProfileDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines")); len(files) != numberOfMachineDirs {
		t.Fatal("Deleted a machine config when it should not")
	}

	viper.Set(config.MachineProfile, "")
}

func TestDeleteProfileWithEmptyMachineConfig(t *testing.T) {
	testMinikubeDir := "../../../pkg/minikube/config/testdata/delete-single/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs := len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	profileToDelete := "p6_empty_machine_config"
	profile, _ := config.LoadProfile(profileToDelete)

	errs := DeleteProfiles([]*config.Profile{profile})

	if len(errs) > 0 {
		HandleDeletionErrors(errs)
		t.Fatal("Errors while deleting profiles")
	}

	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	pathToMachine := cluster.MachinePath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles")); len(files) != (numberOfProfileDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines")); len(files) != (numberOfMachineDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	viper.Set(config.MachineProfile, "")
}

func TestDeleteProfileWithInvalidMachineConfig(t *testing.T) {
	testMinikubeDir := "../../../pkg/minikube/config/testdata/delete-single/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs := len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	profileToDelete := "p7_invalid_machine_config"
	profile, _ := config.LoadProfile(profileToDelete)

	errs := DeleteProfiles([]*config.Profile{profile})

	if len(errs) > 0 {
		HandleDeletionErrors(errs)
		t.Fatal("Errors while deleting profiles")
	}

	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	pathToMachine := cluster.MachinePath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles")); len(files) != (numberOfProfileDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines")); len(files) != (numberOfMachineDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	viper.Set(config.MachineProfile, "")
}

func TestDeleteProfileWithPartialMachineConfig(t *testing.T) {
	testMinikubeDir := "../../../pkg/minikube/config/testdata/delete-single/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs := len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	profileToDelete := "p8_partial_machine_config"
	profile, _ := config.LoadProfile(profileToDelete)

	errs := DeleteProfiles([]*config.Profile{profile})

	if len(errs) > 0 {
		HandleDeletionErrors(errs)
		t.Fatal("Errors while deleting profiles")
	}

	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	pathToMachine := cluster.MachinePath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		t.Fatalf("Profile folder of profile \"%s\" was not deleted", profile.Name)
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles")); len(files) != (numberOfProfileDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	if files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines")); len(files) != (numberOfMachineDirs - 1) {
		t.Fatal("Did not delete exactly one profile")
	}

	viper.Set(config.MachineProfile, "")
}

func TestDeleteAllProfiles(t *testing.T) {
	const numberOfTotalProfileDirs = 8
	const numberOfTotalMachineDirs = 7

	testMinikubeDir := "../../../pkg/minikube/config/testdata/delete-all/.minikube"
	miniDir, err := filepath.Abs(testMinikubeDir)

	if err != nil {
		t.Errorf("error getting dir path for %s : %v", testMinikubeDir, err)
	}

	err = os.Setenv(localpath.MinikubeHome, miniDir)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	files, _ := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs := len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs := len(files)

	if numberOfTotalProfileDirs != numberOfProfileDirs {
		t.Error("invalid testdata")
	}

	if numberOfTotalMachineDirs != numberOfMachineDirs {
		t.Error("invalid testdata")
	}

	validProfiles, inValidProfiles, err := config.ListProfiles()

	if err != nil {
		t.Error(err)
	}

	if numberOfTotalProfileDirs != len(validProfiles)+len(inValidProfiles) {
		t.Error("invalid testdata")
	}

	profiles := append(validProfiles, inValidProfiles...)
	errs := DeleteProfiles(profiles)

	if errs != nil {
		t.Errorf("errors while deleting all profiles: %v", errs)
	}

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "profiles"))
	numberOfProfileDirs = len(files)

	files, _ = ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	numberOfMachineDirs = len(files)

	if numberOfProfileDirs != 0 {
		t.Errorf("Did not delete all profiles: still %d profiles left", numberOfProfileDirs)
	}

	if numberOfMachineDirs != 0 {
		t.Errorf("Did not delete all profiles: still %d machines left", numberOfMachineDirs)
	}

	viper.Set(config.MachineProfile, "")
}
