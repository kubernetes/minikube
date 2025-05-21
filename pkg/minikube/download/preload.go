/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package download

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	// PreloadVersion is the current version of the preloaded tarball
	//
	// NOTE: You may need to bump this version up when upgrading auxiliary docker images
	PreloadVersion = "v18"
	// PreloadBucket is the name of the GCS bucket where preloaded volume tarballs exist
	PreloadBucket = "minikube-preloaded-volume-tarballs"
)

var (
	preloadStates = make(map[string]map[string]bool)
)

// TarballName returns name of the tarball
func TarballName(k8sVersion, containerRuntime string) string {
	if containerRuntime == "crio" {
		containerRuntime = "cri-o"
	}
	var storageDriver string
	if containerRuntime == "cri-o" {
		storageDriver = "overlay"
	} else {
		storageDriver = "overlay2"
	}
	return fmt.Sprintf("preloaded-images-k8s-%s-%s-%s-%s-%s.tar.lz4", PreloadVersion, k8sVersion, containerRuntime, storageDriver, runtime.GOARCH)
}

// returns the name of the checksum file
func checksumName(k8sVersion, containerRuntime string) string {
	return fmt.Sprintf("%s.checksum", TarballName(k8sVersion, containerRuntime))
}

// returns target dir for all cached items related to preloading
func targetDir() string {
	return localpath.MakeMiniPath("cache", "preloaded-tarball")
}

// PreloadChecksumPath returns the local path to the cached checksum file
func PreloadChecksumPath(k8sVersion, containerRuntime string) string {
	return filepath.Join(targetDir(), checksumName(k8sVersion, containerRuntime))
}

// TarballPath returns the local path to the cached preload tarball
func TarballPath(k8sVersion, containerRuntime string) string {
	return filepath.Join(targetDir(), TarballName(k8sVersion, containerRuntime))
}

// remoteTarballURL returns the URL for the remote tarball in GCS
func remoteTarballURL(k8sVersion, containerRuntime string) string {
	return fmt.Sprintf("https://%s/%s/%s/%s/%s", downloadHost, PreloadBucket, PreloadVersion, k8sVersion, TarballName(k8sVersion, containerRuntime))
}

func setPreloadState(k8sVersion, containerRuntime string, value bool) {
	cRuntimes, ok := preloadStates[k8sVersion]
	if !ok {
		cRuntimes = make(map[string]bool)
		preloadStates[k8sVersion] = cRuntimes
	}
	cRuntimes[containerRuntime] = value
}

var checkRemotePreloadExists = func(k8sVersion, containerRuntime string) bool {
	url := remoteTarballURL(k8sVersion, containerRuntime)
	resp, err := http.Head(url)
	if err != nil {
		klog.Warningf("%s fetch error: %v", url, err)
		return false
	}

	// note: err won't be set if it's a 404
	if resp.StatusCode != http.StatusOK {
		klog.Warningf("%s status code: %d", url, resp.StatusCode)
		return false
	}

	klog.Infof("Found remote preload: %s", url)
	return true
}

// PreloadExists returns true if there is a preloaded tarball that can be used
func PreloadExists(k8sVersion, containerRuntime, driverName string, forcePreload ...bool) bool {
	// TODO (#8166): Get rid of the need for this and viper at all
	force := false
	if len(forcePreload) > 0 {
		force = forcePreload[0]
	}

	// TODO: debug why this func is being called two times
	klog.Infof("Checking if preload exists for k8s version %s and runtime %s", k8sVersion, containerRuntime)
	// If `driverName` is BareMetal, there is no preload. Note: some uses of
	// `PreloadExists` assume that the driver is irrelevant unless BareMetal.
	if !driver.AllowsPreload(driverName) || !viper.GetBool("preload") && !force {
		return false
	}

	// If the preload existence is cached, just return that value.
	if preloadState, ok := preloadStates[k8sVersion][containerRuntime]; ok {
		return preloadState
	}

	// Omit remote check if tarball exists locally
	targetPath := TarballPath(k8sVersion, containerRuntime)
	if f, err := checkCache(targetPath); err == nil && f.Size() != 0 {
		klog.Infof("Found local preload: %s", targetPath)
		setPreloadState(k8sVersion, containerRuntime, true)
		return true
	}

	existence := checkRemotePreloadExists(k8sVersion, containerRuntime)
	setPreloadState(k8sVersion, containerRuntime, existence)
	return existence
}

var checkPreloadExists = PreloadExists

// Preload caches the preloaded images tarball on the host machine
func Preload(k8sVersion, containerRuntime, driverName string) error {
	targetPath := TarballPath(k8sVersion, containerRuntime)
	targetLock := targetPath + ".lock"

	releaser, err := lockDownload(targetLock)
	if releaser != nil {
		defer releaser.Release()
	}
	if err != nil {
		return err
	}

	if f, err := checkCache(targetPath); err == nil && f.Size() != 0 {
		klog.Infof("Found %s in cache, skipping download", targetPath)
		return nil
	}

	// Make sure we support this k8s version
	if !checkPreloadExists(k8sVersion, containerRuntime, driverName) {
		klog.Infof("Preloaded tarball for k8s version %s does not exist", k8sVersion)
		return nil
	}

	out.Step(style.FileDownload, "Downloading Kubernetes {{.version}} preload ...", out.V{"version": k8sVersion})
	url := remoteTarballURL(k8sVersion, containerRuntime)

	checksum, err := getChecksum(k8sVersion, containerRuntime)
	var realPath string
	if err != nil {
		klog.Warningf("No checksum for preloaded tarball for k8s version %s: %v", k8sVersion, err)
		realPath = targetPath
		tmp, err := os.CreateTemp(targetDir(), TarballName(k8sVersion, containerRuntime)+".*")
		if err != nil {
			return errors.Wrap(err, "tempfile")
		}
		targetPath = tmp.Name()
	} else if checksum != nil {
		// add URL parameter for go-getter to automatically verify the checksum
		url += fmt.Sprintf("?checksum=md5:%s", hex.EncodeToString(checksum))
	}

	if err := download(url, targetPath); err != nil {
		return errors.Wrapf(err, "download failed: %s", url)
	}

	if err := ensureChecksumValid(k8sVersion, containerRuntime, targetPath, checksum); err != nil {
		return err
	}

	if realPath != "" {
		klog.Infof("renaming tempfile to %s ...", TarballName(k8sVersion, containerRuntime))
		err := os.Rename(targetPath, realPath)
		if err != nil {
			return errors.Wrap(err, "rename")
		}
	}

	// If the download was successful, mark off that the preload exists in the cache.
	setPreloadState(k8sVersion, containerRuntime, true)
	return nil
}

func getStorageAttrs(name string) (*storage.ObjectAttrs, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, errors.Wrap(err, "getting storage client")
	}
	attrs, err := client.Bucket(PreloadBucket).Object(name).Attrs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting storage object")
	}
	return attrs, nil
}

// getChecksum returns the MD5 checksum of the preload tarball
var getChecksum = func(k8sVersion, containerRuntime string) ([]byte, error) {
	klog.Infof("getting checksum for %s ...", TarballName(k8sVersion, containerRuntime))
	filename := fmt.Sprintf("%s/%s/%s", PreloadVersion, k8sVersion, TarballName(k8sVersion, containerRuntime))
	attrs, err := getStorageAttrs(filename)
	if err != nil {
		return nil, err
	}
	return attrs.MD5, nil
}

// saveChecksumFile saves the checksum to a local file for later verification
func saveChecksumFile(k8sVersion, containerRuntime string, checksum []byte) error {
	klog.Infof("saving checksum for %s ...", TarballName(k8sVersion, containerRuntime))
	return os.WriteFile(PreloadChecksumPath(k8sVersion, containerRuntime), checksum, 0o644)
}

// verifyChecksum returns true if the checksum of the local binary matches
// the checksum of the remote binary
func verifyChecksum(k8sVersion, containerRuntime, binaryPath string) error {
	klog.Infof("verifying checksum of %s ...", binaryPath)
	// get md5 checksum of tarball path
	contents, err := os.ReadFile(binaryPath)
	if err != nil {
		return errors.Wrap(err, "reading tarball")
	}
	checksum := md5.Sum(contents)

	remoteChecksum, err := os.ReadFile(PreloadChecksumPath(k8sVersion, containerRuntime))
	if err != nil {
		return errors.Wrap(err, "reading checksum file")
	}

	// create a slice of checksum, which is [16]byte
	if string(remoteChecksum) != string(checksum[:]) {
		return fmt.Errorf("checksum of %s does not match remote checksum (%s != %s)", binaryPath, string(remoteChecksum), string(checksum[:]))
	}
	return nil
}

// ensureChecksumValid saves and verifies local binary checksum matches remote binary checksum
var ensureChecksumValid = func(k8sVersion, containerRuntime, targetPath string, checksum []byte) error {
	if err := saveChecksumFile(k8sVersion, containerRuntime, checksum); err != nil {
		return errors.Wrap(err, "saving checksum file")
	}

	if err := verifyChecksum(k8sVersion, containerRuntime, targetPath); err != nil {
		return errors.Wrap(err, "verify")
	}

	return nil
}

// CleanUpOlderPreloads deletes preload files belonging to older minikube versions
// checks the current preload version and then if the saved tar file is belongs to older minikube it will delete it
// in case of failure only logs to the user
func CleanUpOlderPreloads() {
	files, err := os.ReadDir(targetDir())
	if err != nil {
		klog.Warningf("Failed to list preload files: %v", err)
	}

	for _, file := range files {
		splited := strings.Split(file.Name(), "-")
		if len(splited) < 4 {
			continue
		}
		ver := splited[3]
		if ver != PreloadVersion {
			fn := path.Join(targetDir(), file.Name())
			klog.Infof("deleting older generation preload %s", fn)
			err := os.Remove(fn)
			if err != nil {
				klog.Warningf("Failed to clean up older preload files, consider running `minikube delete --all --purge`")
			}
		}
	}
}
