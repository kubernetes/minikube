#!/bin/bash
#
if [[ ! -f cleanup.sh ]]; then
  echo "cleanup.sh is missing"
  exit 1
fi

sudo install cleanup.sh /etc/cron.hourly/cleanup || echo "FAILED TO INSTALL CLEANUP"
(crontab -l 2>/dev/null; echo '@reboot rm -rf /var/run/reboot.in.progress') | crontab -
