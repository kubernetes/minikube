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

package virtualbox

import "strconv"

type VM struct {
	CPUs   int
	Memory int
}

func getVMInfo(name string, vbox VBoxManager) (*VM, error) {
	out, err := vbox.vbmOut("showvminfo", name, "--machinereadable")
	if err != nil {
		return nil, err
	}

	vm := &VM{}

	err = parseKeyValues(out, reEqualLine, func(key, val string) error {
		switch key {
		case "cpus":
			v, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			vm.CPUs = v
		case "memory":
			v, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			vm.Memory = v
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return vm, nil
}
