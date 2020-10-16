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
Script expects the following env variables:
 - UPDATE_TARGET=<string>: optional - if unset/absent, default option is "fs"; valid options are:
   - "fs"  - update only local filesystem repo files [default]
   - "gh"  - update only remote GitHub repo files and create PR (if one does not exist already)
   - "all" - update local and remote repo files and create PR (if one does not exist already)
 - GITHUB_TOKEN=<string>: GitHub [personal] access token
   - note: GITHUB_TOKEN is required if UPDATE_TARGET is "gh" or "all"
*/

package update

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"text/template"
	"time"

	"k8s.io/klog/v2"

	"github.com/cenkalti/backoff/v4"
)

const (
	// FSRoot is relative (to scripts in subfolders) root folder of local filesystem repo to update
	FSRoot = "../../../"
)

var (
	target = os.Getenv("UPDATE_TARGET")
)

// init klog and check general requirements
func init() {
	klog.InitFlags(nil)
	// write log statements to stderr instead of to files
	if err := flag.Set("logtostderr", "true"); err != nil {
		fmt.Printf("Error setting 'logtostderr' klog flag: %v\n", err)
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
}

// Item defines Content where all occurrences of each Replace map key, corresponding to
// GitHub TreeEntry.Path and/or local filesystem repo file path (prefixed with FSRoot),
// would be swapped with its respective actual map value (having placeholders replaced with data),
// creating a concrete update plan.
// Replace map keys can use RegExp and map values can use Golang Text Template
type Item struct {
	Content []byte            `json:"-"`
	Replace map[string]string `json:"replace"`
}

// apply updates Item Content by replacing all occurrences of Replace map's keys
// with their actual map values (with placeholders replaced with data))
func (i *Item) apply(data interface{}) (changed bool, err error) {
	if i.Content == nil || i.Replace == nil {
		return false, fmt.Errorf("want something, got nothing to update")
	}
	org := string(i.Content)
	str := org
	for src, dst := range i.Replace {
		tmpl := template.Must(template.New("").Parse(dst))
		buf := new(bytes.Buffer)
		if err := tmpl.Execute(buf, data); err != nil {
			return false, err
		}
		re := regexp.MustCompile(src)
		str = re.ReplaceAllString(str, buf.String())
	}
	i.Content = []byte(str)

	return str != org, nil
}

// Apply applies concrete update plan (schema + data) to GitHub or local filesystem repo
func Apply(ctx context.Context, schema map[string]Item, data interface{}, prBranchPrefix, prTitle string, prIssue int) {
	plan, err := GetPlan(schema, data)
	if err != nil {
		klog.Fatalf("Error parsing schema: %v\n%s", err, plan)
	}
	klog.Infof("The Plan:\n%s", plan)

	if target == "fs" || target == "all" {
		changed, err := fsUpdate(FSRoot, schema, data)
		if err != nil {
			klog.Errorf("Error updating local repo: %v", err)
		} else if !changed {
			klog.Infof("Local repo update skipped: nothing changed")
		} else {
			klog.Infof("Local repo updated")
		}
	}

	if target == "gh" || target == "all" {
		// update prTitle replacing template placeholders with actual data values
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
			pr, err := ghCreatePR(ctx, ghOwner, ghRepo, ghBase, prBranchPrefix, prTitle, prIssue, ghToken, schema, data)
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

// GetPlan returns concrete plan replacing placeholders in schema with actual data values,
// returns JSON-formatted representation of the plan and any error
func GetPlan(schema map[string]Item, data interface{}) (prettyprint string, err error) {
	for _, item := range schema {
		for src, dst := range item.Replace {
			tmpl := template.Must(template.New("").Parse(dst))
			buf := new(bytes.Buffer)
			if err := tmpl.Execute(buf, data); err != nil {
				return fmt.Sprintf("%+v", schema), err
			}
			item.Replace[src] = buf.String()
		}
	}
	str, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", schema), err
	}
	return string(str), nil
}

// RunWithRetryNotify runs command cmd with stdin using exponential backoff for maxTime duration
// up to maxRetries (negative values will make it ignored),
// notifies about any intermediary errors and return any final error.
// similar to pkg/util/retry/retry.go:Expo(), just for commands with params and also with context
func RunWithRetryNotify(ctx context.Context, cmd *exec.Cmd, stdin io.Reader, maxTime time.Duration, maxRetries uint64) error {
	be := backoff.NewExponentialBackOff()
	be.Multiplier = 2
	be.MaxElapsedTime = maxTime
	bm := backoff.WithMaxRetries(be, maxRetries)
	bc := backoff.WithContext(bm, ctx)

	notify := func(err error, wait time.Duration) {
		klog.Errorf("Temporary error running '%s' (will retry in %s): %v", cmd.String(), wait, err)
	}
	if err := backoff.RetryNotify(func() error {
		cmd.Stdin = stdin
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			time.Sleep(be.NextBackOff().Round(1 * time.Second))
			return fmt.Errorf("%w: %s", err, stderr.String())
		}
		return nil
	}, bc, notify); err != nil {
		return err
	}
	return nil
}

// Run runs command cmd with stdin
func Run(cmd *exec.Cmd, stdin io.Reader) error {
	cmd.Stdin = stdin
	var out bytes.Buffer
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, out.String())
	}
	return nil
}
