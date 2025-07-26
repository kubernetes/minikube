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
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/juju/mutex/v2"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util/lock"
)

type cacheError struct {
	Err error
}

func (f *cacheError) Error() string {
	return f.Err.Error()
}

// errCacheImageDoesntExist is thrown when image that user is trying to add does not exist
var errCacheImageDoesntExist = &cacheError{errors.New("the image you are trying to add does not exist")}

// DeleteFromCacheDir deletes tar files stored in cache dir
func DeleteFromCacheDir(images []string) error {
	for _, image := range images {
		path := filepath.Join(detect.ImageCacheDir(), image)
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
// For example, registry.k8s.io/kube-addon-manager:v6.5 would be
// stored at $CACHE_DIR/registry.k8s.io/kube-addon-manager_v6.5
func SaveToDir(images []string, cacheDir string, overwrite bool) error {
	var g errgroup.Group
	for _, image := range images {
		image := image
		g.Go(func() error {
			dst := filepath.Join(cacheDir, image)
			dst = localpath.SanitizeCacheDir(dst)
			if err := saveToTarFile(image, dst, overwrite); err != nil {
				if err == errCacheImageDoesntExist {
					out.WarningT("The image '{{.imageName}}' was not found; unable to add it to cache.", out.V{"imageName": image})
					return nil
				}
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
func saveToTarFile(iname, rawDest string, overwrite bool) error {
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

	if _, err := os.Stat(dst); !overwrite && err == nil {
		klog.Infof("%s exists", dst)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return errors.Wrapf(err, "making cache image directory: %s", dst)
	}

	// use given short name
	ref, err := name.ParseReference(iname, name.WeakValidation)
	if err != nil {
		return errors.Wrapf(err, "parsing image ref name for %s", iname)
	}
	if ref == nil {
		return errors.Wrapf(err, "nil reference for %s", iname)
	}

	img, cname, err := retrieveImage(ref, iname)
	if err != nil {
		klog.V(2).ErrorS(err, "an error while retrieving the image")
		return errCacheImageDoesntExist
	}
	if img == nil {
		return errors.Wrapf(err, "nil image for %s", iname)
	}

	if cname != iname {
		// use new canonical name
		ref, err = name.ParseReference(cname, name.WeakValidation)
		if err != nil {
			return errors.Wrapf(err, "parsing image ref name for %s", cname)
		}
		if ref == nil {
			return errors.Wrapf(err, "nil reference for %s", cname)
		}
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
	f, err := os.CreateTemp(filepath.Dir(dst), filepath.Base(dst)+".*.tmp")
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
