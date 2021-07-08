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

# Collects all test data manually, processes it, and uploads to GCS. This will
# overwrite any existing data. This should only be done for a dryrun, new data
# should be handled exclusively through upload_tests.sh.
# Example usage: ./collect_data_manual.sh

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# 1) "cat" together all summary files.
# 2) Process all summary files.
# 3) Optimize the resulting data.
# 4) Store in GCS bucket.
gsutil cat gs://minikube-builds/logs/master/*/*_summary.json \
| $DIR/process_data.sh \
| $DIR/optimize_data.sh \
| gsutil cp - gs://minikube-flake-rate/data.csv
