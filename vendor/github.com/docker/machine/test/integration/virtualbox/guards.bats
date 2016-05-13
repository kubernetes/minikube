#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER virtualbox

use_disposable_machine

@test "$DRIVER: Should not allow machine creation with bad ISO" {
  run machine create -d virtualbox --virtualbox-boot2docker-url http://dev.null:9111/bad.iso $NAME
  [[ ${status} -eq 1 ]]
}

@test "$DRIVER: Should not allow machine creation with engine-install-url" {
  run machine create --engine-install-url https://test.docker.com -d virtualbox $NAME
  [[ ${output} == *"--engine-install-url cannot be used"*  ]]
}