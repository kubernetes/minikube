#!/bin/bash

# Copyright 2025 The Kubernetes Authors All rights reserved.
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

set -e
set -x

OS="linux"
ARCH="amd64"
DRIVER="none"
CONTAINER_RUNTIME="docker"
EXTRA_START_ARGS="" 
EXTRA_TEST_ARGS=""
JOB_NAME="None_Docker_Linux_X86_64"
#  marking all directories ('*') as trusted, since .git belongs to root, not minikube  user
git config --global --add safe.directory '*'
COMMIT=$(git rev-parse HEAD)
MINIKUBE_LOCATION=$COMMIT

# conntrack is required for Kubernetes 1.18 and higher for none driver
if ! conntrack --version &>/dev/null; then
  echo "WARNING: conntrack is not installed. will try to install."
  sudo apt-get update -qq
  sudo apt-get -qq -y install conntrack
fi

# socat is required for kubectl port forward which is used in some tests such as validateHelmTillerAddon
if ! which socat &>/dev/null; then
  echo "WARNING: socat is not installed. will try to install."
  sudo apt-get update -qq
  sudo apt-get -qq -y install socat
fi

CRI_DOCKERD_VERSION="0.4.1"
if [[ $(cri-dockerd --version 2>&1 || true) != *"$CRI_DOCKERD_VERSION"* ]]; then
  echo "Installing cri-dockerd…"
  CRI_DOCKERD_COMMIT="55d6e1a1d6f2ee58949e13a0c66afe7d779ac942"
  CRI_DOCKERD_BASE_URL="https://storage.googleapis.com/kicbase-artifacts/cri-dockerd/${CRI_DOCKERD_COMMIT}"

  sudo curl -L "${CRI_DOCKERD_BASE_URL}/amd64/cri-dockerd" -o /usr/bin/cri-dockerd
  sudo chmod +x /usr/bin/cri-dockerd
fi

# If PID 1 is not systemd, don't even try systemctl
if [[ "$(ps -p 1 -o comm=)" != "systemd" ]]; then
  echo "No systemd detected (PID 1 is $(ps -p 1 -o comm=)); starting cri-dockerd directly"

  # Adjust flags to match your setup:
  sudo /usr/bin/cri-dockerd \
    --container-runtime-endpoint=unix:///var/run/docker.sock \
    --cri-dockerd-root-dir=/var/lib/cri-dockerd \
    --network-plugin=cni \
    --socket-path=unix:///var/run/cri-dockerd.sock \
    >/var/log/cri-dockerd.log 2>&1 &

else
  echo "systemd detected, using systemctl"
  sudo systemctl daemon-reload
  sudo systemctl enable cri-docker.service
  sudo systemctl enable --now cri-docker.socket
fi

# crictl is required for Kubernetes v1.24+ with none driver
CRICTL_VERSION="v1.34.0"
if [[ $(crictl --version) != *"$CRICTL_VERSION"* ]]; then
  echo "WARNING: expected version of crictl is not installed. will try to install."
  curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/$CRICTL_VERSION/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz --output crictl-${CRICTL_VERSION}-linux-amd64.tar.gz
  sudo tar zxvf crictl-$CRICTL_VERSION-linux-amd64.tar.gz -C /usr/local/bin
fi

# cni-plugins is required for Kubernetes v1.24+ with none driver
pushd ../jenkins/installers >/dev/null
./check_install_cni_plugins.sh
popd >/dev/null

# by default, prow jobs run in root, so we must switch to a non-root user to run docker driver
source ./hack/prow/common.sh 
