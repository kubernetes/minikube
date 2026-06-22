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
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v85/github"
	_ "modernc.org/sqlite"
)

type options struct {
	Workflow   string
	JobFilter  string
	Since      int
	Conclusion string
	Output     string
	Owner      string
	Repo       string
	Branch     string
	TimeoutMul float64
	DBPath     string
}

type stepStats struct {
	Name             string  `json:"step"`
	Samples          int     `json:"samples"`
	Min              float64 `json:"min_s"`
	Avg              float64 `json:"avg_s"`
	P50              float64 `json:"p50_s"`
	P90              float64 `json:"p90_s"`
	P95              float64 `json:"p95_s"`
	Max              float64 `json:"max_s"`
	SuggestedTimeout float64 `json:"suggested_timeout_s"`
}

type runInfo struct {
	ID        int64
	CreatedAt time.Time
}

func main() {
	opts := parseOptions()

	ctx := context.Background()
	client := ghClient()
	db := openDB(opts.DBPath)
	defer db.Close()

	updateDB(ctx, client, db, opts)
	stats := computeStats(db, opts)
	printStats(stats, opts.Output)
}

func parseOptions() options {
	var opts options
	var repo string

	flag.StringVar(&opts.Output, "o", "table", "Output format: table, markdown, csv, json")
	flag.StringVar(&opts.Output, "output", "table", "Output format: table, markdown, csv, json")
	flag.StringVar(&opts.Workflow, "workflow", "", "Workflow name (e.g. \"Functional Test\")")
	flag.StringVar(&opts.JobFilter, "job", "", "Job name substring filter")
	flag.IntVar(&opts.Since, "since", 14, "Analyze runs from this many days ago")
	flag.StringVar(&opts.Conclusion, "conclusion", "success", "Filter jobs by conclusion: success, failure, or empty for all")
	flag.StringVar(&repo, "repo", "kubernetes/minikube", "GitHub owner/repo")
	flag.StringVar(&opts.Branch, "branch", "", "Filter runs by branch name")
	flag.Float64Var(&opts.TimeoutMul, "timeout-multiplier", 3, "Multiplier applied to p95 for suggested timeout")
	flag.StringVar(&opts.DBPath, "db", "", "Path to SQLite database for cached data")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Analyze GitHub Actions workflow step durations.

Job data is cached in a local SQLite database so repeated queries
and longer time ranges don't waste API calls.

Usage:
  # Stats from last 14 days (default), all job variants:
  workflow-stats -workflow "Functional Test"

  # Last 30 days:
  workflow-stats -workflow "Smoke Test" -since 30

  # Filter by job name (substring match):
  workflow-stats -workflow "Smoke Test" -job "vfkit-docker-macos-15-x86"
  workflow-stats -workflow "Smoke Test" -job "docker-docker-ubuntu-24.04-x86"
  workflow-stats -workflow "Functional Test" -job "docker-docker-ubuntu24.04-x86"

  # Output formats (default: table):
  workflow-stats -workflow "Smoke Test" -o markdown
  workflow-stats -workflow "Smoke Test" -o csv
  workflow-stats -workflow "Smoke Test" -o json

Flags:
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		log.Fatalf("Invalid -repo format %q, expected owner/repo", repo)
	}
	opts.Owner = parts[0]
	opts.Repo = parts[1]

	if opts.DBPath == "" {
		opts.DBPath = defaultDBPath(repo)
	}

	if opts.Workflow == "" {
		flag.Usage()
		os.Exit(1)
	}

	return opts
}

// ── Fetch and cache ──

func updateDB(ctx context.Context, client *github.Client, db *sql.DB, opts options) {
	fetchSince := latestRunDate(db, opts.Workflow)
	requestedSince := time.Now().UTC().AddDate(0, 0, -opts.Since)
	if fetchSince.Before(requestedSince) {
		fetchSince = requestedSince
	}

	fmt.Fprintf(os.Stderr, "Fetching runs since %s ...", fetchSince.Format("2006-01-02"))
	t := time.Now()
	wfID := findWorkflowID(ctx, client, opts.Owner, opts.Repo, opts.Workflow)
	runs := fetchRuns(ctx, client, opts.Owner, opts.Repo, wfID, opts.Branch, fetchSince)
	fmt.Fprintf(os.Stderr, " %d runs (%.1fs)\n", len(runs), time.Since(t).Seconds())

	cached := cachedRunIDs(db)
	var toFetch []runInfo
	for _, r := range runs {
		if _, ok := cached[r.ID]; !ok {
			toFetch = append(toFetch, r)
		}
	}

	if len(toFetch) > 0 {
		fmt.Fprintf(os.Stderr, "Fetching %d new runs ...", len(toFetch))
		t = time.Now()
		for _, r := range toFetch {
			jobs := fetchJobsForRun(ctx, client, opts.Owner, opts.Repo, r.ID)
			if jobs == nil {
				continue
			}
			insertRun(db, r.ID, opts.Workflow, r.CreatedAt, jobs)
		}
		fmt.Fprintf(os.Stderr, " (%.1fs)\n", time.Since(t).Seconds())
	}
}

// ── Compute stats ──

func computeStats(db *sql.DB, opts options) []stepStats {
	since := time.Now().UTC().AddDate(0, 0, -opts.Since)
	runIDs := runIDsSince(db, opts.Workflow, since)
	durations := collectDurations(db, runIDs, opts.JobFilter, opts.Conclusion)

	var stats []stepStats
	for name, durs := range durations {
		sort.Float64s(durs)
		s := stepStats{
			Name:    name,
			Samples: len(durs),
			Min:     durs[0],
			Avg:     mean(durs),
			P50:     percentile(durs, 50),
			P90:     percentile(durs, 90),
			P95:     percentile(durs, 95),
			Max:     durs[len(durs)-1],
		}
		s.SuggestedTimeout = math.Max(math.Ceil(s.P95*opts.TimeoutMul/60)*60, 60)
		stats = append(stats, s)
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Avg > stats[j].Avg
	})
	return stats
}

// ── GitHub API ──

func ghClient() *github.Client {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return github.NewClient(nil).WithAuthToken(token)
	}
	return github.NewClient(nil)
}

func findWorkflowID(ctx context.Context, client *github.Client, owner, repo, name string) int64 {
	opts := &github.ListOptions{PerPage: 100}
	for {
		wfs, resp, err := client.Actions.ListWorkflows(ctx, owner, repo, opts)
		if err != nil {
			log.Fatalf("Listing workflows: %v", err)
		}
		for _, wf := range wfs.Workflows {
			if wf.GetName() == name {
				return wf.GetID()
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	log.Fatalf("Workflow %q not found", name)
	return 0
}

func fetchRuns(ctx context.Context, client *github.Client, owner, repo string, wfID int64, branch string, since time.Time) []runInfo {
	created := ">=" + since.Format("2006-01-02")
	opts := &github.ListWorkflowRunsOptions{
		Status:      "completed",
		Created:     created,
		ListOptions: github.ListOptions{PerPage: 100},
	}
	if branch != "" {
		opts.Branch = branch
	}

	var runs []runInfo
	for {
		result, resp, err := client.Actions.ListWorkflowRunsByID(ctx, owner, repo, wfID, opts)
		if err != nil {
			log.Fatalf("Listing runs: %v", err)
		}
		for _, r := range result.WorkflowRuns {
			runs = append(runs, runInfo{ID: r.GetID(), CreatedAt: r.GetCreatedAt().Time})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return runs
}

func fetchJobsForRun(ctx context.Context, client *github.Client, owner, repo string, runID int64) []*github.WorkflowJob {
	opts := &github.ListWorkflowJobsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	all := []*github.WorkflowJob{}
	for {
		jobs, resp, err := client.Actions.ListWorkflowJobs(ctx, owner, repo, runID, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n  warning: failed to fetch jobs for run %d: %v\n", runID, err)
			return nil
		}
		all = append(all, jobs.Jobs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return all
}

// ── SQLite cache ──

func defaultDBPath(repo string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Cannot determine home directory: %v", err)
	}
	return filepath.Join(home, ".cache", "workflow-stats", repo, "stats.db")
}

func openDB(path string) *sql.DB {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Fatalf("Creating database directory: %v", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("Opening database: %v", err)
	}

	// WAL mode allows concurrent reads during writes and improves write performance.
	if _, err = db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Fatalf("Setting journal mode: %v", err)
	}
	// NORMAL sync is safe with WAL and avoids an fsync on every transaction.
	if _, err = db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		log.Fatalf("Setting synchronous mode: %v", err)
	}

	// run_id PRIMARY KEY: uniqueness + O(1) lookup by run ID (dbCachedRunIDs).
	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS runs (
			run_id        INTEGER PRIMARY KEY,
			workflow_name TEXT NOT NULL DEFAULT '',
			created_at    TEXT NOT NULL DEFAULT ''
		)`); err != nil {
		log.Fatalf("Creating runs table: %v", err)
	}

	// Optimizes dbLatestRunDate (MAX(created_at) per workflow) and
	// dbRunIDsSince (run IDs for a workflow within a date range).
	if _, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_runs_workflow_created ON runs (workflow_name, created_at)"); err != nil {
		log.Fatalf("Creating index: %v", err)
	}

	// PRIMARY KEY (run_id, job_name, step_number): uniqueness + fast lookup
	// by run_id prefix for dbCollectDurations (all steps for a set of runs).
	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS steps (
			run_id      INTEGER NOT NULL,
			job_name    TEXT    NOT NULL,
			job_conclusion TEXT NOT NULL,
			step_number INTEGER NOT NULL,
			step_name   TEXT    NOT NULL,
			step_conclusion TEXT NOT NULL,
			duration_s  REAL    NOT NULL,
			started_at  TEXT    NOT NULL,
			PRIMARY KEY (run_id, job_name, step_number)
		)`); err != nil {
		log.Fatalf("Creating steps table: %v", err)
	}
	return db
}

func cachedRunIDs(db *sql.DB) map[int64]struct{} {
	rows, err := db.Query("SELECT run_id FROM runs")
	if err != nil {
		log.Fatalf("Querying cached runs: %v", err)
	}
	defer rows.Close()
	cached := map[int64]struct{}{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("Scanning run_id: %v", err)
		}
		cached[id] = struct{}{}
	}
	return cached
}

func insertRun(db *sql.DB, runID int64, workflowName string, createdAt time.Time, jobs []*github.WorkflowJob) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Begin transaction: %v", err)
	}
	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO steps
		(run_id, job_name, job_conclusion, step_number, step_name, step_conclusion, duration_s, started_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		log.Fatalf("Prepare insert: %v", err)
	}
	defer stmt.Close()

	for _, job := range jobs {
		for _, step := range job.Steps {
			d := stepDur(step)
			startedAt := ""
			if step.StartedAt != nil {
				startedAt = step.StartedAt.Time.Format(time.RFC3339)
			}
			if _, err = stmt.Exec(runID, job.GetName(), job.GetConclusion(),
				step.GetNumber(), step.GetName(), step.GetConclusion(),
				d.Seconds(), startedAt); err != nil {
				tx.Rollback()
				log.Fatalf("Inserting step: %v", err)
			}
		}
	}
	if _, err = tx.Exec("INSERT OR IGNORE INTO runs (run_id, workflow_name, created_at) VALUES (?, ?, ?)",
		runID, workflowName, createdAt.Format(time.RFC3339)); err != nil {
		tx.Rollback()
		log.Fatalf("Inserting run: %v", err)
	}
	if err = tx.Commit(); err != nil {
		log.Fatalf("Committing transaction: %v", err)
	}
}

func latestRunDate(db *sql.DB, workflowName string) time.Time {
	var s string
	err := db.QueryRow("SELECT COALESCE(MAX(created_at), '') FROM runs WHERE workflow_name = ? AND created_at != ''",
		workflowName).Scan(&s)
	if err != nil || s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func runIDsSince(db *sql.DB, workflowName string, since time.Time) []int64 {
	rows, err := db.Query("SELECT run_id FROM runs WHERE workflow_name = ? AND created_at >= ?",
		workflowName, since.Format(time.RFC3339))
	if err != nil {
		log.Fatalf("Querying runs since %s: %v", since.Format(time.RFC3339), err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("Scanning run_id: %v", err)
		}
		ids = append(ids, id)
	}
	return ids
}

func collectDurations(db *sql.DB, runIDs []int64, jobFilter, conclusion string) map[string][]float64 {
	if len(runIDs) == 0 {
		return nil
	}
	placeholders := make([]string, len(runIDs))
	args := make([]any, len(runIDs))
	for i, id := range runIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := "SELECT step_name, duration_s FROM steps WHERE run_id IN (" +
		strings.Join(placeholders, ",") + ") AND step_conclusion != 'skipped'"
	if conclusion != "" {
		query += " AND job_conclusion = ?"
		args = append(args, conclusion)
	}
	if jobFilter != "" {
		query += " AND job_name LIKE ?"
		args = append(args, "%"+jobFilter+"%")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatalf("Querying steps: %v", err)
	}
	defer rows.Close()

	durations := map[string][]float64{}
	for rows.Next() {
		var name string
		var dur float64
		if err := rows.Scan(&name, &dur); err != nil {
			log.Fatalf("Scanning step: %v", err)
		}
		durations[name] = append(durations[name], dur)
	}
	return durations
}

// ── Output formatting ──

func printStats(stats []stepStats, format string) {
	switch format {
	case "markdown":
		printMarkdown(stats)
	case "csv":
		printCSV(stats)
	case "json":
		printJSON(stats)
	default:
		printTable(stats)
	}
}

func printTable(stats []stepStats) {
	fmt.Printf("\n%-50s  %5s  %8s  %8s  %8s  %8s  %8s  %8s  %9s\n",
		"Step", "N", "Min", "Avg", "P50", "P90", "P95", "Max", "Timeout")
	fmt.Println(strings.Repeat("─", 120))

	for _, s := range stats {
		name := s.Name
		if len(name) > 50 {
			name = name[:49] + "…"
		}
		fmt.Printf("%-50s  %5d  %8s  %8s  %8s  %8s  %8s  %8s  %9s\n",
			name, s.Samples,
			fmtSec(s.Min), fmtSec(s.Avg), fmtSec(s.P50),
			fmtSec(s.P90), fmtSec(s.P95), fmtSec(s.Max),
			fmtSec(s.SuggestedTimeout))
	}
}

func printMarkdown(stats []stepStats) {
	fmt.Println("| Step | N | Min | Avg | P50 | P90 | P95 | Max | Timeout |")
	fmt.Println("|---|--:|--:|--:|--:|--:|--:|--:|--:|")
	for _, s := range stats {
		fmt.Printf("| %s | %d | %s | %s | %s | %s | %s | %s | %s |\n",
			s.Name, s.Samples,
			fmtSec(s.Min), fmtSec(s.Avg), fmtSec(s.P50),
			fmtSec(s.P90), fmtSec(s.P95), fmtSec(s.Max),
			fmtSec(s.SuggestedTimeout))
	}
}

func printCSV(stats []stepStats) {
	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"step", "samples", "min_s", "avg_s", "p50_s", "p90_s", "p95_s", "max_s", "suggested_timeout_s"})
	for _, s := range stats {
		w.Write([]string{
			s.Name,
			strconv.Itoa(s.Samples),
			strconv.FormatFloat(s.Min, 'f', 1, 64),
			strconv.FormatFloat(s.Avg, 'f', 1, 64),
			strconv.FormatFloat(s.P50, 'f', 1, 64),
			strconv.FormatFloat(s.P90, 'f', 1, 64),
			strconv.FormatFloat(s.P95, 'f', 1, 64),
			strconv.FormatFloat(s.Max, 'f', 1, 64),
			strconv.FormatFloat(s.SuggestedTimeout, 'f', 0, 64),
		})
	}
	w.Flush()
}

func printJSON(stats []stepStats) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(stats)
}

// ── Helpers ──

func formatDuration(d time.Duration) string {
	s := int(d.Seconds())
	if s >= 3600 {
		return fmt.Sprintf("%dh%02dm%02ds", s/3600, (s%3600)/60, s%60)
	}
	if s >= 60 {
		return fmt.Sprintf("%dm%02ds", s/60, s%60)
	}
	return fmt.Sprintf("%ds", s)
}

func stepDur(step *github.TaskStep) time.Duration {
	if step.StartedAt == nil || step.CompletedAt == nil {
		return 0
	}
	return step.CompletedAt.Time.Sub(step.StartedAt.Time)
}

func mean(sorted []float64) float64 {
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	return sum / float64(len(sorted))
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100.0*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func fmtSec(s float64) string {
	return formatDuration(time.Duration(s * float64(time.Second)))
}
