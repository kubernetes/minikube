package oci

import (
	"os/exec"
	"strings"
	
	"k8s.io/minikube/pkg/minikube/image"
)

// ToDriverCache
// calls OCIBIN's load command at specified path:
// loads the archived container image at provided PATH.
func ArchiveToDriverCache(ociBin, path string) (string, error) {
	_, err := runCmd(exec.Command(ociBin, "load", "-i", path))
	return "", err
}

// IsInCache
// searches in OCIBIN's cache for the IMG; returns true if found. no error handling
func IsImageInCache(ociBin, img string) (bool) {
	res, err := runCmd(exec.Command(ociBin, "images", "--format", "{{.Repository}}:{{.Tag}}@{{.Digest}}"))
	if err != nil {
		// only the docker binary seems to have this issue..
		// the docker.io/ substring is cut from the output and formatting doesn't help
		if ociBin == Docker {
			img = image.TrimDockerIO(img)
		}
		
		if strings.Contains(res.Stdout.String(), img){
			return true
		}
	}

	return false
}
