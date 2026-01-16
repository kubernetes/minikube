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

package localpath

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/otiai10/copy"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

// MinikubeHome is the name of the minikube home directory environment variable.
const MinikubeHome = "MINIKUBE_HOME"

// ConfigFile is the path of the config file
func ConfigFile() string {
	return MakeMiniPath("config", "config.json")
}

// MiniPath returns the path to the user's minikube dir
func MiniPath() string {
	minikubeHomeEnv := os.Getenv(MinikubeHome)
	if minikubeHomeEnv == "" {
		return filepath.Join(homedir.HomeDir(), ".minikube")
	}
	if filepath.Base(minikubeHomeEnv) == ".minikube" {
		return minikubeHomeEnv
	}
	return filepath.Join(minikubeHomeEnv, ".minikube")
}

// MakeMiniPath is a utility to calculate a relative path to our directory.
func MakeMiniPath(fileName ...string) string {
	args := []string{MiniPath()}
	args = append(args, fileName...)
	return filepath.Join(args...)
}

// Profile returns the path to a profile
func Profile(name string) string {
	return filepath.Join(MiniPath(), "profiles", name)
}

// EventLog returns the path to a CloudEvents log
// This log contains the transient state of minikube and the completed steps on start.
func EventLog(name string) string {
	return filepath.Join(Profile(name), "events.json")
}

// AuditLog returns the path to the audit log.
// This log contains a history of commands run, by who, when, and what arguments.
func AuditLog() string {
	return filepath.Join(MiniPath(), "logs", "audit.json")
}

// LastStartLog returns the path to the last start log.
func LastStartLog() string {
	return filepath.Join(MiniPath(), "logs", "lastStart.txt")
}

// ClientCert returns client certificate path, used by kubeconfig
func ClientCert(name string) string {
	newCert := filepath.Join(Profile(name), "client.crt")
	if _, err := os.Stat(newCert); err == nil {
		return newCert
	}

	// minikube v1.5.x
	legacy := filepath.Join(MiniPath(), "client.crt")
	if _, err := os.Stat(legacy); err == nil {
		klog.Infof("copying %s -> %s", legacy, newCert)
		if err := copy.Copy(legacy, newCert); err != nil {
			klog.Errorf("failed copy %s -> %s: %v", legacy, newCert, err)
			return legacy
		}
	}

	return newCert
}

// PID returns the path to the pid file used by profile for scheduled stop
func PID(profile string) string {
	return path.Join(Profile(profile), "pid")
}

// ClientKey returns client certificate path, used by kubeconfig
func ClientKey(name string) string {
	newKey := filepath.Join(Profile(name), "client.key")
	if _, err := os.Stat(newKey); err == nil {
		return newKey
	}

	// minikube v1.5.x
	legacy := filepath.Join(MiniPath(), "client.key")
	if _, err := os.Stat(legacy); err == nil {
		klog.Infof("copying %s -> %s", legacy, newKey)
		if err := copy.Copy(legacy, newKey); err != nil {
			klog.Errorf("failed copy %s -> %s: %v", legacy, newKey, err)
			return legacy
		}
	}

	return newKey
}

// CACert returns the minikube CA certificate shared between profiles
func CACert() string {
	return filepath.Join(MiniPath(), "ca.crt")
}

// MachinePath returns the minikube machine path of a machine
func MachinePath(machine string, miniHome ...string) string {
	miniPath := MiniPath()
	if len(miniHome) > 0 {
		miniPath = miniHome[0]
	}
	return filepath.Join(miniPath, "machines", machine)
}

// SanitizeCacheDir returns a path without special characters
func SanitizeCacheDir(image string) string {
	if runtime.GOOS == "windows" && hasWindowsDriveLetter(image) {
		// not sanitize Windows drive letter.
		s := image[:2] + strings.ReplaceAll(image[2:], ":", "_")
		klog.Infof("windows sanitize: %s -> %s", image, s)
		return s
	}
	// ParseReference cannot have a : in the directory path
	return strings.ReplaceAll(image, ":", "_")
}

func hasWindowsDriveLetter(s string) bool {
	if len(s) < 3 {
		return false
	}

	drive := s[:3]
	for _, b := range "CDEFGHIJKLMNOPQRSTUVWXYZABcdefghijklmnopqrstuvwxyzab" {
		if d := string(b) + ":"; drive == d+`\` || drive == d+`/` {
			return true
		}
	}

	return false
}

// DstPath returns an os specific
func DstPath(dst string) (string, error) {
	if runtime.GOOS == "windows" && hasWindowsDriveLetter(dst) {
		// ParseReference does not support a Windows drive letter.
		// Therefore, will replace the drive letter to a volume name.
		var err error
		if dst, err = replaceWinDriveLetterToVolumeName(dst); err != nil {
			return "", fmt.Errorf("parsing docker archive dst ref: replace a Win drive letter to a volume name: %w", err)
		}
	}
	return dst, nil
}

// Replace a drive letter to a volume name.
func replaceWinDriveLetterToVolumeName(s string) (string, error) {
	vname, err := getWindowsVolumeName(s[:1])
	if err != nil {
		return "", err
	}
	p := vname + s[3:]

	return p, nil
}

// findPowerShell locates the PowerShell executable
func findPowerShell() (string, error) {
	// First try to find powershell.exe in PATH
	if ps, err := exec.LookPath("powershell.exe"); err == nil {
		return ps, nil
	}

	// Fallback to using SystemRoot environment variable
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = "C:\\Windows"
	}

	// Try common PowerShell locations
	locations := []string{
		filepath.Join(systemRoot, "System32", "WindowsPowerShell", "v1.0", "powershell.exe"),
		filepath.Join(systemRoot, "SysWOW64", "WindowsPowerShell", "v1.0", "powershell.exe"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", fmt.Errorf("PowerShell not found in PATH or common locations")
}

// getWindowsVolumeNameCmd returns the Windows volume GUID for a given drive letter
func getWindowsVolumeNameCmd(d string) (string, error) {
	psPath, err := findPowerShell()
	if err != nil {
		return "", fmt.Errorf("failed to locate PowerShell: %w", err)
	}

	psCommand := `Get-CimInstance -ClassName Win32_Volume -Filter "DriveLetter = '` + d + `:'" | Select-Object -ExpandProperty DeviceID`

	cmd := exec.Command(psPath, "-NoProfile", "-NonInteractive", "-Command", psCommand)

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("PowerShell command failed: %w", err)
	}

	vname := strings.TrimSpace(out.String())
	if !strings.HasPrefix(vname, `\\?\Volume{`) || !strings.HasSuffix(vname, `}\`) {
		return "", fmt.Errorf("failed to get a volume GUID, got: %s", vname)
	}

	return vname, nil
}

var getWindowsVolumeName = getWindowsVolumeNameCmd
