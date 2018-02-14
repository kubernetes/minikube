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

# This script generates the CHANGELOG.md
# This is intended to be run on a new release tag in order to generate the CHANGELOG.md for that release

# The script expects the following env variabls:
# VERSION_MAJOR: The the major version of the tag to be released.
# VERSION_MINOR: The the minor version of the tag to be released.
# VERSION_BUILD: The the build version of the tag to be released.
# OLDTAGNAME: The name of the last taged version that was built and released.

set -e

export TAGNAME=v${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}

# Update CHANGELOG.md w/ changes in github
git config user.name "minikube-bot"
git config user.email "minikube-bot@google.com"

git checkout -b "jenkins-CHANGELOG.md-${TAGNAME}"

git status

#Prepends the new version to the release.json file
# TODO(aprindle) change the Add/Added/Update words to predefined words or tag that specifies use in CHANGELOG.md.
# TODO(aprindle) document the above in README.md
export GIT_DIFF=$(git log --pretty=oneline --no-merges ${OLDTAGNAME}..${TAGNAME} | cut -d ' ' -f2- | grep '^Add\|^Added\|^Update\|^Updated\|^Fix\|^Fixed\|^Enable\|^Enabled')
GIT_DIFF=$(echo "$GIT_DIFF" | sed "s/^/* /")
GIT_DIFF=$(echo "$GIT_DIFF" | tr '\n' '~')

sed -i "3i ##VERSION ${TAGNAME} - $(date +"%d\\\\%m\\\\%y")" CHANGELOG.md
sed -i "4i ${GIT_DIFF}" CHANGELOG.md
sed -i 's/~/\'$'\n/g' CHANGELOG.md

git add -A
git commit -m "Update CHANGELOG.md for ${TAGNAME}"
git remote add minikube-bot git@github.com:minikube-bot/minikube.git
git push -f minikube-bot jenkins-CHANGELOG.md-${TAGNAME}

# Send PR from minikube-bot/minikube to kubernetes/minikube
curl -X POST -u minikube-bot:${BOT_PASSWORD} -k   -d "{\"title\": \"Update CHANGELOG.md for ${TAGNAME}\",\"head\": \"minikube-bot:jenkins-CHANGELOG.md-${TAGNAME}\",\"base\": \"master\"}" https://api.github.com/repos/kubernetes/minikube/pulls

# Upload file to GCS so that minikube can see the new version
gsutil cp deploy/minikube/releases.json gs://minikube/releases.json
