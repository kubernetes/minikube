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

# This script generates the Github Release page and uploads all the binaries/etc to that page
# This is intended to be run on a new release tag in order to generate the github release page for that release

# The script expects the following env variables:
# VERSION_MAJOR: The major version of the tag to be released.
# VERSION_MINOR: The minor version of the tag to be released.
# VERSION_BUILD: The build version of the tag to be released.
# ISO_SHA256: The sha 256 of the minikube-iso for the current release.
# GITHUB_TOKEN: The Github API access token. Injected by the Jenkins credential provider.

set -e
export TAGNAME=v${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}
export DEB_VERSION=${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}
export RPM_VERSION=${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}

export GITHUB_ORGANIZATION="kubernetes"
export GITHUB_REPO="minikube"
export PROJECT_NAME="minikube"
export DARWIN_SHA256=$(cat out/minikube-darwin-amd64.sha256)
export LINUX_SHA256=$(cat out/minikube-linux-amd64.sha256)
export WINDOWS_SHA256=$(cat out/minikube-windows-amd64.exe.sha256)

# Description could be moved into file on machine or fetched via URL.  Doing this for now as it is the simplest, portable solution.
# ================================================================================
export DESCRIPTION="# Minikube ${TAGNAME}
Minikube is still under active development, and features may change at any time. Release notes are available [here](https://github.com/kubernetes/minikube/blob/${TAGNAME}/CHANGELOG.md).

## Distribution
Minikube is distributed in binary form for Linux, OSX, and Windows systems for the ${TAGNAME} release. Please note that Windows support is currently experimental and may have issues.  Binaries are available through GitHub or on Google Cloud Storage. The direct GCS links are:
[Darwin/amd64](https://storage.googleapis.com/minikube/releases/${TAGNAME}/minikube-darwin-amd64)
[Linux/amd64](https://storage.googleapis.com/minikube/releases/${TAGNAME}/minikube-linux-amd64)
[Windows/amd64](https://storage.googleapis.com/minikube/releases/${TAGNAME}/minikube-windows-amd64.exe)

## Installation
### OSX
\`\`\`shell
curl -Lo minikube https://storage.googleapis.com/minikube/releases/${TAGNAME}/minikube-darwin-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube
\`\`\`
Feel free to leave off \`\`\`sudo cp minikube /usr/local/bin/ && rm minikube\`\`\` if you would like to add minikube to your path manually.

Or you can install via homebrew with \`brew cask install minikube\`.

### Linux
\`\`\`shell
curl -Lo minikube https://storage.googleapis.com/minikube/releases/${TAGNAME}/minikube-linux-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube
\`\`\`
Feel free to leave off \`\`\`sudo cp minikube /usr/local/bin/ && rm minikube\`\`\` if you would like to add minikube to your path manually.

### Debian Package (.deb) [Experimental]
Download the \`minikube_${DEB_VERSION}.deb\` file, and install it using \`sudo dpkg -i minikube_${DEB_VERSION}.deb\`

### RPM Package (.rpm) [Experimental]
Download the \`minikube-${RPM_VERSION}.rpm\` file, and install it using \`sudo rpm -i minikube-${RPM_VERSION}.rpm\`

### Windows [Experimental]
Download the \`minikube-windows-amd64.exe\` file, rename it to \`minikube.exe\` and add it to your path.

### Windows Installer [Experimental]
Download the \`minikube-installer.exe\` file, and execute the installer.  This will automatically add minikube.exe to your path with an uninstaller available as well.

## Usage
Documentation is available [here](https://github.com/kubernetes/minikube/blob/${TAGNAME}/README.md).

## Checksums
Minikube consists of a binary executable and a VM image in ISO format. To verify the contents of your distribution, you can compare sha256 hashes with these values:

\`\`\`
$ tail -n +1 -- out/*.sha256
==> out/minikube-darwin-amd64.sha256 <==
${DARWIN_SHA256}

==> out/minikube-linux-amd64.sha256 <==
${LINUX_SHA256}

==> out/minikube-windows-amd64.exe.sha256 <==
${WINDOWS_SHA256}
\`\`\`

### ISO
\`\`\`shell
$ openssl sha256 minikube.iso
SHA256(minikube.iso)=
${ISO_SHA256}
\`\`\`
"
# ================================================================================

# Deleting release from github before creating new one
github-release delete --user ${GITHUB_ORGANIZATION} --repo ${GITHUB_REPO} --tag ${TAGNAME} || true

# Creating a new release in github
github-release release \
    --user ${GITHUB_ORGANIZATION} \
    --repo ${GITHUB_REPO} \
    --tag ${TAGNAME} \
    --name "${TAGNAME}" \
    --description "${DESCRIPTION}"

# Uploading the files into github
FILES_TO_UPLOAD=(
    'minikube-linux-amd64'
    'minikube-linux-amd64.sha256'
    'minikube-darwin-amd64'
    'minikube-darwin-amd64.sha256'
    'minikube-windows-amd64.exe'
    'minikube-windows-amd64.exe.sha256'
    'minikube-installer.exe'
    "minikube_${DEB_VERSION}.deb"
    "minikube-${RPM_VERSION}.rpm"
    'docker-machine-driver-kvm2'
    'docker-machine-driver-kvm2.sha256'
    'docker-machine-driver-hyperkit'
    'docker-machine-driver-hyperkit.sha256'
)

for UPLOAD in "${FILES_TO_UPLOAD[@]}"
do
    n=0
    until [ $n -ge 5 ]
    do
        github-release upload --user ${GITHUB_ORGANIZATION} --repo ${GITHUB_REPO} --tag ${TAGNAME} --name $UPLOAD --file out/$UPLOAD && break
        n=$[$n+1]
        sleep 15
    done
done
