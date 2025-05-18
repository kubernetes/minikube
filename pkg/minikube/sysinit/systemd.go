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

// Package sysinit provides an abstraction over init systems like systemctl
package sysinit

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/minikube/pkg/minikube/assets"
)

// Systemd is a service manager for systemd distributions
type Systemd struct {
	r Runner
}

// Name returns the name of the init system
func (s *Systemd) Name() string {
	return "systemd"
}

// daemonReload reloads systemd configuration
func (s *Systemd) daemonReload() error {
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "daemon-reload"))
	return err
}

// Active checks if a service is running
func (s *Systemd) Active(svc string) bool {
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "is-active", "--quiet", "service", svc))
	return err == nil
}

// Disable disables a service
func (s *Systemd) Disable(svc string) error {
	cmd := exec.Command("sudo", "systemctl", "disable", svc)
	// See https://github.com/kubernetes/minikube/issues/11615#issuecomment-861794258
	cmd.Env = append(cmd.Env, "SYSTEMCTL_SKIP_SYSV=1")
	_, err := s.r.RunCmd(cmd)
	return err
}

// DisableNow disables a service and stops it too (not waiting for next restart)
func (s *Systemd) DisableNow(svc string) error {
	cmd := exec.Command("sudo", "systemctl", "disable", "--now", svc)
	// See https://github.com/kubernetes/minikube/issues/11615#issuecomment-861794258
	cmd.Env = append(cmd.Env, "SYSTEMCTL_SKIP_SYSV=1")
	_, err := s.r.RunCmd(cmd)
	return err
}

// Mask prevents a service from being started
func (s *Systemd) Mask(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "mask", svc))
	return err
}

// Enable enables a service
func (s *Systemd) Enable(svc string) error {
	if svc == "kubelet" {
		return errors.New("please don't enable kubelet as it creates a race condition; if it starts on systemd boot it will pick up /etc/hosts before we have time to configure /etc/hosts")
	}
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "enable", svc))
	return s.appendJournalctlLogsOnFailure(svc, err)
}

// EnableNow enables a service and then activates it too (not waiting for next start)
func (s *Systemd) EnableNow(svc string) error {
	if svc == "kubelet" {
		return errors.New("please don't enable kubelet as it creates a race condition; if it starts on systemd boot it will pick up /etc/hosts before we have time to configure /etc/hosts")
	}
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "enable", "--now", svc))
	return s.appendJournalctlLogsOnFailure(svc, err)
}

// Unmask allows a service to be started
func (s *Systemd) Unmask(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "unmask", svc))
	return err
}

// Start starts a service
func (s *Systemd) Start(svc string) error {
	if err := s.daemonReload(); err != nil {
		return err
	}
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "start", svc))
	return s.appendJournalctlLogsOnFailure(svc, err)
}

// Restart restarts a service
func (s *Systemd) Restart(svc string) error {
	if err := s.daemonReload(); err != nil {
		return err
	}

	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "restart", svc))
	return s.appendJournalctlLogsOnFailure(svc, err)
}

// run systemctl reset-failed for a service
// some services declare a realitive small restart-limit in their .service configuration
// so we reset reset-failed counter to override the limit
func (s *Systemd) ResetFailed(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "reset-failed", svc))
	return s.appendJournalctlLogsOnFailure(svc, err)
}

// Reload reloads a service
func (s *Systemd) Reload(svc string) error {
	if err := s.daemonReload(); err != nil {
		return err
	}
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "reload", svc))
	return s.appendJournalctlLogsOnFailure(svc, err)
}

// Stop stops a service
func (s *Systemd) Stop(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "stop", svc))
	return s.appendJournalctlLogsOnFailure(svc, err)
}

// ForceStop terminates a service with prejudice
func (s *Systemd) ForceStop(svc string) error {
	rr, err := s.r.RunCmd(exec.Command("sudo", "systemctl", "stop", "-f", svc))
	if err == nil {
		return nil
	}
	if strings.Contains(rr.Output(), fmt.Sprintf("Unit %s not loaded", svc)) {
		// already stopped
		return nil
	}
	return err
}

// GenerateInitShim does nothing for systemd
func (s *Systemd) GenerateInitShim(_, _, _ string) ([]assets.CopyableFile, error) {
	return nil, nil
}

func usesSystemd(r Runner) bool {
	_, err := r.RunCmd(exec.Command("systemctl", "--version"))
	return err == nil
}

// appendJournalctlLogsOnFailure appends journalctl logs for the service to the error if err is not nil
func (s *Systemd) appendJournalctlLogsOnFailure(svc string, err error) error {
	if err == nil {
		return nil
	}
	rr, logErr := s.r.RunCmd(exec.Command("sudo", "journalctl", "--no-pager", "-u", svc))
	if logErr != nil {
		return fmt.Errorf("%v\nfailed to get journalctl logs: %v", err, logErr)
	}

	return fmt.Errorf("%v\n%s:\n%s", err, rr.Command(), rr.Output())
}
