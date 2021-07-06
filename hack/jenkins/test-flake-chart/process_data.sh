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

# Takes a series of gopogh summary jsons, and formats them into a CSV file with
# a row for each test.
# Example usage: cat gopogh_1.json gopogh_2.json gopogh_3.json | ./process_data.sh

set -eu -o pipefail

# Print header.
printf "Commit Hash,Test Date,Environment,Test,Status,Duration\n"

# Turn each test in each summary file to a CSV line containing its commit hash, date, environment, test, status, and duration.
# Example line:
# 247982745892,2021-06-10,Docker_Linux,TestFunctional,Passed,0.5
jq -r '((.PassedTests[]? as $name | {commit: (.Detail.Details | split(":") | .[0]), date: (.Detail.Details | split(":") | .[1]), environment: .Detail.Name, test: $name, duration: .Durations[$name], status: "Passed"}),
          (.FailedTests[]? as $name | {commit: (.Detail.Details | split(":") | .[0]), date: (.Detail.Details | split(":") | .[1]), environment: .Detail.Name, test: $name, duration: .Durations[$name], status: "Failed"}),
          (.SkippedTests[]? as $name | {commit: (.Detail.Details | split(":") | .[0]), date: (.Detail.Details | split(":") | .[1]), environment: .Detail.Name, test: $name, duration: 0, status: "Skipped"}))
          | .commit + "," + .date + "," + .environment + "," + .test + "," + .status + "," + (.duration | tostring)'
