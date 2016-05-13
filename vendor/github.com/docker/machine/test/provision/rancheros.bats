#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash


# this should move to the makefile

if [[ "$DRIVER" != "virtualbox" ]]; then
    exit 0
fi

export RANCHEROS_VERSION="v0.3.1"
export RANCHEROS_ISO="https://github.com/rancherio/os/releases/download/$RANCHEROS_VERSION/machine-rancheros.iso"

@test "$DRIVER: create with RancherOS ISO" {
  VIRTUALBOX_BOOT2DOCKER_URL="$RANCHEROS_ISO" run ${BASE_TEST_DIR}/run-bats.sh ${BASE_TEST_DIR}/core
  echo ${output}
  [ ${status} -eq 0 ]
}
