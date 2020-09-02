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

install_release_notes_helper() {
  release_notes_workdir="$(mktemp -d)"
  trap 'rm -rf -- ${release_notes_workdir}' RETURN

  # See https://stackoverflow.com/questions/56842385/using-go-get-to-download-binaries-without-adding-them-to-go-mod for this workaround
  cd "${release_notes_workdir}"
  go mod init release-notes
  GOBIN="$DIR" go get github.com/corneliusweig/release-notes
  cd -
}

if ! [[ -x "${DIR}/release-notes" ]]; then
  echo >&2 'Installing release-notes'
  install_release_notes_helper
fi

git pull git@github.com:kubernetes/minikube master --tags
recent=$(git describe --abbrev=0)

"${DIR}/release-notes" kubernetes minikube --since $recent

echo "Thank you to our contributors for this release! "
git log "$recent".. --format="%aN" --reverse | sort | uniq | awk '{printf "- %s\n", $0 }'
