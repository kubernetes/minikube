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

# This script builds the minikube binary for all 3 platforms, uploads the
# results to GCS, and then runs a series of platform specific integrations tests
#
# This is to done as part of the CI tests for Github PRs
#
# The script expects the following env variables:
# ghprbPullId: The pull request ID, injected from the ghpbr plugin.
# ghprbActualCommit: The commit hash, injected from the ghpbr plugin.
#
# This script relies a reverse proxy having been started on the remote
# test host to a known jumphost.
# Access goes like this:
# buildhost --ssh:localhost:HOSTPORT--> jumphost <--sshReverse-- remotetesthost
#
# To set it up, public keys must be deployed in multiple places:
# - jumphost $USER/.ssh/authorized_keys must include the pubkeys from both
#   the buildhost user and the remotetesthost user
# - remotetesthost $USER/.ssh/authorized_keys must include the pubkey from
#   the buildhost user
#
# With keys in place start the reverse proxy on remotetesthost with:
# nohup ssh -i ~/.ssh/REMOTEHOSTKEY -R REMOTEHOSTPORT:localhost:22 JUMPHOSTIP -N &
# e.g. for the mac do:
# nohup ssh -i ~/.ssh/id_rsa_jumphost_don -R 19999:localhost:22 35.231.224.8 -N &
#


set -eux -o pipefail

readonly bucket="minikube-builds-prow"

declare -rx BUILD_IN_DOCKER=y
declare -rx ISO_BUCKET="${bucket}/${ghprbPullId}"
declare -rx ISO_VERSION="testing"
declare -rx TAG="${ghprbActualCommit}"
declare -rx SCRIPTDIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function build_and_upload() {
  make -j 16 all && failed=$? || failed=$?

  gsutil cp "gs://${bucket}/logs/index.html" \
    "gs://${bucket}/logs/${ghprbPullId}/index.html"

  if [[ "${failed}" -ne 0 ]]; then
    echo "build failed"
    exit "${failed}"
  fi

  git diff ${ghprbActualCommit} --name-only \
    $(git merge-base origin/master ${ghprbActualCommit}) \
    | grep -q deploy/iso/minikube && rebuild=1 || rebuild=0

  if [[ "${rebuild}" -eq 1 ]]; then
    echo "ISO changes detected ... rebuilding ISO"
    make release-iso
  fi

  cp -r test/integration/testdata out/

  # Don't upload the buildroot artifacts if they exist
  rm -r out/buildroot || true

  gsutil -m cp -r out/* "gs://${bucket}/${ghprbPullId}/"
}

function test_connectivity() {
  NODE=$1
  echo "=== RUN ssh to ${NODE}"

  connectstatus=$(ssh -q -F ${SCRIPTDIR}/ssh_config ${NODE} echo ok 2>&1)
  [ "$connectstatus" != "ok" ] && echo "--- FAIL: ssh -F ${SCRIPTDIR}/ssh_config ${NODE}" && exit -1
  echo "--- PASS: ssh to ${NODE}"
}

function runtests() {
  MINIKUBE_LOCATION=$1
  MINIKUBE_TESTNAME=$2
  MINIKUBE_TESTSCRIPT=$3
  MINIKUBE_TESTNODE=$4
  # Delivery dir
  REMOTEBASE="minikube-prow-tests/${MINIKUBE_LOCATION}/${MINIKUBE_TESTNAME}"
  REMOTEDIR="${REMOTEBASE}/src/k8s.io/minikube"

  # Deliver to macnode
  ssh -F ${SCRIPTDIR}/ssh_config ${MINIKUBE_TESTNODE} "mkdir -p ${REMOTEDIR}"
  scp -F ${SCRIPTDIR}/ssh_config ${SCRIPTDIR}/${MINIKUBE_TESTSCRIPT} ${MINIKUBE_TESTNODE}:${REMOTEDIR}
  ssh -F ${SCRIPTDIR}/ssh_config ${MINIKUBE_TESTNODE} "MINIKUBE_LOCATION=${MINIKUBE_LOCATION} cd ${REMOTEDIR} && ${MINIKUBE_TESTSCRIPT}"
}

#First test connectivity to all remote test hosts.  If we can't connect, don't even try to build
test_connectivity("jumphost")
test_connectivity("macnode")
#test_connectivity("winnode")
#test_connectivity("linuxgpunode")
#test_connectivity("linuxnode")

#Now build and upload
build_and_upload()

#Start remote tests
runtests(${ghprbPullId}, "integration_tests_hyperkit", "remote_osx_integration_tests_hyperkit.sh", "macnode")
