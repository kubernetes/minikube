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

	"github.com/cheggaaa/pb/v3"
	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
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

var (
	useDaemon = true
	useRemote = true
)

// UseDaemon is if we should look in local daemon for image ref
func UseDaemon(use bool) {
	useDaemon = use
}

// UseRemote is if we should look in remote registry for image ref
func UseRemote(use bool) {
	useRemote = use
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

	img, _, err := retrieveImage(ref, imgName)
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
	klog.Infof("Checking for %s in local docker daemon", img)
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
	// buffered channel
	c := make(chan v1.Update, 200)

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
		_, err = daemon.Write(ref, i, tarball.WithProgress(c))
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
				return errors.Wrap(err, "writing daemon image")
			}
			return nil
		}
	}
}

func retrieveImage(ref name.Reference, imgName string) (v1.Image, string, error) {
	var err error
	var img v1.Image

	if !useDaemon && !useRemote {
		return nil, "", fmt.Errorf("neither daemon nor remote")
	}

	klog.Infof("retrieving image: %+v", ref)
	if useDaemon {
		if useRemote {
			klog.Infof("checking repository: %+v", ref.Context())
			_, err := remote.Head(ref)
			if err == nil {
				imgName = ref.Name()
				klog.Infof("canonical name: %s", imgName)
			}
			if err != nil {
				klog.Warningf("remote: %v", err)
				klog.Infof("short name: %s", imgName)
			}
		}
		img, err = retrieveDaemon(ref)
		if err == nil {
			return img, imgName, nil
		}
	}
	if useRemote {
		img, err = retrieveRemote(ref, defaultPlatform)
		if err == nil {
			img, err = fixPlatform(ref, img, defaultPlatform)
			if err == nil {
				return img, ref.Name(), nil
			}
		}
	}

	return nil, "", err
}

func retrieveDaemon(ref name.Reference) (v1.Image, error) {
	img, err := daemon.Image(ref)
	if err == nil {
		klog.Infof("found %s locally: %+v", ref.Name(), img)
		return img, nil
	}
	// reference does not exist in the local daemon
	klog.Infof("daemon lookup for %+v: %v", ref, err)
	return img, err
}

func retrieveRemote(ref name.Reference, p v1.Platform) (v1.Image, error) {
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain), remote.WithPlatform(p))
	if err == nil {
		return img, nil
	}

	klog.Warningf("authn lookup for %+v (trying anon): %+v", ref, err)
	img, err = remote.Image(ref, remote.WithPlatform(p))
	// reference does not exist in the remote registry
	if err != nil {
		klog.Infof("remote lookup for %+v: %v", ref, err)
	}
	return img, err
}

// See https://github.com/kubernetes/minikube/issues/10402
// check if downloaded image Architecture field matches the requested and fix it otherwise
func fixPlatform(ref name.Reference, img v1.Image, p v1.Platform) (v1.Image, error) {
	cfg, err := img.ConfigFile()
	if err != nil {
		klog.Warningf("failed to get config for %s: %v", ref, err)
		return img, err
	}

	if cfg.Architecture == p.Architecture {
		return img, nil
	}
	klog.Warningf("image %s arch mismatch: want %s got %s. fixing",
		ref, p.Architecture, cfg.Architecture)

	cfg.Architecture = p.Architecture
	img, err = mutate.ConfigFile(img, cfg)
	if err != nil {
		klog.Warningf("failed to change config for %s: %v", ref, err)
		return img, errors.Wrap(err, "failed to change image config")
	}
	return img, nil
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
