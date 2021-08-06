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
