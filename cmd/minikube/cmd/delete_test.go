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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine"
	"github.com/google/go-cmp/cmp"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"

	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// exclude returns a list of strings, minus the excluded ones
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
	fis, err := os.ReadDir(path)
	if err != nil {
		return result, err
	}
	for _, fi := range fis {
		result = append(result, fi.Name())
	}
	return result, nil
}

func TestDeleteProfile(t *testing.T) {
	td := t.TempDir()

	if err := copy.Copy("../../../pkg/minikube/config/testdata/delete-single", td); err != nil {
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
			t.Setenv(localpath.MinikubeHome, td)

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

var hostAndDirsDeleterMock = func(_ libmachine.API, _ *config.ClusterConfig, _ string) error { return deleteContextTest() }

func deleteContextTest() error {
	if err := cmdcfg.Unset(config.ProfileName); err != nil {
		return DeletionError{Err: fmt.Errorf("unset minikube profile: %v", err), Errtype: Fatal}
	}
	return nil
}

func TestDeleteAllProfiles(t *testing.T) {
	td := t.TempDir()

	if err := copy.Copy("../../../pkg/minikube/config/testdata/delete-all", td); err != nil {
		t.Fatalf("copy: %v", err)
	}

	t.Setenv(localpath.MinikubeHome, td)

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

	profiles := validProfiles
	profiles = append(profiles, inValidProfiles...)
	hostAndDirsDeleter = hostAndDirsDeleterMock
	errs := DeleteProfiles(profiles)

	if errs != nil {
		t.Errorf("errors while deleting all profiles: %v", errs)
	}

	afterProfiles, err := fileNames(filepath.Join(localpath.MiniPath(), "profiles"))
	if err != nil {
		t.Errorf("profiles: %v", err)
	}
	afterMachines, err := os.ReadDir(filepath.Join(localpath.MiniPath(), "machines"))
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

// TestTryKillOne spawns a go child process that waits to be SIGKILLed,
// then tries to execute the tryKillOne function on it;
// if after tryKillOne the process still exists, we consider it a failure
func TestTryKillOne(t *testing.T) {

	var waitForSig = []byte(`
package main

import (
	"os"
	"os/signal"
	"syscall"
)

// This is used to unit test functions that send termination
// signals to processes, in a cross-platform way.
func main() {
	ch := make(chan os.Signal, 1)
	done := make(chan struct{})
	defer close(ch)

	signal.Notify(ch, syscall.SIGHUP)
	defer signal.Stop(ch)

	go func() {
		<-ch
		close(done)
	}()

	<-done
}
`)
	td := t.TempDir()
	tmpfile := filepath.Join(td, "waitForSig.go")

	if err := os.WriteFile(tmpfile, waitForSig, 0o600); err != nil {
		t.Fatalf("copying source to %s: %v\n", tmpfile, err)
	}

	processToKill := exec.Command("go", "run", tmpfile)
	err := processToKill.Start()
	if err != nil {
		t.Fatalf("while execing child process: %v\n", err)
	}
	pid := processToKill.Process.Pid

	isMinikubeProcess = func(int) (bool, error) {
		return true, nil
	}

	err = trySigKillProcess(pid)
	if err != nil {
		t.Fatalf("while trying to kill child proc %d: %v\n", pid, err)
	}

	// waiting for process to exit
	if err := processToKill.Wait(); !strings.Contains(err.Error(), "killed") {
		t.Fatalf("unable to kill process: %v\n", err)
	}
}
