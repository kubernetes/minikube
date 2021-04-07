#!/bin/bash

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

set -x -o pipefail
# Only run this on PRs
if [[ "${MINIKUBE_LOCATION}" == "master" ]]; then
	exit 0
fi

# Make sure docker is installed and configured                                        
#./installers/check_install_docker.sh

# Make sure gh is installed and configured                                            
./installers/check_install_gh.sh 

# Make sure go is installed and configured                                            
./installers/check_install_golang.sh "1.16" "/usr/local" || true

# Grab latest code                                                                    
git clone https://github.com/kubernetes/minikube.git                                  
cd minikube                                                                           

# Build minikube binary and mkcmp binary                                              
make out/minikube out/mkcmp                                                           

# Make sure there aren't any old minikube clusters laying around
out/minikube delete --all

# Run mkcmp                                                                           
out/mkcmp out/minikube pr://${MINIKUBE_LOCATION} | tee mkcmp.log
if [ $? -gt 0 ]; then
       # Comment that mkcmp failed
       gh pr comment ${MINIKUBE_LOCATION} --body "timing minikube failed, please try again"
       exit 1
fi
output=$(cat mkcmp.log)
gh pr comment ${MINIKUBE_LOCATION} --body "${output}"
