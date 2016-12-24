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


# Copy only the files we need to this workspace
mkdir -p out/ testdata/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/minikube-${OS_ARCH} out/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/e2e-${OS_ARCH} out/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/testdata/busybox.yaml testdata/

# Set the executable bit on the e2e binary and out binary
chmod +x out/e2e-${OS_ARCH}
chmod +x out/minikube-${OS_ARCH}
