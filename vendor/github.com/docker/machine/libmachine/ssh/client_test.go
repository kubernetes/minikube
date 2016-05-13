package ssh

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSSHCmdArgs(t *testing.T) {
	cases := []struct {
		binaryPath   string
		args         []string
		expectedArgs []string
	}{
		{
			binaryPath: "/usr/local/bin/ssh",
			args: []string{
				"docker@localhost",
				"apt-get install -y htop",
			},
			expectedArgs: []string{
				"/usr/local/bin/ssh",
				"docker@localhost",
				"apt-get install -y htop",
			},
		},
		{
			binaryPath: "C:\\Program Files\\Git\\bin\\ssh.exe",
			args: []string{
				"docker@localhost",
				"sudo /usr/bin/sethostname foobar && echo 'foobar' | sudo tee /var/lib/boot2docker/etc/hostname",
			},
			expectedArgs: []string{
				"C:\\Program Files\\Git\\bin\\ssh.exe",
				"docker@localhost",
				"sudo /usr/bin/sethostname foobar && echo 'foobar' | sudo tee /var/lib/boot2docker/etc/hostname",
			},
		},
	}

	for _, c := range cases {
		cmd := getSSHCmd(c.binaryPath, c.args...)
		assert.Equal(t, cmd.Args, c.expectedArgs)
	}
}

func TestNewExternalClient(t *testing.T) {
	cases := []struct {
		sshBinaryPath string
		user          string
		host          string
		port          int
		auth          *Auth
		expectedError string
		skipOS        string
	}{
		{
			sshBinaryPath: "/usr/local/bin/ssh",
			user:          "docker",
			host:          "localhost",
			port:          22,
			auth:          &Auth{Keys: []string{"/tmp/private-key-not-exist"}},
			expectedError: "stat /tmp/private-key-not-exist: no such file or directory",
			skipOS:        "none",
		},
		{
			sshBinaryPath: "/usr/local/bin/ssh",
			user:          "docker",
			host:          "localhost",
			port:          22,
			auth:          &Auth{Keys: []string{"/dev/null"}},
			expectedError: "Permissions 0410000666 for '/dev/null' are too open.",
			skipOS:        "windows",
		},
	}

	for _, c := range cases {
		if runtime.GOOS != c.skipOS {
			_, err := NewExternalClient(c.sshBinaryPath, c.user, c.host, c.port, c.auth)
			assert.EqualError(t, err, c.expectedError)
		}
	}
}
