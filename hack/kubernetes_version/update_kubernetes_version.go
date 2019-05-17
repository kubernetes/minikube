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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang/glog"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: go run update_kubernetes_version.go <kubernetes_version>")
		os.Exit(1)
	}

	v := os.Args[1]
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
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

	re := regexp.MustCompile(`var DefaultKubernetesVersion = .*`)
	f := re.ReplaceAllString(string(cf), "var DefaultKubernetesVersion = \""+v+"\"")

	re = regexp.MustCompile(`var NewestKubernetesVersion = .*`)
	f = re.ReplaceAllString(f, "var NewestKubernetesVersion = \""+v+"\"")

	if err := ioutil.WriteFile(constantsFile, []byte(f), mode); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testData := "../../pkg/minikube/bootstrapper/kubeadm/testdata"

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
		cf = []byte(re.ReplaceAllString(string(cf), "kubernetesVersion: "+v))
		return ioutil.WriteFile(path, cf, info.Mode())
	})
	if err != nil {
		glog.Errorf("Walk failed: %v", err)
	}
}
