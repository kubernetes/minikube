package integration

import (
	"fmt"
	"os"
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
	// Warning of issues with specific Kubernetes versions
	`Kubernetes .* has a known `,
	`For more information, see`,
}

// stderrAllowRe combines rootCauses into a single regex
var stderrAllowRe = regexp.MustCompile(strings.Join(stderrAllow, "|"))

// checkSpam checks standard output and standard error for unexpected spam
func checkSpam(t *testing.T, stdout, stderr string) {
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
}

func getLogFiles(logDir string, command string) ([]string, error) {
	return filepath.Glob(filepath.Join(logDir, fmt.Sprintf("minikube_%s*", command)))
}

func checkLogFileCount(command string, logDir string, expectedNumberOfLogFiles int) error {
	// get log files generated above
	logFiles, err := getLogFiles(logDir, command)
	if err != nil {
		return fmt.Errorf("failed to get new log files for command %s : %v", command, err)
	}

	if len(logFiles) != expectedNumberOfLogFiles {
		return fmt.Errorf("Running cmd %q resulted in %d log file(s); expected: %d", command, len(logFiles), expectedNumberOfLogFiles)
	}

	return nil
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
