package version

import (
	"fmt"
	"strings"
)

var (
	// Version should be updated by hand at each release
	Version = "0.8.2"

	// GitCommit will be overwritten automatically by the build system
	GitCommit = "HEAD"
)

// FullVersion formats the version to be printed
func FullVersion() string {
	return fmt.Sprintf("%s, build %s", Version, GitCommit)
}

// RC checks if the Machine version is a release candidate or not
func RC() bool {
	return strings.Contains(Version, "rc")
}
