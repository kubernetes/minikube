/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
)

var schema = map[string]update.Item{
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`Version:\s*".*", // traefik-version`: `Version: "{{.Version}}", // traefik-version`,
		},
	},
}

type Data struct {
	Version string
}

type Output struct {
	OldVersion string `json:"old_version"`
	NewVersion string `json:"new_version"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	addonsPath := filepath.Join(update.FSRoot, "pkg/minikube/assets/addons.go")
	content, err := os.ReadFile(addonsPath)
	if err != nil {
		klog.Fatalf("Unable to read %s: %v", addonsPath, err)
	}

	re := regexp.MustCompile(`Version:\s*"(.*)", // traefik-version`)
	matches := re.FindSubmatch(content)
	oldVersion := ""
	if len(matches) > 1 {
		oldVersion = string(matches[1])
	}

	ghc := update.GHClient()
	release, _, err := ghc.Repositories.GetLatestRelease(ctx, "traefik", "traefik-helm-chart")
	if err != nil {
		klog.Fatalf("Unable to get latest Traefik helm chart version: %v", err)
	}

	newVersion := strings.TrimPrefix(release.GetTagName(), "v")

	if oldVersion != newVersion {
		data := Data{Version: newVersion}
		if err := update.Apply(schema, data); err != nil {
			klog.Fatalf("unable to apply update: %v", err)
		}
	}

	out := Output{
		OldVersion: oldVersion,
		NewVersion: newVersion,
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		klog.Fatalf("Unable to encode JSON output: %v", err)
	}
}
