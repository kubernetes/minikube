#!/bin/bash

# Copyright 2025 The Kubernetes Authors All rights reserved.
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

# The script expects the following env variables:
# OS: The operating system
# ARCH: The architecture
# DRIVER: the driver to use for the test
# CONTAINER_RUNTIME: the container runtime to use for the test
# EXTRA_START_ARGS: additional flags to pass into minikube start
# EXTRA_TEST_ARGS: additional flags to pass into go test
# JOB_NAME: the name of the logfile and check name to update on github
# PULL_NUMBER: the PR number, if applicable

function print_test_info() {
	echo ">> Starting at $(date)"
	echo ""
	echo "user:      $(whoami)"
	echo "arch:      ${OS_ARCH}"
	echo "pr:        ${PULL_NUMBER}"
	echo "driver:    ${DRIVER}"
	echo "runtime:   ${CONTAINER_RUNTIME}"
	echo "job:       ${JOB_NAME}"
	echo "test home: ${TEST_HOME}"
	echo "kernel:    $(uname -v)"
	echo "uptime:    $(uptime)"
	# Setting KUBECONFIG prevents the version check from erroring out due to permission issues
	echo "kubectl:   $(env KUBECONFIG=${TEST_HOME} kubectl version --client)"
	echo "docker:    $(docker version --format '{{ .Client.Version }}')"
	echo "podman:    $(sudo podman version --format '{{.Version}}' || true)"
	echo "go:        $(go version || true)"

	case "${DRIVER}" in
	kvm2)
		echo "virsh:     $(virsh --version)"
		;;
	virtualbox)
		echo "vbox:      $(vboxmanage --version)"
		;;
	vfkit)
		echo "vfkit:     $(vfkit --version)"
		;;
	krunkit)
		echo "krunkit:   $(krunkit --version)"
		;;
	esac

	echo ""
}

function install_dependencies() {
	# We need pstree for the restart cronjobs
	if [ "$(uname)" != "Darwin" ]; then
		sudo apt-get -y install lsof psmisc dnsutils
	else
		# install brew if not present
		if ! command -v brew >/dev/null 2>&1; then
			/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
			echo >> /Users/ec2-user/.zprofile
			echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> /Users/ec2-user/.zprofile
			eval "$(/opt/homebrew/bin/brew shellenv)"
		fi
		# install docker cli
		brew install docker

		
		# install vfkit
		brew update
		brew install vfkit pstree coreutils pidof
		ln -s /usr/local/bin/gtimeout /usr/local/bin/timeout || true
		# install vmnet shared on macos in non-interactive mode
		curl -fsSL https://github.com/minikube-machine/vmnet-helper/releases/latest/download/install.sh | sudo VMNET_INTERACTIVE=0 bash

		# ensure go dirs are owned by current user so that go install can write
		sudo chown -R $(whoami) $HOME/go
		sudo chown -R $(whoami) $HOME/Library/Caches/go-build
	fi
	# do NOT change manually - only using make update-golang-version
	GOLANG_VERSION_TO_INSTALL=1.25.5
	# install golang if not present
	hack/prow/installer/check_install_golang.sh /usr/local $GOLANG_VERSION_TO_INSTALL || true
	# install gotestsum if not present
	hack/prow/installer/check_install_gotestsum.sh || true
	# install gopogh
	hack/prow/installer/check_install_gopogh.sh || true

	# install jq
	if ! type "jq" >/dev/null; then
		echo ">> Installing jq"
		if [ "${ARCH}" == "arm64" && "${OS}" == "linux" ]; then
		# linux arm64
			sudo apt-get install jq -y
		elif [ "${ARCH}" == "arm64" && "${OS}" == "darwin"]; then
		# macos arm64
			brew install jq
		elif [ "${OS}" == "linux" ]; then
		# linux x86
			curl -LO https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 && sudo install jq-linux64 /usr/local/bin/jq
		else
		# macos ax86
			curl -LO https://github.com/stedolan/jq/releases/download/jq-1.6/jq-osx-amd64 && sudo install jq-osx-amd64 /usr/local/bin/jq
		fi
	fi

}

function docker_setup() {
	if [ "$(uname)" != "Darwin" ]; then
		# clean all docker artifacts up
		docker system prune -a --volumes -f || true
		docker system df || true
		docker rm -f -v $(docker ps -aq) >/dev/null 2>&1 || true

		# read only token, never expires
		#todo: do we need this token
		# docker login -u minikubebot -p "$DOCKERHUB_READONLY_TOKEN"
	fi
}

function gvisor_image_build() {
	
	# skipping gvisor mac because ofg https://github.com/kubernetes/minikube/issues/5137
	if [ "$(uname)" != "Darwin" ]; then
		# Build the gvisor image so that we can integration test changes to pkg/gvisor
		chmod +x testdata/gvisor-addon
		# Should match GVISOR_IMAGE_VERSION in Makefile
		docker build -t gcr.io/k8s-minikube/gvisor-addon:2 -f testdata/gvisor-addon-Dockerfile ./testdata
	fi
}

function run_gopogh() {
    # todo: currently we do not save to gopogh db
    echo "Not saving to DB"
    gopogh -in "${JSON_OUT}" -out_html "${HTML_OUT}" -out_summary "${SUMMARY_OUT}" -name "${JOB_NAME}" -pr "${PULL_NUMBER}" -repo github.com/kubernetes/minikube/ -details "${COMMIT}:$(date +%Y-%m-%d)"

}

# this is where the script starts
readonly OS_ARCH="${OS}-${ARCH}"
readonly TEST_ROOT="${PWD}/minikube-integration"
readonly TEST_HOME="$(pwd)/${MINIKUBE_LOCATION}-$$"

export GOPATH="$HOME/go"
export KUBECONFIG="${TEST_HOME}/kubeconfig"
export PATH=$PATH:"/usr/local/bin/:/usr/local/go/bin/:$GOPATH/bin"
export MINIKUBE_SUPPRESS_DOCKER_PERFORMANCE=true

readonly TIMEOUT=120m

cp -r test/integration/testdata .

# Add the out/ directory to the PATH, for using new drivers.
export PATH="$(pwd)/out/":$PATH
mkdir -p "${TEST_ROOT}"
mkdir -p "${TEST_HOME}"
export MINIKUBE_HOME="${TEST_HOME}/.minikube"
export MINIKUBE_BIN="out/minikube-${OS_ARCH}"
export E2E_BIN="out/e2e-${OS_ARCH}"

install_dependencies
docker_setup


if [ "$CONTAINER_RUNTIME" == "containerd" ]; then
	cp out/gvisor-addon testdata/
	gvisor_image_build
fi

print_test_info

readonly TEST_OUT="${TEST_HOME}/testout.txt"
readonly JSON_OUT="${TEST_HOME}/test.json"
readonly JUNIT_OUT="${TEST_HOME}/junit-unit.xml"
readonly HTML_OUT="${TEST_HOME}/test.html"
readonly SUMMARY_OUT="${TEST_HOME}/test_summary.json"

touch "${TEST_OUT}"
touch "${JSON_OUT}"
touch "${JUNIT_OUT}"
touch "${HTML_OUT}"
touch "${SUMMARY_OUT}"

e2e_start_time="$(date -u +%s)"
echo ""
echo ">> Starting ${E2E_BIN} at $(date)"
set -x

EXTRA_START_ARGS="${EXTRA_START_ARGS} --container-runtime=${CONTAINER_RUNTIME}"
echo $PATH
gotestsum --jsonfile "${JSON_OUT}" --junitfile="${JUNIT_OUT}" -f standard-verbose --raw-command -- \
	go tool test2json -t \
	${E2E_BIN} \
	-minikube-start-args="--driver=${DRIVER} ${EXTRA_START_ARGS}" \
	-test.timeout=${TIMEOUT} -test.v \
	${EXTRA_TEST_ARGS} \
	-binary="${MINIKUBE_BIN}" 2>&1 |
	tee "${TEST_OUT}"

result=${PIPESTATUS[0]} # capture the exit code of the first cmd in pipe.
set +x
echo ">> ${E2E_BIN} exited with ${result} at $(date)"
echo ""

# calculate the time took to finish running e2e binary test.
e2e_end_time="$(date -u +%s)"
elapsed=$(($e2e_end_time - $e2e_start_time))
min=$(($elapsed / 60))
sec=$(tail -c 3 <<<$((${elapsed}00 / 60)))
elapsed=$min.$sec

#todo: currently we skip gopogh upload , we shall add it back
run_gopogh

# according to prow's requirement, upload the test report to $ARTIFACTS
cp ${TEST_OUT} .
cp ${JSON_OUT} .
cp ${JUNIT_OUT} .
cp ${HTML_OUT} .
cp ${SUMMARY_OUT} .
if [[ $result -eq 0 ]]; then
	echo "minikube: SUCCESS"
else
	echo "minikube: FAIL"
fi
MINIKUBE_BIN delete --all --purge || true

exit "$result"
