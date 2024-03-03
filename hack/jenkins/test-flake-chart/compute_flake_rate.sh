#!/bin/bash

# Copyright 2021 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eux -o pipefail

# Get directory of script.
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# Update html+js of flake charts.
gsutil cp "${DIR}/flake_chart.html" gs://minikube-flake-rate/flake_chart.html
gsutil cp "${DIR}/flake_chart.js" gs://minikube-flake-rate/flake_chart.js

DATA_CSV=$(mktemp)
FLAKE_RATES_CSV=$(mktemp)
# Get raw test data.
gsutil cp gs://minikube-flake-rate/data.csv "${DATA_CSV}"
# Compute flake rates.
go run "${DIR}/compute_flake_rate.go" --data-csv="${DATA_CSV}" --date-range=15 > "${FLAKE_RATES_CSV}"
# Upload flake rates.
gsutil cp "${FLAKE_RATES_CSV}" gs://minikube-flake-rate/flake_rates.csv

# Threshold to open issues at (either creation or re-opening)
OPEN_ISSUE_THRESHOLD=80
# Threshold to close existing issues at
CLOSE_ISSUE_THRESHOLD=20

"${DIR}/../installers/check_install_gh.sh" || true

# Get a list of issues from GitHub and extract only those that look like flake issues.
# Sort by test name for later usage.
EXISTING_ISSUES_LIST=$(mktemp)
gh issue list -L 10000 -s all -A "minikube-bot" -l kind/failing-test \
  | awk '-F\t' 'BEGIN { OFS="," } {
    where = match($3, /^Frequent test failures of `([a-zA-Z0-9.\/_-]*)`$/, captures)
    if (where != 0) {
      print $1, $2, captures[1]
    }
  }' \
  | sort -t , -k 3,3 \
  > "${EXISTING_ISSUES_LIST}"

# Get a list of only the tests for each issue. 
EXISTING_ISSUES_TESTS_ONLY=$(mktemp)
awk -F, '{ print $3 }' "${EXISTING_ISSUES_LIST}" \
  > "${EXISTING_ISSUES_TESTS_ONLY}"

# Get a list of all tests present in the flake rate CSV.
FLAKES_TESTS_ONLY=$(mktemp)
awk -F, 'NR>1 {
  print $2
}' "${FLAKE_RATES_CSV}" \
  | sort \
  | uniq \
  > "${FLAKES_TESTS_ONLY}"

# 1) Get only entries above the close threshold
# 2) Sort by the test name
# 3) Ensure the list is unique
# 4) Store in $MID_FLAKES_DATA
MID_FLAKES_DATA=$(mktemp)
awk -F, 'BEGIN { OFS="," } NR>1 {
    if ($3 >= '${CLOSE_ISSUE_THRESHOLD}') {
      print $1, $2, $3
    }
  }' "${FLAKE_RATES_CSV}" \
  | sort -t , -k 2,2\
  | uniq \
  > "${MID_FLAKES_DATA}"

# 1) Get only the test names from the $MID_FLAKES_DATA
# 2) Ensure the list is unique
# 3) Get only tests not present in the $MID_FLAKES_DATA
CLOSE_ISSUES_LIST=$(mktemp)
awk -F, '{ print $2 }' "${MID_FLAKES_DATA}" \
  | uniq \
  | comm -13 - "${FLAKES_TESTS_ONLY}" \
  > "${CLOSE_ISSUES_LIST}"

# Get test names of issues that are not present in the flake rate CSV and append
# to the close-issues list.
awk -F, 'NR>1 { print $2 }' "${FLAKE_RATES_CSV}" \
  | sort \
  | uniq \
  | comm -13 - "${EXISTING_ISSUES_TESTS_ONLY}" \
  >> "${CLOSE_ISSUES_LIST}"

# 1) Sort the close-issues list
# 2) Ensure the list is unique
# 3) Filter the existing issues to only include issues we intend to close
# 4) Extract only the issue number
# 5) Close the issue
sort "${CLOSE_ISSUES_LIST}" \
  | uniq \
  | join -t , -1 1 -2 3 - "${EXISTING_ISSUES_LIST}" \
  | awk -F, '{ if ($3 == "OPEN") { print $2 } }' \
  | xargs -I % gh issue close %

# Filter the $MID_FLAKES_DATA for tests that surpass the $OPEN_ISSUE_THRESHOLD.
# Also, only return the test name
OPEN_ISSUES_LIST=$(mktemp)
awk -F, '{
    if ($3 >= '${OPEN_ISSUE_THRESHOLD}') {
      print $2
    }
  }' "${MID_FLAKES_DATA}" \
  | uniq \
  > "${OPEN_ISSUES_LIST}"

# 1) Get existing issues that we want to be open
# 2) Filter for only closed issues, and get just the issue number
# 3) Reopen the issue
join -t , -1 1 -2 3 "${OPEN_ISSUES_LIST}" "${EXISTING_ISSUES_LIST}" \
  | awk -F, '{
      if ($3 == "CLOSED") {
        print $2
      }
    }' \
  | xargs -I % gh issue reopen %

# 1) Get only tests without an existing issue
# 2) For each test, create an issue for it and format into a row for $EXISTING_ISSUES_LIST
# 3) Append to $EXISTING_ISSUES_LIST
comm -13 "${EXISTING_ISSUES_TESTS_ONLY}" "${OPEN_ISSUES_LIST}" \
  | xargs -I % sh -c \
    'gh issue create -b "Will be filled in with details" -l kind/failing-test -l priority/backlog -t "Frequent test failures of \`%\`" \
      | sed -n -r "s~^https://github.com/kubernetes/minikube/issues/([0-9]*)$~\1,OPEN,%~p"' \
  >> "${EXISTING_ISSUES_LIST}"

# Re-sort $EXISTING_ISSUES_LIST to account for any newly created issues.
sort -t , -k 3,3 "${EXISTING_ISSUES_LIST}" -o "${EXISTING_ISSUES_LIST}"

# Join the existing issues with those that we wish to report.
# Only take the test name and issue number.
MID_FLAKES_ISSUES=$(mktemp)
join -t , -1 2 -2 3 "${MID_FLAKES_DATA}" "${EXISTING_ISSUES_LIST}" \
  | awk -F, 'BEGIN { OFS="," } { print $1, $4 }' \
  | uniq \
  > "${MID_FLAKES_ISSUES}"

# Go through each high-flake issue.
ISSUE_BODY_TMP=$(mktemp)
for ROW in $(cat ${MID_FLAKES_ISSUES}); do
  # Parse the row into its test name and issue number.
  IFS=','; ROW_ENTRIES=($ROW); unset IFS
  TEST_NAME=${ROW_ENTRIES[0]}
  ISSUE_NUMBER=${ROW_ENTRIES[1]}

  # Clear $ISSUE_BODY_TMP and fill with the standard header.
  printf "This test has high flake rates for the following environments:\n\n|Environment|Flake Rate (%%)|\n|---|---|\n" > "${ISSUE_BODY_TMP}"

  TEST_CHART_LINK_FORMAT='https://gopogh-server-tts3vkcpgq-uc.a.run.app/?env=%1$s&test='${TEST_NAME}
  # 1) Filter $MID_FLAKES_DATA to only include entries with the given test name
  # 2) Sort by flake rates in descending order
  # 3) Format the entry into a row in the table
  # 4) Append all entries to $ISSUE_BODY_TMP
  echo "${TEST_NAME}" \
    | join -t , -1 1 -2 2 - "${MID_FLAKES_DATA}" \
    | sort -t , -g -r -k 3,3 \
    | awk -F, '{ printf "|[%1$s]('$TEST_CHART_LINK_FORMAT')|%2$s|\n", $2, $3 }' \
    >> "${ISSUE_BODY_TMP}"
  
  # Edit the issue body to use $ISSUE_BODY_TMP
  gh issue edit "${ISSUE_NUMBER}" --body "$(cat "$ISSUE_BODY_TMP")"
done
