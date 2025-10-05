#!/bin/bash

set -e
set -x

OS="linux"
ARCH="amd64"
DRIVER="kvm2"
CONTAINER_RUNTIME="docker"
# in prow, if you want libvirtd to be run, you have to start a privileged container as root
EXTRA_START_ARGS="--force" 
EXTRA_TEST_ARGS="-gvisor" # We pick kvm as our gvisor testbed because it is fast & reliable
JOB_NAME="KVM_Linux"

sudo apt-get update
sudo apt-get -y install qemu-system libvirt-clients libvirt-daemon-system ebtables iptables dnsmasq
sudo adduser $(whoami) libvirt || true

# start libvirtd 
sudo systemctl start libvirtd
sleep 5  # wait for libvirtd to be running
echo "=========libvirtd status=========="
sudo systemctl status libvirtd

source ./hack/prow/common.sh 


