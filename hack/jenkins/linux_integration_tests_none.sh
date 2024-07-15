#!/bin/bash

# Copyright 2016 The Kubernetes Authors All rights reserved.
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


# This script runs the integration tests on a Linux machine for the none Driver

# The script expects the following env variables:
# MINIKUBE_LOCATION: GIT_COMMIT from upstream build.
# COMMIT: Actual commit ID from upstream build
# EXTRA_BUILD_ARGS (optional): Extra args to be passed into the minikube integrations tests
# access_token: The GitHub API access token. Injected by the Jenkins credential provider.

set -e

OS="linux"
ARCH="amd64"
DRIVER="none"
JOB_NAME="none_Linux"
EXTRA_START_ARGS="--bootstrapper=kubeadm"

# add install location of iptables and conntrack to PATH
PATH="/usr/sbin:$PATH"

if ! kubeadm &>/dev/null; then
  echo "WARNING: kubeadm is not installed. will try to install."
  curl -LO "https://dl.k8s.io/release/$(curl -sSL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubeadm"
  sudo install kubeadm /usr/local/bin/kubeadm
fi
# "none" driver specific cleanup from previous runs.
sudo kubeadm reset -f --cri-socket unix:///var/run/cri-dockerd.sock || true
# kubeadm reset may not stop pods immediately
docker rm -f $(docker ps -aq) >/dev/null 2>&1 || true

# Cleanup data directory
sudo rm -rf /data/*
# Cleanup old Kubernetes configs
sudo rm -rf /etc/kubernetes/*
sudo rm -rf /var/lib/minikube/* 

# Stop any leftover kubelet
sudo systemctl is-active --quiet kubelet \
  && echo "stopping kubelet" \
  && sudo systemctl stop -f kubelet

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

# cri-dockerd is required for Kubernetes v1.24+ with none driver
CRI_DOCKERD_VERSION="0.3.15"
if [[ $(cri-dockerd --version 2>&1) != *"$CRI_DOCKERD_VERSION"* ]]; then
  echo "WARNING: expected version of cri-dockerd is not installed. will try to install."
  sudo systemctl stop cri-docker.socket || true
  sudo systemctl stop cri-docker.service || true
  CRI_DOCKERD_COMMIT="c1c566e0cc84abe6972f0bf857ecd8fe306258d9"
  CRI_DOCKERD_BASE_URL="https://storage.googleapis.com/kicbase-artifacts/cri-dockerd/${CRI_DOCKERD_COMMIT}"
  sudo curl -L "${CRI_DOCKERD_BASE_URL}/amd64/cri-dockerd" -o /usr/bin/cri-dockerd
  sudo curl -L "${CRI_DOCKERD_BASE_URL}/cri-docker.socket" -o /usr/lib/systemd/system/cri-docker.socket
  sudo curl -L "${CRI_DOCKERD_BASE_URL}/cri-docker.service" -o /usr/lib/systemd/system/cri-docker.service
  sudo chmod +x /usr/bin/cri-dockerd
  sudo systemctl daemon-reload
  sudo systemctl enable cri-docker.service
  sudo systemctl enable --now cri-docker.socket
fi

# crictl is required for Kubernetes v1.24+ with none driver
CRICTL_VERSION="v1.28.0"
if [[ $(crictl --version) != *"$CRICTL_VERSION"* ]]; then
  echo "WARNING: expected version of crictl is not installed. will try to install."
  curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/$CRICTL_VERSION/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz --output crictl-${CRICTL_VERSION}-linux-amd64.tar.gz
  sudo tar zxvf crictl-$CRICTL_VERSION-linux-amd64.tar.gz -C /usr/local/bin
fi

# cni-plugins is required for Kubernetes v1.24+ with none driver
./installers/check_install_cni_plugins.sh

# We need this for reasons now
sudo sysctl fs.protected_regular=0

source ./common.sh
