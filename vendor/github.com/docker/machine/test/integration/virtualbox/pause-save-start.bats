#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER virtualbox

use_shared_machine

@test "$DRIVER: VBoxManage pause" {
  run VBoxManage controlvm $NAME pause
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show paused after VBoxManage pause" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Paused"*  ]]
}

@test "$DRIVER: start after paused" {
  run machine start $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after start" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: VBoxManage savestate" {
  run VBoxManage controlvm $NAME savestate
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show saved after VBoxManage savestate" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"$NAME"*  ]]
  [[ ${lines[1]} == *"Saved"*  ]]
}

@test "$DRIVER: start after saved" {
  run machine start $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after start" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}
