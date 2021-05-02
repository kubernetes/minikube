#!/usr/bin/env bash

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

# wrapper.sh handles setting up things before / after the test command $@
#
# usage: wrapper.sh my-test-command [my-test-args]
#
# Things wrapper.sh handles:
# - starting / stopping docker-in-docker
# - ensuring GOPATH/bin is in PATH
#
# After handling these things / before cleanup, my-test-command will be invoked,
# and the exit code of my-test-command will be preserved by wrapper.sh

set -o errexit
set -o pipefail
set -o nounset

>&2 echo "wrapper.sh] [INFO] Wrapping Test Command: \`$*\`"
printf '%0.s=' {1..80} >&2; echo >&2
>&2 echo "wrapper.sh] [SETUP] Performing pre-test setup ..."

cleanup(){
  >&2 echo "wrapper.sh] [CLEANUP] Cleaning up after Docker in Docker ..."
  docker ps -aq | xargs -r docker rm -f || true
  service docker stop || true
  >&2 echo "wrapper.sh] [CLEANUP] Done cleaning up after Docker in Docker."
}

early_exit_handler() {
  >&2 echo "wrapper.sh] [EARLY EXIT] Interrupted, entering handler ..."
  if [ -n "${EXIT_VALUE:-}" ]; then
    >&2 echo "Original exit code was ${EXIT_VALUE}, not preserving due to interrupt signal"
  fi
  cleanup
  >&2 echo "wrapper.sh] [EARLY EXIT] Completed handler ..."
  exit 1
}

trap early_exit_handler TERM INT


>&2 echo "wrapper.sh] [SETUP] Docker in Docker enabled, initializing ..."
# If we have opted in to docker in docker, start the docker daemon,
service docker start
# the service can be started but the docker socket not ready, wait for ready
WAIT_N=0
while true; do
  # docker ps -q should only work if the daemon is ready
  docker ps -q > /dev/null 2>&1 && break
  if [[ ${WAIT_N} -lt 5 ]]; then
    WAIT_N=$((WAIT_N+1))
    echo "wrapper.sh] [SETUP] Waiting for Docker to be ready, sleeping for ${WAIT_N} seconds ..."
    sleep ${WAIT_N}
  else
    echo "wrapper.sh] [SETUP] Reached maximum attempts, not waiting any longer ..."
    break
  fi
done
echo "wrapper.sh] [SETUP] Done setting up Docker in Docker."

# add $GOPATH/bin to $PATH
export GOPATH="${GOPATH:-${HOME}/go}"
export PATH="${GOPATH}/bin:${PATH}"
mkdir -p "${GOPATH}/bin"

# actually run the user supplied command
printf '%0.s=' {1..80}; echo
>&2 echo "wrapper.sh] [TEST] Running Test Command: \`$*\` ..."
set +o errexit
"$@"
EXIT_VALUE=$?
set -o errexit
>&2 echo "wrapper.sh] [TEST] Test Command exit code: ${EXIT_VALUE}"

# cleanup
cleanup

# preserve exit value from user supplied command
printf '%0.s=' {1..80} >&2; echo >&2
>&2 echo "wrapper.sh] Exiting ${EXIT_VALUE}"
exit ${EXIT_VALUE}
