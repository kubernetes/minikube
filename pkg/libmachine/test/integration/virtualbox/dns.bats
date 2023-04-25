#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER virtualbox

use_disposable_machine

@test "$DRIVER: Create a vm with a dns proxy set" {
  run machine create -d $DRIVER --virtualbox-dns-proxy=true $NAME
  [[ ${status} -eq 0 ]]
}

@test "$DRIVER: Check DNSProxy flag is properly set during machine creation" {
  run bash -c "cat ${MACHINE_STORAGE_PATH}/machines/$NAME/$NAME/Logs/VBox.log | grep DNSProxy | grep '(1)'"
  [[ ${status} -eq 0 ]]
}