#!/bin/sh

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

# Periodically cleanup and reboot if no Jenkins subprocesses are running.
set -uf -o pipefail

PATH=/usr/local/bin:/sbin:/usr/local/sbin:$PATH

exit_if_jenkins() {
  jenkins=$(pgrep java)
  if [[ "$jenkins" -- "" ]]; then
    echo "no java, no jenkins"
    return 0
  fi
  pstree $jenkins | grep -v java && echo "jenkins is running..." && exit 1
}

exit_if_jenkins
echo "waiting to see if any jobs are coming in..."
sleep 15
exit_if_jenkins
echo "doing it"
killall java
sudo rm -Rf ~jenkins/.minikube || echo "could not delete minikube"
sudo rm -Rf ~/jenkins/minikube-integration/* || true
sudo reboot
