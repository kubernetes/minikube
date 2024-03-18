/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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
	"bytes"
	"context"
	"html/template"
	"sort"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/mod/semver"
	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
)

const (
	cxTimeout                  = 5 * time.Minute
	kubernetesVersionsTemplate = `{{range .}}{{"\t"}}"{{.}}",
{{end}}`
)

var schema = map[string]update.Item{
	"pkg/minikube/constants/constants_kubernetes_versions.go": {
		Replace: map[string]string{
			`ValidKubernetesVersions = \[]string{((.|\n)*)}`: `ValidKubernetesVersions = []string{
{{.VersionsList}}}`,
		},
	},
}

type Data struct {
	VersionsList string
}

func main() {
	releases := []string{}

	ghc := github.NewClient(nil)

	opts := &github.ListOptions{PerPage: 100}
	for {
		rls, resp, err := ghc.Repositories.ListReleases(context.Background(), "kubernetes", "kubernetes", opts)
		if err != nil {
			klog.Fatal(err)
		}
		for _, rl := range rls {
			ver := rl.GetTagName()
			if !semver.IsValid(ver) {
				continue
			}
			releases = append([]string{ver}, releases...)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	sort.Slice(releases, func(i, j int) bool { return semver.Compare(releases[i], releases[j]) == 1 })

	formatted, err := formatKubernetesVersionsList(releases)
	if err != nil {
		klog.Fatal(err)
	}

	update.Apply(schema, Data{formatted})
}

func formatKubernetesVersionsList(versions []string) (string, error) {
	imageTemplate := template.New("kubernetesVersionsList")
	t, err := imageTemplate.Parse(kubernetesVersionsTemplate)
	if err != nil {
		klog.Errorf("failed to create kubernetes versions template: %v", err)
		return "", err
	}

	var bytesBuffer bytes.Buffer
	if err := t.Execute(&bytesBuffer, &versions); err != nil {
		return "", err
	}

	return bytesBuffer.String(), nil
}
