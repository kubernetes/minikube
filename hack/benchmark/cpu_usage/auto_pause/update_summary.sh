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

NONAUTOPAUSE_RESULTS=()
AUTOPAUSE_RESULTS=()
OS=$(uname)

if [[ ${OS} == "Darwin" ]]; then
  TESTS_TARGETS=("idle" "minikube.hyperkit" "minikube.virtualbox" "minikube.docker" "docker" "k3d" "kind")
elif [[ ${OS} == "Linux" ]]; then
  TESTS_TARGETS=("idle" "minikube.kvm2" "minikube.virtualbox" "minikube.docker" "docker" "k3d" "kind")
fi

# calc average each non-autopause test target
calcAverageNonAutopause() {
  for target in ${TESTS_TARGETS[@]}; do
    nap_count=0;
    nap_total=0;
    if [[ "${target}" == "minikube."* ]]; then
      FILES=$(ls out/benchmark-results/${SESSION_ID} | grep cstat.${target} | grep nonautopause)
    else
      FILES=$(ls out/benchmark-results/${SESSION_ID} | grep cstat.${target})
    fi

    # calc average per test target
    for file in ${FILES[@]}; do
      NAP_MEASURED=$(cat out/benchmark-results/${SESSION_ID}/${file} | tail -n 1) 
      nap_total=$(echo ${nap_total}+${NAP_MEASURED} | bc )
      ((nap_count++))
    done

    NONAUTOPAUSE_RESULT=$(echo "scale=4; ${nap_total} / ${nap_count}" | bc | awk '{printf "%.4f\n", $0}')
    NONAUTOPAUSE_RESULTS=("${NONAUTOPAUSE_RESULTS[@]}" ${NONAUTOPAUSE_RESULT})
  done
}

# calc average each autopause test target
calcAverageAutopause() {
  for target in ${TESTS_TARGETS[@]}; do
    if [[ "${target}" == "minikube."* ]]; then
      ap_count=0;
      ap_total=0;
      FILES=$(ls out/benchmark-results/${SESSION_ID} | grep cstat.${target} | grep autopause)
      # calc average per test target
      for file in ${FILES[@]}; do
        AP_MEASURED=$(cat out/benchmark-results/${SESSION_ID}/${file} | tail -n 1)
        ap_total=$(echo ${ap_total}+${AP_MEASURED} | bc )
        ((ap_count++))
      done

      AUTOPAUSE_RESULT=$(echo "scale=4; ${ap_total} / ${ap_count}" | bc | awk '{printf "%.4f\n", $0}')
      AUTOPAUSE_RESULTS=("${AUTOPAUSE_RESULTS[@]}" ${AUTOPAUSE_RESULT})
    else
      AUTOPAUSE_RESULTS=("${AUTOPAUSE_RESULTS[@]}" 0)
    fi
  done
}

# create non-autopause summary csv
updateNonAutopauseSummary() {
  for ((i = 0; i < ${#NONAUTOPAUSE_RESULTS[@]}; i++)) {
    echo "${NONAUTOPAUSE_RESULTS[i]}" >> out/benchmark-results/${SESSION_ID}/cstat.nonautopause.summary
  }
}

# create autopause summary csv
updateAutopauseSummary() {
  for ((i = 0; i < ${#AUTOPAUSE_RESULTS[@]}; i++)) {
    echo "${AUTOPAUSE_RESULTS[i]}" >> out/benchmark-results/${SESSION_ID}/cstat.autopause.summary
  }
}

calcAverageNonAutopause
updateNonAutopauseSummary

calcAverageAutopause
updateAutopauseSummary
