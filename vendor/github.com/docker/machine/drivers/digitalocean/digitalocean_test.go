package digitalocean

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"digitalocean-access-token": "TOKEN",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)

	assert.Equal(t, driver.ResolveStorePath("id_rsa"), driver.GetSSHKeyPath())
}

func TestDefaultSSHUserAndPort(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"digitalocean-access-token": "TOKEN",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)

	sshPort, err := driver.GetSSHPort()
	assert.Equal(t, "root", driver.GetSSHUsername())
	assert.Equal(t, 22, sshPort)
	assert.NoError(t, err)
}

func TestCustomSSHUserAndPort(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"digitalocean-access-token": "TOKEN",
			"digitalocean-ssh-user":     "user",
			"digitalocean-ssh-port":     2222,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)

	sshPort, err := driver.GetSSHPort()
	assert.Equal(t, "user", driver.GetSSHUsername())
	assert.Equal(t, 2222, sshPort)
	assert.NoError(t, err)
}

func TestSSHKeyFingerprint(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"digitalocean-access-token":        "TOKEN",
			"digitalocean-ssh-key-fingerprint": "64:51:2b:9b:8b:f0:95:3c:f9:36:4d:8b:80:a8:8f:1e",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)

	assert.Equal(t, "64:51:2b:9b:8b:f0:95:3c:f9:36:4d:8b:80:a8:8f:1e", driver.SSHKeyFingerprint)
	assert.Equal(t, "", driver.GetSSHKeyPath())
}
