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

	"github.com/golang/glog"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// DeleteFromCacheDir deletes tar files stored in cache dir
func DeleteFromCacheDir(images []string) error {
	for _, image := range images {
		path := filepath.Join(constants.ImageCacheDir, image)
		path = localpath.SanitizeCacheDir(path)
		glog.Infoln("Deleting image in cache at ", path)
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return cleanImageCacheDir()
}

// CacheImagesToTar will cache images on the host
//
// The cache directory currently caches images using the imagename_tag
// For example, k8s.gcr.io/kube-addon-manager:v6.5 would be
// stored at $CACHE_DIR/k8s.gcr.io/kube-addon-manager_v6.5
func CacheImagesToTar(images []string, cacheDir string) error {
	var g errgroup.Group
	for _, image := range images {
		image := image
		g.Go(func() error {
			dst := filepath.Join(cacheDir, image)
			dst = localpath.SanitizeCacheDir(dst)
			if err := saveToTarFile(image, dst); err != nil {
				glog.Errorf("CacheImage %s -> %s failed: %v", image, dst, err)
				return errors.Wrapf(err, "caching image %q", dst)
			}
			glog.Infof("CacheImage %s -> %s succeeded", image, dst)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "caching images")
	}
	glog.Infoln("Successfully cached all images.")
	return nil
}

// saveToTarFile caches an image
func saveToTarFile(image, dst string) error {
	start := time.Now()
	glog.Infof("CacheImage: %s -> %s", image, dst)
	defer func() {
		glog.Infof("CacheImage: %s -> %s completed in %s", image, dst, time.Since(start))
	}()

	if _, err := os.Stat(dst); err == nil {
		glog.Infof("%s exists", dst)
		return nil
	}

	dstPath, err := localpath.DstPath(dst)
	if err != nil {
		return errors.Wrap(err, "getting destination path")
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0777); err != nil {
		return errors.Wrapf(err, "making cache image directory: %s", dst)
	}

	ref, err := name.ParseReference(image, name.WeakValidation)
	if err != nil {
		return errors.Wrapf(err, "parsing image ref name for %s", image)
	}

	img, err := retrieveImage(ref)
	if err != nil {
		glog.Warningf("unable to retrieve image: %v", err)
	}
	glog.Infoln("opening: ", dstPath)
	f, err := ioutil.TempFile(filepath.Dir(dstPath), filepath.Base(dstPath)+".*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		// If we left behind a temp file, remove it.
		_, err := os.Stat(f.Name())
		if err == nil {
			os.Remove(f.Name())
			if err != nil {
				glog.Warningf("Failed to clean up the temp file %s: %v", f.Name(), err)
			}
		}
	}()
	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		return errors.Wrap(err, "newtag")
	}
	err = tarball.Write(tag, img, &tarball.WriteOptions{}, f)
	if err != nil {
		return errors.Wrap(err, "write")
	}
	err = f.Close()
	if err != nil {
		return errors.Wrap(err, "close")
	}
	err = os.Rename(f.Name(), dstPath)
	if err != nil {
		return errors.Wrap(err, "rename")
	}
	glog.Infof("%s exists", dst)
	return nil
}
