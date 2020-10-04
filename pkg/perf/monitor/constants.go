/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package monitor

const (
	// GithubAccessTokenEnvVar is the env var name to use
	GithubAccessTokenEnvVar = "GITHUB_ACCESS_TOKEN"

	// OkToTestLabel is the github label for ok-to-test
	OkToTestLabel           = "ok-to-test"

	// GithubOwner is the owner of the github repository
	GithubOwner             = "kubernetes"

	// GithubRepo is the name of the github repository
	GithubRepo              = "minikube"

	// BotName is the name of the minikube Pull Request bot
	BotName                 = "minikube-pr-bot"
)
