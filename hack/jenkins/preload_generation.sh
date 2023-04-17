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

# This script builds all the minikube binary for all 3 platforms as well as Windows-installer and .deb
# This is intended to be run on a new release tag in order to build/upload the required files for a release


set -eux -o pipefail

source ./hack/jenkins/installers/check_install_linux_crons.sh

# Make sure the right golang version is installed based on Makefile
./hack/jenkins/installers/check_install_golang.sh /usr/local

make update-kubeadm-constants
make upload-preloaded-images-tar
make clean
