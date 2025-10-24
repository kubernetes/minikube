/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package run

// CommandOptions are minikube command line options.
type CommandOptions struct {
	// NonInteractive is true if the minikube command run with the
	// --interactive=false flag and we can not interact with the user.
	NonInteractive bool

	// DownloadOnly is true if the minikube command run with the --download-only
	// flag and we should If only download and cache files for later use and
	// don't install or start anything.
	DownloadOnly bool
}
