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
	"context"
	"time"

	"github.com/google/go-github/v43/github"
	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

var schema = map[string]update.Item{
	".github/workflows/hide-minikube-bot-comments.yml": {
		Replace: map[string]string{
			`spowelljr/hide-minikube-bot-comments.*`: `spowelljr/hide-minikube-bot-comments@{{.SHA}}`,
		},
	},
}

type Data struct {
	SHA string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	sha := lastCommitSHA(ctx)

	data := Data{SHA: sha}

	update.Apply(schema, data)
}

func lastCommitSHA(ctx context.Context) string {
	opts := &github.CommitsListOptions{ListOptions: github.ListOptions{PerPage: 1}}
	ghc := github.NewClient(nil)
	commits, _, err := ghc.Repositories.ListCommits(ctx, "spowelljr", "hide-minikube-bot-comments", opts)
	if err != nil {
		klog.Fatalf("failed to list commits: %v", err)
	}
	return *commits[0].SHA
}
