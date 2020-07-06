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

// Log represents the different types of logs that can be output as JSON
// This includes: Step, Download, DownloadProgress, Warning, Info, Error
type Log interface {
	Type() string
}

type Step struct {
	data map[string]string
}

func (s *Step) Type() string {
	return "io.k8s.sigs.minikube.step"
}

func NewStep(message string) *Step {
	return &Step{data: map[string]string{
		"totalsteps":  string(Reg.totalSteps()),
		"currentstep": string(Reg.currentStep()),
		"message":     message,
		"name":        string(Reg.current),
	}}
}

type Download struct {
}

func (s *Download) Type() string {
	return "io.k8s.sigs.minikube.download"
}

type DownloadProgress struct {
}

func (s *DownloadProgress) Type() string {
	return "io.k8s.sigs.minikube.download.progress"
}

type Warning struct {
}

func (s *Warning) Type() string {
	return "io.k8s.sigs.minikube.warning"
}

type Info struct {
}

func (s *Info) Type() string {
	return "io.k8s.sigs.minikube.info"
}

type Error struct {
}

func (s *Error) Type() string {
	return "io.k8s.sigs.minikube.error"
}
