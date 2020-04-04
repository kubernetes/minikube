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

var KubeletInitTemplate = template.Must(template.New("kubeletSysVTemplate").Parse(`#!/bin/bash
# SysV style init script for kubelet

readonly KUBELET="{{.KubeletPath}}"
readonly KUBELET_WRAPPER="{{.WrapperPath}}"
readonly KUBELET_CONF="{{.ConfPath}}"
readonly KUBELET_PIDFILE="/var/run/kubelet.pid"
readonly KUBELET_LOGFILE="/var/run/nohup.out"

if [[ ! -x "${KUBELET}" ]]; then
	echo "$KUBELET not present or not executable"
	exit 1
fi

function start() {
    start-stop-daemon --pid "${KUBELET_PIDFILE}" --background --start --make-pidfile --exec "${KUBELET_WRAPPER}" "${KUBELET}" "${KUBELET_CONF}"
}

function stop() {
    start-stop-daemon --pid "${KUBELET_PIDFILE}" --stop
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
        start-stop-daemon --pid "${KUBELET_PIDFILE}" --status
		;;
	*)
		echo "Usage: service kubelet {start|stop|restart|status}"
		exit 1
		;;
esac
`))
