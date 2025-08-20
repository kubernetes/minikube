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
	// Skip PRs tarting with these prefixies in the Changelog
	SkipPrefixes: []string{
		"ci:",
		"test:",
		"Build(deps):",
		"site:",
		"docs:",
		"chore:",
		"Post-release:"},
	// Group PRs with these prefixes into their own group heding
	PrefixGroups: map[string]string{
		"addon:":       "Addons",
		"addon ":       "Addons",
		"ISO:":         "Base image versions",
		"ISO/Kicabse:": "Base image versions",
		"Kicbase":      "Base image versions",
		"cni":          "CNI",
	},
	ContainGroups: map[string]string{
		"bug":         "Bug fixes",
		"fix":         "Bug fixes",
		"vfkit:":      "Drivers",
		"qemu:":       "Drivers",
		"driver":      "Drivers",
		"krunkit:":    "Drivers",
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

	ctx := context.Background()
	gh := newGitHubClient(ctx)
	start := resolveStartRef(ctx, gh, *owner, *repo, *startRef)
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
	cmp, _, err := gh.Repositories.CompareCommits(ctx, owner, repo, start, end, nil)
	if err != nil {
		log.Fatalf("compare commits: %v", err)
	}
	prsMap := make(map[int]*github.PullRequest)
	for _, c := range cmp.Commits {
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
	for _, p := range cfg.SkipPrefixes {
		if strings.HasPrefix(lower, strings.ToLower(p)) {
			return "", true
		}
	}
	for pref, g := range cfg.PrefixGroups {
		if strings.HasPrefix(lower, strings.ToLower(pref)) {
			return g, false
		}
	}
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
