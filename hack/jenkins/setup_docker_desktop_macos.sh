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

set -x

if docker system info > /dev/null 2>&1; then
  echo "Docker is already running, exiting"
  exit 0
fi

# kill docker first
osascript -e 'quit app "Docker"'

# wait 2 minutes for it to start back up
timeout=120
elapsed=0
echo "Starting Docker Desktop..."
open --background -a Docker
echo "Waiting at most two minutes..."
while ! docker system info > /dev/null 2>&1;
do
  sleep 1
  elapsed=$((elapsed+1))
  if [ $elapsed -gt $timeout ]; then
	  echo "Start Docker Desktop failed"
	  exit 1
  fi
done

echo "Docker Desktop started!"
