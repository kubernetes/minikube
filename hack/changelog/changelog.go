/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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
	"flag"
	"fmt"
	"log"
	"maps"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/google/go-github/v81/github"
	"golang.org/x/oauth2"
)

// Config holds knobs for filtering and grouping pull requests when generating
// release notes.
// All string comparisons are case-insensitive.
type Config struct {
	// all PRs with these prefixes will be skipped in the Changelog
	SkipPrefixes []string

	// Group PRs with these prefixes into their own group heading based on their Prefix
	PrefixGroups map[string]string

	//Group PRs with these prefixes into their own group heading based the string they contain
	ContainGroups map[string]string

	// AllowedLabels enumerates the labels that can be used for grouping.
	// Labels outside this list are ignored.
	AllowedLabels []string
}

var cfg = Config{
	// if PR title "starts with" these words, skip them from change log
	SkipPrefixes: []string{
		"ci:",
		"test:",
		"build:",
		"Build(deps):",
		"site:",
		"docs:",
		"chore:",
		"Post-release:"},
	// if PR title "starts with" these words, put them in these group
	PrefixGroups: map[string]string{
		"addon:":       "Addons",
		"addon ":       "Addons",
		"ISO:":         "Base image versions",
		"ISO/Kicabse:": "Base image versions",
		"Kicbase":      "Base image versions",
		"cni":          "CNI",
	},
	// if PR title "contains" these words, put them in these group
	ContainGroups: map[string]string{
		"bug":         "Bug fixes",
		"fix":         "Bug fixes",
		"vfkit:":      "Drivers",
		"qemu:":       "Drivers",
		"driver":      "Drivers",
		"krunkit:":    "Drivers",
		"kvm:":        "Drivers",
		"hyperv:":     "Drivers",
		"hyperkit:":   "Drivers",
		"vbox:":       "Drivers",
		"improve":     "Improvements",
		"translation": "UI",
	},
	AllowedLabels: []string{"kind/feature", "kind/bug", "kind/enhancement"},
}

func main() {

	var (
		owner    = flag.String("owner", "kubernetes", "repository owner")
		repo     = flag.String("repo", "minikube", "repository name")
		startRef = flag.String("start", "", "start git ref (exclusive) (defaults to last release tag)")
		endRef   = flag.String("end", "HEAD", "end git ref (inclusive)")
	)
	flag.Parse()
	fmt.Println("Generating changelog, this might take a few minutes ...")
	ctx := context.Background()
	gh := newGitHubClient(ctx)

	// Determine the starting reference; use the latest release if none is provided.
	start := resolveStartRef(ctx, gh, *owner, *repo, *startRef)

	// Collect and group pull requests between the refs, then print the result.
	prs := pullRequestsBetween(ctx, gh, *owner, *repo, start, *endRef)
	groups := groupPullRequests(prs, cfg)
	printGroups(groups)
}

// newGitHubClient constructs a GitHub client. If GITHUB_TOKEN is set the client
// uses it for authentication to avoid strict rate limits.
func newGitHubClient(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	httpClient := http.DefaultClient
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = oauth2.NewClient(ctx, ts)
	}
	return github.NewClient(httpClient)
}

// resolveStartRef returns the starting git reference. When start is empty the
// most recent GitHub release tag is used.
func resolveStartRef(ctx context.Context, gh *github.Client, owner, repo, start string) string {
	if start != "" {
		return start
	}
	rel, _, err := gh.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		log.Fatalf("get latest release: %v", err)
	}
	return rel.GetTagName()
}

// pullRequestsBetween finds all pull requests merged between start and end by
// walking the commit comparison and querying PRs for each commit.
func pullRequestsBetween(ctx context.Context, gh *github.Client, owner, repo, start, end string) map[int]*github.PullRequest {
	var allCommits []*github.RepositoryCommit
	opts := &github.ListOptions{PerPage: 100}
	for {
		cmp, resp, err := gh.Repositories.CompareCommits(ctx, owner, repo, start, end, opts)
		if err != nil {
			log.Fatalf("compare commits: %v", err)
		}
		allCommits = append(allCommits, cmp.Commits...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	prsMap := make(map[int]*github.PullRequest)
	for _, c := range allCommits {
		prs, _, err := gh.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, c.GetSHA(), nil)
		if err != nil {
			log.Printf("list PRs for commit %s: %v", c.GetSHA(), err)
			continue
		}
		for _, pr := range prs {
			prsMap[pr.GetNumber()] = pr
		}
	}
	return prsMap
}

// groupPullRequests assigns each pull request to a group based on the provided
// configuration and allowed label list.
func groupPullRequests(prs map[int]*github.PullRequest, cfg Config) map[string][]*github.PullRequest {
	groups := make(map[string][]*github.PullRequest)

	// Build a quick lookup for allowed labels.
	allowed := make(map[string]struct{})
	for _, l := range cfg.AllowedLabels {
		allowed[strings.ToLower(l)] = struct{}{}
	}

	for _, pr := range prs {
		group, skip := classify(pr, cfg, allowed)
		if skip {
			continue
		}
		groups[group] = append(groups[group], pr)
	}
	return groups
}

// classify returns the group name for a pull request and whether it should be
// skipped entirely.
func classify(pr *github.PullRequest, cfg Config, allowed map[string]struct{}) (string, bool) {
	title := pr.GetTitle()
	lower := strings.ToLower(title)

	// Skip PRs with any configured prefix.
	for _, p := range cfg.SkipPrefixes {
		if strings.HasPrefix(lower, strings.ToLower(p)) {
			return "", true
		}
	}

	// Group by specific prefixes first.
	for pref, g := range cfg.PrefixGroups {
		if strings.HasPrefix(lower, strings.ToLower(pref)) {
			return g, false
		}
	}

	// Then look for substring matches. Iterate in a deterministic order so
	// overlapping keywords behave predictably. Longer substrings are checked
	// first to favor more specific matches.
	subs := slices.Collect(maps.Keys(cfg.ContainGroups))
	sort.Slice(subs, func(i, j int) bool {
		if len(subs[i]) == len(subs[j]) {
			return subs[i] < subs[j]
		}
		return len(subs[i]) > len(subs[j])
	})
	for _, sub := range subs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return cfg.ContainGroups[sub], false
		}
	}

	// Finally, use an allowed label if present.
	for _, l := range pr.Labels {
		name := l.GetName()
		if _, ok := allowed[strings.ToLower(name)]; ok {
			return name, false
		}
	}
	return "other", false
}

// printGroups writes the grouped pull requests to stdout. Groups and the
// entries within each group are sorted for stable output.
func printGroups(groups map[string][]*github.PullRequest) {
	groupNames := make([]string, 0, len(groups))
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)
	for _, g := range groupNames {
		fmt.Printf("## %s\n", g)
		prs := groups[g]

		// Sort entries alphabetically by title with PR number as a tiebreaker.
		sort.Slice(prs, func(i, j int) bool {
			ti := strings.ToLower(prs[i].GetTitle())
			tj := strings.ToLower(prs[j].GetTitle())
			if ti == tj {
				return prs[i].GetNumber() < prs[j].GetNumber()
			}
			return ti < tj
		})
		for _, pr := range prs {
			fmt.Printf("* %s (#%d)\n", pr.GetTitle(), pr.GetNumber())
		}
		fmt.Println()
	}
}
