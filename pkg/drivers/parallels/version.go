/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package parallels

import "fmt"

// GitCommit that was compiled. This will be filled in by the compiler.
var GitCommit string

// Version number that is being run at the moment.
const Version = "2.0.1"

// FullVersion formats the version to be printed.
func FullVersion() string {
	return fmt.Sprintf("%s, build %s", Version, GitCommit)
}
