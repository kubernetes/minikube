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

package mcnflag

import "fmt"

type Flag interface {
	fmt.Stringer
	Default() interface{}
}

type StringFlag struct {
	Name   string
	Usage  string
	EnvVar string
	Value  string
}

// TODO: Could this be done more succinctly using embedding?
func (f StringFlag) String() string {
	return f.Name
}

func (f StringFlag) Default() interface{} {
	return f.Value
}

type StringSliceFlag struct {
	Name   string
	Usage  string
	EnvVar string
	Value  []string
}

// TODO: Could this be done more succinctly using embedding?
func (f StringSliceFlag) String() string {
	return f.Name
}

func (f StringSliceFlag) Default() interface{} {
	return f.Value
}

type IntFlag struct {
	Name   string
	Usage  string
	EnvVar string
	Value  int
}

// TODO: Could this be done more succinctly using embedding?
func (f IntFlag) String() string {
	return f.Name
}

func (f IntFlag) Default() interface{} {
	return f.Value
}

type BoolFlag struct {
	Name   string
	Usage  string
	EnvVar string
}

// TODO: Could this be done more succinctly using embedding?
func (f BoolFlag) String() string {
	return f.Name
}

func (f BoolFlag) Default() interface{} {
	return nil
}
