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

# Gather data comparing the overhead of multiple local Kubernetes (macOS and linux)
readonly TESTS=$1

# How many iterations to cycle through
readonly TEST_ITERATIONS=10

# How long to poll CPU usage for (each point is an average over this period)
readonly POLL_DURATION=5s

# How long to measure background usage for. 5 minutes too short, 10 minutes too long
readonly TOTAL_DURATION=5m

# How all tests will be identified
readonly SESSION_ID="$(date +%Y%m%d-%H%M%S)-$$"

# OS Type
readonly OS=$(uname)

measure() {
  local name=$1
  local iteration=$2
  local filename="benchmark-results/${SESSION_ID}/cstat.${name}.$$-${iteration}"

  echo ""
  echo "  >> Current top processes by CPU:"
  if [[ "${OS}" == "Darwin" ]]; then
    top -n 3 -l 2 -s 2 -o cpu  | tail -n4 | awk '{ print $1 " " $2 " " $3 " " $4 }'
  elif [[ "${OS}" == "Linux" ]]; then
    top -b -n 3 -o %CPU | head -n 9
  fi

  if [[ "${iteration}" == 0 ]]; then
    echo "NOTE: dry-run iteration: will not record measurements"
    cstat --poll "${POLL_DURATION}" --for "${POLL_DURATION}" --busy
    return
  fi

  echo ""
  echo "  >> Measuring ${name} and saving to out/${filename} ..."
  cstat --poll "${POLL_DURATION}" --for "${TOTAL_DURATION}" --busy --header=false | tee "$(pwd)/out/${filename}"
}


cleanup() {
  echo "  >> Deleting local clusters and Docker containers ..."
  out/minikube delete --all 2>/dev/null >/dev/null
  k3d cluster delete 2>/dev/null >/dev/null
  kind delete cluster 2>/dev/null >/dev/null
  docker stop $(docker ps -q) 2>/dev/null
  docker kill $(docker ps -q) 2>/dev/null
  docker rm $(docker ps -a -q) 2>/dev/null
  sleep 2
}

pause_if_running_apps() {
  while true; do
    local apps=$(osascript -e 'tell application "System Events" to get name of (processes where background only is false)'  | tr ',' '\n' | sed s/"^ "//g)
    local quiet=0

    for app in $apps; do
      quiet=1
      if [[ "${app}" != "Terminal" && "${app}" != "Finder" ]]; then
        echo "Unexpected application running: \"${app}\" - will sleep"
        quiet=0
      fi
    done

    pmset -g batt | grep 'AC Power'
    if [[ "$?" != 0 ]]; then
      echo "waiting to be plugged in ..."
      sleep 5
      continue
    fi

    if [[ "${quiet}" == 1 ]]; then
      break
    else
      echo "waiting for apps to be closed ..."
      sleep 5
    fi

  done
}

fail() {
  local name=$1
  local iteration=$2

  echo '***********************************************************************'
  echo "${name} failed on iteration ${iteration} - will not record measurement"
  echo '***********************************************************************'

  if [[ "${iteration}" == 0 ]]; then
    echo "test environment appears invalid, exiting"
    exit 90
  fi
}

start_docker() {
    local docker_up=0
    local started=0

    while [[ "${docker_up}" == 0 ]]; do
      docker info >/dev/null && docker_up=1 || docker_up=0

      if [[ "${docker_up}" == 0 && "${started}" == 0 ]]; then
          if [[ "${OS}" == "Darwin" ]]; then
            echo ""
            echo "  >> Starting Docker for Desktop ..."
            open -a Docker
            started=1
          elif [[ "${OS}" == "Linux" ]]; then
            echo ""
            echo "  >> Starting Docker Engine ..."
            sudo systemctl start docker
            started=1
          fi
      fi

      sleep 1
    done

    # Give time for d4d Kubernetes to begin, if it's around
    if [[ "${started}" == 1 ]]; then
      sleep 60
    fi
}


main() {
  # check if cstat is installed
  CSTAT=$(which cstat)
  if [[ "$?" != 0 ]]; then
    echo "cstat in not installed. Install cstat at https://github.com/tstromberg/cstat"
    exit 1
  fi

  echo "----[ versions ]------------------------------------"
  k3d version || { echo "k3d version failed. Please install latest k3d"; exit 1; }
  kind version || { echo "kind version failed. Please install latest kind"; exit 1; }
  out/minikube version || { echo "minikube version failed"; exit 1; }
  docker version
  echo "----------------------------------------------------"
  echo ""

  echo "Session ID: ${SESSION_ID}"
  mkdir -p "out/benchmark-results/${SESSION_ID}"

  echo ""

  if [[ "${OS}" == "Darwin" ]]; then
    echo "Turning on Wi-Fi for initial downloads"
    networksetup -setairportpower Wi-Fi on
  fi

  for i in $(seq 0 ${TEST_ITERATIONS}); do
    echo ""
    echo "==> session ${SESSION_ID}, iteration $i"

    cleanup

    if [[ "$i" = 0 ]]; then
      echo "NOTE: The 0 iteration is an unmeasured dry run!"
    else
      if [[ "${OS}" == "Darwin" ]]; then
        pause_if_running_apps
        echo "Turning off Wi-Fi to remove background noise"
        networksetup -setairportpower Wi-Fi off

        echo "  >> Killing Docker for Desktop ..."
        osascript -e 'quit app "Docker"'
      elif [[ "${OS}" == "Linux" ]]; then
        echo "  >> Killing Docker Engine ..."
        sudo systemctl stop docker
      fi

      # Measure the background noise on this system
      sleep 15
      measure idle $i
    fi

    # Run cleanup once we can assert that Docker is up
    start_docker
    cleanup

    docker_k8s=0
    # depending on whether Docker for Mac Kubernetes is enabled
    if [[ "${OS}" == "Darwin" ]]; then
      # wait kubernetes system pods for Docker for Mac, if it is enabled
      sleep 60
      kubectl --context docker-desktop version

      # measure Docker for Mac Kubernetes
      if  [[ $? == 0 ]]; then
        echo "Kubernetes is running in Docker for Desktop - adjusting tests"
        docker_k8s=1
        measure docker_k8s $i
      # measure Docker idle
      else
        measure docker $i
      fi
    # measure Docker idle only
    elif [[ "${OS}" == "Linux" ]]; then
      measure docker $i
    fi

    # measure k3d and kind
    echo ""
    echo "-> k3d"
    time k3d cluster create && measure k3d $i || fail k3d $i
    cleanup

    echo ""
    echo "-> kind"
    time kind create cluster && measure kind $i || fail kind $i
    cleanup

    # test different drivers
    if [[ "${OS}" == "Darwin" ]]; then
      drivers=(docker hyperkit virtualbox)
    elif [[ "${OS}" == "Linux" ]]; then
      drivers=(docker kvm2 virtualbox)
    fi

    for driver in "${drivers[@]}"; do
      echo ""
      echo "-> out/minikube --driver=${driver}"
      time out/minikube start --driver "${driver}" && measure "minikube.${driver}" $i || fail "minikube.${driver}" $i
      cleanup

      # We won't be needing docker for the remaining tests this iteration
      if [[ "${OS}" == "Darwin" && "${driver}" == "docker" ]]; then
        echo "  >> Quitting Docker for Desktop ..."
        osascript -e 'quit app "Docker"'
      elif [[ "${OS}" == "Linux" && ${driver} == "docker" ]]; then
        echo "  >> Quitting Docker Engine ..."
        sudo systemctl stop docker
      fi
    done ## driver
  done ## iteration
}

main "$@"
# update benchmark result into docs contents
./hack/benchmark/cpu_usage/idle_only/update_summary.sh "${SESSION_ID}"
go run ./hack/benchmark/cpu_usage/idle_only/chart.go "${SESSION_ID}"
