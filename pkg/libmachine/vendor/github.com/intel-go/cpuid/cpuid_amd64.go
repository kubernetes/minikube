// +build amd64
// Copyright 2015 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cpuid provides access to the information available
// through the CPUID instruction.
// All information is gathered during package initialization phase
// so package's public interface doesn't call CPUID instruction.
package cpuid

func cpuid_low(arg1, arg2 uint32) (eax, ebx, ecx, edx uint32) // implemented in cpuidlow_amd64.s
func xgetbv_low(arg1 uint32) (eax, edx uint32)                // implemented in cpuidlow_amd64.s

func init() {
	detectFeatures()
}

func detectFeatures() {
	leaf0()
	leaf1()
	leaf2()
	leaf3()
	leaf4()
	leaf5()
	leaf6()
	leaf7()
	leaf0x80000000()
	leaf0x80000001()
	leaf0x80000004()
	leaf0x80000005()
	leaf0x80000006()

	if HasFeature(OSXSAVE) {
		eax, _ := xgetbv_low(0)
		if (eax & 0x6) == 0x6 {
			EnabledAVX = true
		}
		if (eax & 0xE0) == 0xE0 {
			EnabledAVX512 = true
		}
	}
}

var leaf02Names = [...]string{
	"NULL",
	"DATA_CACHE",
	"INSTRUCTION_CACHE",
	"UNIFIED_CACHE",
	"TLB",
	"DTLB",
	"STLB",
	"PREFETCH",
}

func leaf0() {

	eax, ebx, ecx, edx := cpuid_low(0, 0)

	maxInputValue = eax

	VendorIdentificatorString = string(int32sToBytes(ebx, edx, ecx))
	brandId = brandStrings[VendorIdentificatorString]
}

func leaf1() {

	if maxInputValue < 1 {
		return
	}

	eax, ebx, ecx, edx := cpuid_low(1, 0)
	// Parse EAX
	SteppingId = (eax & 0xF)
	modelId := (eax >> 4) & 0xF
	familyId := (eax >> 8) & 0xF
	ProcessorType = (eax >> 12) & 0x3
	ExtendedModelId := (eax >> 16) & 0xF
	extendedFamilyId := (eax >> 20) & 0xFF

	DisplayFamily = familyId
	DisplayModel = modelId

	if familyId == 0xF {
		DisplayFamily = extendedFamilyId + familyId
	}

	if familyId == 0x6 || familyId == 0xF {
		DisplayModel = ExtendedModelId<<4 + modelId
	}

	// Parse EBX
	brandIndex = ebx & 0xFF
	CacheLineSize = ((ebx >> 8) & 0xFF) << 3
	MaxLogocalCPUId = (ebx >> 16) & 0xFF
	InitialAPICId = (ebx >> 24)

	// Parse ECX & EDX not needed. Ask through HasFeature function
	featureFlags = (uint64(edx) << 32) | uint64(ecx)
}

func leaf2() {

	if brandId != INTEL {
		return
	}
	if maxInputValue < 2 {
		return
	}

	bytes := int32sToBytes(cpuid_low(2, 0))

	for i := 0; i < len(bytes); i++ {
		if (i%4 == 0) && (bytes[i+3]&(1<<7) != 0) {
			i += 4
			continue
		}
		if bytes[i] == 0xFF { // it means that we should use leaf 4 for cache info
			CacheDescriptors = CacheDescriptors[0:0]
			break
		}
		CacheDescriptors = append(CacheDescriptors, leaf02Descriptors[int16(bytes[i])])
	}
}

func leaf3() {
	if brandId != INTEL {
		return
	}

	if maxInputValue < 3 {
		return
	}
	// TODO SerialNumber for < Pentium 4
}

func leaf4() {

	if brandId != INTEL {
		return
	}

	if maxInputValue < 4 {
		return
	}

	cacheId := 0
	for {
		eax, ebx, ecx, _ := cpuid_low(4, uint32(cacheId))
		cacheId++
		cacheType := eax & 0xF

		if cacheType == NULL {
			break
		}

		cacheLevel := (eax >> 5) & 0x7
		//		selfInitializingCacheLevel := eax & (1<<8)
		//		fullyAssociativeCache      := eax & (1<<9)
		//		maxNumLogicalCoresSharing  := (eax >> 14) & 0x3FF
		//		maxNumPhisCores            := (eax >> 26) & 0x3F
		systemCoherencyLineSize := (ebx & 0xFFF) + 1
		physicalLinePartions := (ebx>>12)&0x3FF + 1
		waysOfAssiociativity := (ebx>>22)&0x3FF + 1
		numberOfSets := ecx + 1
		//		writeBackInvalidate        := edx & 1
		//		cacheInclusiveness         := edx & (1<<1)
		//		complexCacheIndexing       := edx & (1<<2)
		cacheSize := (waysOfAssiociativity * physicalLinePartions *
			systemCoherencyLineSize * numberOfSets) >> 10
		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{int(cacheLevel),
				int(cacheType),
				"",
				int(cacheSize),
				int(waysOfAssiociativity),
				int(systemCoherencyLineSize),
				int(numberOfSets),
				int(physicalLinePartions),
			})
	}
}

func leaf5() {
	if maxInputValue < 5 {
		return
	}

	eax, ebx, ecx, _ := cpuid_low(4, 0) // TODO process EDX with C0-C7 C-states
	MonLineSizeMax = eax & (0xFFFF)
	MonLineSizeMax = ebx & (0xFFFF)
	MonitorEMX = (ecx & (1 << 0)) != 0
	MonitorIBE = (ecx & (1 << 1)) != 0

}

func leaf6() {
	// Thermal and Power Management Features for Intel
	if maxInputValue < 6 {
		return
	}

	eax, ebx, ecx, _ := cpuid_low(6, 0)
	thermalAndPowerFeatureFlags = (eax & 0xFFFF) | (ecx << 16)
	ThermalSensorInterruptThresholds = ebx & 7
}

func leaf7() {
	_, ebx, ecx, _ := cpuid_low(7, 0)
	extendedFeatureFlags = (uint64(ecx) << 32) | uint64(ebx)
}

func leaf0x80000000() {
	maxExtendedInputValue, _, _, _ = cpuid_low(0x80000000, 0)
}

func leaf0x80000001() {
	if maxExtendedInputValue < 0x80000001 {
		return
	}
	_, _, ecx, edx := cpuid_low(0x80000001, 0)
	//extendedProcessorSignatureAndFeatureBits := eax
	extraFeatureFlags = (uint64(edx) << 32) | uint64(ecx)
}

// leaf0x80000004 looks at the Processor Brand String in leaves 0x80000002 through 0x80000004
func leaf0x80000004() {
	if maxExtendedInputValue < 0x80000004 {
		return
	}

	ProcessorBrandString += string(int32sToBytes(cpuid_low(0x80000002, 0)))
	ProcessorBrandString += string(int32sToBytes(cpuid_low(0x80000003, 0)))
	ProcessorBrandString += string(int32sToBytes(cpuid_low(0x80000004, 0)))
}

func leaf0x80000005() {
	// AMD L1 Cache and TLB Information
	if maxExtendedInputValue < 0x80000005 {
		return
	}

	if brandId != AMD {
		return
	}

	eax, ebx, ecx, edx := cpuid_low(0x80000005, 0)

	L1DTlb2and4MAssoc := (eax >> 24) & 0xFF
	L1DTlb2and4MSize := (eax >> 16) & 0xFF

	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			DTLB,
			"DTLB 2M/4M",
			2 * 1024,
			int(L1DTlb2and4MAssoc),
			-1,
			int(L1DTlb2and4MSize),
			0,
		})

	L1ITlb2and4MAssoc := (eax >> 8) & 0xFF
	L1ITlb2and4MSize := (eax) & 0xFF

	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			TLB,
			"ITLB 2M/4M",
			2 * 1024,
			int(L1ITlb2and4MAssoc),
			-1,
			int(L1ITlb2and4MSize),
			0,
		})

	L1DTlb4KAssoc := (ebx >> 24) & 0xFF
	L1DTlb4KSize := (ebx >> 16) & 0xFF

	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			DTLB,
			"DTLB 4K",
			4,
			int(L1DTlb4KAssoc),
			-1,
			int(L1DTlb4KSize),
			0,
		})

	L1ITlb4KAssoc := (ebx >> 8) & 0xFF
	L1ITlb4KSize := (ebx) & 0xFF

	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			TLB,
			"ITLB 4K",
			4,
			int(L1ITlb4KAssoc),
			-1,
			int(L1ITlb4KSize),
			0,
		})

	L1DcSize := (ecx >> 24) & 0xFF
	L1DcAssoc := (ecx >> 16) & 0xFF
	L1DcLinesPerTag := (ecx >> 8) & 0xFF
	L1DcLineSize := (ecx >> 0) & 0xFF
	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			DATA_CACHE,
			"L1 Data cache",
			int(L1DcSize),
			int(L1DcAssoc),
			int(L1DcLineSize),
			-1,
			int(L1DcLinesPerTag),
		})

	L1IcSize := (edx >> 24) & 0xFF
	L1IcAssoc := (edx >> 16) & 0xFF
	L1IcLinesPerTag := (edx >> 8) & 0xFF
	L1IcLineSize := (edx >> 0) & 0xFF
	CacheDescriptors = append(CacheDescriptors,
		CacheDescriptor{1,
			INSTRUCTION_CACHE,
			"L1 Instruction cache",
			int(L1IcSize),
			int(L1IcAssoc),
			int(L1IcLineSize),
			-1,
			int(L1IcLinesPerTag),
		})
}

func leaf0x80000006() {

	if maxExtendedInputValue < 0x80000006 {
		return
	}

	var associativityEncodings = map[uint]uint{
		0x00: 0,
		0x01: 1,
		0x02: 2,
		0x04: 4,
		0x06: 8,
		0x08: 16,
		0x0A: 32,
		0x0B: 48,
		0x0C: 64,
		0x0D: 96,
		0x0E: 128,
		0x0F: 0xFF, // - Fully associative
	}

	eax, ebx, ecx, edx := cpuid_low(0x80000006, 0)

	if brandId == INTEL {

		CacheLineSize := (ecx >> 0) & 0xFF
		L2Associativity := uint((ecx >> 12) & 0xF)
		CacheSize := (ecx >> 16) & 0xFFFF
		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				0,
				"Cache info from leaf 0x80000006 for Intel",
				int(CacheSize),
				int(associativityEncodings[L2Associativity]),
				int(CacheLineSize),
				-1,
				0,
			})
	}

	if brandId == AMD {

		L2DTlb2and4MAssoc := uint((eax >> 28) & 0xF)
		L2DTlb2and4MSize := (eax >> 16) & 0xFFF

		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				DTLB,
				"DTLB 2M/4M",
				2 * 1024,
				int(associativityEncodings[L2DTlb2and4MAssoc]),
				-1,
				int(L2DTlb2and4MSize),
				0,
			})

		L2ITlb2and4MAssoc := uint((eax >> 12) & 0xF)
		L2ITlb2and4MSize := (eax) & 0xFFF

		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				TLB,
				"ITLB 2M/4M",
				2 * 1024,
				int(associativityEncodings[L2ITlb2and4MAssoc]),
				-1,
				int(L2ITlb2and4MSize),
				0,
			})

		L2DTlb4KAssoc := uint((ebx >> 28) & 0xF)
		L2DTlb4KSize := (ebx >> 16) & 0xFFF

		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				DTLB,
				"DTLB 4K",
				4,
				int(associativityEncodings[L2DTlb4KAssoc]),
				-1,
				int(L2DTlb4KSize),
				0,
			})

		L2ITlb4KAssoc := uint((ebx >> 12) & 0xF)
		L2ITlb4KSize := (ebx) & 0xFFF

		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				TLB,
				"ITLB 4K",
				4,
				int(associativityEncodings[L2ITlb4KAssoc]),
				-1,
				int(L2ITlb4KSize),
				0,
			})

		L2Size := (ecx >> 16) & 0xFFFF
		L2Assoc := uint((ecx >> 12) & 0xF)
		L2LinesPerTag := (ecx >> 8) & 0xF
		L2LineSize := (ecx >> 0) & 0xFF
		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{2,
				DATA_CACHE,
				"L2 Data cache",
				int(L2Size),
				int(associativityEncodings[L2Assoc]),
				int(L2LineSize),
				-1,
				int(L2LinesPerTag),
			})

		L3Size := ((edx >> 18) & 0xF) * 512
		L3Assoc := uint((edx >> 12) & 0xF)
		L3LinesPerTag := (edx >> 8) & 0xF
		L3LineSize := (edx >> 0) & 0xFF
		CacheDescriptors = append(CacheDescriptors,
			CacheDescriptor{3,
				DATA_CACHE,
				"L3 Data cache",
				int(L3Size),
				int(associativityEncodings[L3Assoc]),
				int(L3LineSize),
				-1,
				int(L3LinesPerTag),
			})
	}
}

// TODO split fused descritops with bits in high key's byte like for 0x49
var leaf02Descriptors = map[int16]CacheDescriptor{
	0x01: {-1, TLB, "Instruction TLB", 4, 4, -1, 32, 0},
	0x02: {-1, TLB, "Instruction TLB", 4 * 1024, 0xFF, -1, 2, 0},
	0x03: {-1, TLB, "Data TLB", 4, 4, -1, 64, 0},
	0x04: {-1, TLB, "Data TLB", 4 * 1024, 4, -1, 8, 0},
	0x05: {-1, TLB, "Data TLB1", 4 * 1024, 4, -1, 32, 0},
	0x06: {1, INSTRUCTION_CACHE, "1st-level instruction cache", 8, 4, 32, -1, 0},
	0x08: {1, INSTRUCTION_CACHE, "1st-level instruction cache", 16, 4, 32, -1, 0},
	0x09: {1, INSTRUCTION_CACHE, "1st-level instruction cache", 32, 4, 64, -1, 0},
	0x0A: {1, DATA_CACHE, "1st-level data cache", 8, 2, 32, -1, 0},
	0x0B: {-1, TLB, "Instruction TLB", 4 * 1024, 4, -1, 4, 0},
	0x0C: {1, DATA_CACHE, "1st-level data cache", 16, 4, 32, -1, 0},
	0x0D: {1, DATA_CACHE, "1st-level data cache", 16, 4, 64, -1, 0},
	0x0E: {1, DATA_CACHE, "1st-level data cache", 24, 6, 64, -1, 0},
	0x1D: {2, DATA_CACHE, "2nd-level cache", 128, 2, 64, -1, 0},
	0x21: {2, DATA_CACHE, "2nd-level cache", 256, 8, 64, -1, 0},
	0x22: {3, DATA_CACHE, "3nd-level cache", 512, 4, 64, -1, 2},
	0x23: {3, DATA_CACHE, "3nd-level cache", 1 * 1024, 8, 64, -1, 2},
	0x24: {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 16, 64, -1, 0},
	0x25: {3, DATA_CACHE, "3nd-level cache", 2 * 1024, 8, 64, -1, 2},
	0x29: {3, DATA_CACHE, "2nd-level cache", 4 * 1024, 8, 64, -1, 2},
	0x2C: {1, DATA_CACHE, "1st-level cache", 32, 8, 64, -1, 0},
	0x30: {1, INSTRUCTION_CACHE, "1st-level instruction cache", 32, 8, 64, -1, 0},
	0x40: {-1, DATA_CACHE, "No 2nd-level cache or, if processor contains a " +
		"valid 2nd-level cache, no 3rd-level cache", -1, -1, -1, -1, 0},
	0x41: {2, DATA_CACHE, "2nd-level cache", 128, 4, 32, -1, 0},
	0x42: {2, DATA_CACHE, "2nd-level cache", 256, 4, 32, -1, 0},
	0x43: {2, DATA_CACHE, "2nd-level cache", 512, 4, 32, -1, 0},
	0x44: {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 4, 32, -1, 0},
	0x45: {2, DATA_CACHE, "2nd-level cache", 2 * 1024, 4, 32, -1, 0},
	0x46: {3, DATA_CACHE, "3nd-level cache", 4 * 1024, 4, 64, -1, 0},
	0x47: {3, DATA_CACHE, "3nd-level cache", 8 * 1024, 8, 64, -1, 0},
	0x48: {2, DATA_CACHE, "2nd-level cache", 3 * 1024, 12, 64, -1, 0},
	0x49: {2, DATA_CACHE, "2nd-level cache", 4 * 1024, 16, 64, -1, 0},
	// (Intel Xeon processor MP, Family 0FH, Model 06H)
	(0x49 | (1 << 8)): {3, DATA_CACHE, "3nd-level cache", 4 * 1024, 16, 64, -1, 0},
	0x4A:              {3, DATA_CACHE, "3nd-level cache", 6 * 1024, 12, 64, -1, 0},
	0x4B:              {3, DATA_CACHE, "3nd-level cache", 8 * 1024, 16, 64, -1, 0},
	0x4C:              {3, DATA_CACHE, "3nd-level cache", 12 * 1024, 12, 64, -1, 0},
	0x4D:              {3, DATA_CACHE, "3nd-level cache", 16 * 1024, 16, 64, -1, 0},
	0x4E:              {2, DATA_CACHE, "3nd-level cache", 6 * 1024, 24, 64, -1, 0},
	0x4F:              {-1, TLB, "Instruction TLB", 4, -1, -1, 32, 0},
	0x50:              {-1, TLB, "Instruction TLB: 4 KByte and 2-MByte or 4-MByte pages", 4, -1, -1, 64, 0},
	0x51:              {-1, TLB, "Instruction TLB: 4 KByte and 2-MByte or 4-MByte pages", 4, -1, -1, 128, 0},
	0x52:              {-1, TLB, "Instruction TLB: 4 KByte and 2-MByte or 4-MByte pages", 4, -1, -1, 256, 0},
	0x55:              {-1, TLB, "Instruction TLB: 2-MByte or 4-MByte pages", 2 * 1024, 0xFF, -1, 7, 0},
	0x56:              {-1, TLB, "Data TLB0", 4 * 1024, 4, -1, 16, 0},
	0x57:              {-1, TLB, "Data TLB0", 4, 4, -1, 16, 0},
	0x59:              {-1, TLB, "Data TLB0", 4, 0xFF, -1, 16, 0},
	0x5A:              {-1, TLB, "Data TLB0 2-MByte or 4 MByte pages", 2 * 1024, 4, -1, 32, 0},
	0x5B:              {-1, TLB, "Data TLB 4 KByte and 4 MByte pages", 4, -1, -1, 64, 0},
	0x5C:              {-1, TLB, "Data TLB 4 KByte and 4 MByte pages", 4, -1, -1, 128, 0},
	0x5D:              {-1, TLB, "Data TLB 4 KByte and 4 MByte pages", 4, -1, -1, 256, 0},
	0x60:              {1, DATA_CACHE, "1st-level data cache", 16, 8, 64, -1, 0},
	0x61:              {-1, TLB, "Instruction TLB", 4, 0xFF, -1, 48, 0},
	0x63:              {-1, TLB, "Data TLB", 1 * 1024 * 1024, 4, -1, 4, 0},
	0x66:              {1, DATA_CACHE, "1st-level data cache", 8, 4, 64, -1, 0},
	0x67:              {1, DATA_CACHE, "1st-level data cache", 16, 4, 64, -1, 0},
	0x68:              {1, DATA_CACHE, "1st-level data cache", 32, 4, 64, -1, 0},
	0x70:              {1, INSTRUCTION_CACHE, "Trace cache (size in K of uop)", 12, 8, -1, -1, 0},
	0x71:              {1, INSTRUCTION_CACHE, "Trace cache (size in K of uop)", 16, 8, -1, -1, 0},
	0x72:              {1, INSTRUCTION_CACHE, "Trace cache (size in K of uop)", 32, 8, -1, -1, 0},
	0x76:              {-1, TLB, "Instruction TLB: 2M/4M pages", 2 * 1024, 0xFF, -1, 8, 0},
	0x78:              {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 4, 64, -1, 0},
	0x79:              {2, DATA_CACHE, "2nd-level cache", 128, 8, 64, -1, 2},
	0x7A:              {2, DATA_CACHE, "2nd-level cache", 256, 8, 64, -1, 2},
	0x7B:              {2, DATA_CACHE, "2nd-level cache", 512, 8, 64, -1, 2},
	0x7C:              {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 8, 64, -1, 2},
	0x7D:              {2, DATA_CACHE, "2nd-level cache", 2 * 1024, 8, 64, -1, 0},
	0x7F:              {2, DATA_CACHE, "2nd-level cache", 512, 2, 64, -1, 0},
	0x80:              {2, DATA_CACHE, "2nd-level cache", 512, 8, 64, -1, 0},
	0x82:              {2, DATA_CACHE, "2nd-level cache", 256, 8, 32, -1, 0},
	0x83:              {2, DATA_CACHE, "2nd-level cache", 512, 8, 32, -1, 0},
	0x84:              {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 8, 32, -1, 0},
	0x85:              {2, DATA_CACHE, "2nd-level cache", 2 * 1024, 8, 32, -1, 0},
	0x86:              {2, DATA_CACHE, "2nd-level cache", 512, 4, 32, -1, 0},
	0x87:              {2, DATA_CACHE, "2nd-level cache", 1 * 1024, 8, 64, -1, 0},
	0xA0:              {-1, DTLB, "DTLB", 4, 0xFF, -1, 32, 0},
	0xB0:              {-1, TLB, "Instruction TLB", 4, 4, -1, 128, 0},
	0xB1: {-1, TLB, "Instruction TLB 2M pages 4 way 8 entries or" +
		"4M pages 4-way, 4 entries", 2 * 1024, 4, -1, 8, 0},
	0xB2: {-1, TLB, "Instruction TLB", 4, 4, -1, 64, 0},
	0xB3: {-1, TLB, "Data TLB", 4, 4, -1, 128, 0},
	0xB4: {-1, TLB, "Data TLB1", 4, 4, -1, 256, 0},
	0xB5: {-1, TLB, "Instruction TLB", 4, 8, -1, 64, 0},
	0xB6: {-1, TLB, "Instruction TLB", 4, 8, -1, 128, 0},
	0xBA: {-1, TLB, "Data TLB1", 4, 4, -1, 64, 0},
	0xC0: {-1, TLB, "Data TLB: 4 KByte and 4 MByte pages", 4, 4, -1, 8, 0},
	0xC1: {-1, STLB, "Shared 2nd-Level TLB: 4Kbyte and 2Mbyte pages", 4, 8, -1, 1024, 0},
	0xC2: {-1, DTLB, "DTLB 4KByte/2 MByte pages", 4, 4, -1, 16, 0},
	0xC3: {-1, STLB, "Shared 2nd-Level TLB: " +
		"4 KByte /2 MByte pages, 6-way associative, 1536 entries." +
		"Also 1GBbyte pages, 4-way,16 entries.", 4, 6, -1, 1536, 0},
	0xCA: {-1, STLB, "Shared 2nd-Level TLB", 4, 4, -1, 512, 0},
	0xD0: {3, DATA_CACHE, "3nd-level cache", 512, 4, 64, -1, 0},
	0xD1: {3, DATA_CACHE, "3nd-level cache", 1 * 1024, 4, 64, -1, 0},
	0xD2: {3, DATA_CACHE, "3nd-level cache", 2 * 1024, 4, 64, -1, 0},
	0xD6: {3, DATA_CACHE, "3nd-level cache", 1 * 1024, 8, 64, -1, 0},
	0xD7: {3, DATA_CACHE, "3nd-level cache", 2 * 1024, 8, 64, -1, 0},
	0xD8: {3, DATA_CACHE, "3nd-level cache", 4 * 1024, 8, 64, -1, 0},
	0xDC: {3, DATA_CACHE, "3nd-level cache", 1 * 1536, 12, 64, -1, 0},
	0xDD: {3, DATA_CACHE, "3nd-level cache", 3 * 1024, 12, 64, -1, 0},
	0xDE: {3, DATA_CACHE, "3nd-level cache", 6 * 1024, 12, 64, -1, 0},
	0xE2: {3, DATA_CACHE, "3nd-level cache", 2 * 1024, 16, 64, -1, 0},
	0xE3: {3, DATA_CACHE, "3nd-level cache", 4 * 1024, 16, 64, -1, 0},
	0xE4: {3, DATA_CACHE, "3nd-level cache", 8 * 1024, 16, 64, -1, 0},
	0xEA: {3, DATA_CACHE, "3nd-level cache", 12 * 1024, 24, 64, -1, 0},
	0xEB: {3, DATA_CACHE, "3nd-level cache", 18 * 1024, 24, 64, -1, 0},
	0xEC: {3, DATA_CACHE, "3nd-level cache", 24 * 1024, 24, 64, -1, 0},
	0xF0: {-1, PREFETCH, "", 64, -1, -1, -1, 0},
	0xF1: {-1, PREFETCH, "", 128, -1, -1, -1, 0},
	0xFF: {-1, NULL, "CPUID leaf 2 does not report cache descriptor " +
		"information, use CPUID leaf 4 to query cache parameters",
		-1, -1, -1, -1, 0},
}

func int32sToBytes(args ...uint32) []byte {
	var result []byte

	for _, arg := range args {
		result = append(result,
			byte((arg)&0xFF),
			byte((arg>>8)&0xFF),
			byte((arg>>16)&0xFF),
			byte((arg>>24)&0xFF))
	}

	return result
}
