// Copyright 2016 The go-qcow2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// QCow2 image format specifications is under the QEMU license.

package qcow2

import (
	"math"
	"os"
	"syscall"
)

// ---------------------------------------------------------------------------
// go-qcow2

type writeStatus int

const (
	BLK_DATA writeStatus = iota
	BLK_ZERO
	BLK_BACKING_FILE
)

// QCow2 represents a QEMU QCow2 image format.
type QCow2 struct {
	blk *BlockBackend

	// ImgConvertState
	src              *BlockBackend
	srcSectors       int64         // int64_t
	srcCur, srcNum   int           // int
	srcCurOffset     int64         // int64_t
	totalSectors     int64         // int64_t
	allocatedSectors int64         // int64_t
	status           writeStatus   // ImgConvertBlockStatus
	sectorNextStatus int64         // int64_t
	target           *BlockBackend // BlockBackend
	hasZeroInit      bool          // bool
	compressed       bool          // bool
	targetHasBacking bool          // bool
	minSparse        int           // int
	clusterSectors   int           // size_t
	bufSectors       int           // size_t

}

const (
	// UINT16_SIZE results of sizeof(uint16_t) in C.
	UINT16_SIZE = 2
	// UINT32_SIZE results of sizeof(uint32_t) in C.
	UINT32_SIZE = 4
	// UINT64_SIZE results of sizeof(uint64_t) in C.
	UINT64_SIZE = 8
)

// Version represents a version number of qcow2 image format.
// The valid values are 2 or 3.
type Version uint32

const (
	// Version2 qcow2 image format version2.
	Version2 Version = 2
	// Version3 qcow2 image format version3.
	Version3 Version = 3
)

const (
	// Version2HeaderSize is the image header at the beginning of the file.
	Version2HeaderSize = 72
	// Version3HeaderSize is directly following the v2 header, up to 104.
	Version3HeaderSize = 104
)

// FeatureNameTable represents a optional header extension that contains the name for features used by the image.
type FeatureNameTable struct {
	// Type type of feature [0:1]
	Type int
	// BitNumber bit number within the selected feature bitmap [1:2]
	BitNumber int
	// FeatureName feature name. padded with zeros [2:48]
	FeatureName int
}

// BitmapExtension represents a optional header extension.
type BitmapExtension struct {
	// NbBitmaps the number of bitmaps contained in the image. Must be greater than ro equal to 1. [1:4]
	NbBitmaps int
	// Reserved reserved, must be zero. [4:8]
	Reserved int
	// BitmapDirectorySize size of the bitmap directory in bytes. It is the cumulative size of all (nb_bitmaps) bitmap headers. [8:16]
	BitmapDirectorySize int
	// BitmapDirectoryOffset offste into the image file at which the bitmap directory starts. [16:24]
	BitmapDirectoryOffset int
}

const BDRV_SECTOR_BITS = 9

var (
	BDRV_SECTOR_SIZE = 1 << BDRV_SECTOR_BITS   // (1ULL << BDRV_SECTOR_BITS)
	BDRV_SECTOR_MASK = ^(BDRV_SECTOR_SIZE - 1) // ~(BDRV_SECTOR_SIZE - 1)
)

var BDRV_REQUEST_MAX_SECTORS = MIN(SIZE_MAX>>BDRV_SECTOR_BITS, INT_MAX>>BDRV_SECTOR_BITS)

// ---------------------------------------------------------------------------
// block/qcow2.c

// Extension represents a optional header extension.
type Extension struct {
	Magic HeaderExtensionType // [:4] Header extension type
	Len   uint32              // [4:8] Length of the header extension data
}

// HeaderExtensionType represents a indicators the the entries in the optional header area
type HeaderExtensionType uint32

const (
	// HeaderExtensionEndOfArea End of the header extension area.
	HeaderExtensionEndOfArea HeaderExtensionType = 0x00000000

	// HeaderExtensionBackingFileFormat Backing file format name.
	HeaderExtensionBackingFileFormat HeaderExtensionType = 0xE2792ACA

	// HeaderExtensionFeatureNameTable Feature name table.
	HeaderExtensionFeatureNameTable HeaderExtensionType = 0x6803f857

	// HeaderExtensionBitmapsExtension Bitmaps extension.
	// TODO(zchee): qemu does not implements?
	HeaderExtensionBitmapsExtension HeaderExtensionType = 0x23852875

	// Safely ignored other unknown header extension
)

// ---------------------------------------------------------------------------
// block/qcow2.h

// MAGIC qemu QCow(2) magic ("QFI\xfb").
//  #define QCOW_MAGIC (('Q' << 24) | ('F' << 16) | ('I' << 8) | 0xfb)
var MAGIC = []byte{0x51, 0x46, 0x49, 0xFB}

// CryptMethod represents a whether encrypted qcow2 image.
// 0 for no enccyption
// 1 for AES encryption
type CryptMethod uint32

const (
	// CRYPT_NONE no encryption.
	CRYPT_NONE CryptMethod = iota
	// CRYPT_AES AES encryption.
	CRYPT_AES

	MAX_CRYPT_CLUSTERS = 32
	MAX_SNAPSHOTS      = 65536
)

// String implementations of fmt.Stringer.
func (cm CryptMethod) String() string {
	if cm == 1 {
		return "AES"
	}
	return "none"
}

// MAX_REFTABLE_SIZE 8 MB refcount table is enough for 2 PB images at 64k cluster size
// (128 GB for 512 byte clusters, 2 EB for 2 MB clusters)
const MAX_REFTABLE_SIZE = 0x800000

// MAX_L1_SIZE 32 MB L1 table is enough for 2 PB images at 64k cluster size
// (128 GB for 512 byte clusters, 2 EB for 2 MB clusters)
const MAX_L1_SIZE = 0x2000000

/* Allow for an average of 1k per snapshot table entry, should be plenty of
 * space for snapshot names and IDs */
const MAX_SNAPSHOTS_SIZE = 1024 * MAX_SNAPSHOTS

const (
	// indicate that the refcount of the referenced cluster is exactly one.
	OFLAG_COPIED = 1 << 63
	// indicate that the cluster is compressed (they never have the copied flag)
	OFLAG_COMPRESSED = 1 << 62
	// The cluster reads as all zeros
	OFLAG_ZERO = 1 << 0
)

const (
	// MIN_CLUSTER_BITS minimum of cluster bits size.
	MIN_CLUSTER_BITS = 9
	// MAX_CLUSTER_BITS maximum of cluster bits size.
	MAX_CLUSTER_BITS = 21
)

// MIN_L2_CACHE_SIZE must be at least 2 to cover COW.
const MIN_L2_CACHE_SIZE = 2 // clusters

// MIN_REFCOUNT_CACHE_SIZE must be at least 4 to cover all cases of refcount table growth.
const MIN_REFCOUNT_CACHE_SIZE = 4 // clusters

/* Whichever is more */
const DEFAULT_L2_CACHE_CLUSTERS = 8        // clusters
const DEFAULT_L2_CACHE_BYTE_SIZE = 1048576 // bytes

// DEFAULT_L2_REFCOUNT_SIZE_RATIO the refblock cache needs only a fourth of the L2 cache size to cover as many
// clusters.
const DEFAULT_L2_REFCOUNT_SIZE_RATIO = 4

const DEFAULT_CLUSTER_SIZE = 65536

// Header represents a header of qcow2 image format.
type Header struct {
	Magic                 uint32      //     [0:3] magic: QCOW magic string ("QFI\xfb")
	Version               Version     //     [4:7] Version number
	BackingFileOffset     uint64      //    [8:15] Offset into the image file at which the backing file name is stored.
	BackingFileSize       uint32      //   [16:19] Length of the backing file name in bytes.
	ClusterBits           uint32      //   [20:23] Number of bits that are used for addressing an offset whithin a cluster.
	Size                  uint64      //   [24:31] Virtual disk size in bytes
	CryptMethod           CryptMethod //   [32:35] Crypt method
	L1Size                uint32      //   [36:39] Number of entries in the active L1 table
	L1TableOffset         uint64      //   [40:47] Offset into the image file at which the active L1 table starts
	RefcountTableOffset   uint64      //   [48:55] Offset into the image file at which the refcount table starts
	RefcountTableClusters uint32      //   [56:59] Number of clusters that the refcount table occupies
	NbSnapshots           uint32      //   [60:63] Number of snapshots contained in the image
	SnapshotsOffset       uint64      //   [64:71] Offset into the image file at which the snapshot table starts
	IncompatibleFeatures  uint64      //   [72:79] for version >= 3: Bitmask of incomptible feature
	CompatibleFeatures    uint64      //   [80:87] for version >= 3: Bitmask of compatible feature
	AutoclearFeatures     uint64      //   [88:95] for version >= 3: Bitmask of auto-clear feature
	RefcountOrder         uint32      //   [96:99] for version >= 3: Describes the width of a reference count block entry
	HeaderLength          uint32      // [100:103] for version >= 3: Length of the header structure in bytes
}

// SnapshotHeader represents a header of snapshot.
type SnapshotHeader struct {
}

// SnapshotExtraData represents a extra data of snapshot.
type SnapshotExtraData struct {
}

// Snapshot represents a snapshot.
type Snapshot struct {
}

// Cache represents a cache.
type Cache struct {
}

// UnknownHeaderExtension represents a unknown of header extension.
type UnknownHeaderExtension struct {
	Magic uint32
	Len   uint32
	// Next QLIST_ENTRY(Qcow2UnknownHeaderExtension)
	Data []int8
}

// FeatureType represents a type of feature.
type FeatureType uint8

const (
	// FEAT_TYPE_INCOMPATIBLE incompatible feature.
	FEAT_TYPE_INCOMPATIBLE FeatureType = iota
	// FEAT_TYPE_COMPATIBLE compatible feature.
	FEAT_TYPE_COMPATIBLE
	// FEAT_TYPE_AUTOCLEAR Autoclear feature.
	FEAT_TYPE_AUTOCLEAR
)

const (
	// INCOMPAT_DIRTY_BITNR represents a incompatible dirty bit number.
	INCOMPAT_DIRTY_BITNR = 0

	// INCOMPAT_CORRUPT_BITNR represents a incompatible corrupt bit number.
	INCOMPAT_CORRUPT_BITNR = 1

	// INCOMPAT_DIRTY incompatible corrupt bit number.
	INCOMPAT_DIRTY = 1 << INCOMPAT_DIRTY_BITNR
	// INCOMPAT_CORRUPT incompatible corrupt bit number.
	INCOMPAT_CORRUPT = 1 << INCOMPAT_CORRUPT_BITNR

	// INCOMPAT_MASK mask of incompatible feature.
	INCOMPAT_MASK = INCOMPAT_DIRTY | INCOMPAT_CORRUPT
)

const (
	// COMPAT_LAZY_REFCOUNTS_BITNR represents a compatible dirty bit number.
	COMPAT_LAZY_REFCOUNTS_BITNR = 0
	// COMPAT_LAZY_REFCOUNTS refcounts of lazy compatible.
	COMPAT_LAZY_REFCOUNTS = 1 << COMPAT_LAZY_REFCOUNTS_BITNR

	// COMPAT_FEAT_MASK mask of compatible feature.
	COMPAT_FEAT_MASK = COMPAT_LAZY_REFCOUNTS
)

// DiscardType represents a type of discard.
type DiscardType int

const (
	// DISCARD_NEVER discard never.
	DISCARD_NEVER DiscardType = iota
	// DISCARD_ALWAYS discard always.
	DISCARD_ALWAYS
	// DISCARD_REQUEST discard request.
	DISCARD_REQUEST
	// DISCARD_SNAPSHOT discard snapshot.
	DISCARD_SNAPSHOT
	// DISCARD_OTHER discard other.
	DISCARD_OTHER
	// DISCARD_MAX discard max.
	DISCARD_MAX
)

type Feature struct {
	Type uint8  // uint8_t
	Bit  uint8  // uint8_t
	Name string // char    name[46];
	byt  []byte
}

type DiscardRegion struct {
	Bs     *BlockDriverState
	Offset uint64 // uint64_t
	byt    uint64 // uint64_t
	// next QTAILQ_ENTRY(Qcow2DiscardRegion)
}

// GetRefcountFunc typedef uint64_t Qcow2GetRefcountFunc(const void *refcount_array, uint64_t index);
func GetRefcountFunc(refcountArray map[uint64]uintptr, index uint64) uint64 {
	// ro0 := (refcountArray[index/8] >> (index % 8)) & 0x1
	// ro1 := (refcountArray)[index/4] >> (2 * (index % 4))
	// ro2 := (refcountArray)[index/2] >> (4 * (index % 2))
	// ro3 := (refcountArray)[index]
	// ro4 := BEUvarint16(uint16(refcountArray[index]))
	// ro5 := BEUvarint32(uint32(refcountArray[index]))
	// ro6 := BEUvarint64(uint64(refcountArray[index]))

	// TODO(zchee): WIP
	return 0
}

// SetRefcountFunc typedef void Qcow2SetRefcountFunc(void *refcount_array, uint64_t index, uint64_t value);
func SetRefcountFunc(refcountArray map[uint64]uintptr, index uint64) {
	// TODO(zchee): WIP
	return
}

type BDRVState struct {
	ClusterBits       int    // int
	ClusterSize       int    // int
	ClusterSectors    int    // int
	L2Bits            int    // int
	L2Size            int    // int
	L1Size            int    // int
	L1VmStateIndex    int    // int
	RefcountBlockBits int    // int
	RefcountBlockSize int    // int
	Csize_shift       int    // int
	Csize_mask        int    // int
	ClusterOffsetMask uint64 // uint64_t
	L1TableOffset     uint64 // uint64_t
	L1Table           uint64 // uint64_t

	L2TableCache       *Cache // *Qcow2Cache
	RefcountBlockCache *Cache // *Qcow2Cache
	// cache_clean_timer    *QEMUTimer
	CacheCleanInterval uintptr // unsigned

	ClusterCache       uint8  // uint8_t
	ClusterData        uint8  // uint8_t
	ClusterCacheOffset uint64 // uint64_t
	// cluster_allocs QLIST_HEAD(QCowClusterAlloc, QCowL2Meta)

	// RefcountTable       map[uint64]int64 // uint64_t
	RefcountTable       [8][]byte // uint64_t
	RefcountTableOffset uint64    // uint64_t
	RefcountTableSize   uint32    // uint32_t
	FreeClusterIndex    uint64    // uint64_t
	FreeByteOffset      uint64    // uint64_t

	// lock CoMutex // CoMutex

	// cipher              *QCryptoCipher // current cipher, nil if no key yet
	CryptMethodHeader uint32  // uint32_t
	SnapshotsOffset   uint64  // uint64_t
	SnapshotsSize     int     // int
	NbSnapshots       uintptr // unsigend int
	// snapshots           *QCowSnapshot

	Flags            int     // int
	Version          Version // int
	UseLazyRefcounts bool    // bool
	RefcountOrder    int     // int
	RefcountBits     int     // int
	RefcountMax      uint64  // uint64_t

	GetRefcount func(refcountArray interface{}, index uint64) uint64        // *Qcow2GetRefcountFunc
	SetRefcount func(refcountArray interface{}, index uint64, value uint64) // *Qcow2SetRefcountFunc

	DiscardPassthrough bool // bool discard_passthrough[QCOW2_DISCARD_MAX]

	OverlapCheck       int  // int: bitmask of Qcow2MetadataOverlap values
	SignaledCorruption bool // bool

	IncompatibleFeatures uint64 // uint64_t
	CompatibleFeatures   uint64 // uint64_t
	AutoclearFeatures    uint64 // uint64_t

	UnknownheaderFieldsSize int    // size_t
	UnknownHeaderFields     []byte // void*
	// unknown_header_ext QLIST_HEAD(, Qcow2UnknownHeaderExtension)
	// discards QTAILQ_HEAD (, Qcow2DiscardRegion)
	CacheDiscards bool // bool

	// Backing file path and format as stored in the image (this is not the
	// effective path/format, which may be the result of a runtime option
	// override)
	ImageBackingFile   string // char *
	ImageBackingFormat []byte // char *
}

type CLUSTER uint64

const (
	CLUSTER_UNALLOCATED CLUSTER = iota
	CLUSTER_NORMAL
	CLUSTER_COMPRESSED
	CLUSTER_ZERO
)

const (
	L1E_OFFSET_MASK                 = uint64(72057594037927424)   // 0x00fffffffffffe00ULL
	L2E_OFFSET_MASK                 = uint64(72057594037927424)   // 0x00fffffffffffe00ULL
	L2E_COMPRESSED_OFFSET_SIZE_MASK = uint64(4611686018427387903) // 0x3fffffffffffffffULL
	REFT_OFFSET_MASK                = uint64(1844674407370955110) // 0xfffffffffffffe00ULL
)

// ---------------------------------------------------------------------------
// include/block/block.h

type BlockDriverInfo struct {
	// in bytes, 0 if irrelevant
	clusterSize int // int
	// offset at which the VM state can be saved (0 if not possible)
	vmStateOffset int64 // int64_t
	isDirty       bool  // bool

	// True if unallocated blocks read back as zeroes. This is equivalent
	// to the LBPRZ flag in the SCSI logical block provisioning page.
	unallocatedBlocksAreZero bool // bool

	// True if the driver can optimize writing zeroes by unmapping
	// sectors. This is equivalent to the BLKDISCARDZEROES ioctl in Linux
	// with the difference that in qemu a discard is allowed to silently
	// fail. Therefore we have to use bdrv_pwrite_zeroes with the
	// BDRV_REQ_MAY_UNMAP flag for an optimized zero write with unmapping.
	// After this call the driver has to guarantee that the contents read
	// back as zero. It is additionally required that the block device is
	// opened with BDRV_O_UNMAP flag for this to work.
	canWriteZeroesWithUnmap bool // bool

	// True if this block driver only supports compressed writes
	needsCompressedWrites bool // bool
}

const (
	BDRV_BLOCK_DATA         = 0x01
	BDRV_BLOCK_ZERO         = 0x02
	BDRV_BLOCK_OFFSET_VALID = 0x04
	BDRV_BLOCK_RAW          = 0x08
	BDRV_BLOCK_ALLOCATED    = 0x10
)

var BDRV_BLOCK_OFFSET_MASK = BDRV_SECTOR_MASK

// ---------------------------------------------------------------------------
// include/block/block_int.h

// BlockDriver represents a block driver.
type BlockDriver struct {
	formatName   DriverFmt
	instanceSize int

	/* set to true if the BlockDriver is a block filter */
	isFilter bool
	/* for snapshots block filter like Quorum can implement the
	 * following recursive callback.
	 * It's purpose is to recurse on the filter children while calling
	 * bdrv_recurse_is_first_non_filter on them.
	 * For a sample implementation look in the future Quorum block filter.
	 */
	// bool (*bdrv_recurse_is_first_non_filter)(BlockDriverState *bs, BlockDriverState *candidate)

	// (*bdrv_probe)(const uint8_t *buf, int buf_size, const char *filename) int // TODO
	// (*bdrv_probe_device)(const char *filename) int // TODO

	/* Any driver implementing this callback is expected to be able to handle
	 * nil file names in its .bdrv_open() implementation */
	// (*bdrv_parse_filename)(const char *filename, QDict *options, Error **errp) func() // TODO
	/* Drivers not implementing bdrv_parse_filename nor bdrv_open should have
	 * this field set to true, except ones that are defined only by their
	 * child's bs.
	 * An example of the last type will be the quorum block driver.
	 */
	bdrvNeedsFilename bool

	/* Set if a driver can support backing files */
	supportsBacking bool

	/* For handling image reopen for split or non-split files */
	// int (*bdrv_reopen_prepare)(BDRVReopenState *reopen_state,
	//                            BlockReopenQueue *queue, Error **errp);
	// void (*bdrv_reopen_commit)(BDRVReopenState *reopen_state);
	// void (*bdrv_reopen_abort)(BDRVReopenState *reopen_state);
	// void (*bdrv_join_options)(QDict *options, QDict *old_options);

	// int (*bdrv_open)(BlockDriverState *bs, QDict *options, int flags,
	//                  Error **errp);
	// int (*bdrv_file_open)(BlockDriverState *bs, QDict *options, int flags,
	//                       Error **errp);
	// void (*bdrv_close)(BlockDriverState *bs);
	// int (*bdrv_create)(const char *filename, QemuOpts *opts, Error **errp);
	// int (*bdrv_set_key)(BlockDriverState *bs, const char *key);
	// int (*bdrv_make_empty)(BlockDriverState *bs);

	// void (*bdrv_refresh_filename)(BlockDriverState *bs, QDict *options);

	// aio
	// BlockAIOCB *(*bdrv_aio_readv)(BlockDriverState *bs, int64_t sector_num, QEMUIOVector *qiov, int nb_sectors, BlockCompletionFunc *cb, void *opaque);
	// BlockAIOCB *(*bdrv_aio_writev)(BlockDriverState *bs, int64_t sector_num, QEMUIOVector *qiov, int nb_sectors, BlockCompletionFunc *cb, void *opaque);
	// BlockAIOCB *(*bdrv_aio_flush)(BlockDriverState *bs, BlockCompletionFunc *cb, void *opaque);
	// BlockAIOCB *(*bdrv_aio_pdiscard)(BlockDriverState *bs, int64_t offset, int count, BlockCompletionFunc *cb, void *opaque);

	// int coroutine_fn (*bdrv_co_readv)(BlockDriverState *bs, int64_t sector_num, int nb_sectors, QEMUIOVector *qiov);
	// int coroutine_fn (*bdrv_co_preadv)(BlockDriverState *bs, uint64_t offset, uint64_t bytes, QEMUIOVector *qiov, int flags);
	// int coroutine_fn (*bdrv_co_writev)(BlockDriverState *bs, int64_t sector_num, int nb_sectors, QEMUIOVector *qiov);
	// int coroutine_fn (*bdrv_co_writev_flags)(BlockDriverState *bs, int64_t sector_num, int nb_sectors, QEMUIOVector *qiov, int flags);
	// int coroutine_fn (*bdrv_co_pwritev)(BlockDriverState *bs, uint64_t offset, uint64_t bytes, QEMUIOVector *qiov, int flags);

	// Efficiently zero a region of the disk image.  Typically an image format
	// would use a compact metadata representation to implement this.  This
	// function pointer may be nil or return -ENOSUP and .bdrv_co_writev()
	// will be called instead.
	//
	// int coroutine_fn (*bdrv_co_pwrite_zeroes)(BlockDriverState *bs, int64_t offset, int count, BdrvRequestFlags flags);
	// int coroutine_fn (*bdrv_co_pdiscard)(BlockDriverState *bs, int64_t offset, int count);
	// int64_t coroutine_fn (*bdrv_co_get_block_status)(BlockDriverState *bs, int64_t sector_num, int nb_sectors, int *pnum, BlockDriverState **file);

	// Invalidate any cached meta-data.
	// void (*bdrv_invalidate_cache)(BlockDriverState *bs, Error **errp);
	// int (*bdrv_inactivate)(BlockDriverState *bs);

	// Flushes all data for all layers by calling bdrv_co_flush for underlying
	// layers, if needed. This function is needed for deterministic
	// synchronization of the flush finishing callback.
	// int coroutine_fn (*bdrv_co_flush)(BlockDriverState *bs);

	// Flushes all data that was already written to the OS all the way down to
	// the disk (for example raw-posix calls fsync()).
	// int coroutine_fn (*bdrv_co_flush_to_disk)(BlockDriverState *bs);

	// Flushes all internal caches to the OS. The data may still sit in a
	// writeback cache of the host OS, but it will survive a crash of the qemu
	// process.
	// int coroutine_fn (*bdrv_co_flush_to_os)(BlockDriverState *bs);

	protocol_name string
	// bdrvTruncate  func(bs *BlockDriverState, offset int64) error // NOTE: implemented use interface

	bdrvGetlength     func(bs *BlockDriverState) (int64, error)
	hasVariableLength bool
	// int64_t (*bdrv_get_allocated_file_size)(BlockDriverState *bs);

	// int (*bdrv_write_compressed)(BlockDriverState *bs, int64_t sector_num, const uint8_t *buf, int nb_sectors);

	// int (*bdrv_snapshot_create)(BlockDriverState *bs, QEMUSnapshotInfo *sn_info);
	// int (*bdrv_snapshot_goto)(BlockDriverState *bs, const char *snapshot_id);
	// int (*bdrv_snapshot_delete)(BlockDriverState *bs, const char *snapshot_id, const char *name, Error **errp);
	// int (*bdrv_snapshot_list)(BlockDriverState *bs, QEMUSnapshotInfo **psn_info);
	// int (*bdrv_snapshot_load_tmp)(BlockDriverState *bs, const char *snapshot_id, const char *name, Error **errp);
	// int (*bdrv_get_info)(BlockDriverState *bs, BlockDriverInfo *bdi);
	// ImageInfoSpecific *(*bdrv_get_specific_info)(BlockDriverState *bs);

	// int coroutine_fn (*bdrv_save_vmstate)(BlockDriverState *bs, QEMUIOVector *qiov, int64_t pos);
	// int coroutine_fn (*bdrv_load_vmstate)(BlockDriverState *bs, QEMUIOVector *qiov, int64_t pos);

	// int (*bdrv_change_backing_file)(BlockDriverState *bs, const char *backing_file, const char *backing_fmt);

	// removable device specific
	// bool (*bdrv_is_inserted)(BlockDriverState *bs);
	// int (*bdrv_media_changed)(BlockDriverState *bs);
	// void (*bdrv_eject)(BlockDriverState *bs, bool eject_flag);
	// void (*bdrv_lock_medium)(BlockDriverState *bs, bool locked);

	// to control generic scsi devices
	// BlockAIOCB *(*bdrv_aio_ioctl)(BlockDriverState *bs, unsigned long int req, void *buf, BlockCompletionFunc *cb, void *opaque);

	// List of options for creating images, terminated by name == nil
	createOpts *OptsList

	//
	// Returns 0 for completed check, -errno for internal errors.
	// The check results are stored in result.
	//
	// int (*bdrv_check)(BlockDriverState* bs, BdrvCheckResult *result, BdrvCheckMode fix);

	// int (*bdrv_amend_options)(BlockDriverState *bs, QemuOpts *opts, BlockDriverAmendStatusCB *status_cb, void *cb_opaque);

	// void (*bdrv_debug_event)(BlockDriverState *bs, BlkdebugEvent event);

	// TODO Better pass a option string/QDict/QemuOpts to add any rule?
	// int (*bdrv_debug_breakpoint)(BlockDriverState *bs, const char *event, const char *tag);
	// int (*bdrv_debug_remove_breakpoint)(BlockDriverState *bs, const char *tag);
	// int (*bdrv_debug_resume)(BlockDriverState *bs, const char *tag);
	// bool (*bdrv_debug_is_suspended)(BlockDriverState *bs, const char *tag);

	// void (*bdrv_refresh_limits)(BlockDriverState *bs, Error **errp);

	//
	// Returns 1 if newly created images are guaranteed to contain only
	// zeros, 0 otherwise.
	//
	// int (*bdrv_has_zero_init)(BlockDriverState *bs);

	// Remove fd handlers, timers, and other event loop callbacks so the event
	// loop is no longer in use.  Called with no in-flight requests and in
	// depth-first traversal order with parents before child nodes.
	//
	// void (*bdrv_detach_aio_context)(BlockDriverState *bs);

	// Add fd handlers, timers, and other event loop callbacks so I/O requests
	// can be processed again.  Called with no in-flight requests and in
	// depth-first traversal order with child nodes before parent nodes.
	//
	// void (*bdrv_attach_aio_context)(BlockDriverState *bs, AioContext *new_context);

	// io queue for linux-aio
	// void (*bdrv_io_plug)(BlockDriverState *bs);
	// void (*bdrv_io_unplug)(BlockDriverState *bs);

	//
	// Try to get @bs's logical and physical block size.
	// On success, store them in @bsz and return zero.
	// On failure, return negative errno.
	//
	// int (*bdrv_probe_blocksizes)(BlockDriverState *bs, BlockSizes *bsz);

	//
	// Try to get @bs's geometry (cyls, heads, sectors)
	// On success, store them in @geo and return 0.
	// On failure return -errno.
	// Only drivers that want to override guest geometry implement this
	// callback; see hd_geometry_guess().
	//
	// int (*bdrv_probe_geometry)(BlockDriverState *bs, HDGeometry *geo);

	//
	// Drain and stop any internal sources of requests in the driver, and
	// remain so until next I/O callback (e.g. bdrv_co_writev) is called.
	//
	// void (*bdrv_drain)(BlockDriverState *bs);

	// void (*bdrv_add_child)(BlockDriverState *parent, BlockDriverState *child, Error **errp);
	// void (*bdrv_del_child)(BlockDriverState *parent, BdrvChild *child, Error **errp);

	// QLIST_ENTRY(BlockDriver) list;
}

type Truncater interface {
	Truncate(bs *BlockDriverState, offset int64) error
}

const pagesize = 4096

func bdrvOptMemAlign(bs *BlockDriverState) uint32 {
	if bs == nil || bs.Drv == nil {
		// page size or 4k (hdd sector size) should be on the safe side
		return uint32(MAX(4096, pagesize))
	}

	return bs.BL.OptMemAlignment
}

// nbSectors return number of sectors.
func nbSectors(bs *BlockDriverState) (int64, error) {
	drv := bs.Drv

	if drv == nil {
		return 0, ENOMEDIUM
	}

	if drv.hasVariableLength {
		err := refreshTotalSectors(bs, bs.TotalSectors)
		if err != nil {
			return 0, err
		}
	}

	return bs.TotalSectors, nil
}

// getlength return length in bytes.
// The length is always a multiple of BDRV_SECTOR_SIZE.
func getlength(bs *BlockDriverState) (int64, error) {
	length, err := nbSectors(bs)
	if err != nil {
		return 0, err
	}

	if length > int64(INT64_MAX/BDRV_SECTOR_SIZE) {
		return 0, syscall.EFBIG
	}
	return length * int64(BDRV_SECTOR_SIZE), nil
}

const BLOCK_FLAG_ENCRYPT = 1

const BLOCK_FLAG_LAZY_REFCOUNTS = 8

const BLOCK_PROBE_BUF_SIZE = 512

type DriverFmt string

const (
	// DriverRaw raw driver format.
	DriverRaw DriverFmt = "raw"
	// DriverQCow2 qcow2 driver format.
	DriverQCow2 DriverFmt = "qcow2"
)

type BlockLimits struct {
	// RequestAlignment alignment requirement, in bytes, for offset/length of I/O
	// requests. Must be a power of 2 less than INT_MAX; defaults to
	// 1 for drivers with modern byte interfaces, and to 512
	// otherwise.
	RequestAlignment uint32 // uint32_t

	// MaxPdiscard maximum number of bytes that can be discarded at once (since it
	// is signed, it must be < 2G, if set). Must be multiple of
	// pdiscard_alignment, but need not be power of 2. May be 0 if no
	// inherent 32-bit limit
	MaxPdiscard int32 // int32_t

	// PdiscardAlignment optimal alignment for discard requests in bytes. A power of 2
	// is best but not mandatory.  Must be a multiple of
	// bl.request_alignment, and must be less than max_pdiscard if
	// that is set. May be 0 if bl.request_alignment is good enough
	PdiscardAlignment uint32 // uint32_t

	// MaxPwriteZeroes maximum number of bytes that can zeroized at once (since it is
	// signed, it must be < 2G, if set). Must be multiple of
	// pwrite_zeroes_alignment. May be 0 if no inherent 32-bit limit
	MaxPwriteZeroes int32 // int32_t

	// PwriteZeroesAlignment optimal alignment for write zeroes requests in bytes. A power
	// of 2 is best but not mandatory.  Must be a multiple of
	// bl.request_alignment, and must be less than max_pwrite_zeroes
	// if that is set. May be 0 if bl.request_alignment is good
	// enough
	PwriteZeroesAlignment uint32 // uint32_t

	// OptTransfer optimal transfer length in bytes.  A power of 2 is best but not
	// mandatory.  Must be a multiple of bl.request_alignment, or 0 if
	// no preferred size
	OptTransfer uint32 // uint32_t

	// MaxTransfer maximal transfer length in bytes.  Need not be power of 2, but
	// must be multiple of opt_transfer and bl.request_alignment, or 0
	// for no 32-bit limit.  For now, anything larger than INT_MAX is
	// clamped down.
	MaxTransfer uint32 // uint32_t

	// MinMemAlignment memory alignment, in bytes so that no bounce buffer is needed
	MinMemAlignment uint32 // size_t

	// OptMemAlignment memory alignment, in bytes, for bounce buffer
	OptMemAlignment uint32 // size_t

	// MaxIov maximum number of iovec elements
	MaxIov int // int
}

// BlockDriverState represents a state of block driver.
//
// Note: the function bdrv_append() copies and swaps contents of
// BlockDriverStates, so if you add new fields to this struct, please
// inspect bdrv_append() to determine if the new fields need to be
// copied as well.
type BlockDriverState struct {
	TotalSectors int64 // int64_t: if we are reading a disk image, give its size in sectors
	OpenFlags    int   // int:     flags used to open the file, re-used for re-open
	ReadOnly     bool  // bool:    if true, the media is read only
	Encrypted    bool  // bool:    if true, the media is encrypted
	ValidKey     bool  // bool:    if true, a valid encryption key has been set
	SG           bool  // bool:    if true, the device is a /dev/sg* (scsi-generic devices)
	Probed       bool  // bool:    if true, format was probed rather than specified

	CopyOnRead int // int: if nonzero, copy read backing sectors into image. note this is a reference count.

	// flush_queue // CoQueue: Serializing flush queue // TODO
	// active_flush_req // *BdrvTrackedRequest: Flush request in flight // TODO
	WriteGen   uint // unsigned int: Current data generation
	FlushedGen uint // unsigned int: Flushed write generation

	Drv    *BlockDriver // BlockDriver *: NULL means no media
	Opaque *BDRVState   // void *

	// AioContext *AioContext // event loop used for fd handlers, timers, etc // TODO

	// long-running tasks intended to always use the same AioContext as this
	// BDS may register themselves in this list to be notified of changes
	// regarding this BDS's context
	// AioNotifiers QLIST_HEAD(, BdrvAioNotifier) // TODO
	WalkingAioNotifiers bool // bool: to make removal during iteration safe

	Filename      string // char: filename[PATH_MAX]
	BackingFile   string // char: if non zero, the image is a diff of this file image
	BackingFormat string // char: if non-zero and backing_file exists

	// FullOpenOptions *QDict // *QDict * TODO
	ExactFilename string // char: exact_filename[PATH_MAX]

	Backing *BdrvChild
	File    *os.File
	file    *BdrvChild

	// BeforeWriteNotifiers Callback before write request is processed
	// BeforeWriteNotifiers NotifierWithReturnList // TODO

	// SerialisingInFlight number of in-flight serialising requests
	SerialisingInFlight uint // unsigned int

	// Offset after the highest byte written to
	WrHighestOffset uint64 // uint64_t

	// I/O Limits
	BL BlockLimits

	// unsigned int: Flags honored during pwrite (so far: BDRV_REQ_FUA)
	SupportedWriteFlags uint
	// unsigned int: Flags honored during pwrite_zeroes (so far: BDRV_REQ_FUA, *BDRV_REQ_MAY_UNMAP)
	SupportedZeroFlags uint

	// NodeName the following member gives a name to every node on the bs graph.
	NodeName string // char node_name[32]
	// NodeList element of the list of named nodes building the graph
	// NodeList QTAILQ_ENTRY(BlockDriverState) // TODO
	// BsList element of the list of all BlockDriverStates (all_bdrv_states)
	// BsList QTAILQ_ENTRY(BlockDriverState) // TODO
	// MonitorList element of the list of monitor-owned BDS
	// MonitorList QTAILQ_ENTRY(BlockDriverState) // TODO
	// DirtyBitmaps QLIST_HEAD(, BdrvDirtyBitmap) // TODO
	Refcnt int // int

	// TrackedRequests QLIST_HEAD(, BdrvTrackedRequest) // TODO

	// operation blockers
	// OpBlockers [BLOCK_OP_TYPE_MAX]QLIST_HEAD(, BdrvOpBlocker) // operation blockers TODO

	// long-running background operation
	// Job *BlockJob // TODO

	// The node that this node inherited default options from (and a reopen on
	// which can affect this node by changing these defaults). This is always a
	// parent node of this node.
	// InheritsFrom *BlockDriverState // BlockDriverState *: TODO
	// Children QLIST_HEAD(, BdrvChild) // TODO
	// Parents QLIST_HEAD(, BdrvChild) // TODO

	// Options         *QDict                      // TODO
	// ExplicitOptions *QDict                      // TODO
	// DetectZeroes    BlockdevDetectZeroesOptions // TODO

	// The error object in use for blocking operations on backing_hd
	BackingBlocker error

	// threshold limit for writes, in bytes. "High water mark"
	WriteThresholdOffset uint64
	// WriteThresholdNotifier NotifierWithReturn // TODO

	// Counters for nested bdrv_io_plug and bdrv_io_unplugged_begin
	IOPlugged      uintptr // unsigned: TODO
	IOPlugDisabled uintptr // unsigned: TODO

	QuiesceCounter int // int
}

type BdrvChild struct {
	bs   *BlockDriverState
	Name string
	// Role   *BdrvChildRole
	Opaque *BDRVState
	// next QLIST_ENTRY(BdrvChild)
	// next_parent QLIST_ENTRY(BdrvChild)
}

// ---------------------------------------------------------------------------
// qapi-types.h

type QType int

const (
	QTYPE_NONE QType = iota
	QTYPE_QNULL
	QTYPE_QINT
	QTYPE_QSTRING
	QTYPE_QDICT
	QTYPE_QLIST
	QTYPE_QFLOAT
	QTYPE_QBOOL
	QTYPE__MAX
)

// PreallocMode represents a mode of Pre-allocation feature.
type PreallocMode int

const (
	// PREALLOC_MODE_OFF turn off preallocation.
	PREALLOC_MODE_OFF PreallocMode = iota
	// PREALLOC_MODE_METADATA preallocation of metadata only mode.
	PREALLOC_MODE_METADATA
	// PREALLOC_MODE_FALLOC preallocation of falloc only mode.
	PREALLOC_MODE_FALLOC
	// PREALLOC_MODE_FULL full preallocation mode.
	PREALLOC_MODE_FULL
	// PREALLOC_MODE__MAX preallocation maximum preallocation mode.
	PREALLOC_MODE__MAX
)

// ---------------------------------------------------------------------------
// include/qemu/option.h

type OptType int

const (
	OPT_STRING OptType = 0 // no parsing (use string as-is)
	OPT_BOOL               // on/off
	OPT_NUMBER             // simple number
	OPT_SIZE               // size, accepts (K)ilo, (M)ega, (G)iga, (T)era postfix
)

type OptDesc struct {
	Name        string
	Type        OptType
	Help        string
	DefValueStr string
}

type OptsList struct {
	Name           string
	ImpliedOptName string
	MergeLists     bool // Merge multiple uses of option into a single list?
	// head QTAILQ_HEAD(, QemuOpts)
	Desc []OptDesc
}

// ---------------------------------------------------------------------------
// include/qemu/option_int.h

type Opt struct {
	Name string
	Str  string

	desc  *OptDesc
	Value struct {
		Boolean bool
		_uint   uint64
	}

	Opts *Opts
	// QTAILQ_ENTRY(QemuOpt) next;
}

type QemuOpts struct {
	ID   string
	List *OptsList
	// Location loc; // in error-report.h, not needed.
	// QTAILQ_HEAD(QemuOptHead, QemuOpt) head;
	// QTAILQ_ENTRY(QemuOpts) next;
}

// ---------------------------------------------------------------------------
// include/qapi/qmp/qobject.h

type QObject struct {
	typ    QType
	refcnt int64
}

// ---------------------------------------------------------------------------
// include/qapi/qmp/qdict.h

type QDict struct {
	base QObject
	size int64
	// QLIST_HEAD(,QDictEntry) table[QDICT_BUCKET_MAX];
}

// ---------------------------------------------------------------------------
// glib-2.0/glib/gmacros.h

// MIN returns the whichever small of a and b.
func MIN(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MAX returns the whichever larger of a and b.
func MAX(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// stdint.h

// Just wrapped of the math stdlib package const variable.
// Right side comments carried from the C header.

const (
	INT8_MAX  = math.MaxInt8  // 127
	INT16_MAX = math.MaxInt16 // 32767
	INT32_MAX = math.MaxInt32 // 2147483647
	INT64_MAX = math.MaxInt64 // 9223372036854775807LL
	INT_MAX   = math.MaxInt32 // INT_MAX == INT32_MAX on darwin,amd64
)

const (
	INT8_MIN  = math.MinInt8  // -128
	INT16_MIN = math.MinInt16 // -32768
	// Note (from stdint.h):
	// the literal "most negative int" cannot be written in C --
	// the rules in the standard (section 6.4.4.1 in C99) will give it
	// an unsigned type, so INT32_MIN (and the most negative member of
	// any larger signed type) must be written via a constant expression.
	INT32_MIN = math.MinInt32 // (-INT32_MAX-1)
	INT64_MIN = math.MinInt64 // (-INT64_MAX-1)
)

const (
	UINT8_MAX  = math.MaxUint8  // 255
	UINT16_MAX = math.MaxUint16 // 65535
	UINT32_MAX = math.MaxUint32 // 4294967295U
	UINT64_MAX = math.MaxUint64 // 18446744073709551615ULL
)

const (
	SIZE_MAX = math.MaxUint64 // #if __WORDSIZE == 64; UINT64_MAX; #else; #define SIZE_MAX	UINT32_MAX; #endif
)
