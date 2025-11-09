#!/bin/bash

# Copyright 2025 The Kubernetes Authors All rights reserved.
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

set -e
set -x

# when docker is the driver, we run integration tests directly in prow cluster
# by default, prow jobs run in root, so we must switch to a non-root user to run docker driver
NEW_USER="minikube"
TARGET_SCRIPT=$1

if [ "$(whoami)" == "root" ]; then
    useradd -m -s /bin/bash "$NEW_USER"
fi
chown -R "$NEW_USER":"$NEW_USER" .
# install sudo if not present
apt-get update && apt-get install -y sudo
# give the new user passwordless sudo
echo "$NEW_USER ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/$NEW_USER"
chmod 440 "/etc/sudoers.d/$NEW_USER"
# add the new user to the docker group
usermod -aG docker "$NEW_USER"
# exec the target script as the new user
su "$NEW_USER" -c "$TARGET_SCRIPT"