#!/usr/bin/bash
shopt -s nullglob

SYSCTL=/usr/bin/systemctl

if [ $# -eq 1 ]; then
    app=$1
    status=$(${SYSCTL} show --property ExecMainStatus "${app}.service")
    echo "${status#*=}" > "/rkt/status/$app"
    exit 0
fi
