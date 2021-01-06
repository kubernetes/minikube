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

package image

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/juju/mutex"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/util/lock"
)

// DeleteFromCacheDir deletes tar files stored in cache dir
func DeleteFromCacheDir(images []string) error {
	for _, image := range images {
		path := filepath.Join(constants.ImageCacheDir, image)
		path = localpath.SanitizeCacheDir(path)
		klog.Infoln("Deleting image in cache at ", path)
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return cleanImageCacheDir()
}

// SaveToDir will cache images on the host
//
// The cache directory currently caches images using the imagename_tag
// For example, k8s.gcr.io/kube-addon-manager:v6.5 would be
// stored at $CACHE_DIR/k8s.gcr.io/kube-addon-manager_v6.5
func SaveToDir(images []string, cacheDir string) error {
	var g errgroup.Group
	for _, image := range images {
		image := image
		g.Go(func() error {
			dst := filepath.Join(cacheDir, image)
			dst = localpath.SanitizeCacheDir(dst)
			if err := saveToTarFile(image, dst); err != nil {
				klog.Errorf("save image to file %q -> %q failed: %v", image, dst, err)
				return errors.Wrapf(err, "caching image %q", dst)
			}
			klog.Infof("save to tar file %s -> %s succeeded", image, dst)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "caching images")
	}
	klog.Infoln("Successfully saved all images to host disk.")
	return nil
}

// saveToTarFile caches an image
func saveToTarFile(iname, rawDest string) error {
	iname = normalizeTagName(iname)
	start := time.Now()
	defer func() {
		klog.Infof("cache image %q -> %q took %s", iname, rawDest, time.Since(start))
	}()

	// OS-specific mangling of destination path
	dst, err := localpath.DstPath(rawDest)
	if err != nil {
		return errors.Wrap(err, "getting destination path")
	}

	spec := lock.PathMutexSpec(dst)
	spec.Timeout = 10 * time.Minute
	klog.Infof("acquiring lock: %+v", spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return errors.Wrapf(err, "unable to acquire lock for %+v", spec)
	}
	defer releaser.Release()

	if _, err := os.Stat(dst); err == nil {
		klog.Infof("%s exists", dst)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return errors.Wrapf(err, "making cache image directory: %s", dst)
	}

	ref, err := name.ParseReference(iname, name.WeakValidation)
	if err != nil {
		return errors.Wrapf(err, "parsing image ref name for %s", iname)
	}
	if ref == nil {
		return errors.Wrapf(err, "nil reference for %s", iname)
	}

	img, err := retrieveImage(ref)
	if err != nil {
		klog.Warningf("unable to retrieve image: %v", err)
	}
	if img == nil {
		return errors.Wrapf(err, "nil image for %s", iname)
	}

	err = writeImage(img, dst, ref)
	if err != nil {
		return err
	}

	klog.Infof("%s exists", dst)
	return nil
}

func writeImage(img v1.Image, dst string, ref name.Reference) error {
	klog.Infoln("opening: ", dst)
	f, err := ioutil.TempFile(filepath.Dir(dst), filepath.Base(dst)+".*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		// If we left behind a temp file, remove it.
		_, err := os.Stat(f.Name())
		if err == nil {
			err = os.Remove(f.Name())
			if err != nil {
				klog.Warningf("failed to clean up the temp file %s: %v", f.Name(), err)
			}
		}
	}()

	err = tarball.Write(ref, img, f)
	if err != nil {
		return errors.Wrap(err, "write")
	}
	err = f.Close()
	if err != nil {
		return errors.Wrap(err, "close")
	}
	err = os.Rename(f.Name(), dst)
	if err != nil {
		return errors.Wrap(err, "rename")
	}
	return nil
}
