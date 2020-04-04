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

// sysinit provides an abstraction over init systems like systemctl
package sysinit

import (
	"context"
	"os/exec"
	"time"

	"github.com/golang/glog"
)

const SysVName = "Sys-V"

// SysV is a service manager for SysV-like init systems
type SysV struct {
	r Runner
}

// Name returns the name of the init system
func (s *SysV) Name() string {
	return SysVName
}

// Active checks if a service is running
func (s *SysV) Active(svc string) bool {
	_, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "status"))
	return err == nil
}

// Start starts a service idempotently
func (s *SysV) Start(svc string) error {
	if s.Active(svc) {
		return nil
	}
	ctx, cb := context.WithTimeout(context.Background(), 5*time.Second)
	defer cb()

	rr, err := s.r.RunCmd(exec.CommandContext(ctx, "sudo", "service", svc, "start"))
	glog.Infof("rr: %v", rr.Output())
	return err
}

// Restart restarts a service
func (s *SysV) Restart(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "restart"))
	return err
}

// Stop stops a service
func (s *SysV) Stop(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "stop"))
	return err
}

// ForceStop stops a service with prejuidice
func (s *SysV) ForceStop(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "stop", "-f"))
	return err
}
