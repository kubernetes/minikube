// Copyright 2015 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cpuid provides access to the information available
// through the CPUID instruction.
// All information is gathered during package initialization phase
// so package's public interface doesn't call CPUID instruction.
package cpuid

// VendorIdentificationString like "GenuineIntel" or "AuthenticAMD"
var VendorIdentificatorString string

// ProcessorBrandString like "Intel(R) Core(TM) i7-4770HQ CPU @ 2.20GHz"
var ProcessorBrandString string

// SteppingId is Processor Stepping ID as described in
// Intel® 64 and IA-32 Architectures Software Developer’s Manual
var SteppingId uint32

// ProcessorType obtained from processor Version Information, according to
// Intel® 64 and IA-32 Architectures Software Developer’s Manual
var ProcessorType uint32

// DisplayFamily is Family of processors obtained from processor Version Information, according to
// Intel® 64 and IA-32 Architectures Software Developer’s Manual
var DisplayFamily uint32

// Display Model is Model of processor obtained from processor Version Information, according to
// Intel® 64 and IA-32 Architectures Software Developer’s Manual
var DisplayModel uint32

// Cache line size in bytes
var CacheLineSize uint32

// Maximum number of addressable IDs for logical processors in this physical package
var MaxLogocalCPUId uint32

// Initial APIC ID
var InitialAPICId uint32

// Cache descriptor's array
// You can iterate like there:
// 	for _, cacheDescription := range cpuid.CacheDescriptors {
//		fmt.Printf("CacheDescriptor: %v\n", cacheDescription)
//	}
// See CacheDescriptor type for more information
var CacheDescriptors []CacheDescriptor

// Smallest monitor-line size in bytes (default is processor's monitor granularity)
var MonLineSizeMin uint32

// Largest monitor-line size in bytes (default is processor's monitor granularity)
var MonLineSizeMax uint32

// Enumeration of Monitor-Mwait extensions availability status
var MonitorEMX bool

// Supports treating interrupts as break-event for MWAIT flag
var MonitorIBE bool

// EnabledAVX flag allows to check if feature AVX is enabled by OS/BIOS
var EnabledAVX bool = false

// EnabledAVX512 flag allows to check if features AVX512xxx are enabled by OS/BIOS
var EnabledAVX512 bool = false

type CacheDescriptor struct {
	Level      int    // Cache level
	CacheType  int    // Cache type
	CacheName  string // Name
	CacheSize  int    // in KBytes (of page size for TLB)
	Ways       int    // Associativity, 0 undefined, 0xFF fully associate
	LineSize   int    // Cache line size in bytes
	Entries    int    // number of entries for TLB
	Partioning int    // partitioning
}

// ThermalSensorInterruptThresholds is the number of interrupt thresholds in digital thermal sensor.
var ThermalSensorInterruptThresholds uint32

// HasFeature to check if features from FeatureNames map are available on the current processor
func HasFeature(feature uint64) bool {
	return (featureFlags & feature) != 0
}

// HasExtendedFeature to check if features from ExtendedFeatureNames map are available on the current processor
func HasExtendedFeature(feature uint64) bool {
	return (extendedFeatureFlags & feature) != 0
}

// HasExtraFeature to check if features from ExtraFeatureNames map are available on the current processor
func HasExtraFeature(feature uint64) bool {
	return (extraFeatureFlags & feature) != 0
}

// HasThermalAndPowerFeature to check if features from ThermalAndPowerFeatureNames map are available on the current processor
func HasThermalAndPowerFeature(feature uint32) bool {
	return (thermalAndPowerFeatureFlags & feature) != 0
}

var FeatureNames = map[uint64]string{
	SSE3:         "SSE3",
	PCLMULQDQ:    "PCLMULQDQ",
	DTES64:       "DTES64",
	MONITOR:      "MONITOR",
	DSI_CPL:      "DSI_CPL",
	VMX:          "VMX",
	SMX:          "SMX",
	EST:          "EST",
	TM2:          "TM2",
	SSSE3:        "SSSE3",
	CNXT_ID:      "CNXT_ID",
	SDBG:         "SDBG",
	FMA:          "FMA",
	CX16:         "CX16",
	XTPR:         "XTPR",
	PDCM:         "PDCM",
	PCID:         "PCID",
	DCA:          "DCA",
	SSE4_1:       "SSE4_1",
	SSE4_2:       "SSE4_2",
	X2APIC:       "X2APIC",
	MOVBE:        "MOVBE",
	POPCNT:       "POPCNT",
	TSC_DEADLINE: "TSC_DEADLINE",
	AES:          "AES",
	XSAVE:        "XSAVE",
	OSXSAVE:      "OSXSAVE",
	AVX:          "AVX",
	F16C:         "F16C",
	RDRND:        "RDRND",
	HYPERVISOR:   "HYPERVISOR",
	FPU:          "FPU",
	VME:          "VME",
	DE:           "DE",
	PSE:          "PSE",
	TSC:          "TSC",
	MSR:          "MSR",
	PAE:          "PAE",
	MCE:          "MCE",
	CX8:          "CX8",
	APIC:         "APIC",
	SEP:          "SEP",
	MTRR:         "MTRR",
	PGE:          "PGE",
	MCA:          "MCA",
	CMOV:         "CMOV",
	PAT:          "PAT",
	PSE_36:       "PSE_36",
	PSN:          "PSN",
	CLFSH:        "CLFSH",
	DS:           "DS",
	ACPI:         "ACPI",
	MMX:          "MMX",
	FXSR:         "FXSR",
	SSE:          "SSE",
	SSE2:         "SSE2",
	SS:           "SS",
	HTT:          "HTT",
	TM:           "TM",
	IA64:         "IA64",
	PBE:          "PBE",
}

var ThermalAndPowerFeatureNames = map[uint32]string{ // From leaf06
	ARAT:                      "ARAT",
	PLN:                       "PLN",
	ECMD:                      "ECMD",
	PTM:                       "PTM",
	HDC:                       "HDC",
	HCFC:                      "HCFC",
	HWP:                       "HWP",
	HWP_NOTIF:                 "HWP_NOTIF",
	HWP_ACTIVITY_WINDOW:       "HWP_ACTIVITY_WINDOW",
	HWP_ENERGY_PERFORMANCE:    "HWP_ENERGY_PERFORMANCE",
	HWP_PACKAGE_LEVEL_REQUEST: "HWP_PACKAGE_LEVEL_REQUEST",
	PERFORMANCE_ENERGY_BIAS:   "PERFORMANCE_ENERGY_BIAS",
	TEMPERATURE_SENSOR:        "TEMPERATURE_SENSOR",
	TURBO_BOOST:               "TURBO_BOOST",
	TURBO_BOOST_MAX:           "TURBO_BOOST_MAX",
}

var ExtendedFeatureNames = map[uint64]string{ // From leaf07
	FSGSBASE:        "FSGSBASE",
	IA32_TSC_ADJUST: "IA32_TSC_ADJUST",
	BMI1:            "BMI1",
	HLE:             "HLE",
	AVX2:            "AVX2",
	SMEP:            "SMEP",
	BMI2:            "BMI2",
	ERMS:            "ERMS",
	INVPCID:         "INVPCID",
	RTM:             "RTM",
	PQM:             "PQM",
	DFPUCDS:         "DFPUCDS",
	MPX:             "MPX",
	PQE:             "PQE",
	AVX512F:         "AVX512F",
	AVX512DQ:        "AVX512DQ",
	RDSEED:          "RDSEED",
	ADX:             "ADX",
	SMAP:            "SMAP",
	AVX512IFMA:      "AVX512IFMA",
	PCOMMIT:         "PCOMMIT",
	CLFLUSHOPT:      "CLFLUSHOPT",
	CLWB:            "CLWB",
	INTEL_PROCESSOR_TRACE: "INTEL_PROCESSOR_TRACE",
	AVX512PF:              "AVX512PF",
	AVX512ER:              "AVX512ER",
	AVX512CD:              "AVX512CD",
	SHA:                   "SHA",
	AVX512BW:              "AVX512BW",
	AVX512VL:              "AVX512VL",
	PREFETCHWT1:           "PREFETCHWT1",
	AVX512VBMI:            "AVX512VBMI",
}

var ExtraFeatureNames = map[uint64]string{ // From leaf 8000 0001
	LAHF_LM:      "LAHF_LM",
	CMP_LEGACY:   "CMP_LEGACY",
	SVM:          "SVM",
	EXTAPIC:      "EXTAPIC",
	CR8_LEGACY:   "CR8_LEGACY",
	ABM:          "ABM",
	SSE4A:        "SSE4A",
	MISALIGNSSE:  "MISALIGNSSE",
	PREFETCHW:    "PREFETCHW",
	OSVW:         "OSVW",
	IBS:          "IBS",
	XOP:          "XOP",
	SKINIT:       "SKINIT",
	WDT:          "WDT",
	LWP:          "LWP",
	FMA4:         "FMA4",
	TCE:          "TCE",
	NODEID_MSR:   "NODEID_MSR",
	TBM:          "TBM",
	TOPOEXT:      "TOPOEXT",
	PERFCTR_CORE: "PERFCTR_CORE",
	PERFCTR_NB:   "PERFCTR_NB",
	SPM:          "SPM",
	DBX:          "DBX",
	PERFTSC:      "PERFTSC",
	PCX_L2I:      "PCX_L2I",
	FPU_2:        "FPU",
	VME_2:        "VME",
	DE_2:         "DE",
	PSE_2:        "PSE",
	TSC_2:        "TSC",
	MSR_2:        "MSR",
	PAE_2:        "PAE",
	MCE_2:        "MCE",
	CX8_2:        "CX8",
	APIC_2:       "APIC",
	SYSCALL:      "SYSCALL",
	MTRR_2:       "MTRR",
	PGE_2:        "PGE",
	MCA_2:        "MCA",
	CMOV_2:       "CMOV",
	PAT_2:        "PAT",
	PSE36:        "PSE36",
	MP:           "MP",
	NX:           "NX",
	MMXEXT:       "MMXEXT",
	MMX_2:        "MMX",
	FXSR_2:       "FXSR",
	FXSR_OPT:     "FXSR_OPT",
	PDPE1GB:      "PDPE1GB",
	RDTSCP:       "RDTSCP",
	LM:           "LM",
	_3DNOWEXT:    "3DNOWEXT",
	_3DNOW:       "3DNOW",
}

var brandStrings = map[string]int{
	"AMDisbetter!": AMD,
	"AuthenticAMD": AMD,
	"CentaurHauls": CENTAUR,
	"CyrixInstead": CYRIX,
	"GenuineIntel": INTEL,
	"TransmetaCPU": TRANSMETA,
	"GenuineTMx86": TRANSMETA,
	"Geode by NSC": NATIONALSEMICONDUCTOR,
	"NexGenDriven": NEXGEN,
	"RiseRiseRise": RISE,
	"SiS SiS SiS ": SIS,
	"UMC UMC UMC ": UMC,
	"VIA VIA VIA ": VIA,
	"Vortex86 SoC": VORTEX,
	"KVMKVMKVM":    KVM,
	"Microsoft Hv": HYPERV,
	"VMwareVMware": VMWARE,
	"XenVMMXenVMM": XEN,
}

var maxInputValue uint32
var maxExtendedInputValue uint32
var extendedModelId uint32
var extendedFamilyId uint32
var brandIndex uint32
var brandId int
var featureFlags uint64
var thermalAndPowerFeatureFlags uint32
var extendedFeatureFlags uint64
var extraFeatureFlags uint64

const (
	UKNOWN = iota
	AMD
	CENTAUR
	CYRIX
	INTEL
	TRANSMETA
	NATIONALSEMICONDUCTOR
	NEXGEN
	RISE
	SIS
	UMC
	VIA
	VORTEX
	KVM
	HYPERV
	VMWARE
	XEN
)

const (
	SSE3 = uint64(1) << iota
	PCLMULQDQ
	DTES64
	MONITOR
	DSI_CPL
	VMX
	SMX
	EST
	TM2
	SSSE3
	CNXT_ID
	SDBG
	FMA
	CX16
	XTPR
	PDCM
	_
	PCID
	DCA
	SSE4_1
	SSE4_2
	X2APIC
	MOVBE
	POPCNT
	TSC_DEADLINE
	AES
	XSAVE
	OSXSAVE
	AVX
	F16C
	RDRND
	HYPERVISOR
	FPU
	VME
	DE
	PSE
	TSC
	MSR
	PAE
	MCE
	CX8
	APIC
	_
	SEP
	MTRR
	PGE
	MCA
	CMOV
	PAT
	PSE_36
	PSN
	CLFSH
	_
	DS
	ACPI
	MMX
	FXSR
	SSE
	SSE2
	SS
	HTT
	TM
	IA64
	PBE
)

const (
	FSGSBASE = uint64(1) << iota
	IA32_TSC_ADJUST
	_
	BMI1
	HLE
	AVX2
	_
	SMEP
	BMI2
	ERMS
	INVPCID
	RTM
	PQM
	DFPUCDS
	MPX
	PQE
	AVX512F
	AVX512DQ
	RDSEED
	ADX
	SMAP
	AVX512IFMA
	PCOMMIT
	CLFLUSHOPT
	CLWB
	INTEL_PROCESSOR_TRACE
	AVX512PF
	AVX512ER
	AVX512CD
	SHA
	AVX512BW
	AVX512VL
	// ECX's const from there
	PREFETCHWT1
	AVX512VBMI
)

const (
	LAHF_LM = uint64(1) << iota
	CMP_LEGACY
	SVM
	EXTAPIC
	CR8_LEGACY
	ABM
	SSE4A
	MISALIGNSSE
	PREFETCHW
	OSVW
	IBS
	XOP
	SKINIT
	WDT
	_
	LWP
	FMA4
	TCE
	_
	NODEID_MSR
	_
	TBM
	TOPOEXT
	PERFCTR_CORE
	PERFCTR_NB
	SPM
	DBX
	PERFTSC
	PCX_L2I
	_
	_
	_
	// EDX features from there
	FPU_2
	VME_2
	DE_2
	PSE_2
	TSC_2
	MSR_2
	PAE_2
	MCE_2
	CX8_2
	APIC_2
	_
	SYSCALL
	MTRR_2
	PGE_2
	MCA_2
	CMOV_2
	PAT_2
	PSE36
	_
	MP
	NX
	_
	MMXEXT
	MMX_2
	FXSR_2
	FXSR_OPT
	PDPE1GB
	RDTSCP
	_
	LM
	_3DNOWEXT
	_3DNOW
)

// Thermal and Power Management features
const (
	// EAX bits 0-15
	TEMPERATURE_SENSOR        = uint32(1) << iota // Digital temperature sensor
	TURBO_BOOST                                   // Intel Turbo Boost Technology available
	ARAT                                          // APIC-Timer-always-running feature is supported if set.
	_                                             // Reserved
	PLN                                           // Power limit notification controls
	ECMD                                          // Clock modulation duty cycle extension
	PTM                                           // Package thermal management
	HWP                                           // HWP base registers (IA32_PM_ENABLE[bit 0], IA32_HWP_CAPABILITIES, IA32_HWP_REQUEST, IA32_HWP_STATUS)
	HWP_NOTIF                                     // IA32_HWP_INTERRUPT MSR
	HWP_ACTIVITY_WINDOW                           // IA32_HWP_REQUEST[bits 41:32]
	HWP_ENERGY_PERFORMANCE                        // IA32_HWP_REQUEST[bits 31:24]
	HWP_PACKAGE_LEVEL_REQUEST                     // IA32_HWP_REQUEST_PKG MSR
	_                                             // Reserved (eax bit 12)
	HDC                                           // HDC base registers IA32_PKG_HDC_CTL, IA32_PM_CTL1, IA32_THREAD_STALL MSRs
	TURBO_BOOST_MAX                               // Intel® Turbo Boost Max Technology
	_                                             // Reserved (eax bit 15)

	// ECX bits 0-15
	HCFC // Hardware Coordination Feedback Capability
	_
	_
	PERFORMANCE_ENERGY_BIAS // Processor supports performance-energy bias preference
)

const (
	NULL = iota
	DATA_CACHE
	INSTRUCTION_CACHE
	UNIFIED_CACHE
	TLB
	DTLB
	STLB
	PREFETCH
)
