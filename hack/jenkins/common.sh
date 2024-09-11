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


# This script downloads the test files from the build bucket and makes some executable.

# The script expects the following env variables:
# OS: The operating system
# ARCH: The architecture
# DRIVER: the driver to use for the test
# CONTAINER_RUNTIME: the container runtime to use for the test
# EXTRA_START_ARGS: additional flags to pass into minikube start
# EXTRA_TEST_ARGS: additional flags to pass into go test
# JOB_NAME: the name of the logfile and check name to update on github

set -x

readonly OS_ARCH="${OS}-${ARCH}"
readonly TEST_ROOT="${HOME}/minikube-integration"
readonly TEST_HOME="${TEST_ROOT}/${MINIKUBE_LOCATION}-$$"

export GOPATH="$HOME/go"
export KUBECONFIG="${TEST_HOME}/kubeconfig"
export PATH=$PATH:"/usr/local/bin/:/usr/local/go/bin/:$GOPATH/bin"
export MINIKUBE_SUPPRESS_DOCKER_PERFORMANCE=true

readonly TIMEOUT=${1:-120m}

public_log_url="https://storage.googleapis.com/minikube-builds/logs/${MINIKUBE_LOCATION}/${ROOT_JOB_ID}/${JOB_NAME}.html"

# retry_github_status provides reliable github status updates
function retry_github_status() {
  local commit=$1
  local context=$2
  local state=$3
  local token=$4
  local target=$5
  local desc=$6

   # Retry in case we hit our GitHub API quota or fail other ways.
  local attempt=0
  local timeout=2
  local code=-1

  echo "set GitHub status $context to $desc"

  while [[ "${attempt}" -lt 8 ]]; do
    local out=$(mktemp)
    code=$(curl -o "${out}" -s --write-out "%{http_code}" -L -u minikube-bot:${token} \
      "https://api.github.com/repos/kubernetes/minikube/statuses/${commit}" \
      -H "Content-Type: application/json" \
      -X POST \
      -d "{\"state\": \"${state}\", \"description\": \"Jenkins: ${desc}\", \"target_url\": \"${target}\", \"context\": \"${context}\"}" || echo 999)

    # 2xx HTTP codes
    if [[ "${code}" =~ ^2 ]]; then
      break
    fi

    cat "${out}" && rm -f "${out}"
    echo "HTTP code ${code}! Retrying in ${timeout} .."
    sleep "${timeout}"
    attempt=$(( attempt + 1 ))
    timeout=$(( timeout * 5 ))
  done
}

if [ "$(uname)" = "Darwin" ]; then
  if [ "$ARCH" = "arm64" ]; then
    export PATH=$PATH:/opt/homebrew/bin
  fi

  if ! bash setup_docker_desktop_macos.sh; then
    retry_github_status "${COMMIT}" "${JOB_NAME}" "failure" "${access_token}" "${public_log_url}" "Jenkins: docker failed to start"
    exit 1
  fi
fi

# We need pstree for the restart cronjobs
if [ "$(uname)" != "Darwin" ]; then
  sudo apt-get -y install lsof psmisc dnsutils
else
  brew install pstree coreutils pidof
  ln -s /usr/local/bin/gtimeout /usr/local/bin/timeout || true
fi

# installing golang so we can go install gopogh
./installers/check_install_golang.sh "/usr/local" || true

# install docker and kubectl if not present
sudo ARCH="$ARCH" ./installers/check_install_docker.sh || true

# install gotestsum if not present
GOROOT="/usr/local/go" ./installers/check_install_gotestsum.sh || true

# install cron jobs
if [ "$OS" == "linux" ]; then
  source ./installers/check_install_linux_crons.sh
else
  source ./installers/check_install_osx_crons.sh
fi

# let's just clean all docker artifacts up
docker system prune -a --volumes -f || true
docker system df || true

echo ">> Starting at $(date)"
echo ""
echo "arch:      ${OS_ARCH}"
echo "build:     ${MINIKUBE_LOCATION}"
echo "driver:    ${DRIVER}"
echo "runtime:   ${CONTAINER_RUNTIME}"
echo "job:       ${JOB_NAME}"
echo "test home: ${TEST_HOME}"
echo "kernel:    $(uname -v)"
echo "uptime:    $(uptime)"
# Setting KUBECONFIG prevents the version check from erroring out due to permission issues
echo "kubectl:   $(env KUBECONFIG=${TEST_HOME} kubectl version --client --short=true)"
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
esac

echo ""
mkdir -p out/ testdata/

# Install gsutil if necessary.
if ! type -P gsutil >/dev/null; then
  if [[ ! -x "out/gsutil/gsutil" ]]; then
    echo "Installing gsutil to $(pwd)/out ..."
    curl -s https://storage.googleapis.com/pub/gsutil.tar.gz | tar -C out/ -zxf -
  fi
  PATH="$(pwd)/out/gsutil:$PATH"
fi

# Add the out/ directory to the PATH, for using new drivers.
export PATH="$(pwd)/out/":$PATH

echo
echo ">> Downloading test inputs from ${MINIKUBE_LOCATION} ..."
gsutil -qm cp \
  "gs://minikube-builds/${MINIKUBE_LOCATION}/minikube-${OS_ARCH}" \
  "gs://minikube-builds/${MINIKUBE_LOCATION}/docker-machine-driver"-* \
  "gs://minikube-builds/${MINIKUBE_LOCATION}/e2e-${OS_ARCH}" out

gsutil -qm cp -r "gs://minikube-builds/${MINIKUBE_LOCATION}/testdata"/* testdata/

gsutil -qm cp "gs://minikube-builds/${MINIKUBE_LOCATION}/gvisor-addon" testdata/


# Set the executable bit on the e2e binary and out binary
export MINIKUBE_BIN="out/minikube-${OS_ARCH}"
export E2E_BIN="out/e2e-${OS_ARCH}"
chmod +x "${MINIKUBE_BIN}" "${E2E_BIN}" out/docker-machine-driver-*
"${MINIKUBE_BIN}" version

procs=$(pgrep "minikube-${OS_ARCH}|e2e-${OS_ARCH}" || true)
if [[ "${procs}" != "" ]]; then
  echo "Warning: found stale test processes to kill:"
  ps -f -p ${procs} || true
  kill ${procs} || true
  kill -9 ${procs} || true
fi

# Quickly notice misconfigured test roots
mkdir -p "${TEST_ROOT}"

# Cleanup stale test outputs.
echo
echo ">> Cleaning up after previous test runs ..."
for test_path in ${TEST_ROOT}; do
  ls -lad "${test_path}" || continue

  echo "* Cleaning stale test path: ${test_path}"
  for tunnel in $(find ${test_path} -name tunnels.json -type f); do
    env MINIKUBE_HOME="$(dirname ${tunnel})" ${MINIKUBE_BIN} tunnel --cleanup || true
  done

  for home in $(find ${test_path} -name .minikube -type d); do
    env MINIKUBE_HOME="$(dirname ${home})" ${MINIKUBE_BIN} delete --all || true
    sudo rm -Rf "${home}"
  done

  for kconfig in $(find ${test_path} -name kubeconfig -type f); do
    sudo rm -f "${kconfig}"
  done

  ## ultimate shotgun clean up docker after we tried all
  docker rm -f -v $(docker ps -aq) >/dev/null 2>&1 || true

  # Be very specific to avoid accidentally deleting other items, like wildcards or devices
  if [[ -d "${test_path}" ]]; then
    rm -Rf "${test_path}" || true
  elif [[ -f "${test_path}" ]]; then
    rm -f "${test_path}" || true
  fi
done

function cleanup_procs() {
  # sometimes tests left over zombie procs that won't exit
  # for example:
  # jenkins  20041  0.0  0.0      0     0 ?        Z    Aug19   0:00 [minikube-linux-] <defunct>
  pgrep docker > d.pids
  zombie_defuncts=$(ps -A -ostat,ppid | grep -v -f d.pids | awk '/[zZ]/ && !a[$2]++ {print $2}')
  if [[ "${zombie_defuncts}" != "" ]]; then
    echo "Found zombie defunct procs to kill..."
    ps -f -p ${zombie_defuncts} || true
    kill ${zombie_defuncts} || true
  fi

  if type -P virsh; then
    virsh -c qemu:///system list --all --uuid \
      | xargs -I {} sh -c "virsh -c qemu:///system destroy {}; virsh -c qemu:///system undefine {}" \
      || true
    echo ">> virsh VM list after clean up (should be empty):"
    virsh -c qemu:///system list --all || true
  fi

  if type -P vboxmanage; then
    killall VBoxHeadless || true
    sleep 1
    killall -9 VBoxHeadless || true

    for guid in $(vboxmanage list vms | grep -Eo '\{[a-zA-Z0-9-]+\}'); do
      echo "- Removing stale VirtualBox VM: $guid"
      vboxmanage startvm "${guid}" --type emergencystop || true
      vboxmanage unregistervm "${guid}" || true
    done

    ifaces=$(vboxmanage list hostonlyifs | grep -E "^Name:" | awk '{ print $2 }')
    for if in $ifaces; do
      vboxmanage hostonlyif remove "${if}" || true
    done

    echo ">> VirtualBox VM list after clean up (should be empty):"
    vboxmanage list vms || true
    echo ">> VirtualBox interface list after clean up (should be empty):"
    vboxmanage list hostonlyifs || true
  fi

  if type -P hdiutil; then
    hdiutil info | grep -E "/dev/disk[1-9][^s]" || true
    hdiutil info \
        | grep -E "/dev/disk[1-9][^s]" \
        | awk '{print $1}' \
        | xargs -I {} sh -c "hdiutil detach {}" \
        || true
  fi

  # cleaning up stale hyperkits
  if type -P hyperkit; then
    for pid in $(pgrep hyperkit); do
      info=$(ps -f -p "$pid")
      if [[ $info == *"com.docker.hyperkit"* ]]; then
        continue
      fi
      echo "Killing stale hyperkit $pid"
      echo "$info" || true
      kill "$pid" || true
      kill -9 "$pid" || true
    done
  fi

  if [[ "${DRIVER}" == "hyperkit" ]]; then
    # even though Internet Sharing is disabled in the UI settings, it's still preventing HyperKit from starting
    # the error is "Could not create vmnet interface, permission denied or no entitlement?"
    # I've discovered that if you kill the "InternetSharing" process that this resolves the error and HyperKit starts normally
    sudo pkill InternetSharing
    if [[ -e out/docker-machine-driver-hyperkit ]]; then
      sudo chown root:wheel out/docker-machine-driver-hyperkit || true
      sudo chmod u+s out/docker-machine-driver-hyperkit || true
    fi
  fi

  kprocs=$(pgrep kubectl || true)
  if [[ "${kprocs}" != "" ]]; then
    echo "error: killing hung kubectl processes ..."
    ps -f -p ${kprocs} || true
    kill ${kprocs} || true
  fi


  # clean up none drivers binding on 8443
  none_procs=$(sudo lsof -i :8443 | tail -n +2 | awk '{print $2}' || true)
  if [[ "${none_procs}" != "" ]]; then
    echo "Found stale api servers listening on 8443 processes to kill: "
    for p in $none_procs
    do
      echo "Killing stale none driver: $p"
      ps -f -p $p || true
      kill $p || true
      kill -9 $p || true
    done
  fi
}

function cleanup_stale_routes() {
  local show="netstat -rn -f inet"
  local del="sudo route -n delete"

  if [[ "$(uname)" == "Linux" ]]; then
    show="ip route show"
    del="sudo ip route delete"
  fi

  local troutes=$($show | awk '{ print $1 }' | grep 10.96.0.0 || true)
  for route in ${troutes}; do
    echo "WARNING: deleting stale tunnel route: ${route}"
    $del "${route}" || true
  done
}

cleanup_procs || true
cleanup_stale_routes || true

mkdir -p "${TEST_HOME}"
export MINIKUBE_HOME="${TEST_HOME}/.minikube"


# Build the gvisor image so that we can integration test changes to pkg/gvisor
chmod +x ./testdata/gvisor-addon
# skipping gvisor mac because ofg https://github.com/kubernetes/minikube/issues/5137
if [ "$(uname)" != "Darwin" ]; then
  # Should match GVISOR_IMAGE_VERSION in Makefile
  docker build -t gcr.io/k8s-minikube/gvisor-addon:2 -f testdata/gvisor-addon-Dockerfile ./testdata
fi

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

readonly TEST_OUT="${TEST_HOME}/testout.txt"
readonly JSON_OUT="${TEST_HOME}/test.json"
readonly HTML_OUT="${TEST_HOME}/test.html"
readonly SUMMARY_OUT="${TEST_HOME}/test_summary.json"

e2e_start_time="$(date -u +%s)"
echo ""
echo ">> Starting ${E2E_BIN} at $(date)"
set -x

if test -f "${TEST_OUT}"; then
  rm "${TEST_OUT}" || true # clean up previous runs of same build
fi
touch "${TEST_OUT}"

if [ ! -z "${CONTAINER_RUNTIME}" ]
then
    EXTRA_START_ARGS="${EXTRA_START_ARGS} --container-runtime=${CONTAINER_RUNTIME}"
fi

if test -f "${JSON_OUT}"; then
  rm "${JSON_OUT}" || true # clean up previous runs of same build
fi

touch "${JSON_OUT}"

gotestsum --jsonfile "${JSON_OUT}" -f standard-verbose --raw-command -- \
  go tool test2json -t \
  ${E2E_BIN} \
    -minikube-start-args="--driver=${DRIVER} ${EXTRA_START_ARGS}" \
    -test.timeout=${TIMEOUT} -test.v \
    ${EXTRA_TEST_ARGS} \
    -binary="${MINIKUBE_BIN}" 2>&1 \
  | tee "${TEST_OUT}"

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
elapsed=$(($e2e_end_time-$e2e_start_time))
min=$(($elapsed/60))
sec=$(tail -c 3 <<< $((${elapsed}00/60)))
elapsed=$min.$sec

if ! type "jq" > /dev/null; then
    echo ">> Installing jq"
    if [ "${ARCH}" == "arm64" && "${OS}" == "linux" ]; then
      sudo apt-get install jq -y
    elif [ "${ARCH}" == "arm64" ]; then
      echo "Unable to install 'jq' automatically for arm64 on Darwin, please install 'jq' manually."
      exit 5
    elif [ "${OS}" != "darwin" ]; then
      curl -LO https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 && sudo install jq-linux64 /usr/local/bin/jq
    else
      curl -LO https://github.com/stedolan/jq/releases/download/jq-1.6/jq-osx-amd64 && sudo install jq-osx-amd64 /usr/local/bin/jq
    fi
fi

echo ">> Installing gopogh"
./installers/check_install_gopogh.sh

echo ">> Running gopogh"
if test -f "${HTML_OUT}"; then
    rm "${HTML_OUT}" || true # clean up previous runs of same build
fi

touch "${HTML_OUT}"
touch "${SUMMARY_OUT}"
echo "EXTERNAL: *$EXTERNAL*"
echo "MINIKUBE_LOCATION: *$MINIKUBE_LOCATION*"
if [ "$EXTERNAL" != "yes" ] && [ "$MINIKUBE_LOCATION" = "master" ]
then
	echo "Saving to DB"
	gopogh -in "${JSON_OUT}" -out_html "${HTML_OUT}" -out_summary "${SUMMARY_OUT}" -name "${JOB_NAME}" -pr "${MINIKUBE_LOCATION}" -repo github.com/kubernetes/minikube/  -details "${COMMIT}:$(date +%Y-%m-%d):${ROOT_JOB_ID}" -db_backend "${GOPOGH_DB_BACKEND}" -db_host "${GOPOGH_DB_HOST}" -db_path "${GOPOGH_DB_PATH}" -use_cloudsql -use_iam_auth
	echo "Exit code: $?"
else
	echo "Not saving to DB"
	gopogh -in "${JSON_OUT}" -out_html "${HTML_OUT}" -out_summary "${SUMMARY_OUT}" -name "${JOB_NAME}" -pr "${MINIKUBE_LOCATION}" -repo github.com/kubernetes/minikube/  -details "${COMMIT}:$(date +%Y-%m-%d):${ROOT_JOB_ID}"
fi
gopogh_status=$(cat "${SUMMARY_OUT}")
fail_num=$(echo $gopogh_status | jq '.NumberOfFail')
test_num=$(echo $gopogh_status | jq '.NumberOfTests')
pessimistic_status="${fail_num} / ${test_num} failures"
description="completed with ${status} in ${elapsed} minutes."
if [ "$status" = "failure" ]; then
  description="completed with ${pessimistic_status} in ${elapsed} minutes."
fi
echo "$description"

REPORT_URL_BASE="https://storage.googleapis.com"

if [ -z "${EXTERNAL}" ]; then
  # If we're already in GCP, then upload results to GCS directly
  SHORT_COMMIT=${COMMIT:0:7}
  JOB_GCS_BUCKET="minikube-builds/logs/${MINIKUBE_LOCATION}/${ROOT_JOB_ID}/${JOB_NAME}"

  echo ">> Copying ${TEST_OUT} to gs://${JOB_GCS_BUCKET}.out.txt"
  echo ">>   public URL: ${REPORT_URL_BASE}/${JOB_GCS_BUCKET}.out.txt"
  gsutil -qm cp "${TEST_OUT}" "gs://${JOB_GCS_BUCKET}.out.txt"

  echo ">> uploading ${JSON_OUT} to gs://${JOB_GCS_BUCKET}.json"
  echo ">>   public URL:  ${REPORT_URL_BASE}/${JOB_GCS_BUCKET}.json"
  gsutil -qm cp "${JSON_OUT}" "gs://${JOB_GCS_BUCKET}.json" || true

  echo ">> uploading ${HTML_OUT} to gs://${JOB_GCS_BUCKET}.html"
  echo ">>   public URL:  ${REPORT_URL_BASE}/${JOB_GCS_BUCKET}.html"
  gsutil -qm cp "${HTML_OUT}" "gs://${JOB_GCS_BUCKET}.html" || true

  echo ">> uploading ${SUMMARY_OUT} to gs://${JOB_GCS_BUCKET}_summary.json"
  echo ">>   public URL:  ${REPORT_URL_BASE}/${JOB_GCS_BUCKET}_summary.json"
  gsutil -qm cp "${SUMMARY_OUT}" "gs://${JOB_GCS_BUCKET}_summary.json" || true
else 
  # Otherwise, put the results in a predictable spot so the upload job can find them
  REPORTS_PATH=test_reports
  mkdir -p "$REPORTS_PATH"
  cp "${TEST_OUT}" "$REPORTS_PATH/out.txt"
  cp "${JSON_OUT}" "$REPORTS_PATH/out.json"
  cp "${HTML_OUT}" "$REPORTS_PATH/out.html"
  cp "${SUMMARY_OUT}" "$REPORTS_PATH/summary.json"
fi

echo ">> Cleaning up after ourselves ..."
timeout 3m ${MINIKUBE_BIN} tunnel --cleanup || true
timeout 5m ${MINIKUBE_BIN} delete --all --purge >/dev/null 2>/dev/null || true
cleanup_stale_routes || true

rm -Rf "${MINIKUBE_HOME}" || true
rm -f "${KUBECONFIG}" || true
rm -f "${TEST_OUT}" || true
rm -f "${JSON_OUT}" || true
rm -f "${HTML_OUT}" || true
rm -f "${SUMMARY_OUT}" || true

rmdir "${TEST_HOME}" || true
echo ">> ${TEST_HOME} completed at $(date)"

if [[ "${MINIKUBE_LOCATION}" == "master" ]]; then
  exit "$result"
fi

retry_github_status "${COMMIT}" "${JOB_NAME}" "${status}" "${access_token}" "${public_log_url}" "${description}"

exit "$result"
