// +build 386 amd64

package virtualbox

import "github.com/intel-go/cpuid"

// IsVTXDisabled checks if VT-x is disabled in the CPU.
func (d *Driver) IsVTXDisabled() bool {
	if cpuid.HasFeature(cpuid.VMX) || cpuid.HasExtraFeature(cpuid.SVM) {
		return false
	}

	return true
}
