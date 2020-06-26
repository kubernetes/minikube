/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"html/template"
	"os/exec"
	"path"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

var sysvRestartWrapper = `#!/bin/bash
# Wrapper script to emulate systemd restart on non-systemd systems
readonly UNIT_PATH=$1

while true; do
  if [[ -f "${UNIT_PATH}" ]]; then
	eval $(egrep "^ExecStart=" "${UNIT_PATH}" | cut -d"=" -f2-)
  fi
  sleep 1
done
`

var sysvInitScriptTmpl = template.Must(template.New("initScript").Parse(`#!/bin/bash
# SysV init script shim for systemd units
readonly NAME="{{.Name}}"
readonly BINARY="{{.Binary}}"
readonly RESTART_WRAPPER="{{.Wrapper}}"
readonly UNIT_PATH="{{.Unit}}"
readonly PID_PATH="/var/run/${NAME}.pid"

if [ -r /etc/rc.d/init.d/functions ]; then
  lsb=false
  . /etc/rc.d/init.d/functions
else
  lsb=true
  . /lib/lsb/init-functions
fi

function start() {
    if $lsb; then
        start_daemon -p "${PID_PATH}" "${RESTART_WRAPPER}" "${UNIT_PATH}"
    else
        daemon --pidfile "${PID_PATH}" "${RESTART_WRAPPER}" "${UNIT_PATH}"
    fi
}

function stop() {
    if $lsb; then
        killproc -p "${PID_PATH}" $BINARY $NAME
    else
        killproc -p "${PID_PATH}" -b $BINARY $NAME
    fi
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
        if $lsb; then
            status_of_proc -p "${PID_PATH}" $BINARY $NAME
        else
            status -p "${PID_PATH}" -b $BINARY $NAME
        fi
		;;
	*)
	    echo "Usage: {{.Name}} {start|stop|restart|status}"
		exit 1
		;;
esac
`))

// SysV is a service manager for SysV-like init systems
type SysV struct {
	r Runner
}

// Name returns the name of the init system
func (s *SysV) Name() string {
	return "SysV"
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
	glog.Infof("start output: %s", rr.Output())
	return err
}

// Disable does nothing
func (s *SysV) Disable(svc string) error {
	return nil
}

// Enable does nothing
func (s *SysV) Enable(svc string) error {
	return nil
}

// Restart restarts a service
func (s *SysV) Restart(svc string) error {
	rr, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "restart"))
	glog.Infof("restart output: %s", rr.Output())
	return err
}

// Stop stops a service
func (s *SysV) Stop(svc string) error {
	rr, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "stop"))
	glog.Infof("stop output: %s", rr.Output())
	return err
}

// ForceStop stops a service with prejuidice
func (s *SysV) ForceStop(svc string) error {
	return s.Stop(svc)
}

// GenerateInitShim generates any additional init files required for this service
func (s *SysV) GenerateInitShim(svc string, binary string, unit string) ([]assets.CopyableFile, error) {
	restartWrapperPath := path.Join(vmpath.GuestPersistentDir, "sysv-restart-wrapper.sh")

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
	if err := sysvInitScriptTmpl.Execute(&b, opts); err != nil {
		return nil, errors.Wrap(err, "template execute")
	}

	files := []assets.CopyableFile{
		assets.NewMemoryAssetTarget([]byte(sysvRestartWrapper), restartWrapperPath, "0755"),
		assets.NewMemoryAssetTarget(b.Bytes(), path.Join("/etc/init.d/", svc), "0755"),
	}

	return files, nil
}
