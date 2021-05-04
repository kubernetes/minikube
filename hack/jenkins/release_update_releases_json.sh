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

# This script generates the Github Release page and uploads all the binaries/etc to that page
# This is intended to be run on a new release tag in order to generate the github release page for that release

# The script expects the following env variables:
# VERSION_MAJOR: The major version of the tag to be released.
# VERSION_MINOR: The minor version of the tag to be released.
# VERSION_BUILD: The build version of the tag to be released.

set -e

export TAGNAME=v${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}

export DARWIN_SHA256=$(cat out/minikube-darwin-amd64.sha256)
export LINUX_SHA256=$(cat out/minikube-linux-amd64.sha256)
export WINDOWS_SHA256=$(cat out/minikube-windows-amd64.exe.sha256)

# Update releases.json w/ new release in gcs and github
git config user.name "minikube-bot"
git config user.email "minikube-bot@google.com"

git checkout -b "jenkins-releases.json-${TAGNAME}"

git status

if ! [[ "${VERSION_BUILD}" =~ ^[0-9]+$ ]]; then
  #Prepends the new version to the release-beta.json file
  sed -i "0,/{/s/{/{\n      \"name\": \"${TAGNAME}\",\n      \"checksums\": {\n          \"darwin\": \"${DARWIN_SHA256}\",\n          \"linux\": \"${LINUX_SHA256}\",\n          \"windows\": \"${WINDOWS_SHA256}\"\n      }\n  },\n  {"/ deploy/minikube/releases-beta.json

  git add -A
  git commit -m "Update releases-beta.json to include ${TAGNAME}"
  git remote add minikube-bot git@github.com:minikube-bot/minikube.git
  git push -f minikube-bot jenkins-releases.json-${TAGNAME}

  # Send PR from minikube-bot/minikube to kubernetes/minikube
  curl -X POST -u minikube-bot:${BOT_PASSWORD} -k   -d "{\"title\": \"update releases-beta.json to include ${TAGNAME}\",\"head\": \"minikube-bot:jenkins-releases.json-${TAGNAME}\",\"base\": \"master\"}" https://api.github.com/repos/kubernetes/minikube/pulls

  # Upload file to GCS so that minikube can see the new version
  gsutil cp deploy/minikube/releases-beta.json gs://minikube/releases-beta.json
else
  #Prepends the new version to the release.json file
  sed -i "0,/{/s/{/{\n      \"name\": \"${TAGNAME}\",\n      \"checksums\": {\n          \"darwin\": \"${DARWIN_SHA256}\",\n          \"linux\": \"${LINUX_SHA256}\",\n          \"windows\": \"${WINDOWS_SHA256}\"\n      }\n  },\n  {"/ deploy/minikube/releases.json

  #Update the front page of our documentation
  now=$(date +"%b %d, %Y")
  sed -i "s/Latest Release: .* (/Latest Release: ${TAGNAME} - ${now} (/" site/content/en/docs/_index.md

  git add -A
  git commit -m "Update releases.json to include ${TAGNAME}"
  git remote add minikube-bot git@github.com:minikube-bot/minikube.git
  git push -f minikube-bot jenkins-releases.json-${TAGNAME}

  # Send PR from minikube-bot/minikube to kubernetes/minikube
  curl -X POST -u minikube-bot:${BOT_PASSWORD} -k   -d "{\"title\": \"update releases.json to include ${TAGNAME}\",\"head\": \"minikube-bot:jenkins-releases.json-${TAGNAME}\",\"base\": \"master\"}" https://api.github.com/repos/kubernetes/minikube/pulls

  # Upload file to GCS so that minikube can see the new version
  gsutil cp deploy/minikube/releases.json gs://minikube/releases.json
fi
