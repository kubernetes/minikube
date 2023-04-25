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
