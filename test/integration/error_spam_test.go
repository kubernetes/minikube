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
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// stderrAllow are regular expressions acceptable to find in normal stderr
var stderrAllow = []string{
	// kubectl out of date warning
	`kubectl`,
	// slow docker warning
	`slow|long time|Restarting the docker service may improve`,
	// don't care if we can't push images to other profiles
	`cache_images.go:.*error getting status`,
	// don't care if we can't push images to other profiles which are deleted.
	`cache_images.go:.*Failed to load profile`,
	// ! 'docker' driver reported a issue that could affect the performance."
	`docker.*issue.*performance`,
	// "* Suggestion: enable overlayfs kernel module on your Linux"
	`Suggestion.*overlayfs`,
	// jenkins VMs (debian 9) cgoups don't allow setting memory
	`Your cgroup does not allow setting memory.`,
}

// stderrAllowRe combines rootCauses into a single regex
var stderrAllowRe = regexp.MustCompile(strings.Join(stderrAllow, "|"))

// TestErrorSpam asserts that there are no errors displayed in UI.
func TestErrorSpam(t *testing.T) {
	if NoneDriver() {
		t.Skip("none driver always shows a warning")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("nospam")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(25))
	defer CleanupWithLogs(t, profile, cancel)

	// This should likely use multi-node once it's ready
	args := append([]string{"start", "-p", profile, "-n=1", "--memory=2250", "--wait=false"}, StartArgs()...)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("%q failed: %v", rr.Command(), err)
	}

	stdout := rr.Stdout.String()
	stderr := rr.Stderr.String()

	for _, line := range strings.Split(stderr, "\n") {
		if stderrAllowRe.MatchString(line) {
			t.Logf("acceptable stderr: %q", line)
			continue
		}

		if len(strings.TrimSpace(line)) > 0 {
			t.Errorf("unexpected stderr: %q", line)
		}
	}

	for _, line := range strings.Split(stdout, "\n") {
		keywords := []string{"error", "fail", "warning", "conflict"}
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				t.Errorf("unexpected %q in stdout: %q", keyword, line)
			}
		}
	}

	if t.Failed() {
		t.Logf("minikube stdout:\n%s", stdout)
		t.Logf("minikube stderr:\n%s", stderr)
	}

	steps := []string{
		"Generating certificates and keys ...",
		"Booting up control plane ...",
		"Configuring RBAC rules ...",
	}
	for _, step := range steps {
		if !strings.Contains(stdout, step) {
			t.Errorf("missing kubeadm init sub-step %q", step)
		}
	}

	logTests := []struct {
		command          string
		args             []string
		runCount         int // number of times to run command
		expectedLogFiles int // number of logfiles expected after running command runCount times
	}{
		{
			command:          "logs",
			runCount:         15, // calling this 15 times should create 2 files with 1 greater than 1M
			expectedLogFiles: 2,
		},
		{
			command:          "status",
			runCount:         100,
			expectedLogFiles: 1,
		}, {
			command:          "pause",
			runCount:         5,
			expectedLogFiles: 1,
		}, {
			command:          "unpause",
			runCount:         1,
			expectedLogFiles: 1,
		}, {
			command:          "stop",
			runCount:         1,
			expectedLogFiles: 1,
		},
	}

	for _, test := range logTests {
		t.Run(test.command, func(t *testing.T) {
			args := []string{test.command, "-p", profile}
			args = append(args, test.args...)
			// run command runCount times
			for i := 0; i < test.runCount; i++ {
				rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
				if err != nil {
					t.Fatalf("%q failed: %v", rr.Command(), err)
				}
			}
			// get log files generated above
			logFiles, err := getLogFiles(test.command)
			if err != nil {
				t.Errorf("failed to find tmp log files: command %s : %v", test.command, err)
			}
			// cleanup generated logfiles
			defer cleanupLogFiles(t, logFiles)
			// if not the expected number of files, throw err
			if len(logFiles) != test.expectedLogFiles {
				t.Errorf("failed to find expected number of log files: cmd %s: expected: %d got %d", test.command, test.expectedLogFiles, len(logFiles))
			}
			// if more than 1 logfile is expected, only one file should be less than 1M
			if test.expectedLogFiles > 1 {
				foundSmall := false
				maxSize := 1024 * 1024 // 1M
				for _, logFile := range logFiles {
					isSmall := int(logFile.Size()) < maxSize
					if isSmall && !foundSmall {
						foundSmall = true
					} else if isSmall && foundSmall {
						t.Errorf("expected to find only one file less than 1M: cmd %s:", test.command)
					}
				}
			}
		})

	}
}

// getLogFiles returns logfiles corresponding to cmd
func getLogFiles(cmdName string) ([]os.FileInfo, error) {
	var logFiles []os.FileInfo
	err := filepath.Walk(os.TempDir(), func(path string, info os.FileInfo, err error) error {
		if strings.Contains(info.Name(), fmt.Sprintf("minikube_%s", cmdName)) {
			logFiles = append(logFiles, info)
		}
		return nil
	})
	return logFiles, err
}

// cleanupLogFiles removes logfiles generated during testing
func cleanupLogFiles(t *testing.T, logFiles []os.FileInfo) {
	for _, logFile := range logFiles {
		logFilePath := filepath.Join(os.TempDir(), logFile.Name())
		t.Logf("Cleaning up logfile %s ...", logFilePath)
		if err := os.Remove(logFilePath); err != nil {
			t.Errorf("failed to cleanup log file: %s : %v", logFilePath, err)
		}
	}
}
