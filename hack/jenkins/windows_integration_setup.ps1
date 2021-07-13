# Copyright 2021 The Kubernetes Authors All rights reserved.
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

$test_home="$env:HOMEDRIVE$env:HOMEPATH\minikube-integration"
$env:KUBECONFIG="$test_home\kubeconfig"
$env:MINIKUBE_HOME="$test_home\.minikube"

if ($driver -eq "docker") {
  # Remove unused images and containers
  docker system prune --all --force --volumes
  docker ps -aq | ForEach -Process {docker rm -fv $_}
}

# delete in case previous test was unexpectedly ended and teardown wasn't run
rm -r -Force $test_home
mkdir -p $test_home
