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

package localkube

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	download "github.com/jimmidyson/go-download"
	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

// localkubeCacher is a struct with methods designed for caching localkube
type localkubeCacher struct {
	k8sConf bootstrapper.KubernetesConfig
}

func (l *localkubeCacher) getLocalkubeCacheFilepath() string {
	return filepath.Join(constants.GetMinipath(), "cache", "localkube",
		filepath.Base(url.QueryEscape("localkube-"+l.k8sConf.KubernetesVersion)))
}

func (l *localkubeCacher) getLocalkubeSha256CacheFilepath() string {
	return l.getLocalkubeCacheFilepath() + ".sha256"
}

func localkubeURIWasSpecified(config bootstrapper.KubernetesConfig) bool {
	// see if flag is different than default -> it was passed by user
	return config.KubernetesVersion != constants.DefaultKubernetesVersion
}

func (l *localkubeCacher) isLocalkubeCached() bool {
	url, err := util.GetLocalkubeDownloadURL(l.k8sConf.KubernetesVersion, constants.LocalkubeLinuxFilename)
	if err != nil {
		glog.Warningf("Unable to get localkube checksum url...continuing.")
		return true
	}
	opts := download.FileOptions{
		Mkdirs: download.MkdirAll,
	}

	if err := download.ToFile(url+".sha256", l.getLocalkubeSha256CacheFilepath(), opts); err != nil {
		glog.Warningf("Unable to check localkube checksum... continuing.")
		return true
	}

	if _, err := os.Stat(l.getLocalkubeCacheFilepath()); os.IsNotExist(err) {
		return false
	}

	localkubeSha256, err := ioutil.ReadFile(l.getLocalkubeSha256CacheFilepath())
	if err != nil {
		glog.Infof("Error reading localkube checksum: %s", err)
		return false
	}

	h := sha256.New()
	f, err := os.Open(l.getLocalkubeCacheFilepath())
	if err != nil {
		glog.Infof("Error opening localkube for checksum verification: %s", err)
		return false
	}
	if _, err := io.Copy(h, f); err != nil {
		glog.Infof("Error copying contents to hasher: %s", err)
	}

	actualChecksum := hex.EncodeToString(h.Sum(nil))
	if strings.TrimSpace(string(localkubeSha256)) != actualChecksum {
		glog.Infof("Localkube checksums do not match actual: %s expected: %s", actualChecksum, string(localkubeSha256))
		return false
	}
	return true
}

func (l *localkubeCacher) downloadAndCacheLocalkube() error {
	url, err := util.GetLocalkubeDownloadURL(l.k8sConf.KubernetesVersion, constants.LocalkubeLinuxFilename)
	if err != nil {
		return errors.Wrap(err, "Error getting localkube download url")
	}
	opts := download.FileOptions{
		Mkdirs: download.MkdirAll,
		Options: download.Options{
			ProgressBars: &download.ProgressBarOptions{
				MaxWidth: 80,
			},
		},
	}
	fmt.Println("Downloading localkube binary")
	if err := download.ToFile(url, l.getLocalkubeCacheFilepath(), opts); err != nil {
		return errors.Wrap(err, "downloading localkube")
	}
	if err := download.ToFile(url+".sha256", l.getLocalkubeSha256CacheFilepath(), opts); err != nil {
		return errors.Wrap(err, "downloading localkube checksum")
	}

	return nil
}

func (l *localkubeCacher) fetchLocalkubeFromURI() (assets.CopyableFile, error) {
	urlObj, err := url.Parse(l.k8sConf.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing --kubernetes-version url")
	}
	if urlObj.Scheme == constants.FileScheme {
		return l.genLocalkubeFileFromFile()
	}
	return l.genLocalkubeFileFromURL()
}

func (l *localkubeCacher) genLocalkubeFileFromURL() (assets.CopyableFile, error) {
	if !l.isLocalkubeCached() {
		glog.Infoln("Localkube not cached or checksum does not match, downloading...")
		if err := l.downloadAndCacheLocalkube(); err != nil {
			return nil, errors.Wrap(err, "Error attempting to download and cache localkube")
		}
	} else {
		glog.Infoln("Using cached localkube")
	}
	localkubeFile, err := assets.NewFileAsset(l.getLocalkubeCacheFilepath(), "/usr/local/bin", "localkube", "0777")
	if err != nil {
		return nil, errors.Wrap(err, "Error creating localkube asset from url")
	}
	return localkubeFile, nil
}

func (l *localkubeCacher) genLocalkubeFileFromFile() (assets.CopyableFile, error) {
	path := strings.TrimPrefix(l.k8sConf.KubernetesVersion, "file://")
	path = filepath.FromSlash(path)
	localkubeFile, err := assets.NewFileAsset(path, "/usr/local/bin", "localkube", "0777")
	if err != nil {
		return nil, errors.Wrap(err, "Error creating localkube asset from file")
	}
	return localkubeFile, nil
}
