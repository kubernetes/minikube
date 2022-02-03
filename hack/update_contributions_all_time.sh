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

TMP_TOKEN=$(mktemp)
gh auth status -t 2>&1 | sed -n -r 's/^.*Token: ([a-zA-Z0-9_]*)/\1/p' > "$TMP_TOKEN"
if [ ! -s "$TMP_TOKEN" ]; then
  echo "Failed to acquire token from 'gh auth'. Ensure 'gh' is authenticated." 1>&2
  exit 1
fi
# Ensure the token is deleted when the script exits, so the token is not leaked.
function cleanup_token() {
  rm -f "$TMP_TOKEN"
}
trap cleanup_token EXIT

echo "Generating leaderboard for all-time"
printf -- "---\ntitle: \"All-time\"\nlinkTitle: \"All-time\"\nweight: -99999999\n---\n" > "$destination/all-time.html"
$DIR/pullsheet leaderboard --token-path "$TMP_TOKEN" --repos kubernetes/minikube --logtostderr=false --stderrthreshold=2 \
    | sed -r -e "/Command\-line/,/pullsheet/d" >> "$destination/all-time.html"
