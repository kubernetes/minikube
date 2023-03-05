package sudominikube

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/version"
)

const sudoersd = `
%minikube ALL=(root:root) NOPASSWD:NOSETENV: /opt/minikube/bin/sudominikube
`
const SudominikubeLocation = "/opt/minikube/bin"

// InstallSudoMinikube installs the sudominikube
// param location:  path to the folder where sudominikube is downloaded
func InstallSudoMinikube(location string) error {
	data, err := exec.Command("sudo", "groupadd", "-f", "minikube").CombinedOutput()
	if err != nil {
		return errors.WithMessagef(err, "failed to run groupadd: %s", string(data))
	}

	user := os.Getenv("USER")
	data, err = exec.Command("sudo", "usermod", "-a", "-G", "minikube", user).CombinedOutput()
	if err != nil {
		return errors.WithMessagef(err, "failed to run usermod: %s", string(data))
	}
	// verify the download
	if err := VerifySudoMinikube(filepath.Join(location, "sudominikube")); err != nil {
		return errors.WithMessage(err, "invalid sudominikube ")
	}
	// move it to  /opt/minikube/bin with sudo permission
	data, err = exec.Command("sudo", "mkdir", "-p", SudominikubeLocation).CombinedOutput()
	if err != nil {
		return errors.WithMessagef(err, "failed to create folder %s bin: %s", SudominikubeLocation, string(data))
	}
	data, err = exec.Command("sudo", "cp", filepath.Join(location, "sudominikube"), SudominikubeLocation).CombinedOutput()
	if err != nil {
		return errors.WithMessagef(err, "failed to copy sudominikube to %s: %s", SudominikubeLocation, string(data))
	}

	// generate sudoer.d/sudominikube
	f, err := os.OpenFile(filepath.Join(location, "sudo_minikube"), os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return errors.WithMessage(err, "failed to create sudo_minikube file in download folder")
	}
	_, err = f.WriteString(sudoersd)
	if err != nil {
		return errors.WithMessage(err, "failed to write sudo_minikube file")
	}
	f.Close()
	data, err = exec.Command("sudo", "mv", filepath.Join(location, "sudo_minikube"), "/etc/sudoers.d").CombinedOutput()
	if err != nil {
		return errors.WithMessagef(err, "failed to move sudo_minikube to /etc/sudoers.d: %s", string(data))
	}
	data, err = exec.Command("sudo", "chown", "root:root", "/etc/sudoers.d/sudo_minikube").CombinedOutput()
	if err != nil {
		return errors.WithMessagef(err, "failed to change owner /etc/sudoers.d/sudo_minikube: %s", string(data))
	}
	return nil
}

// VerifySudoMinikube checks whether sudominikube binary is compatible with minikube
// parameter path: location of sudominikube binary
func VerifySudoMinikube(path string) error {
	// sudo is required here because the sudominikube binary may be
	// downloaded to somewhere where the user has no permission
	cmd := exec.Command("sudo", path, "version", "-o", "json")
	data, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "failed to execute sudominikube version")
	}
	var ver map[string]string
	if err := json.Unmarshal(data, &ver); err != nil {
		return err
	}

	minikubeVersion := version.GetVersion()
	// gitCommitID := version.GetGitCommitID()
	if ver["minikubeVersion"] != minikubeVersion {
		return fmt.Errorf("inconsistent minikubeVersion, minikube: %s, sudominikube: %s", minikubeVersion, ver["minikubeVersion"])
	}
	// if ver["commit"] != gitCommitID {
	// 	return fmt.Errorf("inconsistent commitID, minikube: %s, sudominikbube: %s", gitCommitID, ver["commit"])
	// }

	return nil
}

// return whether current program is sudominikube or minikube
func IsSudoMinikube() bool {
	return strings.HasSuffix(os.Args[0], "sudominikube")
}
