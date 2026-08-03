/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v85/github"
	"gopkg.in/yaml.v3"
)

// workflowNameFromFile reads a workflow's top-level "name:", the same name
// GitHub uses to label its runs. A file with no name (rare, e.g. vex.yml)
// is labeled by GitHub using its repo path instead, so that's the fallback.
func workflowNameFromFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return "", fmt.Errorf("parsing %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return "", fmt.Errorf("%s: empty document", path)
	}
	if v := mapValue(doc.Content[0], "name"); v != nil && v.Value != "" {
		return v.Value, nil
	}
	return ".github/workflows/" + filepath.Base(path), nil
}

// step holds one parsed workflow step, plus its source line so we can edit
// it directly instead of rewriting the whole file.
type step struct {
	JobName      string
	DisplayName  string // human-readable, for the report
	MatchKey     string // computed GitHub-reported step name, used to look up stats
	AnchorLine   int    // 1-based line of the step's first key; new keys are inserted right after it
	Indent       int    // 0-based column where the step's keys begin
	TimeoutLine  int    // 1-based line of an existing timeout-minutes value, 0 if absent
	TimeoutValue int    // existing timeout-minutes value, -1 if absent
}

type change struct {
	step     step
	oldValue int // -1 = none
	newValue int
}

// applyResult counts what applyToFile found, so a caller batching many
// files can build one summary across all of them.
type applyResult struct {
	Changed   int
	Unmatched int
}

// applyToFile checks one workflow file against computed stats, then either
// reports the differences (write=false) or edits the file (write=true).
func applyToFile(path string, stats []stepStats, write bool) (applyResult, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return applyResult{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return applyResult{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	steps, err := extractSteps(&doc)
	if err != nil {
		return applyResult{}, fmt.Errorf("%s: %w", path, err)
	}

	statsByName := map[string]stepStats{}
	for _, s := range stats {
		if isPseudoStep(s.Name) {
			continue
		}
		statsByName[s.Name] = s
	}

	var changes []change
	var unmatched []step
	for _, st := range steps {
		s, ok := statsByName[st.MatchKey]
		if !ok {
			unmatched = append(unmatched, st)
			continue
		}
		newVal := int(math.Round(s.SuggestedTimeout / 60))
		if newVal != st.TimeoutValue {
			changes = append(changes, change{st, st.TimeoutValue, newVal})
		}
	}

	printApplyReport(path, changes, unmatched)
	result := applyResult{Changed: len(changes), Unmatched: len(unmatched)}

	if !write || len(changes) == 0 {
		return result, nil
	}
	return result, writeChanges(path, raw, changes)
}

func printApplyReport(path string, changes []change, unmatched []step) {
	fmt.Printf("\n%s\n", path)
	if len(changes) == 0 && len(unmatched) == 0 {
		fmt.Println("  no changes needed")
	}
	for _, c := range changes {
		old := "(none)"
		if c.oldValue >= 0 {
			old = fmt.Sprintf("%dm", c.oldValue)
		}
		fmt.Printf("  job %-14s line %4d  %-55s %6s -> %dm\n",
			c.step.JobName, c.step.AnchorLine, truncate(c.step.DisplayName, 55), old, c.newValue)
	}
	for _, u := range unmatched {
		fmt.Printf("  job %-14s line %4d  %-55s %s\n",
			u.JobName, u.AnchorLine, truncate(u.DisplayName, 55), "no stats data, skipping")
	}
}

// writeChanges edits from the bottom of the file up, so an insert near the
// top never shifts the line numbers of edits still waiting below it.
func writeChanges(path string, raw []byte, changes []change) error {
	crlf := bytes.Contains(raw, []byte("\r\n"))
	trailingNewline := bytes.HasSuffix(raw, []byte("\n"))
	lines := splitLines(raw)

	sort.Slice(changes, func(i, j int) bool { return changes[i].step.AnchorLine > changes[j].step.AnchorLine })
	for _, c := range changes {
		indent := strings.Repeat(" ", c.step.Indent)
		newLine := fmt.Sprintf("%stimeout-minutes: %d", indent, c.newValue)
		if c.step.TimeoutLine > 0 {
			lines[c.step.TimeoutLine-1] = newLine
		} else {
			idx := c.step.AnchorLine // AnchorLine is 1-based, so this index is right after it
			lines = append(lines[:idx:idx], append([]string{newLine}, lines[idx:]...)...)
		}
	}

	eol := "\n"
	if crlf {
		eol = "\r\n"
	}
	out := strings.Join(lines, eol)
	if trailingNewline {
		out += eol
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

func splitLines(raw []byte) []string {
	s := strings.ReplaceAll(string(raw), "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// extractSteps walks the parsed YAML and returns every step in every job,
// in file order. It only reads structure, so comments and formatting are
// never touched.
func extractSteps(doc *yaml.Node) ([]step, error) {
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("empty document")
	}
	root := doc.Content[0]
	jobsNode := mapValue(root, "jobs")
	if jobsNode == nil {
		return nil, fmt.Errorf("no top-level 'jobs' key found")
	}

	var steps []step
	for i := 0; i+1 < len(jobsNode.Content); i += 2 {
		jobName := jobsNode.Content[i].Value
		stepsNode := mapValue(jobsNode.Content[i+1], "steps")
		if stepsNode == nil || stepsNode.Kind != yaml.SequenceNode {
			continue
		}
		for _, stepMap := range stepsNode.Content {
			if stepMap.Kind != yaml.MappingNode {
				continue
			}
			steps = append(steps, parseStep(jobName, stepMap))
		}
	}
	return steps, nil
}

func parseStep(jobName string, stepMap *yaml.Node) step {
	st := step{
		JobName:      jobName,
		AnchorLine:   stepMap.Line,
		Indent:       stepMap.Column - 1,
		TimeoutValue: -1,
	}

	// GitHub reports a step under its own name when one is set.
	if v := mapValue(stepMap, "name"); v != nil {
		st.DisplayName = v.Value
		st.MatchKey = v.Value
	}
	// With no name, GitHub reports an action step as "Run <uses>".
	if v := mapValue(stepMap, "uses"); v != nil && st.MatchKey == "" {
		st.DisplayName = "uses: " + v.Value
		st.MatchKey = "Run " + v.Value
	}
	// With no name and no uses, GitHub reports "Run <first line of the script>".
	if v := mapValue(stepMap, "run"); v != nil && st.MatchKey == "" {
		first := strings.SplitN(strings.TrimSpace(v.Value), "\n", 2)[0]
		st.DisplayName = "run: " + first
		st.MatchKey = "Run " + first
	}
	// Remember the existing line so we overwrite it instead of adding a duplicate.
	if kn, vn := mapKeyValue(stepMap, "timeout-minutes"); kn != nil {
		st.TimeoutLine = vn.Line
		if n, err := strconv.Atoi(strings.TrimSpace(vn.Value)); err == nil {
			st.TimeoutValue = n
		}
	}
	return st
}

func mapValue(n *yaml.Node, key string) *yaml.Node {
	_, v := mapKeyValue(n, key)
	return v
}

func mapKeyValue(n *yaml.Node, key string) (*yaml.Node, *yaml.Node) {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil, nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i], n.Content[i+1]
		}
	}
	return nil, nil
}

// isPseudoStep skips job-lifecycle and action-cleanup entries that GitHub
// reports but that don't exist as a line in the workflow YAML.
func isPseudoStep(name string) bool {
	return name == "Set up job" || name == "Complete job" || strings.HasPrefix(name, "Post Run ") || strings.HasPrefix(name, "Post ")
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

// fileResult is one file's outcome from a batch run, kept so the final
// summary can group computed updates and missing-data cases.
type fileResult struct {
	Name string
	applyResult
}

// runApplyDir batch-processes every *.yml file in dir. Each file gets its
// own workflow name and stats fetch, same as running -apply by hand.
func runApplyDir(ctx context.Context, client *github.Client, db *sql.DB, opts options) error {
	files, err := filepath.Glob(filepath.Join(opts.ApplyDir, "*.yml"))
	if err != nil {
		return fmt.Errorf("listing %s: %w", opts.ApplyDir, err)
	}
	sort.Strings(files)

	var results []fileResult
	for _, file := range files {
		base := filepath.Base(file)

		name, err := workflowNameFromFile(file)
		if err != nil {
			return err
		}
		fileOpts := opts
		fileOpts.Workflow = name

		var stats []stepStats
		if err := updateDB(ctx, client, db, fileOpts); err != nil {
			if !errors.Is(err, errWorkflowNotFound) {
				return err
			}
			fmt.Fprintf(os.Stderr, " %v; reporting missing stats\n", err)
		} else {
			stats = computeStats(db, fileOpts)
		}

		res, err := applyToFile(file, stats, opts.Write)
		if err != nil {
			return err
		}
		results = append(results, fileResult{Name: base, applyResult: res})
	}

	printBatchSummary(results)
	if opts.PRBody != "" {
		return os.WriteFile(opts.PRBody, []byte(buildPRBody(results)), 0o644)
	}
	return nil
}

func printBatchSummary(results []fileResult) {
	var changedFiles, unchangedFiles, unmatchedFiles int
	for _, r := range results {
		if r.Changed > 0 {
			changedFiles++
		}
		if r.Unmatched > 0 {
			unmatchedFiles++
		}
		if r.Changed == 0 && r.Unmatched == 0 {
			unchangedFiles++
		}
	}
	fmt.Println("\n=== Summary ===")
	fmt.Printf("Files with updated timeouts: %d\n", changedFiles)
	fmt.Printf("Files already up to date:    %d\n", unchangedFiles)
	fmt.Printf("Files with unmatched steps:  %d\n", unmatchedFiles)
}

// buildPRBody turns a batch run's results into updated and needs-attention
// buckets. Missing data is reported from the run, not a hard-coded list.
func buildPRBody(results []fileResult) string {
	var computed, unchanged, unmatched []string
	for _, r := range results {
		switch {
		case r.Changed > 0:
			computed = append(computed, fmt.Sprintf("- `%s` (%d %s updated)", r.Name, r.Changed, plural(r.Changed, "step")))
		case r.Unmatched == 0:
			unchanged = append(unchanged, fmt.Sprintf("- `%s`", r.Name))
		}
		if r.Unmatched > 0 {
			unmatched = append(unmatched, fmt.Sprintf("- `%s` (%d %s with no stats data)", r.Name, r.Unmatched, plural(r.Unmatched, "step")))
		}
	}

	var b strings.Builder
	fmt.Fprintln(&b, "Automated timeout-minutes update from workflow-stats.")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "### Workflows updated (%d)\n\n", len(computed))
	writeLinesOrNone(&b, computed)
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "### Workflows already up to date (%d)\n\n", len(unchanged))
	writeLinesOrNone(&b, unchanged)
	if len(unmatched) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "### Needs attention (steps with no matching stats data)")
		fmt.Fprintln(&b)
		for _, line := range unmatched {
			fmt.Fprintln(&b, line)
		}
	}
	return b.String()
}

func writeLinesOrNone(b *strings.Builder, lines []string) {
	if len(lines) == 0 {
		fmt.Fprintln(b, "- None")
		return
	}
	for _, line := range lines {
		fmt.Fprintln(b, line)
	}
}
