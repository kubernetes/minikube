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
	"strings"

	"k8s.io/minikube/pkg/minikube/constants"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

func PrintKubernetesVersionsFromGCS(output io.Writer) {
	PrintKubernetesVersions(output, constants.KubernetesVersionGCSURL)
}

func PrintKubernetesVersions(output io.Writer, url string) {
	k8sVersions, err := GetK8sVersionsFromURL(url)
	if err != nil {
		glog.Errorln(err)
		return
	}
	fmt.Fprint(output, "The following Kubernetes versions are available when using the localkube bootstrapper: \n")

	for _, k8sVersion := range k8sVersions {
		fmt.Fprintf(output, "\t- %s\n", k8sVersion.Version)
	}
}

type K8sRelease struct {
	Version string
}

type K8sReleases []K8sRelease

func getJson(url string, target *K8sReleases) error {
	r, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "Error getting json from url: %s via http", url)
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

var cachedK8sVersions = make(K8sReleases, 0)

func GetK8sVersionsFromURL(url string) (K8sReleases, error) {
	if len(cachedK8sVersions) != 0 {
		return cachedK8sVersions, nil
	}
	var k8sVersions K8sReleases
	if err := getJson(url, &k8sVersions); err != nil {
		return K8sReleases{}, errors.Wrapf(err, "Error getting json via http with url: %s", url)
	}
	if len(k8sVersions) == 0 {
		return K8sReleases{}, errors.Errorf("There were no json k8s Releases at the url specified: %s", url)
	}

	cachedK8sVersions = k8sVersions
	return k8sVersions, nil
}

func IsValidLocalkubeVersion(v string, url string) (bool, error) {
	if strings.HasPrefix(v, "file://") || strings.HasPrefix(v, "http") {
		return true, nil
	}
	k8sReleases, err := GetK8sVersionsFromURL(url)
	glog.Infoln(k8sReleases)
	if err != nil {
		return false, errors.Wrap(err, "Error getting the localkube versions")
	}

	isValidVersion := false
	for _, version := range k8sReleases {
		if version.Version == v {
			isValidVersion = true
			break
		}
	}

	return isValidVersion, nil
}
