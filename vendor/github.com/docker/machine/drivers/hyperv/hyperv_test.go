package hyperv

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromDefaultFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)

	sshPort, err := driver.GetSSHPort()
	assert.Equal(t, 22, sshPort)
	assert.NoError(t, err)

	assert.Equal(t, "", driver.Boot2DockerURL)
	assert.Equal(t, "", driver.VSwitch)
	assert.Equal(t, defaultDiskSize, driver.DiskSize)
	assert.Equal(t, defaultMemory, driver.MemSize)
	assert.Equal(t, defaultCPU, driver.CPU)
	assert.Equal(t, "", driver.MacAddr)
	assert.Equal(t, defaultVLanID, driver.VLanID)
	assert.Equal(t, "docker", driver.GetSSHUsername())
}

func TestSetConfigFromCustomFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"hyperv-boot2docker-url":   "B2D_URL",
			"hyperv-virtual-switch":    "TheSwitch",
			"hyperv-disk-size":         100000,
			"hyperv-memory":            4096,
			"hyperv-cpu-count":         4,
			"hyperv-static-macaddress": "00:0a:95:9d:68:16",
			"hyperv-vlan-id":           2,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)

	sshPort, err := driver.GetSSHPort()
	assert.Equal(t, 22, sshPort)
	assert.NoError(t, err)

	assert.Equal(t, "B2D_URL", driver.Boot2DockerURL)
	assert.Equal(t, "TheSwitch", driver.VSwitch)
	assert.Equal(t, 100000, driver.DiskSize)
	assert.Equal(t, 4096, driver.MemSize)
	assert.Equal(t, 4, driver.CPU)
	assert.Equal(t, "00:0a:95:9d:68:16", driver.MacAddr)
	assert.Equal(t, 2, driver.VLanID)
	assert.Equal(t, "docker", driver.GetSSHUsername())
}
