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

import "strings"

// IsVTXDisabledInTheVM checks if VT-X is disabled in the started vm.
func (d *Driver) IsVTXDisabledInTheVM() (bool, error) {
	lines, err := d.readVBoxLog()
	if err != nil {
		return true, err
	}

	for _, line := range lines {
		if strings.Contains(line, "VT-x is disabled") && !strings.Contains(line, "Falling back to raw-mode: VT-x is disabled in the BIOS for all CPU modes") {
			return true, nil
		}
		if strings.Contains(line, "the host CPU does NOT support HW virtualization") {
			return true, nil
		}
		if strings.Contains(line, "VERR_VMX_UNABLE_TO_START_VM") {
			return true, nil
		}
		if strings.Contains(line, "Power up failed") && strings.Contains(line, "VERR_VMX_NO_VMX") {
			return true, nil
		}
	}

	return false, nil
}
