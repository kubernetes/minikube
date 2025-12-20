/*
Copyright 2022 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package virtualbox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var stdOutVMInfo = `
storagecontrollerbootable0="on"
memory=1024
cpus=2
"SATA-0-0"="/home/ehazlett/.boot2docker/boot2docker.iso"
"SATA-IsEjected"="off"
"SATA-1-0"="/home/ehazlett/vm/test/disk.vmdk"
"SATA-ImageUUID-1-0"="12345-abcdefg"
"SATA-2-0"="none"
nic1="nat"`

func TestVMInfo(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "showvminfo host --machinereadable",
		stdOut: stdOutVMInfo,
	}

	vm, err := getVMInfo("host", vbox)

	assert.Equal(t, 2, vm.CPUs)
	assert.Equal(t, 1024, vm.Memory)
	assert.NoError(t, err)
}

func TestVMInfoError(t *testing.T) {
	vbox := &VBoxManagerMock{
		args: "showvminfo host --machinereadable",
		err:  errors.New("BUG"),
	}

	vm, err := getVMInfo("host", vbox)

	assert.Nil(t, vm)
	assert.EqualError(t, err, "BUG")
}
