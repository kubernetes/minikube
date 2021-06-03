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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/machine/libmachine"
	"github.com/google/go-cmp/cmp"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"

	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// except returns a list of strings, minus the excluded ones
func exclude(vals []string, exclude []string) []string {
	result := []string{}
	for _, v := range vals {
		excluded := false
		for _, e := range exclude {
			if e == v {
				excluded = true
				continue
			}
		}
		if !excluded {
			result = append(result, v)
		}
	}
	return result
}

func fileNames(path string) ([]string, error) {
	result := []string{}
	fis, err := ioutil.ReadDir(path)
	if err != nil {
		return result, err
	}
	for _, fi := range fis {
		result = append(result, fi.Name())
	}
	return result, nil
}

func TestDeleteProfile(t *testing.T) {
	td, err := ioutil.TempDir("", "single")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}

	t.Cleanup(func() {
		err := os.RemoveAll(td)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", td)
		}
	})

	err = copy.Copy("../../../pkg/minikube/config/testdata/delete-single", td)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}

	tests := []struct {
		name     string
		profile  string
		expected []string
	}{
		{"normal", "p1", []string{"p1"}},
		{"empty-profile", "p2_empty_profile_config", []string{"p2_empty_profile_config"}},
		{"invalid-profile", "p3_invalid_profile_config", []string{"p3_invalid_profile_config"}},
		{"partial-profile", "p4_partial_profile_config", []string{"p4_partial_profile_config"}},
		{"missing-mach", "p5_missing_machine_config", []string{"p5_missing_machine_config"}},
		{"empty-mach", "p6_empty_machine_config", []string{"p6_empty_machine_config"}},
		{"invalid-mach", "p7_invalid_machine_config", []string{"p7_invalid_machine_config"}},
		{"partial-mach", "p8_partial_machine_config", []string{"p8_partial_machine_config"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = os.Setenv(localpath.MinikubeHome, td)
			if err != nil {
				t.Errorf("setenv: %v", err)
			}

			beforeProfiles, err := fileNames(filepath.Join(localpath.MiniPath(), "profiles"))
			if err != nil {
				t.Errorf("readdir: %v", err)
			}
			beforeMachines, err := fileNames(filepath.Join(localpath.MiniPath(), "machines"))
			if err != nil {
				t.Errorf("readdir: %v", err)
			}

			profile, err := config.LoadProfile(tt.profile)
			if err != nil {
				t.Logf("load failure: %v", err)
			}

			hostAndDirsDeleter = hostAndDirsDeleterMock
			errs := DeleteProfiles([]*config.Profile{profile})
			if len(errs) > 0 {
				HandleDeletionErrors(errs)
				t.Errorf("Errors while deleting profiles: %v", errs)
			}
			pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
			if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
				t.Errorf("Profile folder of profile \"%s\" was not deleted", profile.Name)
			}

			pathToMachine := localpath.MachinePath(profile.Name, localpath.MiniPath())
			if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
				t.Errorf("Profile folder of profile \"%s\" was not deleted", profile.Name)
			}

			afterProfiles, err := fileNames(filepath.Join(localpath.MiniPath(), "profiles"))
			if err != nil {
				t.Errorf("readdir profiles: %v", err)
			}

			afterMachines, err := fileNames(filepath.Join(localpath.MiniPath(), "machines"))
			if err != nil {
				t.Errorf("readdir machines: %v", err)
			}

			expectedProfiles := exclude(beforeProfiles, tt.expected)
			if diff := cmp.Diff(expectedProfiles, afterProfiles); diff != "" {
				t.Errorf("profiles mismatch (-want +got):\n%s", diff)
			}

			expectedMachines := exclude(beforeMachines, tt.expected)
			if diff := cmp.Diff(expectedMachines, afterMachines); diff != "" {
				t.Errorf("machines mismatch (-want +got):\n%s", diff)
			}

			viper.Set(config.ProfileName, "")
		})
	}
}

var hostAndDirsDeleterMock = func(api libmachine.API, cc *config.ClusterConfig, profileName string) error {
	return deleteContextTest()
}

func deleteContextTest() error {
	if err := cmdcfg.Unset(config.ProfileName); err != nil {
		return DeletionError{Err: fmt.Errorf("unset minikube profile: %v", err), Errtype: Fatal}
	}
	return nil
}

func TestDeleteAllProfiles(t *testing.T) {
	td, err := ioutil.TempDir("", "all")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	defer func() { // clean up tempdir
		err := os.RemoveAll(td)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", td)
		}
	}()

	err = copy.Copy("../../../pkg/minikube/config/testdata/delete-all", td)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}

	err = os.Setenv(localpath.MinikubeHome, td)
	if err != nil {
		t.Errorf("error setting up test environment. could not set %s", localpath.MinikubeHome)
	}

	pFiles, err := fileNames(filepath.Join(localpath.MiniPath(), "profiles"))
	if err != nil {
		t.Errorf("filenames: %v", err)
	}
	mFiles, err := fileNames(filepath.Join(localpath.MiniPath(), "machines"))
	if err != nil {
		t.Errorf("filenames: %v", err)
	}

	const numberOfTotalProfileDirs = 8
	if numberOfTotalProfileDirs != len(pFiles) {
		t.Errorf("got %d test profiles, expected %d: %s", len(pFiles), numberOfTotalProfileDirs, pFiles)
	}
	const numberOfTotalMachineDirs = 7
	if numberOfTotalMachineDirs != len(mFiles) {
		t.Errorf("got %d test machines, expected %d: %s", len(mFiles), numberOfTotalMachineDirs, mFiles)
	}

	config.DockerContainers = func() ([]string, error) {
		return []string{}, nil
	}
	validProfiles, inValidProfiles, err := config.ListProfiles()
	if err != nil {
		t.Error(err)
	}

	if numberOfTotalProfileDirs != len(validProfiles)+len(inValidProfiles) {
		t.Errorf("ListProfiles length = %d, expected %d\nvalid: %v\ninvalid: %v\n", len(validProfiles)+len(inValidProfiles), numberOfTotalProfileDirs, validProfiles, inValidProfiles)
	}

	profiles := append(validProfiles, inValidProfiles...)
	hostAndDirsDeleter = hostAndDirsDeleterMock
	errs := DeleteProfiles(profiles)

	if errs != nil {
		t.Errorf("errors while deleting all profiles: %v", errs)
	}

	afterProfiles, err := fileNames(filepath.Join(localpath.MiniPath(), "profiles"))
	if err != nil {
		t.Errorf("profiles: %v", err)
	}
	afterMachines, err := ioutil.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
	if err != nil {
		t.Errorf("machines: %v", err)
	}
	if len(afterProfiles) != 0 {
		t.Errorf("Did not delete all profiles, remaining: %v", afterProfiles)
	}

	if len(afterMachines) != 0 {
		t.Errorf("Did not delete all machines, remaining: %v", afterMachines)
	}

	viper.Set(config.ProfileName, "")
}
