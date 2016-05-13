package virtualbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	amdCPUInfo = `
processor	: 0
vendor_id	: AuthenticAMD
cpu family	: 20
model		: 1
model name	: AMD C-50 Processor
stepping	: 0
microcode	: 0x5000026
cpu MHz		: 800.000
cache size	: 512 KB
physical id	: 0
siblings	: 2
core id		: 0
cpu cores	: 2
apicid		: 0
initial apicid	: 0
fpu		: yes
fpu_exception	: yes
cpuid level	: 6
wp		: yes
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ht syscall nx mmxext fxsr_opt pdpe1gb rdtscp lm constant_tsc rep_good nopl nonstop_tsc extd_apicid aperfmperf pni monitor ssse3 cx16 popcnt lahf_lm cmp_legacy svm extapic cr8_legacy abm sse4a misalignsse 3dnowprefetch ibs skinit wdt arat hw_pstate npt lbrv svm_lock nrip_save pausefilter vmmcall
bugs		: fxsave_leak sysret_ss_attrs
bogomips	: 1995.09
TLB size	: 1024 4K pages
clflush size	: 64
cache_alignment	: 64
address sizes	: 36 bits physical, 48 bits virtual
power management: ts ttp tm stc 100mhzsteps hwpstate
`
	intelCPUInfo = `
processor	: 0
vendor_id	: GenuineIntel
cpu family	: 6
model		: 70
model name	: Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz
stepping	: 1
microcode	: 0x19
cpu MHz		: 2294.688
cache size	: 6144 KB
physical id	: 0
siblings	: 1
core id		: 0
cpu cores	: 1
apicid		: 0
initial apicid	: 0
fpu		: yes
fpu_exception	: yes
cpuid level	: 13
wp		: yes
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr vmx pge mca cmov pat pse36 clflush mmx fxsr sse sse2 syscall nx rdtscp lm constant_tsc rep_good nopl xtopology nonstop_tsc pni pclmulqdq monitor ssse3 cx16 sse4_1 sse4_2 movbe popcnt aes xsave avx rdrand hypervisor lahf_lm abm
bugs		:
bogomips	: 4589.37
clflush size	: 64
cache_alignment	: 64
address sizes	: 39 bits physical, 48 bits virtual
power management:
`

	faultyCPUInfo = `
processor	: 0
vendor_id	: GenuineIntel
cpu family	: 6
model		: 70
model name	: Intel(R) Core(TM) i7-4850HQ CPU @ 2.30GHz
stepping	: 1
microcode	: 0x19
cpu MHz		: 2294.688
cache size	: 6144 KB
physical id	: 0
siblings	: 1
core id		: 0
cpu cores	: 1
apicid		: 0
initial apicid	: 0
fpu		: yes
fpu_exception	: yes
cpuid level	: 13
wp		: yes
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 syscall nx rdtscp lm constant_tsc rep_good nopl xtopology nonstop_tsc pni pclmulqdq monitor ssse3 cx16 sse4_1 sse4_2 movbe popcnt aes xsave avx rdrand hypervisor lahf_lm abm
bugs		:
bogomips	: 4589.37
clflush size	: 64
cache_alignment	: 64
address sizes	: 39 bits physical, 48 bits virtual
power management:
`
)

func TestShareName(t *testing.T) {
	name, dir := getShareDriveAndName()

	assert.Equal(t, name, "hosthome")
	assert.Equal(t, dir, "/home")
}

func TestCpuInfoOnAMD(t *testing.T) {
	assert.False(t, isVTXDisabled([]byte(amdCPUInfo)))
}

func TestCpuInfoOnIntel(t *testing.T) {
	assert.False(t, isVTXDisabled([]byte(intelCPUInfo)))
}

func TestCpuInfoOnNone(t *testing.T) {
	assert.True(t, isVTXDisabled([]byte(faultyCPUInfo)))
}
