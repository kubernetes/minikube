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

package register

// PrintStep prints a Step type in JSON format
func PrintStep(message string) {
	s := NewStep(message)
	printAsCloudEvent(s, s.data)
}

// PrintInfo prints an Info type in JSON format
func PrintInfo(message string) {
	s := NewInfo(message)
	printAsCloudEvent(s, s.data)
}

// PrintDownload prints a Download type in JSON format
func PrintDownload(artifact string) {
	s := NewDownload(artifact)
	printAsCloudEvent(s, s.data)
}

// PrintDownloadProgress prints a DownloadProgress type in JSON format
func PrintDownloadProgress(artifact, progress string) {
	s := NewDownloadProgress(artifact, progress)
	printAsCloudEvent(s, s.data)
}

// PrintError prints an Error type in JSON format
func PrintError(err string) {
	e := NewError(err)
	printAsCloudEvent(e, e.data)
}

// PrintErrorExitCode prints an error in JSON format and includes an exit code
func PrintErrorExitCode(err string, exitcode int, additionalArgs ...map[string]string) {
	e := NewErrorExitCode(err, exitcode, additionalArgs...)
	printAsCloudEvent(e, e.data)
}
