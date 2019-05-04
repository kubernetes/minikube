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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	units "github.com/docker/go-units"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// ErrPrefix notes an error
const ErrPrefix = "! "

// OutPrefix notes output
const OutPrefix = "> "

const (
	downloadURL = "https://storage.googleapis.com/minikube/releases/%s/minikube-%s-amd64%s"
)

// RetriableError is an error that can be tried again
type RetriableError struct {
	Err error
}

func (r RetriableError) Error() string { return "Temporary Error: " + r.Err.Error() }

// CalculateDiskSizeInMB returns the number of MB in the human readable string
func CalculateDiskSizeInMB(humanReadableDiskSize string) int {
	diskSize, err := units.FromHumanSize(humanReadableDiskSize)
	if err != nil {
		glog.Errorf("Invalid disk size: %v", err)
	}
	return int(diskSize / units.MB)
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

// Retry retries a number of attempts
func Retry(attempts int, callback func() error) (err error) {
	return RetryAfter(attempts, callback, 0)
}

// RetryAfter retries a number of attempts, after a delay
func RetryAfter(attempts int, callback func() error, d time.Duration) (err error) {
	m := MultiError{}
	for i := 0; i < attempts; i++ {
		if i > 0 {
			glog.V(1).Infof("retry loop %d", i)
		}
		err = callback()
		if err == nil {
			return nil
		}
		m.Collect(err)
		if _, ok := err.(*RetriableError); !ok {
			glog.Infof("non-retriable error: %v", err)
			return m.ToError()
		}
		glog.V(2).Infof("error: %v - sleeping %s", err, d)
		time.Sleep(d)
	}
	return m.ToError()
}

// ParseSHAFromURL downloads and reads a SHA checksum from an URL
func ParseSHAFromURL(url string) (string, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "Error downloading checksum.")
	} else if r.StatusCode != http.StatusOK {
		return "", errors.Errorf("Error downloading checksum. Got HTTP Error: %s", r.Status)
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", errors.Wrap(err, "Error reading checksum.")
	}

	return strings.Trim(string(body), "\n"), nil
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

// MultiError holds multiple errors
type MultiError struct {
	Errors []error
}

// Collect adds the error
func (m *MultiError) Collect(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

// ToError converts all errors into one
func (m MultiError) ToError() error {
	if len(m.Errors) == 0 {
		return nil
	}

	errStrings := []string{}
	for _, err := range m.Errors {
		errStrings = append(errStrings, err.Error())
	}
	return errors.New(strings.Join(errStrings, "\n"))
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

// TeePrefix copies bytes from a reader to writer, logging each new line.
func TeePrefix(prefix string, r io.Reader, w io.Writer, logger func(format string, args ...interface{})) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)
	var line bytes.Buffer

	for scanner.Scan() {
		b := scanner.Bytes()
		if _, err := w.Write(b); err != nil {
			return err
		}

		if bytes.IndexAny(b, "\r\n") == 0 {
			if line.Len() > 0 {
				logger("%s%s", prefix, line.String())
				line.Reset()
			}
			continue
		}
		line.Write(b)
	}
	// Catch trailing output in case stream does not end with a newline
	if line.Len() > 0 {
		logger("%s%s", prefix, line.String())
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

// ContainsString checks if a given slice of strings contains the provided string.
// If a modifier func is provided, it is called with the slice item before the comparation.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
