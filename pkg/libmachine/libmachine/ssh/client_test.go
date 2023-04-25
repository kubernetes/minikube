package ssh

import (
	"fmt"
	"io/ioutil"
	"os"
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
	keyFile, err := ioutil.TempFile("", "docker-machine-tests-dummy-private-key")
	if err != nil {
		t.Fatal(err)
	}
	defer keyFile.Close()

	keyFilename := keyFile.Name()
	defer os.Remove(keyFilename)

	cases := []struct {
		sshBinaryPath string
		user          string
		host          string
		port          int
		auth          *Auth
		perm          os.FileMode
		expectedError string
		skipOS        string
	}{
		{
			auth:          &Auth{Keys: []string{"/tmp/private-key-not-exist"}},
			expectedError: "stat /tmp/private-key-not-exist: no such file or directory",
			skipOS:        "none",
		},
		{
			auth:   &Auth{Keys: []string{keyFilename}},
			perm:   0400,
			skipOS: "windows",
		},
		{
			auth:          &Auth{Keys: []string{keyFilename}},
			perm:          0100,
			expectedError: fmt.Sprintf("'%s' is not readable", keyFilename),
			skipOS:        "windows",
		},
		{
			auth:          &Auth{Keys: []string{keyFilename}},
			perm:          0644,
			expectedError: fmt.Sprintf("permissions 0644 for '%s' are too open", keyFilename),
			skipOS:        "windows",
		},
	}

	for _, c := range cases {
		if runtime.GOOS != c.skipOS {
			keyFile.Chmod(c.perm)
			_, err := NewExternalClient(c.sshBinaryPath, c.user, c.host, c.port, c.auth)
			if c.expectedError != "" {
				assert.EqualError(t, err, c.expectedError)
			} else {
				assert.Equal(t, err, nil)
			}
		}
	}
}
