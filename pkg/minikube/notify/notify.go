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
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/golang/glog"
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

func GetUpdateText() string {
	latestUpdateTime := getTimeFromFileIfExists(lastUpdateCheckFilePath)
	localVersion, err := version.GetSemverVersion()
	if err != nil {
		glog.Errorln(err)
		return ""
	}
	latestVersion, err := getLatestVersionFromGithub()
	if err != nil {
		glog.Errorln(err)
		return ""
	}
	return checkUpdateTimeAndVersions(latestUpdateTime, localVersion, latestVersion)
}

func checkUpdateTimeAndVersions(lastUpdateTime time.Time, localVersion semver.Version, latestVersion semver.Version) string {
	if time.Since(lastUpdateTime).Hours() < viper.GetFloat64(config.ReminderWaitPeriodInHours) {
		return ""
	}
	if localVersion.Compare(latestVersion) < 0 {
		writeTimeToFile(lastUpdateCheckFilePath, time.Now().UTC())
		return fmt.Sprintf("There is a newer version of minikube available (v%s).  Download it here:\n%s%s\n",
			latestVersion, updateLinkPrefix, latestVersion)
	}
	return ""
}

type release struct {
	Name string
}

type releases []release

func getJson(url string, target *releases) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getLatestVersionFromGithub() (semver.Version, error) {
	var releases releases
	getJson("https://api.github.com/repos/kubernetes/minikube/releases", &releases)
	latestVersionString := releases[0].Name
	return semver.Make(strings.TrimPrefix(latestVersionString, version.VersionPrefix))
}

func writeTimeToFile(path string, inputTime time.Time) error {
	err := ioutil.WriteFile(path, []byte(inputTime.Format(timeLayout)), 0644)
	if err != nil {
		return fmt.Errorf("Error writing current update time to file: ", err)
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
