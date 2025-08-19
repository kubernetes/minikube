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

// stringSlice returns a slice from a comma separated flag value.
func stringSlice(val string) []string {
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// prefixMap parses mapping of prefix=group pairs separated by comma.
func prefixMap(val string) map[string]string {
	m := make(map[string]string)
	for _, pair := range stringSlice(val) {
		eq := strings.Index(pair, "=")
		if eq == -1 {
			continue
		}
		prefix := strings.TrimSpace(pair[:eq])
		group := strings.TrimSpace(pair[eq+1:])
		if prefix != "" && group != "" {
			m[prefix] = group
		}
	}
	return m
}

func main() {
	var (
		owner        = flag.String("owner", "kubernetes", "repository owner")
		repo         = flag.String("repo", "minikube", "repository name")
		startRef     = flag.String("start", "", "start git ref (exclusive) (defaults to last release tag)")
		endRef       = flag.String("end", "HEAD", "end git ref (inclusive)")
		skipPrefixes = flag.String("skip-prefixes", "ci:,test:,Build(deps):,site:,Post-release:", "comma separated list of title prefixes to skip")
		groupPref    = flag.String("group-prefixes", "addon =addons,Kicbase/ISO:=Base Image Versions,cni:=CNI,ui:=UI, translations:UI", "comma separated prefix=group mappings")
		labelAllow   = flag.String("labels", "kind/feature,kind/bug,kind/enhancement", "comma separated list of labels allowed for grouping")
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

	skip := stringSlice(*skipPrefixes)
	groups := make(map[string][]*github.PullRequest)
	prefixGroups := prefixMap(*groupPref)
	allowed := make(map[string]struct{})
	for _, l := range stringSlice(*labelAllow) {
		allowed[strings.ToLower(l)] = struct{}{}
	}

	for _, pr := range prsMap {
		title := pr.GetTitle()
		lower := strings.ToLower(title)
		skipThis := false
		for _, p := range skip {
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
			return prs[i].GetNumber() < prs[j].GetNumber()
		})
		for _, pr := range prs {
			fmt.Printf("- %s (#%d)\n", pr.GetTitle(), pr.GetNumber())
		}
		fmt.Println()
	}
}
