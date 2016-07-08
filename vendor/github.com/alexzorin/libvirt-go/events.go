package libvirt

import (
	"sync"
	"unsafe"
)

/*
 * Golang 1.6 doesn't support C pointers to go memory.
 * A hacky-solution might be some multi-threaded approach to support domain events, but let's make it work
 * without domain events for now.
 */

/*
#cgo LDFLAGS: -lvirt
#include <libvirt/libvirt.h>

int domainEventLifecycleCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     int event, int detail, void* data);

int domainEventGenericCallback_cgo(virConnectPtr c, virDomainPtr d, void* data);

int domainEventRTCChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     long long utcoffset, void* data);

int domainEventWatchdogCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int action, void* data);

int domainEventIOErrorCallback_cgo(virConnectPtr c, virDomainPtr d,
                                   const char *srcPath, const char *devAlias,
                                   int action, void* data);

int domainEventGraphicsCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int phase, const virDomainEventGraphicsAddress *local,
                                    const virDomainEventGraphicsAddress *remote,
                                    const char *authScheme,
                                    const virDomainEventGraphicsSubject *subject, void* data);

int domainEventIOErrorReasonCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *srcPath, const char *devAlias,
                                         int action, const char *reason, void* data);

int domainEventBlockJobCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    const char *disk, int type, int status, void* data);

int domainEventDiskChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *oldSrcPath, const char *newSrcPath,
                                      const char *devAlias, int reason, void* data);

int domainEventTrayChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *devAlias, int reason, void* data);

int domainEventReasonCallback_cgo(virConnectPtr c, virDomainPtr d,
                                  int reason, void* data);

int domainEventBalloonChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         unsigned long long actual, void* data);

int domainEventDeviceRemovedCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *devAlias, void* data);

int virConnectDomainEventRegisterAny_cgo(virConnectPtr c,  virDomainPtr d,
						                             int eventID, virConnectDomainEventGenericCallback cb,
                                         int goCallbackId);
*/
import "C"

type DomainLifecycleEvent struct {
	Event  int
	Detail int
}

type DomainRTCChangeEvent struct {
	Utcoffset int64
}

type DomainWatchdogEvent struct {
	Action int
}

type DomainIOErrorEvent struct {
	SrcPath  string
	DevAlias string
	Action   int
}

type DomainEventGraphicsAddress struct {
	Family  int
	Node    string
	Service string
}

type DomainEventGraphicsSubjectIdentity struct {
	Type string
	Name string
}

type DomainGraphicsEvent struct {
	Phase      int
	Local      DomainEventGraphicsAddress
	Remote     DomainEventGraphicsAddress
	AuthScheme string
	Subject    []DomainEventGraphicsSubjectIdentity
}

type DomainIOErrorReasonEvent struct {
	DomainIOErrorEvent
	Reason string
}

type DomainBlockJobEvent struct {
	Disk   string
	Type   int
	Status int
}

type DomainDiskChangeEvent struct {
	OldSrcPath string
	NewSrcPath string
	DevAlias   string
	Reason     int
}

type DomainTrayChangeEvent struct {
	DevAlias string
	Reason   int
}

type DomainReasonEvent struct {
	Reason int
}

type DomainBalloonChangeEvent struct {
	Actual uint64
}

type DomainDeviceRemovedEvent struct {
	DevAlias string
}

//export domainEventLifecycleCallback
func domainEventLifecycleCallback(c C.virConnectPtr, d C.virDomainPtr,
	event int, detail int,
	opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainLifecycleEvent{
		Event:  event,
		Detail: detail,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventGenericCallback
func domainEventGenericCallback(c C.virConnectPtr, d C.virDomainPtr,
	opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	return callCallback(opaque, &connection, &domain, nil)
}

//export domainEventRTCChangeCallback
func domainEventRTCChangeCallback(c C.virConnectPtr, d C.virDomainPtr,
	utcoffset int64, opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainRTCChangeEvent{
		Utcoffset: utcoffset,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventWatchdogCallback
func domainEventWatchdogCallback(c C.virConnectPtr, d C.virDomainPtr,
	action int, opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainWatchdogEvent{
		Action: action,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventIOErrorCallback
func domainEventIOErrorCallback(c C.virConnectPtr, d C.virDomainPtr,
	srcPath string, devAlias string, action int, opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainIOErrorEvent{
		SrcPath:  srcPath,
		DevAlias: devAlias,
		Action:   action,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventGraphicsCallback
func domainEventGraphicsCallback(c C.virConnectPtr, d C.virDomainPtr,
	phase int,
	local C.virDomainEventGraphicsAddressPtr,
	remote C.virDomainEventGraphicsAddressPtr,
	authScheme string,
	subject C.virDomainEventGraphicsSubjectPtr,
	opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

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

	eventDetails := DomainGraphicsEvent{
		Phase: phase,
		Local: DomainEventGraphicsAddress{
			Family:  int(local.family),
			Node:    C.GoString(local.node),
			Service: C.GoString(local.service),
		},
		Remote: DomainEventGraphicsAddress{
			Family:  int(remote.family),
			Node:    C.GoString(remote.node),
			Service: C.GoString(remote.service),
		},
		AuthScheme: authScheme,
		Subject:    subjectGo,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventIOErrorReasonCallback
func domainEventIOErrorReasonCallback(c C.virConnectPtr, d C.virDomainPtr,
	srcPath string, devAlias string, action int, reason string,
	opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainIOErrorReasonEvent{
		DomainIOErrorEvent: DomainIOErrorEvent{
			SrcPath:  srcPath,
			DevAlias: devAlias,
			Action:   action,
		},
		Reason: reason,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventBlockJobCallback
func domainEventBlockJobCallback(c C.virConnectPtr, d C.virDomainPtr,
	disk string, _type int, status int, opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainBlockJobEvent{
		Disk:   disk,
		Type:   _type,
		Status: status,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventDiskChangeCallback
func domainEventDiskChangeCallback(c C.virConnectPtr, d C.virDomainPtr,
	oldSrcPath string, newSrcPath string, devAlias string,
	reason int, opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainDiskChangeEvent{
		OldSrcPath: oldSrcPath,
		NewSrcPath: newSrcPath,
		DevAlias:   devAlias,
		Reason:     reason,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventTrayChangeCallback
func domainEventTrayChangeCallback(c C.virConnectPtr, d C.virDomainPtr,
	devAlias string, reason int, opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainTrayChangeEvent{
		DevAlias: devAlias,
		Reason:   reason,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventReasonCallback
func domainEventReasonCallback(c C.virConnectPtr, d C.virDomainPtr,
	reason int, opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainReasonEvent{
		Reason: reason,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventBalloonChangeCallback
func domainEventBalloonChangeCallback(c C.virConnectPtr, d C.virDomainPtr,
	actual uint64, opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainBalloonChangeEvent{
		Actual: actual,
	}

	return callCallback(opaque, &connection, &domain, eventDetails)
}

//export domainEventDeviceRemovedCallback
func domainEventDeviceRemovedCallback(c C.virConnectPtr, d C.virDomainPtr,
	devAlias string, opaque int) int {

	domain := VirDomain{ptr: d}
	connection := VirConnection{ptr: c}

	eventDetails := DomainDeviceRemovedEvent{
		DevAlias: devAlias,
	}
	return callCallback(opaque, &connection, &domain, eventDetails)
}

type DomainEventCallback func(c *VirConnection, d *VirDomain,
	event interface{}, f func()) int

type domainCallbackContext struct {
	cb *DomainEventCallback
	f  func()
}

var goCallbackLock sync.RWMutex
var goCallbacks = make(map[int]*domainCallbackContext)
var nextGoCallbackId int

//export freeCallbackId
func freeCallbackId(goCallbackId int) {
	goCallbackLock.Lock()
	delete(goCallbacks, goCallbackId)
	goCallbackLock.Unlock()
}

func callCallback(goCallbackId int, c *VirConnection, d *VirDomain,
	event interface{}) int {
	goCallbackLock.RLock()
	ctx := goCallbacks[goCallbackId]
	goCallbackLock.RUnlock()
	if ctx == nil {
		// If this happens there must be a bug in libvirt
		panic("Callback arrived after freeCallbackId was called")
	}
	return (*ctx.cb)(c, d, event, ctx.f)
}

func (c *VirConnection) DomainEventRegister(dom VirDomain,
	eventId int,
	callback *DomainEventCallback,
	opaque func()) int {
	var callbackPtr unsafe.Pointer
	context := &domainCallbackContext{
		cb: callback,
		f:  opaque,
	}
	goCallbackLock.Lock()
	goCallBackId := nextGoCallbackId
	nextGoCallbackId++
	goCallbacks[goCallBackId] = context
	goCallbackLock.Unlock()

	switch eventId {
	case VIR_DOMAIN_EVENT_ID_LIFECYCLE:
		callbackPtr = unsafe.Pointer(C.domainEventLifecycleCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_REBOOT:
	case VIR_DOMAIN_EVENT_ID_CONTROL_ERROR:
		callbackPtr = unsafe.Pointer(C.domainEventGenericCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_RTC_CHANGE:
		callbackPtr = unsafe.Pointer(C.domainEventRTCChangeCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_WATCHDOG:
		callbackPtr = unsafe.Pointer(C.domainEventWatchdogCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_IO_ERROR:
		callbackPtr = unsafe.Pointer(C.domainEventIOErrorCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_GRAPHICS:
		callbackPtr = unsafe.Pointer(C.domainEventGraphicsCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_IO_ERROR_REASON:
		callbackPtr = unsafe.Pointer(C.domainEventIOErrorReasonCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_BLOCK_JOB:
		// TODO Post 1.2.4, uncomment later
		// case VIR_DOMAIN_EVENT_ID_BLOCK_JOB_2:
		callbackPtr = unsafe.Pointer(C.domainEventBlockJobCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_DISK_CHANGE:
		callbackPtr = unsafe.Pointer(C.domainEventDiskChangeCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_TRAY_CHANGE:
		callbackPtr = unsafe.Pointer(C.domainEventTrayChangeCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_PMWAKEUP:
	case VIR_DOMAIN_EVENT_ID_PMSUSPEND:
	case VIR_DOMAIN_EVENT_ID_PMSUSPEND_DISK:
		callbackPtr = unsafe.Pointer(C.domainEventReasonCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_BALLOON_CHANGE:
		callbackPtr = unsafe.Pointer(C.domainEventBalloonChangeCallback_cgo)
	case VIR_DOMAIN_EVENT_ID_DEVICE_REMOVED:
		callbackPtr = unsafe.Pointer(C.domainEventDeviceRemovedCallback_cgo)
	default:
	}
	ret := C.virConnectDomainEventRegisterAny_cgo(c.ptr, dom.ptr, C.VIR_DOMAIN_EVENT_ID_LIFECYCLE,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.int(goCallBackId))
	if ret == -1 {
		goCallbackLock.Lock()
		delete(goCallbacks, goCallBackId)
		goCallbackLock.Unlock()
		return -1
	}
	return goCallBackId
}

func (c *VirConnection) DomainEventDeregister(callbackId int) int {
	// Deregister the callback
	return int(C.virConnectDomainEventDeregisterAny(c.ptr, C.int(callbackId)))
}

func EventRegisterDefaultImpl() int {
	return int(C.virEventRegisterDefaultImpl())
}

func EventRunDefaultImpl() int {
	return int(C.virEventRunDefaultImpl())
}
