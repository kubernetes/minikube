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

	"golang.org/x/mod/semver"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v32/github"
	"k8s.io/klog/v2"
)

const (
	// ghListPerPage uses max value (100) for PerPage to avoid hitting the rate limits.
	// (ref: https://godoc.org/github.com/google/go-github/github#hdr-Rate_Limiting)
	ghListPerPage = 100

	// ghSearchLimit limits the number of searched items to be <= N * ghListPerPage.
	ghSearchLimit = 100
)

var (
	// GitHub repo data
	ghToken = os.Getenv("GITHUB_TOKEN")
	ghOwner = "kubernetes"
	ghRepo  = "minikube"
	ghBase  = "master" // could be "main" in the near future?
)

// ghCreatePR returns PR created in the GitHub owner/repo, applying the changes to the base head commit fork, as defined by the schema and data.
// Returns any error occurred.
// PR branch will be named by the branch, sufixed by '_' and first 7 characters of the fork commit SHA.
// PR itself will be named by the title and will reference the issue.
func ghCreatePR(ctx context.Context, owner, repo, base, branch, title string, issue int, token string, schema map[string]Item, data interface{}) (*github.PullRequest, error) {
	ghc := ghClient(ctx, token)

	// get base branch
	baseBranch, _, err := ghc.Repositories.GetBranch(ctx, owner, repo, base)
	if err != nil {
		return nil, fmt.Errorf("unable to get base branch: %w", err)
	}

	// get base commit
	baseCommit, _, err := ghc.Repositories.GetCommit(ctx, owner, repo, *baseBranch.Commit.SHA)
	if err != nil {
		return nil, fmt.Errorf("unable to get base commit: %w", err)
	}

	// get base tree
	baseTree, _, err := ghc.Git.GetTree(ctx, owner, repo, baseCommit.GetSHA(), true)
	if err != nil {
		return nil, fmt.Errorf("unable to get base tree: %w", err)
	}

	// update files
	changes, err := ghUpdate(ctx, owner, repo, token, schema, data)
	if err != nil {
		return nil, fmt.Errorf("unable to update files: %w", err)
	}
	if changes == nil {
		return nil, nil
	}

	// create fork
	fork, resp, err := ghc.Repositories.CreateFork(ctx, owner, repo, nil)
	// "This method might return an *AcceptedError and a status code of 202.
	// This is because this is the status that GitHub returns to signify that it is now computing creating the fork in a background task.
	// In this event, the Repository value will be returned, which includes the details about the pending fork.
	// A follow up request, after a delay of a second or so, should result in a successful request."
	// (ref: https://pkg.go.dev/github.com/google/go-github/v32@v32.1.0/github#RepositoriesService.CreateFork)
	if resp.StatusCode == 202 { // *AcceptedError
		time.Sleep(time.Second * 5)
	} else if err != nil {
		return nil, fmt.Errorf("unable to create fork: %w", err)
	}

	// create fork tree from base and changed files
	forkTree, _, err := ghc.Git.CreateTree(ctx, *fork.Owner.Login, *fork.Name, *baseTree.SHA, changes)
	if err != nil {
		return nil, fmt.Errorf("unable to create fork tree: %w", err)
	}

	// create fork commit
	forkCommit, _, err := ghc.Git.CreateCommit(ctx, *fork.Owner.Login, *fork.Name, &github.Commit{
		Message: github.String(title),
		Tree:    &github.Tree{SHA: forkTree.SHA},
		Parents: []*github.Commit{{SHA: baseCommit.SHA}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create fork commit: %w", err)
	}
	klog.Infof("PR commit '%s' successfully created: %s", forkCommit.GetSHA(), forkCommit.GetHTMLURL())

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
		return nil, fmt.Errorf("unable to create PR branch: %w", err)
	}
	klog.Infof("PR branch '%s' successfully created: %s", prBranch, prRef.GetURL())

	// create PR
	_, pretty, err := GetPlan(schema, data)
	if err != nil {
		klog.Fatalf("Unable to parse schema: %v\n%s", err, pretty)
	}
	modifiable := true
	pr, _, err := ghc.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(*fork.Owner.Login + ":" + prBranch),
		Base:                github.String(base),
		Body:                github.String(fmt.Sprintf("fixes: #%d\n\nAutomatically created PR to update repo according to the Plan:\n\n```\n%s\n```", issue, pretty)),
		MaintainerCanModify: &modifiable,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create PR: %w", err)
	}
	return pr, nil
}

// ghFindPR returns URL of the PR if found in the given GitHub ower/repo base and any error occurred.
func ghFindPR(ctx context.Context, title, owner, repo, base, token string) (url string, err error) {
	ghc := ghClient(ctx, token)

	// walk through the paginated list of up to ghSearchLimit newest pull requests
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

// ghUpdate updates remote GitHub owner/repo tree according to the given token, schema and data.
// Returns resulting changes, and any error occurred.
func ghUpdate(ctx context.Context, owner, repo string, token string, schema map[string]Item, data interface{}) (changes []*github.TreeEntry, err error) {
	ghc := ghClient(ctx, token)

	// load each schema item content and update it creating new GitHub TreeEntries
	for path, item := range schema {
		// if the item's content is already set, give it precedence over any current file content
		var content string
		if item.Content == nil {
			file, _, _, err := ghc.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: ghBase})
			if err != nil {
				return nil, fmt.Errorf("unable to get file content: %w", err)
			}
			content, err = file.GetContent()
			if err != nil {
				return nil, fmt.Errorf("unable to read file content: %w", err)
			}
			item.Content = []byte(content)
		}
		if err := item.apply(data); err != nil {
			return nil, fmt.Errorf("unable to update file: %w", err)
		}
		if content != string(item.Content) {
			// add github.TreeEntry that will replace original path content with the updated one or add new if one doesn't exist already
			// ref: https://developer.github.com/v3/git/trees/#tree-object
			rcPath := path // make sure to copy path variable as its reference (not value!) is passed to changes
			rcMode := "100644"
			rcType := "blob"
			changes = append(changes, &github.TreeEntry{
				Path:    &rcPath,
				Mode:    &rcMode,
				Type:    &rcType,
				Content: github.String(string(item.Content)),
			})
		}
	}
	return changes, nil
}

// GHReleases returns greatest current stable release and greatest latest rc or beta pre-release from GitHub owner/repo repository, and any error occurred.
// If latest pre-release version is lower than the current stable release, then it will return current stable release for both.
func GHReleases(ctx context.Context, owner, repo string) (stable, latest string, err error) {
	ghc := ghClient(ctx, ghToken)

	// walk through the paginated list of up to ghSearchLimit newest releases
	opts := &github.ListOptions{PerPage: ghListPerPage}
	for (opts.Page+1)*ghListPerPage <= ghSearchLimit {
		rls, resp, err := ghc.Repositories.ListTags(ctx, owner, repo, opts)
		if err != nil {
			return "", "", err
		}
		for _, rl := range rls {
			ver := *rl.Name
			if !semver.IsValid(ver) {
				continue
			}
			// check if ver version is release (ie, 'v1.19.2') or pre-release (ie, 'v1.19.3-rc.0' or 'v1.19.0-beta.2')
			prerls := semver.Prerelease(ver)
			if prerls == "" {
				if semver.Compare(ver, stable) == 1 {
					stable = ver
				}
			} else if strings.HasPrefix(prerls, "-rc") || strings.HasPrefix(prerls, "-beta") {
				if semver.Compare(ver, latest) == 1 {
					latest = ver
				}
			}
			// make sure that latest >= stable
			if semver.Compare(latest, stable) == -1 {
				latest = stable
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return stable, latest, nil
}

// ghClient returns GitHub Client with a given context and optional token for authenticated requests.
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
