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

package kubernetes_versions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

const kubernetesVersionGCSURL = "https://storage.googleapis.com/minikube/k8s_releases.json"

func PrintKubernetesVersionsFromGCS(output io.Writer) {
	PrintKubernetesVersions(output, kubernetesVersionGCSURL)
}

func PrintKubernetesVersions(output io.Writer, url string) {
	k8sVersions, err := getK8sVersionsFromURL(url)
	if err != nil {
		glog.Errorln(err)
		return
	}
	fmt.Fprint(output, "The following Kubernetes versions are available: \n")

	for _, k8sVersion := range k8sVersions {
		fmt.Fprintf(output, "\t- %s\n", k8sVersion.Version)
	}
}

type k8sRelease struct {
	Version string
}

type k8sReleases []k8sRelease

func getJson(url string, target *k8sReleases) error {
	r, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "Error getting json from url: %s via http", url)
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getK8sVersionsFromURL(url string) (k8sReleases, error) {
	var k8sVersions k8sReleases
	if err := getJson(url, &k8sVersions); err != nil {
		return k8sReleases{}, errors.Wrapf(err, "Error getting json via http with url: %s", url)
	}
	if len(k8sVersions) == 0 {
		return k8sReleases{}, errors.Errorf("There were no json k8s Releases at the url specified: %s", url)
	}
	return k8sVersions, nil
}
