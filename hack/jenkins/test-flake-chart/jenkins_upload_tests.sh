#!/bin/bash

set -x -o pipefail

if [ "$#" -ne 1 ]; then
  echo "Wrong number of arguments. Usage: jenkins_upload_tests.sh <gopogh_summary.json>" 1>&2
  exit 1
fi

TMP_DATA=$(mktemp)

# Use the gopogh summary, process it, optimize the data, remove the header, and store.
<"$1" ./test-flake-chart/process_data.sh \
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
