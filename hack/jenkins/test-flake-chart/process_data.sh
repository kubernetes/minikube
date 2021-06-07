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

set -eu -o pipefail

# Print header.
printf "Commit Hash,Test Date,Environment,Test,Status,Duration\n"

# Turn each test in each summary file to a CSV line containing its commit hash, date, environment, test, and status.
jq -r '((.PassedTests[]? as $name | {commit: (.Detail.Details | split(":") | .[0]), date: (.Detail.Details | split(":") | .[1]), environment: .Detail.Name, test: $name, duration: .Durations[$name], status: "Passed"}),
          (.FailedTests[]? as $name | {commit: (.Detail.Details | split(":") | .[0]), date: (.Detail.Details | split(":") | .[1]), environment: .Detail.Name, test: $name, duration: .Durations[$name], status: "Failed"}),
          (.SkippedTests[]? as $name | {commit: (.Detail.Details | split(":") | .[0]), date: (.Detail.Details | split(":") | .[1]), environment: .Detail.Name, test: $name, duration: 0, status: "Skipped"}))
          | .commit + "," + .date + "," + .environment + "," + .test + "," + .status + "," + (.duration | tostring)'
