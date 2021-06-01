#!/bin/bash

# Create temp path for partial data (storing everything but the commit date.)
PARTIAL_DATA_PATH=$(mktemp)
# Write 
echo "Partial path: $PARTIAL_DATA_PATH" 1>&2

# Print header.
printf "Commit Hash,Commit Date,Environment,Test,Status,Duration\n"

# 1) "cat" together all summary files.
# 2) Turn each test in each summary file to a CSV line containing its commit hash, environment, test, and status.
# 3) Copy partial data to $PARTIAL_DATA_PATH to join with date later.
# 4) Extract only commit hash for each row
# 5) Make the commit hashes unique (we assume that gsutil cats files from the same hash next to each other).
#   Also force buffering to occur per line so remainder of pipe can continue to process.
# 6) Execute git log for each commit to get the date of each.
# 7) Join dates with test data.
gsutil cat gs://minikube-builds/logs/master/*/*_summary.json \
| jq -r '((.PassedTests[]? as $name | {commit: .Detail.Details, environment: .Detail.Name, test: $name, duration: .Durations[$name], status: "Passed"}),
          (.FailedTests[]? as $name | {commit: .Detail.Details, environment: .Detail.Name, test: $name, duration: .Durations[$name], status: "Failed"}),
          (.SkippedTests[]? as $name | {commit: .Detail.Details, environment: .Detail.Name, test: $name, duration: 0, status: "Skipped"}))
          | .commit + "," + .environment + "," + .test + "," + .status + "," + (.duration | tostring)' \
| tee $PARTIAL_DATA_PATH \
| sed -r -n 's/^([^,]+),.*/\1/p' \
| stdbuf -oL -eL uniq \
| xargs -I {} git log -1 --pretty=format:"{},%as%n" {} \
| join -t "," - $PARTIAL_DATA_PATH
