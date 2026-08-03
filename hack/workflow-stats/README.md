# workflow-stats

Analyze GitHub Actions workflow step durations for the minikube project.

Fetches completed workflow runs via the GitHub API, caches results in a
local SQLite database, and reports per-step timing statistics (min, avg,
p50, p90, p95, max) along with a suggested timeout.

## Usage

```
# Run from the hack directory
cd hack

# Export GITHUB_TOKEN to avoid rate limits
export GITHUB_TOKEN=$(gh auth token)

# Stats from last 14 days (default), all job variants:
go run workflow-stats/workflow_stats.go -workflow "Functional Test"

# Last 30 days:
go run workflow-stats/workflow_stats.go -workflow "Smoke Test" -since 30

# Filter by job name (substring match):
go run workflow-stats/workflow_stats.go -workflow "Smoke Test" -job "vfkit-docker-macos-15-x86"

# Output formats (table, markdown, csv, json):
go run workflow-stats/workflow_stats.go -workflow "Smoke Test" -o markdown
```

Set `GITHUB_TOKEN` to avoid hitting unauthenticated API rate limits.

Run `go run workflow-stats/workflow_stats.go -help` for the full list of flags.

## Example — Tuning Functional Test Timeouts

### Step 1: Analyze historic data

Run the tool from the `hack/` directory to collect step durations:

```
$ go run workflow-stats/workflow_stats.go -workflow "Functional Test"
Fetching runs since 2026-05-24 ... 50 runs (2.6s)
Fetching 44 new runs ... (46.7s)

Step                                                    N       Min       Avg       P50       P90       P95       Max    Timeout
────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
Run Functional Test                                   932     2m47s     4m05s     3m58s     5m09s     5m23s     9m27s     17m00s
Build minikube and e2e test binaries                  200     1m07s     1m39s     1m38s     2m07s     2m09s     2m21s      7m00s
Set up Rootless Docker (rootless)                      87       42s       47s       46s       53s       55s     1m01s      3m00s
Run actions/setup-go@4a3601121dd01d1626a1e23e3721…   1132        7s       20s       21s       26s       29s     1m01s      2m00s
Update apt-get package index (ubuntu)                 475        5s       11s        8s       20s       23s       48s      2m00s
...
```

The **Timeout** column shows `p95 × 3` (configurable via `-timeout-multiplier`),
rounded up to the nearest minute. This provides enough headroom for normal
variance while catching steps that are genuinely stuck.

### Step 2: Update workflow timeouts

Use the suggested timeout to set `timeout-minutes` on workflow steps. For
example, based on the output above:

```
$ git diff .github/workflows/functional_test.yml
diff --git a/.github/workflows/functional_test.yml b/.github/workflows/functional_test.yml
index f3eb2591e..5797b48ca 100644
--- a/.github/workflows/functional_test.yml
+++ b/.github/workflows/functional_test.yml
@@ -56,6 +56,7 @@ jobs:
       - name: Download Dependencies
         run: go mod download
       - name: Build minikube and e2e test binaries
+        timeout-minutes: 7
         run: |
           make ${{ matrix.make-targets }}
           cp -r test/integration/testdata ./out
@@ -161,6 +162,7 @@ jobs:
           EOF
           sudo systemctl daemon-reload
       - name: Set up Rootless Docker (rootless)
+        timeout-minutes: 3
         if: ${{ matrix.rootless }}
         run: |
           sudo apt-get remove moby-engine-*
@@ -171,6 +173,7 @@ jobs:
         if: contains(matrix.os, 'macos')
         run: brew update
       - name: Update apt-get package index (ubuntu)
+        timeout-minutes: 2
         if: runner.os == 'Linux' && (matrix.driver == 'podman' || matrix.driver == 'none')
         run: sudo apt-get update -qq
       - name: Install cri_dockerd (baremetal only)
@@ -289,6 +292,7 @@ jobs:
           containerd config default | sudo tee /etc/containerd/config.toml
           sudo systemctl restart containerd
       - name: Run Functional Test
+        timeout-minutes: 17
         id: run_test
         continue-on-error: true
         shell: bash
```

## Automation

The `-o json` flag produces machine-readable output suitable for automated
timeout tuning. See [#23043](https://github.com/kubernetes/minikube/issues/23043)
for a proposal to run this automatically as a weekly workflow.
