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

	"github.com/cenkalti/backoff/v4"

	"k8s.io/klog/v2"
)

const (
	// FSRoot is a relative (to scripts in subfolders) root folder of local filesystem repo to update
	FSRoot = "../../../"
)

var (
	target = os.Getenv("UPDATE_TARGET")
)

// init klog and check general requirements
func init() {
	klog.InitFlags(nil)
	if err := flag.Set("logtostderr", "false"); err != nil {
		klog.Warningf("Unable to set flag value for logtostderr: %v", err)
	}
	if err := flag.Set("alsologtostderr", "true"); err != nil {
		klog.Warningf("Unable to set flag value for alsologtostderr: %v", err)
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

// Item defines Content where all occurrences of each Replace map key,
// corresponding to GitHub TreeEntry.Path and/or local filesystem repo file path (prefixed with FSRoot),
// would be swapped with its respective actual map value (having placeholders replaced with data), creating a concrete update plan.
// Replace map keys can use RegExp and map values can use Golang Text Template.
type Item struct {
	Content []byte            `json:"-"`
	Replace map[string]string `json:"replace"`
}

// apply updates Item Content by replacing all occurrences of Replace map's keys with their actual map values (with placeholders replaced with data).
func (i *Item) apply(data interface{}) error {
	if i.Content == nil {
		return fmt.Errorf("unable to update content: nothing to update")
	}
	org := string(i.Content)
	str := org
	for src, dst := range i.Replace {
		out, err := ParseTmpl(dst, data, "")
		if err != nil {
			return err
		}
		re := regexp.MustCompile(src)
		str = re.ReplaceAllString(str, out)
	}
	i.Content = []byte(str)

	return nil
}

// Apply applies concrete update plan (schema + data) to GitHub or local filesystem repo
func Apply(ctx context.Context, schema map[string]Item, data interface{}, prBranchPrefix, prTitle string, prIssue int) {
	schema, pretty, err := GetPlan(schema, data)
	if err != nil {
		klog.Fatalf("Unable to parse schema: %v\n%s", err, pretty)
	}
	klog.Infof("The Plan:\n%s", pretty)

	if target == "fs" || target == "all" {
		changed, err := fsUpdate(FSRoot, schema, data)
		if err != nil {
			klog.Errorf("Unable to update local repo: %v", err)
		} else if !changed {
			klog.Infof("Local repo update skipped: nothing changed")
		} else {
			klog.Infof("Local repo successfully updated")
		}
	}

	if target == "gh" || target == "all" {
		// update prTitle replacing template placeholders with actual data values
		if prTitle, err = ParseTmpl(prTitle, data, "prTitle"); err != nil {
			klog.Fatalf("Unable to parse PR Title: %v", err)
		}

		// check if PR already exists
		prURL, err := ghFindPR(ctx, prTitle, ghOwner, ghRepo, ghBase, ghToken)
		if err != nil {
			klog.Errorf("Unable to check if PR already exists: %v", err)
		} else if prURL != "" {
			klog.Infof("PR create skipped: already exists (%s)", prURL)
		} else {
			// create PR
			pr, err := ghCreatePR(ctx, ghOwner, ghRepo, ghBase, prBranchPrefix, prTitle, prIssue, ghToken, schema, data)
			if err != nil {
				klog.Fatalf("Unable to create PR: %v", err)
			} else if pr == nil {
				klog.Infof("PR create skipped: nothing changed")
			} else {
				klog.Infof("PR successfully created: %s", *pr.HTMLURL)
			}
		}
	}
}

// GetPlan returns concrete plan replacing placeholders in schema with actual data values, returns JSON-formatted representation of the plan and any error occurred.
func GetPlan(schema map[string]Item, data interface{}) (plan map[string]Item, prettyprint string, err error) {
	plan = make(map[string]Item)
	for p, item := range schema {
		path, err := ParseTmpl(p, data, "")
		if err != nil {
			return plan, fmt.Sprintf("%+v", schema), err
		}
		plan[path] = item
	}

	for _, item := range plan {
		for src, dst := range item.Replace {
			out, err := ParseTmpl(dst, data, "")
			if err != nil {
				return plan, fmt.Sprintf("%+v", schema), err
			}
			item.Replace[src] = out
		}
	}
	str, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return plan, fmt.Sprintf("%+v", schema), err
	}

	return plan, string(str), nil
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

// ParseTmpl replaces placeholders in text with actual data values
func ParseTmpl(text string, data interface{}, name string) (string, error) {
	tmpl := template.Must(template.New(name).Parse(text))
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
