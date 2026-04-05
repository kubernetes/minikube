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
		sshBinaryPath    string
		user             string
		host             string
		port             int
		auth             *Auth
		perm             os.FileMode
		expectedError    string
		expectedNotExist bool
		skipOS           string
	}{
		{
			auth:             &Auth{Keys: []string{"/tmp/private-key-not-exist"}},
			expectedError:    "stat /tmp/private-key-not-exist: no such file or directory",
			expectedNotExist: true,
			skipOS:           "none",
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
			assert.NoError(t, keyFile.Chmod(c.perm))
			_, err := NewExternalClient(c.sshBinaryPath, c.user, c.host, c.port, c.auth)
			if c.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error %q but got nil", c.expectedError)
				}
				if c.expectedNotExist {
					assert.True(t, os.IsNotExist(err), "expected a not-exist error but got: %v", err)
				} else {
					assert.EqualError(t, err, c.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		}
	}
}
