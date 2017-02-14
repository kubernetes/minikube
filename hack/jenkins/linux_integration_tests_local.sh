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


# This script runs the integration tests on a Linux machine for the Local Driver

# The script expects the following env variables:
# MINIKUBE_LOCATION: GIT_COMMIT from upstream build.
# COMMIT: Actual commit ID from upstream build
# EXTRA_BUILD_ARGS (optional): Extra args to be passed into the minikube integrations tests
# access_token: The Github API access token. Injected by the Jenkins credential provider. 

set -e

OS_ARCH="linux-amd64"
VM_DRIVER="local"
JOB_NAME="Linux-Local"

# Debug information
# print jenkins environment info
env
# see if the keys needed for local driver are present in the environment
ls -altr ~/.ssh/

ssh -tt -o BatchMode=yes -o StrictHostKeyChecking=no -o CheckHostIP=no -o UserKnownHostsFile=/dev/null 127.0.0.1 <<EOF
    echo "`whoami` ALL=(ALL) NOPASSWD: ALL" | sudo tee -a /etc/sudoers
EOF

# test if a ssh actually works, if this does not work local driver
# wont work either as it depends on using ~/.ssh/id_rsa to ssh into localhost
# and needs authorized_keys to be setup correctly as well.
ssh -o BatchMode=yes -o StrictHostKeyChecking=no -o CheckHostIP=no -o UserKnownHostsFile=/dev/null 127.0.0.1 ls -altr ~

# Download files and set permissions
source common.sh
