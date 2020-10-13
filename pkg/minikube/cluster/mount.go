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

package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
)

// MountConfig defines the options available to the Mount command
type MountConfig struct {
	// Type is the filesystem type (Typically 9p)
	Type string
	// UID is the User ID which this path will be mounted as
	UID string
	// GID is the Group ID which this path will be mounted as
	GID string
	// Version is the 9P protocol version. Valid options: 9p2000, 9p200.u, 9p2000.L
	Version string
	// MSize is the number of bytes to use for 9p packet payload
	MSize int
	// Port is the port to connect to on the host
	Port int
	// Mode is the file permissions to set the mount to (octals)
	Mode os.FileMode
	// Extra mount options. See https://www.kernel.org/doc/Documentation/filesystems/9p.txt
	Options map[string]string
}

// mountRunner is the subset of CommandRunner used for mounting
type mountRunner interface {
	RunCmd(*exec.Cmd) (*command.RunResult, error)
}

// Mount runs the mount command from the 9p client on the VM to the 9p server on the host
func Mount(r mountRunner, source string, target string, c *MountConfig) error {
	if err := Unmount(r, target); err != nil {
		return errors.Wrap(err, "umount")
	}

	if _, err := r.RunCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf("sudo mkdir -m %o -p %s", c.Mode, target))); err != nil {
		return errors.Wrap(err, "create folder pre-mount")
	}

	rr, err := r.RunCmd(exec.Command("/bin/bash", "-c", mntCmd(source, target, c)))
	if err != nil {
		return errors.Wrapf(err, "mount with cmd %s ", rr.Command())
	}

	klog.Infof("mount successful: %q", rr.Output())
	return nil
}

// returns either a raw UID number, or the subshell to resolve it.
func resolveUID(id string) string {
	_, err := strconv.ParseInt(id, 10, 64)
	if err == nil {
		return id
	}
	// Preserve behavior where unset ID == 0
	if id == "" {
		return "0"
	}
	return fmt.Sprintf(`$(id -u %s)`, id)
}

// returns either a raw GID number, or the subshell to resolve it.
func resolveGID(id string) string {
	_, err := strconv.ParseInt(id, 10, 64)
	if err == nil {
		return id
	}
	// Preserve behavior where unset ID == 0
	if id == "" {
		return "0"
	}
	// Because `getent` isn't part of our ISO
	return fmt.Sprintf(`$(grep ^%s: /etc/group | cut -d: -f3)`, id)
}

// mntCmd returns a mount command based on a config.
func mntCmd(source string, target string, c *MountConfig) string {
	options := map[string]string{
		"dfltgid": resolveGID(c.GID),
		"dfltuid": resolveUID(c.UID),
		"trans":   "tcp",
	}

	if c.Port != 0 {
		options["port"] = strconv.Itoa(c.Port)
	}
	if c.Version != "" {
		options["version"] = c.Version
	}
	if c.MSize != 0 {
		options["msize"] = strconv.Itoa(c.MSize)
	}

	// Copy in all of the user-supplied keys and values
	for k, v := range c.Options {
		options[k] = v
	}

	// Convert everything into a sorted list for better test results
	opts := []string{}
	for k, v := range options {
		// Mount option with no value, such as "noextend"
		if v == "" {
			opts = append(opts, k)
			continue
		}
		opts = append(opts, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(opts)
	return fmt.Sprintf("sudo mount -t %s -o %s %s %s", c.Type, strings.Join(opts, ","), source, target)
}

// Unmount unmounts a path
func Unmount(r mountRunner, target string) error {
	// grep because findmnt will also display the parent!
	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("[ \"x$(findmnt -T %s | grep %s)\" != \"x\" ] && sudo umount -f %s || echo ", target, target, target))
	if _, err := r.RunCmd(c); err != nil {
		return errors.Wrap(err, "unmount")
	}
	klog.Infof("unmount for %s ran successfully", target)
	return nil
}
