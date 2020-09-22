/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"

	"github.com/google/go-github/v32/github"
)

func main() {
	// init glog: by default, all log statements write to files in a temporary directory, also
	// flag.Parse must be called before any logging is done
	flag.Parse()
	_ = flag.Set("logtostderr", "true")

	// fetch respective current stable (vDefault as DefaultKubernetesVersion) and
	// latest rc or beta (vDefault as NewestKubernetesVersion) Kubernetes GitHub Releases
	vDefault, vNewest, err := fetchKubernetesReleases()
	if err != nil {
		glog.Errorf("Fetching current GitHub Releases failed: %v", err)
	}
	if vDefault == "" || vNewest == "" {
		glog.Fatalf("Cannot determine current 'DefaultKubernetesVersion' and 'NewestKubernetesVersion'")
	}
	glog.Infof("Current Kubernetes GitHub Releases: 'stable' is %s and 'latest' is %s", vDefault, vNewest)

	if err := updateKubernetesVersions(vDefault, vNewest); err != nil {
		glog.Fatalf("Updating 'DefaultKubernetesVersion' and 'NewestKubernetesVersion' failed: %v", err)
	}
	glog.Infof("Update successful: 'DefaultKubernetesVersion' was set to %s and 'NewestKubernetesVersion' was set to %s", vDefault, vNewest)

	// Flush before exiting to guarantee all log output is written
	glog.Flush()
}

// fetchKubernetesReleases returns respective current stable (as vDefault) and
// latest rc or beta (as vNewest) Kubernetes GitHub Releases, and any error
func fetchKubernetesReleases() (vDefault, vNewest string, err error) {
	client := github.NewClient(nil)

	// set a context with a deadline - timeout after at most 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// walk through the paginated list of all 'kubernetes/kubernetes' repo releases
	// from latest to older releases, until latest release and pre-release are found
	// use max value (100) for PerPage to avoid hitting the rate limits (60 per hour, 10 per minute)
	// see https://godoc.org/github.com/google/go-github/github#hdr-Rate_Limiting
	opt := &github.ListOptions{PerPage: 100}
	for {
		rels, resp, err := client.Repositories.ListReleases(ctx, "kubernetes", "kubernetes", opt)
		if err != nil {
			return "", "", err
		}

		for _, r := range rels {
			// GetName returns the Name field if it's non-nil, zero value otherwise.
			ver := r.GetName()
			if ver == "" {
				continue
			}

			rel := strings.Split(ver, "-")
			// check if it is a release channel (ie, 'v1.19.2') or a
			// pre-release channel (ie, 'v1.19.3-rc.0' or 'v1.19.0-beta.2')
			if len(rel) == 1 && vDefault == "" {
				vDefault = ver
			} else if len(rel) > 1 && vNewest == "" {
				if strings.HasPrefix(rel[1], "rc") || strings.HasPrefix(rel[1], "beta") {
					vNewest = ver
				}
			}

			if vDefault != "" && vNewest != "" {
				// make sure that vNewest >= vDefault
				if vNewest < vDefault {
					vNewest = vDefault
				}
				return vDefault, vNewest, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return vDefault, vNewest, nil
}

// updateKubernetesVersions updates DefaultKubernetesVersion to vDefault release and
// NewestKubernetesVersion to vNewest release, and returns any error
func updateKubernetesVersions(vDefault, vNewest string) error {
	if err := replaceAllString("../../pkg/minikube/constants/constants.go", map[string]string{
		`DefaultKubernetesVersion = \".*`: "DefaultKubernetesVersion = \"" + vDefault + "\"",
		`NewestKubernetesVersion = \".*`:  "NewestKubernetesVersion = \"" + vNewest + "\"",
	}); err != nil {
		return err
	}

	if err := replaceAllString("../../site/content/en/docs/commands/start.md", map[string]string{
		`'stable' for .*,`:  "'stable' for " + vDefault + ",",
		`'latest' for .*\)`: "'latest' for " + vNewest + ")",
	}); err != nil {
		return err
	}

	// update testData just for the latest 'v<MAJOR>.<MINOR>.0' from vDefault
	vDefaultMM := vDefault[:strings.LastIndex(vDefault, ".")]
	testData := "../../pkg/minikube/bootstrapper/bsutil/testdata/" + vDefaultMM

	return filepath.Walk(testData, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, "default.yaml") {
			return nil
		}
		return replaceAllString(path, map[string]string{
			`kubernetesVersion: .*`: "kubernetesVersion: " + vDefaultMM + ".0",
		})
	})
}

// replaceAllString replaces all occuranes of map's keys with their respective values in the file
func replaceAllString(path string, pairs map[string]string) error {
	fb, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := info.Mode()

	f := string(fb)
	for org, new := range pairs {
		re := regexp.MustCompile(org)
		f = re.ReplaceAllString(f, new)
	}
	if err := ioutil.WriteFile(path, []byte(f), mode); err != nil {
		return err
	}

	return nil
}
