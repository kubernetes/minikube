package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Azure/azure-sdk-for-go/tools/apidiff/ioext"

	"k8s.io/minikube/pkg/version"
)

func TestHyperkitDriverSkipUpgrade(t *testing.T) {
	MaybeParallel(t)
	tests := []struct {
		name            string
		path            string
		expectedVersion string
	}{
		{
			name:            "upgrade-v1.11.0-to-current",
			path:            filepath.Join(*testdataDir, "hyperkit-driver-version-1.11.0"),
			expectedVersion: "v1.11.0",
		},
		{
			name:            "upgrade-v1.2.0-to-current",
			path:            filepath.Join(*testdataDir, "hyperkit-driver-older-version"),
			expectedVersion: version.GetVersion(),
		},
	}

	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		t.Fatalf("No sudo in path: %v", err)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mkDir, drvPath, err := tempMinikubeDir(tc.name, tc.path)
			if err != nil {
				t.Fatalf("Failed to prepare tempdir. test: %s, got: %v", tc.name, err)
			}
			defer func() {
				if err := os.RemoveAll(mkDir); err != nil {
					t.Errorf("Failed to remove mkDir %q: %v", mkDir, err)
				}
			}()

			// start "minikube start --download-only --interactive=false --driver=hyperkit --preload=false"
			cmd := exec.Command(Target(), "start", "--download-only", "--interactive=false", "--driver=hyperkit", "--preload=false")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			// set PATH=<tmp_minikube>/bin:<path_to_sudo>
			//     MINIKUBE_PATH=<tmp_minikube>
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PATH=%v%c%v", filepath.Dir(drvPath), filepath.ListSeparator, filepath.Dir(sudoPath)),
				"MINIKUBE_HOME="+mkDir)
			if err = cmd.Run(); err != nil {
				t.Fatalf("failed to run minikube. got: %v", err)
			}

			upgradedVersion, err := driverVersion(drvPath)
			if err != nil {
				t.Fatalf("failed to check driver version. got: %v", err)
			}

			if upgradedVersion != tc.expectedVersion {
				t.Fatalf("invalid driver version. expected: %v, got: %v", tc.expectedVersion, upgradedVersion)
			}
		})
	}
}

func driverVersion(path string) (string, error) {
	output, err := exec.Command(path, "version").Output()
	if err != nil {
		return "", err
	}

	var resultVersion string
	_, err = fmt.Sscanf(string(output), "version: %s\n", &resultVersion)
	if err != nil {
		return "", err
	}
	return resultVersion, nil
}

func tempMinikubeDir(name, driver string) (string, string, error) {
	temp, err := ioutil.TempDir("", name)
	if err != nil {
		return "", "", fmt.Errorf("failed to create tempdir: %v", err)
	}

	mkDir := filepath.Join(temp, ".minikube")
	mkBinDir := filepath.Join(mkDir, "bin")
	err = os.MkdirAll(mkBinDir, 0777)
	if err != nil {
		return "", "", fmt.Errorf("failed to prepare tempdir: %v", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get working directory: %v", err)
	}

	testDataDriverPath := filepath.Join(pwd, driver, "docker-machine-driver-hyperkit")
	if _, err = os.Stat(testDataDriverPath); err != nil {
		return "", "", fmt.Errorf("expected driver to exist: %v", err)
	}

	// copy driver to temp bin
	testDriverPath := filepath.Join(mkBinDir, "docker-machine-driver-hyperkit")
	if err = ioext.CopyFile(testDataDriverPath, testDriverPath, false); err != nil {
		return "", "", fmt.Errorf("failed to setup current hyperkit driver: %v", err)
	}
	// change permission to allow driver to be executable
	if err = os.Chmod(testDriverPath, 0777); err != nil {
		return "", "", fmt.Errorf("failed to set driver permission: %v", err)
	}
	return temp, testDriverPath, nil
}
