#!/bin/bash

# The script expects the following env variables:
# OS: The operating system
# ARCH: The architecture
# DRIVER: the driver to use for the test
# CONTAINER_RUNTIME: the container runtime to use for the test
# EXTRA_START_ARGS: additional flags to pass into minikube start
# EXTRA_TEST_ARGS: additional flags to pass into go test
# JOB_NAME: the name of the logfile and check name to update on github


MINIKUBE_LOCATION="minikube_location" #what should we set here

readonly OS_ARCH="${OS}-${ARCH}"
readonly TEST_ROOT="${HOME}/minikube-integration"
readonly TEST_HOME="${TEST_ROOT}/${MINIKUBE_LOCATION}-$$"

export GOPATH="$HOME/go"
export KUBECONFIG="${TEST_HOME}/kubeconfig"
export PATH=$PATH:"/usr/local/bin/:/usr/local/go/bin/:$GOPATH/bin"
export MINIKUBE_SUPPRESS_DOCKER_PERFORMANCE=true

readonly TIMEOUT=120m

mkdir -p out/ testdata/
# Add the out/ directory to the PATH, for using new drivers.
export PATH="$(pwd)/out/":$PATH
mkdir -p "${TEST_ROOT}"
mkdir -p "${TEST_HOME}"
export MINIKUBE_HOME="${TEST_HOME}/.minikube"
export MINIKUBE_BIN="out/minikube-${OS_ARCH}"
export E2E_BIN="out/e2e-${OS_ARCH}"

function print_test_info() {
    echo ">> Starting at $(date)"
    echo ""
    echo "user:      $(whoami)"
    echo "arch:      ${OS_ARCH}"
    echo "build:     ${MINIKUBE_LOCATION}"
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
        brew install pstree coreutils pidof
        ln -s /usr/local/bin/gtimeout /usr/local/bin/timeout || true
    fi
    # install golang if not present
    sudo hack/prow/installer/check_install_golang.sh /usr/local 1.24.5 || true
    # install gotestsum if not present
    GOROOT="/usr/local/go" hack/prow/installer/check_install_gotestsum.sh || true
    # instal docker if not present
    ARCH="$ARCH" hack/prow/installer/check_install_docker.sh || true
    sudo usermod -aG docker minitest || true
    sudo adduser $(whoami) docker || true
    newgrp docker

    # install jq
    if ! type "jq" >/dev/null; then
        echo ">> Installing jq"
        if [ "${ARCH}" == "arm64" && "${OS}" == "linux" ]; then
            sudo apt-get install jq -y
        elif [ "${ARCH}" == "arm64" ]; then
            echo "Unable to install 'jq' automatically for arm64 on Darwin, please install 'jq' manually."
            exit 5
        elif [ "${OS}" != "darwin" ]; then
            curl -LO https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 &&sudo  install jq-linux64 /usr/local/bin/jq
        else
            curl -LO https://github.com/stedolan/jq/releases/download/jq-1.6/jq-osx-amd64 && sudo install jq-osx-amd64 /usr/local/bin/jq
        fi
    fi

}

function docker_setup() {

    # clean all docker artifacts up
    docker system prune -a --volumes -f || true
    docker system df || true
    docker rm -f -v $(docker ps -aq) >/dev/null 2>&1 || true

    # read only token, never expires
    #todo: do we need this token
    # docker login -u minikubebot -p "$DOCKERHUB_READONLY_TOKEN"
}

function gvisor_image_build() {
    # Build the gvisor image so that we can integration test changes to pkg/gvisor
    chmod +x out/testdata/gvisor-addon
    # skipping gvisor mac because ofg https://github.com/kubernetes/minikube/issues/5137
    if [ "$(uname)" != "Darwin" ]; then
        # Should match GVISOR_IMAGE_VERSION in Makefile
        docker build -t gcr.io/k8s-minikube/gvisor-addon:2 -f out/testdata/gvisor-addon-Dockerfile ./out/testdata
    fi
}

function sleep_on_high_load() {
    readonly LOAD=$(uptime | grep -E -o "load average.*: [0-9]+" | cut -d" " -f3)
    if [[ "${LOAD}" -gt 2 ]]; then
        echo ""
        echo "********************** LOAD WARNING ********************************"
        echo "Load average is very high (${LOAD}), which may cause failures. Top:"
        if [[ "$(uname)" == "Darwin" ]]; then
            # Two samples, macOS does not calculate CPU usage on the first one
            top -l 2 -o cpu -n 5 | tail -n 15
        else
            top -b -n1 | head -n 15
        fi
        echo "********************** LOAD WARNING ********************************"
        echo "Sleeping 30s to see if load goes down ...."
        sleep 30
        uptime
    fi
}

install_dependencies
docker_setup
print_test_info
gvisor_image_build
sleep_on_high_load

readonly TEST_OUT="${TEST_HOME}/testout.txt"
readonly JSON_OUT="${TEST_HOME}/test.json"
readonly JUNIT_OUT="${TEST_HOME}/junit-unit.xml"

touch "${TEST_OUT}"
touch "${JSON_OUT}"
touch "${JUNIT_OUT}"

e2e_start_time="$(date -u +%s)"
echo ""
echo ">> Starting ${E2E_BIN} at $(date)"
set -x

EXTRA_START_ARGS="${EXTRA_START_ARGS} --container-runtime=${CONTAINER_RUNTIME}"

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

if [[ $result -eq 0 ]]; then
    status="success"
    echo "minikube: SUCCESS"
else
    status="failure"
    echo "minikube: FAIL"
fi

# calculate the time took to finish running e2e binary test.
e2e_end_time="$(date -u +%s)"
elapsed=$(($e2e_end_time - $e2e_start_time))
min=$(($elapsed / 60))
sec=$(tail -c 3 <<<$((${elapsed}00 / 60)))
elapsed=$min.$sec

#todo: currently we skip gopogh upload , we shall add it back

# according to prow's requirement, upload the test report to $ARTIFACTS
cp ${TEST_OUT} .
cp ${JSON_OUT} .
cp ${JUNIT_OUT} .
