#!/bin/bash

# Copyright 2016 The Kubernetes Authors All rights reserved.
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

# This script can take the following env variables
# ARGS: args to pass into the make rule
# 	ISO_BUCKET = the bucket location to upload the ISO (e.g. minikube-builds/PR_NUMBER)
# 	ISO_VERSION = the suffix for the iso (i.e. minikube-$(ISO_VERSION).iso)

set -x -o pipefail

# Clean build artifacts first - if available space is very low, anything else may fail.
echo "Cleaning out directory"
rm -rf out

# Clean stale Go caches that accumulate over time and can consume tens of GB.
# The ISO build uses buildroot and does not need the host Go cache.
# Clean it to prevent unbounded growth - only make generate-docs uses host Go.
echo "Cleaning Go caches"
go clean -cache -modcache || true
chmod -R u+w "${GOPATH}/src" || true
rm -rf "${GOPATH}/src" || true

# Clean Jenkins deferred wipeout leftovers and stale workspaces that may still occupy disk.
# Matches both @* (Jenkins deferred wipeout) and _ws-cleanup_* (WS-CLEANUP plugin) leftovers.
echo "Stale workspace copies:"
find "$(dirname "${WORKSPACE}")" -maxdepth 1 -name "$(basename "${WORKSPACE}")_*" -o -name "$(basename "${WORKSPACE}")@*" | sed 's/^/  /'
echo ""
chmod -R u+w "${WORKSPACE}"@* "${WORKSPACE}"_* || true
rm -rf "${WORKSPACE}"@* "${WORKSPACE}"_* || true

# Trim systemd journal to 1GB - no reason to keep gigabytes of logs on a build machine.
sudo journalctl --vacuum-size=1G || true

# Make sure gh is installed and configured
./hack/jenkins/installers/check_install_gh.sh

# Make sure golang is installed and configured
./hack/jenkins/installers/check_install_golang.sh "/usr/local"
export PATH=/usr/local/go/bin:$PATH

# install cron jobs
source ./hack/jenkins/installers/check_install_linux_crons.sh

# Generate changelog for latest github PRs merged
./hack/jenkins/build_changelog.sh deploy/iso/minikube-iso/board/minikube/aarch64/rootfs-overlay/CHANGELOG
./hack/jenkins/build_changelog.sh deploy/iso/minikube-iso/board/minikube/x86_64/rootfs-overlay/CHANGELOG

# Make sure all required packages are installed
sudo apt-get update
sudo apt-get -y install build-essential unzip rsync bc python3 p7zip-full cmake

# Log Cmake version
CMAKE_VERSION=$(cmake --version | head -n1 | awk '{print $3}')
echo "Start of ISO build: CMake version: $CMAKE_VERSION"
if dpkg --compare-versions "$CMAKE_VERSION" lt "3.20"; then
	echo "WARNING: CMake version $CMAKE_VERSION is less than 3.20. this will cause a slower build due to rebuidling cmake ..."
fi

# Let's make sure we have the newest ISO reference
curl -L https://github.com/kubernetes/minikube/raw/master/Makefile --output Makefile-head
# ISO tags are of the form VERSION-TIMESTAMP-PR, so this grep finds that TIMESTAMP in the middle
# if it doesn't exist, it will just return VERSION, which is covered in the if statement below
HEAD_ISO_TIMESTAMP=$(grep -E "ISO_VERSION \?= " Makefile-head | cut -d \" -f 2 | cut -d "-" -f 2)
CURRENT_ISO_TS=$(grep -E "ISO_VERSION \?= " Makefile | cut -d \" -f 2 | cut -d "-" -f 2)
if [[ $HEAD_ISO_TIMESTAMP != v* ]]; then
	diff=$((CURRENT_ISO_TS-HEAD_ISO_TIMESTAMP))
	if [[ $CURRENT_ISO_TS == v* ]] || [ $diff -lt 0 ]; then
		gh pr comment ${ghprbPullId} --body "Hi ${ghprbPullAuthorLoginMention}, your ISO info is out of date. Please rebase."
		exit 1
	fi
fi
rm Makefile-head

if [[ -z $ISO_VERSION ]]; then
	release=false
	IV=$(grep -E "ISO_VERSION \?=" Makefile | cut -d " " -f 3 | cut -d "-" -f 1)
	now=$(date +%s)
	export ISO_VERSION=$IV-$now-$ghprbPullId
	export ISO_BUCKET=minikube-builds/iso/$ghprbPullId
else
	release=true
	export ISO_VERSION
	export ISO_BUCKET
fi

# Check available space after all installs - fail early if insufficient
echo "Disk space before ISO build:"
df -h .

required_gb=100
available_gb=$(df --output=avail . | tail -1 | awk '{printf "%.0f", $1/1024/1024}')
echo "Available disk space: ${available_gb}GB"
if [ "$available_gb" -lt "$required_gb" ]; then
    echo "ERROR: Not enough disk space for ISO build. Available: ${available_gb}GB, required: at least ${required_gb}GB"
    ./hack/jenkins/investigate_disk_usage.sh || true
    body=$(cat << EOF
Hi ${ghprbPullAuthorLoginMention}, building a new ISO failed for Commit ${ghprbActualCommit}
Not enough disk space on build machine (${available_gb}GB available, need ${required_gb}GB).

See the logs at:
https://storage.cloud.google.com/minikube-builds/logs/${ghprbPullId}/${ghprbActualCommit::7}/iso_build.txt

Please contact the maintainers to clean up the ISO build node.
EOF
)
    gh pr comment "${ghprbPullId}" --body "$body"
    exit 1
fi

if ! make release-iso 2>&1 | tee iso-logs.txt; then
    # Exit of `make` (PIPESTATUS[0]); fallback to 1 if unavailable
    ec=${PIPESTATUS[0]:-1}

    # Only comment on non-release; default release=false if unset
    if [[ ${release:-false} != "true" ]]; then
        body=$(cat << EOF
Hi ${ghprbPullAuthorLoginMention}, building a new ISO failed for Commit ${ghprbActualCommit}
See the logs at:
https://storage.cloud.google.com/minikube-builds/logs/${ghprbPullId}/${ghprbActualCommit::7}/iso_build.txt
EOF
)
	    gh pr comment "${ghprbPullId}" --body "$body"
    fi
    exit "$ec"
fi

git config user.name "minikube-bot"
git config user.email "20374350+minikube-bot@users.noreply.github.com"

if [ "$release" = false ]; then
	# Update the user's PR with the newly built ISO.
	# We use `gh pr checkout` because it sets up the fork remote over HTTPS
	# using the gh auth token, which allows pushing to fork PRs when "Allow
	# edits from maintainers" is enabled. SSH-based push cannot work since
	# the bot does not have SSH access to contributors' forks.

	gh auth setup-git
	gh pr checkout ${ghprbPullId}

	sed -i "s/ISO_VERSION ?= .*/ISO_VERSION ?= ${ISO_VERSION}/" Makefile
	sed -i "s|isoBucket := .*|isoBucket := \"${ISO_BUCKET}\"|" pkg/minikube/download/iso.go
	make generate-docs

	git add Makefile pkg/minikube/download/iso.go site/content/en/docs/commands/start.md
	git commit -m "Updating ISO to ${ISO_VERSION}"
	if git push; then
		message=$(cat <<EOF
Hi ${ghprbPullAuthorLoginMention}, we have updated your PR with the reference to newly built ISO.
Pull the changes locally if you want to test with them or update your PR further.

Build logs (for reference):
https://storage.cloud.google.com/minikube-builds/logs/${ghprbPullId}/${ghprbActualCommit::7}/iso_build.txt
EOF
)
	else
		message=$(cat <<EOF
Hi ${ghprbPullAuthorLoginMention}, we built a new ISO but failed to push the update to your PR.
Please run the following commands and push manually:

\`\`\`
sed -i 's/ISO_VERSION ?= .*/ISO_VERSION ?= ${ISO_VERSION}/' Makefile
sed -i 's|isoBucket := .*|isoBucket := "${ISO_BUCKET}"|' pkg/minikube/download/iso.go
make generate-docs
\`\`\`

See the build logs for more details:
https://storage.cloud.google.com/minikube-builds/logs/${ghprbPullId}/${ghprbActualCommit::7}/iso_build.txt
EOF
)
	fi
	gh pr comment ${ghprbPullId} --body "${message}"
else
	# Release!
	branch=iso-release-${ISO_VERSION}
	git checkout -b ${branch}

	sed -i "s/ISO_VERSION ?= .*/ISO_VERSION ?= ${ISO_VERSION}/" Makefile
	sed -i "s|isoBucket := .*|isoBucket := \"${ISO_BUCKET}\"|" pkg/minikube/download/iso.go
	make generate-docs

	git add Makefile pkg/minikube/download/iso.go site/content/en/docs/commands/start.md
	git commit -m "Release: Update ISO to ${ISO_VERSION}"
	git remote add minikube-bot git@github.com:minikube-bot/minikube.git
	git push -f minikube-bot ${branch}

	gh pr create --fill --base master --head minikube-bot:${branch} -l "ok-to-test"
fi
