#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

use_shared_machine

@test "$DRIVER: test basic bash / zsh notation" {
  run machine env $NAME
  [[ ${lines[0]} == "export DOCKER_TLS_VERIFY=\"1\"" ]]
  [[ ${lines[1]} == "export DOCKER_HOST=\"$(machine url $NAME)\"" ]]
  [[ ${lines[2]} == "export DOCKER_CERT_PATH=\"$MACHINE_STORAGE_PATH/machines/$NAME\"" ]]
  [[ ${lines[3]} == "export DOCKER_MACHINE_NAME=\"$NAME\"" ]]
}

@test "$DRIVER: test powershell notation" {
  run machine env --shell powershell --no-proxy $NAME
  [[ ${lines[0]} == "\$Env:DOCKER_TLS_VERIFY = \"1\"" ]]
  [[ ${lines[1]} == "\$Env:DOCKER_HOST = \"$(machine url $NAME)\"" ]]
  [[ ${lines[2]} == "\$Env:DOCKER_CERT_PATH = \"$MACHINE_STORAGE_PATH/machines/$NAME\"" ]]
  [[ ${lines[3]} == "\$Env:DOCKER_MACHINE_NAME = \"$NAME\"" ]]
  [[ ${lines[4]} == "\$Env:NO_PROXY = \"$(machine ip $NAME)\"" ]]
}

@test "$DRIVER: test bash / zsh notation with no-proxy" {
  run machine env --no-proxy $NAME
  [[ ${lines[0]} == "export DOCKER_TLS_VERIFY=\"1\"" ]]
  [[ ${lines[1]} == "export DOCKER_HOST=\"$(machine url $NAME)\"" ]]
  [[ ${lines[2]} == "export DOCKER_CERT_PATH=\"$MACHINE_STORAGE_PATH/machines/$NAME\"" ]]
  [[ ${lines[3]} == "export DOCKER_MACHINE_NAME=\"$NAME\"" ]]
  [[ ${lines[4]} == "export NO_PROXY=\"$(machine ip $NAME)\"" ]]
}

@test "$DRIVER: test cmd.exe notation" {
  run machine env --shell cmd --no-proxy $NAME
  [[ ${lines[0]} == "SET DOCKER_TLS_VERIFY=1" ]]
  [[ ${lines[1]} == "SET DOCKER_HOST=$(machine url $NAME)" ]]
  [[ ${lines[2]} == "SET DOCKER_CERT_PATH=$MACHINE_STORAGE_PATH/machines/$NAME" ]]
  [[ ${lines[3]} == "SET DOCKER_MACHINE_NAME=$NAME" ]]
  [[ ${lines[4]} == "SET NO_PROXY=$(machine ip $NAME)" ]]
}

@test "$DRIVER: test fish notation" {
  run machine env --shell fish --no-proxy $NAME
  [[ ${lines[0]} == "set -gx DOCKER_TLS_VERIFY \"1\";" ]]
  [[ ${lines[1]} == "set -gx DOCKER_HOST \"$(machine url $NAME)\";" ]]
  [[ ${lines[2]} == "set -gx DOCKER_CERT_PATH \"$MACHINE_STORAGE_PATH/machines/$NAME\";" ]]
  [[ ${lines[3]} == "set -gx DOCKER_MACHINE_NAME \"$NAME\";" ]]
  [[ ${lines[4]} == "set -gx NO_PROXY \"$(machine ip $NAME)\";" ]]
}

@test "$DRIVER: test emacs notation" {
  run machine env --shell emacs --no-proxy $NAME
  [[ ${lines[0]} == "(setenv \"DOCKER_TLS_VERIFY\" \"1\")" ]]
  [[ ${lines[1]} == "(setenv \"DOCKER_HOST\" \"$(machine url $NAME)\")" ]]
  [[ ${lines[2]} == "(setenv \"DOCKER_CERT_PATH\" \"$MACHINE_STORAGE_PATH/machines/$NAME\")" ]]
  [[ ${lines[3]} == "(setenv \"DOCKER_MACHINE_NAME\" \"$NAME\")" ]]
  [[ ${lines[4]} == "(setenv \"NO_PROXY\" \"$(machine ip $NAME)\")" ]]
}

@test "$DRIVER: test no proxy with NO_PROXY already set" {
  export NO_PROXY=localhost
  run machine env --no-proxy $NAME
  [[ ${lines[4]} == "export NO_PROXY=\"localhost,$(machine ip $NAME)\"" ]]
}

@test "$DRIVER: test unset with an args should fail" {
  run machine env -u $NAME
  [ "$status" -eq 1 ]
  [[ ${lines} == "Error: Expected no machine name when the -u flag is present" ]]
}


@test "$DRIVER: test bash/zsh unset" {
  run machine env -u
  [[ ${lines[0]} == "unset DOCKER_TLS_VERIFY" ]]
  [[ ${lines[1]} == "unset DOCKER_HOST" ]]
  [[ ${lines[2]} == "unset DOCKER_CERT_PATH" ]]
  [[ ${lines[3]} == "unset DOCKER_MACHINE_NAME" ]]
}

@test "$DRIVER: test unset killing no proxy" {
  run machine env --no-proxy -u
  [[ ${lines[0]} == "unset DOCKER_TLS_VERIFY" ]]
  [[ ${lines[1]} == "unset DOCKER_HOST" ]]
  [[ ${lines[2]} == "unset DOCKER_CERT_PATH" ]]
  [[ ${lines[3]} == "unset DOCKER_MACHINE_NAME" ]]
  [[ ${lines[4]} == "unset NO_PROXY" ]]
}

@test "$DRIVER: unset powershell" {
  run machine env --shell powershell -u
  [[ ${lines[0]} == 'Remove-Item Env:\\DOCKER_TLS_VERIFY' ]]
  [[ ${lines[1]} == 'Remove-Item Env:\\DOCKER_HOST' ]]
  [[ ${lines[2]} == 'Remove-Item Env:\\DOCKER_CERT_PATH' ]]
  [[ ${lines[3]} == 'Remove-Item Env:\\DOCKER_MACHINE_NAME' ]]
}

@test "$DRIVER: unset with fish shell" {
  run machine env --shell fish -u
  [[ ${lines[0]} == "set -e DOCKER_TLS_VERIFY;" ]]
  [[ ${lines[1]} == "set -e DOCKER_HOST;" ]]
  [[ ${lines[2]} == "set -e DOCKER_CERT_PATH;" ]]
  [[ ${lines[3]} == "set -e DOCKER_MACHINE_NAME;" ]]
}

@test "$DRIVER: unset with cmd shell" {
  run machine env --shell cmd -u
  [[ ${lines[0]} == "SET DOCKER_TLS_VERIFY=" ]]
  [[ ${lines[1]} == "SET DOCKER_HOST=" ]]
  [[ ${lines[2]} == "SET DOCKER_CERT_PATH=" ]]
  [[ ${lines[3]} == "SET DOCKER_MACHINE_NAME=" ]]
}

@test "$DRIVER: unset with emacs shell" {
  run machine env --shell emacs -u
  [[ ${lines[0]} == "(setenv \"DOCKER_TLS_VERIFY\" nil)" ]]
  [[ ${lines[1]} == "(setenv \"DOCKER_HOST\" nil)" ]]
  [[ ${lines[2]} == "(setenv \"DOCKER_CERT_PATH\" nil)" ]]
  [[ ${lines[3]} == "(setenv \"DOCKER_MACHINE_NAME\" nil)" ]]
}
