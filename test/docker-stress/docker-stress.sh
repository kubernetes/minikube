#!/bin/sh
version="v1.11.0"
old="/tmp/minikube-${version}"
curl -L -C - -o "${old}" https://storage.googleapis.com/minikube/releases/${version}/minikube-darwin-amd64
chmod 755 "${old}"

git fetch
git pull
make

for i in $(seq 1 20); do
  echo "LOOP ${i}: ${version} to HEAD"

  ./out/minikube delete


  logfile=$(mktemp)

  # upgrade hot
  ${old} start --driver=docker || { echo "${old} failed, that's OK though"; }
  time ./out/minikube start --driver=docker --alsologtostderr -v=1 2>${logfile} || { echo "fail hot upgrade (loop $i): see ${logfile}"; docker logs minikube; exit 1; }
  ./out/minikube delete

  # upgrade cold
  ${old} start --driver=docker || { echo "${old} failed, that's OK though"; }
  ${old} stop
  time ./out/minikube start --driver=docker --alsologtostderr -v=1 2>{$logfile} || { echo "fail cold upgrade (loop $i): see ${logfile}"; docker logs minikube; exit 2; }

  # restart hot
  time ./out/minikube start --driver=docker --alsologtostderr -v=1 2>{$logfile} || { echo "fail hot restart (loop $i): see ${logfile}"; docker logs minikube; exit 2; }

  # restart cold
  ./out/minikube stop
  ./out/minikube start --driver=docker --alsologtostderr -v=1 2>{$logfile} || { echo "fail cold restart (loop $i): see ${logfile}"; docker logs minikube; exit 2; }
  ./out/minikube delete
done

