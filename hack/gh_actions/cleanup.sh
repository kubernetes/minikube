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

PROGRESS_MARK=/var/run/job.in.progress
REBOOT_MARK=/var/run/reboot.in.progress

timeout=900 # 15 minutes

# $PROGRESS_MARK file is touched when a new GitHub Actions job is started
# check that no job was started in the last 15 minutes
function check_running_job() {
  if [[ -f "$PROGRESS_MARK" ]]; then
    started=$(date -r "$PROGRESS_MARK" +%s)
    elapsed=$(($(date +%s) - started))
    if (( elapsed > timeout )); then
      echo "Job started ${elapsed} seconds ago, going to restart"
      sudo rm -rf "$PROGRESS_MARK"
    else
      echo "Job is running. exit."
      exit 1
    fi
  fi
}

check_running_job
sudo touch "$REBOOT_MARK"
# avoid race if a job was started between two lines above and recheck
check_running_job

echo "cleanup docker..."
docker kill $(docker ps -aq) >/dev/null 2>&1 || true
docker system prune --volumes --force || true

echo "rebooting..."
sudo reboot
