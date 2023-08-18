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

# Creates a comment on the provided PR number, using the provided gopogh summary
# to list out the flake rates of all failing tests.
# Example usage: ./report_flakes.sh 11602 gopogh.json Docker_Linux

set -eu -o pipefail

if [ "$#" -ne 3 ]; then
  echo "Wrong number of arguments. Usage: report_flakes.sh <PR number> <Root job id> <environment list file>" 1>&2
  exit 1
fi

PR_NUMBER=$1
ROOT_JOB=$2
ENVIRONMENT_LIST=$3

# To prevent having a super-long comment, add a maximum number of tests to report.
MAX_REPORTED_TESTS=30

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

TMP_DATA=$(mktemp)
# 1) Process the ENVIRONMENT_LIST to turn them into valid GCS URLs.
# 2) Check to see if the files are present. Ignore any missing files.
# 3) Cat the gopogh summaries together.
# 4) Process the data in each gopogh summary.
# 5) Filter tests to only include failed tests (and only get their names and environment).
# 6) Sort by environment, then test name.
# 7) Store in file $TMP_DATA.
sed -r "s|^|gs://minikube-builds/logs/${PR_NUMBER}/${ROOT_JOB}/|; s|$|_summary.json|" "${ENVIRONMENT_LIST}" \
  | (xargs gsutil ls || true) \
  | xargs gsutil cat \
  | "$DIR/process_data.sh" \
  | awk -F, 'NR>1 {
      if ($5 == "Failed") {
        printf "%s:%s\n", $3, $4
      }
    }' \
  | sort \
  > "$TMP_DATA"

# Download the precomputed flake rates from the GCS bucket into file $TMP_FLAKE_RATES.
TMP_FLAKE_RATES=$(mktemp)
gsutil cp gs://minikube-flake-rate/flake_rates.csv "$TMP_FLAKE_RATES"

TMP_FAILED_RATES=$(mktemp)
# 1) Parse the flake rates to only include the environment and test name.
# 2) Sort the environment+test names.
# 3) Get all lines in $TMP_DATA not present in $TMP_FLAKE_RATES.
# 4) Append column containing "n/a" to data.
# 4) Store in $TMP_FAILED_RATES
awk -F, 'NR>1 {
  printf "%s:%s\n", $1, $2
}' "$TMP_FLAKE_RATES" \
  | sort \
  | comm -13 - "$TMP_DATA" \
  | sed -r -e 's|$|,n/a|' \
  > "$TMP_FAILED_RATES"

# 1) Parse the flake rates to only include the environment, test name, and flake rates.
# 2) Sort the flake rates based on environment+test name.
# 3) Join the flake rates with the failing tests to only get flake rates of failing tests.
# 4) Sort failed test flake rates based on the flakiness of that test - stable tests should be first on the list.
# 5) Append to file $TMP_FAILED_RATES.
awk -F, 'NR>1 {
  if ($3 < 50) printf "%s:%s,%s\n", $1, $2, $3
}' "$TMP_FLAKE_RATES" \
  | sort -t, -k1,1 \
  | join -t , -j 1 "$TMP_DATA" - \
  | sort -g -t, -k2,2 \
  >> "$TMP_FAILED_RATES"

FAILED_RATES_LINES=$(wc -l < "$TMP_FAILED_RATES")
if [[ "$FAILED_RATES_LINES" -eq 0 ]]; then
  echo "No failed tests! Aborting without commenting..." 1>&2
  exit 0
fi

# Create the comment template.
TMP_COMMENT=$(mktemp)
printf "These are the flake rates of all failed tests.\n|Environment|Failed Tests|Flake Rate (%%)|\n|---|---|---|\n" > "$TMP_COMMENT"

# Create variables to use for sed command.
ENV_CHART_LINK_FORMAT='https://gopogh-server-tts3vkcpgq-uc.a.run.app/?env=%1$s'
TEST_CHART_LINK_FORMAT=${ENV_CHART_LINK_FORMAT}'&test=%2$s'
TEST_GOPOGH_LINK_FORMAT='https://storage.googleapis.com/minikube-builds/logs/'${PR_NUMBER}'/'${ROOT_JOB}'/%1$s.html#fail_%2$s'
# 1) Get the first $MAX_REPORTED_TESTS lines.
# 2) Print a row in the table with the environment, test name, flake rate, and a link to the flake chart for that test.
# 3) Append these rows to file $TMP_COMMENT.
head -n "$MAX_REPORTED_TESTS" "$TMP_FAILED_RATES" \
  | awk '-F[:,]' '{
      if ($3 != "n/a") {
        rate_text = sprintf("%3$s ([chart]('$TEST_CHART_LINK_FORMAT'))", $1, $2, $3)
      } else {
        rate_text = $3
      }
      printf "|[%1$s]('$ENV_CHART_LINK_FORMAT')|%2$s ([gopogh]('$TEST_GOPOGH_LINK_FORMAT'))|%3$s|\n", $1, $2, rate_text
    }' \
  >> "$TMP_COMMENT"

# If there are too many failing tests, add an extra row explaining this, and a message after the table.
if [[ "$FAILED_RATES_LINES" -gt 30 ]]; then
  printf "|More tests...|Continued...|\n\nToo many tests failed - See test logs for more details." >> "$TMP_COMMENT"
fi

printf "\n\nTo see the flake rates of all tests by environment, click [here](https://minikube.sigs.k8s.io/docs/contrib/test_flakes/)." >> "$TMP_COMMENT"

# install gh if not present
"$DIR/../installers/check_install_gh.sh"

gh pr comment "https://github.com/kubernetes/minikube/pull/$PR_NUMBER" --body "$(cat $TMP_COMMENT)"
