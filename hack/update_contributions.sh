#!/bin/bash

# Copyright 2018 The Kubernetes Authors All rights reserved.
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

if ! [[ -r "${DIR}/gh_token.txt" ]]; then
  echo "Missing '${DIR}/gh_token.txt'. Please create a GitHub token at https://github.com/settings/tokens and store in '${DIR}/gh_token.txt'."
  exit 1
fi

install_pullsheet() {
  pullsheet_workdir="$(mktemp -d)"
  trap 'rm -rf -- ${pullsheet_workdir}' RETURN

  # See https://stackoverflow.com/questions/56842385/using-go-get-to-download-binaries-without-adding-them-to-go-mod for this workaround
  cd "${pullsheet_workdir}"
  go mod init ps
  GOBIN="$DIR" go get github.com/google/pullsheet
  cd -
}

if ! [[ -x "${DIR}/pullsheet" ]]; then
  echo >&2 'Installing pullsheet'
  install_pullsheet
fi

git pull git@github.com:kubernetes/minikube master --tags

tags_to_generate=${1:-1}

# 1) Get tags.
# 2) Filter out beta tags.
# 3) Parse tag name into its version numbers.
# 4) Sort by ascending version numbers.
# 5) Reform tag name from version numbers.
# 6) Pair up current and previous tags. Format: (previous tag, current tag)
# 7) Get dates of previous and current tag. Format: (current tag, prev date, current date)
# 8) Add negative line numbers to each tag. Format: (negative index, current tag, prev date, current date)
#   - Negative line numbers are used since entries are sorted in descending order.
# 9) Take most recent $tags_to_generate tags.
tags_with_range=$(
  git --no-pager tag \
  | grep -v -e "beta" \
  | sed -r "s/v([0-9]*)\.([0-9]*)\.([0-9]*)/\1 \2 \3/" \
  | sort -k1n -k2n -k3n \
  | sed -r "s/([0-9]*) ([0-9]*) ([0-9]*)/v\1.\2.\3/" \
  | sed -n -r "x; G; s/\n/ /; p"\
  | sed -n -r "s/([v.0-9]+) ([v.0-9]+)/sh -c '{ echo -n \2; git log -1 --pretty=format:\" %as \" \1; git log -1 --pretty=format:\"%as\" \2;}'/ep" \
  | sed "=" | sed -r "N;s/\n/ /;s/^/-/" \
  | tail -n "$tags_to_generate")

destination="$DIR/../site/content/en/docs/contrib/leaderboard"
mkdir -p "$destination"

while read -r tag_index tag_name tag_start tag_end; do
  echo "Generating leaderboard for" "$tag_name" "(from $tag_start to $tag_end)"
  # Print header for page.
  printf -- "---\ntitle: \"$tag_name - $tag_end\"\nlinkTitle: \"$tag_name - $tag_end\"\nweight: $tag_index\n---\n" > "$destination/$tag_name.html"
  # Add pullsheet content (deleting the lines consisting of the command used to generate it).
  $DIR/pullsheet leaderboard --token-path "$DIR/gh_token.txt" --repos kubernetes/minikube --since "$tag_start" --until "$tag_end" --logtostderr=false --stderrthreshold=2 \
    | sed -r -e "/Command\-line/,/pullsheet/d" >> "$destination/$tag_name.html"
done <<< "$tags_with_range"
