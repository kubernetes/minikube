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
	"path/filepath"

	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/version"
)

func addonStatusURL() string {
	return fmt.Sprintf("https://%s/minikube-builds/addons/%s/status.yaml", downloadHost, version.GetVersion())
}

func addonListURL() string {
	return fmt.Sprintf("https://%s/minikube-builds/addons/%s/addons.yaml", downloadHost, version.GetVersion())
}

func AddonStatus() error {
	return downloadNoProgressBar(addonStatusURL(), filepath.Join(localpath.MiniPath(), "addons", "status.yaml"))
}

func AddonList() error {
	return downloadWithProgressBar(addonListURL(), filepath.Join(localpath.MiniPath(), "addons", "addons.yaml"))
}
