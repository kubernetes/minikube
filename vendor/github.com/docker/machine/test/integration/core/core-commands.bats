#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

use_shared_machine

@test "$DRIVER: machine should not exist" {
  run machine inspect UNKNOWN
  echo ${output}
  [ "$status" -eq 1 ]
  [[ ${lines[0]} =~ "Host does not exist: \"UNKNOWN\"" ]]
}

@test "$DRIVER: appears with ls" {
  run machine ls -q
  echo ${output}
  [ "$status" -eq 0  ]
  [[ ${lines[0]} == "$NAME" ]]
}

@test "$DRIVER: has status 'started' appearing in ls" {
  run machine ls -q --filter state=Running
  echo ${output}
  [ "$status" -eq 0  ]
  [[ ${lines[0]} == "$NAME" ]]
}

@test "$DRIVER: create with same name again fails" {
  run machine create -d $DRIVER $NAME
  echo ${output}
  [ "$status" -eq 1  ]
  [[ ${lines[0]} == "Host already exists: \"$NAME\"" ]]
}

@test "$DRIVER: run busybox container" {
  run docker $(machine config $NAME) run busybox echo hello world
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: url" {
  run machine url $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ip" {
  run machine ip $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ssh" {
  run machine ssh $NAME -- ls -lah /
  echo ${output}
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "total"  ]]
}

@test "$DRIVER: version" {
  run machine version $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: docker commands with the socket should work" {
  run machine ssh $NAME -- sudo docker version
  echo ${output}
}

@test "$DRIVER: stop" {
  run machine stop $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show stopped after stop" {
  run machine ls
  echo ${output}
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Stopped"*  ]]
}

@test "$DRIVER: url should show an error when machine is stopped" {
  run machine url $NAME
  echo ${output}
  [ "$status" -eq 1 ]
  [[ ${output} == *"Host is not running"* ]]
}

@test "$DRIVER: env should show an error when machine is stopped" {
  run machine env $NAME
  echo ${output}
  [ "$status" -eq 1 ]
  [[ ${output} == *"Host is not running"* ]]
}

@test "$DRIVER: version should show an error when machine is stopped" {
  run machine version $NAME
  echo ${output}
  [ "$status" -eq 1 ]
  [[ ${output} == *"Host is not running"* ]]
}


@test "$DRIVER: machine should not allow upgrade when stopped" {
  run machine upgrade $NAME
  echo ${output}
  [[ "$status" -eq 1 ]]
}

@test "$DRIVER: start" {
  run machine start $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after start" {
  run machine ls --timeout 20
  echo ${output}
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: kill" {
  run machine kill $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show stopped after kill" {
  run machine ls
  echo ${output}
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Stopped"*  ]]
}

@test "$DRIVER: restart" {
  run machine restart $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after restart" {
  run machine ls --timeout 20
  echo ${output}
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: status" {
  run machine status $NAME
  echo ${output}
  [ "$status" -eq 0 ]
  [[ ${output} == *"Running"* ]]
}
