#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER virtualbox

use_disposable_machine

export CACHE_DIR="$MACHINE_STORAGE_PATH/cache"
export ISO_PATH="$CACHE_DIR/boot2docker.iso"
export OLD_ISO_URL="https://github.com/boot2docker/boot2docker/releases/download/v1.4.1/boot2docker.iso"

@test "$DRIVER: download the old version iso" {
  run mkdir -p $CACHE_DIR
  run curl $OLD_ISO_URL -L -o $ISO_PATH
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: create with upgrading" {
  run machine create -d $DRIVER $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: create is correct version" {
  SERVER_VERSION=$(docker $(machine config $NAME) version | grep 'Server version' | awk '{ print $3; }')
  [[ "$SERVER_VERSION" != "1.4.1" ]]
}
