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

package util

import (
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/hashicorp/go-getter"
	"github.com/juju/mutex"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util/lock"
)

const fileScheme = "file"

// ISODownloader downloads an ISO
type ISODownloader interface {
	GetISOFileURI(isoURL string) string
	CacheMinikubeISOFromURL(isoURL string) error
}

// DefaultDownloader is the default ISODownloader
type DefaultDownloader struct{}

// GetISOFileURI gets the local destination for a remote source
func (f DefaultDownloader) GetISOFileURI(isoURL string) string {
	urlObj, err := url.Parse(isoURL)
	if err != nil {
		return isoURL
	}
	if urlObj.Scheme == fileScheme {
		return isoURL
	}
	isoPath := filepath.Join(localpath.MiniPath(), "cache", "iso", filepath.Base(isoURL))
	// As this is a file URL there should be no backslashes regardless of platform running on.
	return "file://" + filepath.ToSlash(isoPath)
}

// CacheMinikubeISOFromURL downloads the ISO, if it doesn't exist in cache
func (f DefaultDownloader) CacheMinikubeISOFromURL(url string) error {
	dst := f.GetISOCacheFilepath(url)

	// Lock before we check for existence to avoid thundering herd issues
	spec := lock.PathMutexSpec(dst)
	spec.Timeout = 10 * time.Minute
	glog.Infof("acquiring lock: %+v", spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return errors.Wrapf(err, "unable to acquire lock for %+v", spec)
	}
	defer releaser.Release()

	if !f.ShouldCacheMinikubeISO(url) {
		glog.Infof("Not caching ISO, using %s", url)
		return nil
	}

	urlWithChecksum := url
	if url == constants.DefaultISOURL {
		urlWithChecksum = url + "?checksum=file:" + constants.DefaultISOSHAURL
	}

	// Predictable temp destination so that resume can function
	tmpDst := dst + ".download"

	opts := []getter.ClientOption{getter.WithProgress(DefaultProgressBar)}
	client := &getter.Client{
		Src:     urlWithChecksum,
		Dst:     tmpDst,
		Mode:    getter.ClientModeFile,
		Options: opts,
	}

	glog.Infof("full url: %s", urlWithChecksum)
	out.T(out.ISODownload, "Downloading VM boot image ...")
	if err := client.Get(); err != nil {
		return errors.Wrap(err, url)
	}
	return os.Rename(tmpDst, dst)
}

// ShouldCacheMinikubeISO returns if we need to download the ISO
func (f DefaultDownloader) ShouldCacheMinikubeISO(isoURL string) bool {
	// store the minikube-iso inside the .minikube dir

	urlObj, err := url.Parse(isoURL)
	if err != nil {
		return false
	}
	if urlObj.Scheme == fileScheme {
		return false
	}
	if f.IsMinikubeISOCached(isoURL) {
		return false
	}
	return true
}

// GetISOCacheFilepath returns the path of an ISO in the local cache
func (f DefaultDownloader) GetISOCacheFilepath(isoURL string) string {
	return filepath.Join(localpath.MiniPath(), "cache", "iso", filepath.Base(isoURL))
}

// IsMinikubeISOCached returns if an ISO exists in the local cache
func (f DefaultDownloader) IsMinikubeISOCached(isoURL string) bool {
	if _, err := os.Stat(f.GetISOCacheFilepath(isoURL)); os.IsNotExist(err) {
		return false
	}
	return true
}
