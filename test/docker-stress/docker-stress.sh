#!/bin/sh
#
# Docker stress test. Usage:
#
# ./test/docker-stress/docker-stress.sh
#
# Tests:
# - Binary Upgrade (cold, hot)
# - Restart (cold, hot)
# 
# Attempts 20 times.
#
version="v1.11.0"
old="/tmp/minikube-${version}"
curl -L -C - -o "${old}" https://storage.googleapis.com/minikube/releases/${version}/minikube-darwin-amd64
chmod 755 "${old}"

git fetch || { echo "failed to fetch"; exit 1; }
git pull || { echo "failed to pull"; exit 2; }
make || { echo "failed to run make"; exit 3; }

for i in $(seq 1 20); do
  echo ""
  echo "LOOP ${i}: ${version} to HEAD"

  ./out/minikube delete
  logfile=$(mktemp)

  echo "Upgrade $version to HEAD hot test: loop ${i}"
  ${old} start --driver=docker || { echo "${old} failed, that's OK though"; }
  time ./out/minikube start --driver=docker --alsologtostderr -v=1 2>${logfile} || { docker logs minikube; echo "fail hot upgrade (loop $i): see ${logfile}";  exit 1; }
  ./out/minikube delete

  echo "Upgrade $version to HEAD cold test: loop ${i}"
  ${old} start --driver=docker || { echo "${old} failed, that's OK though"; }
  ${old} stop
  time ./out/minikube start --driver=docker --alsologtostderr -v=1 2>${logfile} || { docker logs minikube; echo "fail cold upgrade (loop $i): see ${logfile}"; exit 2; }

  echo "Restart HEAD hot test: loop ${i}"
  time ./out/minikube start --driver=docker --alsologtostderr -v=1 2>${logfile} || { docker logs minikube; echo "fail hot restart (loop $i): see ${logfile}"; exit 3; }

  echo "Restart HEAD cold test: loop ${i}"
  ./out/minikube stop
  ./out/minikube start --driver=docker --alsologtostderr -v=1 2>${logfile} || { docker logs minikube; echo "fail cold restart (loop $i): see ${logfile}"; exit 4; }
  ./out/minikube delete
done

