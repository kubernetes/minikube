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

if [ "$#" -ne 1 ]; then
	# there's no secret and therefore no reason to run this script
	exit 0
fi

install_gh() {
	export access_token="$1"

	# Make sure gh is installed and configured
	./hack/jenkins/installers/check_install_gh.sh
}

config_git() {
	git config user.name "minikube-bot"
	git config user.email "minikube-bot@google.com"
}

make generate-docs

# If there are changes, open a PR
changes=$(git status --porcelain)
if [ "$changes" != "" ]; then
	install_gh $1
	config_git
	
	branch=gendocs$(date +%s%N)
	git checkout -b $branch
	
	git add .
	git commit -m "Update generate-docs"

	git remote add minikube-bot https://minikube-bot:$access_token@github.com/minikube-bot/minikube.git
	git push -u minikube-bot $branch
	gh pr create --base master --head minikube-bot:$branch --title "Update generate-docs" --body "Committing changes resulting from \`make generate-docs\`"
fi
