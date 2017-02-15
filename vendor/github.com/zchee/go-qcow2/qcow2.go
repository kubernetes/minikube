// Copyright 2016 The go-qcow2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// QCow2 image format specifications is under the QEMU license.

package qcow2

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

const IO_BUF_SIZE = (2 * 1024 * 1024)

// New return the new Qcow.
func New(config *Opts) *QCow2 {
	return &QCow2{}
}

// Opts options of the create qcow2 image format.
type Opts struct {
	// Filename filename of create image.
	Filename string
	// Fmt format of create image.
	Fmt DriverFmt
	// BaseFliename base filename of create image.
	BaseFilename string
	// BaseFmt base format of create image.
	BaseFmt string

	// BLOCK_OPT
	// Size size of create image virtual size.
	Size int64

	//  Encryption option is if this option is set to "on", the image is encrypted with 128-bit AES-CBC.
	Encryption bool

	//  BackingFile file name of a base image (see create subcommand).
	BackingFile string

	//  BackingFormat image format of the base image.
	BackingFormat string

	//  ClusterSize option is changes the qcow2 cluster size (must be between 512 and 2M).
	//  Smaller cluster sizes can improve the image file size whereas larger cluster sizes generally provide better performance.
	ClusterSize int

	TableSize int

	//  Preallocation mode of pre-allocation metadata (allowed values: "off", "metadata", "falloc", "full").
	//  An image with preallocated metadata is initially larger but can improve performance when the image needs to grow.
	//  "falloc" and "full" preallocations are like the same options of "raw" format, but sets up metadata also.
	Preallocation PreallocMode

	SubFormat string

	//  Compat QCow2 image format compatible. "compat=0.10": uses the traditional image format that can be read by any QEMU since 0.10.
	//  "compat=1.1":  enables image format extensions that only QEMU 1.1 and newer understand (this is the default).
	Compat string

	//  LazyRefcounts option is if this option is set to "on", reference count updates are postponed with the goal of avoiding metadata I/O and improving performance.
	//  This is particularly interesting with cache=writethrough which doesn't batch metadata updates.
	//  The tradeoff is that after a host crash, the reference count tables must be rebuilt,
	//  i.e. on the next open an (automatic) "qemu-img check -r all" is required, which may take some time.
	//  This option can only be enabled if "compat=1.1" is specified.
	LazyRefcounts bool // LazyRefcounts Avoiding metadata I/O and improving performance with the postponed updates reference count.

	AdapterType string

	Redundancy bool

	//  NoCow option is if this option is set to "on", it will turn off COW of the file. It's only valid on btrfs, no effect on other file systems.
	//  Btrfs has low performance when hosting a VM image file, even more when the guest on the VM also using btrfs as file system.
	//  Turning off COW is a way to mitigate this bad performance. Generally there are two ways to turn off COW on btrfs: a)
	//  Disable it by mounting with nodatacow, then all newly created files will be NOCOW. b)
	//  For an empty file, add the NOCOW file attribute. That's what this option does.
	//  Note: this option is only valid to new or empty files.
	//  If there is an existing file which is COW and has data blocks already, it couldn't be changed to NOCOW by setting "nocow=on".
	//  One can issue "lsattr filename" to check if the NOCOW flag is set or not (Capital 'C' is NOCOW flag).
	NoCow bool

	ObjectSize int

	RefcountBits int
}

func (q *QCow2) Len() (int64, error) {
	stat, err := q.blk.bs().File.Stat()
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}

// Create creates the new QCow2 virtual disk image by the qemu style.
func Create(opts *Opts) (*QCow2, error) {
	if opts.Filename == "" {
		err := errors.New("Expecting image file name")
		return nil, err
	}

	// TODO(zchee): implements file size eror handling
	// sval = qemu_strtosz_suffix(argv[optind++], &end,
	// QEMU_STRTOSZ_DEFSUFFIX_B);
	// if (sval < 0 || *end) {
	// 	if (sval == -ERANGE) {
	// 		error_report("Image size must be less than 8 EiB!");
	// 	} else {
	// 		error_report("Invalid image size specified! You may use k, M, "
	// 		"G, T, P or E suffixes for ");
	// 		error_report("kilobytes, megabytes, gigabytes, terabytes, "
	// 		"petabytes and exabytes.");
	// 	}
	// 	goto fail;
	// }

	img := new(QCow2)
	blk, err := create(opts.Filename, opts)
	if err != nil {
		return nil, err
	}
	img.blk = blk
	return img, nil
}

func create(filename string, opts *Opts) (*BlockBackend, error) {

	// ------------------------------------------------------------------------
	// static int qcow2_create(const char *filename, QemuOpts *opts,
	//                         Error **errp)

	var (
		flags int
		// default is version3
		version = Version3
	)

	size := roundUp(int(opts.Size), BDRV_SECTOR_SIZE)
	backingFile := opts.BackingFile
	// backingFormat := opts.BackingFormat

	if opts.Encryption {
		flags |= BLOCK_FLAG_ENCRYPT
	}

	clusterSize := int64(opts.ClusterSize)
	if clusterSize == 0 {
		clusterSize = DEFAULT_CLUSTER_SIZE
	}

	// TODO(zchee): error handle
	prealloc := opts.Preallocation

	compat := opts.Compat
	switch compat {
	case "":
		compat = "1.1" // automatically set to latest compatible version
	case "0.10":
		version = Version2
	case "1.1":
		// nothing to do
	default:
		err := errors.Errorf("Invalid compatibility level: '%s'", compat)
		return nil, err
	}

	if opts.LazyRefcounts {
		flags |= BLOCK_FLAG_LAZY_REFCOUNTS
	}

	if backingFile != "" && prealloc != PREALLOC_MODE_OFF {
		err := errors.New("Backing file and preallocation cannot be used at the same time")
		return nil, err
	}

	if version < 3 && (flags&BLOCK_FLAG_LAZY_REFCOUNTS) == 0 {
		err := errors.New("Lazy refcounts only supported with compatibility level 1.1 and above (use compat=1.1 or greater)")
		return nil, err
	}

	refcountBits := opts.RefcountBits
	if refcountBits == 0 {
		refcountBits = 16 // defaults
	}
	if refcountBits > 64 {
		err := errors.New("Refcount width must be a power of two and may not exceed 64 bits")
		return nil, err
	}

	refcountOrder := ctz32(uint32(refcountBits))

	// ------------------------------------------------------------------------
	// static int qcow2_create2(const char *filename, int64_t total_size,
	//                          const char *backing_file,
	//                          const char *backing_format,
	//                          int flags, size_t cluster_size,
	//                          PreallocMode prealloc,
	//                          QemuOpts *opts, int version,
	//                          int refcount_order,
	//                          Error **errp)

	// Calculate cluster_bits
	clusterBits := ctz32(uint32(clusterSize))
	if clusterBits < MIN_CLUSTER_BITS || clusterBits > MAX_CLUSTER_BITS || (1<<uint(clusterBits)) != opts.ClusterSize {
		err := errors.Errorf("Cluster size must be a power of two between %d and %dk", 1<<MIN_CLUSTER_BITS, 1<<(MAX_CLUSTER_BITS-10))
		return nil, err
	}

	// Open the image file and write a minimal qcow2 header.
	//
	// We keep things simple and start with a zero-sized image. We also
	// do without refcount blocks or a L1 table for now. We'll fix the
	// inconsistency later.
	//
	// We do need a refcount table because growing the refcount table means
	// allocating two new refcount blocks - the seconds of which would be at
	// 2 GB for 64k clusters, and we don't want to have a 2 GB initial file
	// size for any qcow2 image.

	if prealloc == PREALLOC_MODE_FULL || prealloc == PREALLOC_MODE_FALLOC {
		// Note: The following calculation does not need to be exact; if it is a
		// bit off, either some bytes will be "leaked" (which is fine) or we
		// will need to increase the file size by some bytes (which is fine,
		// too, as long as the bulk is allocated here). Therefore, using
		// floating point arithmetic is fine.
		var metaSize int64
		alignedTotalZize := alignOffset(size, int(clusterSize))
		rces := int64(1<<uint(refcountOrder)) / 8.

		refblockBits := clusterBits - (refcountOrder - 3)
		refblockSize := 1 << uint(refblockBits)

		metaSize += int64(clusterSize)

		nl2e := alignedTotalZize / clusterSize
		nl2e = alignOffset(nl2e, int(clusterSize/int64(UINT64_SIZE)))
		metaSize += nl2e * UINT64_SIZE

		nl1e := nl2e * UINT64_SIZE / clusterSize
		nl1e = alignOffset(nl1e, int(clusterSize/int64(UINT64_SIZE)))
		metaSize += nl1e * UINT64_SIZE

		// total size of refcount blocks
		//
		// note: every host cluster is reference-counted, including metadata
		// (even refcount blocks are recursively included).
		// Let:
		//   a = total_size (this is the guest disk size)
		//   m = meta size not including refcount blocks and refcount tables
		//   c = cluster size
		//   y1 = number of refcount blocks entries
		//   y2 = meta size including everything
		//   rces = refcount entry size in bytes
		// then,
		//   y1 = (y2 + a)/c
		//   y2 = y1 * rces + y1 * rces * sizeof(u64) / c + m
		// we can get y1:
		//   y1 = (a + m) / (c - rces - rces * sizeof(u64) / c)
		nrefblocke := (alignedTotalZize + metaSize + int64(clusterSize)) / (int64(clusterSize) - rces - rces*UINT64_SIZE) / int64(clusterSize)
		metaSize += divRoundUp(int(nrefblocke), refblockSize) * int64(clusterSize)

		// total size of refcount tables
		nreftablee := nrefblocke / int64(refblockSize)
		nreftablee = alignOffset(nreftablee, int(clusterSize/int64(UINT64_SIZE)))
		metaSize += nreftablee * UINT64_SIZE

		size = alignedTotalZize + metaSize
	}

	blkOption := new(BlockOption)
	diskImage, err := CreateFile(filename, blkOption)
	if err != nil {
		return nil, err
	}
	defer diskImage.Close()

	blk := new(BlockBackend)
	blk.BlockDriverState = &BlockDriverState{
		file: &BdrvChild{
			Name: diskImage.Name(),
		},
	}

	// TODO(zchee): should use func Open(bs BlockDriverState, options *QDict, flag int) error
	// if err := Open(blk.bs(), nil, flags); err != nil {
	if err := blk.Open(diskImage.Name(), "", nil, os.O_RDWR|os.O_CREATE); err != nil {
		return nil, err
	}

	blk.BlockDriverState.Opaque = new(BDRVState)

	blk.allowBeyondEOF = true

	blk.Header = Header{
		Magic:                 BEUint32(MAGIC), // uint32
		Version:               version,         // uint32
		BackingFileOffset:     uint64(0),
		BackingFileSize:       uint32(0),
		ClusterBits:           uint32(clusterBits),
		Size:                  uint64(size),   // TODO(zchee): Sets to when initializing of the header? qemu is after initialization.
		CryptMethod:           CRYPT_NONE,     // uint32
		L1Size:                uint32(128),    // TODO(zchee): hardcoded
		L1TableOffset:         uint64(458752), // TODO(zchee): hardcoded
		RefcountTableOffset:   uint64(clusterSize),
		RefcountTableClusters: uint32(1),
		NbSnapshots:           uint32(0),
		SnapshotsOffset:       uint64(0),
		IncompatibleFeatures:  uint64(0),
		CompatibleFeatures:    uint64(0),
		AutoclearFeatures:     uint64(0),
		RefcountOrder:         uint32(refcountOrder), // NOTE: qemu now supported only refcount_order = 4
		HeaderLength:          uint32(unsafe.Sizeof(Header{})),
	}

	if opts.Encryption {
		blk.Header.CryptMethod = CRYPT_AES
	}

	if opts.LazyRefcounts {
		blk.Header.CompatibleFeatures |= uint64(COMPAT_LAZY_REFCOUNTS)
	}

	// Write a header data to blk.buf
	binary.Write(&blk.buf, binary.BigEndian, blk.Header)

	if blk.Header.Version >= Version3 {
		binary.Write(&blk.buf, binary.BigEndian, uint32(HeaderExtensionFeatureNameTable))

		features := []Feature{
			Feature{
				Type: uint8(FEAT_TYPE_INCOMPATIBLE),
				Bit:  uint8(INCOMPAT_DIRTY_BITNR),
				Name: "dirty bit",
			},
			Feature{
				Type: uint8(FEAT_TYPE_INCOMPATIBLE),
				Bit:  uint8(INCOMPAT_CORRUPT_BITNR),
				Name: "corrupt bit",
			},
			Feature{
				Type: uint8(FEAT_TYPE_COMPATIBLE),
				Bit:  uint8(COMPAT_LAZY_REFCOUNTS_BITNR),
				Name: "lazy refcounts",
			},
		}

		binary.Write(&blk.buf, binary.BigEndian, uint32(unsafe.Sizeof(Feature{}))*uint32(len(features)))

		for _, f := range features {
			binary.Write(&blk.buf, binary.BigEndian, f.Type)
			binary.Write(&blk.buf, binary.BigEndian, f.Bit)
			binary.Write(&blk.buf, binary.BigEndian, []byte(f.Name))
			zeroFill(&blk.buf, int64(46-uint8(len([]byte(f.Name)))))
		}
	}

	// Write a header data to image file
	writeFile(blk.bs(), 0, blk.buf.Bytes(), blk.buf.Len())

	// Write a refcount table with one refcount block
	refcountTable := make([][]byte, 2*clusterSize)
	refcountTable[0] = BEUvarint64(uint64(2 * clusterSize))

	// TODO(zchee): int(2*clusterSize))?
	writeFile(blk.bs(), clusterSize, bytes.Join(refcountTable, []byte{}), int(clusterSize))

	blk.BlockDriverState.Drv = new(BlockDriver)
	blk.BlockDriverState.Drv.bdrvGetlength = getlength
	// bs.Drv.bdrvTruncate = bdrvTruncate

	blk.BlockDriverState.Opaque = &BDRVState{
		ClusterSize:   int(clusterSize),
		ClusterBits:   clusterBits,
		RefcountOrder: refcountOrder,
	}

	if _, err := AllocClusters(blk.bs(), uint64(3*clusterSize)); err != nil {
		if err != syscall.Errno(0) {
			err = errors.Wrap(err, "Huh, first cluster in empty image is already in use?")
			return nil, err
		}

		err = errors.Wrap(err, "Could not allocate clusters for qcow2 header and refcount table")
		return nil, err
	}

	// Create a full header (including things like feature table)
	// ret = qcow2_update_header(blk_bs(blk));
	// if (ret < 0) {
	// 	error_setg_errno(errp, -ret, "Could not update qcow2 header");
	// 	goto out;
	// }

	// TODO(zchee): carried from bdrv_open_common, should move to the Open function
	blk.bs().Opaque.L2Bits = blk.bs().Opaque.ClusterBits - 3
	blk.bs().Opaque.L2Size = 1 << uint(blk.bs().Opaque.L2Bits)
	blk.bs().Opaque.RefcountTableOffset = blk.Header.RefcountTableOffset
	// blk.bs().Opaque.RefcountTableSize = blk.Header.RefcountTableClusters << uint(blk.bs().Opaque.ClusterBits-3)
	blk.bs().TotalSectors = int64(blk.Header.Size / 512)

	// Okay, now that we have a valid image, let's give it the right size
	if err := truncate(blk.bs(), size); err != nil {
		err = errors.Wrap(err, "Could not resize image")
		return nil, err
	}

	// Want a backing file? There you go
	if backingFile != "" {
		// TODO(zchee): implements bdrv_change_backing_file
	}

	// And if we're supposed to preallocate metadata, do that now
	if prealloc != PREALLOC_MODE_OFF {
		// TODO(zchee): implements preallocate()
	}

	return blk, nil
}

// refreshTotalSectors sets the current 'total_sectors' value
func refreshTotalSectors(bs *BlockDriverState, hint int64) error {
	drv := bs.Drv

	// Do not attempt drv->bdrv_getlength() on scsi-generic devices
	if bs.SG {
		return nil
	}

	// query actual device if possible, otherwise just trust the hint
	if drv.bdrvGetlength != nil {
		length, err := drv.bdrvGetlength(bs)
		if err != nil {
			return err
		}
		if length < 0 {
			return nil
		}
		hint = divRoundUp(int(length), BDRV_SECTOR_SIZE)
	}

	bs.TotalSectors = hint
	return nil
}

func truncate(bs *BlockDriverState, offset int64) error {
	s := bs.Opaque

	if offset&511 != 0 {
		err := errors.Wrap(syscall.EINVAL, "The new size must be a multiple of 512")
		return err
	}

	// cannot proceed if image has snapshots
	if s.NbSnapshots != 0 {
		err := errors.Wrap(syscall.ENOTSUP, "Can't resize an image which has snapshots")
		return err
	}

	// shrinking is currently not supported
	if offset < bs.TotalSectors*512 {
		err := errors.Wrap(syscall.ENOTSUP, "qcow2 doesn't support shrinking images yet")
		return err
	}

	newL1Size := sizeToL1(s, offset)
	if err := growL1Table(bs, uint64(newL1Size), true); err != nil {
		return err
	}

	// write updated header.size
	// off := BEUvarint64(uint64(offset))
	// if err := bdrvPwriteSync(bs.File, unsafe.Offsetof(Header.Size), &offset, UINT64_SIZE); err != nil {
	// 	return err
	// }

	s.L1VmStateIndex = int(newL1Size)
	return nil
}

// writeFile seeks `offset` size, encodes the `data` to big endian format and
// write to the image file.
// If length is bigger than the writed byte data size, fills with zeros the
// up to the length of the `length`.
// Not only of seek, actually grow the file size.
func writeFile(bs *BlockDriverState, offset int64, data []byte, length int) error {
	if bs.File == nil {
		err := errors.New("Not found BlockBackend file")
		return err
	}

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)

	bs.File.Seek(offset, 0)
	off, err := bs.File.Write(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "Could not write a data")
	}

	if length > off {
		if err := zeroFill(bs.File, int64(length-off)); err != nil {
			return err
		}
	}

	return nil
}

func roundUp(n, d int) int64 {
	return int64((n + d - 1) & -d)
}

func divRoundUp(n, d int) int64 {
	return int64((n + d - 1) / d)
}

// zeroFill writes n zero bytes into w.
func zeroFill(w io.Writer, n int64) error {
	const blocksize = 32 << 10
	zeros := make([]byte, blocksize)
	var k int
	var err error
	for n > 0 {
		if n > blocksize {
			k, err = w.Write(zeros)
			if err != nil {
				return err
			}

		} else {
			k, err = w.Write(zeros[:n])
			if err != nil {
				return err
			}

		}
		if err != nil {
			return err
		}
		n -= int64(k)
	}
	return nil
}

// CreateFile creates the new file based by block driver backend.
func CreateFile(filename string, opts *BlockOption) (*os.File, error) {
	image, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return image, nil
}

// Open open the QCow2 block-backend image file.
func (blk *BlockBackend) Open(filename, reference string, options *BlockOption, flag int) error {
	file, err := os.OpenFile(filename, flag, os.FileMode(0))
	if err != nil {
		return err
	}

	blk.BlockDriverState.File = file

	return nil
}

// Open open the QCow2 block-backend image file.
// callgraph:
//  qemu-img.c:img_create -> bdrv_img_create -> bdrv_open -> bdrv_open_inherit -> bdrv_open_common -> drv->bdrv_open -> .bdrv_open = qcow2_open
func Open(bs *BlockDriverState, options *QDict, flag int) error {
	s := bs.Opaque
	var header Header

	err := bdrvPread(bs.file, 0, &header, unsafe.Sizeof(header))
	if err != nil {
		err = errors.Wrap(err, "Could not read qcow2 header")
		return err
	}

	if !bytes.Equal(BEUvarint32(header.Magic), MAGIC) {
		err := errors.Wrap(syscall.EINVAL, "Image is not in qcow2 format")
		return err
	}
	if header.Version < Version2 || header.Version > Version3 {
		err := errors.Wrapf(syscall.ENOTSUP, "Unsupported qcow2 version %d", header.Version)
		return err
	}

	s.Version = header.Version

	// Initialise cluster size
	if header.ClusterBits < MIN_CLUSTER_BITS || header.ClusterBits > MAX_CLUSTER_BITS {
		err := errors.Wrapf(syscall.EINVAL, "Unsupported cluster size: 2^%d", header.ClusterBits)
		return err
	}

	s.ClusterBits = int(header.ClusterBits)
	s.ClusterSize = 1 << uint(s.ClusterBits)
	s.ClusterSectors = 1 << uint(s.ClusterBits-9)

	// Initialise version 3 header fields
	if header.Version == Version2 {
		header.IncompatibleFeatures = 0
		header.CompatibleFeatures = 0
		header.AutoclearFeatures = 0
		header.RefcountOrder = 4
		header.HeaderLength = 72
	} else {
		if header.HeaderLength < 104 {
			err := errors.Wrap(syscall.EINVAL, "qcow2 header too short")
			return err
		}
	}

	if header.HeaderLength > uint32(s.ClusterSize) {
		err := errors.Wrap(syscall.EINVAL, "qcow2 header exceeds cluster size")
		return err
	}

	hdrSizeof := uint32(unsafe.Sizeof(header))
	if header.HeaderLength > hdrSizeof {
		s.UnknownheaderFieldsSize = int(header.HeaderLength - hdrSizeof)
		s.UnknownHeaderFields = make([]byte, s.UnknownheaderFieldsSize)
		err := bdrvPread(bs.file, int64(hdrSizeof), &s.UnknownHeaderFields, uintptr(s.UnknownheaderFieldsSize))
		if err != nil {
			err = errors.Wrap(err, "Could not read unknown qcow2 header fields")
			return err
		}
	}

	if header.BackingFileOffset > uint64(s.ClusterSize) {
		err := errors.Wrap(syscall.EINVAL, "Invalid backing file offset")
		return err
	}

	// var extEnd uint64
	// if header.BackingFileOffset != 0 {
	// 	extEnd = header.BackingFileOffset
	// } else {
	// 	extEnd = 1 << header.ClusterBits
	// }

	// Handle feature bits
	s.IncompatibleFeatures = header.IncompatibleFeatures
	s.CompatibleFeatures = header.CompatibleFeatures
	s.AutoclearFeatures = header.AutoclearFeatures

	if int(s.IncompatibleFeatures) & ^INCOMPAT_MASK != 0 {
		// TODO(zchee): implements read extensions
		// featureTable := nil
		// qcow2_read_extensions(bs, header.header_length, ext_end, &feature_table, NULL);
		// report_unsupported_feature(errp, feature_table, s->incompatible_features & ~QCOW2_INCOMPAT_MASK);
		// ret = -ENOTSUP;
		// g_free(feature_table);
		// goto fail;
	}

	if s.IncompatibleFeatures&INCOMPAT_CORRUPT != 0 {
		// TODO(zchee): implements
		// Corrupt images may not be written to unless they are being repaired
		// if ((flags & BDRV_O_RDWR) && !(flags & BDRV_O_CHECK)) {
		// 	error_setg(errp, "qcow2: Image is corrupt; cannot be opened read/write");
		// 	ret = -EACCES;
		// 	goto fail;
		// }
	}

	// Check support for various header values
	if header.RefcountOrder > 6 {
		err := errors.Wrap(syscall.EINVAL, "Reference count entry width too large; may not exceed 64 bits")
		return err
	}
	s.RefcountOrder = int(header.RefcountOrder)
	s.RefcountBits = 1 << uint(s.RefcountOrder)
	s.RefcountMax = uint64(1) << uint64(s.RefcountBits-1)
	s.RefcountMax += s.RefcountMax - 1

	if header.CryptMethod > CRYPT_AES {
		err := errors.Wrapf(syscall.EINVAL, "Unsupported encryption method: %d", header.CryptMethod)
		return err
	}
	// TODO(zchee): implements
	// if (!qcrypto_cipher_supports(QCRYPTO_CIPHER_ALG_AES_128)) {
	// 	error_setg(errp, "AES cipher not available");
	// 	ret = -EINVAL;
	// 	goto fail;
	// }
	s.CryptMethodHeader = uint32(header.CryptMethod)
	if s.CryptMethodHeader != 0 {
		// TODO(zchee): implements
		// s->crypt_method_header == QCOW_CRYPT_AES) {
		// 	error_setg(errp, "Use of AES-CBC encrypted qcow2 images is no longer supported in system emulators")
		// 	error_append_hint(errp, "You can use 'qemu-img convert' to convert your image to an alternative supported format, such as unencrypted qcow2, or raw with the LUKS format instead.\n")
		// 	ret = -ENOSYS;
		// 	goto fail;
	}

	s.L2Bits = s.ClusterBits - 3
	s.L2Size = 1 << uint(s.L2Bits)
	// 2^(s->refcount_order - 3) is the refcount width in bytes
	s.RefcountBlockBits = s.ClusterBits - (s.RefcountOrder - 3)
	s.RefcountBlockSize = 1 << uint(s.RefcountBlockBits)
	bs.TotalSectors = int64(header.Size / 512)
	s.Csize_shift = (62 - (s.ClusterBits - 8))
	s.Csize_mask = (1 - (s.ClusterBits - 8)) - 1
	s.ClusterOffsetMask = (1 << uint(s.Csize_shift)) - 1

	s.RefcountTableOffset = header.RefcountTableOffset
	s.RefcountTableSize = header.RefcountTableClusters << uint(s.ClusterBits-3)

	if uint64(header.RefcountTableClusters) > maxRefcountClusters(s) {
		err := errors.Wrap(syscall.EINVAL, "Reference count table too large")
		return err
	}

	// ret = validate_table_offset(bs, header.l1_table_offset, header.l1_size, sizeof(uint64_t));

	return nil
}

// getInfo gets the BlockDriverInfo informations.
//  static int qcow2_get_info(BlockDriverState *bs, BlockDriverInfo *bdi)
func getInfo(bs *BlockDriverState) *BlockDriverInfo {
	bdi := new(BlockDriverInfo)
	s := bs.Opaque

	bdi.unallocatedBlocksAreZero = true
	bdi.canWriteZeroesWithUnmap = (s.Version >= 3)
	bdi.clusterSize = s.ClusterSize
	bdi.vmStateOffset = vmStateOffset(s)

	return bdi
}

// selectPart
//  static void convert_select_part(ImgConvertState *s, int64_t sector_num)
func (q *QCow2) selectPart(sectorNum int64) {
	for (sectorNum - q.srcCurOffset) >= int64(BEUvarint64(uint64(q.srcSectors))[q.srcCur]) {
		q.srcCurOffset += int64(BEUvarint64(uint64(q.srcSectors))[q.srcCur])
		q.srcCur++
	}
}

func (q *QCow2) iterationSectors(sectorNum int64) (int, error) {
	// q.selectPart(sectorNum)

	n := MIN(int(q.totalSectors-sectorNum), BDRV_SECTOR_BITS)

	if q.sectorNextStatus <= sectorNum {
		// TODO(zchee): hardcoded BDRV_BLOCK_DATA
		s := BDRV_BLOCK_DATA
		// BlockDriverState *file;
		// ret = bdrv_get_block_status(blk_bs(s->src[s->src_cur]),
		// 							sector_num - s->src_cur_offset,
		// 							n, &n, &file);

		switch {
		case s == BDRV_BLOCK_ZERO:
			q.status = BLK_ZERO
		case s == BDRV_BLOCK_DATA:
			q.status = BLK_DATA
		case !q.targetHasBacking:
			// TODO(zchee): omit
		default:
			q.status = BLK_BACKING_FILE
		}

		q.sectorNextStatus = sectorNum + int64(n)
	}

	n = MIN(n, int(q.sectorNextStatus-sectorNum))
	if q.status == BLK_DATA {
		n = MIN(n, q.bufSectors)
	}

	// TODO(zchee): support zlib compressed write
	// We need to write complete clusters for compressed images, so if an
	// unallocated area is shorter than that, we must consider the whole
	// cluster allocated
	// if (s->compressed) {
	// 	 if (n < s->cluster_sectors) {
	// 		 n = MIN(s->cluster_sectors, s->total_sectors - sector_num);
	// 		 s->status = BLK_DATA;
	// 	 } else {
	// 		 n = QEMU_ALIGN_DOWN(n, s->cluster_sectors);
	// 	 }
	//  }

	return n, nil
}

func (q *QCow2) readData(sectorNum, n int, buf *[]byte) error {
	return nil
}

func (q *QCow2) writeData(sectorNum, n int, buf *[]byte) error {
	return nil
}

func (q *QCow2) Write(data []byte) error {
	bufsectors := IO_BUF_SIZE / BDRV_SECTOR_SIZE
	totalSectors := q.blk.bs().TotalSectors

	// increase bufsectors from the default 4096 (2M) if opt_transfer
	// or discard_alignment of the out_bs is greater. Limit to 32768 (16MB)
	// as maximum.
	bufsectors = MIN(32768,
		MAX(bufsectors,
			MAX(int(q.blk.bs().BL.OptTransfer>>BDRV_SECTOR_BITS), int(q.blk.bs().BL.PdiscardAlignment>>BDRV_SECTOR_BITS))))

	bdi := getInfo(q.blk.bs())

	q.src = q.blk
	// q.srcSectors = bsSectors
	// q.srcNum = bs_n
	q.totalSectors = totalSectors
	q.target = q.blk
	// TODO(zchee): hardcoded false
	q.compressed = false
	// q.targetHasBacking = out_baseimg
	q.minSparse = 8 // Need at least 4k of zeros for sparse detection
	q.clusterSectors = bdi.clusterSize / BDRV_SECTOR_SIZE
	q.bufSectors = bufsectors

	// ------------------------------------------------------------------------
	// static int convert_do_copy(ImgConvertState *s)

	// buf := posixMemalign(4096, uint64(bufsectors*BDRV_SECTOR_SIZE))

	// Calculate allocated sectors for progress
	q.allocatedSectors = 0
	// var sectorNum int64
	// for int64(sectorNum) < q.totalSectors {
	// 	n, err := q.iterationSectors(sectorNum)
	// 	if err != nil && n < 0 {
	// 		return err
	// 	}
	// 	if q.status == BLK_DATA || q.minSparse != 0 && q.status == BLK_ZERO {
	// 		q.allocatedSectors += int64(n)
	// 	}
	// 	sectorNum += int64(n)
	// 	log.Printf("sectorNum: %+v\n", sectorNum)
	// }

	// Do the write
	q.srcCur = 0
	q.srcCurOffset = 0
	q.sectorNextStatus = 0

	// sectorNum = 0
	// allocated_done = 0

	// for int64(sectorNum) < q.totalSectors {
	// 	n, err := q.iterationSectors(sectorNum)
	// 	if err != nil && n < 0 {
	// 		return err
	// 	}
	// 	if q.status == BLK_DATA || q.minSparse != 0 && q.status == BLK_ZERO {
	// 		allocated_done += int64(n)
	// 	}
	//
	// 	if q.status == BLK_DATA {
	// 		if err := q.readData(int(sectorNum), n, &buf); err != nil {
	// 			err = errors.Wrapf(err, "error while reading sector %d", sectorNum)
	// 			return err
	// 		}
	// 	} else if q.minSparse != 0 && q.status == BLK_ZERO {
	// 		n = MIN(n, q.bufSectors)
	// 		mem.Set(buf, 0)
	// 	}
	//
	// 	if err := q.writeData(int(sectorNum), n, &buf); err != nil {
	// 		err = errors.Wrapf(err, "error while writing sector %d", sectorNum)
	// 		return err
	// 	}
	//
	// 	sectorNum += int64(n)
	// }

	// TODO(zchee): support zlib compressed write
	// if (s->compressed) {
	// 	/* signal EOF to align */
	// 	ret = blk_write_compressed(s->target, 0, NULL, 0);
	// 	if (ret < 0) {
	// 		goto fail;
	// 	}
	// }

	// TODO(zchee): tired... hardcoded
	writeFile(q.blk.bs(), 131080, []byte{0, 1, 0, 1, 0, 1, 0, 1}, 8)
	writeFile(q.blk.bs(), 196608, []byte{128, 0, 0, 0, 0, 4}, 6)
	writeFile(q.blk.bs(), 262144, []byte{128, 0, 0, 0, 0, 5}, 6)
	writeFile(q.blk.bs(), 262264, []byte{128, 0, 0, 0, 0, 6}, 6)
	writeFile(q.blk.bs(), 327680, data, len(data))

	return nil
}

// ---------------------------------------------------------------------------
// block/qcow2.h static inline functions

// startOfCluster return the start of cluster.
//  static inline int64_t start_of_cluster(BDRVQcow2State *s, int64_t offset)
func startOfCluster(clusterSize int64, offset int64) int64 {
	return offset &^ (clusterSize - 1)
}

// offsetIntoCluster return the offset into cluster.
//  static inline int64_t offset_into_cluster(BDRVQcow2State *s, int64_t offset)
func offsetIntoCluster(s *BDRVState, offset int64) uint64 {
	return uint64(offset & (int64(s.ClusterSize) - 1))
}

// sizeToClusters return the size to clusters.
//  static inline uint64_t size_to_clusters(BDRVQcow2State *s, uint64_t size)
func sizeToClusters(s *BDRVState, size uint64) uint64 {
	return (size + uint64(s.ClusterSize-1)) >> uint(s.ClusterBits)
}

// sizeToL1 return the L1 size.
//  static inline int64_t size_to_l1(BDRVQcow2State *s, int64_t size)
func sizeToL1(s *BDRVState, size int64) int64 {
	shift := s.ClusterBits + s.L2Bits
	return (size + (1 << uint(shift)) - 1) >> uint(shift)
}

// offsetToL2Index return the L2 index offset.
//  static inline int offset_to_l2_index(BDRVQcow2State *s, int64_t offset)
func offsetToL2Index(s *BDRVState, offset int64) int {
	return int(offset >> uint(s.ClusterBits) & int64(s.L2Size-1))
}

// alignOffset return the aligned offset size.
//  static inline int64_t align_offset(int64_t offset, int n)
func alignOffset(offset int64, n int) int64 {
	offset = offset + int64(n) - 1 & ^(int64(n)-1)
	return offset
}

// vmStateOffset return the offset of vm state.
//  static inline int64_t qcow2_vm_state_offset(BDRVQcow2State *s)
func vmStateOffset(s *BDRVState) int64 {
	return int64(s.L1VmStateIndex << uint(s.ClusterBits+s.L2Bits))
}

// maxRefcountClusters return the maximum size of refcount clusters.
//  static inline uint64_t qcow2_max_refcount_clusters(BDRVQcow2State *s)
func maxRefcountClusters(s *BDRVState) uint64 {
	return MAX_REFTABLE_SIZE >> uint(s.ClusterBits)
}

// getClusterType return the type of cluster.
//  static inline int qcow2_get_cluster_type(uint64_t l2_entry)
func getClusterType(l2Entry uint64) int {
	switch l2Entry {
	case OFLAG_COMPRESSED:
		return int(CLUSTER_COMPRESSED)
	case OFLAG_ZERO:
		return int(CLUSTER_ZERO)
	case L2E_OFFSET_MASK:
		return int(CLUSTER_UNALLOCATED)
	default:
		return int(CLUSTER_NORMAL)
	}
}

// needAccurateRefcounts check whether refcounts are eager or lazy.
//  static inline bool qcow2_need_accurate_refcounts(BDRVQcow2State *s)
func needAccurateRefcounts(s *BDRVState) bool {
	return s.IncompatibleFeatures != INCOMPAT_DIRTY
}

// l2metaCowStart return the start of l2 meta cow.
// TODO(zchee): implements type L2meta struct
//  static inline uint64_t l2meta_cow_start(QCowL2Meta *m)
// func l2metaCowStart(m *L2meta) uint64 {
//     return m.offset + m.cowStart.offset
// }

// l2metaCowEnd return the end of l2meta cow.
// TODO(zchee): implements type L2meta struct
//  static inline uint64_t l2meta_cow_end(QCowL2Meta *m)
// func l2metaCowEnd(m *L2Meta) uint64 {
//     return m.offset + m.cowEnd.offset + m.cowEnd.nbBytes
// }

// refcountDiff return the diff of refcount.
//  static inline uint64_t refcount_diff(uint64_t r1, uint64_t r2)
func refcountDiff(r1, r2 uint64) uint64 {
	if r1 > r2 {
		return r1 - r2
	}
	return r2 - r1
}
