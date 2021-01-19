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

	"github.com/hashicorp/go-getter"
	"k8s.io/minikube/pkg/util"
)

const newPath = "../../out/minikube"

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
	t.Log(url)
	err = getter.GetFile(oldPath, url)
	if err != nil {
		t.Fatalf("failed download minikube %s: %v", *upgradeFrom, err)
	}

	if runtime.GOOS != "windows" {
		os.Chmod(oldPath, 0755)
	}

	for i := 1; i <= *loops; i++ {
		t.Logf("Loop %d of %d: %s to HEAD", i, *loops, *upgradeFrom)
		runStress(t, oldPath, profile, i)
	}
}

// This run the guts of the actual test
func runStress(t *testing.T, oldPath string, profile string, i int) {
	// Cleanup old runs
	exec.Command(newPath, "delete", "-p", profile).Run()

	t.Logf("Hot upgrade from %s to HEAD", *upgradeFrom)
	oldStart := exec.Command(oldPath, "start", "-p", profile, *startArgs, "--alsologtostderr")
	t.Logf(oldStart.String())
	err := oldStart.Run()
	if err != nil {
		t.Logf("old start failed, which is OK: %v", err)
	}

	newStart := exec.Command(newPath, "start", "-p", profile, *startArgs, "--alsologtostderr")
	err = newStart.Run()
	if err != nil {
		t.Errorf("***hot upgrade loop %d FAILED with %s: %v", i, *startArgs, err)
	}

	delete := exec.Command(newPath, "delete", "-p", profile)
	delete.Run()

	t.Logf("Cold upgrade from %s to HEAD", *upgradeFrom)
	oldStart2 := exec.Command(oldPath, "start", "-p", profile, *startArgs, "--alsologtostderr")
	err = oldStart2.Run()
	if err != nil {
		t.Logf("old start failed, which is OK: %v", err)
	}

	oldStop := exec.Command(oldPath, "stop", "-p", profile)
	oldStop.Run()

	err = exec.Command(newPath, "start", "-p", profile, *startArgs, "--alsologtostderr").Run()
	if err != nil {
		t.Errorf("***cold upgrade loop %d FAILED with %s: %v", i, *startArgs, err)
	}

	t.Logf("Restart HEAD test")
	err = exec.Command(newPath, "start", "-p", profile, *startArgs, "--alsologtostderr").Run()
	if err != nil {
		t.Errorf("***hot restart loop %d FAILED with %s: %v", i, *startArgs, err)
	}

	t.Logf("Cold HEAD restart")
	newStop := exec.Command(newPath, "stop", "-p", profile)
	newStop.Run()

	err = exec.Command(newPath, "start", "-p", profile, *startArgs, "--alsologtostderr").Run()
	if err != nil {
		t.Errorf("***cold restart loop %d FAILED with %s: %v", i, *startArgs, err)
	}

	exec.Command(newPath, "delete", "-p", profile).Run()

	t.Logf("Loop %d of %d done.", i, *loops)
}
