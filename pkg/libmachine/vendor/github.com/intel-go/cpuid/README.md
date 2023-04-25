# cpuid
### Intel CPUID library for Go Programming Language

The cpuid package provides convenient and fast access to information from 
the x86 CPUID instruction. 
The package gathers all information during package initialization phase 
so its public interface will not need to execute the CPUID instruction at runtime.
Frequent calls to the CPUID instruction can hurt performance, 
so this package makes it easier to do CPU-specific optimizations.

[![GoDoc](https://godoc.org/github.com/intel-go/cpuid?status.svg)](https://godoc.org/github.com/intel-go/cpuid)

### You can get it with

```shell
go get github.com/intel-go/cpuid
```

### Example:

```go
package main

import (
    "github.com/intel-go/cpuid"
    "fmt"
)

func main() {
    fmt.Printf("VendorString:   %s\n", cpuid.VendorIdentificatorString)

    fmt.Printf("Features: ")
    for i := uint64(0); i < 64; i++ {
        if cpuid.HasFeature(1 << i) {
            fmt.Printf("%s ", cpuid.FeatureNames[1<<i])
        }
    }
    fmt.Printf("\n")

    fmt.Printf("ExtendedFeatures: ")
    for i := uint64(0); i < 64; i++ {
        if cpuid.HasExtendedFeature(1 << i) {
            fmt.Printf("%s ", cpuid.ExtendedFeatureNames[1<<i])
        }
    }
    fmt.Printf("\n")

    fmt.Printf("ExtraFeatures: ")
    for i := uint64(0); i < 64; i++ {
        if cpuid.HasExtraFeature(1 << i) {
            fmt.Printf("%s ", cpuid.ExtraFeatureNames[1<<i])
        }
    }
    fmt.Printf("\n")
}
```

### API description
Most data is available with simple variables:
 
* **SteppingId uint32** Processor Stepping ID
* **ProcessorType uint32** Processor type
* **DisplayFamily uint32** Processor family
* **DisplayModel uint32** Processor model
* **CacheLineSize uint32** Cache line size in bytes
* **MaxLogocalCPUId uint32** Maximum number of addressable IDs for logical processors in this physical package
* **InitialAPICId uint32** Initial APIC ID
* **CacheDescriptors []CacheDescriptor** Cache descriptor's array

You can iterate over them as follows:
     
```go
for _, cacheDescription := range cpuid.CacheDescriptors {
    fmt.Printf("CacheDescriptor: %v\n", cacheDescription)
}
```

* **MonLineSizeMin uint32** Smallest monitor-line size in bytes (default is processor's monitor granularity) 
* **MonLineSizeMax uint32** Largest monitor-line size in bytes (default is processor's monitor granularity)
* **MonitorEMX bool** Enumeration of Monitor-Mwait extensions availability status
* **MonitorIBE bool** Supports treating interrupts as break-event for MWAIT flag
* **EnabledAVX bool** EnabledAVX flag allows to check if feature AVX is enabled by OS/BIOS
* **EnabledAVX512 bool** EnabledAVX512 flag allows to check if features AVX512xxx are enabled by OS/BIOS
* **func HasFeature(feature uint64) bool** to check for the following features:

  > **SSE3**         Prescott New Instructions-SSE3 (PNI) <br/>
  > **PCLMULQDQ**    PCLMULQDQ support <br/>
  > **DTES64**       64-bit debug store (edx bit 21) <br/>
  > **MONITOR**      MONITOR and MWAIT instructions (SSE3) <br/>
  > **DSI_CPL**      CPL qualified debug store <br/>
  > **VMX**          Virtual Machine eXtensions <br/>
  > **SMX**          Safer Mode Extensions (LaGrande) <br/>
  > **EST**          Enhanced SpeedStep <br/>
  > **TM2**          Thermal Monitor 2 <br/>
  > **SSSE3**        Supplemental SSE3 instructions <br/>
  > **CNXT_ID**      L1 Context ID <br/>
  > **SDBG**         Silicon Debug interface <br/>
  > **FMA**          Fused multiply-add (FMA3) <br/>
  > **CX16**         CMPXCHG16B instruction <br/>
  > **XTPR**         Can disable sending task priority messages <br/>
  > **PDCM**         Perfmon & debug capability <br/>
  > **PCID**         Process context identifiers (CR4 bit 17) <br/>
  > **DCA**          Direct cache access for DMA writes[10][11] <br/>
  > **SSE4_1**       SSE4.1 instructions <br/>
  > **SSE4_2**       SSE4.2 instructions <br/>
  > **X2APIC**       x2APIC support <br/>
  > **MOVBE**        MOVBE instruction (big-endian) <br/>
  > **POPCNT**       POPCNT instruction <br/>
  > **TSC_DEADLINE** line APIC supports one-shot operation using a TSC deadline value <br/>
  > **AES**          AES instruction set <br/>
  > **XSAVE**        XSAVE, XRESTOR, XSETBV, XGETBV <BR/>
  > **OSXSAVE**      XSAVE enabled by OS <br/>
  > **AVX**          Advanced Vector Extensions <br/>
  > **F16C**         F16C (half-precision) FP support <br/>
  > **RDRND**        RDRAND (on-chip random number generator) support <br/>
  > **HYPERVISOR**   Running on a hypervisor (always 0 on a real CPU, but also with some hypervisors) <br/>
  > **FPU**          Onboard x87 FPU <br/>
  > **VME**          Virtual 8086 mode extensions (such as VIF, VIP, PIV) <br/>
  > **DE**           Debugging extensions (CR4 bit 3) <br/>
  > **PSE**          Page Size Extension <br/>
  > **TSC**          Time Stamp Counter <br/>
  > **MSR**          Model-specific registers <br/>
  > **PAE**          Physical Address Extension <br/>
  > **MCE**          Machine Check Exception <br/>
  > **CX8**          CMPXCHG8 (compare-and-swap) instruction <br/>
  > **APIC**         Onboard Advanced Programmable Interrupt Controller <br/>
  > **SEP**          SYSENTER and SYSEXIT instructions <br/>
  > **MTRR**         Memory Type Range Registers <br/>
  > **PGE**          Page Global Enable bit in CR4 <br/>
  > **MCA**          Machine check architecture <br/>
  > **CMOV**         Conditional move and FCMOV instructions <br/>
  > **PAT**          Page Attribute Table <br/>
  > **PSE_36**       36-bit page size extension <br/>
  > **PSN**          Processor Serial Number <br/>
  > **CLFSH**        CLFLUSH instruction (SSE2) <br/>
  > **DS**           Debug store: save trace of executed jumps <br/>
  > **ACPI**         Onboard thermal control MSRs for ACPI <br/>
  > **MMX**          MMX instructions <br/>
  > **FXSR**         FXSAVE, FXRESTOR instructions, CR4 bit 9 <br/>
  > **SSE**          SSE instructions (a.k.a. Katmai New Instructions) <br/>
  > **SSE2**         SSE2 instructions <br/>
  > **SS**           CPU cache supports self-snoop <br/>
  > **HTT**          Hyper-threading <br/>
  > **TM**           Thermal monitor automatically limits temperature <br/>
  > **IA64**         IA64 processor emulating x86 <br/>
  > **PBE**          Pending Break Enable (PBE# pin) wakeup support <br/>

Usage example:

```go
if EnabledAVX && HasFeature(AVX) {
    fmt.Printf("We can use AVX\n")
}
```

* **func HasExtendedFeature(feature uint64) bool** to check for the following features:
  > **FSGSBASE**                Access to base of %fs and %gs<br/>
  > **IA32_TSC_ADJUST**         IA32_TSC_ADJUST MSR is supported if 1<br/>
  > **BMI1**                    Bit Manipulation Instruction Set 1<br/>
  > **HLE**                     Transactional Synchronization Extensions<br/>
  > **AVX2**                    Advanced Vector Extensions 2<br/>
  > **SMEP**                    Supervisor-Mode Execution Prevention<br/>
  > **BMI2**                    Bit Manipulation Instruction Set 2<br/>
  > **ERMS**                    Enhanced REP MOVSB/STOSB<br/>
  > **INVPCID**                 INVPCID instruction<br/>
  > **RTM**                     Transactional Synchronization Extensions<br/>
  > **PQM**                     Supports Platform Quality of Service Monitoring (PQM) capability if 1<br/>
  > **DFPUCDS**                 Deprecates FPU CS and FPU DS values if 1<br/>
  > **MPX**                     Intel MPX (Memory Protection Extensions)<br/>
  > **PQE**                     Supports Platform Quality of Service Enforcement (PQE) capability if 1<br/>
  > **AVX512F**                 AVX-512 Foundation<br/>
  > **AVX512DQ**                AVX-512 Doubleword and Quadword Instructions<br/>
  > **RDSEED**                  RDSEED instruction<br/>
  > **ADX**                     Intel ADX (Multi-Precision Add-Carry Instruction Extensions)<br/>
  > **SMAP**                    Supervisor Mode Access Prevention<br/>
  > **AVX512IFMA**              AVX-512 Integer Fused Multiply-Add Instructions<br/>
  > **PCOMMIT**                 PCOMMIT instruction<br/>
  > **CLFLUSHOPT**              CLFLUSHOPT instruction<br/>
  > **CLWB**                    CLWB instruction<br/>
  > **INTEL_PROCESSOR_TRACE**   Intel Processor Trace<br/>
  > **AVX512PF**                AVX-512 Prefetch Instructions<br/>
  > **AVX512ER**                AVX-512 Exponential and Reciprocal Instructions<br/>
  > **AVX512CD**                AVX-512 Conflict Detection Instructions<br/>
  > **SHA**                     Intel SHA extensions<br/>
  > **AVX512BW**                AVX-512 Byte and Word Instructions<br/>
  > **AVX512VL**                AVX-512 Vector Length Extensions<br/>
  > **PREFETCHWT1**             PREFETCHWT1 instruction<br/>
  > **AVX512VBMI**              AVX-512 Vector Bit Manipulation Instructions<br/>

* **func HasExtraFeature(feature uint64) bool**
  > **LAHF_LM**           LahfSahf LAHF and SAHF instruction support in 64-bit mod<br/>
  > **CMP_LEGACY**        CmpLegacy Core multi-processing legacy mode.<br/>
  > **SVM**               SVM Secure virtual machine.<br/>
  > **EXTAPIC**           ExtApicSpace Extended APIC space.<br/>
  > **CR8_LEGACY**        AltMovCr8 LOCK MOV CR0 means MOV CR8.<br/>
  > **ABM**               ABM Advanced bit manipulation. LZCNT instruction support.<br/>
  > **SSE4A**             SSE4A EXTRQ, INSERTQ, MOVNTSS, and MOVNTSD instruction support.<br/>
  > **MISALIGNSSE**       Misaligned SSE mode.<br/>
  > **PREFETCHW**         PREFETCH and PREFETCHW instruction support.<br/>
  > **OSVW**              OSVW OS visible workaround. Indicates OS-visible workaround support.<br/>
  > **IBS**               IBS Instruction based sampling.<br/>
  > **XOP**               XOP Extended operation support.<br/>
  > **SKINIT**            SKINIT SKINIT and STGI are supported.<br/>
  > **WDT**               WDT Watchdog timer support.<br/>
  > **LWP**               LWP Lightweight profiling support.<br/>
  > **FMA4**              FMA4 Four-operand FMA instruction support.<br/>
  > **TCE**               Translation Cache Extension<br/>
  > **NODEID_MSR**        NodeID MSR<br/>
  > **TBM**               TBM Trailing bit manipulation instruction support.<br/>
  > **TOPOEXT**           TopologyExtensio Topology extensions support.<br/>
  > **PERFCTR_CORE**      PerfCtrExtCore Processor performance counter extensions support.<br/>
  > **PERFCTR_NB**        PerfCtrExtNB NB performance counter extensions support.<br/>
  > **SPM**               StreamPerfMon Streaming performance monitor architecture.<br/>
  > **DBX**               DataBreakpointEx Data access breakpoint extension.<br/>
  > **PERFTSC**           PerfTsc<br/>
  > **PCX_L2I**           L2I perf counter extensions<br/>
  > **FPU_2**             Onboard x87 FPU<br/>
  > **VME_2**             Virtual mode extensions (VIF)<br/>
  > **DE_2**              Debugging extensions (CR4 bit 3)<br/>
  > **PSE_2**             Page Size Extension<br/>
  > **TSC_2**             Time Stamp Counter<br/>
  > **MSR_2**             Model-specific register<br/>
  > **PAE_2**             Physical Address Extension<br/>
  > **MCE_2**             Machine Check Exception<br/>
  > **CX8_2**             CMPXCHG8 (compare-and-swap) instruction<br/>
  > **APIC_2**            Onboard Advanced Programmable Interrupt Controller<br/>
  > **SYSCALL**           SYSCALL and SYSRET instructions<br/>
  > **MTRR_2**            Memory Type Range Registers<br/>
  > **PGE_2**             Page Global Enable bit in CR4<br/>
  > **MCA_2**             Machine check architecture<br/>
  > **CMOV_2**            Conditional move and FCMOV instructions<br/>
  > **PAT_2**             Page Attribute Table<br/>
  > **PSE36**             36-bit page size extension<br/>
  > **MP**                Multiprocessor Capable<br/>
  > **NX**                NX bit<br/>
  > **MMXEXT**            Extended MMX<br/>
  > **MMX_2**             MMX instructions<br/>
  > **FXSR_2**            FXSAVE, FXRSTOR instructions<br/>
  > **FXSR_OPT**          FXSAVE/FXRSTOR optimizations<br/>
  > **PDPE1GB**           Gibibyte pages<br/>
  > **RDTSCP**            RDTSCP instruction<br/>
  > **LM**                Long mode<br/>
  > **_3DNOWEXT**         Extended 3DNow!<br/>
  > **_3DNOW**            3DNow!<br/>
