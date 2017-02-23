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

package cluster

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	download "github.com/jimmidyson/go-download"
	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

// localkubeCacher is a struct with methods designed for caching localkube
type localkubeCacher struct {
	k8sConf KubernetesConfig
}

func (l *localkubeCacher) getLocalkubeCacheFilepath() string {
	return filepath.Join(constants.GetMinipath(), "cache", "localkube",
		filepath.Base(url.QueryEscape("localkube-"+l.k8sConf.KubernetesVersion)))
}

func (l *localkubeCacher) isLocalkubeCached() bool {
	if _, err := os.Stat(l.getLocalkubeCacheFilepath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (l *localkubeCacher) downloadAndCacheLocalkube() error {
	err := errors.New("")
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
	return download.ToFile(url, l.getLocalkubeCacheFilepath(), opts)
}

func (l *localkubeCacher) fetchLocalkubeFromURI() (assets.CopyableFile, error) {
	urlObj, err := url.Parse(l.k8sConf.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing --kubernetes-version url")
	}
	if urlObj.Scheme == fileScheme {
		return l.genLocalkubeFileFromFile()
	}
	return l.genLocalkubeFileFromURL()
}

func (l *localkubeCacher) genLocalkubeFileFromURL() (assets.CopyableFile, error) {
	if !l.isLocalkubeCached() {
		if err := l.downloadAndCacheLocalkube(); err != nil {
			return nil, errors.Wrap(err, "Error attempting to download and cache localkube")
		}
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
