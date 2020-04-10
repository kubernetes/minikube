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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/golang/glog"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
)

const (
	// PreloadVersion is the current version of the preloaded tarball
	//
	// NOTE: You may need to bump this version up when upgrading auxiliary docker images
	PreloadVersion = "v2"
	// PreloadBucket is the name of the GCS bucket where preloaded volume tarballs exist
	PreloadBucket = "minikube-preloaded-volume-tarballs"
)

// TarballName returns name of the tarball
func TarballName(k8sVersion, containerRuntime string) string {
	return fmt.Sprintf("preloaded-images-k8s-%s-%s-%s-overlay2-%s.tar.lz4", PreloadVersion, k8sVersion, containerRuntime, runtime.GOARCH)
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
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", PreloadBucket, TarballName(k8sVersion, containerRuntime))
}

// PreloadExists returns true if there is a preloaded tarball that can be used
func PreloadExists(k8sVersion, containerRuntime string) bool {
	glog.Infof("Checking if preload exists for k8s version %s and runtime %s", k8sVersion, containerRuntime)
	if !viper.GetBool("preload") {
		return false
	}

	// See https://github.com/kubernetes/minikube/issues/6933
	// and https://github.com/kubernetes/minikube/issues/6934
	// to track status of adding containerd & crio
	if containerRuntime != "docker" {
		glog.Info("Container runtime isn't docker, skipping preload")
		return false
	}

	// Omit remote check if tarball exists locally
	targetPath := TarballPath(k8sVersion, containerRuntime)
	if _, err := os.Stat(targetPath); err == nil {
		glog.Infof("Found local preload: %s", targetPath)
		return true
	}

	url := remoteTarballURL(k8sVersion, containerRuntime)
	resp, err := http.Head(url)
	if err != nil {
		glog.Warningf("%s fetch error: %v", url, err)
		return false
	}

	// note: err won't be set if it's a 404
	if resp.StatusCode != 200 {
		glog.Warningf("%s status code: %d", url, resp.StatusCode)
		return false
	}

	glog.Infof("Found remote preload: %s", url)
	return true
}

// Preload caches the preloaded images tarball on the host machine
func Preload(k8sVersion, containerRuntime string) error {
	targetPath := TarballPath(k8sVersion, containerRuntime)

	if _, err := os.Stat(targetPath); err == nil {
		glog.Infof("Found %s in cache, skipping download", targetPath)
		return nil
	}

	// Make sure we support this k8s version
	if !PreloadExists(k8sVersion, containerRuntime) {
		glog.Infof("Preloaded tarball for k8s version %s does not exist", k8sVersion)
		return nil
	}

	out.T(out.FileDownload, "Downloading Kubernetes {{.version}} preload ...", out.V{"version": k8sVersion})
	url := remoteTarballURL(k8sVersion, containerRuntime)

	tmpDst := targetPath + ".download"
	client := &getter.Client{
		Src:     url,
		Dst:     tmpDst,
		Mode:    getter.ClientModeFile,
		Options: []getter.ClientOption{getter.WithProgress(DefaultProgressBar)},
	}

	glog.Infof("Downloading: %+v", client)
	if err := client.Get(); err != nil {
		return errors.Wrapf(err, "download failed: %s", url)
	}

	if err := saveChecksumFile(k8sVersion, containerRuntime); err != nil {
		return errors.Wrap(err, "saving checksum file")
	}

	if err := verifyChecksum(k8sVersion, containerRuntime, tmpDst); err != nil {
		return errors.Wrap(err, "verify")
	}
	return os.Rename(tmpDst, targetPath)
}

func saveChecksumFile(k8sVersion, containerRuntime string) error {
	glog.Infof("saving checksum for %s ...", TarballName(k8sVersion, containerRuntime))
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return errors.Wrap(err, "getting storage client")
	}
	attrs, err := client.Bucket(PreloadBucket).Object(TarballName(k8sVersion, containerRuntime)).Attrs(ctx)
	if err != nil {
		return errors.Wrap(err, "getting storage object")
	}
	checksum := attrs.MD5
	return ioutil.WriteFile(PreloadChecksumPath(k8sVersion, containerRuntime), checksum, 0644)
}

// verifyChecksum returns true if the checksum of the local binary matches
// the checksum of the remote binary
func verifyChecksum(k8sVersion, containerRuntime, path string) error {
	glog.Infof("verifying checksumm of %s ...", path)
	// get md5 checksum of tarball path
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "reading tarball")
	}
	checksum := md5.Sum(contents)

	remoteChecksum, err := ioutil.ReadFile(PreloadChecksumPath(k8sVersion, containerRuntime))
	if err != nil {
		return errors.Wrap(err, "reading checksum file")
	}

	// create a slice of checksum, which is [16]byte
	if string(remoteChecksum) != string(checksum[:]) {
		return fmt.Errorf("checksum of %s does not match remote checksum (%s != %s)", path, string(remoteChecksum), string(checksum[:]))
	}
	return nil
}
