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

import "fmt"

// Log represents the different types of logs that can be output as JSON
// This includes: Step, Download, DownloadProgress, Warning, Info, Error
type Log interface {
	Type() string
}

// Step represents a normal step in minikube execution
type Step struct {
	data map[string]string
}

// Type returns the cloud events compatible type of this struct
func (s *Step) Type() string {
	return "io.k8s.sigs.minikube.step"
}

// NewStep returns a new step type
func NewStep(message string) *Step {
	return &Step{data: map[string]string{
		"totalsteps":  Reg.totalSteps(),
		"currentstep": Reg.currentStep(),
		"message":     message,
		"name":        string(Reg.current),
	}}
}

// Download will be used to notify the user that a download has begun
type Download struct {
	data map[string]string
}

// Type returns the cloud events compatible type of this struct
func (s *Download) Type() string {
	return "io.k8s.sigs.minikube.download"
}

// NewDownload returns a new download type
func NewDownload(artifact string) *Download {
	return &Download{data: map[string]string{
		"totalsteps":  Reg.totalSteps(),
		"currentstep": Reg.currentStep(),
		"artifact":    artifact,
	}}
}

// DownloadProgress will be used to notify the user around the progress of a download
type DownloadProgress struct {
	data map[string]string
}

// Type returns the cloud events compatible type of this struct
func (s *DownloadProgress) Type() string {
	return "io.k8s.sigs.minikube.download.progress"
}

// NewDownloadProgress returns a new download progress type
func NewDownloadProgress(artifact, progress string) *DownloadProgress {
	return &DownloadProgress{data: map[string]string{
		"totalsteps":  Reg.totalSteps(),
		"currentstep": Reg.currentStep(),
		"progress":    progress,
		"artifact":    artifact,
	}}
}

// Warning will be used to notify the user of warnings
type Warning struct {
	data map[string]string
}

// NewWarning returns a new warning type
func NewWarning(warning string) *Warning {
	return &Warning{
		map[string]string{
			"message": warning,
		},
	}
}

// Type returns the cloud events compatible type of this struct
func (s *Warning) Type() string {
	return "io.k8s.sigs.minikube.warning"
}

// Info will be used to notify users of any extra info (env variables, options)
type Info struct {
	data map[string]string
}

func (s *Info) Type() string {
	return "io.k8s.sigs.minikube.info"
}

// NewInfo returns a new Info type
func NewInfo(message string) *Info {
	return &Info{
		map[string]string{
			"message": message,
		},
	}
}

// Error will be used to notify the user of errors
type Error struct {
	data map[string]string
}

func NewError(err string) *Error {
	return &Error{
		map[string]string{
			"message": err,
		},
	}
}

// NewErrorExitCode returns an error that has an associated exit code
func NewErrorExitCode(err string, exitcode int, additionalData ...map[string]string) *Error {
	e := NewError(err)
	e.data["exitcode"] = fmt.Sprintf("%v", exitcode)
	for _, a := range additionalData {
		for k, v := range a {
			e.data[k] = v
		}
	}
	return e
}

func (s *Error) Type() string {
	return "io.k8s.sigs.minikube.error"
}
