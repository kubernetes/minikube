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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
)

var defaultPlatform = v1.Platform{
	Architecture: runtime.GOARCH,
	OS:           "linux",
}

// DigestByDockerLib uses client by docker lib to return image digest
// img.ID in as same as image digest
func DigestByDockerLib(imgClient *client.Client, imgName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	imgClient.NegotiateAPIVersion(ctx)
	img, _, err := imgClient.ImageInspectWithRaw(ctx, imgName)
	if err != nil && !client.IsErrNotFound(err) {
		klog.Infof("couldn't find image digest %s from local daemon: %v ", imgName, err)
		return ""
	}
	return img.ID
}

// DigestByGoLib gets image digest uses go-containerregistry lib
// which is 4s slower thabn local daemon per lookup https://github.com/google/go-containerregistry/issues/627
func DigestByGoLib(imgName string) string {
	ref, err := name.ParseReference(imgName, name.WeakValidation)
	if err != nil {
		klog.Infof("error parsing image name %s ref %v ", imgName, err)
		return ""
	}

	img, err := retrieveImage(ref)
	if err != nil {
		klog.Infof("error retrieve Image %s ref %v ", imgName, err)
		return ""
	}

	cf, err := img.ConfigName()
	if err != nil {
		klog.Infof("error getting Image config name %s %v ", imgName, err)
		return cf.Hex
	}
	return cf.Hex
}

// ExistsImageInDaemon if img exist in local docker daemon
func ExistsImageInDaemon(img string) bool {
	// Check if image exists locally
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}@{{.Digest}}")
	if output, err := cmd.Output(); err == nil {
		if strings.Contains(string(output), img) {
			klog.Infof("Found %s in local docker daemon, skipping pull", img)
			return true
		}
	}
	// Else, pull it
	return false
}

// LoadFromTarball checks if the image exists as a tarball and tries to load it to the local daemon
// TODO: Pass in if we are loading to docker or podman so this function can also be used for podman
func LoadFromTarball(binary, img string) error {
	p := filepath.Join(constants.ImageCacheDir, img)
	p = localpath.SanitizeCacheDir(p)

	switch binary {
	case driver.Podman:
		return fmt.Errorf("not yet implemented, see issue #8426")
	default:
		tag, err := name.NewTag(Tag(img))
		if err != nil {
			return errors.Wrap(err, "new tag")
		}

		i, err := tarball.ImageFromPath(p, &tag)
		if err != nil {
			return errors.Wrap(err, "tarball")
		}

		_, err = daemon.Write(tag, i)
		return err
	}

}

// SaveToTarball saves img as a tarball at the given path
func SaveToTarball(img, path string) error {
	if !ExistsImageInDaemon(img) {
		return fmt.Errorf("%s does not exist in local daemon, can't save to tarball", img)
	}
	ref, err := name.ParseReference(img)
	if err != nil {
		return errors.Wrap(err, "parsing reference")
	}
	i, err := daemon.Image(ref)
	if err != nil {
		return errors.Wrap(err, "getting image")
	}
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "creating tmp path")
	}
	defer f.Close()
	return tarball.Write(ref, i, f)
}

// Tag returns just the image with the tag
// eg image:tag@sha256:digest -> image:tag if there is an associated tag
// if not possible, just return the initial img
func Tag(img string) string {
	split := strings.Split(img, ":")
	if len(split) == 3 {
		tag := strings.Split(split[1], "@")[0]
		return fmt.Sprintf("%s:%s", split[0], tag)
	}
	return img
}

// WriteImageToDaemon write img to the local docker daemon
func WriteImageToDaemon(img string) error {
	klog.Infof("Writing %s to local daemon", img)
	ref, err := name.ParseReference(img)
	if err != nil {
		return errors.Wrap(err, "parsing reference")
	}
	klog.V(3).Infof("Getting image %v", ref)
	i, err := remote.Image(ref)
	if err != nil {
		if strings.Contains(err.Error(), "GitHub Docker Registry needs login") {
			ErrGithubNeedsLogin = errors.New(err.Error())
			return ErrGithubNeedsLogin
		} else if strings.Contains(err.Error(), "UNAUTHORIZED") {
			ErrNeedsLogin = errors.New(err.Error())
			return ErrNeedsLogin
		}

		return errors.Wrap(err, "getting remote image")
	}
	klog.V(3).Infof("Writing image %v", ref)
	_, err = daemon.Write(ref, i)
	if err != nil {
		return errors.Wrap(err, "writing daemon image")
	}

	return nil
}

func retrieveImage(ref name.Reference) (v1.Image, error) {
	klog.Infof("retrieving image: %+v", ref)
	img, err := daemon.Image(ref)
	if err == nil {
		klog.Infof("found %s locally: %+v", ref.Name(), img)
		return img, nil
	}
	// reference does not exist in the local daemon
	if err != nil {
		klog.Infof("daemon lookup for %+v: %v", ref, err)
	}

	platform := defaultPlatform
	img, err = remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain), remote.WithPlatform(platform))
	if err == nil {
		return img, nil
	}

	klog.Warningf("authn lookup for %+v (trying anon): %+v", ref, err)
	img, err = remote.Image(ref)
	return img, err
}

func cleanImageCacheDir() error {
	err := filepath.Walk(constants.ImageCacheDir, func(path string, info os.FileInfo, err error) error {
		// If error is not nil, it's because the path was already deleted and doesn't exist
		// Move on to next path
		if err != nil {
			return nil
		}
		// Check if path is directory
		if !info.IsDir() {
			return nil
		}
		// If directory is empty, delete it
		entries, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			if err = os.Remove(path); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// normalizeTagName automatically tag latest to image
// Example:
//  nginx -> nginx:latest
//  localhost:5000/nginx -> localhost:5000/nginx:latest
//  localhost:5000/nginx:latest -> localhost:5000/nginx:latest
//  docker.io/dotnet/core/sdk -> docker.io/dotnet/core/sdk:latest
func normalizeTagName(image string) string {
	base := image
	tag := "latest"

	// From google/go-containerregistry/pkg/name/tag.go
	parts := strings.Split(strings.TrimSpace(image), ":")
	if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], "/") {
		base = strings.Join(parts[:len(parts)-1], ":")
		tag = parts[len(parts)-1]
	}
	return base + ":" + tag
}
