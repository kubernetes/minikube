#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

use_shared_machine

@test "$DRIVER: test external ssh backend" {
  run machine ssh $NAME df -h
  [[ "$status" -eq 0 ]]
}

@test "$DRIVER: test command did what it purported to -- external ssh" {
  run machine ssh $NAME echo foo
  [[ "$output" == "foo"  ]]
}

@test "$DRIVER: test native ssh backend" {
  run machine --native-ssh ssh $NAME df -h
  [[ "$status" -eq 0  ]]
}

@test "$DRIVER: test command did what it purported to -- native ssh" {
  run machine --native-ssh ssh $NAME echo foo
  [[ "$output" =~ "foo"  ]]
}

@test "$DRIVER: ensure that ssh extra arguments work" {
  # don't run this test if we can't use external SSH
  which ssh
  if [[ $? -ne 0 ]]; then
    skip
  fi

  # this will not fare well if -C doesn't get interpreted as "use compression"
  # like intended
  run machine ssh $NAME -C echo foo

  [[ "$status" -eq 0 ]]
  [[ "$output" == "foo" ]]
}
