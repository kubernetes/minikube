/*
 * This file is part of the libvirt-go project
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * Copyright (c) 2013 Alex Zorin
 * Copyright (C) 2016 Red Hat, Inc.
 *
 */

package libvirt

import (
	"fmt"
	"unsafe"
)

/*
#cgo pkg-config: libvirt
#include <libvirt/libvirt.h>
#include "domain_events_cfuncs.h"
#include "domain_compat.h"
*/
import "C"

type DomainEventGenericCallback func(c *Connect, d *Domain)

type DomainEventLifecycle struct {
	Event DomainEventType
	// TODO: we can make Detail typesafe somehow ?
	Detail int
}

type DomainEventLifecycleCallback func(c *Connect, d *Domain, event *DomainEventLifecycle)

type DomainEventRTCChange struct {
	Utcoffset int64
}

type DomainEventRTCChangeCallback func(c *Connect, d *Domain, event *DomainEventRTCChange)

type DomainEventWatchdog struct {
	Action DomainEventWatchdogAction
}

type DomainEventWatchdogCallback func(c *Connect, d *Domain, event *DomainEventWatchdog)

type DomainEventIOError struct {
	SrcPath  string
	DevAlias string
	Action   DomainEventIOErrorAction
}

type DomainEventIOErrorCallback func(c *Connect, d *Domain, event *DomainEventIOError)

type DomainEventGraphicsAddress struct {
	Family  DomainEventGraphicsAddressType
	Node    string
	Service string
}

type DomainEventGraphicsSubjectIdentity struct {
	Type string
	Name string
}

type DomainEventGraphics struct {
	Phase      DomainEventGraphicsPhase
	Local      DomainEventGraphicsAddress
	Remote     DomainEventGraphicsAddress
	AuthScheme string
	Subject    []DomainEventGraphicsSubjectIdentity
}

type DomainEventGraphicsCallback func(c *Connect, d *Domain, event *DomainEventGraphics)

type DomainEventIOErrorReason struct {
	SrcPath  string
	DevAlias string
	Action   DomainEventIOErrorAction
	Reason   string
}

type DomainEventIOErrorReasonCallback func(c *Connect, d *Domain, event *DomainEventIOErrorReason)

type DomainEventBlockJob struct {
	Disk   string
	Type   DomainBlockJobType
	Status ConnectDomainEventBlockJobStatus
}

type DomainEventBlockJobCallback func(c *Connect, d *Domain, event *DomainEventBlockJob)

type DomainEventDiskChange struct {
	OldSrcPath string
	NewSrcPath string
	DevAlias   string
	Reason     ConnectDomainEventDiskChangeReason
}

type DomainEventDiskChangeCallback func(c *Connect, d *Domain, event *DomainEventDiskChange)

type DomainEventTrayChange struct {
	DevAlias string
	Reason   ConnectDomainEventTrayChangeReason
}

type DomainEventTrayChangeCallback func(c *Connect, d *Domain, event *DomainEventTrayChange)

type DomainEventPMSuspend struct {
	Reason int
}

type DomainEventPMSuspendCallback func(c *Connect, d *Domain, event *DomainEventPMSuspend)

type DomainEventPMWakeup struct {
	Reason int
}

type DomainEventPMWakeupCallback func(c *Connect, d *Domain, event *DomainEventPMWakeup)

type DomainEventPMSuspendDisk struct {
	Reason int
}

type DomainEventPMSuspendDiskCallback func(c *Connect, d *Domain, event *DomainEventPMSuspendDisk)

type DomainEventBalloonChange struct {
	Actual uint64
}

type DomainEventBalloonChangeCallback func(c *Connect, d *Domain, event *DomainEventBalloonChange)

type DomainEventDeviceRemoved struct {
	DevAlias string
}

type DomainEventDeviceRemovedCallback func(c *Connect, d *Domain, event *DomainEventDeviceRemoved)

type DomainEventTunableCpuPin struct {
	VcpuPinSet     bool
	VcpuPin        [][]bool
	EmulatorPinSet bool
	EmulatorPin    []bool
	IOThreadPinSet bool
	IOThreadPin    [][]bool
}

type DomainEventTunable struct {
	CpuSched      *DomainSchedulerParameters
	CpuPin        *DomainEventTunableCpuPin
	BlkdevDiskSet bool
	BlkdevDisk    string
	BlkdevTune    *DomainBlockIoTuneParameters
}

type DomainEventTunableCallback func(c *Connect, d *Domain, event *DomainEventTunable)

type DomainEventAgentLifecycle struct {
	State  ConnectDomainEventAgentLifecycleState
	Reason ConnectDomainEventAgentLifecycleReason
}

type DomainEventAgentLifecycleCallback func(c *Connect, d *Domain, event *DomainEventAgentLifecycle)

type DomainEventDeviceAdded struct {
	DevAlias string
}

type DomainEventDeviceAddedCallback func(c *Connect, d *Domain, event *DomainEventDeviceAdded)

type DomainEventMigrationIteration struct {
	Iteration int
}

type DomainEventMigrationIterationCallback func(c *Connect, d *Domain, event *DomainEventMigrationIteration)

type DomainEventJobCompleted struct {
	Info DomainJobInfo
}

type DomainEventJobCompletedCallback func(c *Connect, d *Domain, event *DomainEventJobCompleted)

type DomainEventDeviceRemovalFailed struct {
	DevAlias string
}

type DomainEventDeviceRemovalFailedCallback func(c *Connect, d *Domain, event *DomainEventDeviceRemovalFailed)

type DomainEventMetadataChange struct {
	Type  int
	NSURI string
}

type DomainEventMetadataChangeCallback func(c *Connect, d *Domain, event *DomainEventMetadataChange)

type DomainEventBlockThreshold struct {
	Dev       string
	Path      string
	Threshold uint64
	Excess    uint64
}

type DomainEventBlockThresholdCallback func(c *Connect, d *Domain, event *DomainEventBlockThreshold)

//export domainEventLifecycleCallback
func domainEventLifecycleCallback(c C.virConnectPtr, d C.virDomainPtr,
	event int, detail int,
	goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventLifecycle{
		Event:  DomainEventType(event),
		Detail: detail,
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventLifecycleCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)
}

//export domainEventGenericCallback
func domainEventGenericCallback(c C.virConnectPtr, d C.virDomainPtr,
	goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventGenericCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain)
}

//export domainEventRTCChangeCallback
func domainEventRTCChangeCallback(c C.virConnectPtr, d C.virDomainPtr,
	utcoffset int64, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventRTCChange{
		Utcoffset: utcoffset,
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventRTCChangeCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventWatchdogCallback
func domainEventWatchdogCallback(c C.virConnectPtr, d C.virDomainPtr,
	action int, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventWatchdog{
		Action: DomainEventWatchdogAction(action),
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventWatchdogCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventIOErrorCallback
func domainEventIOErrorCallback(c C.virConnectPtr, d C.virDomainPtr,
	srcPath *C.char, devAlias *C.char, action int, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventIOError{
		SrcPath:  C.GoString(srcPath),
		DevAlias: C.GoString(devAlias),
		Action:   DomainEventIOErrorAction(action),
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventIOErrorCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventGraphicsCallback
func domainEventGraphicsCallback(c C.virConnectPtr, d C.virDomainPtr,
	phase int,
	local C.virDomainEventGraphicsAddressPtr,
	remote C.virDomainEventGraphicsAddressPtr,
	authScheme *C.char,
	subject C.virDomainEventGraphicsSubjectPtr,
	goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	subjectGo := make([]DomainEventGraphicsSubjectIdentity, subject.nidentity)
	nidentities := int(subject.nidentity)
	identities := (*[1 << 30]C.virDomainEventGraphicsSubjectIdentity)(unsafe.Pointer(&subject.identities))[:nidentities:nidentities]
	for _, identity := range identities {
		subjectGo = append(subjectGo,
			DomainEventGraphicsSubjectIdentity{
				Type: C.GoString(identity._type),
				Name: C.GoString(identity.name),
			},
		)
	}

	eventDetails := &DomainEventGraphics{
		Phase: DomainEventGraphicsPhase(phase),
		Local: DomainEventGraphicsAddress{
			Family:  DomainEventGraphicsAddressType(local.family),
			Node:    C.GoString(local.node),
			Service: C.GoString(local.service),
		},
		Remote: DomainEventGraphicsAddress{
			Family:  DomainEventGraphicsAddressType(remote.family),
			Node:    C.GoString(remote.node),
			Service: C.GoString(remote.service),
		},
		AuthScheme: C.GoString(authScheme),
		Subject:    subjectGo,
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventGraphicsCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventIOErrorReasonCallback
func domainEventIOErrorReasonCallback(c C.virConnectPtr, d C.virDomainPtr,
	srcPath *C.char, devAlias *C.char, action int, reason *C.char,
	goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventIOErrorReason{
		SrcPath:  C.GoString(srcPath),
		DevAlias: C.GoString(devAlias),
		Action:   DomainEventIOErrorAction(action),
		Reason:   C.GoString(reason),
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventIOErrorReasonCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventBlockJobCallback
func domainEventBlockJobCallback(c C.virConnectPtr, d C.virDomainPtr,
	disk *C.char, _type int, status int, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventBlockJob{
		Disk:   C.GoString(disk),
		Type:   DomainBlockJobType(_type),
		Status: ConnectDomainEventBlockJobStatus(status),
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventBlockJobCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventDiskChangeCallback
func domainEventDiskChangeCallback(c C.virConnectPtr, d C.virDomainPtr,
	oldSrcPath *C.char, newSrcPath *C.char, devAlias *C.char,
	reason int, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventDiskChange{
		OldSrcPath: C.GoString(oldSrcPath),
		NewSrcPath: C.GoString(newSrcPath),
		DevAlias:   C.GoString(devAlias),
		Reason:     ConnectDomainEventDiskChangeReason(reason),
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventDiskChangeCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventTrayChangeCallback
func domainEventTrayChangeCallback(c C.virConnectPtr, d C.virDomainPtr,
	devAlias *C.char, reason int, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventTrayChange{
		DevAlias: C.GoString(devAlias),
		Reason:   ConnectDomainEventTrayChangeReason(reason),
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventTrayChangeCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventPMSuspendCallback
func domainEventPMSuspendCallback(c C.virConnectPtr, d C.virDomainPtr,
	reason int, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventPMSuspend{
		Reason: reason,
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventPMSuspendCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventPMWakeupCallback
func domainEventPMWakeupCallback(c C.virConnectPtr, d C.virDomainPtr,
	reason int, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventPMWakeup{
		Reason: reason,
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventPMWakeupCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventPMSuspendDiskCallback
func domainEventPMSuspendDiskCallback(c C.virConnectPtr, d C.virDomainPtr,
	reason int, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventPMSuspendDisk{
		Reason: reason,
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventPMSuspendDiskCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventBalloonChangeCallback
func domainEventBalloonChangeCallback(c C.virConnectPtr, d C.virDomainPtr,
	actual uint64, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventBalloonChange{
		Actual: actual,
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventBalloonChangeCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventDeviceRemovedCallback
func domainEventDeviceRemovedCallback(c C.virConnectPtr, d C.virDomainPtr,
	devAlias *C.char, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventDeviceRemoved{
		DevAlias: C.GoString(devAlias),
	}
	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventDeviceRemovedCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventMetadataChangeCallback
func domainEventMetadataChangeCallback(c C.virConnectPtr, d C.virDomainPtr,
	mtype int, nsuri *C.char, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventMetadataChange{
		Type:  (int)(mtype),
		NSURI: C.GoString(nsuri),
	}
	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventMetadataChangeCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

func getDomainTuneSchedulerParametersFieldInfo(params *DomainSchedulerParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_TUNABLE_CPU_CPU_SHARES: typedParamsFieldInfo{
			set: &params.CpuSharesSet,
			ul:  &params.CpuShares,
		},
		C.VIR_DOMAIN_TUNABLE_CPU_GLOBAL_PERIOD: typedParamsFieldInfo{
			set: &params.GlobalPeriodSet,
			ul:  &params.GlobalPeriod,
		},
		C.VIR_DOMAIN_TUNABLE_CPU_GLOBAL_QUOTA: typedParamsFieldInfo{
			set: &params.GlobalQuotaSet,
			ul:  &params.GlobalQuota,
		},
		C.VIR_DOMAIN_TUNABLE_CPU_EMULATOR_PERIOD: typedParamsFieldInfo{
			set: &params.EmulatorPeriodSet,
			ul:  &params.EmulatorPeriod,
		},
		C.VIR_DOMAIN_TUNABLE_CPU_EMULATOR_QUOTA: typedParamsFieldInfo{
			set: &params.EmulatorQuotaSet,
			ul:  &params.EmulatorQuota,
		},
		C.VIR_DOMAIN_TUNABLE_CPU_VCPU_PERIOD: typedParamsFieldInfo{
			set: &params.VcpuPeriodSet,
			ul:  &params.VcpuPeriod,
		},
		C.VIR_DOMAIN_TUNABLE_CPU_VCPU_QUOTA: typedParamsFieldInfo{
			set: &params.VcpuQuotaSet,
			ul:  &params.VcpuQuota,
		},
		C.VIR_DOMAIN_TUNABLE_CPU_IOTHREAD_PERIOD: typedParamsFieldInfo{
			set: &params.IothreadPeriodSet,
			ul:  &params.IothreadPeriod,
		},
		C.VIR_DOMAIN_TUNABLE_CPU_IOTHREAD_QUOTA: typedParamsFieldInfo{
			set: &params.IothreadQuotaSet,
			ul:  &params.IothreadQuota,
		},
	}
}

func getTuneBlockIoTuneParametersFieldInfo(params *DomainBlockIoTuneParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_BYTES_SEC: typedParamsFieldInfo{
			set: &params.TotalBytesSecSet,
			ul:  &params.TotalBytesSec,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_READ_BYTES_SEC: typedParamsFieldInfo{
			set: &params.ReadBytesSecSet,
			ul:  &params.ReadBytesSec,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_BYTES_SEC: typedParamsFieldInfo{
			set: &params.WriteBytesSecSet,
			ul:  &params.WriteBytesSec,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_IOPS_SEC: typedParamsFieldInfo{
			set: &params.TotalIopsSecSet,
			ul:  &params.TotalIopsSec,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_READ_IOPS_SEC: typedParamsFieldInfo{
			set: &params.ReadIopsSecSet,
			ul:  &params.ReadIopsSec,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_IOPS_SEC: typedParamsFieldInfo{
			set: &params.WriteIopsSecSet,
			ul:  &params.WriteIopsSec,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_BYTES_SEC_MAX: typedParamsFieldInfo{
			set: &params.TotalBytesSecMaxSet,
			ul:  &params.TotalBytesSecMax,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_READ_BYTES_SEC_MAX: typedParamsFieldInfo{
			set: &params.ReadBytesSecMaxSet,
			ul:  &params.ReadBytesSecMax,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_BYTES_SEC_MAX: typedParamsFieldInfo{
			set: &params.WriteBytesSecMaxSet,
			ul:  &params.WriteBytesSecMax,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_IOPS_SEC_MAX: typedParamsFieldInfo{
			set: &params.TotalIopsSecMaxSet,
			ul:  &params.TotalIopsSecMax,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_READ_IOPS_SEC_MAX: typedParamsFieldInfo{
			set: &params.ReadIopsSecMaxSet,
			ul:  &params.ReadIopsSecMax,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_IOPS_SEC_MAX: typedParamsFieldInfo{
			set: &params.WriteIopsSecMaxSet,
			ul:  &params.WriteIopsSecMax,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_BYTES_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.TotalBytesSecMaxLengthSet,
			ul:  &params.TotalBytesSecMaxLength,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_READ_BYTES_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.ReadBytesSecMaxLengthSet,
			ul:  &params.ReadBytesSecMaxLength,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_BYTES_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.WriteBytesSecMaxLengthSet,
			ul:  &params.WriteBytesSecMaxLength,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_IOPS_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.TotalIopsSecMaxLengthSet,
			ul:  &params.TotalIopsSecMaxLength,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_READ_IOPS_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.ReadIopsSecMaxLengthSet,
			ul:  &params.ReadIopsSecMaxLength,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_IOPS_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.WriteIopsSecMaxLengthSet,
			ul:  &params.WriteIopsSecMaxLength,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_SIZE_IOPS_SEC: typedParamsFieldInfo{
			set: &params.SizeIopsSecSet,
			ul:  &params.SizeIopsSec,
		},
		C.VIR_DOMAIN_TUNABLE_BLKDEV_GROUP_NAME: typedParamsFieldInfo{
			set: &params.GroupNameSet,
			s:   &params.GroupName,
		},
	}
}

type domainEventTunablePinTemp struct {
	VcpuPinSet     bool
	VcpuPin        []string
	EmulatorPinSet bool
	EmulatorPin    string
	IOThreadPinSet bool
	IOThreadPin    []string
}

func getDomainPinTempFieldInfo(numvcpu int, numiothread int, params *domainEventTunablePinTemp) map[string]typedParamsFieldInfo {
	ret := map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_TUNABLE_CPU_EMULATORPIN: typedParamsFieldInfo{
			set: &params.EmulatorPinSet,
			s:   &params.EmulatorPin,
		},
	}
	for i := 0; i < numvcpu; i++ {
		ret[fmt.Sprintf("cputune.vcpupin%d", i)] = typedParamsFieldInfo{
			s: &params.VcpuPin[i],
		}
	}
	for i := 0; i < numiothread; i++ {
		ret[fmt.Sprintf("cputune.iothreadpin%u", i)] = typedParamsFieldInfo{
			s: &params.IOThreadPin[i],
		}
	}

	return ret
}

func countPinInfo(cparams C.virTypedParameterPtr, nparams C.int) (int, int) {
	maxvcpus := 0
	maxiothreads := 0
	for i := 0; i < int(nparams); i++ {
		var cparam *C.virTypedParameter
		cparam = (*C.virTypedParameter)(unsafe.Pointer(uintptr(unsafe.Pointer(cparams)) + unsafe.Sizeof(*cparam)*uintptr(i)))
		name := C.GoString((*C.char)(unsafe.Pointer(&cparam.field)))

		var vcpu int
		_, err := fmt.Scanf(name, "cputune.vcpupin%d", &vcpu)
		if err == nil {
			if vcpu > maxvcpus {
				maxvcpus = vcpu
			}
		}

		var iothread int
		_, err = fmt.Scanf(name, "cputune.iothreadpin%d", &iothread)
		if err == nil {
			if iothread > maxiothreads {
				maxiothreads = iothread
			}
		}
	}

	return maxvcpus + 1, maxiothreads + 1
}

func domainEventTunableGetPin(params C.virTypedParameterPtr, nparams C.int) *DomainEventTunableCpuPin {
	var pin domainEventTunablePinTemp
	numvcpus, numiothreads := countPinInfo(params, nparams)
	pinInfo := getDomainPinTempFieldInfo(numvcpus, numiothreads, &pin)

	num, err := typedParamsUnpackLen(params, int(nparams), pinInfo)
	if num == 0 || err != nil {
		return nil
	}

	info := &DomainEventTunableCpuPin{}

	if pin.VcpuPinSet {
		info.VcpuPinSet = true
		info.VcpuPin = make([][]bool, len(pin.VcpuPin))

		for i := 0; i < len(pin.VcpuPin); i++ {
			bits, err := parseCPUString(pin.VcpuPin[i])
			if err == nil {
				info.VcpuPin[i] = bits
			}
		}
	}

	if pin.EmulatorPinSet {
		bits, err := parseCPUString(pin.EmulatorPin)
		if err == nil {
			info.EmulatorPinSet = true
			info.EmulatorPin = bits
		}
	}

	if pin.IOThreadPinSet {
		info.IOThreadPinSet = true
		info.IOThreadPin = make([][]bool, len(pin.IOThreadPin))

		for i := 0; i < len(pin.IOThreadPin); i++ {
			bits, err := parseCPUString(pin.IOThreadPin[i])
			if err == nil {
				info.IOThreadPin[i] = bits
			}
		}
	}

	return info
}

//export domainEventTunableCallback
func domainEventTunableCallback(c C.virConnectPtr, d C.virDomainPtr, params C.virTypedParameterPtr, nparams C.int, goCallbackId int) {
	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventTunable{}

	pin := domainEventTunableGetPin(params, nparams)
	if pin != nil {
		eventDetails.CpuPin = pin
	}

	var sched DomainSchedulerParameters
	schedInfo := getDomainTuneSchedulerParametersFieldInfo(&sched)

	num, _ := typedParamsUnpackLen(params, int(nparams), schedInfo)
	if num > 0 {
		eventDetails.CpuSched = &sched
	}

	blknameInfo := map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_TUNABLE_BLKDEV_DISK: typedParamsFieldInfo{
			set: &eventDetails.BlkdevDiskSet,
			s:   &eventDetails.BlkdevDisk,
		},
	}
	typedParamsUnpackLen(params, int(nparams), blknameInfo)

	var blktune DomainBlockIoTuneParameters
	blktuneInfo := getTuneBlockIoTuneParametersFieldInfo(&blktune)

	num, _ = typedParamsUnpackLen(params, int(nparams), blktuneInfo)
	if num > 0 {
		eventDetails.BlkdevTune = &blktune
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventTunableCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventAgentLifecycleCallback
func domainEventAgentLifecycleCallback(c C.virConnectPtr, d C.virDomainPtr, state C.int, reason C.int, goCallbackId int) {
	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventAgentLifecycle{
		State:  ConnectDomainEventAgentLifecycleState(state),
		Reason: ConnectDomainEventAgentLifecycleReason(reason),
	}
	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventAgentLifecycleCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventDeviceAddedCallback
func domainEventDeviceAddedCallback(c C.virConnectPtr, d C.virDomainPtr, devalias *C.char, goCallbackId int) {
	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventDeviceAdded{
		DevAlias: C.GoString(devalias),
	}
	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventDeviceAddedCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventMigrationIterationCallback
func domainEventMigrationIterationCallback(c C.virConnectPtr, d C.virDomainPtr, iteration C.int, goCallbackId int) {
	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventMigrationIteration{
		Iteration: int(iteration),
	}
	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventMigrationIterationCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventJobCompletedCallback
func domainEventJobCompletedCallback(c C.virConnectPtr, d C.virDomainPtr, params C.virTypedParameterPtr, nparams C.int, goCallbackId int) {
	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventJobCompleted{}
	info := getDomainJobInfoFieldInfo(&eventDetails.Info)

	typedParamsUnpackLen(params, int(nparams), info)

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventJobCompletedCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventDeviceRemovalFailedCallback
func domainEventDeviceRemovalFailedCallback(c C.virConnectPtr, d C.virDomainPtr, devalias *C.char, goCallbackId int) {
	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventDeviceRemovalFailed{
		DevAlias: C.GoString(devalias),
	}
	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventDeviceRemovalFailedCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

//export domainEventBlockThresholdCallback
func domainEventBlockThresholdCallback(c C.virConnectPtr, d C.virDomainPtr, dev *C.char, path *C.char, threshold C.ulonglong, excess C.ulonglong, goCallbackId int) {
	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainEventBlockThreshold{
		Dev:       C.GoString(dev),
		Path:      C.GoString(path),
		Threshold: uint64(threshold),
		Excess:    uint64(excess),
	}
	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainEventBlockThresholdCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

func (c *Connect) DomainEventLifecycleRegister(dom *Domain, callback DomainEventLifecycleCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventLifecycleCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_LIFECYCLE,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventRebootRegister(dom *Domain, callback DomainEventGenericCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventGenericCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_REBOOT,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventRTCChangeRegister(dom *Domain, callback DomainEventRTCChangeCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventRTCChangeCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_RTC_CHANGE,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventWatchdogRegister(dom *Domain, callback DomainEventWatchdogCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventWatchdogCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_WATCHDOG,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventIOErrorRegister(dom *Domain, callback DomainEventIOErrorCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventIOErrorCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_IO_ERROR,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventGraphicsRegister(dom *Domain, callback DomainEventGraphicsCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventGraphicsCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_GRAPHICS,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventIOErrorReasonRegister(dom *Domain, callback DomainEventIOErrorReasonCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventIOErrorReasonCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_IO_ERROR_REASON,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventControlErrorRegister(dom *Domain, callback DomainEventGenericCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventGenericCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_CONTROL_ERROR,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventBlockJobRegister(dom *Domain, callback DomainEventBlockJobCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventBlockJobCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_BLOCK_JOB,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventDiskChangeRegister(dom *Domain, callback DomainEventDiskChangeCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventDiskChangeCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_DISK_CHANGE,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventTrayChangeRegister(dom *Domain, callback DomainEventTrayChangeCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventTrayChangeCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_TRAY_CHANGE,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventPMWakeupRegister(dom *Domain, callback DomainEventPMWakeupCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventPMWakeupCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_PMWAKEUP,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventPMSuspendRegister(dom *Domain, callback DomainEventPMSuspendCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventPMSuspendCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_PMSUSPEND,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventBalloonChangeRegister(dom *Domain, callback DomainEventBalloonChangeCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventBalloonChangeCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_BALLOON_CHANGE,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventPMSuspendDiskRegister(dom *Domain, callback DomainEventPMSuspendDiskCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventPMSuspendDiskCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_PMSUSPEND_DISK,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventDeviceRemovedRegister(dom *Domain, callback DomainEventDeviceRemovedCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventDeviceRemovedCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_DEVICE_REMOVED,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventBlockJob2Register(dom *Domain, callback DomainEventBlockJobCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventBlockJobCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_BLOCK_JOB_2,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventTunableRegister(dom *Domain, callback DomainEventTunableCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventTunableCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_TUNABLE,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventAgentLifecycleRegister(dom *Domain, callback DomainEventAgentLifecycleCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventAgentLifecycleCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_AGENT_LIFECYCLE,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventDeviceAddedRegister(dom *Domain, callback DomainEventDeviceAddedCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventDeviceAddedCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_DEVICE_ADDED,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventMigrationIterationRegister(dom *Domain, callback DomainEventMigrationIterationCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventMigrationIterationCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_MIGRATION_ITERATION,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventJobCompletedRegister(dom *Domain, callback DomainEventJobCompletedCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventJobCompletedCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_JOB_COMPLETED,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventDeviceRemovalFailedRegister(dom *Domain, callback DomainEventDeviceRemovalFailedCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventDeviceRemovalFailedCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_DEVICE_REMOVAL_FAILED,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventMetadataChangeRegister(dom *Domain, callback DomainEventMetadataChangeCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventMetadataChangeCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_METADATA_CHANGE,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventBlockThresholdRegister(dom *Domain, callback DomainEventBlockThresholdCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainEventBlockThresholdCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, cdom,
		C.VIR_DOMAIN_EVENT_ID_BLOCK_THRESHOLD,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainEventDeregister(callbackId int) error {
	// Deregister the callback
	if i := int(C.virConnectDomainEventDeregisterAny(c.ptr, C.int(callbackId))); i != 0 {
		return GetLastError()
	}
	return nil
}

func (e DomainEventLifecycle) String() string {
	var detail, event string
	switch e.Event {
	case DOMAIN_EVENT_DEFINED:
		event = "defined"
		switch DomainEventDefinedDetailType(e.Detail) {
		case DOMAIN_EVENT_DEFINED_ADDED:
			detail = "added"
		case DOMAIN_EVENT_DEFINED_UPDATED:
			detail = "updated"
		default:
			detail = "unknown"
		}

	case DOMAIN_EVENT_UNDEFINED:
		event = "undefined"
		switch DomainEventUndefinedDetailType(e.Detail) {
		case DOMAIN_EVENT_UNDEFINED_REMOVED:
			detail = "removed"
		default:
			detail = "unknown"
		}

	case DOMAIN_EVENT_STARTED:
		event = "started"
		switch DomainEventStartedDetailType(e.Detail) {
		case DOMAIN_EVENT_STARTED_BOOTED:
			detail = "booted"
		case DOMAIN_EVENT_STARTED_MIGRATED:
			detail = "migrated"
		case DOMAIN_EVENT_STARTED_RESTORED:
			detail = "restored"
		case DOMAIN_EVENT_STARTED_FROM_SNAPSHOT:
			detail = "snapshot"
		default:
			detail = "unknown"
		}

	case DOMAIN_EVENT_SUSPENDED:
		event = "suspended"
		switch DomainEventSuspendedDetailType(e.Detail) {
		case DOMAIN_EVENT_SUSPENDED_PAUSED:
			detail = "paused"
		case DOMAIN_EVENT_SUSPENDED_MIGRATED:
			detail = "migrated"
		case DOMAIN_EVENT_SUSPENDED_IOERROR:
			detail = "I/O error"
		case DOMAIN_EVENT_SUSPENDED_WATCHDOG:
			detail = "watchdog"
		case DOMAIN_EVENT_SUSPENDED_RESTORED:
			detail = "restored"
		case DOMAIN_EVENT_SUSPENDED_FROM_SNAPSHOT:
			detail = "snapshot"
		default:
			detail = "unknown"
		}

	case DOMAIN_EVENT_RESUMED:
		event = "resumed"
		switch DomainEventResumedDetailType(e.Detail) {
		case DOMAIN_EVENT_RESUMED_UNPAUSED:
			detail = "unpaused"
		case DOMAIN_EVENT_RESUMED_MIGRATED:
			detail = "migrated"
		case DOMAIN_EVENT_RESUMED_FROM_SNAPSHOT:
			detail = "snapshot"
		default:
			detail = "unknown"
		}

	case DOMAIN_EVENT_STOPPED:
		event = "stopped"
		switch DomainEventStoppedDetailType(e.Detail) {
		case DOMAIN_EVENT_STOPPED_SHUTDOWN:
			detail = "shutdown"
		case DOMAIN_EVENT_STOPPED_DESTROYED:
			detail = "destroyed"
		case DOMAIN_EVENT_STOPPED_CRASHED:
			detail = "crashed"
		case DOMAIN_EVENT_STOPPED_MIGRATED:
			detail = "migrated"
		case DOMAIN_EVENT_STOPPED_SAVED:
			detail = "saved"
		case DOMAIN_EVENT_STOPPED_FAILED:
			detail = "failed"
		case DOMAIN_EVENT_STOPPED_FROM_SNAPSHOT:
			detail = "snapshot"
		default:
			detail = "unknown"
		}

	case DOMAIN_EVENT_SHUTDOWN:
		event = "shutdown"
		switch DomainEventShutdownDetailType(e.Detail) {
		case DOMAIN_EVENT_SHUTDOWN_FINISHED:
			detail = "finished"
		default:
			detail = "unknown"
		}

	default:
		event = "unknown"
	}

	return fmt.Sprintf("Domain event=%q detail=%q", event, detail)
}

func (e DomainEventRTCChange) String() string {
	return fmt.Sprintf("RTC change offset=%d", e.Utcoffset)
}

func (e DomainEventWatchdog) String() string {
	return fmt.Sprintf("Watchdog action=%d", e.Action)
}

func (e DomainEventIOError) String() string {
	return fmt.Sprintf("I/O error path=%q alias=%q action=%d",
		e.SrcPath, e.DevAlias, e.Action)
}

func (e DomainEventGraphics) String() string {
	var phase string
	switch e.Phase {
	case DOMAIN_EVENT_GRAPHICS_CONNECT:
		phase = "connected"
	case DOMAIN_EVENT_GRAPHICS_INITIALIZE:
		phase = "initialized"
	case DOMAIN_EVENT_GRAPHICS_DISCONNECT:
		phase = "disconnected"
	default:
		phase = "unknown"
	}

	return fmt.Sprintf("Graphics phase=%q", phase)
}

func (e DomainEventIOErrorReason) String() string {
	return fmt.Sprintf("IO error path=%q alias=%q action=%d reason=%q",
		e.SrcPath, e.DevAlias, e.Action, e.Reason)
}

func (e DomainEventBlockJob) String() string {
	return fmt.Sprintf("Block job disk=%q status=%d type=%d",
		e.Disk, e.Status, e.Type)
}

func (e DomainEventDiskChange) String() string {
	return fmt.Sprintf("Disk change old=%q new=%q alias=%q reason=%d",
		e.OldSrcPath, e.NewSrcPath, e.DevAlias, e.Reason)
}

func (e DomainEventTrayChange) String() string {
	return fmt.Sprintf("Tray change dev=%q reason=%d",
		e.DevAlias, e.Reason)
}

func (e DomainEventBalloonChange) String() string {
	return fmt.Sprintf("Ballon change %d", e.Actual)
}

func (e DomainEventDeviceRemoved) String() string {
	return fmt.Sprintf("Device %q removed ", e.DevAlias)
}
