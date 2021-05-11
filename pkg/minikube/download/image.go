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
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
)

var (
	defaultPlatform = v1.Platform{
		Architecture: runtime.GOARCH,
		OS:           "linux",
	}
)

// imageExistsInCache if img exist in local cache directory
func imageExistsInCache(img string) bool {
	f := filepath.Join(constants.KICCacheDir, path.Base(img)+".tar")
	f = localpath.SanitizeCacheDir(f)

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

var checkImageExistsInCache = imageExistsInCache

// ImageExistsInDaemon if img exist in local docker daemon
func ImageExistsInDaemon(img string) bool {
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

var checkImageExistsInDaemon = ImageExistsInDaemon

// ImageToCache downloads img (if not present in cache) and writes it to the local cache directory
func ImageToCache(img string) error {
	f := filepath.Join(constants.KICCacheDir, path.Base(img)+".tar")
	f = localpath.SanitizeCacheDir(f)
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
		err = tarball.WriteToFile(f, ref, i, tarball.WithProgress(c))
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

// ImageToDaemon downloads img (if not present in daemon) and writes it to the local docker daemon
func ImageToDaemon(img string) error {
	fileLock := filepath.Join(constants.KICCacheDir, path.Base(img)+".d.lock")
	fileLock = localpath.SanitizeCacheDir(fileLock)

	releaser, err := lockDownload(fileLock)
	if releaser != nil {
		defer releaser.Release()
	}
	if err != nil {
		return err
	}

	if checkImageExistsInDaemon(img) {
		klog.Infof("%s exists in daemon, skipping pull", img)
		return nil
	}
	// buffered channel
	c := make(chan v1.Update, 200)

	klog.Infof("Writing %s to local daemon", img)
	ref, err := name.ParseReference(img)
	if err != nil {
		return errors.Wrap(err, "parsing reference")
	}

	if DownloadMock != nil {
		klog.Infof("Mock download: %s -> daemon", img)
		return DownloadMock(img, "daemon")
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
