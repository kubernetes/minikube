// +build !386,!amd64

package virtualbox

// IsVTXDisabled checks if VT-x is disabled in the CPU.
func (d *Driver) IsVTXDisabled() bool {
	return true
}
