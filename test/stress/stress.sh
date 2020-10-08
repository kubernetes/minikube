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
readonly UPGRADE_FROM=${2:-"v1.11.0"}
readonly TOTAL=${3:-20}
readonly OLD_PATH="./out/minikube-${UPGRADE_FROM}"
readonly NEW_PATH="./out/minikube"
readonly LOG_PATH="$(mktemp)"

if [[ -z "${START_FLAGS}" ]]; then
  readonly PROFILE="stress"
else
  readonly PROFILE="stress$(echo -n $START_FLAGS | openssl md5 | awk '{print $NF}' | cut -c1-5)"
fi

if [[ ! -x "${NEW_PATH}" ]]; then
  echo "${NEW_PATH} is missing, please run 'make'"
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

lecho "Downloading minikube ${UPGRADE_FROM} for upgrade test ..."
curl -L -C - -o "${OLD_PATH}" https://storage.googleapis.com/minikube/releases/${UPGRADE_FROM}/minikube-darwin-amd64
chmod 755 "${OLD_PATH}"


for i in $(seq 1 ${TOTAL}); do
  lecho ""
  lecho "*** LOOP ${i} of ${TOTAL}: ${UPGRADE_FROM} to HEAD, logging to ${LOG_PATH}" 
  lecho "***   start flags: -p ${PROFILE} ${START_FLAGS}"
  lecho ""
  "${NEW_PATH}" delete -p ${PROFILE}

  lecho ""
  lecho "Upgrade ${UPGRADE_FROM} to HEAD hot test: loop ${i}"
  time ${OLD_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr 2>>${LOG_PATH} || { lecho "${OLD_PATH} start -p ${PROFILE} ${START_FLAGS} failed, which is OK"; }
  lecho "Starting cluster built-by ${UPGRADE_FROM} with ${NEW_PATH}"
  time ${NEW_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr 2>>${LOG_PATH} || { fail "hot upgrade (loop $i)";  exit 1; }
  lecho "Deleting ${UPGRADE_FROM} built-cluster"
  ${NEW_PATH} delete -p ${PROFILE}

  lecho ""
  lecho "Upgrade ${UPGRADE_FROM} to HEAD cold test: loop ${i}"
  time ${OLD_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr 2>>${LOG_PATH} || { lecho "${OLD_PATH} start -p ${PROFILE} ${START_FLAGS} failed, which is OK"; }

  lecho "Stopping ${UPGRADE_FROM}} built-cluster"
  ${OLD_PATH} stop -p ${PROFILE} 2>>${LOG_PATH}

  lecho "Starting cluster built-by ${UPGRADE_FROM} with ${NEW_PATH}"
  time ${NEW_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr 2>>${LOG_PATH} || { fail "hot upgrade (loop $i)";  exit 1; }

  lecho ""
  lecho "Restart HEAD hot test: loop ${i}"
  time ${NEW_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr 2>>${LOG_PATH} || { fail "hot HEAD restart (loop $i)"; exit 3; }

  lecho ""
  lecho "Restart HEAD cold test: loop ${i}"
  ${NEW_PATH} stop -p ${PROFILE}

  time ${NEW_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr 2>>${LOG_PATH} || { fail "cold HEAD restart (loop $i)"; exit 4; }

  ${NEW_PATH} delete -p ${PROFILE}

  lecho ""
  lecho "****************************************************"
  lecho "Congratulations - ${PROFILE} survived loop ${i} of ${TOTAL}"
  lecho "****************************************************"
  lecho ""
done
