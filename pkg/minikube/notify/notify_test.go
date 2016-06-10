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
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/version"
)

func TestShouldCheckURL(t *testing.T) {
	viper.Set(config.WantUpdateNotification, false)
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)
	lastUpdateCheckFilePath := filepath.Join(tempDir, "last_update_check")

	if shouldCheckURLVersion(lastUpdateCheckFilePath) {
		t.Fatalf("Error: shouldCheckURLVersion returned true even though config had WantUpdateNotification: false")
	}
	viper.Set(config.WantUpdateNotification, true)

	if shouldCheckURLVersion(lastUpdateCheckFilePath) == false {
		t.Fatalf("Error: shouldCheckURLVersion returned false even though there was no last_update_check file")
	}
	viper.Set(config.ReminderWaitPeriodInHours, 24)

	writeTimeToFile(lastUpdateCheckFilePath, time.Time{})
	if shouldCheckURLVersion(lastUpdateCheckFilePath) == false {
		t.Fatalf("Error: shouldCheckURLVersion returned false even though longer than 24 hours since last update")
	}

}

type URLHandlerCorrect struct {
	releases releases
}

func (h *URLHandlerCorrect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(h.releases)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, string(b))
}

func TestGetLatestVersionFromURLCorrect(t *testing.T) {
	latestVersionFromURL := "0.0.0-dev"
	handler := &URLHandlerCorrect{
		releases: []release{{Name: version.VersionPrefix + latestVersionFromURL}},
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
	handler := &URLHandlerMalformed{}
	server := httptest.NewServer(handler)

	_, err := getLatestVersionFromURL(server.URL)
	if err == nil {
		t.Fatalf("Error: ")
	}
}

func TestMaybePrintUpdateText(t *testing.T) {
	tempDir := tests.MakeTempDir()
	defer os.RemoveAll(tempDir)

	viper.Set(config.WantUpdateNotification, true)
	viper.Set(config.ReminderWaitPeriodInHours, 24)

	outputFilePath := filepath.Join(tempDir, "maybePrintUpdateTestFileShouldOutput")
	outputFile, _ := os.Create(outputFilePath) //Is there a better way to make a mock file?
	//also is it okay to ignore the errors for things you're not testing to avoid repeated error checks?
	lastUpdateCheckFilePath := filepath.Join(tempDir, "last_update_check")

	latestVersionFromURL := "100.0.0-dev"
	handler := &URLHandlerCorrect{
		releases: []release{{Name: version.VersionPrefix + latestVersionFromURL}},
	}
	server := httptest.NewServer(handler)

	MaybePrintUpdateText(outputFile, server.URL, lastUpdateCheckFilePath)
	outputByteArr, _ := ioutil.ReadFile(outputFilePath)
	outputString := string(outputByteArr)
	fmt.Println(outputString)
	if len(outputString) == 0 {
		t.Fatalf("Expected MaybePrintUpdateText to output text as the current version is %s and version %s was served from URL but output was [%s]",
			version.GetVersion(), latestVersionFromURL, outputString)
	}

	outputFilePath = filepath.Join(tempDir, "maybePrintUpdateTestFileShouldNotOutput")
	outputFile, _ = os.Create(outputFilePath) //Is there a better way to make a mock file?

	latestVersionFromURL = "0.0.0-dev"
	handler = &URLHandlerCorrect{
		releases: []release{{Name: version.VersionPrefix + latestVersionFromURL}},
	}
	server = httptest.NewServer(handler)

	MaybePrintUpdateText(outputFile, server.URL, lastUpdateCheckFilePath)
	outputByteArr, _ = ioutil.ReadFile(outputFilePath)
	outputString = string(outputByteArr)
	fmt.Println(outputString)
	if len(outputString) != 0 {
		t.Fatalf("Expected MaybePrintUpdateText to not output text as the current version is %s and version %s was served from URL but output was [%s]",
			version.GetVersion(), latestVersionFromURL, outputString)
	}
}
