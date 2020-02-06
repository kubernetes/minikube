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

set -ex 

PROFILE=generate-preloaded-images-tar
KUBERNETES_VERSION=${KUBERNETES_VERSION:-""}
TARBALL_FILENAME=preloaded-images-k8s-$KUBERNETES_VERSION.tar

function delete_minikube {
    out/minikube delete --profile=$PROFILE
}

trap "delete_minikube" ERR

out/minikube start --memory=10000 --profile=$PROFILE --kubernetes-version=$KUBERNETES_VERSION
out/minikube ssh --profile=$PROFILE -- sudo tar cvf $TARBALL_FILENAME /var/lib/docker 
scp -o StrictHostKeyChecking=no  -i $(out/minikube ssh-key --profile=$PROFILE) docker@$(out/minikube ip --profile=$PROFILE):/home/docker/$TARBALL_FILENAME out/$TARBALL_FILENAME
delete_minikube
