//Package provisiontest provides utilities for testing provisioners
package provisiontest

import "errors"

//FakeSSHCommanderOptions is intended to create a FakeSSHCommander without actually knowing the underlying sshcommands by passing it to NewSSHCommander
type FakeSSHCommanderOptions struct {
	//Result of the ssh command to look up the FilesystemType
	FilesystemType string
}

//FakeSSHCommander is an implementation of provision.SSHCommander to provide predictable responses set by testing code
//Extend it when needed
type FakeSSHCommander struct {
	Responses map[string]string
}

//NewFakeSSHCommander creates a FakeSSHCommander without actually knowing the underlying sshcommands
func NewFakeSSHCommander(options FakeSSHCommanderOptions) *FakeSSHCommander {
	if options.FilesystemType == "" {
		options.FilesystemType = "ext4"
	}
	sshCmder := &FakeSSHCommander{
		Responses: map[string]string{
			"stat -f -c %T /var/lib": options.FilesystemType + "\n",
		},
	}

	return sshCmder
}

//SSHCommand is an implementation of provision.SSHCommander.SSHCommand to provide predictable responses set by testing code
func (sshCmder *FakeSSHCommander) SSHCommand(args string) (string, error) {
	response, commandRegistered := sshCmder.Responses[args]
	if !commandRegistered {
		return "", errors.New("Command not registered in FakeSSHCommander")
	}
	return response, nil
}
