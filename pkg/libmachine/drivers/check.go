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

package drivers

import "k8s.io/minikube/pkg/libmachine/mcnflag"

// CheckDriverOptions implements DriverOptions and is used to validate flag parsing
type CheckDriverOptions struct {
	FlagsValues  map[string]interface{}
	CreateFlags  []mcnflag.Flag
	InvalidFlags []string
}

func (o *CheckDriverOptions) String(key string) string {
	for _, flag := range o.CreateFlags {
		if flag.String() == key {
			f, ok := flag.(mcnflag.StringFlag)
			if !ok {
				o.InvalidFlags = append(o.InvalidFlags, flag.String())
			}

			value, present := o.FlagsValues[key].(string)
			if present {
				return value
			}
			return f.Value
		}
	}

	return ""
}

func (o *CheckDriverOptions) StringSlice(key string) []string {
	for _, flag := range o.CreateFlags {
		if flag.String() == key {
			f, ok := flag.(mcnflag.StringSliceFlag)
			if !ok {
				o.InvalidFlags = append(o.InvalidFlags, flag.String())
			}

			value, present := o.FlagsValues[key].([]string)
			if present {
				return value
			}
			return f.Value
		}
	}

	return nil
}

func (o *CheckDriverOptions) Int(key string) int {
	for _, flag := range o.CreateFlags {
		if flag.String() == key {
			f, ok := flag.(mcnflag.IntFlag)
			if !ok {
				o.InvalidFlags = append(o.InvalidFlags, flag.String())
			}

			value, present := o.FlagsValues[key].(int)
			if present {
				return value
			}
			return f.Value
		}
	}

	return 0
}

func (o *CheckDriverOptions) Bool(key string) bool {
	for _, flag := range o.CreateFlags {
		if flag.String() == key {
			_, ok := flag.(mcnflag.BoolFlag)
			if !ok {
				o.InvalidFlags = append(o.InvalidFlags, flag.String())
			}
		}
	}

	value, present := o.FlagsValues[key].(bool)
	if present {
		return value
	}
	return false
}
