#!/bin/bash

# Copyright 2022 The Kubernetes Authors All rights reserved.
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

set -e

mkdir -p cron && gsutil -qm rsync "gs://minikube-builds/${MINIKUBE_LOCATION}/cron" cron || echo "FAILED TO GET CRON FILES"
install cron/cleanup_and_reboot_Darwin.sh $HOME/cleanup_and_reboot.sh || echo "FAILED TO INSTALL CLEANUP AND REBOOT"
echo "*/30 * * * * $HOME/cleanup_and_reboot.sh" | crontab
install cron/cleanup_go_modules.sh $HOME/cleanup_go_modules.sh || echo "FAILED TO INSTALL GO MODULES CLEANUP"
(crontab -l ; echo "0 0 1 * * $HOME/cleanup_go_modules.sh")| crontab -
crontab -l
