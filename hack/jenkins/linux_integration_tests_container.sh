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


# This script runs the integration tests on a Linux machine for the Virtualbox Driver

# The script expects the following env variables:
# MINIKUBE_LOCATION: GIT_COMMIT from upstream build.
# COMMIT: Actual commit ID from upstream build
# EXTRA_BUILD_ARGS (optional): Extra args to be passed into the minikube integrations tests
# access_token: The Github API access token. Injected by the Jenkins credential provider.



set -e
result=0
JOB_NAME="Linux-Container"

CID="$(docker run -d -p 127.0.0.1:4321:4321 --privileged gcr.io/k8s-minikube/localkube-dind-image-devshell:$COMMIT /start.sh)"

kubectl config set-cluster minikube --server=http://127.0.0.1:4321
kubectl config set-context minikube --cluster=minikube
kubectl config use-context minikube
# this for loop waits until kubectl can access the api server that minikube has created
KUBECTL_UP="false"
set +e
for i in {1..150} # timeout for 5 minutes
do
   kubectl get po
   if [ $? -ne 1 ]; then
      KUBECTL_UP="true"
      echo "INIT SUCCESS: kubectl could reached api-server in allotted time"
      break
  fi
  sleep 2
done
if [ "$KUBECTL_UP" != "true" ]; then
  echo "INIT FAILURE: kubectl could not reach api-server in allotted time"
  result=1
fi
# kubectl commands are now able to interact with minikube cluster

set -e
docker stop $CID
docker rm $CID

if [[ $result -eq 0 ]]; then
  status="success"
else
  status="failure"
  source print-debug-info.sh
fi

set +x
target_url="https://storage.googleapis.com/minikube-builds/logs/${MINIKUBE_LOCATION}/${JOB_NAME}.txt"
curl "https://api.github.com/repos/kubernetes/minikube/statuses/${COMMIT}?access_token=$access_token" \
  -H "Content-Type: application/json" \
  -X POST \
  -d "{\"state\": \"$status\", \"description\": \"Jenkins\", \"target_url\": \"$target_url\", \"context\": \"${JOB_NAME}\"}"
set -x

exit $result
