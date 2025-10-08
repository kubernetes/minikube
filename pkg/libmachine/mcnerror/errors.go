/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package mcnerror

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/minikube/pkg/libmachine/state"
)

var (
	ErrInvalidHostname = errors.New("Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -")
)

type ErrHostDoesNotExist struct {
	Name string
}

func (e ErrHostDoesNotExist) Error() string {
	return fmt.Sprintf("Docker machine %q does not exist. Use \"docker-machine ls\" to list machines. Use \"docker-machine create\" to add a new one.", e.Name)
}

type ErrHostAlreadyExists struct {
	Name string
}

func (e ErrHostAlreadyExists) Error() string {
	return fmt.Sprintf("Docker machine %q already exists", e.Name)
}

type ErrDuringPreCreate struct {
	Cause error
}

func (e ErrDuringPreCreate) Error() string {
	return fmt.Sprintf("Error with pre-create check: %q", e.Cause)
}

type ErrHostAlreadyInState struct {
	Name  string
	State state.State
}

func (e ErrHostAlreadyInState) Error() string {
	return fmt.Sprintf("Machine %q is already %s.", e.Name, strings.ToLower(e.State.String()))
}
