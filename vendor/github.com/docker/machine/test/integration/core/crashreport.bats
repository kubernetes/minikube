#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER virtualbox

use_disposable_machine

@test "$DRIVER: should send bugsnag report" {
  # we exploit a 'bug' where vboxmanage wont allow a machine created with 1mb of RAM
  run machine --bugsnag-api-token nonexisting -D create -d virtualbox --virtualbox-memory 1 $NAME
  echo ${output}
  [ "$status" -eq 1 ]
  [[ ${output} == *"notifying bugsnag"* ]]
}
