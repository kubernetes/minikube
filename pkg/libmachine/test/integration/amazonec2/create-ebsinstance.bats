#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER amazonec2

use_disposable_machine

require_env AWS_ACCESS_KEY_ID
require_env AWS_SECRET_ACCESS_KEY
require_env AWS_VPC_ID

require_env AWS_DEFAULT_REGION
require_env AWS_ZONE

@test "$DRIVER: Should Create an EBS Optimized Instance" {
    #Use Instance Type that supports EBS Optimize
    run machine create -d amazonec2 --amazonec2-instance-type=m4.large --amazonec2-use-ebs-optimized-instance $NAME
    echo ${output}
    [ "$status" -eq 0 ]
}

@test "$DRIVER: Check the machine is up" {
    run docker $(machine config $NAME) run --rm -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e AWS_DEFAULT_REGION=$AWS_DEFAULT_REGION -e AWS_ZONE=$AWS_ZONE -e AWS_VPC_ID=$AWS_VPC_ID blendle/aws-cli ec2 describe-instances --filters Name=tag:Name,Values=$NAME Name=instance-state-name,Values=running --query 'Reservations[0].Instances[0].EbsOptimized' --output text
    echo ${output}
    [[ ${lines[*]:-1} =~ "True" ]]
}