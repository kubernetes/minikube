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

#ifndef LIBVIRT_GO_DOMAIN_COMPAT_H__
#define LIBVIRT_GO_DOMAIN_COMPAT_H__

/* 1.2.2 */

#ifndef VIR_DOMAIN_BLKIO_DEVICE_READ_IOPS
#define VIR_DOMAIN_BLKIO_DEVICE_READ_IOPS "device_read_iops_sec"
#endif

#ifndef VIR_DOMAIN_BLKIO_DEVICE_WRITE_IOPS
#define VIR_DOMAIN_BLKIO_DEVICE_WRITE_IOPS "device_write_iops_sec"
#endif

#ifndef VIR_DOMAIN_BLKIO_DEVICE_READ_BPS
#define VIR_DOMAIN_BLKIO_DEVICE_READ_BPS "device_read_bytes_sec"
#endif

#ifndef VIR_DOMAIN_BLKIO_DEVICE_WRITE_BPS
#define VIR_DOMAIN_BLKIO_DEVICE_WRITE_BPS "device_write_bytes_sec"
#endif


/* 1.2.3 */

#ifndef VIR_DOMAIN_CORE_DUMP_FORMAT_RAW
#define VIR_DOMAIN_CORE_DUMP_FORMAT_RAW 0
#endif

#ifndef VIR_DOMAIN_CORE_DUMP_FORMAT_KDUMP_ZLIB
#define VIR_DOMAIN_CORE_DUMP_FORMAT_KDUMP_ZLIB 1
#endif

#ifndef VIR_DOMAIN_CORE_DUMP_FORMAT_KDUMP_LZO
#define VIR_DOMAIN_CORE_DUMP_FORMAT_KDUMP_LZO 2
#endif

#ifndef VIR_DOMAIN_CORE_DUMP_FORMAT_KDUMP_SNAPPY
#define VIR_DOMAIN_CORE_DUMP_FORMAT_KDUMP_SNAPPY 3
#endif

#ifndef VIR_MIGRATE_AUTO_CONVERGE
#define VIR_MIGRATE_AUTO_CONVERGE 1 << 13
#endif

int virDomainCoreDumpWithFormatCompat(virDomainPtr domain,
				      const char *to,
				      unsigned int dumpformat,
				      unsigned int flags);


/* 1.2.5 */

#ifndef VIR_DOMAIN_REBOOT_PARAVIRT
#define VIR_DOMAIN_REBOOT_PARAVIRT 1 << 4
#endif

#ifndef VIR_DOMAIN_SHUTDOWN_PARAVIRT
#define VIR_DOMAIN_SHUTDOWN_PARAVIRT 1 << 4
#endif

#ifndef VIR_DOMAIN_TIME_SYNC
#define VIR_DOMAIN_TIME_SYNC 1 << 0
#endif

int virDomainGetTimeCompat(virDomainPtr dom,
			   long long *seconds,
			   unsigned int *nseconds,
			   unsigned int flags);

int virDomainSetTimeCompat(virDomainPtr dom,
			   long long seconds,
			   unsigned int nseconds,
			   unsigned int flags);

int virDomainFSFreezeCompat(virDomainPtr dom,
			    const char **mountpoints,
			    unsigned int nmountpoints,
			    unsigned int flags);

int virDomainFSThawCompat(virDomainPtr dom,
			  const char **mountpoints,
			  unsigned int nmountpoints,
			  unsigned int flags);


/* 1.2.6 */

#ifndef VIR_DOMAIN_BLOCK_COMMIT_ACTIVE
#define VIR_DOMAIN_BLOCK_COMMIT_ACTIVE 1 << 2
#endif

#ifndef VIR_DOMAIN_BLOCK_JOB_TYPE_ACTIVE_COMMIT
#define VIR_DOMAIN_BLOCK_JOB_TYPE_ACTIVE_COMMIT 4
#endif

#ifndef VIR_DOMAIN_EVENT_ID_BLOCK_JOB_2
#define VIR_DOMAIN_EVENT_ID_BLOCK_JOB_2 16
#endif


/* 1.2.7 */

#ifndef VIR_DOMAIN_BLOCK_COMMIT_RELATIVE
#define VIR_DOMAIN_BLOCK_COMMIT_RELATIVE 1 << 3
#endif

#ifndef VIR_DOMAIN_BLOCK_REBASE_RELATIVE
#define VIR_DOMAIN_BLOCK_REBASE_RELATIVE 1 << 4
#endif


/* 1.2.8 */

#ifndef VIR_DOMAIN_BLOCK_COPY_SHALLOW
#define VIR_DOMAIN_BLOCK_COPY_SHALLOW 1 << 0
#endif

#ifndef VIR_DOMAIN_BLOCK_COPY_REUSE_EXT
#define VIR_DOMAIN_BLOCK_COPY_REUSE_EXT 1 << 1
#endif

#ifndef VIR_DOMAIN_BLOCK_COPY_BANDWIDTH
#define VIR_DOMAIN_BLOCK_COPY_BANDWIDTH "bandwidth"
#endif

#ifndef VIR_DOMAIN_BLOCK_COPY_GRANULARITY
#define VIR_DOMAIN_BLOCK_COPY_GRANULARITY "granularity"
#endif

#ifndef VIR_DOMAIN_BLOCK_COPY_BUF_SIZE
#define VIR_DOMAIN_BLOCK_COPY_BUF_SIZE "buf-size"
#endif

#ifndef VIR_DOMAIN_STATS_STATE
#define VIR_DOMAIN_STATS_STATE 1 << 0
#endif

int virDomainBlockCopyCompat(virDomainPtr dom, const char *disk,
			     const char *destxml,
			     virTypedParameterPtr params,
			     int nparams,
			     unsigned int flags);

int virDomainOpenGraphicsFDCompat(virDomainPtr dom,
				  unsigned int idx,
				  unsigned int flags);


/* 1.2.9 */

#ifndef VIR_DOMAIN_BLOCK_COMMIT_BANDWIDTH_BYTES
#define VIR_DOMAIN_BLOCK_COMMIT_BANDWIDTH_BYTES 1 << 4
#endif

#ifndef VIR_DOMAIN_BLOCK_JOB_INFO_BANDWIDTH_BYTES
#define VIR_DOMAIN_BLOCK_JOB_INFO_BANDWIDTH_BYTES 1 << 0
#endif

#ifndef VIR_DOMAIN_BLOCK_JOB_SPEED_BANDWIDTH_BYTES
#define VIR_DOMAIN_BLOCK_JOB_SPEED_BANDWIDTH_BYTES 1 << 0
#endif

#ifndef VIR_DOMAIN_BLOCK_PULL_BANDWIDTH_BYTES
#define VIR_DOMAIN_BLOCK_PULL_BANDWIDTH_BYTES 1 << 6
#endif

#ifndef VIR_DOMAIN_BLOCK_REBASE_COPY_DEV
#define VIR_DOMAIN_BLOCK_REBASE_COPY_DEV 1 << 5
#endif

#ifndef VIR_DOMAIN_BLOCK_REBASE_BANDWIDTH_BYTES
#define VIR_DOMAIN_BLOCK_REBASE_BANDWIDTH_BYTES 1 << 6
#endif

#ifndef VIR_DOMAIN_JOB_DISK_BPS
#define VIR_DOMAIN_JOB_DISK_BPS "disk_bps"
#endif

#ifndef VIR_DOMAIN_JOB_MEMORY_BPS
#define VIR_DOMAIN_JOB_MEMORY_BPS "memory_bps"
#endif

#ifndef VIR_DOMAIN_JOB_SETUP_TIME
#define VIR_DOMAIN_JOB_SETUP_TIME "setup_time"
#endif

#ifndef VIR_DOMAIN_JOB_STATS_COMPLETED
#define VIR_DOMAIN_JOB_STATS_COMPLETED 1 << 0
#endif

#ifndef VIR_DOMAIN_STATS_CPU_TOTAL
#define VIR_DOMAIN_STATS_CPU_TOTAL 1 << 1
#endif

#ifndef VIR_DOMAIN_STATS_BALLOON
#define VIR_DOMAIN_STATS_BALLOON 1 << 2
#endif

#ifndef VIR_DOMAIN_STATS_VCPU
#define VIR_DOMAIN_STATS_VCPU 1 << 3
#endif

#ifndef VIR_DOMAIN_STATS_INTERFACE
#define VIR_DOMAIN_STATS_INTERFACE 1 << 4
#endif

#ifndef VIR_DOMAIN_STATS_BLOCK
#define VIR_DOMAIN_STATS_BLOCK 1 << 5
#endif

#ifndef VIR_DOMAIN_UNDEFINE_NVRAM
#define VIR_DOMAIN_UNDEFINE_NVRAM 1 << 2
#endif

#ifndef VIR_MIGRATE_RDMA_PIN_ALL
#define VIR_MIGRATE_RDMA_PIN_ALL 1 << 14
#endif

#ifndef VIR_DOMAIN_EVENT_ID_TUNABLE
#define VIR_DOMAIN_EVENT_ID_TUNABLE 17
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_DISK
#define VIR_DOMAIN_TUNABLE_BLKDEV_DISK "blkdeviotune.disk"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_BYTES_SEC
#define VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_BYTES_SEC "blkdeviotune.total_bytes_sec"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_READ_BYTES_SEC
#define VIR_DOMAIN_TUNABLE_BLKDEV_READ_BYTES_SEC "blkdeviotune.read_bytes_sec"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_BYTES_SEC
#define VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_BYTES_SEC "blkdeviotune.write_bytes_sec"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_IOPS_SEC
#define VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_IOPS_SEC "blkdeviotune.total_iops_sec"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_READ_IOPS_SEC
#define VIR_DOMAIN_TUNABLE_BLKDEV_READ_IOPS_SEC "blkdeviotune.read_iops_sec"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_IOPS_SEC
#define VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_IOPS_SEC "blkdeviotune.write_iops_sec"
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_CPU_SHARES
#define VIR_DOMAIN_TUNABLE_CPU_CPU_SHARES "cputune.cpu_shares"
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_EMULATORPIN
#define VIR_DOMAIN_TUNABLE_CPU_EMULATORPIN "cputune.emulatorpin"
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_EMULATOR_PERIOD
#define VIR_DOMAIN_TUNABLE_CPU_EMULATOR_PERIOD "cputune.emulator_period"
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_EMULATOR_QUOTA
#define VIR_DOMAIN_TUNABLE_CPU_EMULATOR_QUOTA "cputune.emulator_quota"
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_VCPU_PERIOD
#define VIR_DOMAIN_TUNABLE_CPU_VCPU_PERIOD "cputune.vcpu_period"
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_VCPU_QUOTA
#define VIR_DOMAIN_TUNABLE_CPU_VCPU_QUOTA "cputune.vcpu_quota"
#endif



/* 1.2.11 */

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_BYTES_SEC_MAX
#define VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_BYTES_SEC_MAX "blkdeviotune.total_bytes_sec_max"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_READ_BYTES_SEC_MAX
#define VIR_DOMAIN_TUNABLE_BLKDEV_READ_BYTES_SEC_MAX "blkdeviotune.read_bytes_sec_max"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_BYTES_SEC_MAX
#define VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_BYTES_SEC_MAX "blkdeviotune.write_bytes_sec_max"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_IOPS_SEC_MAX
#define VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_IOPS_SEC_MAX "blkdeviotune.total_iops_sec_max"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_READ_IOPS_SEC_MAX
#define VIR_DOMAIN_TUNABLE_BLKDEV_READ_IOPS_SEC_MAX "blkdeviotune.read_iops_sec_max"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_IOPS_SEC_MAX
#define VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_IOPS_SEC_MAX "blkdeviotune.write_iops_sec_max"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_SIZE_IOPS_SEC
#define VIR_DOMAIN_TUNABLE_BLKDEV_SIZE_IOPS_SEC "blkdeviotune.size_iops_sec"
#endif

#ifndef VIR_DOMAIN_EVENT_ID_AGENT_LIFECYCLE
#define VIR_DOMAIN_EVENT_ID_AGENT_LIFECYCLE 18
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_BYTES_SEC_MAX
#define VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_BYTES_SEC_MAX "total_bytes_sec_max"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_READ_BYTES_SEC_MAX
#define VIR_DOMAIN_BLOCK_IOTUNE_READ_BYTES_SEC_MAX "read_bytes_sec_max"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_WRITE_BYTES_SEC_MAX
#define VIR_DOMAIN_BLOCK_IOTUNE_WRITE_BYTES_SEC_MAX "write_bytes_sec_max"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_IOPS_SEC_MAX
#define VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_IOPS_SEC_MAX "total_iops_sec_max"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_READ_IOPS_SEC_MAX
#define VIR_DOMAIN_BLOCK_IOTUNE_READ_IOPS_SEC_MAX "read_iops_sec_max"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_WRITE_IOPS_SEC_MAX
#define VIR_DOMAIN_BLOCK_IOTUNE_WRITE_IOPS_SEC_MAX "write_iops_sec_max"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_SIZE_IOPS_SEC
#define VIR_DOMAIN_BLOCK_IOTUNE_SIZE_IOPS_SEC "size_iops_sec"
#endif

#if LIBVIR_VERSION_NUMBER < 1002011
typedef struct _virDomainFSInfo virDomainFSInfo;
typedef virDomainFSInfo *virDomainFSInfoPtr;
struct _virDomainFSInfo {
    char *mountpoint; /* path to mount point */
    char *name;       /* device name in the guest (e.g. "sda1") */
    char *fstype;     /* filesystem type */
    size_t ndevAlias; /* number of elements in devAlias */
    char **devAlias;  /* array of disk device aliases */
};
#endif

void virDomainFSInfoFreeCompat(virDomainFSInfoPtr info);

int virDomainGetFSInfoCompat(virDomainPtr dom,
			     virDomainFSInfoPtr **info,
			     unsigned int flags);


/* 1.2.12 */

#ifndef VIR_DOMAIN_DEFINE_VALIDATE
#define VIR_DOMAIN_DEFINE_VALIDATE 1 << 0
#endif

#ifndef VIR_DOMAIN_START_VALIDATE
#define VIR_DOMAIN_START_VALIDATE 1 << 4
#endif


/* 1.2.14 */

#ifndef VIR_DOMAIN_CONTROL_ERROR_REASON_NONE
#define VIR_DOMAIN_CONTROL_ERROR_REASON_NONE 0
#endif

#ifndef VIR_DOMAIN_CONTROL_ERROR_REASON_UNKNOWN
#define VIR_DOMAIN_CONTROL_ERROR_REASON_UNKNOWN 1
#endif

#ifndef VIR_DOMAIN_CONTROL_ERROR_REASON_MONITOR
#define VIR_DOMAIN_CONTROL_ERROR_REASON_MONITOR 2
#endif

#ifndef VIR_DOMAIN_CONTROL_ERROR_REASON_INTERNAL
#define VIR_DOMAIN_CONTROL_ERROR_REASON_INTERNAL 3
#endif

#ifndef VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE
#define VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE 0
#endif

#ifndef VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT
#define VIR_DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT 1
#endif

#ifndef VIR_DOMAIN_PAUSED_STARTING_UP
#define VIR_DOMAIN_PAUSED_STARTING_UP 11
#endif

#if LIBVIR_VERSION_NUMBER < 1002014
typedef struct _virDomainIOThreadInfo virDomainIOThreadInfo;
typedef virDomainIOThreadInfo *virDomainIOThreadInfoPtr;
struct _virDomainIOThreadInfo {
    unsigned int iothread_id;          /* IOThread ID */
    unsigned char *cpumap;             /* CPU map for thread. A pointer to an */
                                       /* array of real CPUs (in 8-bit bytes) */
    int cpumaplen;                     /* cpumap size */
};

typedef struct _virDomainInterfaceIPAddress virDomainIPAddress;
typedef virDomainIPAddress *virDomainIPAddressPtr;
struct _virDomainInterfaceIPAddress {
    int type;                /* virIPAddrType */
    char *addr;              /* IP address */
    unsigned int prefix;     /* IP address prefix */
};

typedef struct _virDomainInterface virDomainInterface;
typedef virDomainInterface *virDomainInterfacePtr;
struct _virDomainInterface {
    char *name;                     /* interface name */
    char *hwaddr;                   /* hardware address, may be NULL */
    unsigned int naddrs;            /* number of items in @addrs */
    virDomainIPAddressPtr addrs;    /* array of IP addresses */
};
#endif

int virDomainInterfaceAddressesCompat(virDomainPtr dom,
				      virDomainInterfacePtr **ifaces,
				      unsigned int source,
				      unsigned int flags);

void virDomainInterfaceFreeCompat(virDomainInterfacePtr iface);

void virDomainIOThreadInfoFreeCompat(virDomainIOThreadInfoPtr info);

int virDomainGetIOThreadInfoCompat(virDomainPtr domain,
				   virDomainIOThreadInfoPtr **info,
				   unsigned int flags);
int virDomainPinIOThreadCompat(virDomainPtr domain,
			       unsigned int iothread_id,
			       unsigned char *cpumap,
			       int maplen,
			       unsigned int flags);


/* 1.2.15 */

#ifndef VIR_DOMAIN_JOB_DOWNTIME_NET
#define VIR_DOMAIN_JOB_DOWNTIME_NET "downtime_net"
#endif

#ifndef VIR_DOMAIN_JOB_TIME_ELAPSED_NET
#define VIR_DOMAIN_JOB_TIME_ELAPSED_NET "time_elapsed_net"
#endif

#ifndef VIR_DOMAIN_EVENT_ID_DEVICE_ADDED
#define VIR_DOMAIN_EVENT_ID_DEVICE_ADDED 19
#endif

int virDomainAddIOThreadCompat(virDomainPtr domain,
			       unsigned int iothread_id,
			       unsigned int flags);
int virDomainDelIOThreadCompat(virDomainPtr domain,
			       unsigned int iothread_id,
			       unsigned int flags);


/* 1.2.16 */

#ifndef VIR_DOMAIN_PASSWORD_ENCRYPTED
#define VIR_DOMAIN_PASSWORD_ENCRYPTED 1 << 0
#endif

int virDomainSetUserPasswordCompat(virDomainPtr dom,
				   const char *user,
				   const char *password,
				   unsigned int flags);


/* 1.2.17 */

#ifndef VIR_DOMAIN_EVENT_WATCHDOG_INJECTNMI
#define VIR_DOMAIN_EVENT_WATCHDOG_INJECTNMI 6
#endif

#ifndef VIR_MIGRATE_PARAM_MIGRATE_DISKS
#define VIR_MIGRATE_PARAM_MIGRATE_DISKS "migrate_disks"
#endif


/* 1.2.19 */

#ifndef VIR_DOMAIN_BANDWIDTH_IN_FLOOR
#define VIR_DOMAIN_BANDWIDTH_IN_FLOOR "inbound.floor"
#endif

#ifndef VIR_DOMAIN_EVENT_DEFINED_RENAMED
#define VIR_DOMAIN_EVENT_DEFINED_RENAMED 2
#endif

#ifndef VIR_DOMAIN_EVENT_UNDEFINED_RENAMED
#define VIR_DOMAIN_EVENT_UNDEFINED_RENAMED 1
#endif

int virDomainRenameCompat(virDomainPtr dom,
			  const char *new_name,
			  unsigned int flags);


/* 1.3.1 */

#ifndef VIR_DOMAIN_JOB_MEMORY_DIRTY_RATE
#define VIR_DOMAIN_JOB_MEMORY_DIRTY_RATE "memory_dirty_rate"
#endif

#ifndef VIR_DOMAIN_JOB_MEMORY_ITERATION
#define VIR_DOMAIN_JOB_MEMORY_ITERATION "memory_iteration"
#endif


/* 1.3.2 */

#ifndef VIR_DOMAIN_EVENT_ID_MIGRATION_ITERATION
#define VIR_DOMAIN_EVENT_ID_MIGRATION_ITERATION 20
#endif


/* 1.3.3 */

#ifndef VIR_DOMAIN_EVENT_DEFINED_FROM_SNAPSHOT
#define VIR_DOMAIN_EVENT_DEFINED_FROM_SNAPSHOT 3
#endif

#ifndef VIR_DOMAIN_EVENT_RESUMED_POSTCOPY
#define VIR_DOMAIN_EVENT_RESUMED_POSTCOPY 3
#endif

#ifndef VIR_DOMAIN_EVENT_SUSPENDED_POSTCOPY
#define VIR_DOMAIN_EVENT_SUSPENDED_POSTCOPY 7
#endif

#ifndef VIR_DOMAIN_EVENT_SUSPENDED_POSTCOPY_FAILED
#define VIR_DOMAIN_EVENT_SUSPENDED_POSTCOPY_FAILED 8
#endif

#ifndef VIR_DOMAIN_PAUSED_POSTCOPY
#define VIR_DOMAIN_PAUSED_POSTCOPY 12
#endif

#ifndef VIR_DOMAIN_PAUSED_POSTCOPY_FAILED
#define VIR_DOMAIN_PAUSED_POSTCOPY_FAILED 13
#endif

#ifndef VIR_DOMAIN_RUNNING_POSTCOPY
#define VIR_DOMAIN_RUNNING_POSTCOPY 10
#endif

#ifndef VIR_DOMAIN_SCHEDULER_GLOBAL_PERIOD
#define VIR_DOMAIN_SCHEDULER_GLOBAL_PERIOD "global_period"
#endif

#ifndef VIR_DOMAIN_SCHEDULER_GLOBAL_QUOTA
#define VIR_DOMAIN_SCHEDULER_GLOBAL_QUOTA "global_quota"
#endif

#ifndef VIR_DOMAIN_STATS_PERF
#define VIR_DOMAIN_STATS_PERF (1 << 6)
#endif

#ifndef VIR_MIGRATE_PARAM_DISKS_PORT
#define VIR_MIGRATE_PARAM_DISKS_PORT "disks_port"
#endif

#ifndef VIR_PERF_PARAM_CMT
#define VIR_PERF_PARAM_CMT "cmt"
#endif

#ifndef VIR_MIGRATE_POSTCOPY
#define VIR_MIGRATE_POSTCOPY (1 << 15)
#endif

#ifndef VIR_DOMAIN_EVENT_ID_JOB_COMPLETED
#define VIR_DOMAIN_EVENT_ID_JOB_COMPLETED 21
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_GLOBAL_PERIOD
#define VIR_DOMAIN_TUNABLE_CPU_GLOBAL_PERIOD "cputune.global_period"
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_GLOBAL_QUOTA
#define VIR_DOMAIN_TUNABLE_CPU_GLOBAL_QUOTA "cputune.global_quota"
#endif

int virDomainGetPerfEventsCompat(virDomainPtr dom,
				 virTypedParameterPtr *params,
				 int *nparams,
				 unsigned int flags);
int virDomainSetPerfEventsCompat(virDomainPtr dom,
				 virTypedParameterPtr params,
				 int nparams,
				 unsigned int flags);
int virDomainMigrateStartPostCopyCompat(virDomainPtr domain,
					unsigned int flags);


/* 1.3.4 */

#ifndef VIR_MIGRATE_PARAM_COMPRESSION
#define VIR_MIGRATE_PARAM_COMPRESSION  "compression"
#endif

#ifndef VIR_MIGRATE_PARAM_COMPRESSION_MT_THREADS
#define VIR_MIGRATE_PARAM_COMPRESSION_MT_THREADS "compression.mt.threads"
#endif

#ifndef VIR_MIGRATE_PARAM_COMPRESSION_MT_DTHREADS
#define VIR_MIGRATE_PARAM_COMPRESSION_MT_DTHREADS "compression.mt.dthreads"
#endif

#ifndef VIR_MIGRATE_PARAM_COMPRESSION_MT_LEVEL
#define VIR_MIGRATE_PARAM_COMPRESSION_MT_LEVEL "compression.mt.level"
#endif

#ifndef VIR_MIGRATE_PARAM_COMPRESSION_XBZRLE_CACHE
#define VIR_MIGRATE_PARAM_COMPRESSION_XBZRLE_CACHE "compression.xbzrle.cache"
#endif

#ifndef VIR_MIGRATE_PARAM_PERSIST_XML
#define VIR_MIGRATE_PARAM_PERSIST_XML "persistent_xml"
#endif

#ifndef VIR_DOMAIN_EVENT_ID_DEVICE_REMOVAL_FAILED
#define VIR_DOMAIN_EVENT_ID_DEVICE_REMOVAL_FAILED 22
#endif


/* 1.3.5 */

#ifndef VIR_PERF_PARAM_MBML
#define VIR_PERF_PARAM_MBML "mbml"
#endif

#ifndef VIR_PERF_PARAM_MBMT
#define VIR_PERF_PARAM_MBMT "mbmt"
#endif


/* 2.0.0 */

#ifndef VIR_DOMAIN_JOB_AUTO_CONVERGE_THROTTLE
#define VIR_DOMAIN_JOB_AUTO_CONVERGE_THROTTLE "auto_converge_throttle"
#endif

#ifndef VIR_MIGRATE_PARAM_AUTO_CONVERGE_INITIAL
#define VIR_MIGRATE_PARAM_AUTO_CONVERGE_INITIAL "auto_converge.initial"
#endif

#ifndef VIR_MIGRATE_PARAM_AUTO_CONVERGE_INCREMENT
#define VIR_MIGRATE_PARAM_AUTO_CONVERGE_INCREMENT "auto_converge.increment"
#endif

int virDomainGetGuestVcpusCompat(virDomainPtr domain,
				 virTypedParameterPtr *params,
				 unsigned int *nparams,
				 unsigned int flags);

int virDomainSetGuestVcpusCompat(virDomainPtr domain,
				 const char *cpumap,
				 int state,
				 unsigned int flags);


/* 2.1.0 */

#ifndef VIR_DOMAIN_MEMORY_STAT_USABLE
#define VIR_DOMAIN_MEMORY_STAT_USABLE 8
#endif

#ifndef VIR_DOMAIN_MEMORY_STAT_LAST_UPDATE
#define VIR_DOMAIN_MEMORY_STAT_LAST_UPDATE 9
#endif

/* 2.2.0 */

#ifndef VIR_DOMAIN_SCHEDULER_IOTHREAD_PERIOD
#define VIR_DOMAIN_SCHEDULER_IOTHREAD_PERIOD "iothread_period"
#endif

#ifndef VIR_DOMAIN_SCHEDULER_IOTHREAD_QUOTA
#define VIR_DOMAIN_SCHEDULER_IOTHREAD_QUOTA "iothread_quota"
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_IOTHREAD_PERIOD
#define VIR_DOMAIN_TUNABLE_CPU_IOTHREAD_PERIOD "cputune.iothread_period"
#endif

#ifndef VIR_DOMAIN_TUNABLE_CPU_IOTHREAD_QUOTA
# define VIR_DOMAIN_TUNABLE_CPU_IOTHREAD_QUOTA "cputune.iothread_quota"
#endif


/* 2.3.0 */

#ifndef VIR_DOMAIN_UNDEFINE_KEEP_NVRAM
#define VIR_DOMAIN_UNDEFINE_KEEP_NVRAM (1 << 3)
#endif

#ifndef VIR_PERF_PARAM_CACHE_MISSES
#define VIR_PERF_PARAM_CACHE_MISSES "cache_misses"
#endif

#ifndef VIR_PERF_PARAM_CACHE_REFERENCES
#define VIR_PERF_PARAM_CACHE_REFERENCES "cache_references"
#endif

#ifndef VIR_PERF_PARAM_INSTRUCTIONS
#define VIR_PERF_PARAM_INSTRUCTIONS "instructions"
#endif

#ifndef VIR_PERF_PARAM_CPU_CYCLES
#define VIR_PERF_PARAM_CPU_CYCLES "cpu_cycles"
#endif


/* 2.4.0 */

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_READ_BYTES_SEC_MAX_LENGTH
#define VIR_DOMAIN_BLOCK_IOTUNE_READ_BYTES_SEC_MAX_LENGTH "read_bytes_sec_max_length"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_READ_IOPS_SEC_MAX_LENGTH
#define VIR_DOMAIN_BLOCK_IOTUNE_READ_IOPS_SEC_MAX_LENGTH "read_iops_sec_max_length"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_BYTES_SEC_MAX_LENGTH
#define VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_BYTES_SEC_MAX_LENGTH "total_bytes_sec_max_length"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_IOPS_SEC_MAX_LENGTH
#define VIR_DOMAIN_BLOCK_IOTUNE_TOTAL_IOPS_SEC_MAX_LENGTH "total_iops_sec_max_length"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_WRITE_BYTES_SEC_MAX_LENGTH
#define VIR_DOMAIN_BLOCK_IOTUNE_WRITE_BYTES_SEC_MAX_LENGTH "write_bytes_sec_max_length"
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_WRITE_IOPS_SEC_MAX_LENGTH
#define VIR_DOMAIN_BLOCK_IOTUNE_WRITE_IOPS_SEC_MAX_LENGTH "write_iopcs_sec_max_length"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_BYTES_SEC_MAX_LENGTH
#define VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_BYTES_SEC_MAX_LENGTH "blkdeviotune.total_bytes_sec_max_length"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_READ_BYTES_SEC_MAX_LENGTH
#define VIR_DOMAIN_TUNABLE_BLKDEV_READ_BYTES_SEC_MAX_LENGTH "blkdeviotune.read_bytes_sec_max_length"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_BYTES_SEC_MAX_LENGTH
#define VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_BYTES_SEC_MAX_LENGTH "blkdeviotune.write_bytes_sec_max_length"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_IOPS_SEC_MAX_LENGTH
#define VIR_DOMAIN_TUNABLE_BLKDEV_TOTAL_IOPS_SEC_MAX_LENGTH "blkdeviotune.total_iops_sec_max_length"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_READ_IOPS_SEC_MAX_LENGTH
#define VIR_DOMAIN_TUNABLE_BLKDEV_READ_IOPS_SEC_MAX_LENGTH "blkdeviotune.read_iops_sec_max_length"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_IOPS_SEC_MAX_LENGTH
#define VIR_DOMAIN_TUNABLE_BLKDEV_WRITE_IOPS_SEC_MAX_LENGTH "blkdeviotune.write_iops_sec_max_length"
#endif

#ifndef VIR_DOMAIN_VCPU_HOTPLUGGABLE
#define VIR_DOMAIN_VCPU_HOTPLUGGABLE (1 << 4)
#endif

/* 3.0.0 */

#ifndef VIR_PERF_PARAM_BRANCH_INSTRUCTIONS
#define VIR_PERF_PARAM_BRANCH_INSTRUCTIONS "branch_instructions"
#endif

#ifndef VIR_PERF_PARAM_BRANCH_MISSES
#define VIR_PERF_PARAM_BRANCH_MISSES "branch_misses"
#endif

#ifndef VIR_PERF_PARAM_BUS_CYCLES
#define VIR_PERF_PARAM_BUS_CYCLES "bus_cycles"
#endif

#ifndef VIR_PERF_PARAM_STALLED_CYCLES_FRONTEND
#define VIR_PERF_PARAM_STALLED_CYCLES_FRONTEND "stalled_cycles_frontend"
#endif

#ifndef VIR_PERF_PARAM_STALLED_CYCLES_BACKEND
#define VIR_PERF_PARAM_STALLED_CYCLES_BACKEND "stalled_cycles_backend"
#endif

#ifndef VIR_PERF_PARAM_REF_CPU_CYCLES
#define VIR_PERF_PARAM_REF_CPU_CYCLES "ref_cpu_cycles"
#endif

#ifndef VIR_PERF_PARAM_CPU_CLOCK
#define VIR_PERF_PARAM_CPU_CLOCK "cpu_clock"
#endif

#ifndef VIR_PERF_PARAM_TASK_CLOCK
#define VIR_PERF_PARAM_TASK_CLOCK "task_clock"
#endif

#ifndef VIR_PERF_PARAM_PAGE_FAULTS
#define VIR_PERF_PARAM_PAGE_FAULTS "page_faults"
#endif

#ifndef VIR_PERF_PARAM_CONTEXT_SWITCHES
#define VIR_PERF_PARAM_CONTEXT_SWITCHES "context_switches"
#endif

#ifndef VIR_PERF_PARAM_CPU_MIGRATIONS
#define VIR_PERF_PARAM_CPU_MIGRATIONS "cpu_migrations"
#endif

#ifndef VIR_PERF_PARAM_PAGE_FAULTS_MIN
#define VIR_PERF_PARAM_PAGE_FAULTS_MIN "page_faults_min"
#endif

#ifndef VIR_PERF_PARAM_PAGE_FAULTS_MAJ
#define VIR_PERF_PARAM_PAGE_FAULTS_MAJ "page_faults_maj"
#endif

#ifndef VIR_PERF_PARAM_ALIGNMENT_FAULTS
#define VIR_PERF_PARAM_ALIGNMENT_FAULTS "alignment_faults"
#endif

#ifndef VIR_PERF_PARAM_EMULATION_FAULTS
#define VIR_PERF_PARAM_EMULATION_FAULTS "emulation_faults"
#endif

#ifndef VIR_DOMAIN_EVENT_ID_METADATA_CHANGE
#define VIR_DOMAIN_EVENT_ID_METADATA_CHANGE 23
#endif

#ifndef VIR_DOMAIN_BLOCK_IOTUNE_GROUP_NAME
#define VIR_DOMAIN_BLOCK_IOTUNE_GROUP_NAME "group_name"
#endif

#ifndef VIR_DOMAIN_TUNABLE_BLKDEV_GROUP_NAME
#define VIR_DOMAIN_TUNABLE_BLKDEV_GROUP_NAME "blkdeviotune.group_name"
#endif

/* 3.1.0 */

int virDomainSetVcpuCompat(virDomainPtr domain,
			   const char *cpumap,
			   int state,
			   unsigned int flags);

/* 3.2.0 */

#ifndef VIR_MIGRATE_TLS
#define VIR_MIGRATE_TLS 1 << 16
#endif

#ifndef VIR_DOMAIN_EVENT_ID_BLOCK_THRESHOLD
#define VIR_DOMAIN_EVENT_ID_BLOCK_THRESHOLD 24
#endif

int virDomainSetBlockThresholdCompat(virDomainPtr domain,
                                     const char *dev,
                                     unsigned long long threshold,
                                     unsigned int flags);

/* 3.3.0 */

#ifndef VIR_DOMAIN_JOB_OPERATION
#define VIR_DOMAIN_JOB_OPERATION "operation"
#endif

#ifndef VIR_DOMAIN_JOB_OPERATION_UNKNOWN
#define VIR_DOMAIN_JOB_OPERATION_UNKNOWN 0
#endif

#ifndef VIR_DOMAIN_JOB_OPERATION_START
#define VIR_DOMAIN_JOB_OPERATION_START 1
#endif

#ifndef VIR_DOMAIN_JOB_OPERATION_SAVE
#define VIR_DOMAIN_JOB_OPERATION_SAVE 2
#endif

#ifndef VIR_DOMAIN_JOB_OPERATION_RESTORE
#define VIR_DOMAIN_JOB_OPERATION_RESTORE 3
#endif

#ifndef VIR_DOMAIN_JOB_OPERATION_MIGRATION_IN
#define VIR_DOMAIN_JOB_OPERATION_MIGRATION_IN 4
#endif

#ifndef VIR_DOMAIN_JOB_OPERATION_MIGRATION_OUT
#define VIR_DOMAIN_JOB_OPERATION_MIGRATION_OUT 5
#endif

#ifndef VIR_DOMAIN_JOB_OPERATION_SNAPSHOT
#define VIR_DOMAIN_JOB_OPERATION_SNAPSHOT 6
#endif

#ifndef VIR_DOMAIN_JOB_OPERATION_SNAPSHOT_REVERT
#define VIR_DOMAIN_JOB_OPERATION_SNAPSHOT_REVERT 7
#endif

#ifndef VIR_DOMAIN_JOB_OPERATION_DUMP
#define VIR_DOMAIN_JOB_OPERATION_DUMP 8
#endif


/* 3.4.0 */

#ifndef VIR_DOMAIN_EVENT_SHUTDOWN_GUEST
#define VIR_DOMAIN_EVENT_SHUTDOWN_GUEST 1
#endif

#ifndef VIR_DOMAIN_EVENT_SHUTDOWN_HOST
#define VIR_DOMAIN_EVENT_SHUTDOWN_HOST 2
#endif


#endif /* LIBVIRT_GO_DOMAIN_COMPAT_H__ */
