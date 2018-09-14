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

env
${SUDO_PREFIX} cat $KUBECONFIG

kubectl get pods --all-namespaces
kubectl cluster-info dump

cat $HOME/.kube/config
echo $PATH

docker ps

MINIKUBE=${SUDO_PREFIX}out/minikube-${OS_ARCH}
${MINIKUBE} status
${MINIKUBE} ip
${MINIKUBE} ssh -- cat /etc/VERSION
${MINIKUBE} ssh -- docker ps
${MINIKUBE} logs

set -e
