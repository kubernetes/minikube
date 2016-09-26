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
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/util"
)

func updateLocalkubeFromAsset(client *ssh.Client) error {
	contents, err := Asset("out/localkube")
	if err != nil {
		return errors.Wrap(err, "Error loading asset out/localkube")
	}
	if err := sshutil.Transfer(bytes.NewReader(contents), len(contents), "/usr/local/bin",
		"localkube", "0777", client); err != nil {
		return errors.Wrap(err, "Error transferring localkube via ssh")
	}
	return nil
}

// localkubeCacher is a struct with methods designed for caching localkube
type localkubeCacher struct {
	k8sConf KubernetesConfig
}

func (l *localkubeCacher) getLocalkubeCacheFilepath() string {
	return filepath.Join(constants.Minipath, "cache", "localkube",
		filepath.Base(url.QueryEscape("localkube-"+l.k8sConf.KubernetesVersion)))
}

func (l *localkubeCacher) updateLocalkubeFromURI(client *ssh.Client) error {
	urlObj, err := url.Parse(l.k8sConf.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "Error parsing --kubernetes-version url")
	}
	if urlObj.Scheme == fileScheme {
		return l.updateLocalkubeFromFile(client)
	} else {
		return l.updateLocalkubeFromURL(client)
	}
}

func (l *localkubeCacher) updateLocalkubeFromURL(client *ssh.Client) error {
	url, err := util.GetLocalkubeDownloadURL(l.k8sConf.KubernetesVersion, constants.LocalkubeLinuxFilename)
	if err != nil {
		return errors.Wrap(err, "Error parsing localkube download url:")
	}

	var cache util.DiskCache
	localkubeCache := util.CacheItem{
		FilePath: l.getLocalkubeCacheFilepath(),
		URL:      url,
		ShaURL:   url + ".sha256",
	}
	f, err := cache.GetFile(localkubeCache)
	if err != nil {
		return errors.Wrap(err, "Error getting localkube from file")
	}
	defer f.Close()
	if err = l.transferCachedLocalkubeToVM(client); err != nil {
		return errors.Wrap(err, "Error transferring cached localkube to VM")
	}
	return nil
}

func (l *localkubeCacher) transferCachedLocalkubeToVM(client *ssh.Client) error {
	contents, err := ioutil.ReadFile(l.getLocalkubeCacheFilepath())
	if err != nil {
		return errors.Wrap(err, "Error reading file: localkube cache filepath")
	}

	if err = sshutil.Transfer(bytes.NewReader(contents), len(contents), "/usr/local/bin",
		"localkube", "0777", client); err != nil {
		return errors.Wrap(err, "Error transferring cached localkube to VM via ssh")
	}
	return nil
}

func (l *localkubeCacher) updateLocalkubeFromFile(client *ssh.Client) error {
	path := strings.TrimPrefix(l.k8sConf.KubernetesVersion, "file://")
	path = filepath.FromSlash(path)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "Error reading localkube file at %s", path)
	}
	if err := sshutil.Transfer(bytes.NewReader(contents), len(contents), "/usr/local/bin",
		"localkube", "0777", client); err != nil {
		return errors.Wrapf(err, "Error transferring specified localkube file at %s to VM via ssh", path)
	}
	return nil
}
