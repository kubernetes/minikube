/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package stress

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/hashicorp/go-getter"
	"k8s.io/minikube/pkg/util"
)

const newPath = "../../out/minikube"
const cmdTimeout = int64(90 * 1000)

// TestStress runs the stress test
func TestStress(t *testing.T) {
	profile := "stress"
	if *startArgs != "" {
		h := md5.Sum([]byte(*startArgs))
		profile += hex.EncodeToString(h[:])
	}

	_, err := os.Stat(newPath)

	if os.IsNotExist(err) {
		t.Fatalf("latest minikube binary is missing, run make")
	}

	oldPath := fmt.Sprintf("../../out/minikube-%s", *upgradeFrom)
	url := util.GetBinaryDownloadURL(*upgradeFrom, runtime.GOOS)
	t.Logf("Downloading minikube %s from %s", *upgradeFrom, url)
	err = getter.GetFile(oldPath, url)
	if err != nil {
		t.Fatalf("failed download minikube %s: %v", *upgradeFrom, err)
	}

	if runtime.GOOS != "windows" {
		err := os.Chmod(oldPath, 0755)
		if err != nil {
			t.Fatalf("Failed to chmod old binary: %v", err)
		}
	}

	for i := 1; i <= *loops; i++ {
		t.Logf("Loop %d of %d: %s to HEAD", i, *loops, *upgradeFrom)
		runStress(t, oldPath, profile, i)
	}
}

// This run the guts of the actual test
func runStress(t *testing.T, oldPath string, profile string, i int) {
	// Cleanup old runs
	runCommand(t, false, newPath, "delete", "-p", profile)

	t.Logf("Hot upgrade from %s to HEAD", *upgradeFrom)
	runCommand(t, true, oldPath, "start", "-p", profile, *startArgs, "--alsologtostderr")

	runCommand(t, true, newPath, "start", "-p", profile, *startArgs, "--alsologtostderr")

	runCommand(t, false, newPath, "delete", "-p", profile)

	t.Logf("Cold upgrade from %s to HEAD", *upgradeFrom)
	runCommand(t, false, oldPath, "start", "-p", profile, *startArgs, "--alsologtostderr")

	runCommand(t, false, oldPath, "stop", "-p", profile)

	runCommand(t, true, newPath, "start", "-p", profile, *startArgs, "--alsologtostderr")

	t.Logf("Restart HEAD test")
	runCommand(t, true, newPath, "start", "-p", profile, *startArgs, "--alsologtostderr")

	t.Logf("Cold HEAD restart")
	runCommand(t, false, newPath, "stop", "-p", profile)

	runCommand(t, true, newPath, "start", "-p", profile, *startArgs, "--alsologtostderr")

	runCommand(t, false, newPath, "delete", "-p", profile)

	t.Logf("Loop %d of %d done.", i, *loops)
}

func runCommand(t *testing.T, errorOut bool, mkPath string, args ...string) {
	c := exec.Command(mkPath, args...)
	t.Logf("Running: %s, %v", mkPath, args)
	start := time.Now()
	out, err := c.CombinedOutput()
	if err != nil {
		if errorOut {
			t.Errorf("Error running %s %v: %v", mkPath, args, err)
			t.Logf("Command output: %q", out)
		} else {
			t.Logf("Error running %s %v: %v", mkPath, args, err)
		}
		return
	}
	elapsed := time.Since(start)
	if elapsed.Milliseconds() > cmdTimeout && errorOut {
		t.Errorf("Command took too long: %s %v took %f seconds", mkPath, args, elapsed.Seconds())
	}
}
