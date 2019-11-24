/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package mount

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/command"
	"os/exec"
	"os"
)

var (
	ErrNotImplemented 	=	errors.New("This type of Mount is not available on this system")
)


// MountManager is the common interface which is used for Mounting across various operating systems
type Manager interface {

	// Share is used to share the folder on the host.
	Share() error

	// Unshare is used to unshare the folder on the host
	Unshare() error

	// Mount is used to create the link between minikube and the host
	Mount(runner mountRunner) error

	// Unmount is used to destroy the link between minikube and the host
	Unmount(runner mountRunner) error

}

// CommandRunner is the subset of command.Runner this package consumes
type CommandRunner interface {
	Run(string) error
	CombinedOutput(string) (string, error)
}

// mountRunner is the subset of CommandRunner used for mounting
type mountRunner interface {
	RunCmd(*exec.Cmd) (*command.RunResult, error)
}

// MountConfig defines the options available to the Mount command
type Config struct {
	// Type is the filesystem type (Typically 9p)
	Type string
	// UID is the User ID which this path will be mounted as
	UID string
	// GID is the Group ID which this path will be mounted as
	GID string
	// Path on the Host Machine to Mount
	HostPath string
	// The path on minikube where the Share would be mounted
	VmDestinationPath string
	// Version is the 9P protocol version. Valid options: 9p2000, 9p200.u, 9p2000.L
	Version string
	// Mode is the file permissions to set the mount to (octals)
	Mode os.FileMode
	// Extra mount options. See https://www.kernel.org/doc/Documentation/filesystems/9p.txt
	Options map[string]string
}

type Cifs struct {
	// NotImplemented -- UID is the User ID which this path will be mounted as
	UID string
	// NotImplemented -- GID is the Group ID which this path will be mounted as
	GID string
	// Name of the Share
	HostShareName string
	// Path on the Host Machine to Mount
	HostPath string
	// The path on minikube where the Share would be mounted
	VmDestinationPath string
	// NotImplemented -- Version is the SMB protocol version. Valid options:
	Version string
	// NotImplemented -- Mode is the file permissions to set the mount to (octals)
	Mode os.FileMode
	// NotImplemented -- Extra mount options. See https://linux.die.net/man/8/mount.cifs
	Options map[string]string
}


func New(m Config) (Manager, error) {
	switch m.Type {
	case "cifs":
		var shareName = "minikube"
		return &Cifs{
			UID:               m.UID,
			GID:               m.GID,
			HostShareName:     shareName,
			HostPath:          m.HostPath,
			VmDestinationPath: m.VmDestinationPath,
			Version:           m.Version,
			Mode:              m.Mode,
			Options:           m.Options,
		}, nil
	default:
		return nil, ErrNotImplemented
	}
}

// umountCmd returns a command for unmounting
func umountCmd(target string) string {
	// grep because findmnt will also display the parent!
	return fmt.Sprintf("[ \"x$(findmnt -T %s | grep %s)\" != \"x\" ] && sudo umount -f %s || echo ", target, target, target)
}