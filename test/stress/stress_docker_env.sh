#!/bin/bash

# Copyright 2020 The Kubernetes Authors All rights reserved.
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

#
# Stress test for start, restart, upgrade.
# 
# Usage:
#
#   ./test/stress/stress.sh "<optional start flags>" <optional version to upgrade from> <optional count>
#
# Example, testing upgrades from v1.11.0 + fresh starts from HEAD with flags:
#
#   ./test/stress/stress.sh "--driver=docker --base-image gcr.io/k8s-minikube/kic:ubuntu-upgrade"
#
# Example, testing v1.13.0 + HEAD with environment variables, but no flags
#
#   env MINIKUBE_FORCE_SYSTEMD=true MINIKUBE_HOME=/google/minikube ./test/stress/stress.sh "" v1.13.1
#

readonly START_FLAGS=${1:-""}
readonly BIN_PATH="./out/minikube"
readonly LOG_PATH="$(mktemp)"
readonly TOTAL=${2:-4}

if [[ -z "${START_FLAGS}" ]]; then
  PROFILE="stressdenv"
else
  PROFILE="stressdenv$(echo -n $START_FLAGS | openssl md5 | awk '{print $NF}' | cut -c1-5)"
fi

if [[ ! -x "${BIN_PATH}" ]]; then
  echo "${BIN_PATH} is missing, please run 'make'"
  exit 4
fi


fail() {
  local msg=$1

  lecho "** FAILED with ${START_FLAGS}: ${msg} -- docker logs follow: **"
  docker logs "${PROFILE}" | tee -a $LOG_PATH
  lecho "** FULL FAILURE LOGS are available in ${LOG_PATH}"
}

lecho() {
  local msg=$1
  echo "$msg" | tee -a "${LOG_PATH}"
}


${BIN_PATH} delete --all

for i in $(seq 1 ${TOTAL}); do
  lecho ""
  lecho "*** LOOP ${i} of ${TOTAL}: logging to ${LOG_PATH}" 
  lecho "***   start flags: -p ${PROFILE} ${START_FLAGS}"
  lecho ""
  lecho "start $PROFILE"
  PROFILE="${PROFILE}${i}"
  lecho "start"
  time ${BIN_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr 2>>${LOG_PATH} || { fail "minikube start (loop $i)";  exit 1; }
 
  # lecho "docker-env"
  # time ${BIN_PATH} docker-env -p ${PROFILE} --alsologtostderr 2>>${LOG_PATH} || { fail "minikube docker-env (loop $i)";  exit 1; }

  # lecho "status ${i}"
  # time ${BIN_PATH} status -p ${PROFILE}


  # lecho "stop loop ${i}"
  # ${BIN_PATH} stop -p ${PROFILE}

  # lecho "delete loop ${i}"
  #   ${BIN_PATH} delete -p ${PROFILE}

  lecho ""
  lecho "****************************************************"
  lecho "Congratulations - ${PROFILE} survived loop ${i} of ${TOTAL}"
  lecho "****************************************************"
  lecho ""
done
