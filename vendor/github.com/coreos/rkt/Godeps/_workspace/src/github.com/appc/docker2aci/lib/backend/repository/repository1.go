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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/appc/docker2aci/lib/common"
	"github.com/appc/docker2aci/lib/types"
	"github.com/appc/docker2aci/lib/util"
	"github.com/appc/spec/schema"
	"github.com/coreos/ioprogress"
)

type RepoData struct {
	Tokens    []string
	Endpoints []string
	Cookie    []string
}

func (rb *RepositoryBackend) getImageInfoV1(dockerURL *types.ParsedDockerURL) ([]string, *types.ParsedDockerURL, error) {
	repoData, err := rb.getRepoDataV1(dockerURL.IndexURL, dockerURL.ImageName)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting repository data: %v", err)
	}

	// TODO(iaguis) check more endpoints
	appImageID, err := rb.getImageIDFromTagV1(repoData.Endpoints[0], dockerURL.ImageName, dockerURL.Tag, repoData)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting ImageID from tag %s: %v", dockerURL.Tag, err)
	}

	ancestry, err := rb.getAncestryV1(appImageID, repoData.Endpoints[0], repoData)
	if err != nil {
		return nil, nil, err
	}

	rb.repoData = repoData

	return ancestry, dockerURL, nil
}

func (rb *RepositoryBackend) buildACIV1(layerNumber int, layerID string, dockerURL *types.ParsedDockerURL, outputDir string, tmpBaseDir string, curPwl []string, compression common.Compression) (string, *schema.ImageManifest, error) {
	tmpDir, err := ioutil.TempDir(tmpBaseDir, "docker2aci-")
	if err != nil {
		return "", nil, fmt.Errorf("error creating dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	j, size, err := rb.getJsonV1(layerID, rb.repoData.Endpoints[0], rb.repoData)
	if err != nil {
		return "", nil, fmt.Errorf("error getting image json: %v", err)
	}

	layerData := types.DockerImageData{}
	if err := json.Unmarshal(j, &layerData); err != nil {
		return "", nil, fmt.Errorf("error unmarshaling layer data: %v", err)
	}

	layerFile, err := rb.getLayerV1(layerID, rb.repoData.Endpoints[0], rb.repoData, size, tmpDir)
	if err != nil {
		return "", nil, fmt.Errorf("error getting the remote layer: %v", err)
	}
	defer layerFile.Close()

	util.Debug("Generating layer ACI...")
	aciPath, manifest, err := common.GenerateACI(layerNumber, layerData, dockerURL, outputDir, layerFile, curPwl, compression)
	if err != nil {
		return "", nil, fmt.Errorf("error generating ACI: %v", err)
	}

	return aciPath, manifest, nil
}

func (rb *RepositoryBackend) getRepoDataV1(indexURL string, remote string) (*RepoData, error) {
	client := &http.Client{}
	repositoryURL := rb.schema + path.Join(indexURL, "v1", "repositories", remote, "images")

	req, err := http.NewRequest("GET", repositoryURL, nil)
	if err != nil {
		return nil, err
	}

	if rb.username != "" && rb.password != "" {
		req.SetBasicAuth(rb.username, rb.password)
	}

	req.Header.Set("X-Docker-Token", "true")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP code: %d, URL: %s", res.StatusCode, req.URL)
	}

	var tokens []string
	if res.Header.Get("X-Docker-Token") != "" {
		tokens = res.Header["X-Docker-Token"]
	}

	var cookies []string
	if res.Header.Get("Set-Cookie") != "" {
		cookies = res.Header["Set-Cookie"]
	}

	var endpoints []string
	if res.Header.Get("X-Docker-Endpoints") != "" {
		endpoints = makeEndpointsListV1(res.Header["X-Docker-Endpoints"])
	} else {
		// Assume same endpoint
		endpoints = append(endpoints, indexURL)
	}

	return &RepoData{
		Endpoints: endpoints,
		Tokens:    tokens,
		Cookie:    cookies,
	}, nil
}

func (rb *RepositoryBackend) getImageIDFromTagV1(registry string, appName string, tag string, repoData *RepoData) (string, error) {
	client := &http.Client{}
	// we get all the tags instead of directly getting the imageID of the
	// requested one (.../tags/TAG) because even though it's specified in the
	// Docker API, some registries (e.g. Google Container Registry) don't
	// implement it.
	req, err := http.NewRequest("GET", rb.schema+path.Join(registry, "repositories", appName, "tags"), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Image ID: %s, URL: %s", err, req.URL)
	}

	setAuthTokenV1(req, repoData.Tokens)
	setCookieV1(req, repoData.Cookie)
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get Image ID: %s, URL: %s", err, req.URL)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("HTTP code: %d. URL: %s", res.StatusCode, req.URL)
	}

	j, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	var tags map[string]string

	if err := json.Unmarshal(j, &tags); err != nil {
		return "", fmt.Errorf("error unmarshaling: %v", err)
	}

	imageID, ok := tags[tag]
	if !ok {
		return "", fmt.Errorf("tag %s not found", tag)
	}

	return imageID, nil
}

func (rb *RepositoryBackend) getAncestryV1(imgID, registry string, repoData *RepoData) ([]string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", rb.schema+path.Join(registry, "images", imgID, "ancestry"), nil)
	if err != nil {
		return nil, err
	}

	setAuthTokenV1(req, repoData.Tokens)
	setCookieV1(req, repoData.Cookie)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP code: %d. URL: %s", res.StatusCode, req.URL)
	}

	var ancestry []string

	j, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read downloaded json: %s (%s)", err, j)
	}

	if err := json.Unmarshal(j, &ancestry); err != nil {
		return nil, fmt.Errorf("error unmarshaling: %v", err)
	}

	return ancestry, nil
}

func (rb *RepositoryBackend) getJsonV1(imgID, registry string, repoData *RepoData) ([]byte, int64, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", rb.schema+path.Join(registry, "images", imgID, "json"), nil)
	if err != nil {
		return nil, -1, err
	}
	setAuthTokenV1(req, repoData.Tokens)
	setCookieV1(req, repoData.Cookie)
	res, err := client.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, -1, fmt.Errorf("HTTP code: %d, URL: %s", res.StatusCode, req.URL)
	}

	imageSize := int64(-1)

	if hdr := res.Header.Get("X-Docker-Size"); hdr != "" {
		imageSize, err = strconv.ParseInt(hdr, 10, 64)
		if err != nil {
			return nil, -1, err
		}
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to read downloaded json: %v (%s)", err, b)
	}

	return b, imageSize, nil
}

func (rb *RepositoryBackend) getLayerV1(imgID, registry string, repoData *RepoData, imgSize int64, tmpDir string) (*os.File, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", rb.schema+path.Join(registry, "images", imgID, "layer"), nil)
	if err != nil {
		return nil, err
	}

	setAuthTokenV1(req, repoData.Tokens)
	setCookieV1(req, repoData.Cookie)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		res.Body.Close()
		return nil, fmt.Errorf("HTTP code: %d. URL: %s", res.StatusCode, req.URL)
	}

	// if we didn't receive the size via X-Docker-Size when we retrieved the
	// layer's json, try Content-Length
	if imgSize == -1 {
		if hdr := res.Header.Get("Content-Length"); hdr != "" {
			imgSize, err = strconv.ParseInt(hdr, 10, 64)
			if err != nil {
				return nil, err
			}
		}
	}

	prefix := "Downloading " + imgID[:12]
	fmtBytesSize := 18
	barSize := int64(80 - len(prefix) - fmtBytesSize)
	bar := ioprogress.DrawTextFormatBarForW(barSize, os.Stderr)
	fmtfunc := func(progress, total int64) string {
		return fmt.Sprintf(
			"%s: %s %s",
			prefix,
			bar(progress, total),
			ioprogress.DrawTextFormatBytes(progress, total),
		)
	}

	progressReader := &ioprogress.Reader{
		Reader:       res.Body,
		Size:         imgSize,
		DrawFunc:     ioprogress.DrawTerminalf(os.Stderr, fmtfunc),
		DrawInterval: 500 * time.Millisecond,
	}

	layerFile, err := ioutil.TempFile(tmpDir, "dockerlayer-")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(layerFile, progressReader)
	if err != nil {
		return nil, err
	}

	if err := layerFile.Sync(); err != nil {
		return nil, err
	}

	return layerFile, nil
}

func setAuthTokenV1(req *http.Request, token []string) {
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Token "+strings.Join(token, ","))
	}
}

func setCookieV1(req *http.Request, cookie []string) {
	if req.Header.Get("Cookie") == "" {
		req.Header.Set("Cookie", strings.Join(cookie, ""))
	}
}

func makeEndpointsListV1(headers []string) []string {
	var endpoints []string

	for _, ep := range headers {
		endpointsList := strings.Split(ep, ",")
		for _, endpointEl := range endpointsList {
			endpoints = append(
				endpoints,
				path.Join(strings.TrimSpace(endpointEl), "v1"))
		}
	}

	return endpoints
}
