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

./hack/jenkins/installers/check_install_docker.sh
yes|gcloud auth configure-docker
now=$(date +%s)
KV=$(egrep "Version =" pkg/drivers/kic/types.go | cut -d \" -f 2 | cut -d "-" -f 1)
export KIC_VERSION=$KV-$now-$ghprbPullId
export KICBASE_IMAGE_REGISTRIES=gcr.io/k8s-minikube/kicbase-builds:$KIC_VERSION

curl -L https://github.com/kubernetes/minikube/raw/master/pkg/drivers/kic/types.go --output types-head.go
HEAD_KIC_TIMESTAMP=$(egrep "Version =" types-head.go | cut -d \" -f 2 | cut -d "-" -f 2)
CURRENT_KIC_TS=$(egrep "Version =" pkg/drivers/kic/types.go | cut -d \" -f 2 | cut -d "-" -f 2)
if [[ $HEAD_KIC_TIMESTAMP != v* ]] && [[ $CURRENT_KIC_TS != v* ]]; then
	diff=$((CURRENT_KIC_TS-HEAD_KIC_TIMESTAMP))
	if [ $diff -lt 0 ]; then
		curl -s -H "Authorization: token ${access_token}" \
		-H "Accept: application/vnd.github.v3+json" \
		-X POST -d "{\"body\": \"Hi ${ghprbPullAuthorLoginMention}, your kicbase info is out of date. Please rebase.\"}" "https://api.github.com/repos/kubernetes/minikube/issues/$ghprbPullId/comments"
		exit 0
	fi
fi
rm types-head.go
yes|make push-kic-base-image

docker pull $KICBASE_IMAGE_REGISTRIES
fullsha=$(docker inspect --format='{{index .RepoDigests 0}}' $KICBASE_IMAGE_REGISTRIES)
sha=$(echo ${fullsha} | cut -d ":" -f 2)

message="Hi ${ghprbPullAuthorLoginMention},\\n\\nA new kicbase image is available, please update your PR with the new tag and SHA.\\nIn pkg/drivers/kic/types.go:\\n\\n\\t// Version is the current version of kic\\n\\tVersion = \\\"${KIC_VERSION}\\\"\\n\\t// SHA of the kic base image\\n\\tbaseImageSHA = \\\"${sha}\\\"\\nThen run \`make generate-docs\` to update our documentation to reference the new image."

curl -s -H "Authorization: token ${access_token}" \
	 -H "Accept: application/vnd.github.v3+json" \
	 -X POST -d "{\"body\": \"${message}\"}" "https://api.github.com/repos/kubernetes/minikube/issues/$ghprbPullId/comments"
