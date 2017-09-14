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
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/version"
)

const updateLinkPrefix = "https://github.com/kubernetes/minikube/releases/tag/v"

var (
	timeLayout              = time.RFC1123
	lastUpdateCheckFilePath = constants.MakeMiniPath("last_update_check")
)

func MaybePrintUpdateTextFromGithub(output io.Writer) {
	MaybePrintUpdateText(output, constants.GithubMinikubeReleasesURL, lastUpdateCheckFilePath)
}

func MaybePrintUpdateText(output io.Writer, url string, lastUpdatePath string) {
	if !shouldCheckURLVersion(lastUpdatePath) {
		return
	}
	latestVersion, err := getLatestVersionFromURL(url)
	if err != nil {
		glog.Warning(err)
		return
	}
	localVersion, err := version.GetSemverVersion()
	if err != nil {
		glog.Warning(err)
		return
	}
	if localVersion.Compare(latestVersion) < 0 {
		writeTimeToFile(lastUpdateCheckFilePath, time.Now().UTC())
		fmt.Fprintf(output, `There is a newer version of minikube available (%s%s).  Download it here:
%s%s

To disable this notification, run the following:
minikube config set WantUpdateNotification false
`,
			version.VersionPrefix, latestVersion, updateLinkPrefix, latestVersion)
	}
}

func shouldCheckURLVersion(filePath string) bool {
	if !viper.GetBool(config.WantUpdateNotification) {
		return false
	}
	lastUpdateTime := getTimeFromFileIfExists(filePath)
	return time.Since(lastUpdateTime).Hours() >= viper.GetFloat64(config.ReminderWaitPeriodInHours)
}

type Release struct {
	Name      string
	Checksums map[string]string
}

type Releases []Release

func getJson(url string, target *Releases) error {
	r, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "Error getting minikube version url via http")
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getLatestVersionFromURL(url string) (semver.Version, error) {
	r, err := GetAllVersionsFromURL(url)
	if err != nil {
		return semver.Version{}, err
	}
	return semver.Make(strings.TrimPrefix(r[0].Name, version.VersionPrefix))
}

func GetAllVersionsFromURL(url string) (Releases, error) {
	var releases Releases
	glog.Infof("Checking for updates...")
	if err := getJson(url, &releases); err != nil {
		return releases, errors.Wrap(err, "Error getting json from minikube version url")
	}
	if len(releases) == 0 {
		return releases, errors.Errorf("There were no json releases at the url specified: %s", url)
	}
	return releases, nil
}

func writeTimeToFile(path string, inputTime time.Time) error {
	err := ioutil.WriteFile(path, []byte(inputTime.Format(timeLayout)), 0644)
	if err != nil {
		return errors.Wrap(err, "Error writing current update time to file: ")
	}
	return nil
}

func getTimeFromFileIfExists(path string) time.Time {
	lastUpdateCheckTime, err := ioutil.ReadFile(path)
	if err != nil {
		return time.Time{}
	}
	timeInFile, err := time.Parse(timeLayout, string(lastUpdateCheckTime))
	if err != nil {
		return time.Time{}
	}
	return timeInFile
}
