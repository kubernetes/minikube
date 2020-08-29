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

package notify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/version"
)

func TestMaybePrintUpdateTextFromGithub(t *testing.T) {
	if MaybePrintUpdateTextFromGithub() {
		t.Fatal("MaybePrintUpdateTextFromGithub() expected to return false for basic setup, bot got true")
	}
}

func TestShouldCheckURL(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	lastUpdateCheckFilePath := filepath.Join(tempDir, "last_update_check")

	// test that if users disable update notification in config, the URL version does not get checked
	viper.Set(config.WantUpdateNotification, false)
	if shouldCheckURLVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLVersion returned true even though config had WantUpdateNotification: false")
	}

	// test that if users want update notification, the URL version does get checked
	viper.Set(config.WantUpdateNotification, true)
	if !shouldCheckURLVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLVersion returned false even though there was no last_update_check file")
	}

	// test that update notifications get triggered if it has been longer than 24 hours
	viper.Set(config.ReminderWaitPeriodInHours, 24)

	// time.Time{} returns time -> January 1, year 1, 00:00:00.000000000 UTC.
	if err := writeTimeToFile(lastUpdateCheckFilePath, time.Time{}); err != nil {
		t.Errorf("write failed: %v", err)
	}
	if !shouldCheckURLVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLVersion returned false even though longer than 24 hours since last update")
	}

	// test that update notifications do not get triggered if it has been less than 24 hours
	if err := writeTimeToFile(lastUpdateCheckFilePath, time.Now().UTC()); err != nil {
		t.Errorf("write failed: %v", err)
	}
	if shouldCheckURLVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLVersion returned true even though less than 24 hours since last update")
	}
}

type URLHandlerCorrect struct {
	releases Releases
}

func (h *URLHandlerCorrect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(h.releases)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	_, err = fmt.Fprint(w, string(b))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func TestGetLatestVersionFromURLCorrect(t *testing.T) {
	// test that the version is correctly parsed if returned if valid JSON is returned the url endpoint
	latestVersionFromURL := "0.0.0-dev"
	handler := &URLHandlerCorrect{
		releases: []Release{{Name: version.VersionPrefix + latestVersionFromURL}},
	}
	server := httptest.NewServer(handler)

	latestVersion, err := getLatestVersionFromURL(server.URL)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expectedVersion, _ := semver.Make(latestVersionFromURL)
	if latestVersion.Compare(expectedVersion) != 0 {
		t.Fatalf("Expected latest version from URL to be %s, it was instead %s", expectedVersion, latestVersion)
	}
}

type URLHandlerNone struct{}

func (h *URLHandlerNone) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestGetLatestVersionFromURLNone(t *testing.T) {
	// test that an error is returned if nothing is returned at the url endpoint
	handler := &URLHandlerNone{}
	server := httptest.NewServer(handler)

	_, err := getLatestVersionFromURL(server.URL)
	if err == nil {
		t.Fatalf("No version value was returned from URL but no error was thrown")
	}
}

type URLHandlerMalformed struct{}

func (h *URLHandlerMalformed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, "Malformed JSON")
}

func TestGetLatestVersionFromURLMalformed(t *testing.T) {
	// test that an error is returned if malformed JSON is at the url endpoint
	handler := &URLHandlerMalformed{}
	server := httptest.NewServer(handler)

	_, err := getLatestVersionFromURL(server.URL)
	if err == nil {
		t.Fatalf("Malformed version value was returned from URL but no error was thrown")
	}
}

func TestMaybePrintUpdateText(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)
	outputBuffer := tests.NewFakeFile()
	out.SetErrFile(outputBuffer)

	tc := []struct {
		len                     int
		wantUpdateNotification  bool
		latestVersionFromURL    string
		description             string
		status                  bool
		url                     string
		lastUpdateCheckFilePath string
	}{
		{
			len:                    1,
			latestVersionFromURL:   "0.0.0-dev",
			wantUpdateNotification: true,
			description:            "latest version lower or equal",
		},
		{
			len:                    0,
			latestVersionFromURL:   "100.0.0-dev",
			wantUpdateNotification: true,
			description:            "latest version greater",
			status:                 true,
		},
		{
			len:                    1,
			latestVersionFromURL:   "100.0.0-dev",
			wantUpdateNotification: false,
			description:            "notification unwanted",
		},
		{
			len:                    1,
			latestVersionFromURL:   "100.0.0-dev",
			wantUpdateNotification: true,
			description:            "bad url",
			url:                    "this is not valid url",
			status:                 true,
		},
		{
			len:                     1,
			latestVersionFromURL:    "10.0.0-dev",
			wantUpdateNotification:  true,
			description:             "bad lastUpdateCheckFilePath",
			lastUpdateCheckFilePath: "/etc/passwd",
			status:                  true,
		},
	}

	viper.Set(config.ReminderWaitPeriodInHours, 24)
	for _, test := range tc {
		t.Run(test.description, func(t *testing.T) {
			viper.Set(config.WantUpdateNotification, test.wantUpdateNotification)
			lastUpdateCheckFilePath = filepath.Join(tempDir, "last_update_check")
			if test.lastUpdateCheckFilePath != "" {
				lastUpdateCheckFilePath = test.lastUpdateCheckFilePath
			}
			latestVersionFromURL := test.latestVersionFromURL
			handler := &URLHandlerCorrect{
				releases: []Release{{Name: version.VersionPrefix + latestVersionFromURL}},
			}
			server := httptest.NewServer(handler)
			defer server.Close()
			if test.url == "" {
				test.url = server.URL
			}
			tmpfile, err := ioutil.TempFile("", "")
			if err != nil {
				t.Fatalf("Cannot create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())
			status := MaybePrintUpdateText(test.url, tmpfile.Name())
			if test.status != status {
				t.Fatalf("MaybePrintUpdateText expected to return %v, but got %v", test.status, status)
			}
			if len(outputBuffer.String()) == test.len {
				t.Fatalf("Expected MaybePrintUpdateText to output text as the current version is %s and version %s was served from URL but output was [%s]",
					version.GetVersion(), latestVersionFromURL, outputBuffer.String())
			}
		})
	}
}
