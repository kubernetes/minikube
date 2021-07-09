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

# This script can take the following env variables
# ARGS: args to pass into the make rule
# 	ISO_BUCKET = the bucket location to upload the ISO (e.g. minikube-builds/PR_NUMBER)
# 	ISO_VERSION = the suffix for the iso (i.e. minikube-$(ISO_VERSION).iso)

set -x -o pipefail

# Make sure gh is installed and configured
./hack/jenkins/installers/check_install_gh.sh

# Make sure all required packages are installed
sudo apt-get update
sudo apt-get -y install build-essential unzip rsync bc python2 p7zip-full

if [[ -z $ISO_VERSION ]]; then
	release=false
	IV=$(egrep "ISO_VERSION \?=" Makefile | cut -d " " -f 3 | cut -d "-" -f 1)
	now=$(date +%s)
	export ISO_VERSION=$IV-$now-$ghprbPullId
	export ISO_BUCKET=minikube-builds/iso/$ghprbPullId
else
	release=true
	export ISO_VERSION
	export ISO_BUCKET
fi

make release-iso | tee iso-logs.txt
# Abort with error message if above command failed
ec=$?
if [ $ec -gt 0 ]; then
	if [ "$release" = false ]; then
		gh pr comment ${ghprbPullId} --body "Hi ${ghprbPullAuthorLoginMention}, building a new ISO failed.  
		See the logs at: 
		```
		https://storage.cloud.google.com/minikube-builds/logs/${ghprbPullId}/${ghprbActualCommit:0:7}/iso_build.txt
		```
		"
	fi
	exit $ec
fi

git config user.name "minikube-bot"
git config user.email "minikube-bot@google.com"

if [ "$release" = false ]; then
	# Update the user's PR with newly build ISO

	git remote add ${ghprbPullAuthorLogin} git@github.com:${ghprbPullAuthorLogin}/minikube.git
	git fetch ${ghprbPullAuthorLogin}
	git checkout -b ${ghprbPullAuthorLogin}-${ghprbSourceBranch} ${ghprbPullAuthorLogin}/${ghprbSourceBranch}

	sed -i "s/ISO_VERSION ?= .*/ISO_VERSION ?= ${ISO_VERSION}/" Makefile
	sed -i "s|isoBucket := .*|isoBucket := \"${ISO_BUCKET}\"|" pkg/minikube/download/iso.go
	make generate-docs

	git add Makefile pkg/minikube/download/iso.go site/content/en/docs/commands/start.md
	git commit -m "Updating ISO to ${ISO_VERSION}"
	git push ${ghprbPullAuthorLogin} HEAD:${ghprbSourceBranch}

	message="Hi ${ghprbPullAuthorLoginMention}, we have updated your PR with the reference to newly built ISO. Pull the changes locally if you want to test with them or update your PR further."
	if [ $? -gt 0 ]; then
		message="Hi ${ghprbPullAuthorLoginMention}, we failed to push the reference to the ISO to your PR. Please run the following command and push manually.

		sed -i 's/ISO_VERSION ?= .*/ISO_VERSION ?= ${ISO_VERSION}/' Makefile; sed -i 's|isoBucket := .*|isoBucket := "${ISO_BUCKET}"|' pkg/minikube/download/iso.go; make generate-docs;
		"
	fi
	
	gh pr comment ${ghprbPullId} --body "${message}"
else
	# Release!
	branch=iso-release-${ISO_VERSION}
	git checkout -b ${branch}

	sed -i "s/ISO_VERSION ?= .*/ISO_VERSION ?= ${ISO_VERSION}/" Makefile
	sed -i "s|isoBucket := .*|isoBucket := \"${ISO_BUCKET}\"|" pkg/minikube/download/iso.go
	make generate-docs

	git add Makefile pkg/minikube/download/iso.go site/content/en/docs/commands/start.md
	git commit -m "Update ISO to ${ISO_VERSION}"
	git remote add minikube-bot git@github.com:minikube-bot/minikube.git
	git push -f minikube-bot ${branch}

	gh pr create --fill --base master --head minikube-bot:${branch}
fi	
