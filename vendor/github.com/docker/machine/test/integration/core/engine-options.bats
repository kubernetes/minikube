#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

use_disposable_machine

@test "$DRIVER: create with supported engine options" {
  run machine create -d $DRIVER \
    --engine-label spam=eggs \
    --engine-storage-driver overlay \
    --engine-insecure-registry registry.myco.com \
    --engine-env=TEST=VALUE \
    --engine-opt log-driver=none \
    $NAME
  echo "$output"
  [ $status -eq 0 ]
}

@test "$DRIVER: check for engine label" {
  spamlabel=$(docker $(machine config $NAME) info | grep spam)
  [[ $spamlabel =~ "spam=eggs" ]]
}

@test "$DRIVER: check for engine storage driver" {
  storage_driver_info=$(docker $(machine config $NAME) info | grep "Storage Driver")
  [[ $storage_driver_info =~ "overlay" ]]
}

@test "$DRIVER: test docker process envs" {
  # get pid of docker process, check process envs for set Environment Variable from above test
  run machine ssh $NAME 'sudo cat /proc/$(pgrep -f "docker [d]aemon")/environ'
  echo ${output}
  [ $status -eq 0 ]
  [[ "${output}" =~ "TEST=VALUE" ]]
}

@test "$DRIVER: check created engine option (log driver)" {
  docker $(machine config $NAME) run --name nolog busybox echo this should not be logged
  run docker $(machine config $NAME) inspect -f '{{.HostConfig.LogConfig.Type}}' nolog
  echo ${output}
  [ ${output} == "none" ]
}
