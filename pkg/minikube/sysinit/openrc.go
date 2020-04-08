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

const SysVName = "OpenRC"


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

package ktmpl

import "text/template"

var RestartWrapper = `#!/bin/bash
# Wrapper script to emulate systemd restart on non-systemd systems
binary=$1
conf=$2
args=""

while [[ -x "${binary}" ]]; do
  if [[ -f "${conf}" ]]; then
          args=$(egrep "^ExecStart=${binary}" "${conf}" | cut -d" " -f2-)
  fi
  echo "$(date) binary=${binary} args=${args}"
  ${binary} ${args}
  echo ""
  sleep 1
done
`

var InitScript = template.Must(template.New("initScript").Parse(`#!/bin/bash
# OpenRC init script for systemd units

readonly BINARY="{{.BinaryPath}}"
readonly NAME="$(basename ${BINARY})"

readonly RESTART_WRAPPER="{{.WrapperPath}}"
readonly UNIT_PATH="{{.UnitPath}}"
readonly PID_PATH="/var/run/${NAME}.pid"

if [[ ! -x "${BINARY}" ]]; then
	echo "$BINARY not present or not executable"
	exit 1
fi

function start() {
    start-stop-daemon --oknodo --PID_PATH "${PID_PATH}" --background --start --make-PID_PATH --exec "${WRAPPER}" "${BINARY}" "${UNIT_PATH}"
}

function stop() {
    start-stop-daemon --oknodo --PID_PATH "${PID_PATH}" --stop
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
        start-stop-daemon --pid "${PID_PATH}" --status
		;;
	*)
	    echo "Usage: service BINARY {start|stop|restart|status}"
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
	return SysVName
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
	glog.Infof("rr: %v", rr.Output())
	return err
}

// Restart restarts a service
func (s *OpenRC) Restart(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "restart"))
	return err
}

// Stop stops a service
func (s *OpenRC) Stop(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "stop"))
	return err
}

// ForceStop stops a service with prejuidice
func (s *OpenRC) ForceStop(svc string) error {
	_, err := s.r.RunCmd(exec.Command("sudo", "service", svc, "stop", "-f"))
	return err
}
