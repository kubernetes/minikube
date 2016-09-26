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
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type FileCache interface {
	GetFile(CacheItem) (*os.File, error)
	SetFile(CacheItem, *[]byte) error
}

type CacheItem struct {
	FilePath string
	URL      string
	ShaURL   string
}

type DiskCache struct{}

func (d *DiskCache) GetFile(c CacheItem) (*os.File, error) {
	if c.isFileCached() {
		return os.Open(c.FilePath)
	}

	urlObj, err := url.Parse(c.URL)
	if err != nil {
		glog.Errorf("Error parsing URI: %v", err)
	}

	// No need to cache if its a local file
	if urlObj.Scheme == "file" {
		return os.Open(c.URL)
	}

	data, err := c.getFromURL()
	if err != nil {
		return nil, err
	}

	if !c.isSha256ValidFromURL(&data) {
		glog.Warningf("Warning: Unable to verify checksum of file %s", c.FilePath)
	}

	err = d.SetFile(c, data)
	if err != nil {
		return nil, err
	}

	return os.Open(c.FilePath)
}

func (d *DiskCache) SetFile(c CacheItem, b []byte) error {
	os.MkdirAll(path.Dir(c.FilePath), 0777)
	out, err := os.Create(c.FilePath)
	if err != nil {
		return errors.Wrapf(err, "Error caching file %s", c.FilePath)
	}
	defer out.Close()
	_, err = out.Write(b)
	if err != nil {
		return errors.Wrapf(err, "Error writing to cache %s", c.FilePath)
	}
	return nil
}

func (c *CacheItem) isSha256ValidFromURL(data *[]byte) bool {
	r, err := http.Get(c.ShaURL)
	if err != nil {
		glog.Infof("Error downloading checksum: %s", err)
		return false
	} else if r.StatusCode != http.StatusOK {
		glog.Infof("Error downloading checksum. Got HTTP Error: %s", r.Status)
		return false
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("Error reading checksum: %s", err)
		return false
	}

	expectedSum := strings.Trim(string(body), "\n")

	b := sha256.Sum256(*data)
	actualSum := hex.EncodeToString(b[:])
	if string(expectedSum) != actualSum {
		glog.Infof("Downloaded checksum does not match expected value. Actual: %s. Expected: %s", actualSum, expectedSum)
		return false
	}
	return true
}

func (c *CacheItem) getFromURL() ([]byte, error) {
	response, err := http.Get(c.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting file at %s via http", c.URL)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Error getting file at %s via http", c.URL)
	}

	return data, nil
}

func (c *CacheItem) isFileCached() bool {
	if _, err := os.Stat(c.FilePath); os.IsNotExist(err) {
		return false
	}

	return true
}
