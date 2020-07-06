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

package out

const (
	InitialSetup         Step = "Initial Minikube Setup"
	SelectingDriver      Step = "Selecting Driver"
	DownloadingArtifacts Step = "Downloading Artifacts"
	StartingNode         Step = "Starting Node"
	PreparingKubernetes  Step = "Preparing Kubernetes"
	VerifyingKubernetes  Step = "Verifying Kubernetes"
	EnablingAddons       Step = "Enabling Addons"
	Done                 Step = "Done"
)

// Step is a type representing a distinct step of `minikube start`
type Step string

// Register holds all of the steps we could see in `minikube start`
// and keeps track of the current step
type Register struct {
	steps   []Step
	current Step
}

// Reg is a package level register that keep track
// of the current step we are on
var Reg Register

func init() {
	Reg = Register{
		steps: []Step{
			InitialSetup,
			SelectingDriver,
			DownloadingArtifacts,
			StartingNode,
			PreparingKubernetes,
			VerifyingKubernetes,
			EnablingAddons,
			Done,
		},
		current: InitialSetup,
	}
}

// totalSteps returns the total number of steps in the register
func (r *Register) totalSteps() int {
	return len(r.steps)
}

// currentStep returns the current step we are on
func (r *Register) currentStep() int {
	for i, s := range r.steps {
		if r.current == s {
			return i
		}
	}
	return -1
}

// SetStep sets the current step
func (r *Register) SetStep(s Step) {
	r.current = s
}
