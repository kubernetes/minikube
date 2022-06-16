/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
)

const (
	preloadFile      = "pkg/minikube/download/preload.go"
	preloadVersionRE = `v[0-9]*`
)

var (
	schema = map[string]update.Item{
		preloadFile: {
			Replace: map[string]string{
				`PreloadVersion = "v[0-9]*"`: `PreloadVersion = "v{{.UpdateVersion}}"`,
			},
		},
	}
)

// Data holds updated preload version.
type Data struct {
	UpdateVersion string
}

func main() {
	// Get current preload version.
	vCurrent, err := getPreloadVersion()
	if err != nil {
		klog.Fatalf("failed to get current preload version: %v", err)
	}
	if vCurrent == 0 {
		klog.Fatalf("cannot determine current preload version")
	}
	klog.Infof("current preload version: %d", vCurrent)

	// Update preload version.
	updatedVersion := vCurrent + 1

	data := Data{UpdateVersion: fmt.Sprint(updatedVersion)}
	klog.Infof("updated preload version: %s", data.UpdateVersion)

	update.Apply(schema, data)
}

// getPreloadVersion returns current preload version and any error.
func getPreloadVersion() (int, error) {
	blob, err := os.ReadFile(filepath.Join(update.FSRoot, preloadFile))
	if err != nil {
		return 0, err
	}
	// Match PreloadVersion.
	re := regexp.MustCompile(fmt.Sprintf(`PreloadVersion = "%s"`, preloadVersionRE))
	version := re.FindSubmatch(blob)
	if version == nil {
		return 0, nil
	}
	// Match version within PreloadVersion.
	re = regexp.MustCompile(preloadVersionRE)
	version = re.FindSubmatch(version[0])
	if version == nil {
		return 0, nil
	}
	// Convert to integer, drop 'v'.
	current, err := strconv.Atoi(string(version[0])[1:])
	if err != nil {
		return 0, err
	}
	return current, nil
}
