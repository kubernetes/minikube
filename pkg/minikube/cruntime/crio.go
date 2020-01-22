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

package cruntime

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/out"
)

const (
	// CRIOConfFile is the path to the CRI-O configuration
	crioConfigFile = "/etc/crio/crio.conf"
)

// CRIO contains CRIO runtime state
type CRIO struct {
	Socket            string
	Runner            CommandRunner
	ImageRepository   string
	KubernetesVersion string
}

// generateCRIOConfig sets up /etc/crio/crio.conf
func generateCRIOConfig(cr CommandRunner, imageRepository string) error {
	cPath := crioConfigFile
	pauseImage := images.Pause(imageRepository)

	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("sudo sed -e 's|^pause_image = .*$|pause_image = \"%s\"|' -i %s", pauseImage, cPath))
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "generateCRIOConfig.")
	}
	return nil
}

// Name is a human readable name for CRIO
func (r *CRIO) Name() string {
	return "CRI-O"
}

// Style is the console style for CRIO
func (r *CRIO) Style() out.StyleEnum {
	return out.CRIO
}

// Version retrieves the current version of this runtime
func (r *CRIO) Version() (string, error) {
	c := exec.Command("crio", "--version")
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return "", errors.Wrap(err, "crio version.")
	}

	// crio version 1.13.0
	// commit: ""
	line := strings.Split(rr.Stdout.String(), "\n")[0]
	return strings.Replace(line, "crio version ", "", 1), nil
}

// SocketPath returns the path to the socket file for CRIO
func (r *CRIO) SocketPath() string {
	if r.Socket != "" {
		return r.Socket
	}
	return "/var/run/crio/crio.sock"
}

// DefaultCNI returns whether to use CNI networking by default
func (r *CRIO) DefaultCNI() bool {
	return true
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *CRIO) Available() error {
	c := exec.Command("which", "crio")
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrapf(err, "check crio available.")
	}
	return nil

}

// Active returns if CRIO is active on the host
func (r *CRIO) Active() bool {
	c := exec.Command("systemctl", "is-active", "--quiet", "service", "crio")
	_, err := r.Runner.RunCmd(c)
	return err == nil
}

// Enable idempotently enables CRIO on a host
func (r *CRIO) Enable(disOthers bool) error {
	if disOthers {
		if err := disableOthers(r, r.Runner); err != nil {
			glog.Warningf("disableOthers: %v", err)
		}
	}
	if err := populateCRIConfig(r.Runner, r.SocketPath()); err != nil {
		return err
	}
	if err := generateCRIOConfig(r.Runner, r.ImageRepository); err != nil {
		return err
	}
	if err := enableIPForwarding(r.Runner); err != nil {
		return err
	}

	if _, err := r.Runner.RunCmd(exec.Command("sudo", "systemctl", "restart", "crio")); err != nil {
		return errors.Wrapf(err, "enable crio.")
	}
	return nil
}

// Disable idempotently disables CRIO on a host
func (r *CRIO) Disable() error {
	if _, err := r.Runner.RunCmd(exec.Command("sudo", "systemctl", "stop", "crio")); err != nil {
		return errors.Wrapf(err, "disable crio.")
	}
	return nil
}

// ImageExists checks if an image exists
func (r *CRIO) ImageExists(name string, sha string) bool {
	// expected output looks like [NAME@sha256:SHA]
	c := exec.Command("sudo", "podman", "inspect", "--format='{{.Id}}'", name)
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return false
	}
	if !strings.Contains(rr.Output(), sha) {
		return false
	}
	return true
}

// LoadImage loads an image into this runtime
func (r *CRIO) LoadImage(path string) error {
	glog.Infof("Loading image: %s", path)
	c := exec.Command("sudo", "podman", "load", "-i", path)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "crio load image")
	}
	return nil
}

// CGroupDriver returns cgroup driver ("cgroupfs" or "systemd")
func (r *CRIO) CGroupDriver() (string, error) {
	c := exec.Command("crio", "config")
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return "", err
	}
	cgroupManager := "cgroupfs" // default
	for _, line := range strings.Split(rr.Stdout.String(), "\n") {
		if strings.HasPrefix(line, "cgroup_manager") {
			// cgroup_manager = "cgroupfs"
			f := strings.Split(strings.TrimSpace(line), " = ")
			if len(f) == 2 {
				cgroupManager = strings.Trim(f[1], "\"")
			}
		}
	}
	return cgroupManager, nil
}

// KubeletOptions returns kubelet options for a runtime.
func (r *CRIO) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime":          "remote",
		"container-runtime-endpoint": r.SocketPath(),
		"image-service-endpoint":     r.SocketPath(),
		"runtime-request-timeout":    "15m",
	}
}

// ListContainers returns a list of managed by this container runtime
func (r *CRIO) ListContainers(o ListOptions) ([]string, error) {
	return listCRIContainers(r.Runner, "", o)
}

// PauseContainers pauses a running container based on ID
func (r *CRIO) PauseContainers(ids []string) error {
	return pauseCRIContainers(r.Runner, "", ids)
}

// PauseContainers pauses a running container based on ID
func (r *CRIO) UnpauseContainers(ids []string) error {
	return unpauseCRIContainers(r.Runner, "", ids)
}

// KillContainers removes containers based on ID
func (r *CRIO) KillContainers(ids []string) error {
	return killCRIContainers(r.Runner, ids)
}

// StopContainers stops containers based on ID
func (r *CRIO) StopContainers(ids []string) error {
	return stopCRIContainers(r.Runner, ids)
}

// ContainerLogCmd returns the command to retrieve the log for a container based on ID
func (r *CRIO) ContainerLogCmd(id string, len int, follow bool) string {
	return criContainerLogCmd(r.Runner, id, len, follow)
}

// SystemLogCmd returns the command to retrieve system logs
func (r *CRIO) SystemLogCmd(len int) string {
	return fmt.Sprintf("sudo journalctl -u crio -n %d", len)
}
