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


# This script uploads the test reports to the GCS bucket

# The script expects the following env variables:
# UPSTREAM_JOB: the name of the job that needs logs uploaded
# FILE_NAME: the name of the file upload to

set -x

# upload results to GCS
UPSTREAM_JOB=${UPSTREAM_JOB%"_integration"}

JOB_GCS_BUCKET="minikube-builds/logs/${MINIKUBE_LOCATION}/${COMMIT:0:7}/${UPSTREAM_JOB}"

ARTIFACTS=artifacts/test_reports

ls -l $ARTIFACTS

TEST_OUT="$ARTIFACTS/out.txt"
echo ">> uploading ${TEST_OUT} to gs://${JOB_GCS_BUCKET}.txt"
gsutil -qm cp "${TEST_OUT}" "gs://${JOB_GCS_BUCKET}.txt" || true

JSON_OUT="$ARTIFACTS/out.json"
echo ">> uploading ${JSON_OUT}"
gsutil -qm cp "${JSON_OUT}" "gs://${JOB_GCS_BUCKET}.json" || true

HTML_OUT="$ARTIFACTS/out.html"
echo ">> uploading ${HTML_OUT}"
gsutil -qm cp "${HTML_OUT}" "gs://${JOB_GCS_BUCKET}.html" || true

SUMMARY_OUT="$ARTIFACTS/summary.txt"
echo ">> uploading ${SUMMARY_OUT}"
gsutil -qm cp "${SUMMARY_OUT}" "gs://${JOB_GCS_BUCKET}_summary.json" || true
