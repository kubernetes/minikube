package virtualbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	featuresWithVMX = "FPU VME DE PSE TSC MSR PAE MCE CX8 APIC SEP MTRR PGE MCA CMOV PAT PSE36 VMX PBE SSE3 PCLMULQDQ DTES64 AVX1.0 RDRAND F16C"
	featuresNoVMX   = "FPU VME DE PSE TSC MSR PAE MCE CX8 APIC SEP MTRR PGE MCA CMOV PAT PSE36 PBE SSE3 PCLMULQDQ DTES64 AVX1.0 RDRAND F16C"
)

func TestShareName(t *testing.T) {
	name, dir := getShareDriveAndName()

	assert.Equal(t, name, "Users")
	assert.Equal(t, dir, "/Users")

}

func TestIsVTXEnabled(t *testing.T) {
	assert.False(t, isVTXDisabled(featuresWithVMX))
	assert.True(t, isVTXDisabled(featuresNoVMX))
}
