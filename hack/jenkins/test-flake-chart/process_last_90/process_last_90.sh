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

set -eux -o pipefail

# Get directory of script.
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

"${DIR}/../../installers/check_install_golang.sh" "/usr/local" || true

DATA_CSV=$(mktemp)
DATA_LAST_90_CSV=$(mktemp)

gsutil cp gs://minikube-flake-rate/data.csv "$DATA_CSV"

go run "${DIR}/process_last_90.go" --source "$DATA_CSV" --target "$DATA_LAST_90_CSV"

gsutil cp "$DATA_LAST_90_CSV" gs://minikube-flake-rate/data-last-90.csv

rm "$DATA_CSV" "$DATA_LAST_90_CSV"
