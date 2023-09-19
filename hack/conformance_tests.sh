#!/bin/bash

# Copyright 2019 The Kubernetes Authors All rights reserved.
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

# This script executes the Kubernetes conformance tests in accordance with:
# https://github.com/cncf/k8s-conformance/blob/master/instructions.md
#
# Usage:
#   conformance_tests.sh <path to minikube> <flags>
#
# Example:
#   conformance_tests.sh ./out/minikube --driver=hyperkit
set -ex -o pipefail

readonly PROFILE_NAME="k8sconformance"
readonly MINIKUBE=${1:-./out/minikube}
shift || true
readonly START_ARGS=${@:-}

# Requires a fully running Kubernetes cluster.
"${MINIKUBE}" delete -p "${PROFILE_NAME}" || true
"${MINIKUBE}" start -p "${PROFILE_NAME}" ${START_ARGS} --wait=all --nodes=2
kubectl --context "${PROFILE_NAME}" get pods --all-namespaces
"${MINIKUBE}" status -p "${PROFILE_NAME}"

# Make sure jq is installed
sudo apt-get install jq -y

# Remove old sonobuoy installation
rm -rf sonobuoy

# Get latest sonobuoy version
sonobuoy=$(curl -s https://api.github.com/repos/vmware-tanzu/sonobuoy/releases/latest | jq .assets[].browser_download_url | grep linux_amd64 | cut -d '"' -f 2)
curl -LO $sonobuoy
tarball=$(echo $sonobuoy | awk -F "/" '{print $(NF)}')
tar -xzf $tarball

./sonobuoy run --plugin-env=e2e.E2E_EXTRA_ARGS="--ginkgo.v" --mode=certified-conformance --wait --alsologtostderr
outdir="$(mktemp -d)"
./sonobuoy retrieve "${outdir}"

"${MINIKUBE}" delete -p "${PROFILE_NAME}"

cwd=$(pwd)

cd "${outdir}"
mkdir ./results; tar xzf *.tar.gz -C ./results

version=$(${cwd}/${MINIKUBE} version  | cut -d" " -f3)

mkdir -p "minikube-${version}"
cd "minikube-${version}"

cat <<EOF >PRODUCT.yaml
vendor: minikube
name: minikube
version: ${version}
website_url: https://github.com/kubernetes/minikube
repo_url: https://github.com/kubernetes/minikube
documentation_url: https://minikube.sigs.k8s.io/docs/
product_logo_url: https://raw.githubusercontent.com/kubernetes/minikube/master/images/logo/logo.svg
type: installer
description: minikube runs a local Kubernetes cluster on macOS, Linux, and Windows.
contact_email_address: minikube-dev@googlegroups.com
EOF

cat <<"EOF" >README.md
# Reproducing the test results

## Run minikube with docker driver

Install [docker](https://docs.docker.com/engine/install/)
Install [kubectl](https://v1-18.docs.kubernetes.io/docs/tasks/tools/install-kubectl/)
Clone the [minikube repo](https://github.com/kubernetes/minikube)

## Compile the latest minikube binary
```console
% cd <minikube dir>
% make
```

## Trigger the tests and get back the results

We follow the [official instructions](https://github.com/cncf/k8s-conformance/blob/master/instructions.md):

```console
% cd <minikube dir>
./hack/conformance_tests.sh ${MINIKUBE} ${START_ARGS}
```

This script will run sonobuoy against a minikube cluster with two nodes and the provided parameters.
EOF

cp -r ../results/plugins/e2e/results/global/* .
cd ..
cp -r "minikube-${version}" "${cwd}"
