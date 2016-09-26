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
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

type httpHandler struct{}

// The test HTTP handler returns the request url path
// or the sha256 of the request url path if .sha256 is the suffix
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".sha256") {
		spl := strings.Split(r.URL.Path, ".")
		b := sha256.Sum256([]byte(spl[0]))
		actualSum := hex.EncodeToString(b[:])
		w.Write([]byte(actualSum))
	} else {
		w.Write([]byte(r.URL.Path))
	}
}

func setupFakeServer() *httptest.Server {
	handler := &httpHandler{}
	return httptest.NewServer(handler)
}

func TestChecksumValidation(t *testing.T) {
	server := setupFakeServer()
	defer server.Close()

	var checksumTests = []struct {
		c       CacheItem
		data    []byte
		isValid bool
	}{
		{
			CacheItem{
				ShaURL: server.URL + "/test1.sha256",
			},
			[]byte("/test1"),
			true,
		},
		{
			CacheItem{
				URL:    server.URL + "/test2",
				ShaURL: server.URL + "/incorrect_sha.sha256",
			},
			[]byte("/test2"),
			false,
		},
	}
	for _, test := range checksumTests {
		valid := test.c.isSha256ValidFromURL(&test.data)
		if valid && !test.isValid {
			t.Errorf("Checksum validation returned true when actually false %v", test)

		}
		if !valid && test.isValid {
			t.Errorf("Checksum validation failed even though was valid %v", test)
		}
	}
}

func TestGetCachedFile(t *testing.T) {
	s := setupFakeServer()
	defer s.Close()
	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)

	d := DiskCache{}

	cacheTests := []CacheItem{
		{
			FilePath: path.Join(dir, "testfile"),
			URL:      s.URL + "/testfile",
			ShaURL:   s.URL + "/testfile.sha256",
		},
		{
			FilePath: path.Join(dir, "testfile2"),
			URL:      s.URL + "/testfile2",
			ShaURL:   "incorrect sha but should still cache",
		},
	}
	for _, c := range cacheTests {
		f, err := d.GetFile(c)
		defer f.Close()
		if err != nil {
			t.Errorf("Error caching file, %v", err)
		}
		var b []byte
		f.Read(b)
		data := strings.Split(c.URL, "/")
		if data[1] != string(b) {
			t.Errorf("Contents of cached file were incorrect, expected %s, got %s", data[1], string(b))
		}
	}
}
