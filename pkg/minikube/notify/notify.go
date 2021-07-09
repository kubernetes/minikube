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
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util/lock"
	"k8s.io/minikube/pkg/version"
)

var (
	timeLayout              = time.RFC1123
	lastUpdateCheckFilePath = localpath.MakeMiniPath("last_update_check")
)

// MaybePrintUpdateTextFromGithub prints update text if needed, from github
func MaybePrintUpdateTextFromGithub() {
	maybePrintUpdateText(GithubMinikubeReleasesURL, GithubMinikubeBetaReleasesURL, lastUpdateCheckFilePath)
}

func maybePrintUpdateText(latestReleasesURL string, betaReleasesURL string, lastUpdatePath string) {
	if !shouldCheckURLVersion(lastUpdatePath) {
		return
	}
	latestVersion, err := latestVersionFromURL(latestReleasesURL)
	if err != nil {
		klog.Warning(err)
		return
	}
	localVersion, err := version.GetSemverVersion()
	if err != nil {
		klog.Warning(err)
		return
	}
	if maybePrintBetaUpdateText(betaReleasesURL, localVersion, latestVersion, lastUpdatePath) {
		return
	}
	if localVersion.Compare(latestVersion) >= 0 {
		return
	}
	printUpdateText(latestVersion)
}

// maybePrintBetaUpdateText returns true if update text is printed
func maybePrintBetaUpdateText(betaReleasesURL string, localVersion semver.Version, latestFullVersion semver.Version, lastUpdatePath string) bool {
	if !shouldCheckURLBetaVersion(lastUpdatePath) {
		return false
	}
	latestBetaVersion, err := latestVersionFromURL(betaReleasesURL)
	if err != nil {
		klog.Warning(err)
		return false
	}
	if latestFullVersion.Compare(latestBetaVersion) >= 0 {
		return false
	}
	if localVersion.Compare(latestBetaVersion) >= 0 {
		return false
	}
	printBetaUpdateText(latestBetaVersion)
	return true
}

func printUpdateTextCommon(version semver.Version) {
	if err := writeTimeToFile(lastUpdateCheckFilePath, time.Now().UTC()); err != nil {
		klog.Errorf("write time failed: %v", err)
	}
	url := "https://github.com/kubernetes/minikube/releases/tag/v" + version.String()
	out.Styled(style.Celebrate, `minikube {{.version}} is available! Download it: {{.url}}`, out.V{"version": version, "url": url})
}

func printUpdateText(version semver.Version) {
	printUpdateTextCommon(version)
	out.Styled(style.Tip, "To disable this notice, run: 'minikube config set WantUpdateNotification false'\n")
}

func printBetaUpdateText(version semver.Version) {
	printUpdateTextCommon(version)
	out.Styled(style.Tip, "To disable beta notices, run: 'minikube config set WantBetaUpdateNotification false'")
	out.Styled(style.Tip, "To disable update notices in general, run: 'minikube config set WantUpdateNotification false'\n")
}

func shouldCheckURLVersion(filePath string) bool {
	if !viper.GetBool(config.WantUpdateNotification) {
		return false
	}
	lastUpdateTime := timeFromFileIfExists(filePath)
	return time.Since(lastUpdateTime).Hours() >= viper.GetFloat64(config.ReminderWaitPeriodInHours)
}

func shouldCheckURLBetaVersion(filePath string) bool {
	if !viper.GetBool(config.WantBetaUpdateNotification) {
		return false
	}

	return shouldCheckURLVersion(filePath)
}

// Release represents a release
type Release struct {
	Name      string
	Checksums map[string]string
}

// Releases represents several release
type Releases []Release

func getJSON(url string, target *Releases) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "error creating new http request")
	}
	ua := fmt.Sprintf("Minikube/%s Minikube-OS/%s",
		version.GetVersion(), runtime.GOOS)

	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "error with http GET for endpoint %s", url)
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

var latestVersionFromURL = func(url string) (semver.Version, error) {
	r, err := AllVersionsFromURL(url)
	if err != nil {
		return semver.Version{}, err
	}
	return semver.Make(strings.TrimPrefix(r[0].Name, version.VersionPrefix))
}

// AllVersionsFromURL get all versions from a JSON URL
func AllVersionsFromURL(url string) (Releases, error) {
	var releases Releases
	klog.Info("Checking for updates...")
	if err := getJSON(url, &releases); err != nil {
		return releases, errors.Wrap(err, "Error getting json from minikube version url")
	}
	if len(releases) == 0 {
		return releases, errors.Errorf("There were no json releases at the url specified: %s", url)
	}
	return releases, nil
}

func writeTimeToFile(path string, inputTime time.Time) error {
	err := lock.WriteFile(path, []byte(inputTime.Format(timeLayout)), 0o644)
	if err != nil {
		return errors.Wrap(err, "Error writing current update time to file: ")
	}
	return nil
}

func timeFromFileIfExists(path string) time.Time {
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

// DownloadURL returns a URL to get minikube binary version ver for platform os/arch
func DownloadURL(ver, os, arch string) string {
	if ver == "" || strings.HasSuffix(ver, "-unset") || os == "" || arch == "" {
		return "https://github.com/kubernetes/minikube/releases"
	}
	sfx := ""
	if os == "windows" {
		sfx = ".exe"
	}
	return fmt.Sprintf("https://github.com/kubernetes/minikube/releases/download/%s/minikube-%s-%s%s",
		ver, os, arch, sfx)
}
