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

package update

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/blang/semver"
	"github.com/bugsnag/osext"
	"github.com/golang/glog"
	update "github.com/inconshreveable/go-update"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/version"
)

const downloadLinkFormat = "https://github.com/kubernetes/minikube/releases/download/v%s/%s"

var (
	timeLayout                = time.RFC1123
	lastUpdateCheckFilePath   = constants.MakeMiniPath("last_update_check")
	githubMinikubeReleasesURL = "https://storage.googleapis.com/minikube/releases.json"
	downloadBinary            = "minikube-" + runtime.GOOS + "-" + runtime.GOARCH
)

func MaybeUpdateFromGithub(output io.Writer) {
	MaybeUpdate(output, githubMinikubeReleasesURL, lastUpdateCheckFilePath)
}

func MaybeUpdate(output io.Writer, url string, lastUpdatePath string) {
	if !shouldCheckURLVersion(lastUpdatePath) {
		return
	}
	latestVersion, err := getLatestVersionFromURL(url)
	if err != nil {
		glog.Errorln(err)
		return
	}
	localVersion, err := version.GetSemverVersion()
	if err != nil {
		glog.Errorln(err)
		return
	}
	if localVersion.Compare(latestVersion) < 0 {
		writeTimeToFile(lastUpdateCheckFilePath, time.Now().UTC())
		fmt.Fprintf(output,
			`There is a newer version of minikube available (%s%s). Do you want to automatically update? [y/N] `,
			version.VersionPrefix, latestVersion)

		var confirm string
		fmt.Scanln(&confirm)

		if strings.ToLower(confirm) == "y" {
			fmt.Printf("Updating to version %s\n", latestVersion)
			updateBinary(latestVersion)
			return
		}

		fmt.Println("Skipping autoupdate")
	}
}

func shouldCheckURLVersion(filePath string) bool {
	if !viper.GetBool(config.WantUpdateNotification) {
		return false
	}
	lastUpdateTime := getTimeFromFileIfExists(filePath)
	if time.Since(lastUpdateTime).Hours() < viper.GetFloat64(config.ReminderWaitPeriodInHours) {
		return false
	}
	return true
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

func getLatestVersionFromURL(url string) (semver.Version, error) {
	var releases releases
	if err := getJson(url, &releases); err != nil {
		return semver.Version{}, err
	}
	if len(releases) == 0 {
		return semver.Version{}, fmt.Errorf("There were no json releases at the url specified: %s", url)
	}
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

func updateBinary(v semver.Version) {
	checksum, err := downloadChecksum(v)
	if err != nil {
		glog.Errorf("Cannot download checksum: %s", err)
		os.Exit(1)
	}
	binary, err := http.Get(fmt.Sprintf(downloadLinkFormat, v, downloadBinary))
	if err != nil {
		glog.Errorf("Cannot download binary: %s", err)
		os.Exit(1)
	}
	defer binary.Body.Close()
	err = update.Apply(binary.Body, update.Options{
		Hash:     crypto.SHA256,
		Checksum: checksum,
	})
	if err != nil {
		glog.Errorf("Cannot apply binary update: %s", err)
		os.Exit(1)
	}

	env := os.Environ()
	args := os.Args
	currentBinary, err := osext.Executable()
	if err != nil {
		glog.Errorf("Cannot find current binary to exec: %s", err)
		os.Exit(1)
	}
	err = syscall.Exec(currentBinary, args, env)
	if err != nil {
		glog.Errorf("Failed to exec updated binary: %s", err)
		os.Exit(1)
	}
}

func downloadChecksum(v semver.Version) ([]byte, error) {
	u := fmt.Sprintf(downloadLinkFormat, v, downloadBinary+".sha256")
	checksumResp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer checksumResp.Body.Close()

	// If no checksum then return nil slice with no error - nothing will be checked
	if checksumResp.StatusCode != 404 {
		return nil, nil
	}

	if checksumResp.StatusCode != 200 {
		return nil, fmt.Errorf("received %d", checksumResp.StatusCode)
	}
	b, err := ioutil.ReadAll(checksumResp.Body)
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(strings.TrimSpace(string(b)))
}
