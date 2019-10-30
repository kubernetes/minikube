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

package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	units "github.com/docker/go-units"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
)

const (
	// ErrPrefix notes an error
	ErrPrefix = "! "

	// OutPrefix notes output
	OutPrefix = "> "

	downloadURL = "https://storage.googleapis.com/minikube/releases/%s/minikube-%s-amd64%s"
)

// CalculateSizeInMB returns the number of MB in the human readable string
func CalculateSizeInMB(humanReadableSize string) int {
	_, err := strconv.ParseInt(humanReadableSize, 10, 64)
	if err == nil {
		humanReadableSize += "mb"
	}
	size, err := units.FromHumanSize(humanReadableSize)
	if err != nil {
		exit.WithCodeT(exit.Config, "Invalid size passed in argument: {{.error}}", out.V{"error": err})
	}

	return int(size / units.MB)
}

// Until endlessly loops the provided function until a message is received on the done channel.
// The function will wait the duration provided in sleep between function calls. Errors will be sent on provider Writer.
func Until(fn func() error, w io.Writer, name string, sleep time.Duration, done <-chan struct{}) {
	var exitErr error
	for {
		select {
		case <-done:
			return
		default:
			exitErr = fn()
			if exitErr == nil {
				fmt.Fprintf(w, Pad("%s: Exited with no errors.\n"), name)
			} else {
				fmt.Fprintf(w, Pad("%s: Exit with error: %v"), name, exitErr)
			}

			// wait provided duration before trying again
			time.Sleep(sleep)
		}
	}
}

// Pad pads the string with newlines
func Pad(str string) string {
	return fmt.Sprintf("\n%s\n", str)
}

// CanReadFile returns true if the file represented
// by path exists and is readable, otherwise false.
func CanReadFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}

	defer f.Close()

	return true
}

// GetBinaryDownloadURL returns a suitable URL for the platform
func GetBinaryDownloadURL(version, platform string) string {
	switch platform {
	case "windows":
		return fmt.Sprintf(downloadURL, version, platform, ".exe")
	default:
		return fmt.Sprintf(downloadURL, version, platform, "")
	}
}

// IsDirectory checks if path is a directory
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, errors.Wrapf(err, "Error calling os.Stat on file %s", path)
	}
	return fileInfo.IsDir(), nil
}

// ChownR does a recursive os.Chown
func ChownR(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
		}
		return err
	})
}

// MaybeChownDirRecursiveToMinikubeUser changes ownership of a dir, if requested
func MaybeChownDirRecursiveToMinikubeUser(dir string) error {
	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") != "" && os.Getenv("SUDO_USER") != "" {
		username := os.Getenv("SUDO_USER")
		usr, err := user.Lookup(username)
		if err != nil {
			return errors.Wrap(err, "Error looking up user")
		}
		uid, err := strconv.Atoi(usr.Uid)
		if err != nil {
			return errors.Wrapf(err, "Error parsing uid for user: %s", username)
		}
		gid, err := strconv.Atoi(usr.Gid)
		if err != nil {
			return errors.Wrapf(err, "Error parsing gid for user: %s", username)
		}
		if err := ChownR(dir, uid, gid); err != nil {
			return errors.Wrapf(err, "Error changing ownership for: %s", dir)
		}
	}
	return nil
}

// ReplaceChars returns a copy of the src slice with each string modified by the replacer
func ReplaceChars(src []string, replacer *strings.Replacer) []string {
	ret := make([]string, len(src))
	for i, s := range src {
		ret[i] = replacer.Replace(s)
	}
	return ret
}

// ConcatStrings concatenates each string in the src slice with prefix and postfix and returns a new slice
func ConcatStrings(src []string, prefix string, postfix string) []string {
	var buf bytes.Buffer
	ret := make([]string, len(src))
	for i, s := range src {
		buf.WriteString(prefix)
		buf.WriteString(s)
		buf.WriteString(postfix)
		ret[i] = buf.String()
		buf.Reset()
	}
	return ret
}
