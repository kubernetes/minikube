package commands

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMountCmd(t *testing.T) {
	hostInfoLoader := MockHostInfoLoader{MockHostInfo{
		ip:          "12.34.56.78",
		sshPort:     234,
		sshUsername: "root",
		sshKeyPath:  "/fake/keypath/id_rsa",
	}}

	path, err := exec.LookPath("sshfs")
	if err != nil {
		t.Skip("sshfs not found (install sshfs ?)")
	}
	cmd, err := getMountCmd("myfunhost:/home/docker/foo", "/tmp/foo", false, &hostInfoLoader)

	expectedArgs := append(
		baseSSHFSArgs,
		"-o",
		"IdentitiesOnly=yes",
		"-o",
		"Port=234",
		"-o",
		"IdentityFile=/fake/keypath/id_rsa",
		"root@12.34.56.78:/home/docker/foo",
		"/tmp/foo",
	)
	expectedCmd := exec.Command(path, expectedArgs...)

	assert.Equal(t, expectedCmd, cmd)
	assert.NoError(t, err)
}

func TestGetMountCmdWithoutSshKey(t *testing.T) {
	hostInfoLoader := MockHostInfoLoader{MockHostInfo{
		ip:          "1.2.3.4",
		sshUsername: "user",
	}}

	path, err := exec.LookPath("sshfs")
	if err != nil {
		t.Skip("sshfs not found (install sshfs ?)")
	}
	cmd, err := getMountCmd("myfunhost:/home/docker/foo", "", false, &hostInfoLoader)

	expectedArgs := append(
		baseSSHFSArgs,
		"user@1.2.3.4:/home/docker/foo",
		"/home/docker/foo",
	)
	expectedCmd := exec.Command(path, expectedArgs...)

	assert.Equal(t, expectedCmd, cmd)
	assert.NoError(t, err)
}

func TestGetMountCmdUnmount(t *testing.T) {
	hostInfoLoader := MockHostInfoLoader{MockHostInfo{
		ip:          "1.2.3.4",
		sshUsername: "user",
	}}

	path, err := exec.LookPath("fusermount")
	if err != nil {
		t.Skip("fusermount not found (install fuse ?)")
	}
	cmd, err := getMountCmd("myfunhost:/home/docker/foo", "/tmp/foo", true, &hostInfoLoader)

	expectedArgs := []string{
		"-u",
		"/tmp/foo",
	}
	expectedCmd := exec.Command(path, expectedArgs...)

	assert.Equal(t, expectedCmd, cmd)
	assert.NoError(t, err)
}
