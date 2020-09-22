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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang/glog"

	"github.com/google/go-github/v32/github"
)

func main() {
	vDefault := ""
	vNewest := ""

	client := github.NewClient(nil)

	// walk through the paginated list of all 'kubernetes/kubernetes' repo releases
	// from latest to older releases, until latest release and pre-release are found
	// use max value (100) for PerPage to avoid hitting the rate limits (60 per hour, 10 per minute)
	// see https://godoc.org/github.com/google/go-github/github#hdr-Rate_Limiting
	opt := &github.ListOptions{PerPage: 100}
out:
	for {
		rels, resp, err := client.Repositories.ListReleases(context.Background(), "kubernetes", "kubernetes", opt)
		if err != nil {
			glog.Errorf("GitHub ListReleases failed: %v", err)
			break
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
				break out
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if vDefault == "" || vNewest == "" {
		glog.Errorf("Cannot determine DefaultKubernetesVersion or NewestKubernetesVersion")
		os.Exit(1)
	}

	constantsFile := "../../pkg/minikube/constants/constants.go"
	cf, err := ioutil.ReadFile(constantsFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	info, err := os.Stat(constantsFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	mode := info.Mode()

	re := regexp.MustCompile(`DefaultKubernetesVersion = \".*`)
	f := re.ReplaceAllString(string(cf), "DefaultKubernetesVersion = \""+vDefault+"\"")

	re = regexp.MustCompile(`NewestKubernetesVersion = \".*`)
	f = re.ReplaceAllString(f, "NewestKubernetesVersion = \""+vNewest+"\"")

	if err := ioutil.WriteFile(constantsFile, []byte(f), mode); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// update testData just for the latest 'v<MAJOR>.<MINOR>' from vDefault
	vDefaultMM := vDefault[:strings.LastIndex(vDefault, ".")]
	testData := "../../pkg/minikube/bootstrapper/bsutil/testdata/" + vDefaultMM

	err = filepath.Walk(testData, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, "default.yaml") {
			return nil
		}
		cf, err = ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		re = regexp.MustCompile(`kubernetesVersion: .*`)
		cf = []byte(re.ReplaceAllString(string(cf), "kubernetesVersion: "+vDefaultMM+".0")) // TODO: let <PATCH> version to be the latest one instead of "0"
		return ioutil.WriteFile(path, cf, info.Mode())
	})
	if err != nil {
		glog.Errorf("Walk failed: %v", err)
	}
}
