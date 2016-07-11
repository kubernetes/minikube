/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package cmd

import "github.com/docker/go-units"

// unitValue represents an int64 flag specified with units as a string.
type unitValue int64

func newUnitValue(v int64) *unitValue {
	return (*unitValue)(&v)
}

func (u *unitValue) Set(s string) error {
	v, err := units.FromHumanSize(s)
	*u = unitValue(v)
	return err
}

func (u *unitValue) Type() string {
	return "unit"
}

func (u *unitValue) String() string {
	return units.HumanSize(float64(*u))
}
