/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

// listpullreqs.go lists pull requests since the last release.
package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v25/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	token   string
	fromTag string
	toTag   string
)

var rootCmd = &cobra.Command{
	Use:        "listpullreqs fromTag toTag",
	Short:      "Lists pull requests between two versions in our changelog markdown format",
	ArgAliases: []string{"fromTag", "toTag"},
	Run: func(cmd *cobra.Command, args []string) {
		printPullRequests()
	},
}

const org = "kubernetes"
const repo = "minikube"

func main() {
	rootCmd.Flags().StringVar(&token, "token", "", "Specify personal Github Token if you are hitting a rate limit anonymously. https://github.com/settings/tokens")
	rootCmd.Flags().StringVar(&fromTag, "fromTag", "", "comparison of commits is based on this tag (defaults to the latest tag in the repo)")
	rootCmd.Flags().StringVar(&toTag, "toTag", "master", "this is the commit that is compared with fromTag")
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func printPullRequests() {
	client := getClient()

	releases, _, err := client.Repositories.ListReleases(context.Background(), org, repo, &github.ListOptions{})
	if err != nil {
		logrus.Fatal(err)
	}
	lastReleaseTime := *releases[0].PublishedAt
	fmt.Println(fmt.Sprintf("Collecting pull request that were merged since the last release: %s (%s)", *releases[0].TagName, lastReleaseTime))

	listSize := 1
	for page := 1; listSize > 0; page++ {
		pullRequests, _, err := client.PullRequests.List(context.Background(), org, repo, &github.PullRequestListOptions{
			State:     "closed",
			Sort:      "updated",
			Direction: "desc",
			ListOptions: github.ListOptions{
				PerPage: 100,
				Page:    page,
			},
		})
		if err != nil {
			logrus.Fatal(err)
		}

		seen := 0
		for idx := range pullRequests {
			pr := pullRequests[idx]
			if pr.MergedAt != nil {
				if pr.GetMergedAt().After(lastReleaseTime.Time) {
					fmt.Printf("* %s [#%d](https://github.com/%s/%s/pull/%d)\n", pr.GetTitle(), *pr.Number, org, repo, *pr.Number)
					seen++
				}
			}
		}
		if seen == 0 {
			break
		}
		listSize = len(pullRequests)
	}
}

func getClient() *github.Client {
	if len(token) == 0 {
		return github.NewClient(nil)
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
