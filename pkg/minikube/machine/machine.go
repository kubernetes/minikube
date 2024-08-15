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

package machine

import (
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	libprovision "github.com/docker/machine/libmachine/provision"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/provision"
	"k8s.io/minikube/pkg/util/retry"
)

// Machine contains information about a machine
type Machine struct {
	*host.Host
}

// IsValid checks if the machine has the essential info needed for a machine
func (h *Machine) IsValid() bool {
	if h == nil {
		return false
	}

	if h.Host == nil {
		return false
	}

	if h.Host.Name == "" {
		return false
	}

	if h.Host.Driver == nil {
		return false
	}

	if h.Host.HostOptions == nil {
		return false
	}

	if h.Host.RawDriver == nil {
		return false
	}
	return true
}

// LoadMachine returns a Machine abstracting a libmachine.Host
func LoadMachine(name string) (*Machine, error) {
	api, err := NewAPIClient()
	if err != nil {
		return nil, err
	}

	h, err := LoadHost(api, name)
	if err != nil {
		return nil, err
	}

	var mm Machine
	if h != nil {
		mm.Host = h
	} else {
		return nil, errors.New("host is nil")
	}
	return &mm, nil
}

// provisionDockerMachine provides fast provisioning of a docker machine
func provisionDockerMachine(h *host.Host) error {
	klog.Infof("provisionDockerMachine start ...")
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s to provisionDockerMachine", time.Since(start))
	}()

	p, err := fastDetectProvisioner(h)
	if err != nil {
		return errors.Wrap(err, "fast detect")
	}

	// avoid costly need to stop/power off/delete and then re-create docker machine due to the un-ready ssh server and hence errors like:
	// 'error starting host: creating host: create: provisioning: ssh command error: command : sudo hostname minikube-m02 && echo "minikube-m02" | sudo tee /etc/hostname; err: exit status 255'
	// so retry only on "exit status 255" ssh error and fall through in all other cases
	trySSH := func() error {
		if _, err := h.RunSSHCommand("hostname"); err != nil && strings.Contains(err.Error(), "exit status 255") {
			klog.Warning("ssh server returned retryable error (will retry)")
			return err
		}
		return nil
	}
	if err := retry.Expo(trySSH, 100*time.Millisecond, 5*time.Second); err != nil {
		klog.Errorf("ssh server returned non-retryable error (will continue): %v", err)
	}

	return p.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
}

// fastDetectProvisioner provides a shortcut for provisioner detection
func fastDetectProvisioner(h *host.Host) (libprovision.Provisioner, error) {
	d := h.Driver.DriverName()
	switch {
	case driver.IsKIC(d):
		return provision.NewUbuntuProvisioner(h.Driver), nil
	case driver.BareMetal(d), driver.IsSSH(d):
		return libprovision.DetectProvisioner(h.Driver)
	default:
		return provision.NewBuildrootProvisioner(h.Driver), nil
	}
}

// saveHost is a wrapper around libmachine's Save function to proactively update the node's IP whenever a host is saved
func saveHost(api libmachine.API, h *host.Host, cfg *config.ClusterConfig, n *config.Node) error {
	if err := api.Save(h); err != nil {
		return errors.Wrap(err, "save")
	}

	// Save IP to config file for subsequent use
	ip, err := h.Driver.GetIP()
	if err != nil {
		return err
	}
	if ip == "127.0.0.1" && driver.IsQEMU(h.Driver.DriverName()) {
		ip = "10.0.2.15"
	}
	n.IP = ip
	return config.SaveNode(cfg, n)
}

// backup copies critical ephemeral vm config files from tmpfs to persistent storage under /var/lib/minikube/backup,
// preserving same perms as original files/folders, from where they can be restored on next start,
// and returns any error occurred.
func backup(h host.Host, files []string) error {
	klog.Infof("backing up vm config to %s: %v", vmpath.GuestBackupDir, files)

	r, err := CommandRunner(&h)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	// ensure target dir exists
	if _, err := r.RunCmd(exec.Command("sudo", "mkdir", "-p", vmpath.GuestBackupDir)); err != nil {
		return errors.Wrapf(err, "create dir")
	}

	errs := []error{}
	for _, src := range []string{"/etc/cni", "/etc/kubernetes"} {
		if _, err := r.RunCmd(exec.Command("sudo", "rsync", "--archive", "--relative", src, vmpath.GuestBackupDir)); err != nil {
			errs = append(errs, errors.Errorf("failed to copy %q to %q (will continue): %v", src, vmpath.GuestBackupDir, err))
		}
	}
	if len(errs) > 0 {
		return errors.Errorf("%v", errs)
	}
	return nil
}

// restore copies back everything from backup folder using relative paths as their absolute restore locations,
// eg, "/var/lib/minikube/backup/etc/kubernetes" will be restored to "/etc/kubernetes",
// preserving same perms as original files/folders,
// files that were updated since last backup should not be overwritten,
func restore(h host.Host) error {
	r, err := CommandRunner(&h)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	// check first if we have anything to restore
	out, err := r.RunCmd(exec.Command("sudo", "ls", "--almost-all", "-1", vmpath.GuestBackupDir))
	if err != nil {
		return errors.Wrapf(err, "read dir")
	}
	files := strings.Split(strings.TrimSpace(out.Stdout.String()), "\n")

	klog.Infof("restoring vm config from %s: %v", vmpath.GuestBackupDir, files)

	errs := []error{}
	for _, dst := range files {
		if len(dst) == 0 {
			continue
		}
		src := path.Join(vmpath.GuestBackupDir, dst)
		if _, err := r.RunCmd(exec.Command("sudo", "rsync", "--archive", "--update", src, "/")); err != nil {
			errs = append(errs, errors.Errorf("failed to copy %q to %q (will continue): %v", src, dst, err))
		}
	}
	if len(errs) > 0 {
		return errors.Errorf("%v", errs)
	}
	return nil
}
