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

package deployer

type MiniTestDeployer interface {
	// Up should provision the environment for testing
	Up() error
	// Down should tear down the environment if any
	Down() error
	// IsUp should return true if a test cluster is successfully provisioned
	IsUp() (bool, error)
	// Execute execute a command in the deployed environment
	Execute(args ...string) error
	// SyncToRemote sync file or folder from src on host to dst on deployed environment
	SyncToRemote(src string, dst string, excludedPattern []string) error
	// SyncToRemote sync file or folder from src on remote to host
	SyncToHost(src string, dst string, excludedPattern []string) error
}
