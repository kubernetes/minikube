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
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/image"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
)

var (
	defaultPlatform = v1.Platform{
		Architecture: runtime.GOARCH,
		OS:           "linux",
	}
)

// ImagePathInCache returns path in local cache directory
func ImagePathInCache(img string) string {
	f := filepath.Join(detect.KICCacheDir(), path.Base(img)+".tar")
	f = localpath.SanitizeCacheDir(f)
	return f
}

// ImageExistsInCache if img exist in local cache directory
func ImageExistsInCache(img string) bool {
	f := ImagePathInCache(img)

	// Check if image exists locally
	klog.Infof("Checking for %s in local cache directory", img)
	if st, err := os.Stat(f); err == nil {
		if st.Size() > 0 {
			klog.Infof("Found %s in local cache directory, skipping pull", img)
			return true
		}
	}
	// Else, pull it
	return false
}

var checkImageExistsInCache = ImageExistsInCache

// ImageExistsInDaemon if img exist in local docker daemon
func ImageExistsInDaemon(img string) bool {
	// Check if image exists locally
	klog.Infof("Checking for %s in local docker daemon", img)
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}@{{.Digest}}")
	if output, err := cmd.Output(); err == nil {
		if strings.Contains(string(output), image.TrimDockerIO(img)) {
			klog.Infof("Found %s in local docker daemon, skipping pull", img)
			return true
		}
	}
	// Else, pull it
	return false
}

// ImageToCache downloads img (if not present in cache) and writes it to the local cache directory
func ImageToCache(img string) error {
	f := ImagePathInCache(img)
	fileLock := f + ".lock"

	releaser, err := lockDownload(fileLock)
	if releaser != nil {
		defer releaser.Release()
	}
	if err != nil {
		return err
	}

	if checkImageExistsInCache(img) {
		klog.Infof("%s exists in cache, skipping pull", img)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(f), 0777); err != nil {
		return errors.Wrapf(err, "making cache image directory: %s", f)
	}

	if DownloadMock != nil {
		klog.Infof("Mock download: %s -> %s", img, f)
		return DownloadMock(img, f)
	}

	// buffered channel
	c := make(chan v1.Update, 200)

	klog.Infof("Writing %s to local cache", img)
	ref, err := name.ParseReference(img)
	if err != nil {
		return errors.Wrap(err, "parsing reference")
	}
	tag, err := name.NewTag(image.Tag(img))
	if err != nil {
		return errors.Wrap(err, "parsing tag")
	}
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
	if out.JSON {
		if err := tarball.WriteToFile(f, tag, i); err != nil {
			return errors.Wrap(err, "writing tarball image")
		}
		return nil
	}
	errchan := make(chan error)
	p := pb.Full.Start64(0)
	fn := image.Tag(ref.Name())
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
	tag, err := name.NewTag(image.Tag(img))
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

// CacheToKicDriver
// loads a kicBase cached image to the OCIBIN kicDriver
func CacheToKicDriver(ociBin, img string) error {
	tarpath := ImagePathInCache(img)
	return oci.LoadTarball(ociBin, tarpath)
}

// CacheToDaemon loads image from tarball in the local cache directory to the local docker daemon
// It returns the img that was loaded into the daemon
// If online it will be: image:tag@sha256
// If offline it will be: image:tag
func CacheToDaemon(img string) (string, error) {
	p := ImagePathInCache(img)

	tag, ref, err := parseImage(img)
	if err != nil {
		return "", err
	}
	// do not use cache if image is set in format <name>:latest
	if _, ok := ref.(name.Tag); ok {
		if tag.Name() == "latest" {
			return "", fmt.Errorf("can't cache 'latest' tag")
		}
	}

	i, err := tarball.ImageFromPath(p, tag)
	if err != nil {
		return "", errors.Wrap(err, "tarball")
	}

	resp, err := daemon.Write(*tag, i)
	klog.V(2).Infof("response: %s", resp)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("docker", "pull", "--quiet", img)
	if output, err := cmd.CombinedOutput(); err != nil {
		klog.Warningf("failed to pull image digest (expected if offline): %s: %v", output, err)
		img = image.Tag(img)
	}

	return img, nil
}
