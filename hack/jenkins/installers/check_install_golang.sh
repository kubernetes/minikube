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


# This script installs the exact golang version if not already installed
# this script is meant to be used by jenkins build agent and works only linux

set -e

VERSION_TO_INSTALL=${1:-"1.12.8"}  
INSTALL_PATH=${2:-"/usr/local"} 


# installs or updates golang if right version doesn't exists
function check_and_install_golang() {
	if ! go version &> /dev/null
	then
		echo "WARNING: No golang installation found in your enviroment."
    install_golang "$VERSION_TO_INSTALL" "$INSTALL_PATH"
    return
	fi
	
	# golang has been installed and check its version
	if [[ $(go version) =~ (([0-9]+)\.([0-9]+).([0-9]+).([\.0-9]*)) ]]
	then
		HOST_VERSION=${BASH_REMATCH[1]}	
    if [ $HOST_VERSION = $VERSION_TO_INSTALL ]; then
      echo "go version on the host looks good : $HOST_VERSION"
    else
      echo "WARNING: expected go version to be $VERSION_TO_INSTALL but got $HOST_VERSION"
      install_golang "$VERSION_TO_INSTALL" "$INSTALL_PATH"
    fi 	  
  else
  	echo "ERROR: Failed to parse golang version."
		return
	fi
}

# install_golang takes two parameters version and path to install.
function install_golang {
    echo "Installing golang version: $1 on $2"  
    pushd /tmp >/dev/null
    sudo curl -qL -O "https://storage.googleapis.com/golang/go${1}.linux-amd64.tar.gz" &&
    sudo tar xfa go${1}.linux-amd64.tar.gz &&
    sudo rm -rf "${2}/go" &&
    sudo mv go "${2}/" && sudo chown -R $(whoami): ${2}/go
    popd >/dev/null
}


check_and_install_golang