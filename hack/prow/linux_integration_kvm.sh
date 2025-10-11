#!/bin/bash

set -e
set -x

OS="linux"
ARCH="amd64"
DRIVER="kvm2"
CONTAINER_RUNTIME="docker"
# in prow, if you want libvirtd to be run, you have to start a privileged container as root
EXTRA_START_ARGS="" 
EXTRA_TEST_ARGS="-gvisor" # We pick kvm as our gvisor testbed because it is fast & reliable
JOB_NAME="KVM_Linux"

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
