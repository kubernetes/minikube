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


# This script runs the integration tests on a Linux machine for the KVM Driver

# The script expects the following env variables:
# MINIKUBE_LOCATION: GIT_COMMIT from upstream build.
# COMMIT: Actual commit ID from upstream build
# EXTRA_BUILD_ARGS (optional): Extra args to be passed into the minikube integrations tests
# access_token: The Github API access token. Injected by the Jenkins credential provider. 

set -e

OS_ARCH="linux-amd64"
VM_DRIVER="kvm"
JOB_NAME="Linux-KVM-Alt"

# Use a checksum here instead of a released version
# until the driver is stable
SHA="ba50f204ccf62c2f1f521fd60aa8eb68c39bcd2f"

pushd "$GOPATH/src/github.com/r2d4/docker-machine-driver-kvm" >/dev/null
    git fetch origin
    git checkout ${SHA}

    # Make will run `go install` and put the binary in $GOBIN
    make
popd >/dev/null

echo "Using driver $(which docker-machine-driver-kvm)"

# Download files and set permissions
source common.sh

# Clean up the driver, so that we use the other KVM driver by default.
rm $GOPATH/bin/docker-machine-driver-kvm
