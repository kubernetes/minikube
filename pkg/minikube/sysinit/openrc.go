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
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os/exec"
	"path"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

var restartWrapper = `#!/bin/bash
# Wrapper script to emulate systemd restart on non-systemd systems
readonly UNIT_PATH=$1

while true; do
  if [[ -f "${UNIT_PATH}" ]]; then
	eval $(egrep "^ExecStart=" "${UNIT_PATH}" | cut -d"=" -f2-)
  fi
  sleep 1
done
`

var initScriptTmpl = template.Must(template.New("initScript").Parse(`#!/bin/bash
# OpenRC init script shim for systemd units
readonly NAME="{{.Name}}"
readonly RESTART_WRAPPER="{{.Wrapper}}"
readonly UNIT_PATH="{{.Unit}}"
readonly PID_PATH="/var/run/${NAME}.pid"

function start() {
    start-stop-daemon --oknodo --pidfile "${PID_PATH}" --background --start --make-pid --exec "${RESTART_WRAPPER}" "${UNIT_PATH}"
}

function stop() {
	if [[ -f "${PID_PATH}" ]]; then
		pkill -P "$(cat ${PID_PATH})"
	fi
	start-stop-daemon --oknodo --pidfile "${PID_PATH}" --stop
}

case "$1" in
    start)
        start
		;;
    stop)
        stop
		;;
    restart)
        stop
        start
		;;
    status)
        start-stop-daemon --pidfile "${PID_PATH}" --status
		;;
	*)
	    echo "Usage: {{.Name}} {start|stop|restart|status}"
		exit 1
		;;
esac
`))

// OpenRC is a service manager for OpenRC-like init systems
type OpenRC struct {
	r Runner
}

// Name returns the name of the init system
func (s *OpenRC) Name() string {
	return "OpenRC"
}

// Active checks if a service is running
func (s *OpenRC) Active(svc string) bool {
	_, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "status"))
	return err == nil
}

// Start starts a service idempotently
func (s *OpenRC) Start(svc string) error {
	if s.Active(svc) {
		return nil
	}
	ctx, cb := context.WithTimeout(context.Background(), 5*time.Second)
	defer cb()

	rr, err := s.r.RunCmd(exec.CommandContext(ctx, "sudo", "service", svc, "start"))
	klog.Infof("start output: %s", rr.Output())
	return err
}

// Disable does nothing
func (s *OpenRC) Disable(svc string) error {
	return nil
}

// DisableNow not implemented for openRC
func (s *OpenRC) DisableNow(svc string) error {
	return fmt.Errorf("disable now is not implemented for OpenRC! PRs to fix are welcomed")
}

// Enable does nothing
func (s *OpenRC) Enable(svc string) error {
	return nil
}

// EnableNow  not implemented for openRC
func (s *OpenRC) EnableNow(svc string) error {
	return fmt.Errorf("enable now is not implemented for OpenRC! PRs to fix are welcomed")
}

// Restart restarts a service
func (s *OpenRC) Restart(svc string) error {
	rr, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "restart"))
	klog.Infof("restart output: %s", rr.Output())
	return err
}

// Reload reloads a service
// currently only used by our docker-env that doesn't need openrc implementation
func (s *OpenRC) Reload(svc string) error {
	return fmt.Errorf("reload is not implemented for OpenRC yet ! Please implement if needed")
}

// Stop stops a service
func (s *OpenRC) Stop(svc string) error {
	rr, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "stop"))
	klog.Infof("stop output: %s", rr.Output())
	return err
}

// ForceStop stops a service with prejuidice
func (s *OpenRC) ForceStop(svc string) error {
	return s.Stop(svc)
}

// GenerateInitShim generates any additional init files required for this service
func (s *OpenRC) GenerateInitShim(svc string, binary string, unit string) ([]assets.CopyableFile, error) {
	restartWrapperPath := path.Join(vmpath.GuestPersistentDir, "openrc-restart-wrapper.sh")

	opts := struct {
		Binary  string
		Wrapper string
		Name    string
		Unit    string
	}{
		Name:    svc,
		Binary:  binary,
		Wrapper: restartWrapperPath,
		Unit:    unit,
	}

	var b bytes.Buffer
	if err := initScriptTmpl.Execute(&b, opts); err != nil {
		return nil, errors.Wrap(err, "template execute")
	}

	files := []assets.CopyableFile{
		assets.NewMemoryAssetTarget([]byte(restartWrapper), restartWrapperPath, "0755"),
		assets.NewMemoryAssetTarget(b.Bytes(), path.Join("/etc/init.d/", svc), "0755"),
	}

	return files, nil
}
