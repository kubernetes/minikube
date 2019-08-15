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
    install_golang $VERSION_TO_INSTALL $INSTALL_PATH
    return
	fi
	
	# golang has been installed and check its version
	if [[ $(go version) =~ (([0-9]+)\.([0-9]+).([0-9]+).([\.0-9]*)) ]]
	then
		host_golang_version=${BASH_REMATCH[1]}	
    if [ $host_golang_version = $VERSION_TO_INSTALL ]; then
      echo "go version on the host looks good : $host_golang_version"
    else
      echo "WARNING: expected go version to be $VERSION_TO_INSTALL but got $host_golang_version"
      install_golang $VERSION_TO_INSTALL $INSTALL_PATH
    fi
  	
  
  else
  	warn "Failed to parse golang version."
		return
	fi
}

# install_golang takes two parameters version and path to install.
function install_golang {
    echo "Installing golang version: $1 on $2"  
    pushd /tmp >/dev/null
    curl -qL -O "https://storage.googleapis.com/golang/go${version_to_install}.linux-amd64.tar.gz" &&
      tar xfa go${version_to_install}.linux-amd64.tar.gz &&
      rm -rf "${INSTALL_PATH}/go" &&
      mv go "${INSTALL_PATH}/" &&
    popd >/dev/null

    pushd "${INSTALL_PATH}/go/src/go/types" > /dev/null
    echo "Installing gotype linter"
    go build gotype.go
    cp gotype "${INSTALL_PATH}/go/bin"
    popd >/dev/null
}


check_and_install_golang