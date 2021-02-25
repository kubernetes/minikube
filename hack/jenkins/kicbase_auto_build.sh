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

set -x

# Make sure docker is installed and configured
./hack/jenkins/installers/check_install_docker.sh
yes|gcloud auth configure-docker
docker login -u ${DOCKERHUB_USER} -p ${DOCKERHUB_PASS}

# Make sure gh is installed and configured
./hack/jenkins/installers/check_install_gh.sh

# Let's make sure we have the newest kicbase reference
curl -L https://github.com/kubernetes/minikube/raw/master/pkg/drivers/kic/types.go --output types-head.go
# kicbase tags are of the form VERSION-TIMESTAMP-PR, so this grep finds that TIMESTAMP in the middle
# if it doesn't exist, it will just return VERSION, which is covered in the if statement below
HEAD_KIC_TIMESTAMP=$(egrep "Version =" types-head.go | cut -d \" -f 2 | cut -d "-" -f 2)
CURRENT_KIC_TS=$(egrep "Version =" pkg/drivers/kic/types.go | cut -d \" -f 2 | cut -d "-" -f 2)
if [[ $HEAD_KIC_TIMESTAMP != v* ]]; then
	diff=$((CURRENT_KIC_TS-HEAD_KIC_TIMESTAMP))
	if [[ $CURRENT_KIC_TS == v* ]] || [ $diff -lt 0 ]; then
		gh pr comment ${ghprbPullId} --body "Hi ${ghprbPullAuthorLoginMention}, your kicbase info is out of date. Please rebase."
		exit 1
	fi
fi
rm types-head.go

# Setup variables
if [[ -z $KIC_VERSION ]]; then
	# Testing PRs here
	release=false
	now=$(date +%s)
	KV=$(egrep "Version =" pkg/drivers/kic/types.go | cut -d \" -f 2 | cut -d "-" -f 1)
	GCR_REPO=gcr.io/k8s-minikube/kicbase-builds
	DH_REPO=kicbase/build
	export KIC_VERSION=$KV-$now-$ghprbPullId
else
	# Actual kicbase release here
	release=true
	GCR_REPO=${GCR_REPO:-gcr.io/k8s-minikube/kicbase}
	DH_REPO=${DH_REPO:-kicbase/stable}
	export KIC_VERSION
fi
GCR_IMG=${GCR_REPO}:${KIC_VERSION}
DH_IMG=${DH_REPO}:${KIC_VERSION}
export KICBASE_IMAGE_REGISTRIES="${GCR_IMG} ${DH_IMG}"


# Build a new kicbase image
yes|make push-kic-base-image

# Abort with error message if above command failed
ec=$?
if [ $ec -gt 0 ]; then
	if [ "$release" = false ]; then
		gh pr comment ${ghprbPullId} --body "Hi ${ghprbPullAuthorLoginMention}, building a new kicbase image failed, please try again."
	fi
	exit $ec
fi

# Retrieve the sha from the new image
docker pull $GCR_IMG
fullsha=$(docker inspect --format='{{index .RepoDigests 0}}' $KICBASE_IMAGE_REGISTRIES)
sha=$(echo ${fullsha} | cut -d ":" -f 2)
git config user.name "minikube-bot"
git config user.email "minikube-bot@google.com"


if [ "$release" = false ]; then
	# Update the user's PR with the newly built kicbase image.

	git remote add ${ghprbPullAuthorLogin} git@github.com:${ghprbPullAuthorLogin}/minikube.git
	git fetch ${ghprbPullAuthorLogin}
	git checkout -b ${ghprbPullAuthorLogin}-${ghprbSourceBranch} ${ghprbPullAuthorLogin}/${ghprbSourceBranch}

	sed -i "s|Version = .*|Version = \"${KIC_VERSION}\"|;s|baseImageSHA = .*|baseImageSHA = \"${sha}\"|;s|gcrRepo = .*|gcrRepo = \"${GCR_REPO}\"|;s|dockerhubRepo = .*|dockerhubRepo = \"${DH_REPO}\"|" pkg/drivers/kic/types.go; make generate-docs;

	git commit -am "Updating kicbase image to ${KIC_VERSION}"
	git push ${ghprbPullAuthorLogin} HEAD:${ghprbSourceBranch}

	gh pr comment ${ghprbPullId} --body "Hi ${ghprbPullAuthorLoginMention}, kicbase has been updated to the newly built image. Please pull the commits locally if you want to test."
else
	# We're releasing, so open a new PR with the newly released kicbase
	
	branch=kicbase-release-${KIC_VERSION}
	git checkout -b ${branch}

	sed -i "s|Version = .*|Version = \"${KIC_VERSION}\"|;s|baseImageSHA = .*|baseImageSHA = \"${sha}\"|;s|gcrRepo = .*|gcrRepo = \"${GCR_REPO}\"|;s|dockerhubRepo = .*|dockerhubRepo = \"${DH_REPO}\"|" pkg/drivers/kic/types.go; make generate-docs;

	git add -A
	git commit -m "Update kicbase to ${KIC_VERSION}"
	git remote add minikube-bot git@github.com:minikube-bot/minikube.git
	git push -f minikube-bot ${branch}

	gh pr create --fill --base master --head minikube-bot:${branch}
fi
