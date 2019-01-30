#!/bin/bash
#
# This script will copy the local minikube build to a remote host,
# execute tests.sh script,
# gather and write the results locally
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
#
set -x

SCRIPTDIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# First test connectivity and build readiness
# If any starting condition tests fail, fail test completely

# Test JUMPHOSTIP connectivity
echo "=== RUN  ssh to jumphost"
jumphoststatus=$(ssh -q -F ${SCRIPTDIR}/ssh_config jumphost echo ok 2>&1)
[ "$jumphoststatus" != "ok" ] && echo "--- FAIL: ssh ${JUMPHOSTIP}" && exit -1
echo "--- PASS: ssh to jumphost"

# Test remote Mac connectivity
echo "=== RUN  ssh to macnode"
macnodestatus=$(ssh -q -F ${SCRIPTDIR}/ssh_config macnode echo ok 2>&1)
[ "$macnodestatus" != "ok" ] && echo "--- FAIL: ssh -F ${SCRIPTDIR}/ssh_config macnode" && exit -1
echo "--- PASS: ssh to macnode"

# Test remote Win connectivity
#echo "=== RUN  ssh to winnode"
#[$(ssh -q -F ${SCRIPTDIR}/ssh_config macnode exit)] && echo "FAIL:ssh -F ${SCRIPTDIR}/ssh_config macnode" && exit -1
#echo "--- PASS: ssh to winnode"

# Test remote LinuxGPU connectivity
#echo "=== RUN  ssh to gpunode"
#[$(ssh -q -F ${SCRIPTDIR}/ssh_config macnode exit)] && echo "FAIL:ssh -F ${SCRIPTDIR}/ssh_config macnode" && exit -1
#echo "--- PASS: ssh to gpunode"

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
   
  # Test build outputs for out/minikube-darwin-amd64 
  ssh -F ${SCRIPTDIR}/ssh_config macnode "cd ${REMOTEDIR} && GOPATH=\$HOME/${REMOTEBASE} ./test.sh"

}

# out/minikube-osx
runtests($MINIKUBE_LOCATION, "Integration_hyperkit", "osx_integration_tests_hyperkit.sh", "macnode")

# out/minikube-linux-amd64

# out/minikube-windows-amd64
