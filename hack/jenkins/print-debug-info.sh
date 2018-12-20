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

# Prints some debug info about the state of the cluster
#
# We don't check the error on these commands, since they might fail depending on
# the cluster state.
set +e

echo ""
echo ">>> print-debug-info at $(date):"
echo ""
${SUDO_PREFIX} cat "${KUBECONFIG}"
kubectl version \
    && kubectl get pods --all-namespaces \
    && kubectl cluster-info dump

# minikube has probably been shut down, so iterate forward each command rather than spamming.
MINIKUBE=${SUDO_PREFIX}out/minikube-${OS_ARCH}
${MINIKUBE} status
${MINIKUBE} ip && ${MINIKUBE} logs

if [[ "${VM_DRIVER}" == "none" ]]; then
  run=""
else
  run="${MINIKUBE} ssh --"
fi

echo "Local date: $(date)"
${run} date
${run} uptime
${run} docker ps
${run} env TERM=dumb systemctl list-units --state=failed
${run} env TERM=dumb journalctl --no-tail --no-pager -p notice
${run} free
${run} cat /etc/VERSION

if type -P virsh; then
  virsh -c qemu:///system list --all
fi

if type -P vboxmanage; then
  vboxmanage list vms
fi

if type -P hdiutil; then
  hdiutil info | grep -E "/dev/disk[1-9][^s]"
fi

netstat -rn -f inet

echo ""
echo ">>> end print-debug-info"
echo ""
set -e
