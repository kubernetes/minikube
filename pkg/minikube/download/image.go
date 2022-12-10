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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/localpath"
)

var (
	defaultPlatform = v1.Platform{
		Architecture: runtime.GOARCH,
		OS:           "linux",
	}
)

// imagePathInCache returns path in local cache directory
func imagePathInMinikubeCache(img string) string {
	f := filepath.Join(detect.KICCacheDir(), path.Base(img)+".tar")
	f = localpath.SanitizeCacheDir(f)
	return f
}

// ImageExistsInCache if img exist in local minikube cache directory
func ImageExistsInMinikubeCache(img string) bool {
	f := imagePathInMinikubeCache(img)

	// Check if image exists locally
	klog.Infof("Checking for %s in local minikube cache directory", img)
	if st, err := os.Stat(f); err == nil {
		if st.Size() > 0 {
			klog.Infof("Found %s in local minikube cache directory, skipping pull", img)
			return true
		}
	}
	// Else, pull it
	return false
}

var checkImageExistsInMinikubeCache = ImageExistsInMinikubeCache

// ImageExistsInKICDriver
// checks for the specified image in the container engine's local cache
func ImageExistsInKicDriver(ociBin, img string) bool {
	klog.Infof("Checking for %s in local KICdriver's cache", img)
	inCache := oci.IsImageInCache(ociBin, img)
	if inCache {
		klog.Infof("Found %s in local KICdriver's cache, skipping pull", img)
		return true
	}
	return false
}

// ImageToCache
// downloads specified container image in tar format, to local minikube's cache
// does nothing if image is already present.
func ImageToMinikubeCache(img string) error {
	tag, ref, err := parseImage(img)
	// do not use cache if image is set in format <name>:latest
	if _, ok := ref.(name.Tag); ok {
		if tag.Name() == "latest" {
			return fmt.Errorf("can't cache 'latest' tag")
		}
	}

	f := imagePathInMinikubeCache(img)
	fileLock := f + ".lock"

	releaser, err := lockDownload(fileLock)
	if releaser != nil {
		defer releaser.Release()
	}
	if err != nil {
		return err
	}

	if checkImageExistsInMinikubeCache(img) {
		klog.Infof("%s exists in minikube cache, skipping pull", img)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(f), 0777); err != nil {
		return errors.Wrapf(err, "making minikube cache image directory: %s", f)
	}

	if DownloadMock != nil {
		klog.Infof("Mock download: %s -> %s", img, f)
		return DownloadMock(img, f)
	}

	// buffered channel
	c := make(chan v1.Update, 200)

	klog.Infof("Writing %s to local minikube cache", img)
	klog.V(3).Infof("Getting image %v", ref)
	i, err := remote.Image(ref, remote.WithPlatform(defaultPlatform))
	if err != nil {
		if strings.Contains(err.Error(), "GitHub Docker Registry needs login") {
			ErrGithubNeedsLogin := errors.New(err.Error())
			return ErrGithubNeedsLogin
		} else if strings.Contains(err.Error(), "UNAUTHORIZED") {
			ErrNeedsLogin := errors.New(err.Error())
			return ErrNeedsLogin
		}

		return errors.Wrap(err, "getting remote image")
	}
	klog.V(3).Infof("Writing image %v", tag)
	errchan := make(chan error)
	p := pb.Full.Start64(0)
	fn := strings.Split(ref.Name(), "@")[0]
	// abbreviate filename for progress
	maxwidth := 30 - len("...")
	if len(fn) > maxwidth {
		fn = fn[0:maxwidth] + "..."
	}
	p.Set("prefix", "    > "+fn+": ")
	p.Set(pb.Bytes, true)

	// Just a hair less than 80 (standard terminal width) for aesthetics & pasting into docs
	p.SetWidth(79)

	go func() {
		err = tarball.WriteToFile(f, tag, i, tarball.WithProgress(c))
		errchan <- err
	}()
	var update v1.Update
	for {
		select {
		case update = <-c:
			p.SetCurrent(update.Complete)
			p.SetTotal(update.Total)
		case err = <-errchan:
			p.Finish()
			if err != nil {
				return errors.Wrap(err, "writing tarball image")
			}
			return nil
		}
	}
}

func parseImage(img string) (*name.Tag, name.Reference, error) {

	var ref name.Reference
	tag, err := name.NewTag(strings.Split(img, "@")[0])
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse image reference")
	}
	digest, err := name.NewDigest(img)
	if err != nil {
		_, ok := err.(*name.ErrBadName)
		if !ok {
			return nil, nil, errors.Wrap(err, "new ref")
		}
		// ErrBadName means img contains no digest
		// It happens if its value is name:tag for example.
		ref = tag
	} else {
		ref = digest
	}
	return &tag, ref, nil
}

// CacheToKICDriver
// loads a locally minikube-cached container image, to the KIC-driver's cache
func CacheToKicDriver(ociBin string, img string) error {
	p := imagePathInMinikubeCache(img)
	err := oci.ArchiveToDriverCache(ociBin, p)
	return err
}

// ImageToKicDriver
// Makes a direct pull of the specified image to the kicdriver's cache
// maintaining reference to the image digest.
func ImageToKicDriver(ociBin, img string) error {
	_, ref, err := parseImage(img)
	if err != nil {
		return err
	}

	fileLock := filepath.Join(detect.KICCacheDir(), path.Base(img)+".d.lock")
	fileLock = localpath.SanitizeCacheDir(fileLock)
	releaser, err := lockDownload(fileLock)
	if releaser != nil {
		defer releaser.Release()
	}
	if err != nil {
		return err
	}

	if ImageExistsInKicDriver(ociBin, img) {
		klog.Infof("%s exists in KicDriver, skipping pull", img)
		return nil
	}

	if DownloadMock != nil {
		klog.Infof("Mock download: %s -> daemon", img)
		return DownloadMock(img, "daemon")
	}

	klog.V(3).Infof("Pulling image %v", ref)
	// an image pull for the digest at this point is not a bad thing..
	// images are pulled by layers and we already have the biggest part
	if err := oci.PullImage(ociBin, img); err != nil {
		return errors.Wrap(err, "pulling remote image")
	}
	return nil
}
