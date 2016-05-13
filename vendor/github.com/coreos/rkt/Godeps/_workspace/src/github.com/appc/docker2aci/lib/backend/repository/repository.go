// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package repository

import (
	"fmt"
	"net/http"
	"path"

	"github.com/appc/docker2aci/lib/common"
	"github.com/appc/docker2aci/lib/types"
	"github.com/appc/spec/schema"
)

type registryVersion int

const (
	registryV1 registryVersion = iota
	registryV2
)

type RepositoryBackend struct {
	repoData          *RepoData
	username          string
	password          string
	insecure          bool
	hostsV2Support    map[string]bool
	hostsV2AuthTokens map[string]map[string]string
	schema            string
	imageManifests    map[types.ParsedDockerURL]v2Manifest
}

func NewRepositoryBackend(username string, password string, insecure bool) *RepositoryBackend {
	return &RepositoryBackend{
		username:          username,
		password:          password,
		insecure:          insecure,
		hostsV2Support:    make(map[string]bool),
		hostsV2AuthTokens: make(map[string]map[string]string),
		imageManifests:    make(map[types.ParsedDockerURL]v2Manifest),
	}
}

func (rb *RepositoryBackend) GetImageInfo(url string) ([]string, *types.ParsedDockerURL, error) {
	dockerURL := common.ParseDockerURL(url)

	var supportsV2, ok bool
	var URLSchema string
	if supportsV2, ok = rb.hostsV2Support[dockerURL.IndexURL]; !ok {
		var err error
		URLSchema, supportsV2, err = rb.supportsRegistry(dockerURL.IndexURL, registryV2)
		if err != nil {
			return nil, nil, err
		}
		rb.schema = URLSchema + "://"
		rb.hostsV2Support[dockerURL.IndexURL] = supportsV2
	}

	if supportsV2 {
		return rb.getImageInfoV2(dockerURL)
	} else {
		URLSchema, supportsV1, err := rb.supportsRegistry(dockerURL.IndexURL, registryV1)
		if err != nil {
			return nil, nil, err
		}
		if !supportsV1 {
			return nil, nil, fmt.Errorf("registry doesn't support API v2 nor v1")
		}
		rb.schema = URLSchema + "://"
		return rb.getImageInfoV1(dockerURL)
	}
}

func (rb *RepositoryBackend) BuildACI(layerNumber int, layerID string, dockerURL *types.ParsedDockerURL, outputDir string, tmpBaseDir string, curPwl []string, compression common.Compression) (string, *schema.ImageManifest, error) {
	if rb.hostsV2Support[dockerURL.IndexURL] {
		return rb.buildACIV2(layerNumber, layerID, dockerURL, outputDir, tmpBaseDir, curPwl, compression)
	} else {
		return rb.buildACIV1(layerNumber, layerID, dockerURL, outputDir, tmpBaseDir, curPwl, compression)
	}
}

func checkRegistryStatus(statusCode int, hdr http.Header, version registryVersion) (bool, error) {
	switch statusCode {
	case http.StatusOK, http.StatusUnauthorized:
		ok := true
		if version == registryV2 {
			// v2 API requires this check
			ok = hdr.Get("Docker-Distribution-API-Version") == "registry/2.0"
		}
		return ok, nil
	case http.StatusNotFound:
		return false, nil
	}

	return false, fmt.Errorf("unexpected http code: %d", statusCode)
}

func (rb *RepositoryBackend) supportsRegistry(indexURL string, version registryVersion) (schema string, ok bool, err error) {
	var URLPath string
	switch version {
	case registryV1:
		// the v1 API defines this URL to check if the registry's status
		URLPath = "v1/_ping"
	case registryV2:
		URLPath = "v2"
	}
	URLStr := path.Join(indexURL, URLPath)

	fetch := func(schema string) (res *http.Response, err error) {
		url := schema + "://" + URLStr
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		rb.setBasicAuth(req)

		client := &http.Client{}
		res, err = client.Do(req)
		return
	}

	schema = "https"
	res, err := fetch(schema)
	if err == nil {
		ok, err = checkRegistryStatus(res.StatusCode, res.Header, version)
		defer res.Body.Close()
	}
	if err != nil || !ok {
		if rb.insecure {
			schema = "http"
			res, err = fetch(schema)
			if err == nil {
				ok, err = checkRegistryStatus(res.StatusCode, res.Header, version)
				defer res.Body.Close()
			}
		}
		return schema, ok, err
	}

	return schema, ok, err
}
