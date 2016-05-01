#!/bin/bash

REPO_PATH="github.com/kubernetes/minikube"

export GOPATH=${PWD}/gopath
export GO15VENDOREXPERIMENT=1
export OS=${OS:-$(go env GOOS)}
export ARCH=${ARCH:-$(go env GOARCH)}

rm -f ${GOPATH}/src/${REPO_PATH}
mkdir -p $(dirname ${GOPATH}/src/${REPO_PATH})
ln -s ${PWD} $GOPATH/src/${REPO_PATH}

CGO_ENABLED=0 GOARCH=${ARCH} GOOS=${OS} go build --installsuffix cgo -a -o minikube cli/main.go
