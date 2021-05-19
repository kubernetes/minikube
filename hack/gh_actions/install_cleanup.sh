#!/bin/bash

# Copyright 2021 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

if [[ ! -f cleanup.sh ]]; then
  echo "cleanup.sh is missing"
  exit 1
fi

# update cron to run the cleanup script every hour
if [ ! -f /etc/cron.hourly/cleanup.sh ]; then
  sudo install cleanup.sh /etc/cron.hourly/cleanup.sh || echo "FAILED TO INSTALL CLEANUP"
fi

# install a cron rule to remove /var/run/reboot.in.progress file immediately after reboot
if (crontab -l 2>/dev/null | grep '@reboot rm -rf /var/run/reboot.in.progress'); then
  echo "reboot cron rule already installed"
  exit 0
fi
(crontab -l 2>/dev/null; echo '@reboot rm -rf /var/run/reboot.in.progress') | crontab -
