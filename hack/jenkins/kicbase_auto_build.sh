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

./hack/jenkins/installers/check_install_docker.sh
yes|gcloud auth configure-docker
now=$(date +%s)
export KV=$(egrep "Version =" pkg/drivers/kic/types.go | cut -d \" -f 2 | cut -d "-" -f 1)
export KIC_VERSION=$KV-$now-$ghprbPullId
export KICBASE_IMAGE_REGISTRIES=gcr.io/k8s-minikube/kicbase-builds:$KIC_VERSION
yes|make push-kic-base-image

docker pull $KICBASE_IMAGE_REGISTRIES
fullsha=$(docker inspect --format='{{index .RepoDigests 0}}' $KICBASE_IMAGE_REGISTRIES)
sha=$(echo ${fullsha} | cut -d ":" -f 2)

message="Hi ${ghprbPullAuthorLoginMention},

A new kicbase image is available, please update your PR with the new tag and SHA.
In pkg/drivers/kic/types.go:

	// Version is the current version of kic
	Version = \"${KICBASE_IMAGE_REGISTRIES}\"
	// SHA of the kic base image
	baseImageSHA = \"${sha}\"
"

curl -s -H "Authorization: token ${access_token}" \
	 -H "Accept: application/vnd.github.v3+json" \
	 -X POST -d "{\"body\": \"${message}\"}" "https://api.github.com/repos/kubernetes/minikube/issues/$ghprbPullId/comments"
