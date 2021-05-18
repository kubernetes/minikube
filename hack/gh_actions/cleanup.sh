#!/bin/bash

PROGRESS_MARK=/var/run/job.in.progress
REBOOT_MARK=/var/run/reboot.in.progress

timeout=900 # 15 minutes

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
check_running_job

echo "cleanup docker..."
docker kill $(docker ps -aq) >/dev/null 2>&1 || true
docker system prune --volumes --force || true
sudo reboot
