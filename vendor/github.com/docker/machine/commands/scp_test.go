package commands

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockHostInfo struct {
	name        string
	ip          string
	sshUsername string
	sshKeyPath  string
}

func (h *MockHostInfo) GetMachineName() string {
	return h.name
}

func (h *MockHostInfo) GetIP() (string, error) {
	return h.ip, nil
}

func (h *MockHostInfo) GetSSHUsername() string {
	return h.sshUsername
}

func (h *MockHostInfo) GetSSHKeyPath() string {
	return h.sshKeyPath
}

type MockHostInfoLoader struct {
	hostInfo MockHostInfo
}

func (l *MockHostInfoLoader) load(name string) (HostInfo, error) {
	info := l.hostInfo
	info.name = name
	return &info, nil
}

func TestGetInfoForLocalScpArg(t *testing.T) {
	host, path, opts, err := getInfoForScpArg("/tmp/foo", nil)
	assert.Nil(t, host)
	assert.Equal(t, "/tmp/foo", path)
	assert.Nil(t, opts)
	assert.NoError(t, err)

	host, path, opts, err = getInfoForScpArg("localhost:C:\\path", nil)
	assert.Nil(t, host)
	assert.Equal(t, "C:\\path", path)
	assert.Nil(t, opts)
	assert.NoError(t, err)
}

func TestGetInfoForRemoteScpArg(t *testing.T) {
	hostInfoLoader := MockHostInfoLoader{MockHostInfo{
		sshKeyPath: "/fake/keypath/id_rsa",
	}}

	host, path, opts, err := getInfoForScpArg("myfunhost:/home/docker/foo", &hostInfoLoader)
	assert.Equal(t, "myfunhost", host.GetMachineName())
	assert.Equal(t, "/home/docker/foo", path)
	assert.Equal(t, []string{"-i", "/fake/keypath/id_rsa"}, opts)
	assert.NoError(t, err)

	host, path, opts, err = getInfoForScpArg("myfunhost:C:\\path", &hostInfoLoader)
	assert.Equal(t, "myfunhost", host.GetMachineName())
	assert.Equal(t, "C:\\path", path)
	assert.NoError(t, err)
}

func TestHostLocation(t *testing.T) {
	arg, err := generateLocationArg(nil, "/home/docker/foo")

	assert.Equal(t, "/home/docker/foo", arg)
	assert.NoError(t, err)
}

func TestRemoteLocation(t *testing.T) {
	hostInfo := MockHostInfo{
		ip:          "12.34.56.78",
		sshUsername: "root",
	}

	arg, err := generateLocationArg(&hostInfo, "/home/docker/foo")

	assert.Equal(t, "root@12.34.56.78:/home/docker/foo", arg)
	assert.NoError(t, err)
}

func TestGetScpCmd(t *testing.T) {
	hostInfoLoader := MockHostInfoLoader{MockHostInfo{
		ip:          "12.34.56.78",
		sshUsername: "root",
		sshKeyPath:  "/fake/keypath/id_rsa",
	}}

	cmd, err := getScpCmd("/tmp/foo", "myfunhost:/home/docker/foo", true, &hostInfoLoader)

	expectedArgs := append(
		baseSSHArgs,
		"-3",
		"-r",
		"-i",
		"/fake/keypath/id_rsa",
		"/tmp/foo",
		"root@12.34.56.78:/home/docker/foo",
	)
	expectedCmd := exec.Command("/usr/bin/scp", expectedArgs...)

	assert.Equal(t, expectedCmd, cmd)
	assert.NoError(t, err)
}

func TestGetScpCmdWithoutSshKey(t *testing.T) {
	hostInfoLoader := MockHostInfoLoader{MockHostInfo{
		ip:          "1.2.3.4",
		sshUsername: "user",
	}}

	cmd, err := getScpCmd("/tmp/foo", "myfunhost:/home/docker/foo", true, &hostInfoLoader)

	expectedArgs := append(
		baseSSHArgs,
		"-3",
		"-r",
		"/tmp/foo",
		"user@1.2.3.4:/home/docker/foo",
	)
	expectedCmd := exec.Command("/usr/bin/scp", expectedArgs...)

	assert.Equal(t, expectedCmd, cmd)
	assert.NoError(t, err)
}
