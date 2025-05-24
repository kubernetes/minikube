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
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/detect"
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

// MaybePrintUpdateTextFromAliyunMirror prints update text if needed, from Aliyun mirror
func MaybePrintUpdateTextFromAliyunMirror() {
	maybePrintUpdateText(GithubMinikubeReleasesAliyunURL, GithubMinikubeBetaReleasesAliyunURL, lastUpdateCheckFilePath)
}

func maybePrintUpdateText(latestReleasesURL string, betaReleasesURL string, lastUpdatePath string) {
	latestVersion, err := latestVersionFromURL(latestReleasesURL)
	if err != nil {
		klog.Warning(err)
		return
	}
	if !shouldCheckURLVersion(lastUpdatePath) {
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

func printUpdateTextCommon(ver semver.Version) {
	if err := writeTimeToFile(lastUpdateCheckFilePath, time.Now().UTC()); err != nil {
		klog.Errorf("write time failed: %v", err)
	}
	url := "https://github.com/kubernetes/minikube/releases/tag/v" + ver.String()
	out.Styled(style.Celebrate, `minikube {{.version}} is available! Download it: {{.url}}`, out.V{"version": ver, "url": url})
}

func printUpdateText(ver semver.Version) {
	printUpdateTextCommon(ver)
	out.Styled(style.Tip, "To disable this notice, run: 'minikube config set WantUpdateNotification false'\n")
}

func printBetaUpdateText(ver semver.Version) {
	printUpdateTextCommon(ver)
	out.Styled(style.Tip, "To disable beta notices, run: 'minikube config set WantBetaUpdateNotification false'")
	out.Styled(style.Tip, "To disable update notices in general, run: 'minikube config set WantUpdateNotification false'\n")
}

func shouldCheckURLVersion(filePath string) bool {
	if !viper.GetBool(config.WantUpdateNotification) {
		return false
	}
	if !viper.GetBool("interactive") {
		return false
	}
	if out.JSON {
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

type operatingSystems struct {
	Darwin  string `json:"darwin,omitempty"`
	Linux   string `json:"linux,omitempty"`
	Windows string `json:"windows,omitempty"`
}

type checksums struct {
	AMD64   *operatingSystems `json:"amd64,omitempty"`
	ARM     *operatingSystems `json:"arm,omitempty"`
	ARM64   *operatingSystems `json:"arm64,omitempty"`
	PPC64LE *operatingSystems `json:"ppc64le,omitempty"`
	S390X   *operatingSystems `json:"s390x,omitempty"`
	operatingSystems
}

type Release struct {
	Checksums checksums `json:"checksums"`
	Name      string    `json:"name"`
}

type Releases struct {
	Releases []Release
}

func (r *Releases) UnmarshalJSON(p []byte) error {
	return json.Unmarshal(p, &r.Releases)
}

func getJSON(url string, target *Releases) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "error creating new http request")
	}
	ua := fmt.Sprintf("Minikube/%s Minikube-OS/%s Minikube-Arch/%s Minikube-Plaform/%s Minikube-Cloud/%s",
		version.GetVersion(), runtime.GOOS, runtime.GOARCH, platform(), cloud())

	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "error with http GET for endpoint %s", url)
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func platform() string {
	if detect.GithubActionRunner() {
		return "GitHub Action"
	}
	if detect.IsCloudShell() {
		return "Cloud Shell"
	}
	if detect.IsMicrosoftWSL() {
		return "WSL"
	}
	return "none"
}

func cloud() string {
	if detect.IsOnGCE() {
		return "GCE"
	}
	return "none"
}

var latestVersionFromURL = func(url string) (semver.Version, error) {
	r, err := AllVersionsFromURL(url)
	if err != nil {
		return semver.Version{}, err
	}
	return semver.Make(strings.TrimPrefix(r.Releases[0].Name, version.VersionPrefix))
}

// AllVersionsFromURL get all versions from a JSON URL
func AllVersionsFromURL(url string) (Releases, error) {
	var releases Releases
	klog.Info("Checking for updates...")
	if err := getJSON(url, &releases); err != nil {
		return releases, errors.Wrap(err, "Error getting json from minikube version url")
	}
	if len(releases.Releases) == 0 {
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
	lastUpdateCheckTime, err := os.ReadFile(path)
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
func DownloadURL(ver, osName, arch string) string {
	if ver == "" || strings.HasSuffix(ver, "-unset") || osName == "" || arch == "" {
		return "https://github.com/kubernetes/minikube/releases"
	}
	sfx := ""
	if osName == "windows" {
		sfx = ".exe"
	}
	return fmt.Sprintf("https://github.com/kubernetes/minikube/releases/download/%s/minikube-%s-%s%s",
		ver, osName, arch, sfx)
}
