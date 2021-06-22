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

SESSION_ID=$1

RESULTS=()
OS=$(uname)

if [[ ${OS} == "Darwin" ]]; then
  TESTS_TARGETS=("idle" "minikube.hyperkit" "minikube.virtualbox" "minikube.docker" "docker" "k3d" "kind")
elif [[ ${OS} == "Linux" ]]; then
  TESTS_TARGETS=("idle" "minikube.kvm2" "minikube.virtualbox" "minikube.docker" "docker" "k3d" "kind")
fi

# calc average each test target
calcAverage() {
  for target in ${TESTS_TARGETS[@]}; do
    count=0;
    total=0;
    FILES=$(ls out/benchmark-results/${SESSION_ID} | grep cstat.${target})

    # calc average per test target
    for file in ${FILES[@]}; do
      MEASURED=$(cat out/benchmark-results/${SESSION_ID}/${file} | tail -n 1) 
      total=$(echo ${total}+${MEASURED} | bc )
      ((count++))
    done

    RESULT=$(echo "scale=4; ${total} / ${count}" | bc | awk '{printf "%.4f\n", $0}')
    RESULTS=("${RESULTS[@]}" ${RESULT})
  done
}

# create summary csv
updateSummary() {
  for ((i = 0; i < ${#RESULTS[@]}; i++)) {
    echo "${RESULTS[i]}" >> out/benchmark-results/${SESSION_ID}/cstat.summary
  }
}

calcAverage
updateSummary
