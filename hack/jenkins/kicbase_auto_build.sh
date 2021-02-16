#!/bin/bash

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
