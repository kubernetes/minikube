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
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"golang.org/x/mod/semver"
	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	cxTimeout = 1 * time.Minute

	// ghListPerPage uses max value (100) for PerPage to avoid hitting the rate limits.
	// (ref: https://pkg.go.dev/github.com/google/go-github/github#hdr-Rate_Limiting)
	ghListPerPage = 100

	// ghSearchLimit limits the number of searched items to be <= N * ghListPerPage.
	ghSearchLimit = 300
)

var schema = map[string]update.Item{
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`ingress-nginx/controller:.*`:          `{{.Controller}}",`,
			`ingress-nginx/kube-webhook-certgen.*`: `{{.Webhook}}",`,
		},
	},
}

type Data struct {
	Controller string
	Webhook    string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	tag, err := LatestControllerTag(ctx)
	if err != nil {
		klog.Fatalf("Unable to get controller tag: %v", err)
	}
	res, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/kubernetes/ingress-nginx/%s/deploy/static/provider/kind/deploy.yaml", tag))
	if err != nil {
		klog.Fatalf("failed to get deploy.yaml: %v", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		klog.Fatalf("failed to read body: %v", err)
	}
	controllerRegex := regexp.MustCompile(`ingress-nginx\/controller.*`)
	controllerImage := controllerRegex.Find(body)
	webhookRegex := regexp.MustCompile(`ingress-nginx\/kube-webhook-certgen.*`)
	webhookImage := webhookRegex.Find(body)

	data := Data{Controller: string(controllerImage), Webhook: string(webhookImage)}

	update.Apply(schema, data)
}

func LatestControllerTag(ctx context.Context) (string, error) {
	latest := "v0.0.0"
	ghc := github.NewClient(nil)
	re := regexp.MustCompile(`controller-(.*)`)

	// walk through the paginated list of up to ghSearchLimit newest releases
	opts := &github.ListOptions{PerPage: ghListPerPage}
	for (opts.Page+1)*ghListPerPage <= ghSearchLimit {
		rls, resp, err := ghc.Repositories.ListReleases(ctx, "kubernetes", "ingress-nginx", opts)
		if err != nil {
			return "", err
		}
		for _, rl := range rls {
			ver := rl.GetTagName()
			if !strings.HasPrefix(ver, "controller-") {
				continue
			}
			s := re.FindStringSubmatch(ver)
			if len(s) < 2 {
				continue
			}
			vTag := s[1]
			if semver.Prerelease(vTag) != "" {
				continue
			}
			if semver.Compare(vTag, latest) == 1 {
				latest = vTag
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return fmt.Sprintf("controller-%s", latest), nil
}
