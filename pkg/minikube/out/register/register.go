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

// Package register contains all the logic to print out `minikube start` in JSON
package register

import (
	"fmt"
)

const (
	InitialSetup         RegStep = "Initial Minikube Setup"
	SelectingDriver      RegStep = "Selecting Driver"
	DownloadingArtifacts RegStep = "Downloading Artifacts"
	StartingNode         RegStep = "Starting Node"
	RunningLocalhost     RegStep = "Running on Localhost"
	LocalOSRelease       RegStep = "Local OS Release"
	CreatingContainer    RegStep = "Creating Container"
	CreatingVM           RegStep = "Creating VM"
	ConfiguringLHEnv     RegStep = "Configuring Localhost Environment"
	PreparingKubernetes  RegStep = "Preparing Kubernetes"
	VerifyingKubernetes  RegStep = "Verifying Kubernetes"
	EnablingAddons       RegStep = "Enabling Addons"
	Done                 RegStep = "Done"
)

// RegStep is a type representing a distinct step of `minikube start`
type RegStep string

// Register holds all of the steps we could see in `minikube start`
// and keeps track of the current step
type Register struct {
	steps   []RegStep
	current RegStep
}

// Reg keeps track of all possible steps and the current step we are on
var Reg Register

func init() {
	Reg = Register{
		steps: []RegStep{
			InitialSetup,
			SelectingDriver,
			DownloadingArtifacts,
			StartingNode,
			RunningLocalhost,
			LocalOSRelease,
			CreatingContainer,
			CreatingVM,
			PreparingKubernetes,
			ConfiguringLHEnv,
			VerifyingKubernetes,
			EnablingAddons,
			Done,
		},
		current: InitialSetup,
	}
}

// totalSteps returns the total number of steps in the register
func (r *Register) totalSteps() string {
	return fmt.Sprintf("%d", len(r.steps)-1)
}

// currentStep returns the current step we are on
func (r *Register) currentStep() string {
	for i, s := range r.steps {
		if r.current == s {
			return fmt.Sprintf("%d", i)
		}
	}
	// all steps should be registered so this shouldn't happen
	// can't call exit.WithError as it creates an import dependency loop
	panic(fmt.Sprintf("%v is not a registered step", r.current))
}

// SetStep sets the current step
func (r *Register) SetStep(s RegStep) {
	r.current = s
}

// recordStep records the current step
