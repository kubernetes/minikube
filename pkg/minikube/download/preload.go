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

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	// PreloadVersion is the current version of the preloaded tarball
	//
	// NOTE: You may need to bump this version up when upgrading auxiliary docker images
	PreloadVersion = "v7"
	// PreloadBucket is the name of the GCS bucket where preloaded volume tarballs exist
	PreloadBucket = "minikube-preloaded-volume-tarballs"
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
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", PreloadBucket, TarballName(k8sVersion, containerRuntime))
}

// PreloadExists returns true if there is a preloaded tarball that can be used
func PreloadExists(k8sVersion, containerRuntime string, forcePreload ...bool) bool {
	// TODO (#8166): Get rid of the need for this and viper at all
	force := false
	if len(forcePreload) > 0 {
		force = forcePreload[0]
	}

	// TODO: debug why this func is being called two times
	klog.Infof("Checking if preload exists for k8s version %s and runtime %s", k8sVersion, containerRuntime)
	if !viper.GetBool("preload") && !force {
		return false
	}

	// Omit remote check if tarball exists locally
	targetPath := TarballPath(k8sVersion, containerRuntime)
	if _, err := os.Stat(targetPath); err == nil {
		klog.Infof("Found local preload: %s", targetPath)
		return true
	}

	url := remoteTarballURL(k8sVersion, containerRuntime)
	resp, err := http.Head(url)
	if err != nil {
		klog.Warningf("%s fetch error: %v", url, err)
		return false
	}

	// note: err won't be set if it's a 404
	if resp.StatusCode != 200 {
		klog.Warningf("%s status code: %d", url, resp.StatusCode)
		return false
	}

	klog.Infof("Found remote preload: %s", url)
	return true
}

// Preload caches the preloaded images tarball on the host machine
func Preload(k8sVersion, containerRuntime string) error {
	targetPath := TarballPath(k8sVersion, containerRuntime)

	if _, err := os.Stat(targetPath); err == nil {
		klog.Infof("Found %s in cache, skipping download", targetPath)
		return nil
	}

	// Make sure we support this k8s version
	if !PreloadExists(k8sVersion, containerRuntime) {
		klog.Infof("Preloaded tarball for k8s version %s does not exist", k8sVersion)
		return nil
	}

	out.Step(style.FileDownload, "Downloading Kubernetes {{.version}} preload ...", false, out.V{"version": k8sVersion})
	url := remoteTarballURL(k8sVersion, containerRuntime)

	if err := download(url, targetPath); err != nil {
		return errors.Wrapf(err, "download failed: %s", url)
	}

	if err := saveChecksumFile(k8sVersion, containerRuntime); err != nil {
		return errors.Wrap(err, "saving checksum file")
	}

	if err := verifyChecksum(k8sVersion, containerRuntime, targetPath); err != nil {
		return errors.Wrap(err, "verify")
	}

	return nil
}

func saveChecksumFile(k8sVersion, containerRuntime string) error {
	klog.Infof("saving checksum for %s ...", TarballName(k8sVersion, containerRuntime))
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
	return ioutil.WriteFile(PreloadChecksumPath(k8sVersion, containerRuntime), checksum, 0o644)
}

// verifyChecksum returns true if the checksum of the local binary matches
// the checksum of the remote binary
func verifyChecksum(k8sVersion, containerRuntime, path string) error {
	klog.Infof("verifying checksumm of %s ...", path)
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
