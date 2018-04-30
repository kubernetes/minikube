/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package update

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blang/semver"
	"github.com/inconshreveable/go-update"
	"gopkg.in/cheggaaa/pb.v1"
	"io/ioutil"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/version"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LatestVersion gives the latest version of minikube
func LatestVersion() (semver.Version, error) {
	r, err := notify.GetAllVersionsFromURL(constants.GithubMinikubeReleasesURL)
	if err != nil {
		return semver.Version{}, errors.New("error fetching latest version from internet")
	}

	if len(r) < 1 {
		return semver.Version{}, errors.New("got empty version list from server")
	}

	return semver.Make(strings.TrimPrefix(r[0].Name, version.VersionPrefix))
}

// IsNewerVersion compares the local and latest versions and returns a boolean
func IsNewerVersion(currentVersion, latestVersion semver.Version) bool {
	if currentVersion.Compare(latestVersion) < 0 {
		return true
	} else {
		return false
	}
}

// Update updates the binary to the latest version
func Update(latestVersion semver.Version) error {
	// temporary directory to store downloaded binary contents
	tmpDir, err := ioutil.TempDir("", "download")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return errors.New(fmt.Sprintf("could not create a temporary directory: %s", err))
	}

	version := fmt.Sprintf("%s%s", version.VersionPrefix, latestVersion)
	downloadURL := util.GetBinaryDownloadURL(version, runtime.GOOS)
	fmt.Printf("Downloading latest minikube version %s\n", version)
	binaryPath, err := downloadBinary(downloadURL, tmpDir)
	if err != nil {
		return err
	}

	// Replace the existing binary with the downloaded binary
	if err := updateBinary(binaryPath); err != nil {
		return err
	}

	return nil
}

// downloadBinary downloads the latest minikube binary from GitHub into a temporary location.
// It returns a string containing path to the downloaded binary.
// It returns an error on failure.
func downloadBinary(url, tmpDir string) (string, error) {
	httpResp, err := http.Get(url)
	if err != nil || httpResp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("cannot download binary from '%s': %s", url, err))
	}
	defer func() { _ = httpResp.Body.Close() }()

	binaryContent := httpResp.Body
	if httpResp.ContentLength > 0 {
		bar := pb.New64(httpResp.ContentLength).SetUnits(pb.U_BYTES)
		bar.Start()
		binaryContent = bar.NewProxyReader(binaryContent)
		defer func() {
			<-time.After(bar.RefreshRate)
			fmt.Println()
		}()
	}

	binaryBytes, err := ioutil.ReadAll(binaryContent)
	if err != nil {
		return "", errors.New(fmt.Sprintf("unable to read downloaded binary: %s", err))
	}

	// Save to binary
	urlSplit := strings.Split(url, "/")
	downloadedBinaryPath := filepath.Join(tmpDir, urlSplit[len(urlSplit)-1])
	if err := ioutil.WriteFile(downloadedBinaryPath, binaryBytes, 0644); err != nil {
		return "", err
	}

	// ignore checksum verification in Windows platform, see issue #2536
	if runtime.GOOS != "windows" {
		checksumURL := fmt.Sprintf(url + ".sha256")
		if err := verifyBinaryChecksum(checksumURL, binaryBytes); err != nil {
			return "", err
		}
	}

	return downloadedBinaryPath, nil
}

// checksumFor evaluates and returns the checksum for the payload passed to it.
// It returns an error if given hash function is not linked into the binary.
// Check "crypto" package for more info on hash function.
func checksumFor(h crypto.Hash, payload []byte) ([]byte, error) {
	if !h.Available() {
		return nil, errors.New("requested hash function not available")
	}

	hash := h.New()
	hash.Write(payload)

	return hash.Sum([]byte{}), nil
}

// verifyBinaryChecksum verify the checksum present in checksumURL with binary content.
// It returns an error if checksum fails to match.
func verifyBinaryChecksum(checksumURL string, binaryBytes []byte) error {
	// Find checksum of binary
	binaryChecksum, err := checksumFor(crypto.SHA256, binaryBytes)
	if err != nil {
		return err
	}

	checksumResp, err := http.Get(checksumURL)
	if err != nil {
		return err
	}
	defer func() { _ = checksumResp.Body.Close() }()

	checksumContent := checksumResp.Body
	// Verify checksum
	checksumBytes, err := ioutil.ReadAll(checksumContent)
	if err != nil {
		return err
	}
	if checksumResp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("received %d during checksum download", checksumResp.StatusCode))
	}

	downloadedChecksum, err := hex.DecodeString(strings.TrimSpace(string(checksumBytes)))
	if err != nil {
		return err
	}

	// Compare checksums of downloaded checksum and binary file
	if !bytes.Equal(binaryChecksum, downloadedChecksum) {
		return fmt.Errorf("binary checksum mismatch, expected: %x, got: %x", binaryChecksum, checksumContent)
	}

	return nil
}

// updateBinary takes the path to the latest downloaded binary and replaces the existing binary.
// It returns an error if update fails.
func updateBinary(binaryPath string) error {
	binaryFile, err := os.Open(binaryPath)
	if err != nil {
		return err
	}

	err = update.Apply(binaryFile, update.Options{
		Hash: crypto.SHA256,
	})
	if err != nil {
		rollbackErr := update.RollbackError(err)
		if rollbackErr != nil {
			return errors.New(fmt.Sprintf("failed to rollback update: %s", rollbackErr))
		}
		return err
	}

	return nil
}
