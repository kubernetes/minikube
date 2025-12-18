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
DRIVER="kvm2"
CONTAINER_RUNTIME="containerd"
# in prow, if you want libvirtd to be run, you have to start a privileged container as root
EXTRA_START_ARGS="" 
EXTRA_TEST_ARGS=""
JOB_NAME="KVM_Containerd_Linux_x86"
#  marking all directories ('*') as trusted, since .git belongs to root, not minikube user
git config --global --add safe.directory '*'
COMMIT=$(git rev-parse HEAD)
MINIKUBE_LOCATION=$COMMIT
echo "running test in $(pwd)"

set +e
sleep 5  # wait for libvirtd to be running
echo "=========libvirtd status=========="
sudo systemctl status libvirtd
echo "=========Check virtualization support=========="
grep -E -q 'vmx|svm' /proc/cpuinfo && echo yes || echo no 
echo "=========virt-host-validate=========="
virt-host-validate

set -e
source ./hack/prow/common.sh 
