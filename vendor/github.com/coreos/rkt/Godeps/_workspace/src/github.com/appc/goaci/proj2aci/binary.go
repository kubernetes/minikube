package proj2aci

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// GetBinaryName checks if useBinary is in binDir and returns it. If
// useBinary is empty it returns a binary name if there is only one
// such in binDir. Otherwise it returns an error.
func GetBinaryName(binDir, useBinary string) (string, error) {
	fi, err := ioutil.ReadDir(binDir)
	if err != nil {
		return "", err
	}

	switch {
	case len(fi) < 1:
		return "", fmt.Errorf("No binaries found in %q", binDir)
	case len(fi) == 1:
		name := fi[0].Name()
		if useBinary != "" && name != useBinary {
			return "", fmt.Errorf("No such binary found in %q: %q. There is only %q", binDir, useBinary, name)
		}
		Debug("found binary: ", name)
		return name, nil
	case len(fi) > 1:
		names := []string{}
		for _, v := range fi {
			names = append(names, v.Name())
		}
		if useBinary == "" {
			return "", fmt.Errorf("Found multiple binaries in %q, but no specific binary is preferred. Please specify which binary to pick up. Following binaries are available: %q", binDir, strings.Join(names, `", "`))
		}
		for _, v := range names {
			if v == useBinary {
				return v, nil
			}
		}
		return "", fmt.Errorf("No such binary found in %q: %q. There are following binaries available: %q", binDir, useBinary, strings.Join(names, `", "`))
	}
	panic("Reaching this point shouldn't be possible.")
}
