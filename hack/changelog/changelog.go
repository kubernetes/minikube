package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

type Config struct {
	SkipPrefixes  []string
	PrefixGroups  map[string]string
	AllowedLabels []string
}

var cfg = Config{
	// Skip PRs tarting with these prefixies in the Changelog
	SkipPrefixes: []string{
		"ci:",
		"test:",
		"Build(deps):",
		"site:",
		"Post-release:"},
	// Group PRs with these prefixes into their own group heding
	PrefixGroups: map[string]string{
		"addon:":       "addons",
		"addon ":       "addons",
		"ISO/Kicabse:": "base image",
		"ISO:":         "base image",
		"Kicbase":      "base image",
		"cni":          "CNI",
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

	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	httpClient := http.DefaultClient
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = oauth2.NewClient(ctx, ts)
	}
	gh := github.NewClient(httpClient)

	if *startRef == "" {
		rel, _, err := gh.Repositories.GetLatestRelease(ctx, *owner, *repo)
		if err != nil {
			log.Fatalf("get latest release: %v", err)
		}
		*startRef = rel.GetTagName()
	}

	cmp, _, err := gh.Repositories.CompareCommits(ctx, *owner, *repo, *startRef, *endRef, nil)
	if err != nil {
		log.Fatalf("compare commits: %v", err)
	}

	prsMap := make(map[int]*github.PullRequest)
	for _, c := range cmp.Commits {
		prs, _, err := gh.PullRequests.ListPullRequestsWithCommit(ctx, *owner, *repo, c.GetSHA(), nil)
		if err != nil {
			log.Printf("list PRs for commit %s: %v", c.GetSHA(), err)
			continue
		}
		for _, pr := range prs {
			prsMap[pr.GetNumber()] = pr
		}
	}

	groups := make(map[string][]*github.PullRequest)
	prefixGroups := cfg.PrefixGroups
	allowed := make(map[string]struct{})
	for _, l := range cfg.AllowedLabels {
		allowed[strings.ToLower(l)] = struct{}{}
	}

	for _, pr := range prsMap {
		title := pr.GetTitle()
		lower := strings.ToLower(title)
		skipThis := false
		for _, p := range cfg.SkipPrefixes {
			if strings.HasPrefix(lower, strings.ToLower(p)) {
				skipThis = true
				break
			}
		}
		if skipThis {
			continue
		}

		group := ""
		for pref, g := range prefixGroups {
			if strings.HasPrefix(lower, strings.ToLower(pref)) {
				group = g
				break
			}
		}
		if group == "" {
			for _, l := range pr.Labels {
				name := l.GetName()
				if _, ok := allowed[strings.ToLower(name)]; ok {
					group = name
					break
				}
			}
		}
		if group == "" {
			group = "other"
		}
		groups[group] = append(groups[group], pr)
	}

	groupNames := make([]string, 0, len(groups))
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)

	for _, g := range groupNames {
		fmt.Printf("## %s\n", g)
		prs := groups[g]
		sort.Slice(prs, func(i, j int) bool {
			ti := strings.ToLower(prs[i].GetTitle())
			tj := strings.ToLower(prs[j].GetTitle())
			if ti == tj {
				return prs[i].GetNumber() < prs[j].GetNumber()
			}
			return ti < tj
		})
		for _, pr := range prs {
			fmt.Printf("- %s (#%d)\n", pr.GetTitle(), pr.GetNumber())
		}
		fmt.Println()
	}
}
