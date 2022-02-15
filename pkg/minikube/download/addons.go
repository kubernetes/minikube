/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package download

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/version"
)

func addonStatusURL() string {
	return fmt.Sprintf("https://%s/minikube-builds/addons/%s/status.yaml", downloadHost, version.GetVersion())
}

func addonListURL() string {
	return fmt.Sprintf("https://%s/minikube-builds/addons/%s/addons.yaml", downloadHost, version.GetVersion())
}

func AddonStatus() (map[string]util.AddonStatus, error) {
	statuses := make(map[string]util.AddonStatus)
	fileLocation := filepath.Join(localpath.MiniPath(), "addons", "status.yaml")
	err := downloadNoProgressBar(addonStatusURL(), fileLocation)
	if err != nil {
		return statuses, err
	}
	statusFile, err := os.ReadFile(fileLocation)
	if err != nil {
		return statuses, err
	}

	err = yaml.Unmarshal(statusFile, statuses)
	if err != nil {
		return statuses, err
	}

	return statuses, nil
}

func AddonList() (map[string]interface{}, error) {
	fileLocation := filepath.Join(localpath.MiniPath(), "addons", "addons.yaml")
	addonMap := make(map[string]interface{})

	err := downloadWithProgressBar(addonListURL(), fileLocation)
	if err != nil {
		return addonMap, err
	}

	addonsFile, err := os.ReadFile(fileLocation)
	if err != nil {
		return addonMap, err
	}
	err = yaml.Unmarshal(addonsFile, addonMap)
	if err != nil {
		return addonMap, err
	}

	return addonMap, nil
}
