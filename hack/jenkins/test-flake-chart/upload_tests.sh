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

# Takes a gopogh summary in a GCS bucket, extracts test data as a CSV and
# appends to the existing CSV data in the flake rate GCS bucket.
# Example usage: ./upload_tests.sh gs://some-bucket/gopogh_summary.json

set -eu -o pipefail

if [ "$#" -ne 1 ]; then
  echo "Wrong number of arguments. Usage: upload_tests.sh <gopogh_summary.json>" 1>&2
  exit 1
fi

TMP_DATA=$(mktemp)

# Use the gopogh summary, process it, optimize the data, remove the header, and store.
gsutil cat "$1" \
  | ./test-flake-chart/process_data.sh \
  | ./test-flake-chart/optimize_data.sh \
  | sed "1d" > $TMP_DATA

GCS_TMP="gs://minikube-flake-rate/$(basename "$TMP_DATA")"

# Copy data to append to GCS
gsutil cp $TMP_DATA $GCS_TMP
# Append data to existing data.
gsutil compose gs://minikube-flake-rate/data.csv $GCS_TMP gs://minikube-flake-rate/data.csv
# Clear all the temp stuff.
rm $TMP_DATA
gsutil rm $GCS_TMP
