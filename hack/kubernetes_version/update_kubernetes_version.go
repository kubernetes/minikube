/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

/*
The script expects the following env variables:
 - UPDATE_TARGET=<string>: optional - if unset/absent, default option is "fs"; valid options are:
   - "fs"  - update only local filesystem repo files [default]
   - "gh"  - update only remote GitHub repo files and create PR (if one does not exist already)
   - "all" - update local and remote repo files and create PR (if one does not exist already)
 - GITHUB_TOKEN=<string>: The Github API access token. Injected by the Jenkins credential provider.
   - note: GITHUB_TOKEN is needed only if UPDATE_TARGET is "gh" or "all"
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"golang.org/x/oauth2"

	"k8s.io/klog/v2"

	"github.com/google/go-github/v32/github"
)

const (
	// default context timeout
	cxTimeout = 300 * time.Second

	// use max value (100) for PerPage to avoid hitting the rate limits (60 per hour, 10 per minute)
	// see https://godoc.org/github.com/google/go-github/github#hdr-Rate_Limiting
	ghListOptionsPerPage = 100
)

var (
	// root directory of the local filesystem repo to update
	fsRoot = "../../"

	// map key corresponds to GitHub TreeEntry.Path and local repo file path (prefixed with fsRoot)
	plan = map[string]Patch{
		"pkg/minikube/constants/constants.go": {
			Replace: map[string]string{
				`DefaultKubernetesVersion = \".*`: `DefaultKubernetesVersion = "{{.K8sStableVersion}}"`,
				`NewestKubernetesVersion = \".*`:  `NewestKubernetesVersion = "{{.K8sLatestVersion}}"`,
			},
		},
		"site/content/en/docs/commands/start.md": {
			Replace: map[string]string{
				`'stable' for .*,`:  `'stable' for {{.K8sStableVersion}},`,
				`'latest' for .*\)`: `'latest' for {{.K8sLatestVersion}})`,
			},
		},
	}

	target = os.Getenv("UPDATE_TARGET")

	// GitHub repo data
	ghToken = os.Getenv("GITHUB_TOKEN")
	ghOwner = "kubernetes"
	ghRepo  = "minikube"
	ghBase  = "master" // could be "main" in the future?

	// PR data
	prBranchPrefix = "update-kubernetes-version_" // will be appended with first 7 characters of the PR commit SHA
	prTitle        = `update_kubernetes_version: {stable:"{{.K8sStableVersion}}", latest:"{{.K8sLatestVersion}}"}`
	prIssue        = 4392
	prSearchLimit  = 100 // limit the number of previous PRs searched for same prTitle to be <= N * ghListOptionsPerPage
)

// Data holds respective stable (release) and latest (pre-release) Kubernetes versions
type Data struct {
	K8sStableVersion string `json:"k8sStableVersion"`
	K8sLatestVersion string `json:"k8sLatestVersion"`
}

// Patch defines content where all occurrences of each replace map key should be swapped with its
// respective value. Replace map keys can use RegExp and values can use Golang Text Template
type Patch struct {
	Content []byte            `json:"-"`
	Replace map[string]string `json:"replace"`
}

// apply patch to content by replacing all occurrences of map's keys with their respective values
func (p *Patch) apply(data interface{}) (changed bool, err error) {
	if p.Content == nil || p.Replace == nil {
		return false, fmt.Errorf("nothing to patch")
	}
	org := string(p.Content)
	str := org
	for src, dst := range p.Replace {
		re := regexp.MustCompile(src)
		tmpl := template.Must(template.New("").Parse(dst))
		buf := new(bytes.Buffer)
		if err := tmpl.Execute(buf, data); err != nil {
			return false, err
		}
		str = re.ReplaceAllString(str, buf.String())
	}
	p.Content = []byte(str)

	return str != org, nil
}

func main() {
	klog.InitFlags(nil)
	// write log statements to stderr instead of to files
	if err := flag.Set("logtostderr", "true"); err != nil {
		fmt.Printf("Error setting 'logtostderr' klog flag: %v", err)
	}
	flag.Parse()
	defer klog.Flush()

	if target == "" {
		target = "fs"
	} else if target != "fs" && target != "gh" && target != "all" {
		klog.Fatalf("Invalid UPDATE_TARGET option: '%s'; Valid options are: unset/absent (defaults to 'fs'), 'fs', 'gh', or 'all'", target)
	} else if (target == "gh" || target == "all") && ghToken == "" {
		klog.Fatalf("GITHUB_TOKEN is required if UPDATE_TARGET is 'gh' or 'all'")
	}

	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// get Kubernetes versions from GitHub Releases
	stable, latest, err := ghReleases(ctx, "kubernetes", "kubernetes", ghToken)
	if err != nil || stable == "" || latest == "" {
		klog.Fatalf("Error getting Kubernetes versions: %v", err)
	}
	data := Data{K8sStableVersion: stable, K8sLatestVersion: latest}
	klog.Infof("Kubernetes versions: 'stable' is %s and 'latest' is %s", data.K8sStableVersion, data.K8sLatestVersion)

	klog.Infof("The Plan:\n%s", thePlan(plan, data))

	if target == "fs" || target == "all" {
		changed, err := fsUpdate(fsRoot, plan, data)
		if err != nil {
			klog.Errorf("Error updating local repo: %v", err)
		} else if !changed {
			klog.Infof("Local repo update skipped: nothing changed")
		} else {
			klog.Infof("Local repo updated")
		}
	}

	if target == "gh" || target == "all" {
		// update prTitle replacing template placeholders with concrete data values
		tmpl := template.Must(template.New("prTitle").Parse(prTitle))
		buf := new(bytes.Buffer)
		if err := tmpl.Execute(buf, data); err != nil {
			klog.Fatalf("Error parsing PR Title: %v", err)
		}
		prTitle = buf.String()

		// check if PR already exists
		prURL, err := ghFindPR(ctx, prTitle, ghOwner, ghRepo, ghBase, ghToken)
		if err != nil {
			klog.Errorf("Error checking if PR already exists: %v", err)
		} else if prURL != "" {
			klog.Infof("PR create skipped: already exists (%s)", prURL)
		} else {
			// create PR
			pr, err := ghCreatePR(ctx, ghOwner, ghRepo, ghBase, prBranchPrefix, prTitle, prIssue, ghToken, plan, data)
			if err != nil {
				klog.Fatalf("Error creating PR: %v", err)
			} else if pr == nil {
				klog.Infof("PR create skipped: nothing changed")
			} else {
				klog.Infof("PR created: %s", *pr.HTMLURL)
			}
		}
	}
}

func fsUpdate(fsRoot string, plan map[string]Patch, data Data) (changed bool, err error) {
	for path, p := range plan {
		path = filepath.Join(fsRoot, path)
		blob, err := ioutil.ReadFile(path)
		if err != nil {
			return false, err
		}
		info, err := os.Stat(path)
		if err != nil {
			return false, err
		}
		mode := info.Mode()

		p.Content = blob
		chg, err := p.apply(data)
		if err != nil {
			return false, err
		}
		if chg {
			changed = true
		}
		if err := ioutil.WriteFile(path, p.Content, mode); err != nil {
			return false, err
		}
	}
	return changed, nil
}

func ghCreatePR(ctx context.Context, owner, repo, base, branch, title string, issue int, token string, plan map[string]Patch, data Data) (*github.PullRequest, error) {
	ghc := ghClient(ctx, token)

	// get base branch
	baseBranch, _, err := ghc.Repositories.GetBranch(ctx, owner, repo, base)
	if err != nil {
		return nil, fmt.Errorf("error getting base branch: %w", err)
	}

	// get base commit
	baseCommit, _, err := ghc.Repositories.GetCommit(ctx, owner, repo, *baseBranch.Commit.SHA)
	if err != nil {
		return nil, fmt.Errorf("error getting base commit: %w", err)
	}

	// get base tree
	baseTree, _, err := ghc.Git.GetTree(ctx, owner, repo, baseCommit.GetSHA(), true)
	if err != nil {
		return nil, fmt.Errorf("error getting base tree: %w", err)
	}

	// update files
	changes, err := ghUpdate(ctx, owner, repo, baseTree, token, plan, data)
	if err != nil {
		return nil, fmt.Errorf("error updating files: %w", err)
	}
	if changes == nil {
		return nil, nil
	}

	// create fork
	fork, resp, err := ghc.Repositories.CreateFork(ctx, owner, repo, nil)
	// https://pkg.go.dev/github.com/google/go-github/v32@v32.1.0/github#RepositoriesService.CreateFork
	// This method might return an *AcceptedError and a status code of 202. This is because this is
	// the status that GitHub returns to signify that it is now computing creating the fork in a
	// background task. In this event, the Repository value will be returned, which includes the
	// details about the pending fork. A follow up request, after a delay of a second or so, should
	// result in a successful request.
	if resp.StatusCode == 202 { // *AcceptedError
		time.Sleep(time.Second * 5)
	} else if err != nil {
		return nil, fmt.Errorf("error creating fork: %w", err)
	}

	// create fork tree from base and changed files
	forkTree, _, err := ghc.Git.CreateTree(ctx, *fork.Owner.Login, *fork.Name, *baseTree.SHA, changes)
	if err != nil {
		return nil, fmt.Errorf("error creating fork tree: %w", err)
	}

	// create fork commit
	forkCommit, _, err := ghc.Git.CreateCommit(ctx, *fork.Owner.Login, *fork.Name, &github.Commit{
		Message: github.String(title),
		Tree:    &github.Tree{SHA: forkTree.SHA},
		Parents: []*github.Commit{{SHA: baseCommit.SHA}},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating fork commit: %w", err)
	}
	klog.Infof("PR commit '%s' created: %s", forkCommit.GetSHA(), forkCommit.GetHTMLURL())

	// create PR branch
	prBranch := branch + forkCommit.GetSHA()[:7]
	prRef, _, err := ghc.Git.CreateRef(ctx, *fork.Owner.Login, *fork.Name, &github.Reference{
		Ref: github.String("refs/heads/" + prBranch),
		Object: &github.GitObject{
			Type: github.String("commit"),
			SHA:  forkCommit.SHA,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating PR branch: %w", err)
	}
	klog.Infof("PR branch '%s' created: %s", prBranch, prRef.GetURL())

	// create PR
	modifiable := true
	pr, _, err := ghc.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(*fork.Owner.Login + ":" + prBranch),
		Base:                github.String(base),
		Body:                github.String(fmt.Sprintf("fixes #%d\n\nAutomatically created PR to update repo according to the Plan:\n\n```\n%s\n```", issue, thePlan(plan, data))),
		MaintainerCanModify: &modifiable,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating pull request: %w", err)
	}
	return pr, nil
}

func ghUpdate(ctx context.Context, owner, repo string, tree *github.Tree, token string, plan map[string]Patch, data Data) (changes []*github.TreeEntry, err error) {
	ghc := ghClient(ctx, token)

	// load each plan's path content and update it creating new GitHub TreeEntries
	cnt := len(plan) // expected number of files to change
	for _, org := range tree.Entries {
		if *org.Type == "blob" {
			if patch, match := plan[*org.Path]; match {
				blob, _, err := ghc.Git.GetBlobRaw(ctx, owner, repo, *org.SHA)
				if err != nil {
					return nil, fmt.Errorf("error getting file: %w", err)
				}
				patch.Content = blob
				changed, err := patch.apply(data)
				if err != nil {
					return nil, fmt.Errorf("error patching file: %w", err)
				}
				if changed {
					// add github.TreeEntry that will replace original path content with patched one
					changes = append(changes, &github.TreeEntry{
						Path:    org.Path,
						Mode:    org.Mode,
						Type:    org.Type,
						Content: github.String(string(patch.Content)),
					})
				}
				if cnt--; cnt == 0 {
					break
				}
			}
		}
	}
	if cnt != 0 {
		return nil, fmt.Errorf("error finding all the files (%d missing) - check the Plan: %w", cnt, err)
	}
	return changes, nil
}

func ghFindPR(ctx context.Context, title, owner, repo, base, token string) (url string, err error) {
	ghc := ghClient(ctx, token)

	// walk through the paginated list of all pull requests, from latest to older releases
	opts := &github.PullRequestListOptions{State: "all", Base: base, ListOptions: github.ListOptions{PerPage: ghListOptionsPerPage}}
	for (opts.Page+1)*ghListOptionsPerPage <= prSearchLimit {
		prs, resp, err := ghc.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return "", err
		}
		for _, pr := range prs {
			if pr.GetTitle() == title {
				return pr.GetHTMLURL(), nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return "", nil
}

// ghReleases returns current stable release and latest rc or beta pre-release
// from GitHub owner/repo repository, and any error;
// if latest pre-release version is lower than current stable release, then it
// will return current stable release for both
func ghReleases(ctx context.Context, owner, repo, token string) (stable, latest string, err error) {
	ghc := ghClient(ctx, token)

	// walk through the paginated list of all owner/repo releases, from newest to oldest
	opts := &github.ListOptions{PerPage: ghListOptionsPerPage}
	for {
		rls, resp, err := ghc.Repositories.ListReleases(ctx, owner, repo, opts)
		if err != nil {
			return "", "", err
		}
		for _, rl := range rls {
			ver := rl.GetName()
			if ver == "" {
				continue
			}
			// check if ver version is a release (ie, 'v1.19.2') or a
			// pre-release (ie, 'v1.19.3-rc.0' or 'v1.19.0-beta.2') channel ch
			// note: github.RepositoryRelease GetPrerelease() bool would be useful for all pre-rels
			ch := strings.Split(ver, "-")
			if len(ch) == 1 && stable == "" {
				stable = ver
			} else if len(ch) > 1 && latest == "" {
				if strings.HasPrefix(ch[1], "rc") || strings.HasPrefix(ch[1], "beta") {
					latest = ver
				}
			}
			if stable != "" && latest != "" {
				// make sure that v.Latest >= stable
				if latest < stable {
					latest = stable
				}
				return stable, latest, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return stable, latest, nil
}

// ghClient returns GitHub Client with a given context and optional token for authenticated requests
func ghClient(ctx context.Context, token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// thePlan parses and returns updated plan replacing template placeholders with concrete data values
func thePlan(plan map[string]Patch, data Data) (prettyprint string) {
	for _, p := range plan {
		for src, dst := range p.Replace {
			tmpl := template.Must(template.New("").Parse(dst))
			buf := new(bytes.Buffer)
			if err := tmpl.Execute(buf, data); err != nil {
				klog.Fatalf("Error parsing the Plan: %v", err)
				return fmt.Sprintf("%+v", plan)
			}
			p.Replace[src] = buf.String()
		}
	}
	str, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		klog.Fatalf("Error parsing the Plan: %v", err)
		return fmt.Sprintf("%+v", plan)
	}
	return string(str)
}
