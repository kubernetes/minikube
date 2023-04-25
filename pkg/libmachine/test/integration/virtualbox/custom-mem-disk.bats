#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER virtualbox

use_disposable_machine

# Default memsize is 1024MB and disksize is 20000MB
# These values are defined in drivers/virtualbox/virtualbox.go
export DEFAULT_MEMSIZE=1024
export DEFAULT_DISKSIZE=20000
export CUSTOM_MEMSIZE=1536
export CUSTOM_DISKSIZE=10000
export CUSTOM_CPUCOUNT=1

function findDiskSize() {
  # SATA-0-0 is usually the boot2disk.iso image
  # We assume that SATA 1-0 is root disk VMDK and grab this UUID
  # e.g. "SATA-ImageUUID-1-0"="fb5f33a7-e4e3-4cb9-877c-f9415ae2adea"
  # TODO(slashk): does this work on Windows ?
  run bash -c "VBoxManage showvminfo --machinereadable $NAME | grep SATA-ImageUUID-1-0 | cut -d'=' -f2"
  run bash -c "VBoxManage showhdinfo $output | grep "Capacity:" | awk -F' ' '{ print $2 }'"
}

function findMemorySize() {
  run bash -c "VBoxManage showvminfo --machinereadable $NAME | grep memory= | cut -d'=' -f2"
}

function findCPUCount() {
  run bash -c "VBoxManage showvminfo --machinereadable $NAME | grep cpus= | cut -d'=' -f2"
}

@test "$DRIVER: create with custom disk, cpu count and memory size flags" {
  run machine create -d $DRIVER --virtualbox-cpu-count $CUSTOM_CPUCOUNT --virtualbox-disk-size $CUSTOM_DISKSIZE --virtualbox-memory $CUSTOM_MEMSIZE $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: check custom machine memory size" {
  findMemorySize
  [[ ${output} == "$CUSTOM_MEMSIZE"  ]]
}

@test "$DRIVER: check custom machine disksize" {
  findDiskSize
  [[ ${output} == *"$CUSTOM_DISKSIZE"* ]]
}

@test "$DRIVER: check custom machine cpucount" {
  findCPUCount
  [[ ${output} == "$CUSTOM_CPUCOUNT" ]]
}

@test "$DRIVER: machine should show running after create" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}
