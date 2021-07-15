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

# Parse the token from `gh auth`
GH_TOKEN=$(mktemp)
gh auth status -t 2>&1 | sed -n -r 's/^.*Token: ([a-zA-Z0-9_]*)/\1/p' > "$GH_TOKEN"

# Ensure the token is deleted when the script exits, so the token is not leaked.
function cleanup_token() {
  rm -f "$GH_TOKEN"
}
trap cleanup_token EXIT

install_release_notes_helper() {
  release_notes_workdir="$(mktemp -d)"
  trap 'rm -rf -- ${release_notes_workdir}' RETURN

  # See https://stackoverflow.com/questions/56842385/using-go-get-to-download-binaries-without-adding-them-to-go-mod for this workaround
  cd "${release_notes_workdir}"
  go mod init release-notes
  GOBIN="$DIR" go get github.com/corneliusweig/release-notes
  GOBIN="$DIR" go get github.com/google/pullsheet
  cd -
}

if ! [[ -x "${DIR}/release-notes" ]] || ! [[ -x "${DIR}/pullsheet" ]]; then
  echo >&2 'Installing release-notes'
  install_release_notes_helper
fi

git pull https://github.com/kubernetes/minikube.git master --tags
recent=$(git describe --abbrev=0)
recent_date=$(git log -1 --format=%as $recent)

"${DIR}/release-notes" kubernetes minikube --since $recent

echo ""
echo "For a more detailed changelog, including changes occuring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md)."
echo ""

echo "Thank you to our contributors for this release!"
echo ""
git log "$recent".. --format="%aN" --reverse | sort | uniq | awk '{printf "- %s\n", $0 }'
echo ""
echo "Thank you to our PR reviewers for this release!"
echo ""
AWK_FORMAT_ITEM='{printf "- %s (%d comments)\n", $2, $1}'
AWK_REVIEW_COMMENTS='NR>1{arr[$4] += $6 + $7}END{for (a in arr) printf "%d %s\n", arr[a], a}'
"${DIR}/pullsheet" reviews --since "$recent_date" --repos kubernetes/minikube --token-path "$GH_TOKEN" --logtostderr=false --stderrthreshold=2 | awk -F ',' "$AWK_REVIEW_COMMENTS" | sort -k1nr -k2d  | awk -F ' ' "$AWK_FORMAT_ITEM"
echo ""
echo "Thank you to our triage members for this release!"
echo ""
AWK_ISSUE_COMMENTS='NR>1{arr[$4] += $7}END{for (a in arr) printf "%d %s\n", arr[a], a}'
"${DIR}/pullsheet" issue-comments --since "$recent_date" --repos kubernetes/minikube --token-path "$GH_TOKEN" --logtostderr=false --stderrthreshold=2 | awk -F ',' "$AWK_ISSUE_COMMENTS" | sort -k1nr -k2d  | awk -F ' ' "$AWK_FORMAT_ITEM" | head -n 5

if [[ "$recent" != *"beta"* ]]; then
  echo ""
  echo "Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/$recent/) for this release!"
fi

echo ""
echo "Don't forget to run `make update-leaderboard` & `make time-to-k8s-benchmark`!"
