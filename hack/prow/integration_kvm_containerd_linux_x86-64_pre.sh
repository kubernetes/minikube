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
JOB_NAME="KVM_containerd_Linux_X86-64"


sudo apt-get update
sudo apt-get -y install qemu-system qemu-kvm libvirt-clients libvirt-daemon-system ebtables iptables dnsmasq
sudo adduser $(whoami) libvirt || true

# start libvirtd 
sudo systemctl start libvirtd
