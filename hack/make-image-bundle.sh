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


images=(
    "gcr.io/google_containers/k8s-dns-sidecar-amd64:1.14.4"
    "gcr.io/google_containers/k8s-dns-kube-dns-amd64:1.14.4"
    "gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.4"
    "gcr.io/google-containers/kube-addon-manager:v6.4-beta.2"
    "gcr.io/google_containers/kubernetes-dashboard-amd64:v1.6.1"
    "gcr.io/google_containers/pause-amd64:3.0"
)

for image in "${images[@]}"
do
    # Image name cannot have forward slash (directory) or colon (for make)
    ${PULLER} --name "$image" --tarball ${TARBALL_LOCATION}/$(basename $image | tr : _).tar
done
