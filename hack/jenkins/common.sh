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


# This script downloads the test files from the build bucket and makes some executable.

# The script expects the following env variables:
# OS_ARCH: The operating system and the architecture separated by a hyphen '-' (e.g. darwin-amd64, linux-amd64, windows-amd64)
# VM_DRIVER: the vm-driver to use for the test
# EXTRA_BUILD_ARGS: additional flags to pass into minikube start
# JOB_NAME: the name of the logfile and check name to update on github


# Copy only the files we need to this workspace
mkdir -p out/ testdata/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/minikube-${OS_ARCH} out/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/docker-machine-driver-* out/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/e2e-${OS_ARCH} out/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/testdata/busybox.yaml testdata/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/testdata/pvc.yaml testdata/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/testdata/busybox-mount-test.yaml testdata/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/testdata/nginx-pod-svc.yaml testdata/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/testdata/nginx-ing.yaml testdata/

# Set the executable bit on the e2e binary and out binary
chmod +x out/e2e-${OS_ARCH}
chmod +x out/minikube-${OS_ARCH}
chmod +x out/docker-machine-driver-*

# Fix permissions in $HOME
sudo chown -R $USER $HOME/.kube || true
sudo chown -R $USER $HOME/.minikube || true

export MINIKUBE_WANTREPORTERRORPROMPT=False
sudo ./out/minikube-${OS_ARCH} delete || true
./out/minikube-${OS_ARCH} delete || true

# Add the out/ directory to the PATH, for using new drivers.
export PATH="$(pwd)/out/":$PATH

# Linux cleanup
virsh -c qemu:///system list --all \
      | sed -n '3,$ p' \
      | cut -d' ' -f 7 \
      | xargs -I {} sh -c "virsh -c qemu:///system destroy {}; virsh -c qemu:///system undefine {}"  \
      || true

# Clean up virtualbox VMs
vboxmanage list vms \
      | cut -d "{" -f2 \
      | cut -d "}" -f1 \
      | xargs -I {} sh -c "vboxmanage startvm {} --type emergencystop; vboxmanage unregistervm {} --delete" \
      || true

# Clean up xhyve disks
hdiutil info \
      | egrep \/dev\/disk[1-9][^s] \
      | awk '{print $1}' \
      | xargs -I {} sh -c "hdiutil detach {}" \
      || true

# Clean up xhyve processes
pgrep xhyve | xargs kill || true


if [ -e out/docker-machine-driver-hyperkit ]; then
  sudo chown root:wheel out/docker-machine-driver-hyperkit || true
  sudo chmod u+s out/docker-machine-driver-hyperkit || true
fi

# See the default image
./out/minikube-${OS_ARCH} start -h | grep iso

# see what driver we are using
which docker-machine-driver-${VM_DRIVER} || true
find ~/.minikube || true

# Allow this to fail, we'll switch on the return code below.
set +e
${SUDO_PREFIX}out/e2e-${OS_ARCH} -minikube-start-args="--vm-driver=${VM_DRIVER} ${EXTRA_START_ARGS}" -minikube-args="--v=10 --logtostderr ${EXTRA_ARGS}" -test.v -test.timeout=30m -binary=out/minikube-${OS_ARCH}
result=$?
set -e

# See the KUBECONFIG file for debugging
sudo cat $KUBECONFIG

MINIKUBE_WANTREPORTERRORPROMPT=False sudo ./out/minikube-${OS_ARCH} delete \
|| MINIKUBE_WANTREPORTERRORPROMPT=False ./out/minikube-${OS_ARCH} delete \
|| true

if [[ $result -eq 0 ]]; then
  status="success"
else
  status="failure"
  source print-debug-info.sh
fi

set +x
target_url="https://storage.googleapis.com/minikube-builds/logs/${MINIKUBE_LOCATION}/${JOB_NAME}.txt"
curl "https://api.github.com/repos/kubernetes/minikube/statuses/${COMMIT}?access_token=$access_token" \
  -H "Content-Type: application/json" \
  -X POST \
  -d "{\"state\": \"$status\", \"description\": \"Jenkins\", \"target_url\": \"$target_url\", \"context\": \"${JOB_NAME}\"}"
set -x

exit $result
