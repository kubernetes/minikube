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
	// "! 'docker' driver reported a issue that could affect the performance."
	`docker.*issue.*performance`,
	// "* Suggestion: enable overlayfs kernel module on your Linux"
	`Suggestion.*overlayfs`,
	// "! docker is currently using the btrfs storage driver, consider switching to overlay2 for better performance"
	`docker.*btrfs storage driver`,
	// jenkins VMs (debian 9) cgoups don't allow setting memory
	`Your cgroup does not allow setting memory.`,
	// progress bar output
	`    > .*`,
}

// stderrAllowRe combines rootCauses into a single regex
var stderrAllowRe = regexp.MustCompile(strings.Join(stderrAllow, "|"))

// TestErrorSpam asserts that there are no unexpected errors displayed in minikube command outputs.
func TestErrorSpam(t *testing.T) {
	if NoneDriver() {
		t.Skip("none driver always shows a warning")
	}

	profile := UniqueProfileName("nospam")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(25))
	defer CleanupWithLogs(t, profile, cancel)

	logDir := filepath.Join(os.TempDir(), profile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Unable to make logDir %s: %v", logDir, err)
	}
	defer os.RemoveAll(logDir)

	// This should likely use multi-node once it's ready
	// use `--log_dir` flag to run isolated and avoid race condition - ie, failing to clean up (locked) log files created by other concurently-run tests, or counting them in results
	args := append([]string{"start", "-p", profile, "-n=1", "--memory=2250", "--wait=false", fmt.Sprintf("--log_dir=%s", logDir)}, StartArgs()...)

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

	logsToBeRemoved, err := filepath.Glob(filepath.Join(logDir, "minikube_*"))
	if err != nil {
		t.Error("failed to clean old log files", err)
	}
	cleanupLogFiles(t, logsToBeRemoved)

	logTests := []struct {
		command          string
		args             []string
		runCount         int // number of times to run command
		expectedLogFiles int // number of logfiles expected after running command runCount times
	}{
		{
			command:          "start",
			args:             []string{"--dry-run"},
			runCount:         175, // calling this 175 times should create 2 files with 1 greater than 1M
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

			// flags can be before subcommand
			args := []string{"-p", profile, "--log_dir", logDir, test.command}
			args = append(args, test.args...)

			// before starting the test, ensure no other logs from the current command are written
			logFiles, err := filepath.Glob(filepath.Join(logDir, fmt.Sprintf("minikube_%s*", test.command)))
			if err != nil {
				t.Errorf("failed to get old log files for command %s : %v", test.command, err)
			}
			cleanupLogFiles(t, logFiles)

			singleRun, err := Run(t, exec.CommandContext(ctx, Target(), args...))
			if err != nil {
				t.Errorf("%q failed: %v", singleRun.Command(), err)
			}

			// get log files generated above for a single run
			logFiles, err = filepath.Glob(filepath.Join(logDir, fmt.Sprintf("minikube_%s*", test.command)))
			if err != nil {
				t.Errorf("failed to get new log files for command %s : %v", test.command, err)
			}
			cleanupLogFiles(t, logFiles)

			// run command runCount times
			for i := 0; i < test.runCount; i++ {
				rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
				if err != nil {
					t.Errorf("%q failed: %v", rr.Command(), err)
				}
			}

			// get log files generated above
			logFiles, err = filepath.Glob(filepath.Join(logDir, fmt.Sprintf("minikube_%s*", test.command)))
			if err != nil {
				t.Errorf("failed to get new log files for command %s : %v", test.command, err)
			}

			// if not the expected number of files, throw err
			if len(logFiles) != test.expectedLogFiles {
				t.Errorf("failed to find expected number of log files: cmd %s: expected: %d got %d", test.command, test.expectedLogFiles, len(logFiles))
			}

			// if more than 1 logfile is expected, only one file should be less than 1M
			if test.expectedLogFiles > 1 {
				foundSmall := false
				var maxSize int64 = 1024 * 1024 // 1M
				for _, logFile := range logFiles {
					finfo, err := os.Stat(logFile)
					if err != nil {
						t.Logf("logfile %q for command %q not found:", logFile, test.command)
						continue
					}
					isSmall := finfo.Size() < maxSize
					if isSmall && !foundSmall {
						foundSmall = true
					} else if isSmall && foundSmall {
						t.Errorf("expected to find only one file less than 1MB: cmd %s:", test.command)
					}
				}
			}
		})

	}
}

// cleanupLogFiles removes logfiles generated during testing
func cleanupLogFiles(t *testing.T, logFiles []string) {
	t.Logf("Cleaning up %d logfile(s) ...", len(logFiles))
	for _, logFile := range logFiles {
		if err := os.Remove(logFile); err != nil {
			t.Logf("failed to cleanup log file: %s : %v", logFile, err)
		}
	}
}
