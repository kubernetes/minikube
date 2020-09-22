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

package monitor

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// Client provides the context and client with necessary auth
// for interacting with the Github API
type Client struct {
	ctx context.Context
	*github.Client
	owner string
	repo  string
}

// NewClient returns a github client with the necessary auth
func NewClient(ctx context.Context, owner, repo string) *Client {
	githubToken := os.Getenv(GithubAccessTokenEnvVar)
	// Setup the token for github authentication
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	// Return a client instance from github
	client := github.NewClient(tc)
	return &Client{
		ctx:    ctx,
		Client: client,
		owner:  owner,
		repo:   repo,
	}
}

// CommentOnPR comments message on the PR
func (g *Client) CommentOnPR(pr int, message string) error {
	comment := &github.IssueComment{
		Body: &message,
	}

	log.Printf("Creating comment on PR %d: %s", pr, message)
	_, _, err := g.Client.Issues.CreateComment(g.ctx, g.owner, g.repo, pr, comment)
	if err != nil {
		return errors.Wrap(err, "creating github comment")
	}
	log.Printf("Successfully commented on PR %d.", pr)
	return nil
}

// ListOpenPRsWithLabel returns all open PRs with the specified label
func (g *Client) ListOpenPRsWithLabel(label string) ([]int, error) {
	validPrs := []int{}
	prs, _, err := g.Client.PullRequests.List(g.ctx, g.owner, g.repo, &github.PullRequestListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "listing pull requests")
	}
	for _, pr := range prs {
		if prContainsLabel(pr.Labels, "ok-to-test") {
			validPrs = append(validPrs, pr.GetNumber())
		}
	}
	return validPrs, nil
}

func prContainsLabel(labels []*github.Label, label string) bool {
	for _, l := range labels {
		if l == nil {
			continue
		}
		if l.GetName() == label {
			return true
		}
	}
	return false
}

// NewCommitsExist checks if new commits exist since minikube-pr-bot
// commented on the PR. If so, return true.
func (g *Client) NewCommitsExist(pr int, login string) (bool, error) {
	lastCommentTime, err := g.timeOfLastComment(pr, login)
	if err != nil {
		return false, errors.Wrapf(err, "getting time of last comment by %s on pr %d", login, pr)
	}
	lastCommitTime, err := g.timeOfLastCommit(pr)
	if err != nil {
		return false, errors.Wrapf(err, "getting time of last commit on pr %d", pr)
	}
	return lastCommentTime.Before(lastCommitTime), nil
}

func (g *Client) timeOfLastCommit(pr int) (time.Time, error) {
	var commits []*github.RepositoryCommit

	page := 0
	resultsPerPage := 30
	for {
		c, _, err := g.Client.PullRequests.ListCommits(g.ctx, g.owner, g.repo, pr, &github.ListOptions{
			Page:    page,
			PerPage: resultsPerPage,
		})
		if err != nil {
			return time.Time{}, nil
		}
		commits = append(commits, c...)
		if len(c) < resultsPerPage {
			break
		}
		page++
	}

	lastCommitTime := time.Time{}
	for _, c := range commits {
		if newCommitTime := c.GetCommit().GetAuthor().GetDate(); newCommitTime.After(lastCommitTime) {
			lastCommitTime = newCommitTime
		}
	}
	return lastCommitTime, nil
}

func (g *Client) timeOfLastComment(pr int, login string) (time.Time, error) {
	var comments []*github.IssueComment

	page := 0
	resultsPerPage := 30
	for {
		c, _, err := g.Client.Issues.ListComments(g.ctx, g.owner, g.repo, pr, &github.IssueListCommentsOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: resultsPerPage,
			},
		})
		if err != nil {
			return time.Time{}, nil
		}
		comments = append(comments, c...)
		if len(c) < resultsPerPage {
			break
		}
		page++
	}

	// go through comments backwards to find the most recent
	lastCommentTime := time.Time{}

	for _, c := range comments {
		if u := c.GetUser(); u != nil {
			if u.GetLogin() == login {
				if c.GetCreatedAt().After(lastCommentTime) {
					lastCommentTime = c.GetCreatedAt()
				}
			}
		}
	}

	return lastCommentTime, nil
}
