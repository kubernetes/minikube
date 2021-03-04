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
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"
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
	new := filepath.Join(Profile(name), "client.crt")
	if _, err := os.Stat(new); err == nil {
		return new
	}

	// minikube v1.5.x
	legacy := filepath.Join(MiniPath(), "client.crt")
	if _, err := os.Stat(legacy); err == nil {
		klog.Infof("copying %s -> %s", legacy, new)
		if err := copy.Copy(legacy, new); err != nil {
			klog.Errorf("failed copy %s -> %s: %v", legacy, new, err)
			return legacy
		}
	}

	return new
}

// PID returns the path to the pid file used by profile for scheduled stop
func PID(profile string) string {
	return path.Join(Profile(profile), "pid")
}

// ClientKey returns client certificate path, used by kubeconfig
func ClientKey(name string) string {
	new := filepath.Join(Profile(name), "client.key")
	if _, err := os.Stat(new); err == nil {
		return new
	}

	// minikube v1.5.x
	legacy := filepath.Join(MiniPath(), "client.key")
	if _, err := os.Stat(legacy); err == nil {
		klog.Infof("copying %s -> %s", legacy, new)
		if err := copy.Copy(legacy, new); err != nil {
			klog.Errorf("failed copy %s -> %s: %v", legacy, new, err)
			return legacy
		}
	}

	return new
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
		s := image[:2] + strings.Replace(image[2:], ":", "_", -1)
		klog.Infof("windows sanitize: %s -> %s", image, s)
		return s
	}
	// ParseReference cannot have a : in the directory path
	return strings.Replace(image, ":", "_", -1)
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
			return "", errors.Wrap(err, "parsing docker archive dst ref: replace a Win drive letter to a volume name")
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
	path := vname + s[3:]

	return path, nil
}

func getWindowsVolumeNameCmd(d string) (string, error) {
	cmd := exec.Command("wmic", "volume", "where", "DriveLetter = '"+d+":'", "get", "DeviceID")

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	outs := strings.Split(strings.Replace(string(stdout), "\r", "", -1), "\n")

	var vname string
	for _, l := range outs {
		s := strings.TrimSpace(l)
		if strings.HasPrefix(s, `\\?\Volume{`) && strings.HasSuffix(s, `}\`) {
			vname = s
			break
		}
	}

	if vname == "" {
		return "", errors.New("failed to get a volume GUID")
	}

	return vname, nil
}

var getWindowsVolumeName = getWindowsVolumeNameCmd
