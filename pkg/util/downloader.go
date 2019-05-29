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
	"crypto"
	"net/url"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	download "github.com/jimmidyson/go-download"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
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
	isoPath := filepath.Join(constants.GetMinipath(), "cache", "iso", filepath.Base(isoURL))
	// As this is a file URL there should be no backslashes regardless of platform running on.
	return "file://" + filepath.ToSlash(isoPath)
}

// CacheMinikubeISOFromURL downloads the ISO, if it doesn't exist in cache
func (f DefaultDownloader) CacheMinikubeISOFromURL(isoURL string) error {
	if !f.ShouldCacheMinikubeISO(isoURL) {
		glog.Infof("Not caching ISO, using %s", isoURL)
		return nil
	}

	options := download.FileOptions{
		Mkdirs: download.MkdirAll,
		Options: download.Options{
			ProgressBars: &download.ProgressBarOptions{
				MaxWidth: 80,
			},
		},
	}

	// Validate the ISO if it was the default URL, before writing it to disk.
	if isoURL == constants.DefaultISOURL {
		options.Checksum = constants.DefaultISOSHAURL
		options.ChecksumHash = crypto.SHA256
	}

	console.OutStyle("iso-download", "Downloading Minikube ISO ...")
	if err := download.ToFile(isoURL, f.GetISOCacheFilepath(isoURL), options); err != nil {
		return errors.Wrap(err, isoURL)
	}

	return nil
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
	return filepath.Join(constants.GetMinipath(), "cache", "iso", filepath.Base(isoURL))
}

// IsMinikubeISOCached returns if an ISO exists in the local cache
func (f DefaultDownloader) IsMinikubeISOCached(isoURL string) bool {
	if _, err := os.Stat(f.GetISOCacheFilepath(isoURL)); os.IsNotExist(err) {
		return false
	}
	return true
}
