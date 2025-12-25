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

const stdOutDiskInfo = `
storagecontrollerbootable0="on"
"SATA-0-0"="/home/ehazlett/.boot2docker/boot2docker.iso"
"SATA-IsEjected"="off"
"SATA-1-0"="/home/ehazlett/vm/test/disk.vmdk"
"SATA-ImageUUID-1-0"="12345-abcdefg"
"SATA-2-0"="none"
nic1="nat"`

func TestVMDiskInfo(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "showvminfo default --machinereadable",
		stdOut: stdOutDiskInfo,
	}

	disk, err := getVMDiskInfo("default", vbox)

	assert.Equal(t, "/home/ehazlett/vm/test/disk.vmdk", disk.Path)
	assert.Equal(t, "12345-abcdefg", disk.UUID)
	assert.NoError(t, err)
}

func TestVMDiskInfoError(t *testing.T) {
	vbox := &VBoxManagerMock{
		args: "showvminfo default --machinereadable",
		err:  errors.New("BUG"),
	}

	disk, err := getVMDiskInfo("default", vbox)

	assert.Nil(t, disk)
	assert.EqualError(t, err, "BUG")
}

func TestVMDiskInfoInvalidOutput(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "showvminfo default --machinereadable",
		stdOut: "INVALID",
	}

	disk, err := getVMDiskInfo("default", vbox)

	assert.Empty(t, disk.Path)
	assert.Empty(t, disk.UUID)
	assert.NoError(t, err)
}
