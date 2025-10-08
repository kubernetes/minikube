#!/usr/bin/env bash
OS=$(go env GOOS)
ARCH=$(go env GOARCH)
REPO_ROOT="$(git rev-parse --show-toplevel)"
make # build the binary first
LATEST_RELEASE=$(curl -sSfL https://dl.k8s.io/release/stable.txt)
"${REPO_ROOT}"/out/minikube start --nodes=2 --driver=docker --cpus=no-limit --memory=no-limit --force --kubernetes-version=$LATEST_RELEASE

kubetest2-tester-ginkgo --test-package-marker stable.txt \
    --parallel=30 \
    --skip-regex='\[Driver:.gcepd\]|\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]'
