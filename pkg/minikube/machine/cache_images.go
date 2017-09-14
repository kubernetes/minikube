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

package machine

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/constants"

	"github.com/containers/image/copy"
	"github.com/containers/image/docker"
	"github.com/containers/image/docker/archive"
	"github.com/containers/image/signature"
	"github.com/containers/image/types"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

const tempLoadDir = "/tmp"

func CacheImagesForBootstrapper(version string, clusterBootstrapper string) error {
	images := bootstrapper.GetCachedImageList(version, clusterBootstrapper)

	if err := CacheImages(images, constants.ImageCacheDir); err != nil {
		return errors.Wrapf(err, "Caching images for %s", clusterBootstrapper)
	}

	return nil
}

// CacheImages will cache images on the host
//
// The cache directory currently caches images using the imagename_tag
// For example, gcr.io/google-containers-kube-addon-manager:v6.4-beta.2 would be
// stored at $CACHE_DIR/gcr.io/google-containers/kube-addon-manager_v6.4-beta.2
func CacheImages(images []string, cacheDir string) error {
	var g errgroup.Group
	for _, image := range images {
		image := image
		g.Go(func() error {
			dst := filepath.Join(cacheDir, image)
			dst = sanitizeCacheDir(dst)
			if err := CacheImage(image, dst); err != nil {
				return errors.Wrapf(err, "caching image %s", dst)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "caching images")
	}
	glog.Infoln("Successfully cached all images.")
	return nil
}

func LoadImages(cmd bootstrapper.CommandRunner, images []string, cacheDir string) error {
	var g errgroup.Group
	for _, image := range images {
		image := image
		g.Go(func() error {
			src := filepath.Join(cacheDir, image)
			src = sanitizeCacheDir(src)
			if err := LoadFromCacheBlocking(cmd, src); err != nil {
				return errors.Wrapf(err, "loading image %s", src)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "loading cached images")
	}
	glog.Infoln("Successfully loaded all cached images.")
	return nil
}

// # ParseReference cannot have a : in the directory path
func sanitizeCacheDir(image string) string {
	return strings.Replace(image, ":", "_", -1)
}

func LoadFromCacheBlocking(cmd bootstrapper.CommandRunner, src string) error {
	glog.Infoln("Loading image from cache at ", src)
	filename := filepath.Base(src)
	for {
		if _, err := os.Stat(src); err == nil {
			break
		}
	}
	dst := filepath.Join(tempLoadDir, filename)
	f, err := assets.NewFileAsset(src, tempLoadDir, filename, "0777")
	if err != nil {
		return errors.Wrapf(err, "creating copyable file asset: %s", filename)
	}
	if err := cmd.Copy(f); err != nil {
		return errors.Wrap(err, "transferring cached image")
	}

	dockerLoadCmd := "docker load -i " + dst

	if err := cmd.Run(dockerLoadCmd); err != nil {
		return errors.Wrapf(err, "loading docker image: %s", dst)
	}

	if err := cmd.Run("rm -rf " + dst); err != nil {
		return errors.Wrap(err, "deleting temp docker image location")
	}

	glog.Infof("Successfully loaded image %s from cache", src)
	return nil
}

func getSrcRef(image string) (types.ImageReference, error) {
	srcRef, err := docker.ParseReference("//" + image)
	if err != nil {
		return nil, errors.Wrap(err, "parsing docker image src ref")
	}
	return srcRef, nil
}

func getDstRef(image, dst string) (types.ImageReference, error) {
	dstRef, err := archive.ParseReference(dst + ":" + image)
	if err != nil {
		return nil, errors.Wrap(err, "parsing docker archive dst ref")
	}
	return dstRef, nil
}

func CacheImage(image, dst string) error {
	glog.Infof("Attempting to cache image: %s at %s\n", image, dst)
	if _, err := os.Stat(dst); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return errors.Wrapf(err, "making cache image directory: %s", dst)
	}

	srcRef, err := getSrcRef(image)
	if err != nil {
		return errors.Wrap(err, "creating docker image src ref")
	}

	dstRef, err := getDstRef(image, dst)
	if err != nil {
		return errors.Wrap(err, "creating docker archive dst ref")
	}

	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		return errors.Wrap(err, "getting policy context")
	}

	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return errors.Wrap(err, "making temp dir")
	}
	defer os.RemoveAll(tmp)
	sourceCtx := &types.SystemContext{
		// By default, the image library will try to look at /etc/docker/certs.d
		// As a non-root user, this would result in a permissions error,
		// so, we skip this step by just looking in a newly created tmpdir.
		DockerCertPath: tmp,
	}

	err = copy.Image(policyContext, dstRef, srcRef, &copy.Options{
		SourceCtx: sourceCtx,
	})
	if err != nil {
		return errors.Wrap(err, "copying image")
	}

	return nil
}
