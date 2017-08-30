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

/*
#cgo pkg-config: libvirt
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
#include "domain_compat.h"
*/
import "C"

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

const (
	DOMAIN_SEND_KEY_MAX_KEYS = uint32(C.VIR_DOMAIN_SEND_KEY_MAX_KEYS)
)

type DomainState int

const (
	DOMAIN_NOSTATE     = DomainState(C.VIR_DOMAIN_NOSTATE)
	DOMAIN_RUNNING     = DomainState(C.VIR_DOMAIN_RUNNING)
	DOMAIN_BLOCKED     = DomainState(C.VIR_DOMAIN_BLOCKED)
	DOMAIN_PAUSED      = DomainState(C.VIR_DOMAIN_PAUSED)
	DOMAIN_SHUTDOWN    = DomainState(C.VIR_DOMAIN_SHUTDOWN)
	DOMAIN_CRASHED     = DomainState(C.VIR_DOMAIN_CRASHED)
	DOMAIN_PMSUSPENDED = DomainState(C.VIR_DOMAIN_PMSUSPENDED)
	DOMAIN_SHUTOFF     = DomainState(C.VIR_DOMAIN_SHUTOFF)
)

type DomainMetadataType int

const (
	DOMAIN_METADATA_DESCRIPTION = DomainMetadataType(C.VIR_DOMAIN_METADATA_DESCRIPTION)
	DOMAIN_METADATA_TITLE       = DomainMetadataType(C.VIR_DOMAIN_METADATA_TITLE)
	DOMAIN_METADATA_ELEMENT     = DomainMetadataType(C.VIR_DOMAIN_METADATA_ELEMENT)
)

type DomainVcpuFlags int

const (
	DOMAIN_VCPU_CONFIG       = DomainVcpuFlags(C.VIR_DOMAIN_VCPU_CONFIG)
	DOMAIN_VCPU_CURRENT      = DomainVcpuFlags(C.VIR_DOMAIN_VCPU_CURRENT)
	DOMAIN_VCPU_LIVE         = DomainVcpuFlags(C.VIR_DOMAIN_VCPU_LIVE)
	DOMAIN_VCPU_MAXIMUM      = DomainVcpuFlags(C.VIR_DOMAIN_VCPU_MAXIMUM)
	DOMAIN_VCPU_GUEST        = DomainVcpuFlags(C.VIR_DOMAIN_VCPU_GUEST)
	DOMAIN_VCPU_HOTPLUGGABLE = DomainVcpuFlags(C.VIR_DOMAIN_VCPU_HOTPLUGGABLE)
)

type DomainModificationImpact int

const (
	DOMAIN_AFFECT_CONFIG  = DomainModificationImpact(C.VIR_DOMAIN_AFFECT_CONFIG)
	DOMAIN_AFFECT_CURRENT = DomainModificationImpact(C.VIR_DOMAIN_AFFECT_CURRENT)
	DOMAIN_AFFECT_LIVE    = DomainModificationImpact(C.VIR_DOMAIN_AFFECT_LIVE)
)

type DomainMemoryModFlags int

const (
	DOMAIN_MEM_CONFIG  = DomainMemoryModFlags(C.VIR_DOMAIN_MEM_CONFIG)
	DOMAIN_MEM_CURRENT = DomainMemoryModFlags(C.VIR_DOMAIN_MEM_CURRENT)
	DOMAIN_MEM_LIVE    = DomainMemoryModFlags(C.VIR_DOMAIN_MEM_LIVE)
	DOMAIN_MEM_MAXIMUM = DomainMemoryModFlags(C.VIR_DOMAIN_MEM_MAXIMUM)
)

type DomainDestroyFlags int

const (
	DOMAIN_DESTROY_DEFAULT  = DomainDestroyFlags(C.VIR_DOMAIN_DESTROY_DEFAULT)
	DOMAIN_DESTROY_GRACEFUL = DomainDestroyFlags(C.VIR_DOMAIN_DESTROY_GRACEFUL)
)

type DomainShutdownFlags int

const (
	DOMAIN_SHUTDOWN_DEFAULT        = DomainShutdownFlags(C.VIR_DOMAIN_SHUTDOWN_DEFAULT)
	DOMAIN_SHUTDOWN_ACPI_POWER_BTN = DomainShutdownFlags(C.VIR_DOMAIN_SHUTDOWN_ACPI_POWER_BTN)
	DOMAIN_SHUTDOWN_GUEST_AGENT    = DomainShutdownFlags(C.VIR_DOMAIN_SHUTDOWN_GUEST_AGENT)
	DOMAIN_SHUTDOWN_INITCTL        = DomainShutdownFlags(C.VIR_DOMAIN_SHUTDOWN_INITCTL)
	DOMAIN_SHUTDOWN_SIGNAL         = DomainShutdownFlags(C.VIR_DOMAIN_SHUTDOWN_SIGNAL)
	DOMAIN_SHUTDOWN_PARAVIRT       = DomainShutdownFlags(C.VIR_DOMAIN_SHUTDOWN_PARAVIRT)
)

type DomainUndefineFlagsValues int

const (
	DOMAIN_UNDEFINE_MANAGED_SAVE       = DomainUndefineFlagsValues(C.VIR_DOMAIN_UNDEFINE_MANAGED_SAVE)       // Also remove any managed save
	DOMAIN_UNDEFINE_SNAPSHOTS_METADATA = DomainUndefineFlagsValues(C.VIR_DOMAIN_UNDEFINE_SNAPSHOTS_METADATA) // If last use of domain, then also remove any snapshot metadata
	DOMAIN_UNDEFINE_NVRAM              = DomainUndefineFlagsValues(C.VIR_DOMAIN_UNDEFINE_NVRAM)              // Also remove any nvram file
	DOMAIN_UNDEFINE_KEEP_NVRAM         = DomainUndefineFlagsValues(C.VIR_DOMAIN_UNDEFINE_KEEP_NVRAM)         // Keep nvram file
)

type DomainDeviceModifyFlags int

const (
	DOMAIN_DEVICE_MODIFY_CONFIG  = DomainDeviceModifyFlags(C.VIR_DOMAIN_DEVICE_MODIFY_CONFIG)
	DOMAIN_DEVICE_MODIFY_CURRENT = DomainDeviceModifyFlags(C.VIR_DOMAIN_DEVICE_MODIFY_CURRENT)
	DOMAIN_DEVICE_MODIFY_LIVE    = DomainDeviceModifyFlags(C.VIR_DOMAIN_DEVICE_MODIFY_LIVE)
	DOMAIN_DEVICE_MODIFY_FORCE   = DomainDeviceModifyFlags(C.VIR_DOMAIN_DEVICE_MODIFY_FORCE)
)

type DomainCreateFlags int

const (
	DOMAIN_NONE               = DomainCreateFlags(C.VIR_DOMAIN_NONE)
	DOMAIN_START_PAUSED       = DomainCreateFlags(C.VIR_DOMAIN_START_PAUSED)
	DOMAIN_START_AUTODESTROY  = DomainCreateFlags(C.VIR_DOMAIN_START_AUTODESTROY)
	DOMAIN_START_BYPASS_CACHE = DomainCreateFlags(C.VIR_DOMAIN_START_BYPASS_CACHE)
	DOMAIN_START_FORCE_BOOT   = DomainCreateFlags(C.VIR_DOMAIN_START_FORCE_BOOT)
	DOMAIN_START_VALIDATE     = DomainCreateFlags(C.VIR_DOMAIN_START_VALIDATE)
)

const DOMAIN_MEMORY_PARAM_UNLIMITED = C.VIR_DOMAIN_MEMORY_PARAM_UNLIMITED

type DomainEventType int

const (
	DOMAIN_EVENT_DEFINED     = DomainEventType(C.VIR_DOMAIN_EVENT_DEFINED)
	DOMAIN_EVENT_UNDEFINED   = DomainEventType(C.VIR_DOMAIN_EVENT_UNDEFINED)
	DOMAIN_EVENT_STARTED     = DomainEventType(C.VIR_DOMAIN_EVENT_STARTED)
	DOMAIN_EVENT_SUSPENDED   = DomainEventType(C.VIR_DOMAIN_EVENT_SUSPENDED)
	DOMAIN_EVENT_RESUMED     = DomainEventType(C.VIR_DOMAIN_EVENT_RESUMED)
	DOMAIN_EVENT_STOPPED     = DomainEventType(C.VIR_DOMAIN_EVENT_STOPPED)
	DOMAIN_EVENT_SHUTDOWN    = DomainEventType(C.VIR_DOMAIN_EVENT_SHUTDOWN)
	DOMAIN_EVENT_PMSUSPENDED = DomainEventType(C.VIR_DOMAIN_EVENT_PMSUSPENDED)
	DOMAIN_EVENT_CRASHED     = DomainEventType(C.VIR_DOMAIN_EVENT_CRASHED)
)

type DomainEventWatchdogAction int

// The action that is to be taken due to the watchdog device firing
const (
	// No action, watchdog ignored
	DOMAIN_EVENT_WATCHDOG_NONE = DomainEventWatchdogAction(C.VIR_DOMAIN_EVENT_WATCHDOG_NONE)

	// Guest CPUs are paused
	DOMAIN_EVENT_WATCHDOG_PAUSE = DomainEventWatchdogAction(C.VIR_DOMAIN_EVENT_WATCHDOG_PAUSE)

	// Guest CPUs are reset
	DOMAIN_EVENT_WATCHDOG_RESET = DomainEventWatchdogAction(C.VIR_DOMAIN_EVENT_WATCHDOG_RESET)

	// Guest is forcibly powered off
	DOMAIN_EVENT_WATCHDOG_POWEROFF = DomainEventWatchdogAction(C.VIR_DOMAIN_EVENT_WATCHDOG_POWEROFF)

	// Guest is requested to gracefully shutdown
	DOMAIN_EVENT_WATCHDOG_SHUTDOWN = DomainEventWatchdogAction(C.VIR_DOMAIN_EVENT_WATCHDOG_SHUTDOWN)

	// No action, a debug message logged
	DOMAIN_EVENT_WATCHDOG_DEBUG = DomainEventWatchdogAction(C.VIR_DOMAIN_EVENT_WATCHDOG_DEBUG)

	// Inject a non-maskable interrupt into guest
	DOMAIN_EVENT_WATCHDOG_INJECTNMI = DomainEventWatchdogAction(C.VIR_DOMAIN_EVENT_WATCHDOG_INJECTNMI)
)

type DomainEventIOErrorAction int

// The action that is to be taken due to an IO error occurring
const (
	// No action, IO error ignored
	DOMAIN_EVENT_IO_ERROR_NONE = DomainEventIOErrorAction(C.VIR_DOMAIN_EVENT_IO_ERROR_NONE)

	// Guest CPUs are paused
	DOMAIN_EVENT_IO_ERROR_PAUSE = DomainEventIOErrorAction(C.VIR_DOMAIN_EVENT_IO_ERROR_PAUSE)

	// IO error reported to guest OS
	DOMAIN_EVENT_IO_ERROR_REPORT = DomainEventIOErrorAction(C.VIR_DOMAIN_EVENT_IO_ERROR_REPORT)
)

type DomainEventGraphicsPhase int

// The phase of the graphics client connection
const (
	// Initial socket connection established
	DOMAIN_EVENT_GRAPHICS_CONNECT = DomainEventGraphicsPhase(C.VIR_DOMAIN_EVENT_GRAPHICS_CONNECT)

	// Authentication & setup completed
	DOMAIN_EVENT_GRAPHICS_INITIALIZE = DomainEventGraphicsPhase(C.VIR_DOMAIN_EVENT_GRAPHICS_INITIALIZE)

	// Final socket disconnection
	DOMAIN_EVENT_GRAPHICS_DISCONNECT = DomainEventGraphicsPhase(C.VIR_DOMAIN_EVENT_GRAPHICS_DISCONNECT)
)

type DomainEventGraphicsAddressType int

const (
	// IPv4 address
	DOMAIN_EVENT_GRAPHICS_ADDRESS_IPV4 = DomainEventGraphicsAddressType(C.VIR_DOMAIN_EVENT_GRAPHICS_ADDRESS_IPV4)

	// IPv6 address
	DOMAIN_EVENT_GRAPHICS_ADDRESS_IPV6 = DomainEventGraphicsAddressType(C.VIR_DOMAIN_EVENT_GRAPHICS_ADDRESS_IPV6)

	// UNIX socket path
	DOMAIN_EVENT_GRAPHICS_ADDRESS_UNIX = DomainEventGraphicsAddressType(C.VIR_DOMAIN_EVENT_GRAPHICS_ADDRESS_UNIX)
)

type DomainBlockJobType int

const (
	// Placeholder
	DOMAIN_BLOCK_JOB_TYPE_UNKNOWN = DomainBlockJobType(C.VIR_DOMAIN_BLOCK_JOB_TYPE_UNKNOWN)

	// Block Pull (virDomainBlockPull, or virDomainBlockRebase without
	// flags), job ends on completion
	DOMAIN_BLOCK_JOB_TYPE_PULL = DomainBlockJobType(C.VIR_DOMAIN_BLOCK_JOB_TYPE_PULL)

	// Block Copy (virDomainBlockCopy, or virDomainBlockRebase with
	// flags), job exists as long as mirroring is active
	DOMAIN_BLOCK_JOB_TYPE_COPY = DomainBlockJobType(C.VIR_DOMAIN_BLOCK_JOB_TYPE_COPY)

	// Block Commit (virDomainBlockCommit without flags), job ends on
	// completion
	DOMAIN_BLOCK_JOB_TYPE_COMMIT = DomainBlockJobType(C.VIR_DOMAIN_BLOCK_JOB_TYPE_COMMIT)

	// Active Block Commit (virDomainBlockCommit with flags), job
	// exists as long as sync is active
	DOMAIN_BLOCK_JOB_TYPE_ACTIVE_COMMIT = DomainBlockJobType(C.VIR_DOMAIN_BLOCK_JOB_TYPE_ACTIVE_COMMIT)
)

type DomainRunningReason int

const (
	DOMAIN_RUNNING_UNKNOWN            = DomainRunningReason(C.VIR_DOMAIN_RUNNING_UNKNOWN)
	DOMAIN_RUNNING_BOOTED             = DomainRunningReason(C.VIR_DOMAIN_RUNNING_BOOTED)             /* normal startup from boot */
	DOMAIN_RUNNING_MIGRATED           = DomainRunningReason(C.VIR_DOMAIN_RUNNING_MIGRATED)           /* migrated from another host */
	DOMAIN_RUNNING_RESTORED           = DomainRunningReason(C.VIR_DOMAIN_RUNNING_RESTORED)           /* restored from a state file */
	DOMAIN_RUNNING_FROM_SNAPSHOT      = DomainRunningReason(C.VIR_DOMAIN_RUNNING_FROM_SNAPSHOT)      /* restored from snapshot */
	DOMAIN_RUNNING_UNPAUSED           = DomainRunningReason(C.VIR_DOMAIN_RUNNING_UNPAUSED)           /* returned from paused state */
	DOMAIN_RUNNING_MIGRATION_CANCELED = DomainRunningReason(C.VIR_DOMAIN_RUNNING_MIGRATION_CANCELED) /* returned from migration */
	DOMAIN_RUNNING_SAVE_CANCELED      = DomainRunningReason(C.VIR_DOMAIN_RUNNING_SAVE_CANCELED)      /* returned from failed save process */
	DOMAIN_RUNNING_WAKEUP             = DomainRunningReason(C.VIR_DOMAIN_RUNNING_WAKEUP)             /* returned from pmsuspended due to wakeup event */
	DOMAIN_RUNNING_CRASHED            = DomainRunningReason(C.VIR_DOMAIN_RUNNING_CRASHED)            /* resumed from crashed */
	DOMAIN_RUNNING_POSTCOPY           = DomainRunningReason(C.VIR_DOMAIN_RUNNING_POSTCOPY)           /* running in post-copy migration mode */
)

type DomainPausedReason int

const (
	DOMAIN_PAUSED_UNKNOWN         = DomainPausedReason(C.VIR_DOMAIN_PAUSED_UNKNOWN)         /* the reason is unknown */
	DOMAIN_PAUSED_USER            = DomainPausedReason(C.VIR_DOMAIN_PAUSED_USER)            /* paused on user request */
	DOMAIN_PAUSED_MIGRATION       = DomainPausedReason(C.VIR_DOMAIN_PAUSED_MIGRATION)       /* paused for offline migration */
	DOMAIN_PAUSED_SAVE            = DomainPausedReason(C.VIR_DOMAIN_PAUSED_SAVE)            /* paused for save */
	DOMAIN_PAUSED_DUMP            = DomainPausedReason(C.VIR_DOMAIN_PAUSED_DUMP)            /* paused for offline core dump */
	DOMAIN_PAUSED_IOERROR         = DomainPausedReason(C.VIR_DOMAIN_PAUSED_IOERROR)         /* paused due to a disk I/O error */
	DOMAIN_PAUSED_WATCHDOG        = DomainPausedReason(C.VIR_DOMAIN_PAUSED_WATCHDOG)        /* paused due to a watchdog event */
	DOMAIN_PAUSED_FROM_SNAPSHOT   = DomainPausedReason(C.VIR_DOMAIN_PAUSED_FROM_SNAPSHOT)   /* paused after restoring from snapshot */
	DOMAIN_PAUSED_SHUTTING_DOWN   = DomainPausedReason(C.VIR_DOMAIN_PAUSED_SHUTTING_DOWN)   /* paused during shutdown process */
	DOMAIN_PAUSED_SNAPSHOT        = DomainPausedReason(C.VIR_DOMAIN_PAUSED_SNAPSHOT)        /* paused while creating a snapshot */
	DOMAIN_PAUSED_CRASHED         = DomainPausedReason(C.VIR_DOMAIN_PAUSED_CRASHED)         /* paused due to a guest crash */
	DOMAIN_PAUSED_STARTING_UP     = DomainPausedReason(C.VIR_DOMAIN_PAUSED_STARTING_UP)     /* the domainis being started */
	DOMAIN_PAUSED_POSTCOPY        = DomainPausedReason(C.VIR_DOMAIN_PAUSED_POSTCOPY)        /* paused for post-copy migration */
	DOMAIN_PAUSED_POSTCOPY_FAILED = DomainPausedReason(C.VIR_DOMAIN_PAUSED_POSTCOPY_FAILED) /* paused after failed post-copy */
)

type DomainXMLFlags int

const (
	DOMAIN_XML_SECURE     = DomainXMLFlags(C.VIR_DOMAIN_XML_SECURE)     /* dump security sensitive information too */
	DOMAIN_XML_INACTIVE   = DomainXMLFlags(C.VIR_DOMAIN_XML_INACTIVE)   /* dump inactive domain information */
	DOMAIN_XML_UPDATE_CPU = DomainXMLFlags(C.VIR_DOMAIN_XML_UPDATE_CPU) /* update guest CPU requirements according to host CPU */
	DOMAIN_XML_MIGRATABLE = DomainXMLFlags(C.VIR_DOMAIN_XML_MIGRATABLE) /* dump XML suitable for migration */
)

type DomainEventDefinedDetailType int

const (
	DOMAIN_EVENT_DEFINED_ADDED         = DomainEventDefinedDetailType(C.VIR_DOMAIN_EVENT_DEFINED_ADDED)
	DOMAIN_EVENT_DEFINED_UPDATED       = DomainEventDefinedDetailType(C.VIR_DOMAIN_EVENT_DEFINED_UPDATED)
	DOMAIN_EVENT_DEFINED_RENAMED       = DomainEventDefinedDetailType(C.VIR_DOMAIN_EVENT_DEFINED_RENAMED)
	DOMAIN_EVENT_DEFINED_FROM_SNAPSHOT = DomainEventDefinedDetailType(C.VIR_DOMAIN_EVENT_DEFINED_FROM_SNAPSHOT)
)

type DomainEventUndefinedDetailType int

const (
	DOMAIN_EVENT_UNDEFINED_REMOVED = DomainEventUndefinedDetailType(C.VIR_DOMAIN_EVENT_UNDEFINED_REMOVED)
	DOMAIN_EVENT_UNDEFINED_RENAMED = DomainEventUndefinedDetailType(C.VIR_DOMAIN_EVENT_UNDEFINED_RENAMED)
)

type DomainEventStartedDetailType int

const (
	DOMAIN_EVENT_STARTED_BOOTED        = DomainEventStartedDetailType(C.VIR_DOMAIN_EVENT_STARTED_BOOTED)
	DOMAIN_EVENT_STARTED_MIGRATED      = DomainEventStartedDetailType(C.VIR_DOMAIN_EVENT_STARTED_MIGRATED)
	DOMAIN_EVENT_STARTED_RESTORED      = DomainEventStartedDetailType(C.VIR_DOMAIN_EVENT_STARTED_RESTORED)
	DOMAIN_EVENT_STARTED_FROM_SNAPSHOT = DomainEventStartedDetailType(C.VIR_DOMAIN_EVENT_STARTED_FROM_SNAPSHOT)
	DOMAIN_EVENT_STARTED_WAKEUP        = DomainEventStartedDetailType(C.VIR_DOMAIN_EVENT_STARTED_WAKEUP)
)

type DomainEventSuspendedDetailType int

const (
	DOMAIN_EVENT_SUSPENDED_PAUSED          = DomainEventSuspendedDetailType(C.VIR_DOMAIN_EVENT_SUSPENDED_PAUSED)
	DOMAIN_EVENT_SUSPENDED_MIGRATED        = DomainEventSuspendedDetailType(C.VIR_DOMAIN_EVENT_SUSPENDED_MIGRATED)
	DOMAIN_EVENT_SUSPENDED_IOERROR         = DomainEventSuspendedDetailType(C.VIR_DOMAIN_EVENT_SUSPENDED_IOERROR)
	DOMAIN_EVENT_SUSPENDED_WATCHDOG        = DomainEventSuspendedDetailType(C.VIR_DOMAIN_EVENT_SUSPENDED_WATCHDOG)
	DOMAIN_EVENT_SUSPENDED_RESTORED        = DomainEventSuspendedDetailType(C.VIR_DOMAIN_EVENT_SUSPENDED_RESTORED)
	DOMAIN_EVENT_SUSPENDED_FROM_SNAPSHOT   = DomainEventSuspendedDetailType(C.VIR_DOMAIN_EVENT_SUSPENDED_FROM_SNAPSHOT)
	DOMAIN_EVENT_SUSPENDED_API_ERROR       = DomainEventSuspendedDetailType(C.VIR_DOMAIN_EVENT_SUSPENDED_API_ERROR)
	DOMAIN_EVENT_SUSPENDED_POSTCOPY        = DomainEventSuspendedDetailType(C.VIR_DOMAIN_EVENT_SUSPENDED_POSTCOPY)
	DOMAIN_EVENT_SUSPENDED_POSTCOPY_FAILED = DomainEventSuspendedDetailType(C.VIR_DOMAIN_EVENT_SUSPENDED_POSTCOPY_FAILED)
)

type DomainEventResumedDetailType int

const (
	DOMAIN_EVENT_RESUMED_UNPAUSED      = DomainEventResumedDetailType(C.VIR_DOMAIN_EVENT_RESUMED_UNPAUSED)
	DOMAIN_EVENT_RESUMED_MIGRATED      = DomainEventResumedDetailType(C.VIR_DOMAIN_EVENT_RESUMED_MIGRATED)
	DOMAIN_EVENT_RESUMED_FROM_SNAPSHOT = DomainEventResumedDetailType(C.VIR_DOMAIN_EVENT_RESUMED_FROM_SNAPSHOT)
	DOMAIN_EVENT_RESUMED_POSTCOPY      = DomainEventResumedDetailType(C.VIR_DOMAIN_EVENT_RESUMED_POSTCOPY)
)

type DomainEventStoppedDetailType int

const (
	DOMAIN_EVENT_STOPPED_SHUTDOWN      = DomainEventStoppedDetailType(C.VIR_DOMAIN_EVENT_STOPPED_SHUTDOWN)
	DOMAIN_EVENT_STOPPED_DESTROYED     = DomainEventStoppedDetailType(C.VIR_DOMAIN_EVENT_STOPPED_DESTROYED)
	DOMAIN_EVENT_STOPPED_CRASHED       = DomainEventStoppedDetailType(C.VIR_DOMAIN_EVENT_STOPPED_CRASHED)
	DOMAIN_EVENT_STOPPED_MIGRATED      = DomainEventStoppedDetailType(C.VIR_DOMAIN_EVENT_STOPPED_MIGRATED)
	DOMAIN_EVENT_STOPPED_SAVED         = DomainEventStoppedDetailType(C.VIR_DOMAIN_EVENT_STOPPED_SAVED)
	DOMAIN_EVENT_STOPPED_FAILED        = DomainEventStoppedDetailType(C.VIR_DOMAIN_EVENT_STOPPED_FAILED)
	DOMAIN_EVENT_STOPPED_FROM_SNAPSHOT = DomainEventStoppedDetailType(C.VIR_DOMAIN_EVENT_STOPPED_FROM_SNAPSHOT)
)

type DomainEventShutdownDetailType int

const (
	DOMAIN_EVENT_SHUTDOWN_FINISHED = DomainEventShutdownDetailType(C.VIR_DOMAIN_EVENT_SHUTDOWN_FINISHED)
	DOMAIN_EVENT_SHUTDOWN_GUEST    = DomainEventShutdownDetailType(C.VIR_DOMAIN_EVENT_SHUTDOWN_GUEST)
	DOMAIN_EVENT_SHUTDOWN_HOST     = DomainEventShutdownDetailType(C.VIR_DOMAIN_EVENT_SHUTDOWN_HOST)
)

type DomainMemoryStatTags int

const (
	DOMAIN_MEMORY_STAT_LAST           = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_NR)
	DOMAIN_MEMORY_STAT_SWAP_IN        = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_SWAP_IN)
	DOMAIN_MEMORY_STAT_SWAP_OUT       = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_SWAP_OUT)
	DOMAIN_MEMORY_STAT_MAJOR_FAULT    = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_MAJOR_FAULT)
	DOMAIN_MEMORY_STAT_MINOR_FAULT    = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_MINOR_FAULT)
	DOMAIN_MEMORY_STAT_UNUSED         = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_UNUSED)
	DOMAIN_MEMORY_STAT_AVAILABLE      = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_AVAILABLE)
	DOMAIN_MEMORY_STAT_ACTUAL_BALLOON = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON)
	DOMAIN_MEMORY_STAT_RSS            = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_RSS)
	DOMAIN_MEMORY_STAT_USABLE         = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_USABLE)
	DOMAIN_MEMORY_STAT_LAST_UPDATE    = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_LAST_UPDATE)
	DOMAIN_MEMORY_STAT_NR             = DomainMemoryStatTags(C.VIR_DOMAIN_MEMORY_STAT_NR)
)

type DomainCPUStatsTags string

const (
	DOMAIN_CPU_STATS_CPUTIME    = DomainCPUStatsTags(C.VIR_DOMAIN_CPU_STATS_CPUTIME)
	DOMAIN_CPU_STATS_SYSTEMTIME = DomainCPUStatsTags(C.VIR_DOMAIN_CPU_STATS_SYSTEMTIME)
	DOMAIN_CPU_STATS_USERTIME   = DomainCPUStatsTags(C.VIR_DOMAIN_CPU_STATS_USERTIME)
	DOMAIN_CPU_STATS_VCPUTIME   = DomainCPUStatsTags(C.VIR_DOMAIN_CPU_STATS_VCPUTIME)
)

type DomainInterfaceAddressesSource int

const (
	DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE = DomainInterfaceAddressesSource(C.VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT = DomainInterfaceAddressesSource(C.VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT)
)

type KeycodeSet int

const (
	KEYCODE_SET_LINUX  = KeycodeSet(C.VIR_KEYCODE_SET_LINUX)
	KEYCODE_SET_XT     = KeycodeSet(C.VIR_KEYCODE_SET_XT)
	KEYCODE_SET_ATSET1 = KeycodeSet(C.VIR_KEYCODE_SET_ATSET1)
	KEYCODE_SET_ATSET2 = KeycodeSet(C.VIR_KEYCODE_SET_ATSET2)
	KEYCODE_SET_ATSET3 = KeycodeSet(C.VIR_KEYCODE_SET_ATSET3)
	KEYCODE_SET_OSX    = KeycodeSet(C.VIR_KEYCODE_SET_OSX)
	KEYCODE_SET_XT_KBD = KeycodeSet(C.VIR_KEYCODE_SET_XT_KBD)
	KEYCODE_SET_USB    = KeycodeSet(C.VIR_KEYCODE_SET_USB)
	KEYCODE_SET_WIN32  = KeycodeSet(C.VIR_KEYCODE_SET_WIN32)
	KEYCODE_SET_RFB    = KeycodeSet(C.VIR_KEYCODE_SET_RFB)
)

type ConnectDomainEventBlockJobStatus int

const (
	DOMAIN_BLOCK_JOB_COMPLETED = ConnectDomainEventBlockJobStatus(C.VIR_DOMAIN_BLOCK_JOB_COMPLETED)
	DOMAIN_BLOCK_JOB_FAILED    = ConnectDomainEventBlockJobStatus(C.VIR_DOMAIN_BLOCK_JOB_FAILED)
	DOMAIN_BLOCK_JOB_CANCELED  = ConnectDomainEventBlockJobStatus(C.VIR_DOMAIN_BLOCK_JOB_CANCELED)
	DOMAIN_BLOCK_JOB_READY     = ConnectDomainEventBlockJobStatus(C.VIR_DOMAIN_BLOCK_JOB_READY)
)

type ConnectDomainEventDiskChangeReason int

const (
	// OldSrcPath is set
	DOMAIN_EVENT_DISK_CHANGE_MISSING_ON_START = ConnectDomainEventDiskChangeReason(C.VIR_DOMAIN_EVENT_DISK_CHANGE_MISSING_ON_START)
	DOMAIN_EVENT_DISK_DROP_MISSING_ON_START   = ConnectDomainEventDiskChangeReason(C.VIR_DOMAIN_EVENT_DISK_DROP_MISSING_ON_START)
)

type ConnectDomainEventTrayChangeReason int

const (
	DOMAIN_EVENT_TRAY_CHANGE_OPEN  = ConnectDomainEventTrayChangeReason(C.VIR_DOMAIN_EVENT_TRAY_CHANGE_OPEN)
	DOMAIN_EVENT_TRAY_CHANGE_CLOSE = ConnectDomainEventTrayChangeReason(C.VIR_DOMAIN_EVENT_TRAY_CHANGE_CLOSE)
)

type DomainProcessSignal int

const (
	DOMAIN_PROCESS_SIGNAL_NOP  = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_NOP)
	DOMAIN_PROCESS_SIGNAL_HUP  = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_HUP)
	DOMAIN_PROCESS_SIGNAL_INT  = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_INT)
	DOMAIN_PROCESS_SIGNAL_QUIT = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_QUIT)
	DOMAIN_PROCESS_SIGNAL_ILL  = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_ILL)
	DOMAIN_PROCESS_SIGNAL_TRAP = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_TRAP)
	DOMAIN_PROCESS_SIGNAL_ABRT = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_ABRT)
	DOMAIN_PROCESS_SIGNAL_BUS  = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_BUS)
	DOMAIN_PROCESS_SIGNAL_FPE  = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_FPE)
	DOMAIN_PROCESS_SIGNAL_KILL = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_KILL)

	DOMAIN_PROCESS_SIGNAL_USR1   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_USR1)
	DOMAIN_PROCESS_SIGNAL_SEGV   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_SEGV)
	DOMAIN_PROCESS_SIGNAL_USR2   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_USR2)
	DOMAIN_PROCESS_SIGNAL_PIPE   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_PIPE)
	DOMAIN_PROCESS_SIGNAL_ALRM   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_ALRM)
	DOMAIN_PROCESS_SIGNAL_TERM   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_TERM)
	DOMAIN_PROCESS_SIGNAL_STKFLT = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_STKFLT)
	DOMAIN_PROCESS_SIGNAL_CHLD   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_CHLD)
	DOMAIN_PROCESS_SIGNAL_CONT   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_CONT)
	DOMAIN_PROCESS_SIGNAL_STOP   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_STOP)

	DOMAIN_PROCESS_SIGNAL_TSTP   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_TSTP)
	DOMAIN_PROCESS_SIGNAL_TTIN   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_TTIN)
	DOMAIN_PROCESS_SIGNAL_TTOU   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_TTOU)
	DOMAIN_PROCESS_SIGNAL_URG    = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_URG)
	DOMAIN_PROCESS_SIGNAL_XCPU   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_XCPU)
	DOMAIN_PROCESS_SIGNAL_XFSZ   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_XFSZ)
	DOMAIN_PROCESS_SIGNAL_VTALRM = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_VTALRM)
	DOMAIN_PROCESS_SIGNAL_PROF   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_PROF)
	DOMAIN_PROCESS_SIGNAL_WINCH  = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_WINCH)
	DOMAIN_PROCESS_SIGNAL_POLL   = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_POLL)

	DOMAIN_PROCESS_SIGNAL_PWR = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_PWR)
	DOMAIN_PROCESS_SIGNAL_SYS = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_SYS)
	DOMAIN_PROCESS_SIGNAL_RT0 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT0)
	DOMAIN_PROCESS_SIGNAL_RT1 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT1)
	DOMAIN_PROCESS_SIGNAL_RT2 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT2)
	DOMAIN_PROCESS_SIGNAL_RT3 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT3)
	DOMAIN_PROCESS_SIGNAL_RT4 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT4)
	DOMAIN_PROCESS_SIGNAL_RT5 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT5)
	DOMAIN_PROCESS_SIGNAL_RT6 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT6)
	DOMAIN_PROCESS_SIGNAL_RT7 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT7)

	DOMAIN_PROCESS_SIGNAL_RT8  = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT8)
	DOMAIN_PROCESS_SIGNAL_RT9  = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT9)
	DOMAIN_PROCESS_SIGNAL_RT10 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT10)
	DOMAIN_PROCESS_SIGNAL_RT11 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT11)
	DOMAIN_PROCESS_SIGNAL_RT12 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT12)
	DOMAIN_PROCESS_SIGNAL_RT13 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT13)
	DOMAIN_PROCESS_SIGNAL_RT14 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT14)
	DOMAIN_PROCESS_SIGNAL_RT15 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT15)
	DOMAIN_PROCESS_SIGNAL_RT16 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT16)
	DOMAIN_PROCESS_SIGNAL_RT17 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT17)
	DOMAIN_PROCESS_SIGNAL_RT18 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT18)

	DOMAIN_PROCESS_SIGNAL_RT19 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT19)
	DOMAIN_PROCESS_SIGNAL_RT20 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT20)
	DOMAIN_PROCESS_SIGNAL_RT21 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT21)
	DOMAIN_PROCESS_SIGNAL_RT22 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT22)
	DOMAIN_PROCESS_SIGNAL_RT23 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT23)
	DOMAIN_PROCESS_SIGNAL_RT24 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT24)
	DOMAIN_PROCESS_SIGNAL_RT25 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT25)
	DOMAIN_PROCESS_SIGNAL_RT26 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT26)
	DOMAIN_PROCESS_SIGNAL_RT27 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT27)

	DOMAIN_PROCESS_SIGNAL_RT28 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT28)
	DOMAIN_PROCESS_SIGNAL_RT29 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT29)
	DOMAIN_PROCESS_SIGNAL_RT30 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT30)
	DOMAIN_PROCESS_SIGNAL_RT31 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT31)
	DOMAIN_PROCESS_SIGNAL_RT32 = DomainProcessSignal(C.VIR_DOMAIN_PROCESS_SIGNAL_RT32)
)

type DomainBlockedReason int

const (
	DOMAIN_BLOCKED_UNKNOWN = DomainBlockedReason(C.VIR_DOMAIN_BLOCKED_UNKNOWN)
)

type DomainControlState int

const (
	DOMAIN_CONTROL_OK       = DomainControlState(C.VIR_DOMAIN_CONTROL_OK)
	DOMAIN_CONTROL_JOB      = DomainControlState(C.VIR_DOMAIN_CONTROL_JOB)
	DOMAIN_CONTROL_OCCUPIED = DomainControlState(C.VIR_DOMAIN_CONTROL_OCCUPIED)
	DOMAIN_CONTROL_ERROR    = DomainControlState(C.VIR_DOMAIN_CONTROL_ERROR)
)

type DomainControlErrorReason int

const (
	DOMAIN_CONTROL_ERROR_REASON_NONE     = DomainControlErrorReason(C.VIR_DOMAIN_CONTROL_ERROR_REASON_NONE)
	DOMAIN_CONTROL_ERROR_REASON_UNKNOWN  = DomainControlErrorReason(C.VIR_DOMAIN_CONTROL_ERROR_REASON_UNKNOWN)
	DOMAIN_CONTROL_ERROR_REASON_MONITOR  = DomainControlErrorReason(C.VIR_DOMAIN_CONTROL_ERROR_REASON_MONITOR)
	DOMAIN_CONTROL_ERROR_REASON_INTERNAL = DomainControlErrorReason(C.VIR_DOMAIN_CONTROL_ERROR_REASON_INTERNAL)
)

type DomainCrashedReason int

const (
	DOMAIN_CRASHED_UNKNOWN  = DomainCrashedReason(C.VIR_DOMAIN_CRASHED_UNKNOWN)
	DOMAIN_CRASHED_PANICKED = DomainCrashedReason(C.VIR_DOMAIN_CRASHED_PANICKED)
)

type DomainEventCrashedDetailType int

const (
	DOMAIN_EVENT_CRASHED_PANICKED = DomainEventCrashedDetailType(C.VIR_DOMAIN_EVENT_CRASHED_PANICKED)
)

type DomainEventPMSuspendedDetailType int

const (
	DOMAIN_EVENT_PMSUSPENDED_MEMORY = DomainEventPMSuspendedDetailType(C.VIR_DOMAIN_EVENT_PMSUSPENDED_MEMORY)
	DOMAIN_EVENT_PMSUSPENDED_DISK   = DomainEventPMSuspendedDetailType(C.VIR_DOMAIN_EVENT_PMSUSPENDED_DISK)
)

type DomainNostateReason int

const (
	DOMAIN_NOSTATE_UNKNOWN = DomainNostateReason(C.VIR_DOMAIN_NOSTATE_UNKNOWN)
)

type DomainPMSuspendedReason int

const (
	DOMAIN_PMSUSPENDED_UNKNOWN = DomainPMSuspendedReason(C.VIR_DOMAIN_PMSUSPENDED_UNKNOWN)
)

type DomainPMSuspendedDiskReason int

const (
	DOMAIN_PMSUSPENDED_DISK_UNKNOWN = DomainPMSuspendedDiskReason(C.VIR_DOMAIN_PMSUSPENDED_DISK_UNKNOWN)
)

type DomainShutdownReason int

const (
	DOMAIN_SHUTDOWN_UNKNOWN = DomainShutdownReason(C.VIR_DOMAIN_SHUTDOWN_UNKNOWN)
	DOMAIN_SHUTDOWN_USER    = DomainShutdownReason(C.VIR_DOMAIN_SHUTDOWN_USER)
)

type DomainShutoffReason int

const (
	DOMAIN_SHUTOFF_UNKNOWN       = DomainShutoffReason(C.VIR_DOMAIN_SHUTOFF_UNKNOWN)
	DOMAIN_SHUTOFF_SHUTDOWN      = DomainShutoffReason(C.VIR_DOMAIN_SHUTOFF_SHUTDOWN)
	DOMAIN_SHUTOFF_DESTROYED     = DomainShutoffReason(C.VIR_DOMAIN_SHUTOFF_DESTROYED)
	DOMAIN_SHUTOFF_CRASHED       = DomainShutoffReason(C.VIR_DOMAIN_SHUTOFF_CRASHED)
	DOMAIN_SHUTOFF_MIGRATED      = DomainShutoffReason(C.VIR_DOMAIN_SHUTOFF_MIGRATED)
	DOMAIN_SHUTOFF_SAVED         = DomainShutoffReason(C.VIR_DOMAIN_SHUTOFF_SAVED)
	DOMAIN_SHUTOFF_FAILED        = DomainShutoffReason(C.VIR_DOMAIN_SHUTOFF_FAILED)
	DOMAIN_SHUTOFF_FROM_SNAPSHOT = DomainShutoffReason(C.VIR_DOMAIN_SHUTOFF_FROM_SNAPSHOT)
)

type DomainBlockCommitFlags int

const (
	DOMAIN_BLOCK_COMMIT_SHALLOW         = DomainBlockCommitFlags(C.VIR_DOMAIN_BLOCK_COMMIT_SHALLOW)
	DOMAIN_BLOCK_COMMIT_DELETE          = DomainBlockCommitFlags(C.VIR_DOMAIN_BLOCK_COMMIT_DELETE)
	DOMAIN_BLOCK_COMMIT_ACTIVE          = DomainBlockCommitFlags(C.VIR_DOMAIN_BLOCK_COMMIT_ACTIVE)
	DOMAIN_BLOCK_COMMIT_RELATIVE        = DomainBlockCommitFlags(C.VIR_DOMAIN_BLOCK_COMMIT_RELATIVE)
	DOMAIN_BLOCK_COMMIT_BANDWIDTH_BYTES = DomainBlockCommitFlags(C.VIR_DOMAIN_BLOCK_COMMIT_BANDWIDTH_BYTES)
)

type DomainBlockCopyFlags int

const (
	DOMAIN_BLOCK_COPY_SHALLOW   = DomainBlockCopyFlags(C.VIR_DOMAIN_BLOCK_COPY_SHALLOW)
	DOMAIN_BLOCK_COPY_REUSE_EXT = DomainBlockCopyFlags(C.VIR_DOMAIN_BLOCK_COPY_REUSE_EXT)
)

type DomainBlockRebaseFlags int

const (
	DOMAIN_BLOCK_REBASE_SHALLOW         = DomainBlockRebaseFlags(C.VIR_DOMAIN_BLOCK_REBASE_SHALLOW)
	DOMAIN_BLOCK_REBASE_REUSE_EXT       = DomainBlockRebaseFlags(C.VIR_DOMAIN_BLOCK_REBASE_REUSE_EXT)
	DOMAIN_BLOCK_REBASE_COPY_RAW        = DomainBlockRebaseFlags(C.VIR_DOMAIN_BLOCK_REBASE_COPY_RAW)
	DOMAIN_BLOCK_REBASE_COPY            = DomainBlockRebaseFlags(C.VIR_DOMAIN_BLOCK_REBASE_COPY)
	DOMAIN_BLOCK_REBASE_RELATIVE        = DomainBlockRebaseFlags(C.VIR_DOMAIN_BLOCK_REBASE_RELATIVE)
	DOMAIN_BLOCK_REBASE_COPY_DEV        = DomainBlockRebaseFlags(C.VIR_DOMAIN_BLOCK_REBASE_COPY_DEV)
	DOMAIN_BLOCK_REBASE_BANDWIDTH_BYTES = DomainBlockRebaseFlags(C.VIR_DOMAIN_BLOCK_REBASE_BANDWIDTH_BYTES)
)

type DomainBlockJobAbortFlags int

const (
	DOMAIN_BLOCK_JOB_ABORT_ASYNC = DomainBlockJobAbortFlags(C.VIR_DOMAIN_BLOCK_JOB_ABORT_ASYNC)
	DOMAIN_BLOCK_JOB_ABORT_PIVOT = DomainBlockJobAbortFlags(C.VIR_DOMAIN_BLOCK_JOB_ABORT_PIVOT)
)

type DomainBlockJobInfoFlags int

const (
	DOMAIN_BLOCK_JOB_INFO_BANDWIDTH_BYTES = DomainBlockJobInfoFlags(C.VIR_DOMAIN_BLOCK_JOB_INFO_BANDWIDTH_BYTES)
)

type DomainBlockJobSetSpeedFlags int

const (
	DOMAIN_BLOCK_JOB_SPEED_BANDWIDTH_BYTES = DomainBlockJobSetSpeedFlags(C.VIR_DOMAIN_BLOCK_JOB_SPEED_BANDWIDTH_BYTES)
)

type DomainBlockPullFlags int

const (
	DOMAIN_BLOCK_PULL_BANDWIDTH_BYTES = DomainBlockPullFlags(C.VIR_DOMAIN_BLOCK_PULL_BANDWIDTH_BYTES)
)

type DomainBlockResizeFlags int

const (
	DOMAIN_BLOCK_RESIZE_BYTES = DomainBlockResizeFlags(C.VIR_DOMAIN_BLOCK_RESIZE_BYTES)
)

type Domain struct {
	ptr C.virDomainPtr
}

type DomainChannelFlags int

const (
	DOMAIN_CHANNEL_FORCE = DomainChannelFlags(C.VIR_DOMAIN_CHANNEL_FORCE)
)

type DomainConsoleFlags int

const (
	DOMAIN_CONSOLE_FORCE = DomainConsoleFlags(C.VIR_DOMAIN_CONSOLE_FORCE)
	DOMAIN_CONSOLE_SAFE  = DomainConsoleFlags(C.VIR_DOMAIN_CONSOLE_SAFE)
)

type DomainCoreDumpFormat int

const (
	DOMAIN_CORE_DUMP_FORMAT_RAW          = DomainCoreDumpFormat(C.VIR_DOMAIN_CORE_DUMP_FORMAT_RAW)
	DOMAIN_CORE_DUMP_FORMAT_KDUMP_ZLIB   = DomainCoreDumpFormat(C.VIR_DOMAIN_CORE_DUMP_FORMAT_KDUMP_ZLIB)
	DOMAIN_CORE_DUMP_FORMAT_KDUMP_LZO    = DomainCoreDumpFormat(C.VIR_DOMAIN_CORE_DUMP_FORMAT_KDUMP_LZO)
	DOMAIN_CORE_DUMP_FORMAT_KDUMP_SNAPPY = DomainCoreDumpFormat(C.VIR_DOMAIN_CORE_DUMP_FORMAT_KDUMP_SNAPPY)
)

type DomainDefineFlags int

const (
	DOMAIN_DEFINE_VALIDATE = DomainDefineFlags(C.VIR_DOMAIN_DEFINE_VALIDATE)
)

type DomainJobType int

const (
	DOMAIN_JOB_NONE      = DomainJobType(C.VIR_DOMAIN_JOB_NONE)
	DOMAIN_JOB_BOUNDED   = DomainJobType(C.VIR_DOMAIN_JOB_BOUNDED)
	DOMAIN_JOB_UNBOUNDED = DomainJobType(C.VIR_DOMAIN_JOB_UNBOUNDED)
	DOMAIN_JOB_COMPLETED = DomainJobType(C.VIR_DOMAIN_JOB_COMPLETED)
	DOMAIN_JOB_FAILED    = DomainJobType(C.VIR_DOMAIN_JOB_FAILED)
	DOMAIN_JOB_CANCELLED = DomainJobType(C.VIR_DOMAIN_JOB_CANCELLED)
)

type DomainGetJobStatsFlags int

const (
	DOMAIN_JOB_STATS_COMPLETED = DomainGetJobStatsFlags(C.VIR_DOMAIN_JOB_STATS_COMPLETED)
)

type DomainNumatuneMemMode int

const (
	DOMAIN_NUMATUNE_MEM_STRICT     = DomainNumatuneMemMode(C.VIR_DOMAIN_NUMATUNE_MEM_STRICT)
	DOMAIN_NUMATUNE_MEM_PREFERRED  = DomainNumatuneMemMode(C.VIR_DOMAIN_NUMATUNE_MEM_PREFERRED)
	DOMAIN_NUMATUNE_MEM_INTERLEAVE = DomainNumatuneMemMode(C.VIR_DOMAIN_NUMATUNE_MEM_INTERLEAVE)
)

type DomainOpenGraphicsFlags int

const (
	DOMAIN_OPEN_GRAPHICS_SKIPAUTH = DomainOpenGraphicsFlags(C.VIR_DOMAIN_OPEN_GRAPHICS_SKIPAUTH)
)

type DomainSetUserPasswordFlags int

const (
	DOMAIN_PASSWORD_ENCRYPTED = DomainSetUserPasswordFlags(C.VIR_DOMAIN_PASSWORD_ENCRYPTED)
)

type DomainRebootFlagValues int

const (
	DOMAIN_REBOOT_DEFAULT        = DomainRebootFlagValues(C.VIR_DOMAIN_REBOOT_DEFAULT)
	DOMAIN_REBOOT_ACPI_POWER_BTN = DomainRebootFlagValues(C.VIR_DOMAIN_REBOOT_ACPI_POWER_BTN)
	DOMAIN_REBOOT_GUEST_AGENT    = DomainRebootFlagValues(C.VIR_DOMAIN_REBOOT_GUEST_AGENT)
	DOMAIN_REBOOT_INITCTL        = DomainRebootFlagValues(C.VIR_DOMAIN_REBOOT_INITCTL)
	DOMAIN_REBOOT_SIGNAL         = DomainRebootFlagValues(C.VIR_DOMAIN_REBOOT_SIGNAL)
	DOMAIN_REBOOT_PARAVIRT       = DomainRebootFlagValues(C.VIR_DOMAIN_REBOOT_PARAVIRT)
)

type DomainSaveRestoreFlags int

const (
	DOMAIN_SAVE_BYPASS_CACHE = DomainSaveRestoreFlags(C.VIR_DOMAIN_SAVE_BYPASS_CACHE)
	DOMAIN_SAVE_RUNNING      = DomainSaveRestoreFlags(C.VIR_DOMAIN_SAVE_RUNNING)
	DOMAIN_SAVE_PAUSED       = DomainSaveRestoreFlags(C.VIR_DOMAIN_SAVE_PAUSED)
)

type DomainSetTimeFlags int

const (
	DOMAIN_TIME_SYNC = DomainSetTimeFlags(C.VIR_DOMAIN_TIME_SYNC)
)

type DomainDiskErrorCode int

const (
	DOMAIN_DISK_ERROR_NONE     = DomainDiskErrorCode(C.VIR_DOMAIN_DISK_ERROR_NONE)
	DOMAIN_DISK_ERROR_UNSPEC   = DomainDiskErrorCode(C.VIR_DOMAIN_DISK_ERROR_UNSPEC)
	DOMAIN_DISK_ERROR_NO_SPACE = DomainDiskErrorCode(C.VIR_DOMAIN_DISK_ERROR_NO_SPACE)
)

type DomainStatsTypes int

const (
	DOMAIN_STATS_STATE     = DomainStatsTypes(C.VIR_DOMAIN_STATS_STATE)
	DOMAIN_STATS_CPU_TOTAL = DomainStatsTypes(C.VIR_DOMAIN_STATS_CPU_TOTAL)
	DOMAIN_STATS_BALLOON   = DomainStatsTypes(C.VIR_DOMAIN_STATS_BALLOON)
	DOMAIN_STATS_VCPU      = DomainStatsTypes(C.VIR_DOMAIN_STATS_VCPU)
	DOMAIN_STATS_INTERFACE = DomainStatsTypes(C.VIR_DOMAIN_STATS_INTERFACE)
	DOMAIN_STATS_BLOCK     = DomainStatsTypes(C.VIR_DOMAIN_STATS_BLOCK)
	DOMAIN_STATS_PERF      = DomainStatsTypes(C.VIR_DOMAIN_STATS_PERF)
)

type DomainCoreDumpFlags int

const (
	DUMP_CRASH        = DomainCoreDumpFlags(C.VIR_DUMP_CRASH)
	DUMP_LIVE         = DomainCoreDumpFlags(C.VIR_DUMP_LIVE)
	DUMP_BYPASS_CACHE = DomainCoreDumpFlags(C.VIR_DUMP_BYPASS_CACHE)
	DUMP_RESET        = DomainCoreDumpFlags(C.VIR_DUMP_RESET)
	DUMP_MEMORY_ONLY  = DomainCoreDumpFlags(C.VIR_DUMP_MEMORY_ONLY)
)

type DomainMemoryFlags int

const (
	MEMORY_VIRTUAL  = DomainMemoryFlags(C.VIR_MEMORY_VIRTUAL)
	MEMORY_PHYSICAL = DomainMemoryFlags(C.VIR_MEMORY_PHYSICAL)
)

type DomainMigrateFlags int

const (
	MIGRATE_LIVE              = DomainMigrateFlags(C.VIR_MIGRATE_LIVE)
	MIGRATE_PEER2PEER         = DomainMigrateFlags(C.VIR_MIGRATE_PEER2PEER)
	MIGRATE_TUNNELLED         = DomainMigrateFlags(C.VIR_MIGRATE_TUNNELLED)
	MIGRATE_PERSIST_DEST      = DomainMigrateFlags(C.VIR_MIGRATE_PERSIST_DEST)
	MIGRATE_UNDEFINE_SOURCE   = DomainMigrateFlags(C.VIR_MIGRATE_UNDEFINE_SOURCE)
	MIGRATE_PAUSED            = DomainMigrateFlags(C.VIR_MIGRATE_PAUSED)
	MIGRATE_NON_SHARED_DISK   = DomainMigrateFlags(C.VIR_MIGRATE_NON_SHARED_DISK)
	MIGRATE_NON_SHARED_INC    = DomainMigrateFlags(C.VIR_MIGRATE_NON_SHARED_INC)
	MIGRATE_CHANGE_PROTECTION = DomainMigrateFlags(C.VIR_MIGRATE_CHANGE_PROTECTION)
	MIGRATE_UNSAFE            = DomainMigrateFlags(C.VIR_MIGRATE_UNSAFE)
	MIGRATE_OFFLINE           = DomainMigrateFlags(C.VIR_MIGRATE_OFFLINE)
	MIGRATE_COMPRESSED        = DomainMigrateFlags(C.VIR_MIGRATE_COMPRESSED)
	MIGRATE_ABORT_ON_ERROR    = DomainMigrateFlags(C.VIR_MIGRATE_ABORT_ON_ERROR)
	MIGRATE_AUTO_CONVERGE     = DomainMigrateFlags(C.VIR_MIGRATE_AUTO_CONVERGE)
	MIGRATE_RDMA_PIN_ALL      = DomainMigrateFlags(C.VIR_MIGRATE_RDMA_PIN_ALL)
	MIGRATE_POSTCOPY          = DomainMigrateFlags(C.VIR_MIGRATE_POSTCOPY)
	MIGRATE_TLS               = DomainMigrateFlags(C.VIR_MIGRATE_TLS)
)

type VcpuState int

const (
	VCPU_OFFLINE = VcpuState(C.VIR_VCPU_OFFLINE)
	VCPU_RUNNING = VcpuState(C.VIR_VCPU_RUNNING)
	VCPU_BLOCKED = VcpuState(C.VIR_VCPU_BLOCKED)
)

type DomainJobOperationType int

const (
	DOMAIN_JOB_OPERATION_UNKNOWN         = DomainJobOperationType(C.VIR_DOMAIN_JOB_OPERATION_UNKNOWN)
	DOMAIN_JOB_OPERATION_START           = DomainJobOperationType(C.VIR_DOMAIN_JOB_OPERATION_START)
	DOMAIN_JOB_OPERATION_SAVE            = DomainJobOperationType(C.VIR_DOMAIN_JOB_OPERATION_SAVE)
	DOMAIN_JOB_OPERATION_RESTORE         = DomainJobOperationType(C.VIR_DOMAIN_JOB_OPERATION_RESTORE)
	DOMAIN_JOB_OPERATION_MIGRATION_IN    = DomainJobOperationType(C.VIR_DOMAIN_JOB_OPERATION_MIGRATION_IN)
	DOMAIN_JOB_OPERATION_MIGRATION_OUT   = DomainJobOperationType(C.VIR_DOMAIN_JOB_OPERATION_MIGRATION_OUT)
	DOMAIN_JOB_OPERATION_SNAPSHOT        = DomainJobOperationType(C.VIR_DOMAIN_JOB_OPERATION_SNAPSHOT)
	DOMAIN_JOB_OPERATION_SNAPSHOT_REVERT = DomainJobOperationType(C.VIR_DOMAIN_JOB_OPERATION_SNAPSHOT_REVERT)
	DOMAIN_JOB_OPERATION_DUMP            = DomainJobOperationType(C.VIR_DOMAIN_JOB_OPERATION_DUMP)
)

type DomainBlockInfo struct {
	Capacity   uint64
	Allocation uint64
	Physical   uint64
}

type DomainInfo struct {
	State     DomainState
	MaxMem    uint64
	Memory    uint64
	NrVirtCpu uint
	CpuTime   uint64
}

type DomainMemoryStat struct {
	Tag int32
	Val uint64
}

type DomainVcpuInfo struct {
	Number  uint32
	State   int32
	CpuTime uint64
	Cpu     int32
	CpuMap  []bool
}

func (d *Domain) Free() error {
	ret := C.virDomainFree(d.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Domain) Ref() error {
	ret := C.virDomainRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) Create() error {
	result := C.virDomainCreate(d.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) CreateWithFlags(flags DomainCreateFlags) error {
	result := C.virDomainCreateWithFlags(d.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) CreateWithFiles(files []os.File, flags DomainCreateFlags) error {
	cfiles := make([]C.int, len(files))
	for i := 0; i < len(files); i++ {
		cfiles[i] = C.int(files[i].Fd())
	}
	result := C.virDomainCreateWithFiles(d.ptr, C.uint(len(files)), &cfiles[0], C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) Destroy() error {
	result := C.virDomainDestroy(d.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) Shutdown() error {
	result := C.virDomainShutdown(d.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) Reboot(flags DomainRebootFlagValues) error {
	result := C.virDomainReboot(d.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) IsActive() (bool, error) {
	result := C.virDomainIsActive(d.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (d *Domain) IsPersistent() (bool, error) {
	result := C.virDomainIsPersistent(d.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (d *Domain) IsUpdated() (bool, error) {
	result := C.virDomainIsUpdated(d.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (d *Domain) SetAutostart(autostart bool) error {
	var cAutostart C.int
	switch autostart {
	case true:
		cAutostart = 1
	default:
		cAutostart = 0
	}
	result := C.virDomainSetAutostart(d.ptr, cAutostart)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) GetAutostart() (bool, error) {
	var out C.int
	result := C.virDomainGetAutostart(d.ptr, (*C.int)(unsafe.Pointer(&out)))
	if result == -1 {
		return false, GetLastError()
	}
	switch out {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (d *Domain) GetBlockInfo(disk string, flag uint) (*DomainBlockInfo, error) {
	var cinfo C.virDomainBlockInfo
	cDisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cDisk))
	result := C.virDomainGetBlockInfo(d.ptr, cDisk, &cinfo, C.uint(flag))
	if result == -1 {
		return nil, GetLastError()
	}

	return &DomainBlockInfo{
		Capacity:   uint64(cinfo.capacity),
		Allocation: uint64(cinfo.allocation),
		Physical:   uint64(cinfo.physical),
	}, nil
}

func (d *Domain) GetName() (string, error) {
	name := C.virDomainGetName(d.ptr)
	if name == nil {
		return "", GetLastError()
	}
	return C.GoString(name), nil
}

func (d *Domain) GetState() (DomainState, int, error) {
	var cState C.int
	var cReason C.int
	result := C.virDomainGetState(d.ptr,
		(*C.int)(unsafe.Pointer(&cState)),
		(*C.int)(unsafe.Pointer(&cReason)),
		0)
	if int(result) == -1 {
		return 0, 0, GetLastError()
	}
	return DomainState(cState), int(cReason), nil
}

func (d *Domain) GetID() (uint, error) {
	id := uint(C.virDomainGetID(d.ptr))
	if id == ^uint(0) {
		return id, GetLastError()
	}
	return id, nil
}

func (d *Domain) GetUUID() ([]byte, error) {
	var cUuid [C.VIR_UUID_BUFLEN](byte)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virDomainGetUUID(d.ptr, (*C.uchar)(cuidPtr))
	if result != 0 {
		return []byte{}, GetLastError()
	}
	return C.GoBytes(cuidPtr, C.VIR_UUID_BUFLEN), nil
}

func (d *Domain) GetUUIDString() (string, error) {
	var cUuid [C.VIR_UUID_STRING_BUFLEN](C.char)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virDomainGetUUIDString(d.ptr, (*C.char)(cuidPtr))
	if result != 0 {
		return "", GetLastError()
	}
	return C.GoString((*C.char)(cuidPtr)), nil
}

func (d *Domain) GetInfo() (*DomainInfo, error) {
	var cinfo C.virDomainInfo
	result := C.virDomainGetInfo(d.ptr, &cinfo)
	if result == -1 {
		return nil, GetLastError()
	}
	return &DomainInfo{
		State:     DomainState(cinfo.state),
		MaxMem:    uint64(cinfo.maxMem),
		Memory:    uint64(cinfo.memory),
		NrVirtCpu: uint(cinfo.nrVirtCpu),
		CpuTime:   uint64(cinfo.cpuTime),
	}, nil
}

func (d *Domain) GetXMLDesc(flags DomainXMLFlags) (string, error) {
	result := C.virDomainGetXMLDesc(d.ptr, C.uint(flags))
	if result == nil {
		return "", GetLastError()
	}
	xml := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return xml, nil
}

type DomainCPUStats struct {
	CpuTimeSet    bool
	CpuTime       uint64
	UserTimeSet   bool
	UserTime      uint64
	SystemTimeSet bool
	SystemTime    uint64
	VcpuTimeSet   bool
	VcpuTime      uint64
}

func getCPUStatsFieldInfo(params *DomainCPUStats) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_CPU_STATS_CPUTIME: typedParamsFieldInfo{
			set: &params.CpuTimeSet,
			ul:  &params.CpuTime,
		},
		C.VIR_DOMAIN_CPU_STATS_USERTIME: typedParamsFieldInfo{
			set: &params.UserTimeSet,
			ul:  &params.UserTime,
		},
		C.VIR_DOMAIN_CPU_STATS_SYSTEMTIME: typedParamsFieldInfo{
			set: &params.SystemTimeSet,
			ul:  &params.SystemTime,
		},
		C.VIR_DOMAIN_CPU_STATS_VCPUTIME: typedParamsFieldInfo{
			set: &params.VcpuTimeSet,
			ul:  &params.VcpuTime,
		},
	}
}

func (d *Domain) GetCPUStats(startCpu int, nCpus uint, flags uint32) ([]DomainCPUStats, error) {
	if nCpus == 0 {
		if startCpu == -1 {
			nCpus = 1
		} else {
			ret := C.virDomainGetCPUStats(d.ptr, nil, 0, 0, 0, 0)
			if ret == -1 {
				return []DomainCPUStats{}, GetLastError()
			}
			nCpus = uint(ret)
		}
	}

	ret := C.virDomainGetCPUStats(d.ptr, nil, 0, C.int(startCpu), C.uint(nCpus), 0)
	if ret == -1 {
		return []DomainCPUStats{}, GetLastError()
	}
	nparams := uint(ret)

	var cparams []C.virTypedParameter
	var nallocparams uint
	if startCpu == -1 {
		nallocparams = nparams
	} else {
		nallocparams = nparams * nCpus
	}
	cparams = make([]C.virTypedParameter, nallocparams)
	ret = C.virDomainGetCPUStats(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), C.uint(nparams), C.int(startCpu), C.uint(nCpus), C.uint(flags))
	if ret == -1 {
		return []DomainCPUStats{}, GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), C.int(nallocparams))

	stats := make([]DomainCPUStats, nCpus)
	for i := 0; i < int(nCpus); i++ {
		offset := i * int(nparams)
		info := getCPUStatsFieldInfo(&stats[i])
		cparamscpu := cparams[offset : offset+int(ret)]
		_, err := typedParamsUnpack(cparamscpu, info)
		if err != nil {
			return []DomainCPUStats{}, err
		}
	}
	return stats, nil
}

type DomainInterfaceParameters struct {
	BandwidthInAverageSet  bool
	BandwidthInAverage     uint
	BandwidthInPeakSet     bool
	BandwidthInPeak        uint
	BandwidthInBurstSet    bool
	BandwidthInBurst       uint
	BandwidthInFloorSet    bool
	BandwidthInFloor       uint
	BandwidthOutAverageSet bool
	BandwidthOutAverage    uint
	BandwidthOutPeakSet    bool
	BandwidthOutPeak       uint
	BandwidthOutBurstSet   bool
	BandwidthOutBurst      uint
}

func getInterfaceParameterFieldInfo(params *DomainInterfaceParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_BANDWIDTH_IN_AVERAGE: typedParamsFieldInfo{
			set: &params.BandwidthInAverageSet,
			ui:  &params.BandwidthInAverage,
		},
		C.VIR_DOMAIN_BANDWIDTH_IN_PEAK: typedParamsFieldInfo{
			set: &params.BandwidthInPeakSet,
			ui:  &params.BandwidthInPeak,
		},
		C.VIR_DOMAIN_BANDWIDTH_IN_BURST: typedParamsFieldInfo{
			set: &params.BandwidthInBurstSet,
			ui:  &params.BandwidthInBurst,
		},
		C.VIR_DOMAIN_BANDWIDTH_IN_FLOOR: typedParamsFieldInfo{
			set: &params.BandwidthInFloorSet,
			ui:  &params.BandwidthInFloor,
		},
		C.VIR_DOMAIN_BANDWIDTH_OUT_AVERAGE: typedParamsFieldInfo{
			set: &params.BandwidthOutAverageSet,
			ui:  &params.BandwidthOutAverage,
		},
		C.VIR_DOMAIN_BANDWIDTH_OUT_PEAK: typedParamsFieldInfo{
			set: &params.BandwidthOutPeakSet,
			ui:  &params.BandwidthOutPeak,
		},
		C.VIR_DOMAIN_BANDWIDTH_OUT_BURST: typedParamsFieldInfo{
			set: &params.BandwidthOutBurstSet,
			ui:  &params.BandwidthOutBurst,
		},
	}
}

func (d *Domain) GetInterfaceParameters(device string, flags DomainModificationImpact) (*DomainInterfaceParameters, error) {
	params := &DomainInterfaceParameters{}
	info := getInterfaceParameterFieldInfo(params)

	var nparams C.int

	cdevice := C.CString(device)
	defer C.free(unsafe.Pointer(cdevice))
	ret := C.virDomainGetInterfaceParameters(d.ptr, cdevice, nil, &nparams, C.uint(0))
	if ret == -1 {
		return nil, GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetInterfaceParameters(d.ptr, cdevice, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	_, err := typedParamsUnpack(cparams, info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Domain) SetInterfaceParameters(device string, params *DomainInterfaceParameters, flags DomainModificationImpact) error {
	info := getInterfaceParameterFieldInfo(params)

	var nparams C.int

	cdevice := C.CString(device)
	defer C.free(unsafe.Pointer(cdevice))
	ret := C.virDomainGetInterfaceParameters(d.ptr, cdevice, nil, &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetInterfaceParameters(d.ptr, cdevice, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	err := typedParamsPack(cparams, info)
	if err != nil {
		return err
	}

	ret = C.virDomainSetInterfaceParameters(d.ptr, cdevice, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams, C.uint(flags))

	return nil
}

func (d *Domain) GetMetadata(tipus DomainMetadataType, uri string, flags DomainModificationImpact) (string, error) {
	var cUri *C.char
	if uri != "" {
		cUri = C.CString(uri)
		defer C.free(unsafe.Pointer(cUri))
	}

	result := C.virDomainGetMetadata(d.ptr, C.int(tipus), cUri, C.uint(flags))
	if result == nil {
		return "", GetLastError()

	}
	defer C.free(unsafe.Pointer(result))
	return C.GoString(result), nil
}

func (d *Domain) SetMetadata(metaDataType DomainMetadataType, metaDataCont, uriKey, uri string, flags DomainModificationImpact) error {
	var cMetaDataCont *C.char
	var cUriKey *C.char
	var cUri *C.char

	if metaDataCont != "" {
		cMetaDataCont = C.CString(metaDataCont)
		defer C.free(unsafe.Pointer(cMetaDataCont))
	}

	if metaDataType == DOMAIN_METADATA_ELEMENT {
		if uriKey != "" {
			cUriKey = C.CString(uriKey)
			defer C.free(unsafe.Pointer(cUriKey))
		}
		cUri = C.CString(uri)
		defer C.free(unsafe.Pointer(cUri))
	}
	result := C.virDomainSetMetadata(d.ptr, C.int(metaDataType), cMetaDataCont, cUriKey, cUri, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) Undefine() error {
	result := C.virDomainUndefine(d.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) UndefineFlags(flags DomainUndefineFlagsValues) error {
	result := C.virDomainUndefineFlags(d.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) SetMaxMemory(memory uint) error {
	result := C.virDomainSetMaxMemory(d.ptr, C.ulong(memory))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) SetMemory(memory uint64) error {
	result := C.virDomainSetMemory(d.ptr, C.ulong(memory))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) SetMemoryFlags(memory uint64, flags DomainMemoryModFlags) error {
	result := C.virDomainSetMemoryFlags(d.ptr, C.ulong(memory), C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) SetMemoryStatsPeriod(period int, flags DomainMemoryModFlags) error {
	result := C.virDomainSetMemoryStatsPeriod(d.ptr, C.int(period), C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) SetVcpus(vcpu uint) error {
	result := C.virDomainSetVcpus(d.ptr, C.uint(vcpu))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) SetVcpusFlags(vcpu uint, flags DomainVcpuFlags) error {
	result := C.virDomainSetVcpusFlags(d.ptr, C.uint(vcpu), C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) Suspend() error {
	result := C.virDomainSuspend(d.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) Resume() error {
	result := C.virDomainResume(d.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) AbortJob() error {
	result := C.virDomainAbortJob(d.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) DestroyFlags(flags DomainDestroyFlags) error {
	result := C.virDomainDestroyFlags(d.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) ShutdownFlags(flags DomainShutdownFlags) error {
	result := C.virDomainShutdownFlags(d.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) AttachDevice(xml string) error {
	cXml := C.CString(xml)
	defer C.free(unsafe.Pointer(cXml))
	result := C.virDomainAttachDevice(d.ptr, cXml)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) AttachDeviceFlags(xml string, flags DomainDeviceModifyFlags) error {
	cXml := C.CString(xml)
	defer C.free(unsafe.Pointer(cXml))
	result := C.virDomainAttachDeviceFlags(d.ptr, cXml, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) DetachDevice(xml string) error {
	cXml := C.CString(xml)
	defer C.free(unsafe.Pointer(cXml))
	result := C.virDomainDetachDevice(d.ptr, cXml)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) DetachDeviceFlags(xml string, flags DomainDeviceModifyFlags) error {
	cXml := C.CString(xml)
	defer C.free(unsafe.Pointer(cXml))
	result := C.virDomainDetachDeviceFlags(d.ptr, cXml, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) UpdateDeviceFlags(xml string, flags DomainDeviceModifyFlags) error {
	cXml := C.CString(xml)
	defer C.free(unsafe.Pointer(cXml))
	result := C.virDomainUpdateDeviceFlags(d.ptr, cXml, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) Screenshot(stream *Stream, screen, flags uint32) (string, error) {
	cType := C.virDomainScreenshot(d.ptr, stream.ptr, C.uint(screen), C.uint(flags))
	if cType == nil {
		return "", GetLastError()
	}
	defer C.free(unsafe.Pointer(cType))

	mimeType := C.GoString(cType)
	return mimeType, nil
}

func (d *Domain) SendKey(codeset, holdtime uint, keycodes []uint, flags uint32) error {
	result := C.virDomainSendKey(d.ptr, C.uint(codeset), C.uint(holdtime), (*C.uint)(unsafe.Pointer(&keycodes[0])), C.int(len(keycodes)), C.uint(flags))
	if result == -1 {
		return GetLastError()
	}

	return nil
}

type DomainBlockStats struct {
	RdBytesSet         bool
	RdBytes            int64
	RdReqSet           bool
	RdReq              int64
	RdTotalTimesSet    bool
	RdTotalTimes       int64
	WrBytesSet         bool
	WrBytes            int64
	WrReqSet           bool
	WrReq              int64
	WrTotalTimesSet    bool
	WrTotalTimes       int64
	FlushReqSet        bool
	FlushReq           int64
	FlushTotalTimesSet bool
	FlushTotalTimes    int64
	ErrsSet            bool
	Errs               int64
}

func getBlockStatsFieldInfo(params *DomainBlockStats) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_BLOCK_STATS_READ_BYTES: typedParamsFieldInfo{
			set: &params.RdBytesSet,
			l:   &params.RdBytes,
		},
		C.VIR_DOMAIN_BLOCK_STATS_READ_REQ: typedParamsFieldInfo{
			set: &params.RdReqSet,
			l:   &params.RdReq,
		},
		C.VIR_DOMAIN_BLOCK_STATS_READ_TOTAL_TIMES: typedParamsFieldInfo{
			set: &params.RdTotalTimesSet,
			l:   &params.RdTotalTimes,
		},
		C.VIR_DOMAIN_BLOCK_STATS_WRITE_BYTES: typedParamsFieldInfo{
			set: &params.WrBytesSet,
			l:   &params.WrBytes,
		},
		C.VIR_DOMAIN_BLOCK_STATS_WRITE_REQ: typedParamsFieldInfo{
			set: &params.WrReqSet,
			l:   &params.WrReq,
		},
		C.VIR_DOMAIN_BLOCK_STATS_WRITE_TOTAL_TIMES: typedParamsFieldInfo{
			set: &params.WrTotalTimesSet,
			l:   &params.WrTotalTimes,
		},
		C.VIR_DOMAIN_BLOCK_STATS_FLUSH_REQ: typedParamsFieldInfo{
			set: &params.FlushReqSet,
			l:   &params.FlushReq,
		},
		C.VIR_DOMAIN_BLOCK_STATS_FLUSH_TOTAL_TIMES: typedParamsFieldInfo{
			set: &params.FlushTotalTimesSet,
			l:   &params.FlushTotalTimes,
		},
		C.VIR_DOMAIN_BLOCK_STATS_ERRS: typedParamsFieldInfo{
			set: &params.ErrsSet,
			l:   &params.Errs,
		},
	}
}

func (d *Domain) BlockStatsFlags(disk string, flags uint32) (*DomainBlockStats, error) {
	params := &DomainBlockStats{}
	info := getBlockStatsFieldInfo(params)

	var nparams C.int

	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))
	ret := C.virDomainBlockStatsFlags(d.ptr, cdisk, nil, &nparams, C.uint(0))
	if ret == -1 {
		return nil, GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainBlockStatsFlags(d.ptr, cdisk, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	_, err := typedParamsUnpack(cparams, info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Domain) BlockStats(path string) (*DomainBlockStats, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	size := C.size_t(unsafe.Sizeof(C.struct__virDomainBlockStats{}))

	cStats := (C.virDomainBlockStatsPtr)(C.malloc(size))
	defer C.free(unsafe.Pointer(cStats))

	result := C.virDomainBlockStats(d.ptr, cPath, (C.virDomainBlockStatsPtr)(cStats), size)

	if result != 0 {
		return nil, GetLastError()
	}
	return &DomainBlockStats{
		WrReqSet:   true,
		WrReq:      int64(cStats.wr_req),
		RdReqSet:   true,
		RdReq:      int64(cStats.rd_req),
		RdBytesSet: true,
		RdBytes:    int64(cStats.rd_bytes),
		WrBytesSet: true,
		WrBytes:    int64(cStats.wr_bytes),
	}, nil
}

type DomainInterfaceStats struct {
	RxBytesSet   bool
	RxBytes      int64
	RxPacketsSet bool
	RxPackets    int64
	RxErrsSet    bool
	RxErrs       int64
	RxDropSet    bool
	RxDrop       int64
	TxBytesSet   bool
	TxBytes      int64
	TxPacketsSet bool
	TxPackets    int64
	TxErrsSet    bool
	TxErrs       int64
	TxDropSet    bool
	TxDrop       int64
}

func (d *Domain) InterfaceStats(path string) (*DomainInterfaceStats, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	size := C.size_t(unsafe.Sizeof(C.struct__virDomainInterfaceStats{}))

	cStats := (C.virDomainInterfaceStatsPtr)(C.malloc(size))
	defer C.free(unsafe.Pointer(cStats))

	result := C.virDomainInterfaceStats(d.ptr, cPath, (C.virDomainInterfaceStatsPtr)(cStats), size)

	if result != 0 {
		return nil, GetLastError()
	}
	return &DomainInterfaceStats{
		RxBytesSet:   true,
		RxBytes:      int64(cStats.rx_bytes),
		RxPacketsSet: true,
		RxPackets:    int64(cStats.rx_packets),
		RxErrsSet:    true,
		RxErrs:       int64(cStats.rx_errs),
		RxDropSet:    true,
		RxDrop:       int64(cStats.rx_drop),
		TxBytesSet:   true,
		TxBytes:      int64(cStats.tx_bytes),
		TxPacketsSet: true,
		TxPackets:    int64(cStats.tx_packets),
		TxErrsSet:    true,
		TxErrs:       int64(cStats.tx_errs),
		TxDropSet:    true,
		TxDrop:       int64(cStats.tx_drop),
	}, nil
}

func (d *Domain) MemoryStats(nrStats uint32, flags uint32) ([]DomainMemoryStat, error) {
	ptr := make([]C.virDomainMemoryStatStruct, nrStats)

	result := C.virDomainMemoryStats(
		d.ptr, (C.virDomainMemoryStatPtr)(unsafe.Pointer(&ptr[0])),
		C.uint(nrStats), C.uint(flags))

	if result == -1 {
		return []DomainMemoryStat{}, GetLastError()
	}

	out := make([]DomainMemoryStat, result)
	for i := 0; i < int(result); i++ {
		out = append(out, DomainMemoryStat{
			Tag: int32(ptr[i].tag),
			Val: uint64(ptr[i].val),
		})
	}
	return out, nil
}

func (d *Domain) GetVcpus() ([]DomainVcpuInfo, error) {
	var cnodeinfo C.virNodeInfo
	ret := C.virNodeGetInfo(C.virDomainGetConnect(d.ptr), &cnodeinfo)
	if ret == -1 {
		return []DomainVcpuInfo{}, GetLastError()
	}

	var cdominfo C.virDomainInfo
	ret = C.virDomainGetInfo(d.ptr, &cdominfo)
	if ret == -1 {
		return []DomainVcpuInfo{}, GetLastError()
	}

	nvcpus := int(cdominfo.nrVirtCpu)
	npcpus := int(cnodeinfo.nodes * cnodeinfo.sockets * cnodeinfo.cores * cnodeinfo.threads)
	maplen := ((npcpus + 7) / 8)
	ccpumaps := make([]C.uchar, maplen*nvcpus)
	cinfo := make([]C.virVcpuInfo, nvcpus)

	ret = C.virDomainGetVcpus(d.ptr, &cinfo[0], C.int(nvcpus), &ccpumaps[0], C.int(maplen))
	if ret == -1 {
		return []DomainVcpuInfo{}, GetLastError()
	}

	info := make([]DomainVcpuInfo, int(ret))
	for i := 0; i < int(ret); i++ {
		affinity := make([]bool, npcpus)
		for j := 0; j < npcpus; j++ {
			byte := (i * maplen) + (j / 8)
			bit := j % 8

			affinity[j] = (ccpumaps[byte] & (1 << uint(bit))) != 0
		}

		info[i] = DomainVcpuInfo{
			Number:  uint32(cinfo[i].number),
			State:   int32(cinfo[i].state),
			CpuTime: uint64(cinfo[i].cpuTime),
			Cpu:     int32(cinfo[i].cpu),
			CpuMap:  affinity,
		}
	}

	return info, nil
}

func (d *Domain) GetVcpusFlags(flags DomainVcpuFlags) (int32, error) {
	result := C.virDomainGetVcpusFlags(d.ptr, C.uint(flags))
	if result == -1 {
		return 0, GetLastError()
	}
	return int32(result), nil
}

func (d *Domain) PinVcpu(vcpu uint, cpuMap []bool) error {
	maplen := (len(cpuMap) + 7) / 8
	ccpumap := make([]C.uchar, maplen)
	for i := 0; i < len(cpuMap); i++ {
		if cpuMap[i] {
			byte := i / 8
			bit := i % 8
			ccpumap[byte] |= (1 << uint(bit))
		}
	}

	result := C.virDomainPinVcpu(d.ptr, C.uint(vcpu), &ccpumap[0], C.int(maplen))

	if result == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) PinVcpuFlags(vcpu uint, cpuMap []bool, flags DomainModificationImpact) error {
	maplen := (len(cpuMap) + 7) / 8
	ccpumap := make([]C.uchar, maplen)
	for i := 0; i < len(cpuMap); i++ {
		if cpuMap[i] {
			byte := i / 8
			bit := i % 8
			ccpumap[byte] |= (1 << uint(bit))
		}
	}

	result := C.virDomainPinVcpuFlags(d.ptr, C.uint(vcpu), &ccpumap[0], C.int(maplen), C.uint(flags))

	if result == -1 {
		return GetLastError()
	}

	return nil
}

type DomainIPAddress struct {
	Type   int
	Addr   string
	Prefix uint
}

type DomainInterface struct {
	Name   string
	Hwaddr string
	Addrs  []DomainIPAddress
}

func (d *Domain) ListAllInterfaceAddresses(src DomainInterfaceAddressesSource) ([]DomainInterface, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002014 {
		return []DomainInterface{}, GetNotImplementedError("virDomainInterfaceAddresses")
	}

	var cList *C.virDomainInterfacePtr
	numIfaces := int(C.virDomainInterfaceAddressesCompat(d.ptr, (**C.virDomainInterfacePtr)(&cList), C.uint(src), 0))
	if numIfaces == -1 {
		return nil, GetLastError()
	}

	ifaces := make([]DomainInterface, numIfaces)

	for i := 0; i < numIfaces; i++ {
		var ciface *C.virDomainInterface
		ciface = *(**C.virDomainInterface)(unsafe.Pointer(uintptr(unsafe.Pointer(cList)) + (unsafe.Sizeof(ciface) * uintptr(i))))

		ifaces[i].Name = C.GoString(ciface.name)
		ifaces[i].Hwaddr = C.GoString(ciface.hwaddr)

		numAddr := int(ciface.naddrs)

		ifaces[i].Addrs = make([]DomainIPAddress, numAddr)

		for k := 0; k < numAddr; k++ {
			var caddr *C.virDomainIPAddress
			caddr = (*C.virDomainIPAddress)(unsafe.Pointer(uintptr(unsafe.Pointer(ciface.addrs)) + (unsafe.Sizeof(*caddr) * uintptr(k))))
			ifaces[i].Addrs[k] = DomainIPAddress{}
			ifaces[i].Addrs[k].Type = int(caddr._type)
			ifaces[i].Addrs[k].Addr = C.GoString(caddr.addr)
			ifaces[i].Addrs[k].Prefix = uint(caddr.prefix)

		}
		C.virDomainInterfaceFreeCompat(ciface)
	}
	C.free(unsafe.Pointer(cList))
	return ifaces, nil
}

func (d *Domain) SnapshotCurrent(flags uint32) (*DomainSnapshot, error) {
	result := C.virDomainSnapshotCurrent(d.ptr, C.uint(flags))
	if result == nil {
		return nil, GetLastError()
	}
	return &DomainSnapshot{ptr: result}, nil

}

func (d *Domain) SnapshotNum(flags DomainSnapshotListFlags) (int, error) {
	result := int(C.virDomainSnapshotNum(d.ptr, C.uint(flags)))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (d *Domain) SnapshotLookupByName(name string, flags uint32) (*DomainSnapshot, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virDomainSnapshotLookupByName(d.ptr, cName, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &DomainSnapshot{ptr: ptr}, nil
}

func (d *Domain) SnapshotListNames(flags DomainSnapshotListFlags) ([]string, error) {
	const maxNames = 1024
	var names [maxNames](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numNames := C.virDomainSnapshotListNames(
		d.ptr,
		(**C.char)(namesPtr),
		maxNames, C.uint(flags))
	if numNames == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numNames)
	for k := 0; k < int(numNames); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (d *Domain) ListAllSnapshots(flags DomainSnapshotListFlags) ([]DomainSnapshot, error) {
	var cList *C.virDomainSnapshotPtr
	numVols := C.virDomainListAllSnapshots(d.ptr, (**C.virDomainSnapshotPtr)(&cList), C.uint(flags))
	if numVols == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numVols),
		Cap:  int(numVols),
	}
	var pools []DomainSnapshot
	slice := *(*[]C.virDomainSnapshotPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		pools = append(pools, DomainSnapshot{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return pools, nil
}

func (d *Domain) BlockCommit(disk string, base string, top string, bandwidth uint64, flags DomainBlockCommitFlags) error {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))
	var cbase *C.char
	if base != "" {
		cbase = C.CString(base)
		defer C.free(unsafe.Pointer(cbase))
	}
	var ctop *C.char
	if top != "" {
		ctop = C.CString(top)
		defer C.free(unsafe.Pointer(ctop))
	}
	ret := C.virDomainBlockCommit(d.ptr, cdisk, cbase, ctop, C.ulong(bandwidth), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

type DomainBlockCopyParameters struct {
	BandwidthSet   bool
	Bandwidth      uint64
	GranularitySet bool
	Granularity    uint
	BufSizeSet     bool
	BufSize        uint64
}

func getBlockCopyParameterFieldInfo(params *DomainBlockCopyParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_BLOCK_COPY_BANDWIDTH: typedParamsFieldInfo{
			set: &params.BandwidthSet,
			ul:  &params.Bandwidth,
		},
		C.VIR_DOMAIN_BLOCK_COPY_GRANULARITY: typedParamsFieldInfo{
			set: &params.GranularitySet,
			ui:  &params.Granularity,
		},
		C.VIR_DOMAIN_BLOCK_COPY_BUF_SIZE: typedParamsFieldInfo{
			set: &params.BufSizeSet,
			ul:  &params.BufSize,
		},
	}
}

func (d *Domain) BlockCopy(disk string, destxml string, params *DomainBlockCopyParameters, flags DomainBlockCopyFlags) error {
	if C.LIBVIR_VERSION_NUMBER < 1002008 {
		return GetNotImplementedError("virDomainBlockCopy")
	}
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))
	cdestxml := C.CString(destxml)
	defer C.free(unsafe.Pointer(cdestxml))

	info := getBlockCopyParameterFieldInfo(params)

	cparams, err := typedParamsPackNew(info)
	if err != nil {
		return err
	}
	nparams := len(*cparams)

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&(*cparams)[0])), C.int(nparams))

	ret := C.virDomainBlockCopyCompat(d.ptr, cdisk, cdestxml, (*C.virTypedParameter)(unsafe.Pointer(&(*cparams)[0])), C.int(nparams), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) BlockJobAbort(disk string, flags DomainBlockJobAbortFlags) error {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))
	ret := C.virDomainBlockJobAbort(d.ptr, cdisk, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) BlockJobSetSpeed(disk string, bandwidth uint64, flags DomainBlockJobSetSpeedFlags) error {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))
	ret := C.virDomainBlockJobSetSpeed(d.ptr, cdisk, C.ulong(bandwidth), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) BlockPull(disk string, bandwidth uint64, flags DomainBlockPullFlags) error {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))
	ret := C.virDomainBlockPull(d.ptr, cdisk, C.ulong(bandwidth), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) BlockRebase(disk string, base string, bandwidth uint64, flags DomainBlockRebaseFlags) error {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))
	var cbase *C.char
	if base != "" {
		cbase := C.CString(base)
		defer C.free(unsafe.Pointer(cbase))
	}
	ret := C.virDomainBlockRebase(d.ptr, cdisk, cbase, C.ulong(bandwidth), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) BlockResize(disk string, size uint64, flags DomainBlockResizeFlags) error {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))
	ret := C.virDomainBlockResize(d.ptr, cdisk, C.ulonglong(size), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) BlockPeek(disk string, offset uint64, size uint64, flags uint32) ([]byte, error) {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))
	data := make([]byte, size)
	ret := C.virDomainBlockPeek(d.ptr, cdisk, C.ulonglong(offset), C.size_t(size),
		unsafe.Pointer(&data[0]), C.uint(flags))
	if ret == -1 {
		return []byte{}, GetLastError()
	}

	return data, nil
}

func (d *Domain) MemoryPeek(start uint64, size uint64, flags DomainMemoryFlags) ([]byte, error) {
	data := make([]byte, size)
	ret := C.virDomainMemoryPeek(d.ptr, C.ulonglong(start), C.size_t(size),
		unsafe.Pointer(&data[0]), C.uint(flags))
	if ret == -1 {
		return []byte{}, GetLastError()
	}

	return data, nil
}

func (d *Domain) Migrate(dconn *Connect, flags DomainMigrateFlags, dname string, uri string, bandwidth uint64) (*Domain, error) {
	var cdname *C.char
	if dname != "" {
		cdname = C.CString(dname)
		defer C.free(unsafe.Pointer(cdname))
	}
	var curi *C.char
	if uri != "" {
		curi = C.CString(uri)
		defer C.free(unsafe.Pointer(curi))
	}

	ret := C.virDomainMigrate(d.ptr, dconn.ptr, C.ulong(flags), cdname, curi, C.ulong(bandwidth))
	if ret == nil {
		return nil, GetLastError()
	}

	return &Domain{
		ptr: ret,
	}, nil
}

func (d *Domain) Migrate2(dconn *Connect, dxml string, flags DomainMigrateFlags, dname string, uri string, bandwidth uint64) (*Domain, error) {
	var cdxml *C.char
	if dxml != "" {
		cdxml = C.CString(dxml)
		defer C.free(unsafe.Pointer(cdxml))
	}
	var cdname *C.char
	if dname != "" {
		cdname = C.CString(dname)
		defer C.free(unsafe.Pointer(cdname))
	}
	var curi *C.char
	if uri != "" {
		curi = C.CString(uri)
		defer C.free(unsafe.Pointer(curi))
	}

	ret := C.virDomainMigrate2(d.ptr, dconn.ptr, cdxml, C.ulong(flags), cdname, curi, C.ulong(bandwidth))
	if ret == nil {
		return nil, GetLastError()
	}

	return &Domain{
		ptr: ret,
	}, nil
}

type DomainMigrateParameters struct {
	URISet                    bool
	URI                       string
	DestNameSet               bool
	DestName                  string
	DestXMLSet                bool
	DestXML                   string
	PersistXMLSet             bool
	PersistXML                string
	BandwidthSet              bool
	Bandwidth                 uint64
	GraphicsURISet            bool
	GraphicsURI               string
	ListenAddressSet          bool
	ListenAddress             string
	MigrateDisksSet           bool
	MigrateDisks              []string
	DisksPortSet              bool
	DisksPort                 int
	CompressionSet            bool
	Compression               string
	CompressionMTLevelSet     bool
	CompressionMTLevel        int
	CompressionMTThreadsSet   bool
	CompressionMTThreads      int
	CompressionMTDThreadsSet  bool
	CompressionMTDThreads     int
	CompressionXBZRLECacheSet bool
	CompressionXBZRLECache    uint64
	AutoConvergeInitialSet    bool
	AutoConvergeInitial       int
	AutoConvergeIncrementSet  bool
	AutoConvergeIncrement     int
}

func getMigrateParameterFieldInfo(params *DomainMigrateParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_MIGRATE_PARAM_URI: typedParamsFieldInfo{
			set: &params.URISet,
			s:   &params.URI,
		},
		C.VIR_MIGRATE_PARAM_DEST_NAME: typedParamsFieldInfo{
			set: &params.DestNameSet,
			s:   &params.DestName,
		},
		C.VIR_MIGRATE_PARAM_DEST_XML: typedParamsFieldInfo{
			set: &params.DestXMLSet,
			s:   &params.DestXML,
		},
		C.VIR_MIGRATE_PARAM_PERSIST_XML: typedParamsFieldInfo{
			set: &params.PersistXMLSet,
			s:   &params.PersistXML,
		},
		C.VIR_MIGRATE_PARAM_BANDWIDTH: typedParamsFieldInfo{
			set: &params.BandwidthSet,
			ul:  &params.Bandwidth,
		},
		C.VIR_MIGRATE_PARAM_GRAPHICS_URI: typedParamsFieldInfo{
			set: &params.GraphicsURISet,
			s:   &params.GraphicsURI,
		},
		C.VIR_MIGRATE_PARAM_LISTEN_ADDRESS: typedParamsFieldInfo{
			set: &params.ListenAddressSet,
			s:   &params.ListenAddress,
		},
		C.VIR_MIGRATE_PARAM_MIGRATE_DISKS: typedParamsFieldInfo{
			set: &params.MigrateDisksSet,
			sl:  &params.MigrateDisks,
		},
		C.VIR_MIGRATE_PARAM_DISKS_PORT: typedParamsFieldInfo{
			set: &params.DisksPortSet,
			i:   &params.DisksPort,
		},
		C.VIR_MIGRATE_PARAM_COMPRESSION: typedParamsFieldInfo{
			set: &params.CompressionSet,
			s:   &params.Compression,
		},
		C.VIR_MIGRATE_PARAM_COMPRESSION_MT_LEVEL: typedParamsFieldInfo{
			set: &params.CompressionMTLevelSet,
			i:   &params.CompressionMTLevel,
		},
		C.VIR_MIGRATE_PARAM_COMPRESSION_MT_THREADS: typedParamsFieldInfo{
			set: &params.CompressionMTThreadsSet,
			i:   &params.CompressionMTThreads,
		},
		C.VIR_MIGRATE_PARAM_COMPRESSION_MT_DTHREADS: typedParamsFieldInfo{
			set: &params.CompressionMTDThreadsSet,
			i:   &params.CompressionMTDThreads,
		},
		C.VIR_MIGRATE_PARAM_COMPRESSION_XBZRLE_CACHE: typedParamsFieldInfo{
			set: &params.CompressionXBZRLECacheSet,
			ul:  &params.CompressionXBZRLECache,
		},
		C.VIR_MIGRATE_PARAM_AUTO_CONVERGE_INITIAL: typedParamsFieldInfo{
			set: &params.AutoConvergeInitialSet,
			i:   &params.AutoConvergeInitial,
		},
		C.VIR_MIGRATE_PARAM_AUTO_CONVERGE_INCREMENT: typedParamsFieldInfo{
			set: &params.AutoConvergeIncrementSet,
			i:   &params.AutoConvergeIncrement,
		},
	}
}

func (d *Domain) Migrate3(dconn *Connect, params *DomainMigrateParameters, flags DomainMigrateFlags) (*Domain, error) {

	info := getMigrateParameterFieldInfo(params)
	cparams, err := typedParamsPackNew(info)
	if err != nil {
		return nil, err
	}
	nparams := len(*cparams)

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&(*cparams)[0])), C.int(nparams))

	ret := C.virDomainMigrate3(d.ptr, dconn.ptr, (*C.virTypedParameter)(unsafe.Pointer(&(*cparams)[0])), C.uint(nparams), C.uint(flags))
	if ret == nil {
		return nil, GetLastError()
	}

	return &Domain{
		ptr: ret,
	}, nil
}

func (d *Domain) MigrateToURI(duri string, flags DomainMigrateFlags, dname string, bandwidth uint64) error {
	cduri := C.CString(duri)
	defer C.free(unsafe.Pointer(cduri))

	var cdname *C.char
	if dname != "" {
		cdname = C.CString(dname)
		defer C.free(unsafe.Pointer(cdname))
	}

	ret := C.virDomainMigrateToURI(d.ptr, cduri, C.ulong(flags), cdname, C.ulong(bandwidth))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) MigrateToURI2(dconnuri string, miguri string, dxml string, flags DomainMigrateFlags, dname string, bandwidth uint64) error {
	var cdconnuri *C.char
	if dconnuri != "" {
		cdconnuri = C.CString(dconnuri)
		defer C.free(unsafe.Pointer(cdconnuri))
	}
	var cmiguri *C.char
	if miguri != "" {
		cmiguri = C.CString(miguri)
		defer C.free(unsafe.Pointer(cmiguri))
	}
	var cdxml *C.char
	if dxml != "" {
		cdxml = C.CString(dxml)
		defer C.free(unsafe.Pointer(cdxml))
	}
	var cdname *C.char
	if dname != "" {
		cdname = C.CString(dname)
		defer C.free(unsafe.Pointer(cdname))
	}

	ret := C.virDomainMigrateToURI2(d.ptr, cdconnuri, cmiguri, cdxml, C.ulong(flags), cdname, C.ulong(bandwidth))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) MigrateToURI3(dconnuri string, params *DomainMigrateParameters, flags DomainMigrateFlags) error {
	var cdconnuri *C.char
	if dconnuri != "" {
		cdconnuri = C.CString(dconnuri)
		defer C.free(unsafe.Pointer(cdconnuri))
	}

	info := getMigrateParameterFieldInfo(params)
	cparams, err := typedParamsPackNew(info)
	if err != nil {
		return err
	}
	nparams := len(*cparams)

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&(*cparams)[0])), C.int(nparams))

	ret := C.virDomainMigrateToURI3(d.ptr, cdconnuri, (*C.virTypedParameter)(unsafe.Pointer(&(*cparams)[0])), C.uint(nparams), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) MigrateGetCompressionCache(flags uint32) (uint64, error) {
	var cacheSize C.ulonglong

	ret := C.virDomainMigrateGetCompressionCache(d.ptr, &cacheSize, C.uint(flags))
	if ret == -1 {
		return 0, GetLastError()
	}

	return uint64(cacheSize), nil
}

func (d *Domain) MigrateSetCompressionCache(size uint64, flags uint32) error {
	ret := C.virDomainMigrateSetCompressionCache(d.ptr, C.ulonglong(size), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) MigrateGetMaxSpeed(flags uint32) (uint64, error) {
	var maxSpeed C.ulong

	ret := C.virDomainMigrateGetMaxSpeed(d.ptr, &maxSpeed, C.uint(flags))
	if ret == -1 {
		return 0, GetLastError()
	}

	return uint64(maxSpeed), nil
}

func (d *Domain) MigrateSetMaxSpeed(speed uint64, flags uint32) error {
	ret := C.virDomainMigrateSetMaxSpeed(d.ptr, C.ulong(speed), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) MigrateSetMaxDowntime(downtime uint64, flags uint32) error {
	ret := C.virDomainMigrateSetMaxDowntime(d.ptr, C.ulonglong(downtime), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) MigrateStartPostCopy(flags uint32) error {
	if C.LIBVIR_VERSION_NUMBER < 1003003 {
		return GetNotImplementedError("virDomainMigrateStartPostCopy")
	}

	ret := C.virDomainMigrateStartPostCopyCompat(d.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

type DomainBlkioParameters struct {
	WeightSet          bool
	Weight             uint
	DeviceWeightSet    bool
	DeviceWeight       string
	DeviceReadIopsSet  bool
	DeviceReadIops     string
	DeviceWriteIopsSet bool
	DeviceWriteIops    string
	DeviceReadBpsSet   bool
	DeviceReadBps      string
	DeviceWriteBpsSet  bool
	DeviceWriteBps     string
}

func getBlkioParametersFieldInfo(params *DomainBlkioParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_BLKIO_WEIGHT: typedParamsFieldInfo{
			set: &params.WeightSet,
			ui:  &params.Weight,
		},
		C.VIR_DOMAIN_BLKIO_DEVICE_WEIGHT: typedParamsFieldInfo{
			set: &params.DeviceWeightSet,
			s:   &params.DeviceWeight,
		},
		C.VIR_DOMAIN_BLKIO_DEVICE_READ_IOPS: typedParamsFieldInfo{
			set: &params.DeviceReadIopsSet,
			s:   &params.DeviceReadIops,
		},
		C.VIR_DOMAIN_BLKIO_DEVICE_WRITE_IOPS: typedParamsFieldInfo{
			set: &params.DeviceWriteIopsSet,
			s:   &params.DeviceWriteIops,
		},
		C.VIR_DOMAIN_BLKIO_DEVICE_READ_BPS: typedParamsFieldInfo{
			set: &params.DeviceReadBpsSet,
			s:   &params.DeviceReadBps,
		},
		C.VIR_DOMAIN_BLKIO_DEVICE_WRITE_BPS: typedParamsFieldInfo{
			set: &params.DeviceWriteBpsSet,
			s:   &params.DeviceWriteBps,
		},
	}
}

func (d *Domain) GetBlkioParameters(flags DomainModificationImpact) (*DomainBlkioParameters, error) {
	params := &DomainBlkioParameters{}
	info := getBlkioParametersFieldInfo(params)

	var nparams C.int
	ret := C.virDomainGetBlkioParameters(d.ptr, nil, &nparams, 0)
	if ret == -1 {
		return nil, GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetBlkioParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	_, err := typedParamsUnpack(cparams, info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Domain) SetBlkioParameters(params *DomainBlkioParameters, flags DomainModificationImpact) error {
	info := getBlkioParametersFieldInfo(params)

	var nparams C.int

	ret := C.virDomainGetBlkioParameters(d.ptr, nil, &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetBlkioParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	err := typedParamsPack(cparams, info)
	if err != nil {
		return err
	}

	ret = C.virDomainSetBlkioParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams, C.uint(flags))

	return nil
}

type DomainBlockIoTuneParameters struct {
	TotalBytesSecSet          bool
	TotalBytesSec             uint64
	ReadBytesSecSet           bool
	ReadBytesSec              uint64
	WriteBytesSecSet          bool
	WriteBytesSec             uint64
	TotalIopsSecSet           bool
	TotalIopsSec              uint64
	ReadIopsSecSet            bool
	ReadIopsSec               uint64
	WriteIopsSecSet           bool
	WriteIopsSec              uint64
	TotalBytesSecMaxSet       bool
	TotalBytesSecMax          uint64
	ReadBytesSecMaxSet        bool
	ReadBytesSecMax           uint64
	WriteBytesSecMaxSet       bool
	WriteBytesSecMax          uint64
	TotalIopsSecMaxSet        bool
	TotalIopsSecMax           uint64
	ReadIopsSecMaxSet         bool
	ReadIopsSecMax            uint64
	WriteIopsSecMaxSet        bool
	WriteIopsSecMax           uint64
	TotalBytesSecMaxLengthSet bool
	TotalBytesSecMaxLength    uint64
	ReadBytesSecMaxLengthSet  bool
	ReadBytesSecMaxLength     uint64
	WriteBytesSecMaxLengthSet bool
	WriteBytesSecMaxLength    uint64
	TotalIopsSecMaxLengthSet  bool
	TotalIopsSecMaxLength     uint64
	ReadIopsSecMaxLengthSet   bool
	ReadIopsSecMaxLength      uint64
	WriteIopsSecMaxLengthSet  bool
	WriteIopsSecMaxLength     uint64
	SizeIopsSecSet            bool
	SizeIopsSec               uint64
	GroupNameSet              bool
	GroupName                 string
}

func getBlockIoTuneParametersFieldInfo(params *DomainBlockIoTuneParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_BYTES_SEC: typedParamsFieldInfo{
			set: &params.TotalBytesSecSet,
			ul:  &params.TotalBytesSec,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_READ_BYTES_SEC: typedParamsFieldInfo{
			set: &params.ReadBytesSecSet,
			ul:  &params.ReadBytesSec,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_WRITE_BYTES_SEC: typedParamsFieldInfo{
			set: &params.WriteBytesSecSet,
			ul:  &params.WriteBytesSec,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_IOPS_SEC: typedParamsFieldInfo{
			set: &params.TotalIopsSecSet,
			ul:  &params.TotalIopsSec,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_READ_IOPS_SEC: typedParamsFieldInfo{
			set: &params.ReadIopsSecSet,
			ul:  &params.ReadIopsSec,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_WRITE_IOPS_SEC: typedParamsFieldInfo{
			set: &params.WriteIopsSecSet,
			ul:  &params.WriteIopsSec,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_BYTES_SEC_MAX: typedParamsFieldInfo{
			set: &params.TotalBytesSecMaxSet,
			ul:  &params.TotalBytesSecMax,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_READ_BYTES_SEC_MAX: typedParamsFieldInfo{
			set: &params.ReadBytesSecMaxSet,
			ul:  &params.ReadBytesSecMax,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_WRITE_BYTES_SEC_MAX: typedParamsFieldInfo{
			set: &params.WriteBytesSecMaxSet,
			ul:  &params.WriteBytesSecMax,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_IOPS_SEC_MAX: typedParamsFieldInfo{
			set: &params.TotalIopsSecMaxSet,
			ul:  &params.TotalIopsSecMax,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_READ_IOPS_SEC_MAX: typedParamsFieldInfo{
			set: &params.ReadIopsSecMaxSet,
			ul:  &params.ReadIopsSecMax,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_WRITE_IOPS_SEC_MAX: typedParamsFieldInfo{
			set: &params.WriteIopsSecMaxSet,
			ul:  &params.WriteIopsSecMax,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_BYTES_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.TotalBytesSecMaxLengthSet,
			ul:  &params.TotalBytesSecMaxLength,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_READ_BYTES_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.ReadBytesSecMaxLengthSet,
			ul:  &params.ReadBytesSecMaxLength,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_WRITE_BYTES_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.WriteBytesSecMaxLengthSet,
			ul:  &params.WriteBytesSecMaxLength,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_IOPS_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.TotalIopsSecMaxLengthSet,
			ul:  &params.TotalIopsSecMaxLength,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_READ_IOPS_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.ReadIopsSecMaxLengthSet,
			ul:  &params.ReadIopsSecMaxLength,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_WRITE_IOPS_SEC_MAX_LENGTH: typedParamsFieldInfo{
			set: &params.WriteIopsSecMaxLengthSet,
			ul:  &params.WriteIopsSecMaxLength,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_SIZE_IOPS_SEC: typedParamsFieldInfo{
			set: &params.SizeIopsSecSet,
			ul:  &params.SizeIopsSec,
		},
		C.VIR_DOMAIN_BLOCK_IOTUNE_GROUP_NAME: typedParamsFieldInfo{
			set: &params.GroupNameSet,
			s:   &params.GroupName,
		},
	}
}

func (d *Domain) GetBlockIoTune(disk string, flags DomainModificationImpact) (*DomainBlockIoTuneParameters, error) {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))

	params := &DomainBlockIoTuneParameters{}
	info := getBlockIoTuneParametersFieldInfo(params)

	var nparams C.int
	ret := C.virDomainGetBlockIoTune(d.ptr, cdisk, nil, &nparams, 0)
	if ret == -1 {
		return nil, GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetBlockIoTune(d.ptr, cdisk, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	_, err := typedParamsUnpack(cparams, info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Domain) SetBlockIoTune(disk string, params *DomainBlockIoTuneParameters, flags DomainModificationImpact) error {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))

	info := getBlockIoTuneParametersFieldInfo(params)

	var nparams C.int

	ret := C.virDomainGetBlockIoTune(d.ptr, cdisk, nil, &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetBlockIoTune(d.ptr, cdisk, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	err := typedParamsPack(cparams, info)
	if err != nil {
		return err
	}

	ret = C.virDomainSetBlockIoTune(d.ptr, cdisk, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams, C.uint(flags))

	return nil
}

type DomainBlockJobInfo struct {
	Type      DomainBlockJobType
	Bandwidth uint64
	Cur       uint64
	End       uint64
}

func (d *Domain) GetBlockJobInfo(disk string, flags DomainBlockJobInfoFlags) (*DomainBlockJobInfo, error) {
	cdisk := C.CString(disk)
	defer C.free(unsafe.Pointer(cdisk))

	var cinfo C.virDomainBlockJobInfo

	ret := C.virDomainGetBlockJobInfo(d.ptr, cdisk, &cinfo, C.uint(flags))

	if ret == -1 {
		return nil, GetLastError()
	}

	return &DomainBlockJobInfo{
		Type:      DomainBlockJobType(cinfo._type),
		Bandwidth: uint64(cinfo.bandwidth),
		Cur:       uint64(cinfo.cur),
		End:       uint64(cinfo.end),
	}, nil
}

type DomainControlInfo struct {
	State     DomainControlState
	Details   int
	StateTime uint64
}

func (d *Domain) GetControlInfo(flags uint32) (*DomainControlInfo, error) {

	var cinfo C.virDomainControlInfo

	ret := C.virDomainGetControlInfo(d.ptr, &cinfo, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	return &DomainControlInfo{
		State:     DomainControlState(cinfo.state),
		Details:   int(cinfo.details),
		StateTime: uint64(cinfo.stateTime),
	}, nil
}

type DomainDiskError struct {
	Disk  string
	Error DomainDiskErrorCode
}

func (d *Domain) GetDiskErrors(flags uint32) ([]DomainDiskError, error) {
	ret := C.virDomainGetDiskErrors(d.ptr, nil, 0, 0)
	if ret == -1 {
		return []DomainDiskError{}, GetLastError()
	}

	maxerrors := ret
	cerrors := make([]C.virDomainDiskError, maxerrors)

	ret = C.virDomainGetDiskErrors(d.ptr, (*C.virDomainDiskError)(unsafe.Pointer(&cerrors[0])), C.uint(maxerrors), C.uint(flags))
	if ret == -1 {
		return []DomainDiskError{}, GetLastError()
	}

	errors := make([]DomainDiskError, maxerrors)

	for i, cerror := range cerrors {
		errors[i] = DomainDiskError{
			Disk:  C.GoString(cerror.disk),
			Error: DomainDiskErrorCode(cerror.error),
		}
		C.free(unsafe.Pointer(cerror.disk))
	}

	return errors, nil
}

func (d *Domain) GetHostname(flags uint32) (string, error) {
	ret := C.virDomainGetHostname(d.ptr, C.uint(flags))
	if ret == nil {
		return "", GetLastError()
	}

	defer C.free(unsafe.Pointer(ret))

	return C.GoString(ret), nil
}

type DomainJobInfo struct {
	Type                      DomainJobType
	TimeElapsedSet            bool
	TimeElapsed               uint64
	TimeElapsedNetSet         bool
	TimeElapsedNet            uint64
	TimeRemainingSet          bool
	TimeRemaining             uint64
	DowntimeSet               bool
	Downtime                  uint64
	DowntimeNetSet            bool
	DowntimeNet               uint64
	SetupTimeSet              bool
	SetupTime                 uint64
	DataTotalSet              bool
	DataTotal                 uint64
	DataProcessedSet          bool
	DataProcessed             uint64
	DataRemainingSet          bool
	DataRemaining             uint64
	MemTotalSet               bool
	MemTotal                  uint64
	MemProcessedSet           bool
	MemProcessed              uint64
	MemRemainingSet           bool
	MemRemaining              uint64
	MemConstantSet            bool
	MemConstant               uint64
	MemNormalSet              bool
	MemNormal                 uint64
	MemNormalBytesSet         bool
	MemNormalBytes            uint64
	MemBpsSet                 bool
	MemBps                    uint64
	MemDirtyRateSet           bool
	MemDirtyRate              uint64
	MemIterationSet           bool
	MemIteration              uint64
	DiskTotalSet              bool
	DiskTotal                 uint64
	DiskProcessedSet          bool
	DiskProcessed             uint64
	DiskRemainingSet          bool
	DiskRemaining             uint64
	DiskBpsSet                bool
	DiskBps                   uint64
	CompressionCacheSet       bool
	CompressionCache          uint64
	CompressionBytesSet       bool
	CompressionBytes          uint64
	CompressionPagesSet       bool
	CompressionPages          uint64
	CompressionCacheMissesSet bool
	CompressionCacheMisses    uint64
	CompressionOverflowSet    bool
	CompressionOverflow       uint64
	AutoConvergeThrottleSet   bool
	AutoConvergeThrottle      int
	OperationSet              bool
	Operation                 DomainJobOperationType
}

func (d *Domain) GetJobInfo() (*DomainJobInfo, error) {
	var cinfo C.virDomainJobInfo

	ret := C.virDomainGetJobInfo(d.ptr, &cinfo)
	if ret == -1 {
		return nil, GetLastError()
	}

	return &DomainJobInfo{
		Type:             DomainJobType(cinfo._type),
		TimeElapsedSet:   true,
		TimeElapsed:      uint64(cinfo.timeElapsed),
		TimeRemainingSet: true,
		TimeRemaining:    uint64(cinfo.timeRemaining),
		DataTotalSet:     true,
		DataTotal:        uint64(cinfo.dataTotal),
		DataProcessedSet: true,
		DataProcessed:    uint64(cinfo.dataProcessed),
		DataRemainingSet: true,
		DataRemaining:    uint64(cinfo.dataRemaining),
		MemTotalSet:      true,
		MemTotal:         uint64(cinfo.memTotal),
		MemProcessedSet:  true,
		MemProcessed:     uint64(cinfo.memProcessed),
		MemRemainingSet:  true,
		MemRemaining:     uint64(cinfo.memRemaining),
		DiskTotalSet:     true,
		DiskTotal:        uint64(cinfo.fileTotal),
		DiskProcessedSet: true,
		DiskProcessed:    uint64(cinfo.fileProcessed),
		DiskRemainingSet: true,
		DiskRemaining:    uint64(cinfo.fileRemaining),
	}, nil
}

func getDomainJobInfoFieldInfo(params *DomainJobInfo) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_JOB_TIME_ELAPSED: typedParamsFieldInfo{
			set: &params.TimeElapsedSet,
			ul:  &params.TimeElapsed,
		},
		C.VIR_DOMAIN_JOB_TIME_ELAPSED_NET: typedParamsFieldInfo{
			set: &params.TimeElapsedNetSet,
			ul:  &params.TimeElapsedNet,
		},
		C.VIR_DOMAIN_JOB_TIME_REMAINING: typedParamsFieldInfo{
			set: &params.TimeRemainingSet,
			ul:  &params.TimeRemaining,
		},
		C.VIR_DOMAIN_JOB_DOWNTIME: typedParamsFieldInfo{
			set: &params.DowntimeSet,
			ul:  &params.Downtime,
		},
		C.VIR_DOMAIN_JOB_DOWNTIME_NET: typedParamsFieldInfo{
			set: &params.DowntimeNetSet,
			ul:  &params.DowntimeNet,
		},
		C.VIR_DOMAIN_JOB_SETUP_TIME: typedParamsFieldInfo{
			set: &params.SetupTimeSet,
			ul:  &params.SetupTime,
		},
		C.VIR_DOMAIN_JOB_DATA_TOTAL: typedParamsFieldInfo{
			set: &params.DataTotalSet,
			ul:  &params.DataTotal,
		},
		C.VIR_DOMAIN_JOB_DATA_PROCESSED: typedParamsFieldInfo{
			set: &params.DataProcessedSet,
			ul:  &params.DataProcessed,
		},
		C.VIR_DOMAIN_JOB_DATA_REMAINING: typedParamsFieldInfo{
			set: &params.DataRemainingSet,
			ul:  &params.DataRemaining,
		},
		C.VIR_DOMAIN_JOB_MEMORY_TOTAL: typedParamsFieldInfo{
			set: &params.MemTotalSet,
			ul:  &params.MemTotal,
		},
		C.VIR_DOMAIN_JOB_MEMORY_PROCESSED: typedParamsFieldInfo{
			set: &params.MemProcessedSet,
			ul:  &params.MemProcessed,
		},
		C.VIR_DOMAIN_JOB_MEMORY_REMAINING: typedParamsFieldInfo{
			set: &params.MemRemainingSet,
			ul:  &params.MemRemaining,
		},
		C.VIR_DOMAIN_JOB_MEMORY_CONSTANT: typedParamsFieldInfo{
			set: &params.MemConstantSet,
			ul:  &params.MemConstant,
		},
		C.VIR_DOMAIN_JOB_MEMORY_NORMAL: typedParamsFieldInfo{
			set: &params.MemNormalSet,
			ul:  &params.MemNormal,
		},
		C.VIR_DOMAIN_JOB_MEMORY_NORMAL_BYTES: typedParamsFieldInfo{
			set: &params.MemNormalBytesSet,
			ul:  &params.MemNormalBytes,
		},
		C.VIR_DOMAIN_JOB_MEMORY_BPS: typedParamsFieldInfo{
			set: &params.MemBpsSet,
			ul:  &params.MemBps,
		},
		C.VIR_DOMAIN_JOB_MEMORY_DIRTY_RATE: typedParamsFieldInfo{
			set: &params.MemDirtyRateSet,
			ul:  &params.MemDirtyRate,
		},
		C.VIR_DOMAIN_JOB_MEMORY_ITERATION: typedParamsFieldInfo{
			set: &params.MemIterationSet,
			ul:  &params.MemIteration,
		},
		C.VIR_DOMAIN_JOB_DISK_TOTAL: typedParamsFieldInfo{
			set: &params.DiskTotalSet,
			ul:  &params.DiskTotal,
		},
		C.VIR_DOMAIN_JOB_DISK_PROCESSED: typedParamsFieldInfo{
			set: &params.DiskProcessedSet,
			ul:  &params.DiskProcessed,
		},
		C.VIR_DOMAIN_JOB_DISK_REMAINING: typedParamsFieldInfo{
			set: &params.DiskRemainingSet,
			ul:  &params.DiskRemaining,
		},
		C.VIR_DOMAIN_JOB_DISK_BPS: typedParamsFieldInfo{
			set: &params.DiskBpsSet,
			ul:  &params.DiskBps,
		},
		C.VIR_DOMAIN_JOB_COMPRESSION_CACHE: typedParamsFieldInfo{
			set: &params.CompressionCacheSet,
			ul:  &params.CompressionCache,
		},
		C.VIR_DOMAIN_JOB_COMPRESSION_BYTES: typedParamsFieldInfo{
			set: &params.CompressionBytesSet,
			ul:  &params.CompressionBytes,
		},
		C.VIR_DOMAIN_JOB_COMPRESSION_PAGES: typedParamsFieldInfo{
			set: &params.CompressionPagesSet,
			ul:  &params.CompressionPages,
		},
		C.VIR_DOMAIN_JOB_COMPRESSION_CACHE_MISSES: typedParamsFieldInfo{
			set: &params.CompressionCacheMissesSet,
			ul:  &params.CompressionCacheMisses,
		},
		C.VIR_DOMAIN_JOB_COMPRESSION_OVERFLOW: typedParamsFieldInfo{
			set: &params.CompressionOverflowSet,
			ul:  &params.CompressionOverflow,
		},
		C.VIR_DOMAIN_JOB_AUTO_CONVERGE_THROTTLE: typedParamsFieldInfo{
			set: &params.AutoConvergeThrottleSet,
			i:   &params.AutoConvergeThrottle,
		},
		C.VIR_DOMAIN_JOB_OPERATION: typedParamsFieldInfo{
			set: &params.OperationSet,
			i:   (*int)(&params.Operation),
		},
	}
}

func (d *Domain) GetJobStats(flags DomainGetJobStatsFlags) (*DomainJobInfo, error) {
	var cparams *C.virTypedParameter
	var nparams C.int
	var jobtype C.int
	ret := C.virDomainGetJobStats(d.ptr, &jobtype, (*C.virTypedParameterPtr)(unsafe.Pointer(&cparams)), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}
	defer C.virTypedParamsFree(cparams, nparams)

	params := DomainJobInfo{}
	info := getDomainJobInfoFieldInfo(&params)

	_, err := typedParamsUnpackLen(cparams, int(nparams), info)
	if err != nil {
		return nil, GetLastError()
	}

	return &params, nil
}

func (d *Domain) GetMaxMemory() (uint64, error) {
	ret := C.virDomainGetMaxMemory(d.ptr)
	if ret == 0 {
		return 0, GetLastError()
	}

	return uint64(ret), nil
}

func (d *Domain) GetMaxVcpus() (uint, error) {
	ret := C.virDomainGetMaxVcpus(d.ptr)
	if ret == -1 {
		return 0, GetLastError()
	}

	return uint(ret), nil
}

func (d *Domain) GetOSType() (string, error) {
	ret := C.virDomainGetOSType(d.ptr)
	if ret == nil {
		return "", GetLastError()
	}

	defer C.free(unsafe.Pointer(ret))

	return C.GoString(ret), nil
}

type DomainMemoryParameters struct {
	HardLimitSet     bool
	HardLimit        uint64
	SoftLimitSet     bool
	SoftLimit        uint64
	MinGuaranteeSet  bool
	MinGuarantee     uint64
	SwapHardLimitSet bool
	SwapHardLimit    uint64
}

func getDomainMemoryParametersFieldInfo(params *DomainMemoryParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_MEMORY_HARD_LIMIT: typedParamsFieldInfo{
			set: &params.HardLimitSet,
			ul:  &params.HardLimit,
		},
		C.VIR_DOMAIN_MEMORY_SOFT_LIMIT: typedParamsFieldInfo{
			set: &params.SoftLimitSet,
			ul:  &params.SoftLimit,
		},
		C.VIR_DOMAIN_MEMORY_MIN_GUARANTEE: typedParamsFieldInfo{
			set: &params.MinGuaranteeSet,
			ul:  &params.MinGuarantee,
		},
		C.VIR_DOMAIN_MEMORY_SWAP_HARD_LIMIT: typedParamsFieldInfo{
			set: &params.SwapHardLimitSet,
			ul:  &params.SwapHardLimit,
		},
	}
}

func (d *Domain) GetMemoryParameters(flags DomainModificationImpact) (*DomainMemoryParameters, error) {
	params := &DomainMemoryParameters{}
	info := getDomainMemoryParametersFieldInfo(params)

	var nparams C.int
	ret := C.virDomainGetMemoryParameters(d.ptr, nil, &nparams, 0)
	if ret == -1 {
		return nil, GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetMemoryParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	_, err := typedParamsUnpack(cparams, info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Domain) SetMemoryParameters(params *DomainMemoryParameters, flags DomainModificationImpact) error {
	info := getDomainMemoryParametersFieldInfo(params)

	var nparams C.int

	ret := C.virDomainGetMemoryParameters(d.ptr, nil, &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetMemoryParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	err := typedParamsPack(cparams, info)
	if err != nil {
		return err
	}

	ret = C.virDomainSetMemoryParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams, C.uint(flags))

	return nil
}

type DomainNumaParameters struct {
	NodesetSet bool
	Nodeset    string
	ModeSet    bool
	Mode       DomainNumatuneMemMode
}

func getDomainNumaParametersFieldInfo(params *DomainNumaParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_NUMA_NODESET: typedParamsFieldInfo{
			set: &params.NodesetSet,
			s:   &params.Nodeset,
		},
		C.VIR_DOMAIN_NUMA_MODE: typedParamsFieldInfo{
			set: &params.ModeSet,
			i:   (*int)(&params.Mode),
		},
	}
}

func (d *Domain) GetNumaParameters(flags DomainModificationImpact) (*DomainNumaParameters, error) {
	params := &DomainNumaParameters{}
	info := getDomainNumaParametersFieldInfo(params)

	var nparams C.int
	ret := C.virDomainGetNumaParameters(d.ptr, nil, &nparams, 0)
	if ret == -1 {
		return nil, GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetNumaParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	_, err := typedParamsUnpack(cparams, info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Domain) SetNumaParameters(params *DomainNumaParameters, flags DomainModificationImpact) error {
	info := getDomainNumaParametersFieldInfo(params)

	var nparams C.int

	ret := C.virDomainGetNumaParameters(d.ptr, nil, &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virDomainGetNumaParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	err := typedParamsPack(cparams, info)
	if err != nil {
		return err
	}

	ret = C.virDomainSetNumaParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams, C.uint(flags))

	return nil
}

type DomainPerfEvents struct {
	CmtSet                   bool
	Cmt                      bool
	MbmtSet                  bool
	Mbmt                     bool
	MbmlSet                  bool
	Mbml                     bool
	CacheMissesSet           bool
	CacheMisses              bool
	CacheReferencesSet       bool
	CacheReferences          bool
	InstructionsSet          bool
	Instructions             bool
	CpuCyclesSet             bool
	CpuCycles                bool
	BranchInstructionsSet    bool
	BranchInstructions       bool
	BranchMissesSet          bool
	BranchMisses             bool
	BusCyclesSet             bool
	BusCycles                bool
	StalledCyclesFrontendSet bool
	StalledCyclesFrontend    bool
	StalledCyclesBackendSet  bool
	StalledCyclesBackend     bool
	RefCpuCyclesSet          bool
	RefCpuCycles             bool
	CpuClockSet              bool
	CpuClock                 bool
	TaskClockSet             bool
	TaskClock                bool
	PageFaultsSet            bool
	PageFaults               bool
	ContextSwitchesSet       bool
	ContextSwitches          bool
	CpuMigrationsSet         bool
	CpuMigrations            bool
	PageFaultsMinSet         bool
	PageFaultsMin            bool
	PageFaultsMajSet         bool
	PageFaultsMaj            bool
	AlignmentFaultsSet       bool
	AlignmentFaults          bool
	EmulationFaultsSet       bool
	EmulationFaults          bool
}

/* Remember to also update DomainStatsPerf in connect.go when adding to the stuct above */

func getDomainPerfEventsFieldInfo(params *DomainPerfEvents) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_PERF_PARAM_CMT: typedParamsFieldInfo{
			set: &params.CmtSet,
			b:   &params.Cmt,
		},
		C.VIR_PERF_PARAM_MBMT: typedParamsFieldInfo{
			set: &params.MbmtSet,
			b:   &params.Mbmt,
		},
		C.VIR_PERF_PARAM_MBML: typedParamsFieldInfo{
			set: &params.MbmlSet,
			b:   &params.Mbml,
		},
		C.VIR_PERF_PARAM_CACHE_MISSES: typedParamsFieldInfo{
			set: &params.CacheMissesSet,
			b:   &params.CacheMisses,
		},
		C.VIR_PERF_PARAM_CACHE_REFERENCES: typedParamsFieldInfo{
			set: &params.CacheReferencesSet,
			b:   &params.CacheReferences,
		},
		C.VIR_PERF_PARAM_INSTRUCTIONS: typedParamsFieldInfo{
			set: &params.InstructionsSet,
			b:   &params.Instructions,
		},
		C.VIR_PERF_PARAM_CPU_CYCLES: typedParamsFieldInfo{
			set: &params.CpuCyclesSet,
			b:   &params.CpuCycles,
		},
		C.VIR_PERF_PARAM_BRANCH_INSTRUCTIONS: typedParamsFieldInfo{
			set: &params.BranchInstructionsSet,
			b:   &params.BranchInstructions,
		},
		C.VIR_PERF_PARAM_BRANCH_MISSES: typedParamsFieldInfo{
			set: &params.BranchMissesSet,
			b:   &params.BranchMisses,
		},
		C.VIR_PERF_PARAM_BUS_CYCLES: typedParamsFieldInfo{
			set: &params.BusCyclesSet,
			b:   &params.BusCycles,
		},
		C.VIR_PERF_PARAM_STALLED_CYCLES_FRONTEND: typedParamsFieldInfo{
			set: &params.StalledCyclesFrontendSet,
			b:   &params.StalledCyclesFrontend,
		},
		C.VIR_PERF_PARAM_STALLED_CYCLES_BACKEND: typedParamsFieldInfo{
			set: &params.StalledCyclesBackendSet,
			b:   &params.StalledCyclesBackend,
		},
		C.VIR_PERF_PARAM_REF_CPU_CYCLES: typedParamsFieldInfo{
			set: &params.RefCpuCyclesSet,
			b:   &params.RefCpuCycles,
		},
		C.VIR_PERF_PARAM_CPU_CLOCK: typedParamsFieldInfo{
			set: &params.CpuClockSet,
			b:   &params.CpuClock,
		},
		C.VIR_PERF_PARAM_TASK_CLOCK: typedParamsFieldInfo{
			set: &params.TaskClockSet,
			b:   &params.TaskClock,
		},
		C.VIR_PERF_PARAM_PAGE_FAULTS: typedParamsFieldInfo{
			set: &params.PageFaultsSet,
			b:   &params.PageFaults,
		},
		C.VIR_PERF_PARAM_CONTEXT_SWITCHES: typedParamsFieldInfo{
			set: &params.ContextSwitchesSet,
			b:   &params.ContextSwitches,
		},
		C.VIR_PERF_PARAM_CPU_MIGRATIONS: typedParamsFieldInfo{
			set: &params.CpuMigrationsSet,
			b:   &params.CpuMigrations,
		},
		C.VIR_PERF_PARAM_PAGE_FAULTS_MIN: typedParamsFieldInfo{
			set: &params.PageFaultsMinSet,
			b:   &params.PageFaultsMin,
		},
		C.VIR_PERF_PARAM_PAGE_FAULTS_MAJ: typedParamsFieldInfo{
			set: &params.PageFaultsMajSet,
			b:   &params.PageFaultsMaj,
		},
		C.VIR_PERF_PARAM_ALIGNMENT_FAULTS: typedParamsFieldInfo{
			set: &params.AlignmentFaultsSet,
			b:   &params.AlignmentFaults,
		},
		C.VIR_PERF_PARAM_EMULATION_FAULTS: typedParamsFieldInfo{
			set: &params.EmulationFaultsSet,
			b:   &params.EmulationFaults,
		},
	}
}

func (d *Domain) GetPerfEvents(flags DomainModificationImpact) (*DomainPerfEvents, error) {
	if C.LIBVIR_VERSION_NUMBER < 1003003 {
		return nil, GetNotImplementedError("virDomainGetPerfEvents")
	}

	params := &DomainPerfEvents{}
	info := getDomainPerfEventsFieldInfo(params)

	var cparams *C.virTypedParameter
	var nparams C.int
	ret := C.virDomainGetPerfEventsCompat(d.ptr, (*C.virTypedParameterPtr)(unsafe.Pointer(&cparams)), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	defer C.virTypedParamsFree(cparams, nparams)

	_, err := typedParamsUnpackLen(cparams, int(nparams), info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Domain) SetPerfEvents(params *DomainPerfEvents, flags DomainModificationImpact) error {
	if C.LIBVIR_VERSION_NUMBER < 1003003 {
		return GetNotImplementedError("virDomainSetPerfEvents")
	}

	info := getDomainPerfEventsFieldInfo(params)

	var cparams *C.virTypedParameter
	var nparams C.int
	ret := C.virDomainGetPerfEventsCompat(d.ptr, (*C.virTypedParameterPtr)(unsafe.Pointer(&cparams)), &nparams, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	defer C.virTypedParamsFree(cparams, nparams)

	err := typedParamsPackLen(cparams, int(nparams), info)
	if err != nil {
		return err
	}

	ret = C.virDomainSetPerfEventsCompat(d.ptr, cparams, nparams, C.uint(flags))

	return nil
}

type DomainSchedulerParameters struct {
	Type              string
	CpuSharesSet      bool
	CpuShares         uint64
	GlobalPeriodSet   bool
	GlobalPeriod      uint64
	GlobalQuotaSet    bool
	GlobalQuota       uint64
	VcpuPeriodSet     bool
	VcpuPeriod        uint64
	VcpuQuotaSet      bool
	VcpuQuota         uint64
	EmulatorPeriodSet bool
	EmulatorPeriod    uint64
	EmulatorQuotaSet  bool
	EmulatorQuota     uint64
	IothreadPeriodSet bool
	IothreadPeriod    uint64
	IothreadQuotaSet  bool
	IothreadQuota     uint64
	WeightSet         bool
	Weight            uint
	CapSet            bool
	Cap               uint
	ReservationSet    bool
	Reservation       int64
	LimitSet          bool
	Limit             int64
	SharesSet         bool
	Shares            int
}

func getDomainSchedulerParametersFieldInfo(params *DomainSchedulerParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_DOMAIN_SCHEDULER_CPU_SHARES: typedParamsFieldInfo{
			set: &params.CpuSharesSet,
			ul:  &params.CpuShares,
		},
		C.VIR_DOMAIN_SCHEDULER_GLOBAL_PERIOD: typedParamsFieldInfo{
			set: &params.GlobalPeriodSet,
			ul:  &params.GlobalPeriod,
		},
		C.VIR_DOMAIN_SCHEDULER_GLOBAL_QUOTA: typedParamsFieldInfo{
			set: &params.GlobalQuotaSet,
			ul:  &params.GlobalQuota,
		},
		C.VIR_DOMAIN_SCHEDULER_EMULATOR_PERIOD: typedParamsFieldInfo{
			set: &params.EmulatorPeriodSet,
			ul:  &params.EmulatorPeriod,
		},
		C.VIR_DOMAIN_SCHEDULER_EMULATOR_QUOTA: typedParamsFieldInfo{
			set: &params.EmulatorQuotaSet,
			ul:  &params.EmulatorQuota,
		},
		C.VIR_DOMAIN_SCHEDULER_VCPU_PERIOD: typedParamsFieldInfo{
			set: &params.VcpuPeriodSet,
			ul:  &params.VcpuPeriod,
		},
		C.VIR_DOMAIN_SCHEDULER_VCPU_QUOTA: typedParamsFieldInfo{
			set: &params.VcpuQuotaSet,
			ul:  &params.VcpuQuota,
		},
		C.VIR_DOMAIN_SCHEDULER_IOTHREAD_PERIOD: typedParamsFieldInfo{
			set: &params.IothreadPeriodSet,
			ul:  &params.IothreadPeriod,
		},
		C.VIR_DOMAIN_SCHEDULER_IOTHREAD_QUOTA: typedParamsFieldInfo{
			set: &params.IothreadQuotaSet,
			ul:  &params.IothreadQuota,
		},
		C.VIR_DOMAIN_SCHEDULER_WEIGHT: typedParamsFieldInfo{
			set: &params.WeightSet,
			ui:  &params.Weight,
		},
		C.VIR_DOMAIN_SCHEDULER_CAP: typedParamsFieldInfo{
			set: &params.CapSet,
			ui:  &params.Cap,
		},
		C.VIR_DOMAIN_SCHEDULER_RESERVATION: typedParamsFieldInfo{
			set: &params.ReservationSet,
			l:   &params.Reservation,
		},
		C.VIR_DOMAIN_SCHEDULER_LIMIT: typedParamsFieldInfo{
			set: &params.LimitSet,
			l:   &params.Limit,
		},
		C.VIR_DOMAIN_SCHEDULER_SHARES: typedParamsFieldInfo{
			set: &params.SharesSet,
			i:   &params.Shares,
		},
	}
}

func (d *Domain) GetSchedulerParameters() (*DomainSchedulerParameters, error) {
	params := &DomainSchedulerParameters{}
	info := getDomainSchedulerParametersFieldInfo(params)

	var nparams C.int
	schedtype := C.virDomainGetSchedulerType(d.ptr, &nparams)
	if schedtype == nil {
		return nil, GetLastError()
	}

	defer C.free(unsafe.Pointer(schedtype))
	if nparams == 0 {
		return &DomainSchedulerParameters{
			Type: C.GoString(schedtype),
		}, nil
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret := C.virDomainGetSchedulerParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams)
	if ret == -1 {
		return nil, GetLastError()
	}
	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	_, err := typedParamsUnpack(cparams, info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Domain) GetSchedulerParametersFlags(flags DomainModificationImpact) (*DomainSchedulerParameters, error) {
	params := &DomainSchedulerParameters{}
	info := getDomainSchedulerParametersFieldInfo(params)

	var nparams C.int
	schedtype := C.virDomainGetSchedulerType(d.ptr, &nparams)
	if schedtype == nil {
		return nil, GetLastError()
	}

	defer C.free(unsafe.Pointer(schedtype))
	if nparams == 0 {
		return &DomainSchedulerParameters{
			Type: C.GoString(schedtype),
		}, nil
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret := C.virDomainGetSchedulerParametersFlags(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}
	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	_, err := typedParamsUnpack(cparams, info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (d *Domain) SetSchedulerParameters(params *DomainSchedulerParameters) error {
	info := getDomainSchedulerParametersFieldInfo(params)

	var nparams C.int
	schedtype := C.virDomainGetSchedulerType(d.ptr, &nparams)
	if schedtype == nil {
		return GetLastError()
	}

	defer C.free(unsafe.Pointer(schedtype))
	if nparams == 0 {
		return nil
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret := C.virDomainGetSchedulerParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams)
	if ret == -1 {
		return GetLastError()
	}
	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	err := typedParamsPack(cparams, info)
	if err != nil {
		return err
	}

	ret = C.virDomainSetSchedulerParameters(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	return nil
}

func (d *Domain) SetSchedulerParametersFlags(params *DomainSchedulerParameters, flags DomainModificationImpact) error {
	info := getDomainSchedulerParametersFieldInfo(params)

	var nparams C.int
	schedtype := C.virDomainGetSchedulerType(d.ptr, &nparams)
	if schedtype == nil {
		return GetLastError()
	}

	defer C.free(unsafe.Pointer(schedtype))
	if nparams == 0 {
		return nil
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret := C.virDomainGetSchedulerParametersFlags(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}
	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	err := typedParamsPack(cparams, info)
	if err != nil {
		return err
	}

	ret = C.virDomainSetSchedulerParametersFlags(d.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams, C.uint(flags))

	return nil
}

type SecurityLabel struct {
	Label     string
	Enforcing bool
}

func (d *Domain) GetSecurityLabel() (*SecurityLabel, error) {
	var clabel C.virSecurityLabel

	ret := C.virDomainGetSecurityLabel(d.ptr, &clabel)
	if ret == -1 {
		return nil, GetLastError()
	}

	return &SecurityLabel{
		Label:     C.GoString((*C.char)(unsafe.Pointer(&clabel.label))),
		Enforcing: clabel.enforcing == 1,
	}, nil
}

func (d *Domain) GetSecurityLabelList() ([]SecurityLabel, error) {
	var clabels *C.virSecurityLabel

	ret := C.virDomainGetSecurityLabelList(d.ptr, (*C.virSecurityLabelPtr)(unsafe.Pointer(&clabels)))
	if ret == -1 {
		return []SecurityLabel{}, GetLastError()
	}

	labels := make([]SecurityLabel, ret)
	for i := 0; i < int(ret); i++ {
		var clabel *C.virSecurityLabel
		clabel = (*C.virSecurityLabel)(unsafe.Pointer(uintptr(unsafe.Pointer(clabels)) + (unsafe.Sizeof(*clabel) * uintptr(i))))
		labels[i] = SecurityLabel{
			Label:     C.GoString((*C.char)(unsafe.Pointer(&clabel.label))),
			Enforcing: clabel.enforcing == 1,
		}
	}

	return labels, nil
}

func (d *Domain) GetTime(flags uint32) (int64, uint, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002005 {
		return 0, 0, GetNotImplementedError("virDomainGetTime")
	}
	var secs C.longlong
	var nsecs C.uint
	ret := C.virDomainGetTimeCompat(d.ptr, &secs, &nsecs, C.uint(flags))
	if ret == -1 {
		return 0, 0, GetLastError()
	}

	return int64(secs), uint(nsecs), nil
}

func (d *Domain) SetTime(secs int64, nsecs uint, flags DomainSetTimeFlags) error {
	if C.LIBVIR_VERSION_NUMBER < 1002005 {
		return GetNotImplementedError("virDomainSetTime")
	}

	ret := C.virDomainSetTimeCompat(d.ptr, C.longlong(secs), C.uint(nsecs), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) SetUserPassword(user string, password string, flags DomainSetUserPasswordFlags) error {
	if C.LIBVIR_VERSION_NUMBER < 1002015 {
		return GetNotImplementedError("virDomainSetUserPassword")
	}
	cuser := C.CString(user)
	cpassword := C.CString(password)

	defer C.free(unsafe.Pointer(cuser))
	defer C.free(unsafe.Pointer(cpassword))

	ret := C.virDomainSetUserPasswordCompat(d.ptr, cuser, cpassword, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) ManagedSave(flags DomainSaveRestoreFlags) error {
	ret := C.virDomainManagedSave(d.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) HasManagedSaveImage(flags uint32) (bool, error) {
	result := C.virDomainHasManagedSaveImage(d.ptr, C.uint(flags))
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (d *Domain) ManagedSaveRemove(flags uint32) error {
	ret := C.virDomainManagedSaveRemove(d.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) Rename(name string, flags uint32) error {
	if C.LIBVIR_VERSION_NUMBER < 1002019 {
		return GetNotImplementedError("virDomainRename")
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ret := C.virDomainRenameCompat(d.ptr, cname, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) Reset(flags uint32) error {
	ret := C.virDomainReset(d.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) SendProcessSignal(pid int64, signum DomainProcessSignal, flags uint32) error {
	ret := C.virDomainSendProcessSignal(d.ptr, C.longlong(pid), C.uint(signum), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) InjectNMI(flags uint32) error {
	ret := C.virDomainInjectNMI(d.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) CoreDump(to string, flags DomainCoreDumpFlags) error {
	cto := C.CString(to)
	defer C.free(unsafe.Pointer(cto))

	ret := C.virDomainCoreDump(d.ptr, cto, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) CoreDumpWithFormat(to string, format DomainCoreDumpFormat, flags DomainCoreDumpFlags) error {
	if C.LIBVIR_VERSION_NUMBER < 1002003 {
		GetNotImplementedError("virDomainCoreDumpWithFormat")
	}
	cto := C.CString(to)
	defer C.free(unsafe.Pointer(cto))

	ret := C.virDomainCoreDumpWithFormatCompat(d.ptr, cto, C.uint(format), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) HasCurrentSnapshot(flags uint32) (bool, error) {
	result := C.virDomainHasCurrentSnapshot(d.ptr, C.uint(flags))
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (d *Domain) FSFreeze(mounts []string, flags uint32) error {
	if C.LIBVIR_VERSION_NUMBER < 1002005 {
		return GetNotImplementedError("virDomainFSFreeze")
	}
	cmounts := make([](*C.char), len(mounts))

	for i := 0; i < len(mounts); i++ {
		cmounts[i] = C.CString(mounts[i])
		defer C.free(unsafe.Pointer(cmounts[i]))
	}

	nmounts := len(mounts)
	ret := C.virDomainFSFreezeCompat(d.ptr, (**C.char)(unsafe.Pointer(&cmounts[0])), C.uint(nmounts), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) FSThaw(mounts []string, flags uint32) error {
	if C.LIBVIR_VERSION_NUMBER < 1002005 {
		return GetNotImplementedError("virDomainFSThaw")
	}
	cmounts := make([](*C.char), len(mounts))

	for i := 0; i < len(mounts); i++ {
		cmounts[i] = C.CString(mounts[i])
		defer C.free(unsafe.Pointer(cmounts[i]))
	}

	nmounts := len(mounts)
	ret := C.virDomainFSThawCompat(d.ptr, (**C.char)(unsafe.Pointer(&cmounts[0])), C.uint(nmounts), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) FSTrim(mount string, minimum uint64, flags uint32) error {
	cmount := C.CString(mount)
	defer C.free(unsafe.Pointer(cmount))

	ret := C.virDomainFSTrim(d.ptr, cmount, C.ulonglong(minimum), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

type DomainFSInfo struct {
	MountPoint string
	Name       string
	FSType     string
	DevAlias   []string
}

func (d *Domain) GetFSInfo(flags uint32) ([]DomainFSInfo, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002011 {
		return []DomainFSInfo{}, GetNotImplementedError("virDomainGetFSInfo")
	}
	var cfsinfolist **C.virDomainFSInfo

	ret := C.virDomainGetFSInfoCompat(d.ptr, (**C.virDomainFSInfoPtr)(unsafe.Pointer(&cfsinfolist)), C.uint(flags))
	if ret == -1 {
		return []DomainFSInfo{}, GetLastError()
	}

	fsinfo := make([]DomainFSInfo, int(ret))

	for i := 0; i < int(ret); i++ {
		cfsinfo := (*C.virDomainFSInfo)(*(**C.virDomainFSInfo)(unsafe.Pointer(uintptr(unsafe.Pointer(cfsinfolist)) + (unsafe.Sizeof(*cfsinfolist) * uintptr(i)))))

		aliases := make([]string, int(cfsinfo.ndevAlias))
		for j := 0; j < int(cfsinfo.ndevAlias); j++ {
			calias := (*C.char)(*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cfsinfo.devAlias)) + (unsafe.Sizeof(*cfsinfo) * uintptr(j)))))
			aliases[j] = C.GoString(calias)
		}
		fsinfo[i] = DomainFSInfo{
			MountPoint: C.GoString(cfsinfo.mountpoint),
			Name:       C.GoString(cfsinfo.name),
			FSType:     C.GoString(cfsinfo.fstype),
			DevAlias:   aliases,
		}

		C.virDomainFSInfoFreeCompat(cfsinfo)
	}
	C.free(unsafe.Pointer(cfsinfolist))

	return fsinfo, nil
}

func (d *Domain) PMSuspendForDuration(target NodeSuspendTarget, duration uint64, flags uint32) error {
	ret := C.virDomainPMSuspendForDuration(d.ptr, C.uint(target), C.ulonglong(duration), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) PMWakeup(flags uint32) error {
	ret := C.virDomainPMWakeup(d.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) AddIOThread(id uint, flags DomainModificationImpact) error {
	if C.LIBVIR_VERSION_NUMBER < 1002015 {
		return GetNotImplementedError("virDomainAddIOThread")
	}
	ret := C.virDomainAddIOThreadCompat(d.ptr, C.uint(id), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) DelIOThread(id uint, flags DomainModificationImpact) error {
	if C.LIBVIR_VERSION_NUMBER < 1002015 {
		return GetNotImplementedError("virDomainDelIOThread")
	}
	ret := C.virDomainDelIOThreadCompat(d.ptr, C.uint(id), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) GetEmulatorPinInfo(flags DomainModificationImpact) ([]bool, error) {
	var cnodeinfo C.virNodeInfo
	ret := C.virNodeGetInfo(C.virDomainGetConnect(d.ptr), &cnodeinfo)
	if ret == -1 {
		return []bool{}, GetLastError()
	}

	ncpus := cnodeinfo.nodes * cnodeinfo.sockets * cnodeinfo.cores * cnodeinfo.threads
	maplen := int((ncpus + 7) / 8)
	ccpumaps := make([]C.uchar, maplen)
	ret = C.virDomainGetEmulatorPinInfo(d.ptr, &ccpumaps[0], C.int(maplen), C.uint(flags))
	if ret == -1 {
		return []bool{}, GetLastError()
	}

	cpumaps := make([]bool, ncpus)
	for i := 0; i < int(ncpus); i++ {
		byte := i / 8
		bit := i % 8
		cpumaps[i] = (ccpumaps[byte] & (1 << uint(bit))) != 0
	}

	return cpumaps, nil
}

type DomainIOThreadInfo struct {
	IOThreadID uint
	CpuMap     []bool
}

func (d *Domain) GetIOThreadInfo(flags DomainModificationImpact) ([]DomainIOThreadInfo, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002014 {
		return []DomainIOThreadInfo{}, GetNotImplementedError("virDomaingetIOThreadInfo")
	}
	var cinfolist **C.virDomainIOThreadInfo

	ret := C.virDomainGetIOThreadInfoCompat(d.ptr, (**C.virDomainIOThreadInfoPtr)(unsafe.Pointer(&cinfolist)), C.uint(flags))
	if ret == -1 {
		return []DomainIOThreadInfo{}, GetLastError()
	}

	info := make([]DomainIOThreadInfo, int(ret))

	for i := 0; i < int(ret); i++ {
		cinfo := (*(**C.virDomainIOThreadInfo)(unsafe.Pointer(uintptr(unsafe.Pointer(cinfolist)) + (unsafe.Sizeof(*cinfolist) * uintptr(i)))))

		ncpus := int(cinfo.cpumaplen * 8)
		cpumap := make([]bool, ncpus)
		for j := 0; j < ncpus; j++ {
			byte := j / 8
			bit := j % 8

			cpumapbyte := *(*C.uchar)(unsafe.Pointer(uintptr(unsafe.Pointer(cinfo.cpumap)) + (unsafe.Sizeof(*cinfo.cpumap) * uintptr(byte))))
			cpumap[j] = (cpumapbyte & (1 << uint(bit))) != 0
		}

		info[i] = DomainIOThreadInfo{
			IOThreadID: uint(cinfo.iothread_id),
			CpuMap:     cpumap,
		}

		C.virDomainIOThreadInfoFreeCompat(cinfo)
	}
	C.free(unsafe.Pointer(cinfolist))

	return info, nil
}

func (d *Domain) GetVcpuPinInfo(flags DomainModificationImpact) ([][]bool, error) {
	var cnodeinfo C.virNodeInfo
	ret := C.virNodeGetInfo(C.virDomainGetConnect(d.ptr), &cnodeinfo)
	if ret == -1 {
		return [][]bool{}, GetLastError()
	}

	var cdominfo C.virDomainInfo
	ret = C.virDomainGetInfo(d.ptr, &cdominfo)
	if ret == -1 {
		return [][]bool{}, GetLastError()
	}

	nvcpus := int(cdominfo.nrVirtCpu)
	npcpus := int(cnodeinfo.nodes * cnodeinfo.sockets * cnodeinfo.cores * cnodeinfo.threads)
	maplen := ((npcpus + 7) / 8)
	ccpumaps := make([]C.uchar, maplen*nvcpus)

	ret = C.virDomainGetVcpuPinInfo(d.ptr, C.int(nvcpus), &ccpumaps[0], C.int(maplen), C.uint(flags))
	if ret == -1 {
		return [][]bool{}, GetLastError()
	}

	cpumaps := make([][]bool, nvcpus)
	for i := 0; i < nvcpus; i++ {
		cpumaps[i] = make([]bool, npcpus)
		for j := 0; j < npcpus; j++ {
			byte := (i * maplen) + (j / 8)
			bit := j % 8

			if (ccpumaps[byte] & (1 << uint(bit))) != 0 {
				cpumaps[i][j] = true
			}
		}
	}

	return cpumaps, nil
}

func (d *Domain) PinEmulator(cpumap []bool, flags DomainModificationImpact) error {

	maplen := (len(cpumap) + 7) / 8
	ccpumaps := make([]C.uchar, maplen)
	for i := 0; i < len(cpumap); i++ {
		if cpumap[i] {
			byte := i / 8
			bit := i % 8

			ccpumaps[byte] |= (1 << uint(bit))
		}
	}

	ret := C.virDomainPinEmulator(d.ptr, &ccpumaps[0], C.int(maplen), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) PinIOThread(iothreadid uint, cpumap []bool, flags DomainModificationImpact) error {
	if C.LIBVIR_VERSION_NUMBER < 1002014 {
		return GetNotImplementedError("virDomainPinIOThread")
	}

	maplen := (len(cpumap) + 7) / 8
	ccpumaps := make([]C.uchar, maplen)
	for i := 0; i < len(cpumap); i++ {
		if cpumap[i] {
			byte := i / 8
			bit := i % 8

			ccpumaps[byte] |= (1 << uint(bit))
		}
	}

	ret := C.virDomainPinIOThreadCompat(d.ptr, C.uint(iothreadid), &ccpumaps[0], C.int(maplen), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) OpenChannel(name string, stream *Stream, flags DomainChannelFlags) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.virDomainOpenChannel(d.ptr, cname, stream.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) OpenConsole(devname string, stream *Stream, flags DomainConsoleFlags) error {
	var cdevname *C.char
	if devname != "" {
		cdevname = C.CString(devname)
		defer C.free(unsafe.Pointer(cdevname))
	}

	ret := C.virDomainOpenConsole(d.ptr, cdevname, stream.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) OpenGraphics(idx uint, file os.File, flags DomainOpenGraphicsFlags) error {
	ret := C.virDomainOpenGraphics(d.ptr, C.uint(idx), C.int(file.Fd()), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) OpenGraphicsFD(idx uint, flags DomainOpenGraphicsFlags) (*os.File, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002008 {
		return nil, GetNotImplementedError("virDomainOpenGraphicsFD")
	}
	ret := C.virDomainOpenGraphicsFDCompat(d.ptr, C.uint(idx), C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	return os.NewFile(uintptr(ret), "graphics"), nil
}

func (d *Domain) CreateSnapshotXML(xml string, flags DomainSnapshotCreateFlags) (*DomainSnapshot, error) {
	cXml := C.CString(xml)
	defer C.free(unsafe.Pointer(cXml))
	result := C.virDomainSnapshotCreateXML(d.ptr, cXml, C.uint(flags))
	if result == nil {
		return nil, GetLastError()
	}
	return &DomainSnapshot{ptr: result}, nil
}

func (d *Domain) Save(destFile string) error {
	cPath := C.CString(destFile)
	defer C.free(unsafe.Pointer(cPath))
	result := C.virDomainSave(d.ptr, cPath)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (d *Domain) SaveFlags(destFile string, destXml string, flags DomainSaveRestoreFlags) error {
	cDestFile := C.CString(destFile)
	cDestXml := C.CString(destXml)
	defer C.free(unsafe.Pointer(cDestXml))
	defer C.free(unsafe.Pointer(cDestFile))
	result := C.virDomainSaveFlags(d.ptr, cDestFile, cDestXml, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

type DomainGuestVcpus struct {
	Vcpus      []bool
	Online     []bool
	Offlinable []bool
}

func getDomainGuestVcpusParametersFieldInfo(VcpusSet *bool, Vcpus *string, OnlineSet *bool, Online *string, OfflinableSet *bool, Offlinable *string) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		"vcpus": typedParamsFieldInfo{
			set: VcpusSet,
			s:   Vcpus,
		},
		"online": typedParamsFieldInfo{
			set: OnlineSet,
			s:   Online,
		},
		"offlinable": typedParamsFieldInfo{
			set: OfflinableSet,
			s:   Offlinable,
		},
	}
}

func parseCPUString(cpumapstr string) ([]bool, error) {
	pieces := strings.Split(cpumapstr, ",")
	var cpumap []bool
	for _, piece := range pieces {
		if len(piece) < 1 {
			return []bool{}, fmt.Errorf("Malformed cpu map string %s", cpumapstr)
		}
		invert := false
		if piece[0] == '^' {
			invert = true
			piece = piece[1:]
		}
		pair := strings.Split(piece, "-")
		var start, end int
		var err error
		if len(pair) == 1 {
			start, err = strconv.Atoi(pair[0])
			if err != nil {
				return []bool{}, fmt.Errorf("Malformed cpu map string %s", cpumapstr)
			}
			end, err = strconv.Atoi(pair[0])
			if err != nil {
				return []bool{}, fmt.Errorf("Malformed cpu map string %s", cpumapstr)
			}
		} else if len(pair) == 2 {
			start, err = strconv.Atoi(pair[0])
			if err != nil {
				return []bool{}, fmt.Errorf("Malformed cpu map string %s", cpumapstr)
			}
			end, err = strconv.Atoi(pair[1])
			if err != nil {
				return []bool{}, fmt.Errorf("Malformed cpu map string %s", cpumapstr)
			}
		} else {
			return []bool{}, fmt.Errorf("Malformed cpu map string %s", cpumapstr)
		}
		if start > end {
			return []bool{}, fmt.Errorf("Malformed cpu map string %s", cpumapstr)
		}
		if (end + 1) > len(cpumap) {
			newcpumap := make([]bool, end+1)
			copy(newcpumap, cpumap)
			cpumap = newcpumap
		}

		for i := start; i <= end; i++ {
			if invert {
				cpumap[i] = false
			} else {
				cpumap[i] = true
			}
		}
	}

	return cpumap, nil
}

func (d *Domain) GetGuestVcpus(flags uint32) (*DomainGuestVcpus, error) {
	if C.LIBVIR_VERSION_NUMBER < 2000000 {
		return nil, GetNotImplementedError("virDomainGetGuestVcpus")
	}

	var VcpusSet, OnlineSet, OfflinableSet bool
	var VcpusStr, OnlineStr, OfflinableStr string
	info := getDomainGuestVcpusParametersFieldInfo(&VcpusSet, &VcpusStr, &OnlineSet, &OnlineStr, &OfflinableSet, &OfflinableStr)

	var cparams C.virTypedParameterPtr
	var nparams C.uint
	ret := C.virDomainGetGuestVcpusCompat(d.ptr, &cparams, &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	defer C.virTypedParamsFree(cparams, C.int(nparams))

	_, err := typedParamsUnpackLen(cparams, int(nparams), info)
	if err != nil {
		return nil, err
	}

	return &DomainGuestVcpus{}, nil
}

func (d *Domain) SetGuestVcpus(cpus []bool, state bool, flags uint32) error {
	if C.LIBVIR_VERSION_NUMBER < 2000000 {
		return GetNotImplementedError("virDomainSetGuestVcpus")
	}

	cpumap := ""
	for i := 0; i < len(cpus); i++ {
		if cpus[i] {
			if cpumap == "" {
				cpumap = string(i)
			} else {
				cpumap += "," + string(i)
			}
		}
	}

	var cstate C.int
	if state {
		cstate = 1
	} else {
		cstate = 0
	}
	ccpumap := C.CString(cpumap)
	defer C.free(unsafe.Pointer(ccpumap))
	ret := C.virDomainSetGuestVcpusCompat(d.ptr, ccpumap, cstate, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) SetVcpu(cpus []bool, state bool, flags uint32) error {
	if C.LIBVIR_VERSION_NUMBER < 3001000 {
		return GetNotImplementedError("virDomainSetVcpu")
	}

	cpumap := ""
	for i := 0; i < len(cpus); i++ {
		if cpus[i] {
			if cpumap == "" {
				cpumap = string(i)
			} else {
				cpumap += "," + string(i)
			}
		}
	}

	var cstate C.int
	if state {
		cstate = 1
	} else {
		cstate = 0
	}
	ccpumap := C.CString(cpumap)
	defer C.free(unsafe.Pointer(ccpumap))
	ret := C.virDomainSetVcpuCompat(d.ptr, ccpumap, cstate, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (d *Domain) SetBlockThreshold(dev string, threshold uint64, flags uint32) error {
	if C.LIBVIR_VERSION_NUMBER < 3002000 {
		return GetNotImplementedError("virDomainSetBlockThreshold")
	}

	cdev := C.CString(dev)
	defer C.free(unsafe.Pointer(cdev))
	ret := C.virDomainSetBlockThresholdCompat(d.ptr, cdev, C.ulonglong(threshold), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}
