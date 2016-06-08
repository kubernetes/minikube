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
	"fmt"
	"io/ioutil"
	"time"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/version"
)

const updateLinkPrefix = "https://github.com/kubernetes/minikube/releases/tag/v"

var timeLayout = time.RFC1123

var lastUpdateCheckFilePath = constants.Minipath + "/last_update_check"

func GetUpdateText() (string, error) {
	lastUpdateTime, err := getTimeFromFile(lastUpdateCheckFilePath)
	if err != nil {
		//This means the file doesn't exist, so don't do the time comparison
		glog.Infof("There was an error with getTimeFromFile (likely the file doesn't exist at %s'): %s",
			lastUpdateCheckFilePath, err)
	} else {
		if time.Since(lastUpdateTime).Hours() < 24 { //use constant instead of 24???
			return "", nil
		}
	}
	//The [1:] slice is used to remove the 'v' prefix from the version as it messes with semver
	localVersion, err := semver.Make(version.GetVersion()[1:])
	if err != nil {
		return "", err
	}
	latestVersionString, err := getLatestVersionFromGithub()
	if err != nil {
		return "", err
	}
	latestVersion, err := semver.Make(latestVersionString)
	if err != nil {
		return "", err
	}
	if localVersion.Compare(latestVersion) < 0 {
		writeTimeToFile(lastUpdateCheckFilePath, time.Now().UTC())
		return fmt.Sprintf("There is a newer version of minikube available (v%s).  Download it here:\n%s%s\n",
			latestVersion, updateLinkPrefix, latestVersion), nil
	}
	return "", nil
}

func getLatestVersionFromGithub() (string, error) {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases("kubernetes", "minikube", nil)
	if err != nil {
		fmt.Errorf("Repositories.ListReleases returned error: %v", err)
	}
	latestVersionString := *releases[0].Name
	//The [1:] slice is used to remove the 'v' prefix from the version as it messes with semver
	return latestVersionString[1:], nil
}

func writeTimeToFile(path string, inputTime time.Time) error {
	err := ioutil.WriteFile(path, []byte(inputTime.Format(timeLayout)), 0644)
	if err != nil {
		return fmt.Errorf("Error writing current update time to file: ", err)
	}
	return nil
}

func getTimeFromFile(path string) (time.Time, error) {
	lastUpdateCheckTime, err := ioutil.ReadFile(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("Error getting current update time to file: ", err)
	}
	return time.Parse(timeLayout, string(lastUpdateCheckTime))
}
