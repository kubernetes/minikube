#!/bin/bash

# Copyright 2019 The Kubernetes Authors All rights reserved.
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

function check_jenkins() {
  jenkins_pid="$(pidof java)"
  if [[ "${jenkins_pid}" = "" ]]; then
          return
  fi
  pstree "${jenkins_pid}" \
        | grep -v java \
        && echo "jenkins is running at pid ${jenkins_pid} ..." \
        && exit 1
}

check_jenkins
logger "cleanup_and_reboot running - may shutdown in 60 seconds"
wall "cleanup_and_reboot running - may shutdown in 60 seconds"

sleep 60

check_jenkins
logger "cleanup_and_reboot is happening!"

# kill jenkins to avoid an incoming request
killall java

# disable localkube, kubelet
systemctl list-unit-files --state=enabled \
        | grep kube \
        | awk '{ print $1 }' \
        | xargs systemctl disable

# update and reboot
apt update -y \
        && apt upgrade -y \
        && reboot
