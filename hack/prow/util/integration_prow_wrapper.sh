#!/bin/bash

# Copyright 2025 The Kubernetes Authors All rights reserved.
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

set -x
TARGET_SCRIPT=$1
# run the target script with minikube user
./hack/prow/util/run_with_minikube_user.sh "$TARGET_SCRIPT"
result=$?
# collect the logs as root user
echo "test finished with exit code $result"

items=("testout.txt" "test.json" "junit-unit.xml" "test.html" "test_summary.json")
for item in "${items[@]}"; do
  if [ -f "${item}" ]; then
    echo "Collecting ${item} to ${ARTIFACTS}/${item}"
    cp "${item}" "${ARTIFACTS}/${item}"
  else
    echo "Warning: ${item} not found, skipping"
  fi
done
exit $result
