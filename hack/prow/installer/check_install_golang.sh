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

# This script requires two parameters:
# $1. INSTALL_PATH: The path to install the golang binary
# $2. GO_VERSION: The version of golang to install



set -eux -o pipefail

if (($# < 2)); then
  echo "ERROR: given ! ($#) parameters but expected 2."
  echo "USAGE: ./check_install_golang.sh INSTALL_PATH GO_VERSION" 
  exit 1
fi

VERSION_TO_INSTALL=${2}
INSTALL_PATH=${1}

function current_arch() {
  case $(arch) in
  "x86_64" | "i386")
     echo "amd64"
  ;;
  "aarch64" | "arm64")
    echo "arm64"
  ;;
  *)
    echo "unexpected arch: $(arch). use amd64" 1>&2
    echo "amd64"
  ;;
  esac
}

ARCH=${ARCH:=$(current_arch)}

# installs or updates golang if right version doesn't exists
function check_and_install_golang() {
	# always reinstall the requested version to avoid permission issues
    install_golang "$VERSION_TO_INSTALL" "$INSTALL_PATH"
  
}

# install_golang takes two parameters version and path to install.
function install_golang() {
  local -r GO_VER="$1"
  local -r GO_DIR="$2/go"
  echo "Installing golang version: $GO_VER in $GO_DIR"

  INSTALLOS=linux
  if [[ "$OSTYPE" == "darwin"* ]]; then
    INSTALLOS=darwin
  fi

  local -r GO_TGZ="go${GO_VER}.${INSTALLOS}-${ARCH}.tar.gz"
  pushd /tmp

  # using sudo because previously installed versions might have been installed by a different user.
  sudo rm -rf "$GO_TGZ"
  curl -qL -O "https://go.dev/dl/$GO_TGZ"
  sudo rm -rf "$GO_DIR"
  sudo mkdir -p "$GO_DIR"
  sudo tar -C "$GO_DIR" --strip-components=1 -xzf "$GO_TGZ"

  popd >/dev/null
  echo "installed in $GO_DIR: $($GO_DIR/bin/go version)"
}

check_and_install_golang
