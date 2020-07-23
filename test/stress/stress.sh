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

curl -L -C - -o "${OLD_PATH}" https://storage.googleapis.com/minikube/releases/${UPGRADE_FROM}/minikube-darwin-amd64
chmod 755 "${OLD_PATH}"

git fetch || { echo "failed to fetch"; exit 1; }
git pull || { echo "failed to pull"; exit 2; }
make || { echo "failed to run make"; exit 3; }

for i in $(seq 1 20); do
  echo ""
  echo "LOOP ${i}: ${UPGRADE_FROM} to HEAD -- logging to ${LOG_PATH}"

  "${NEW_PATH}" delete

  echo ""
  echo "Upgrade $version to HEAD hot test: loop ${i}"
  ${OLD_PATH} start ${START_FLAGS} || { echo "${OLD_PATH} failed, that's OK though"; }
  time ${NEW_PATH} start ${START_FLAGS} --alsologtostderr -v=2 2>${logfile} || { docker logs minikube | tee -a ${logfile}; echo "fail hot upgrade (loop $i): see ${logfile}";  exit 1; }
  ${NEW_PATH} delete

  echo ""
  echo "Upgrade $version to HEAD cold test: loop ${i}"
  ${OLD_PATH} start ${START_FLAGS} || { echo "${OLD_PATH} failed, that's OK though"; }
  ${OLD_PATH} stop
  time ${NEW_PATH} start ${START_FLAGS} --alsologtostderr -v=2 2>${logfile} || { docker logs minikube | tee -a ${logfile}; echo "fail cold upgrade (loop $i): see ${logfile}"; exit 2; }

  echo ""
  echo "Restart HEAD hot test: loop ${i}"
  time ${NEW_PATH} start ${START_FLAGS} --alsologtostderr -v=2 2>${logfile} || { docker logs minikube | tee -a ${logfile}; echo "fail hot HEAD restart (loop $i): see ${logfile}"; exit 3; }

  echo ""
  echo "Restart HEAD cold test: loop ${i}"
  ${NEW_PATH} stop
  time ${NEW_PATH} start ${START_FLAGS} --alsologtostderr -v=2 2>${logfile} || { docker logs minikube | tee -a ${logfile}; echo "fail cold HEAD restart (loop $i): see ${logfile}"; exit 4; }
  ${NEW_PATH} delete

  echo ""
  echo "****************************************************"
  echo "Congratulations - you survived loop ${i} of ${TOTAL}"
  echo "****************************************************"
  echo ""
done

