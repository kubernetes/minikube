package download

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	sudominikubeVersion = "v0.0.1-test"
)

func SudoMinikube(dir, osName, archName string) error {
	targetFileName := filepath.Join(dir, "sudominikube")
	downloadURL := fmt.Sprintf("https://github.com/ComradeProgrammer/minikube/releases/download/%s/sudominikube-%s-%s", sudominikubeVersion, osName, archName)
	if err := download(downloadURL, targetFileName); err != nil {
		return err
	}
	if data, err := exec.Command("sudo", "chmod", "777", targetFileName).CombinedOutput(); err != nil {
		return errors.WithMessagef(err, "failed to give permission: %s", string(data))
	}

	return nil
}
