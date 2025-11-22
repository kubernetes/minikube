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
