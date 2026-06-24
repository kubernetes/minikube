#!/bin/bash

# Copyright 2026 The Kubernetes Authors All rights reserved.
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

# Investigate disk usage on build nodes.
# Shows top 5 entries by size, recursively drilling into directories.
# Usage: ./hack/jenkins/investigate_disk_usage.sh [root_path] [max_depth]

set -o pipefail
shopt -s dotglob

root="${1:-/}"
max_depth="${2:-3}"

investigate() {
    local dir="${1%/}"
    local depth="$2"

    local top5
    top5=$(sudo du -sm "${dir}"/* 2>/dev/null | sort -rn | head -5 || true)

    echo "--- ${dir}/* ---"
    echo "$top5"
    echo ""

    [ "$depth" -ge "$max_depth" ] && return

    for entry in $(echo "$top5" | awk '{print $2}'); do
        [ -d "${entry}" ] || continue
        investigate "${entry}" $((depth + 1))
    done
}

echo "=== Disk usage investigation: ${root} ==="
echo ""
df -h "${root}"
echo ""

investigate "${root}" 1

echo "=== End disk usage investigation ==="
