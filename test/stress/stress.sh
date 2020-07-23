#!/bin/sh
#
# Stress test for start, restart, upgrade.
# 
# Usage:
#
#   ./test/stress/stress.sh "<optional start flags>" <optional version to upgrade from> <optional count>
#
# Example:
#
#   ./test/stress/stress.sh "--driver=docker --base-image gcr.io/k8s-minikube/kic:ubuntu-upgrade"
#

readonly START_FLAGS=${1:-""}
readonly UPGRADE_FROM=${2:-"v1.11.0"}
readonly TOTAL=${3:-20}
readonly OLD_PATH="/tmp/minikube-${UPGRADE_FROM}"
readonly NEW_PATH="./out/minikube"
readonly LOG_PATH="$(mktemp)"

readonly PROFILE="stress$(echo -n $START_FLAGS | openssl md5 | cut -c1-5)"

if [[ ! -x "${NEW_PATH}" ]]; then
  echo "${NEW_PATH} is missing, please run 'make'"
  exit 4
fi

echo "Downloading minikube ${UPGRADE_FROM} for upgrade test ..."
curl -L -C - -o "${OLD_PATH}" https://storage.googleapis.com/minikube/releases/${UPGRADE_FROM}/minikube-darwin-amd64
chmod 755 "${OLD_PATH}"

for i in $(seq 1 ${TOTAL}); do
  echo ""
  echo "*** LOOP ${i} of ${TOTAL}: ${UPGRADE_FROM} to HEAD, logging to ${LOG_PATH}"
  echo "***   start flags: -p ${PROFILE} ${START_FLAGS}"
  echo ""
  "${NEW_PATH}" delete -p ${PROFILE}

  echo ""
  echo "Upgrade ${UPGRADE_FROM} to HEAD hot test: loop ${i}"
  ${OLD_PATH} start -p ${PROFILE} ${START_FLAGS} || { echo "${OLD_PATH} failed, that's OK though"; }

  time ${NEW_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr -v=2 2>${LOG_PATH} \
    || { docker logs ${PROFILE} | tee -a ${LOG_PATH}; echo "fail hot upgrade (loop $i): see ${LOG_PATH}";  exit 1; }

  ${NEW_PATH} delete -p ${PROFILE}

  echo ""
  echo "Upgrade ${UPGRADE_FROM} to HEAD cold test: loop ${i}"
  ${OLD_PATH} start -p ${PROFILE} ${START_FLAGS} || { echo "${OLD_PATH} failed, that's OK though"; }

  ${OLD_PATH} stop -p ${PROFILE}

  time ${NEW_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr -v=2 2>${LOG_PATH} \
    || { docker logs ${PROFILE} | tee -a ${LOG_PATH}; echo "fail cold upgrade (loop $i): see ${LOG_PATH}"; exit 2; }

  echo ""
  echo "Restart HEAD hot test: loop ${i}"
  time ${NEW_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr -v=2 2>${LOG_PATH} \
    || { docker logs ${PROFILE} | tee -a ${LOG_PATH}; echo "fail hot HEAD restart (loop $i): see ${LOG_PATH}"; exit 3; }

  echo ""
  echo "Restart HEAD cold test: loop ${i}"
  ${NEW_PATH} stop

  time ${NEW_PATH} start -p ${PROFILE} ${START_FLAGS} --alsologtostderr -v=2 2>${LOG_PATH} \
    || { docker logs ${PROFILE} | tee -a ${LOG_PATH}; echo "fail cold HEAD restart (loop $i): see ${LOG_PATH}"; exit 4; }

  ${NEW_PATH} delete -p ${PROFILE}

  echo ""
  echo "****************************************************"
  echo "Congratulations - ${PROFILE} survived loop ${i} of ${TOTAL}"
  echo "****************************************************"
  echo ""
done

