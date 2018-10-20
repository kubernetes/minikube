// +build integration

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

package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func testMounting(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "darwin" {
		t.Skip("mount tests disabled in darwin due to timeout (issue#3200)")
	}
	if strings.Contains(*args, "--vm-driver=none") {
		t.Skip("skipping test for none driver as it does not need mount")
	}
	minikubeRunner := NewMinikubeRunner(t)

	tempDir, err := ioutil.TempDir("", "mounttest")
	if err != nil {
		t.Fatalf("Unexpected error while creating tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mountCmd := fmt.Sprintf("mount %s:/mount-9p", tempDir)
	cmd, _ := minikubeRunner.RunDaemon(mountCmd)
	defer func() {
		err := cmd.Process.Kill()
		if err != nil {
			t.Logf("Failed to kill mount command: %v", err)
		}
	}()

	kubectlRunner := util.NewKubectlRunner(t)
	podName := "busybox-mount"
	podPath, _ := filepath.Abs("testdata/busybox-mount-test.yaml")

	// Write file in mounted dir from host
	expected := "test\n"
	files := []string{"fromhost", "fromhostremove"}
	for _, file := range files {
		path := filepath.Join(tempDir, file)
		err = ioutil.WriteFile(path, []byte(expected), 0644)
		if err != nil {
			t.Fatalf("Unexpected error while writing file %s: %v", path, err)
		}
	}

	// Create the pods we need outside the main test loop.
	setupTest := func() error {
		if _, err := kubectlRunner.RunCommand([]string{"create", "-f", podPath}); err != nil {
			return err
		}
		return nil
	}
	defer func() {
		if out, err := kubectlRunner.RunCommand([]string{"delete", "-f", podPath}); err != nil {
			t.Logf("delete -f %s failed: %v\noutput: %s\n", podPath, err, out)
		}
	}()

	if err := util.Retry(t, setupTest, 5*time.Second, 40); err != nil {
		t.Fatal("mountTest failed with error:", err)
	}

	client, err := pkgutil.GetClient()
	if err != nil {
		t.Fatalf("getting kubernetes client: %v", err)
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"integration-test": "busybox-mount"}))
	if err := pkgutil.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		t.Fatalf("Error waiting for busybox mount pod to be up: %v", err)
	}

	mountTest := func() error {
		path := filepath.Join(tempDir, "frompod")
		out, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		// test that file written from pod can be read from host echo test > /mount-9p/frompod; in pod
		if string(out) != expected {
			t.Fatalf("Expected file %s to contain text %s, was %s.", path, expected, out)
		}

		// test that file written from host was read in by the pod via cat /mount-9p/fromhost;
		if out, err = kubectlRunner.RunCommand([]string{"logs", podName}); err != nil {
			return err
		}
		if string(out) != expected {
			t.Fatalf("Expected file %s to contain text %s, was %s.", path, expected, out)
		}

		// test that fromhostremove was deleted by the pod from the mount via rm /mount-9p/fromhostremove
		path = filepath.Join(tempDir, "fromhostremove")
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("Expected file %s to be removed", path)
		}

		// test that frompodremove can be deleted on the host
		path = filepath.Join(tempDir, "frompodremove")
		if err := os.Remove(path); err != nil {
			t.Fatalf("Unexpected error removing file %s: %v", path, err)
		}

		return nil
	}
	if err := util.Retry(t, mountTest, 5*time.Second, 40); err != nil {
		t.Fatalf("mountTest failed with error: %v", err)
	}

}
