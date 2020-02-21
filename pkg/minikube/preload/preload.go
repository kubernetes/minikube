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

package preload

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"cloud.google.com/go/storage"
	"github.com/golang/glog"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util"
)

// returns name of the tarball
func tarballName(k8sVersion string) string {
	return fmt.Sprintf("preloaded-images-k8s-%s.tar.lz4", k8sVersion)
}

// returns the name of the checksum file
func checksumName(k8sVersion string) string {
	return fmt.Sprintf("preloaded-images-k8s-%s.tar.lz4.checksum", k8sVersion)
}

// returns target dir for all cached items related to preloading
func targetDir() string {
	return localpath.MakeMiniPath("cache", "preloaded-tarball")
}

// returns path to checksum file
func checksumFilepath(k8sVersion string) string {
	return path.Join(targetDir(), checksumName(k8sVersion))
}

// TarballFilepath returns the path to the preloaded tarball
func TarballFilepath(k8sVersion string) string {
	return path.Join(targetDir(), tarballName(k8sVersion))
}

// remoteTarballURL returns the URL for the remote tarball in GCS
func remoteTarballURL(k8sVersion string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", constants.PreloadedVolumeTarballsBucket, tarballName(k8sVersion))
}

// CacheTarball caches the preloaded images tarball on the host machine
func CacheTarball(k8sVersion, containerRuntime string) error {
	if containerRuntime != "docker" {
		return nil
	}
	targetFilepath := TarballFilepath(k8sVersion)

	if _, err := os.Stat(targetFilepath); err == nil {
		if err := verifyChecksum(k8sVersion); err == nil {
			glog.Infof("Found %s in cache, skipping downloading", targetFilepath)
			return nil
		}
	}

	url := remoteTarballURL(k8sVersion)

	// Make sure we support this k8s version
	if _, err := http.Get(url); err != nil {
		glog.Infof("Unable to get response from %s, skipping downloading: %v", err)
		return nil
	}

	out.T(out.FileDownload, "Downloading preloaded images tarball for k8s {{.version}}:", out.V{"version": k8sVersion})
	client := &getter.Client{
		Src:     url,
		Dst:     targetFilepath,
		Mode:    getter.ClientModeFile,
		Options: []getter.ClientOption{getter.WithProgress(util.DefaultProgressBar)},
	}

	glog.Infof("Downloading: %+v", client)
	if err := client.Get(); err != nil {
		return errors.Wrapf(err, "download failed: %s", url)
	}
	// Give downloaded drivers a baseline decent file permission
	if err := os.Chmod(targetFilepath, 0755); err != nil {
		return err
	}
	// Save checksum file locally
	if err := saveChecksumFile(k8sVersion); err != nil {
		return errors.Wrap(err, "saving checksum file")
	}
	return verifyChecksum(k8sVersion)
}

func saveChecksumFile(k8sVersion string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	attrs, err := client.Bucket(constants.PreloadedVolumeTarballsBucket).Object(tarballName(k8sVersion)).Attrs(ctx)
	if err != nil {
		return err
	}
	checksum := attrs.MD5
	return ioutil.WriteFile(checksumFilepath(k8sVersion), checksum, 0644)
}

// verifyChecksum returns true if the checksum of the local binary matches
// the checksum of the remote binary
func verifyChecksum(k8sVersion string) error {
	// get md5 checksum of tarball path
	contents, err := ioutil.ReadFile(TarballFilepath(k8sVersion))
	if err != nil {
		return errors.Wrap(err, "reading tarball")
	}
	checksum := md5.Sum(contents)

	remoteChecksum, err := ioutil.ReadFile(checksumFilepath(k8sVersion))
	if err != nil {
		return errors.Wrap(err, "reading checksum file")
	}

	// create a slice of checksum, which is [16]byte
	if string(remoteChecksum) != string(checksum[:]) {
		return fmt.Errorf("checksum of %s does not match remote checksum", TarballFilepath(k8sVersion))
	}
	return nil
}
