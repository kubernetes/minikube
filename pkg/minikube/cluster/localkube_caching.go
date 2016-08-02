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
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"golang.org/x/crypto/ssh"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/util"
)

// localkubeCacher is a struct with methods designed for caching localkube
type localkubeCacher struct {
	k8sConf KubernetesConfig
}

func (l *localkubeCacher) getLocalkubeCacheFilepath() string {
	return filepath.Join(constants.Minipath, "cache", "localkube",
		filepath.Base(url.QueryEscape("localkube-"+l.k8sConf.KubernetesVersion)))
}

func (l *localkubeCacher) isLocalkubeCached() bool {
	if _, err := os.Stat(l.getLocalkubeCacheFilepath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (l *localkubeCacher) cacheLocalkube(body io.ReadCloser) error {
	// store localkube inside the .minikube dir
	out, err := os.Create(l.getLocalkubeCacheFilepath())
	if err != nil {
		return err
	}
	defer out.Close()
	defer body.Close()
	if _, err = io.Copy(out, body); err != nil {
		return err
	}
	return nil
}

func (l *localkubeCacher) downloadAndCacheLocalkube() error {
	resp := &http.Response{}
	err := errors.New("")
	downloader := func() (err error) {
		url, err := util.GetLocalkubeDownloadURL(l.k8sConf.KubernetesVersion,
			constants.LocalkubeLinuxFilename)
		if err != nil {
			return err
		}
		resp, err = http.Get(url)
		return err
	}

	if err = util.Retry(5, downloader); err != nil {
		return err
	}
	if err = l.cacheLocalkube(resp.Body); err != nil {
		return err
	}
	return nil
}

func updateLocalkubeFromURL(lCacher localkubeCacher, client *ssh.Client) error {
	if !lCacher.isLocalkubeCached() {
		if err := lCacher.downloadAndCacheLocalkube(); err != nil {
			return err
		}
	}
	if err := lCacher.transferCachedLocalkubeToVM(client); err != nil {
		return err
	}
	return nil
}

func (l *localkubeCacher) transferCachedLocalkubeToVM(client *ssh.Client) error {
	contents, err := ioutil.ReadFile(l.getLocalkubeCacheFilepath())
	if err != nil {
		glog.Infof("Error loading asset out/localkube: %s", err)
		return err
	}

	if err = sshutil.Transfer(bytes.NewReader(contents), len(contents), "/usr/local/bin",
		"localkube", "0777", client); err != nil {
		return err
	}
	return nil
}

func updateLocalkubeFromAsset(client *ssh.Client) error {
	contents, err := Asset("out/localkube")
	if err != nil {
		glog.Infof("Error loading asset out/localkube: %s", err)
		return err
	}
	if err := sshutil.Transfer(bytes.NewReader(contents), len(contents), "/usr/local/bin",
		"localkube", "0777", client); err != nil {
		return err
	}
	return nil
}
