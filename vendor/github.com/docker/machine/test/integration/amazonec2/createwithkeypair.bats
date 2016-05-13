#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER amazonec2

use_disposable_machine

require_env AWS_ACCESS_KEY_ID

require_env AWS_SECRET_ACCESS_KEY

export AWS_SSH_DIR="$MACHINE_STORAGE_PATH/mcnkeys"

export AWS_SSH_KEYPATH=$AWS_SSH_DIR/id_rsa

@test "$DRIVER: Should Create Instance with Pre existing SSH Key" {

	mkdir -p $AWS_SSH_DIR

	run ssh-keygen -f $AWS_SSH_KEYPATH -t rsa -N ''

	machine create -d amazonec2 $NAME
	
	run diff $AWS_SSH_KEYPATH $MACHINE_STORAGE_PATH/machines/$NAME/id_rsa
	[[ $output == "" ]]

	run diff $AWS_SSH_KEYPATH.pub $MACHINE_STORAGE_PATH/machines/$NAME/id_rsa.pub
	[[ $output == "" ]]


}