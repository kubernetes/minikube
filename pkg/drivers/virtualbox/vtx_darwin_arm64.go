//go:build darwin && arm64

/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

// IsVTXDisabled always returns false on darwin/arm64. VT-X/AMD-v are x86
// concepts; Apple Silicon chips have Armv8-A virtualization extensions that
// VirtualBox 7.1+ accesses via Apple's Hypervisor.framework, and there is no
// user-facing toggle on Apple Silicon to disable virtualization. If the host
// hypervisor actually cannot virtualize, VBoxManage will surface the failure
// when the VM is started, so the precreate check does not need to block here.
func (d *Driver) IsVTXDisabled() bool {
	return false
}
