/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package notify

const (
	// GithubMinikubeReleasesURL is the URL of the minikube github releases JSON file
	GithubMinikubeReleasesURL = "https://storage.googleapis.com/minikube/releases-v2.json"
	// GithubMinikubeBetaReleasesURL is the URL of the minikube GitHub beta releases JSON file
	GithubMinikubeBetaReleasesURL = "https://storage.googleapis.com/minikube/releases-beta-v2.json"

	// GithubMinikubeReleasesAliyunURL is the URL of the minikube github releases JSON file from Aliyun Mirror
	GithubMinikubeReleasesAliyunURL = "https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/releases.json"
	// GithubMinikubeBetaReleasesAliyunURL is the URL of the minikube GitHub beta releases JSON file
	GithubMinikubeBetaReleasesAliyunURL = "https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/releases-beta.json"
)
