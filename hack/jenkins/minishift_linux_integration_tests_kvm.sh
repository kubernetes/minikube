#!/bin/bash

# Copyright 2017 The Kubernetes Authors All rights reserved.
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


# This script runs the integration tests on a Linux machine for the KVM Driver

# The script expects the following env variables:
# MINIKUBE_LOCATION: GIT_COMMIT from upstream build.
# COMMIT: Actual commit ID from upstream build
# EXTRA_BUILD_ARGS (optional): Extra args to be passed into the minikube integrations tests
# access_token: The Github API access token. Injected by the Jenkins credential provider.

set -e

OS_ARCH="linux-amd64"
VM_DRIVER="kvm"
JOB_NAME="Minishift-Linux-KVM"
BINARY=$(pwd)/out/minishift
EXTRA_FLAGS="--show-libmachine-logs"

# Download files and set permissions
mkdir -p out/

curl -L http://artifacts.ci.centos.org/minishift/minishift/master/latest/linux-amd64/minishift -o out/minishift
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/docker-machine-driver-* out/ || true
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/minikube-testing.iso out/ || true

# Set the executable bit on the minishift and driver binary
chmod +x out/minishift
chmod +x out/docker-machine-driver-* || true

# Get version information
./out/minishift version

# Remove pre-exist kube config files
rm -fr $HOME/.kube || true

ISO=https://storage.googleapis.com/minikube/iso/minikube-v0.24.0.iso
if [ -f $(pwd)/out/minikube-testing.iso ]; then
    ISO=file://$(pwd)/out/minikube-testing.iso
fi

export MINIKUBE_WANTREPORTERRORPROMPT=False
./out/minishift delete --force || true

# Add the out/ directory to the PATH, for using new drivers.
export PATH="$(pwd)/out/":$PATH

# Linux cleanup
virsh -c qemu:///system list --all \
      | sed -n '3,$ p' \
      | cut -d' ' -f 7 \
      | xargs -I {} sh -c "virsh -c qemu:///system destroy {}; virsh -c qemu:///system undefine {}"  \
|| true

sudo virsh net-define /usr/share/libvirt/networks/default.xml || true
sudo virsh net-start default || true
sudo virsh net-list

# see what driver we are using
which docker-machine-driver-${VM_DRIVER} || true

result=$?
echo $result

function print_success_message() {
  echo ""
  echo " ------------ [ $1 - Passed ]"
  echo ""
}

function exit_with_message() {
  if [[ "$1" != 0 ]]; then
    echo "$2"
    result=1
  fi
}

function assert_equal() {
  if [ "$1" != "$2" ]; then
    echo "Expected '$1' equal to '$2'"
    result=1
  fi
}

# http://www.linuxjournal.com/content/validating-ip-address-bash-script
function assert_valid_ip() {
  local ip=$1
  local valid=1

  if [[ $ip =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
    OIFS=$IFS
    IFS='.'
    ip=($ip)
    IFS=$OIFS
    [[ ${ip[0]} -le 255 && ${ip[1]} -le 255 \
    && ${ip[2]} -le 255 && ${ip[3]} -le 255 ]]
    valid=$?
  fi

  if [[ "$valid" != 0 ]]; then
    echo "IP '$1' is invalid"
    result=1
  fi
 }

function verify_start_instance() {
  $BINARY start --iso-url $ISO $EXTRA_FLAGS
  exit_with_message "$?" "Error starting Minishift VM"
  output=`$BINARY status | sed -n 1p`
  assert_equal "$output" "Minishift:  Running"
  print_success_message "Starting VM"
}

function verify_stop_instance() {
  $BINARY stop
  exit_with_message "$?" "Error starting Minishift VM"
  output=`$BINARY status | sed -n 1p`
  assert_equal "$output" "Minishift:  Stopped"
  print_success_message "Stopping VM"
}

function verify_swap_space() {
  output=`$BINARY ssh -- free | tail -n 1 | awk '{print $2}'`
  if [ "$output" == "0" ]; then
    echo "Expected '$output' not equal to 0"
    result=1
  fi
  print_success_message "Swap space check"
}

function verify_ssh_connection() {
  output=`$BINARY ssh -- echo hello`
  assert_equal $output "hello"
  print_success_message "SSH Connection"
}

function verify_vm_ip() {
  output=`$BINARY ip`
  assert_valid_ip $output
  print_success_message "Getting VM IP"
}

function verify_cifs_installation() {
  expected="mount.cifs version: 6.6"
  output=`$BINARY ssh -- sudo /sbin/mount.cifs -V`
  assert_equal "$output" "$expected"
  print_success_message "CIFS installation"
}

function verify_nfs_installation() {
  expected="mount.nfs: (linux nfs-utils 1.3.3)"
  output=`$BINARY ssh -- sudo /sbin/mount.nfs -V /need/for/version`
  assert_equal "$output" "$expected"
  print_success_message "NFS installation"
}

function verify_bind_mount() {
  output=`$BINARY ssh -- 'findmnt | grep "\[/var/lib/" | wc -l'`
  assert_equal $output "11"
  print_success_message "Bind mount check"
}

function verify_delete() {
  $BINARY delete --force
  exit_with_message "$?" "Error deleting Minishift VM"
  output=`$BINARY status`
  if [ "$1" != "$2" ]; then
    echo "Expected '$1' equal to '$2'"
    result=1
  fi
  print_success_message "Deleting VM"
}

# Tests
set +e
verify_start_instance
verify_vm_ip
verify_cifs_installation
verify_nfs_installation
verify_bind_mount
verify_swap_space
verify_delete
set -e

if [[ $result -eq 0 ]]; then
  status="success"
else
  status="failure"
fi

set +x
target_url="https://storage.googleapis.com/minikube-builds/logs/${MINIKUBE_LOCATION}/${JOB_NAME}.txt"
curl "https://api.github.com/repos/kubernetes/minikube/statuses/${COMMIT}?access_token=$access_token" \
  -H "Content-Type: application/json" \
  -X POST \
  -d "{\"state\": \"$status\", \"description\": \"Jenkins\", \"target_url\": \"$target_url\", \"context\": \"${JOB_NAME}\"}"
set -x

exit $result
