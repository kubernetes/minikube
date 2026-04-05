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

package provisiontest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFakeSSHCommander(t *testing.T) {
	sshCmder := NewFakeSSHCommander(FakeSSHCommanderOptions{FilesystemType: "btrfs"})
	output, err := sshCmder.SSHCommand("stat -f -c %T /var/lib")
	if err != nil || output != "btrfs\n" {
		t.Fatal("FakeSSHCommander should have returned btrfs and no error but returned '", output, "' and error", err)
	}
}

func TestStatSSHCommand(t *testing.T) {
	sshCmder := FakeSSHCommander{
		Responses: map[string]string{"sshcommand": "sshcommandresponse"},
	}

	output, err := sshCmder.SSHCommand("sshcommand")
	assert.NoError(t, err)
	assert.Equal(t, "sshcommandresponse", output)

	_, err = sshCmder.SSHCommand("errorcommand")
	assert.Error(t, err)
}
