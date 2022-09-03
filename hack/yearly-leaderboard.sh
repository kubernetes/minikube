#!/bin/bash

# Copyright 2022 The Kubernetes Authors All rights reserved.
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

set -eu -o pipefail

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
DESTINATION="$DIR/../site/content/en/docs/contrib/leaderboard"
TMP_TOKEN=$(mktemp)
BUCKET="s3://minikube-leaderboard"
YEAR=$(date +"%Y" --date "last month")
MONTH=$(date +"%m" --date "last month")
DAYS_IN_MONTH=$(cal $(date +"%m %Y" --date "last month") | awk 'NF {DAYS = $NF}; END {print DAYS}')

install_pullsheet() {
	echo >&2 'Installing pullsheet'
	go install github.com/google/pullsheet@latest
}

verify_gh_auth() {
	gh auth status -t 2>&1 | sed -n -r 's/^.*Token: ([a-zA-Z0-9_]*)/\1/p' > "$TMP_TOKEN"
	if [ ! -s "$TMP_TOKEN" ]; then
		echo "Failed to acquire token from 'gh auth'. Ensure 'gh' is authenticated." 1>&2
		exit 1
	fi
}

# Ensure the token is deleted when the script exits, so the token is not leaked.
cleanup_token() {
	rm -f "$TMP_TOKEN"
}

copy() {
	aws s3 cp "$1" "$2"
}

generate_leaderboard() {
	echo "Generating leaderboard for" "$YEAR"
	# Print header for page
	printf -- "---\ntitle: \"$YEAR\"\nlinkTitle: \"$YEAR\"\nweight: -9999$YEAR\n---\n" > "$DESTINATION/$YEAR.html"
	# Add pullsheet content
	pullsheet leaderboard --token-path "$TMP_TOKEN" --repos kubernetes/minikube --since-display "$YEAR-01-01" --since "$YEAR-$MONTH-01" --until "$YEAR-$MONTH-$DAYS_IN_MONTH" --json-files "./$YEAR.json" --json-output "./$YEAR.json" --hide-command --logtostderr=false --stderrthreshold=2 >> "$DESTINATION/$YEAR.html"
}

cleanup() {
	rm "$YEAR.json"
}

install_pullsheet

verify_gh_auth

trap cleanup_token EXIT

copy "$BUCKET/$YEAR.json" "$YEAR.json" || printf -- "{}" > "$YEAR.json"

generate_leaderboard

copy "$YEAR.json" "$BUCKET/$YEAR.json"
copy "$YEAR.json" "$BUCKET/$YEAR-$MONTH.json"

cleanup
