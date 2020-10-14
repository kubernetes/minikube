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

package update

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v32/github"
	"k8s.io/klog/v2"
)

const (
	// ghListPerPage uses max value (100) for PerPage to avoid hitting the rate limits
	// (ref: https://godoc.org/github.com/google/go-github/github#hdr-Rate_Limiting)
	ghListPerPage = 100

	// ghSearchLimit limits the number of searched items to be <= N * ListPerPage
	ghSearchLimit = 100
)

var (
	// GitHub repo data
	ghToken = os.Getenv("GITHUB_TOKEN")
	ghOwner = "kubernetes"
	ghRepo  = "minikube"
	ghBase  = "master" // could be "main" in the future?
)

// ghCreatePR returns PR created in the GitHub owner/repo, applying the changes to the base head
// commit fork, as defined by the schema and data, and also returns any error occurred
// PR branch will be named by the branch, sufixed by '_' and first 7 characters of fork commit SHA
// PR itself will be named by the title and will reference the issue
func ghCreatePR(ctx context.Context, owner, repo, base, branch, title string, issue int, token string, schema map[string]Item, data interface{}) (*github.PullRequest, error) {
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
	changes, err := ghUpdate(ctx, owner, repo, baseTree, token, schema, data)
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
	plan, err := GetPlan(schema, data)
	if err != nil {
		klog.Fatalf("Error parsing schema: %v\n%s", err, plan)
	}
	modifiable := true
	pr, _, err := ghc.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(*fork.Owner.Login + ":" + prBranch),
		Base:                github.String(base),
		Body:                github.String(fmt.Sprintf("fixes #%d\n\nAutomatically created PR to update repo according to the Plan:\n\n```\n%s\n```", issue, plan)),
		MaintainerCanModify: &modifiable,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating pull request: %w", err)
	}
	return pr, nil
}

// ghUpdate updates remote GitHub owner/repo tree according to the given token, schema and data,
// returns resulting changes, and any error occurred
func ghUpdate(ctx context.Context, owner, repo string, tree *github.Tree, token string, schema map[string]Item, data interface{}) (changes []*github.TreeEntry, err error) {
	ghc := ghClient(ctx, token)

	// load each schema item content and update it creating new GitHub TreeEntries
	cnt := len(schema) // expected number of files to change
	for _, org := range tree.Entries {
		if *org.Type == "blob" {
			if item, match := schema[*org.Path]; match {
				blob, _, err := ghc.Git.GetBlobRaw(ctx, owner, repo, *org.SHA)
				if err != nil {
					return nil, fmt.Errorf("error getting file: %w", err)
				}
				item.Content = blob
				changed, err := item.apply(data)
				if err != nil {
					return nil, fmt.Errorf("error updating file: %w", err)
				}
				if changed {
					// add github.TreeEntry that will replace original path content with updated one
					changes = append(changes, &github.TreeEntry{
						Path:    org.Path,
						Mode:    org.Mode,
						Type:    org.Type,
						Content: github.String(string(item.Content)),
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

// ghFindPR returns URL of the PR if found in the given GitHub ower/repo base and any error occurred
func ghFindPR(ctx context.Context, title, owner, repo, base, token string) (url string, err error) {
	ghc := ghClient(ctx, token)

	// walk through the paginated list of all pull requests, from latest to older releases
	opts := &github.PullRequestListOptions{State: "all", Base: base, ListOptions: github.ListOptions{PerPage: ghListPerPage}}
	for (opts.Page+1)*ghListPerPage <= ghSearchLimit {
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

// GHVersions returns current stable release and latest rc or beta pre-release
// from GitHub owner/repo repository, and any error;
// if latest pre-release version is lower than current stable release, then it
// will return current stable release for both
func GHVersions(ctx context.Context, owner, repo string) (stable, latest string, err error) {
	ghc := ghClient(ctx, ghToken)

	// walk through the paginated list of all owner/repo releases, from newest to oldest
	opts := &github.ListOptions{PerPage: ghListPerPage}
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
