#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER amazonec2

use_disposable_machine

require_env AWS_ACCESS_KEY_ID
require_env AWS_SECRET_ACCESS_KEY

@test "$DRIVER: Should Create a default host" {
    run machine create -d amazonec2 $NAME
    echo ${output}
    [ "$status" -eq 0 ]
}
