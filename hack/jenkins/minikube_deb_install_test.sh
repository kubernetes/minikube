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

# This script builds the minikube binary for all 3 platforms and uploads them.
# This is to done as part of the CI tests for Github PRs

# The script expects the following env variables:
# bucket: The GCP bucket to get .deb files from
# debs: .deb packages to test

set -eu -o pipefail

trap exit SIGINT

for pkg in ${debs[@]}; do
    gsutil cp -r "gs://${bucket}/${pkg}.deb" .
done  

declare -ra distros=(\
	   debian:sid\
	   debian:latest\
	   debian:buster\
	   debian:stretch\
	   ubuntu:latest\
	   ubuntu:20.10 \
	   ubuntu:20.04 \
	   ubuntu:19.10 \
	   ubuntu:19.04 \
	   ubuntu:18.10 \
	   ubuntu:18.04)


for distro in "${distros[@]}"; do
    for pkg in ${debs[@]}; do
	echo   "=============================================================="
	printf "Install %s on %s\n" "${pkg}" "${distro}"
	echo   "=============================================================="
	docker run --rm -v "$PWD:/var/tmp" "${distro}" sh -c "apt-get update; \
	    (dpkg -i /var/tmp/${pkg} || 
		(apt-get -fy install && dpkg -i /var/tmp/${pkg})) 
	    || echo "Failed to install ${pkg}.deb on ${distro}"
    done
done

