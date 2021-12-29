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

package assets

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/minikube/deploy/addons"
)

func UnmarshalLoad(url *url.URL, data interface{}) error {
	asset, err := LoadAsset(url, "", "0640", false)
	if err != nil {
		return errors.Wrapf(err, "loading asset %s", url.String())
	}

	configFile, err := ioutil.ReadAll(asset)
	if err != nil {
		return errors.Wrapf(err, "reading asset %s", url.String())
	}

	err = yaml.Unmarshal(configFile, data)
	if err != nil {
		return errors.Wrapf(err, "parsing asset config %s", url.String())
	}

	return nil
}

func LoadAsset(url *url.URL, targetPath, permissions string, isTemplate bool) (*BinAsset, error) {
	scheme := "file"
	if url.Scheme != "" {
		scheme = url.Scheme
	}

	switch scheme {
	case "embedfs":
		return loadBinAsset(url.Host, url.Path, targetPath, permissions, isTemplate)
	default:
		return nil, fmt.Errorf("asset scheme %s is not supported", scheme)
	}
}

func loadBinAsset(packageName, sourcePath, targetPath, permissions string, isTemplate bool) (*BinAsset, error) {
	sourcePath = strings.TrimPrefix(sourcePath, "/")
	fs, ok := addons.Embedded[packageName]

	if ok {
		return NewBinAsset(fs, sourcePath, targetPath, permissions, isTemplate)
	}

	return nil, fmt.Errorf("embedfs package %s does not exist", packageName)
}
