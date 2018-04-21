package parallels

import "fmt"

// GitCommit that was compiled. This will be filled in by the compiler.
var GitCommit string

// Version number that is being run at the moment.
const Version = "1.3.0"

// FullVersion formats the version to be printed.
func FullVersion() string {
	return fmt.Sprintf("%s, build %s", Version, GitCommit)
}
