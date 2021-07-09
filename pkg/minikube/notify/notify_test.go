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
	"strings"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/version"
)

func TestShouldCheckURLVersion(t *testing.T) {
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

func TestShouldCheckURLBetaVersion(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	lastUpdateCheckFilePath := filepath.Join(tempDir, "last_update_check")
	viper.Set(config.WantUpdateNotification, true)

	// test if the user disables beta update notification in config, the URL version does not get checked
	viper.Set(config.WantBetaUpdateNotification, false)
	if shouldCheckURLBetaVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLBetaVersion returned true even though config had WantBetaUpdateNotification: false")
	}

	// test if the user enables beta update notification in config, the URL version does get checked
	viper.Set(config.WantBetaUpdateNotification, true)
	if !shouldCheckURLBetaVersion(lastUpdateCheckFilePath) {
		t.Fatalf("shouldCheckURLBetaVersion returned false even though config had WantBetaUpdateNotification: true")
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

func TestLatestVersionFromURLCorrect(t *testing.T) {
	// test that the version is correctly parsed if returned if valid JSON is returned the url endpoint
	versionFromURL := "0.0.0-dev"
	handler := &URLHandlerCorrect{
		releases: []Release{{Name: version.VersionPrefix + versionFromURL}},
	}
	server := httptest.NewServer(handler)

	latestVersion, err := latestVersionFromURL(server.URL)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expectedVersion, _ := semver.Make(versionFromURL)
	if latestVersion.Compare(expectedVersion) != 0 {
		t.Fatalf("Expected latest version from URL to be %s, it was instead %s", expectedVersion, latestVersion)
	}
}

type URLHandlerNone struct{}

func (h *URLHandlerNone) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestLatestVersionFromURLNone(t *testing.T) {
	// test that an error is returned if nothing is returned at the url endpoint
	handler := &URLHandlerNone{}
	server := httptest.NewServer(handler)

	_, err := latestVersionFromURL(server.URL)
	if err == nil {
		t.Fatalf("No version value was returned from URL but no error was thrown")
	}
}

type URLHandlerMalformed struct{}

func (h *URLHandlerMalformed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, "Malformed JSON")
}

func TestLatestVersionFromURLMalformed(t *testing.T) {
	// test that an error is returned if malformed JSON is at the url endpoint
	handler := &URLHandlerMalformed{}
	server := httptest.NewServer(handler)

	_, err := latestVersionFromURL(server.URL)
	if err == nil {
		t.Fatalf("Malformed version value was returned from URL but no error was thrown")
	}
}

var mockLatestVersionFromURL = semver.Make

func TestMaybePrintUpdateText(t *testing.T) {
	latestVersionFromURL = mockLatestVersionFromURL

	tempDir := tests.MakeTempDir()
	defer tests.RemoveTempDir(tempDir)

	var tc = []struct {
		wantUpdateNotification     bool
		wantBetaUpdateNotification bool
		latestFullVersionFromURL   string
		latestBetaVersionFromURL   string
		description                string
		want                       string
	}{
		{
			wantUpdateNotification:     true,
			wantBetaUpdateNotification: true,
			latestFullVersionFromURL:   "99.0.0",
			latestBetaVersionFromURL:   "99.0.0-beta.0",
			description:                "latest full version greater",
			want:                       "99.0.0 ",
		},
		{
			wantUpdateNotification:     true,
			wantBetaUpdateNotification: true,
			latestFullVersionFromURL:   "97.0.0",
			latestBetaVersionFromURL:   "98.0.0-beta.0",
			description:                "latest beta version greater",
			want:                       "98.0.0-beta.0",
		},
		{
			wantUpdateNotification:     false,
			wantBetaUpdateNotification: true,
			latestFullVersionFromURL:   "97.0.0",
			latestBetaVersionFromURL:   "96.0.0-beta.0",
			description:                "notification unwanted",
		},
		{
			wantUpdateNotification:     true,
			wantBetaUpdateNotification: false,
			latestFullVersionFromURL:   "0.0.0-unset",
			latestBetaVersionFromURL:   "95.0.0-beta.0",
			description:                "beta notification unwanted",
		},
	}

	viper.Set(config.ReminderWaitPeriodInHours, 24)
	for _, tt := range tc {
		t.Run(tt.description, func(t *testing.T) {
			outputBuffer := tests.NewFakeFile()
			out.SetOutFile(outputBuffer)

			viper.Set(config.WantUpdateNotification, tt.wantUpdateNotification)
			viper.Set(config.WantBetaUpdateNotification, tt.wantBetaUpdateNotification)
			lastUpdateCheckFilePath = filepath.Join(tempDir, "last_update_check")

			tmpfile, err := ioutil.TempFile("", "")
			if err != nil {
				t.Fatalf("Cannot create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			maybePrintUpdateText(tt.latestFullVersionFromURL, tt.latestBetaVersionFromURL, tmpfile.Name())
			got := outputBuffer.String()
			if (tt.want == "" && len(got) != 0) || (tt.want != "" && !strings.Contains(got, tt.want)) {
				t.Fatalf("Expected MaybePrintUpdateText to contain the text %q as the current version is %s and full version %s and beta version %s, but output was [%s]",
					tt.want, version.GetVersion(), tt.latestFullVersionFromURL, tt.latestBetaVersionFromURL, outputBuffer.String())
			}
		})
	}
}

func TestDownloadURL(t *testing.T) {
	const urlBase = "https://github.com/kubernetes/minikube/releases/download/"
	type args struct {
		ver  string
		os   string
		arch string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"linux-amd64", args{"foo", "linux", "amd64"}, urlBase + "foo/minikube-linux-amd64"},
		{"linux-arm64", args{"foo", "linux", "arm64"}, urlBase + "foo/minikube-linux-arm64"},
		{"darwin-amd64", args{"foo", "darwin", "amd64"}, urlBase + "foo/minikube-darwin-amd64"},
		{"darwin-arm64", args{"foo", "darwin", "arm64"}, urlBase + "foo/minikube-darwin-arm64"},
		{"windows", args{"foo", "windows", "amd64"}, urlBase + "foo/minikube-windows-amd64.exe"},
		{"linux-unset", args{"foo-unset", "linux", "amd64"}, "https://github.com/kubernetes/minikube/releases"},
		{"linux-unset", args{"foo-unset", "windows", "arm64"}, "https://github.com/kubernetes/minikube/releases"},
		{"windows-zzz", args{"bar", "windows", "zzz"}, urlBase + "bar/minikube-windows-zzz.exe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DownloadURL(tt.args.ver, tt.args.os, tt.args.arch); got != tt.want {
				t.Errorf("DownloadURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
