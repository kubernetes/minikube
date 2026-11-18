#!/bin/bash

# Copyright 2026 The Kubernetes Authors All rights reserved.
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
DRIVER="ssh"
CONTAINER_RUNTIME="containerd"
EXTRA_TEST_ARGS=""
JOB_NAME="SSH_Containerd_Linux_x86"
#  marking all directories ('*') as trusted, since .git belongs to root, not minikube  user
git config --global --add safe.directory '*'
COMMIT=$(git rev-parse HEAD)
MINIKUBE_LOCATION=$COMMIT


# conntrack is required for Kubernetes 1.18 and higher
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

# crictl is required for Kubernetes v1.24+
CRICTL_VERSION="v1.35.0"
if [[ $(crictl --version) != *"$CRICTL_VERSION"* ]]; then
  echo "WARNING: expected version of crictl is not installed. will try to install."
  curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/$CRICTL_VERSION/crictl-${CRICTL_VERSION}-linux-amd64.tar.gz --output crictl-${CRICTL_VERSION}-linux-amd64.tar.gz
  sudo tar zxvf crictl-$CRICTL_VERSION-linux-amd64.tar.gz -C /usr/local/bin
fi

# cni-plugins is required for Kubernetes v1.24+
./hack/jenkins/installers/check_install_cni_plugins.sh

# the ssh driver connects back to this machine with a dedicated key
SSH_KEY="$HOME/.ssh/id_minikube_ssh_driver"
mkdir -p "$HOME/.ssh" && chmod 700 "$HOME/.ssh"
if [ ! -f "$SSH_KEY" ]; then
  ssh-keygen -t ed25519 -N '' -f "$SSH_KEY"
  cat "${SSH_KEY}.pub" >> "$HOME/.ssh/authorized_keys"
  chmod 600 "$HOME/.ssh/authorized_keys"
fi
# use the primary IP, pods cannot reach the apiserver on 127.0.0.1
SSH_IP=$(hostname -I | awk '{print $1}')
ssh -i "$SSH_KEY" -o StrictHostKeyChecking=no -o BatchMode=yes "$(whoami)@${SSH_IP}" sudo -n true

EXTRA_START_ARGS="--ssh-ip-address=${SSH_IP} --ssh-user=$(whoami) --ssh-key=${SSH_KEY}"

source ./hack/prow/common.sh
