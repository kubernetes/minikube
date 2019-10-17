#!/bin/bash
# Periodically cleanup and reboot if no Jenkins subprocesses are running.
#
# Installation:
#   install cleanup-and-reboot.linux /etc/cron.hourly/cleanup-and-reboot

function check_jenkins() {
  jenkins_pid="$(pidof java)"
  if [[ "${jenkins_pid}" = "" ]]; then
          return
  fi
  pstree "${jenkins_pid}" \
        | grep -v java \
        && echo "jenkins is running at pid ${jenkins_pid} ..." \
        && exit 1
}

check_jenkins
logger "cleanup-and-reboot running - may shutdown in 60 seconds"
wall "cleanup-and-reboot running - may shutdown in 60 seconds"

sleep 60

check_jenkins
logger "cleanup-and-reboot is happening!"

# kill jenkins to avoid an incoming request
killall java

# disable localkube, kubelet
systemctl list-unit-files --state=enabled \
        | grep kube \
        | awk '{ print $1 }' \
        | xargs systemctl disable

# update and reboot
apt update -y \
        && apt upgrade -y \
        && reboot
