// Package iso9660 implements ECMA-119 standard, also known as ISO 9660.
//
// References:
//
// * https://en.wikipedia.org/wiki/ISO_9660
//
// * http://alumnus.caltech.edu/~pje/iso9660.html
//
// * http://users.telenet.be/it3.consultants.bvba/handouts/ISO9960.html
//
// * http://www.ecma-international.org/publications/files/ECMA-ST/Ecma-119.pdf
//
// * http://www.drdobbs.com/database/inside-the-iso-9660-filesystem-format/184408899
//
// * http://www.cdfs.com
//
// * http://wiki.osdev.org/ISO_9660
package iso9660

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"strings"
	"time"
)

// File represents a concrete implementation of os.FileInfo interface for
// accessing ISO 9660 file data
type File struct {
	DirectoryRecord
	fileID string
	// We have the raw image here only to be able to access file extents
	image io.ReadSeeker
}

// Name returns the file's name.
func (f *File) Name() string {
	name := strings.Split(f.fileID, ";")[0]
	return strings.ToLower(name)
}

// Size returns the file size in bytes
func (f *File) Size() int64 {
	return int64(f.ExtentLengthBE)
}

// Mode returns file's mode and permissions bits. Since we don't yet support
// Rock Ridge extensions we cannot extract POSIX permissions and the rest of the
// normal metadata. So, right we return 0740 for directories and 0640 for files.
func (f *File) Mode() os.FileMode {
	if f.IsDir() {
		return os.FileMode(0740)
	}
	return os.FileMode(0640)
}

// ModTime returns file's modification time.
func (f *File) ModTime() time.Time {
	return time.Now()
}

// IsDir tells whether the file is a directory or not.
func (f *File) IsDir() bool {
	if (f.FileFlags & 2) == 2 {
		return true
	}
	return false
}

// Sys returns io.Reader instance pointing to the file's content if it is not a directory, nil otherwise.
func (f *File) Sys() interface{} {
	if f.IsDir() {
		return nil
	}

	// if f.ExtentLengthBE <= 0 {
	// 	return bytes.NewReader([]byte(""))
	// }
	// Saves the current position within the ISO image. This is so we can
	// restore it once the file content is read. By doing this we allow
	// reader.Next() to keep working normally.
	curOffset, err := f.image.Seek(0, os.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	_, err = f.image.Seek(int64(f.ExtentLocationBE*sectorSize), os.SEEK_SET)
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, f.ExtentLengthBE)
	err = binary.Read(f.image, binary.BigEndian, buffer)
	if err != nil {
		panic(err)
	}

	// Restores original position within the ISO image after reading file's content.
	_, err = f.image.Seek(curOffset, os.SEEK_SET)
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(buffer)
}

const (
	bootRecord       = 0
	primaryVol       = 1
	supplementaryVol = 2
	volPartition     = 3
	volSetTerminator = 255
	// System area goes from sectors 0x00 to 0x0F. Volume descriptors can be
	// found starting at sector 0x10
	dataAreaSector = 0x10
	sectorSize     = 2048
)

// VolumeDescriptor identify the volume, the partitions recorded on the volume,
// the volume creator(s), certain attributes of the volume, the location of
// other recorded descriptors and the version of the standard which applies
// to the volume descriptor.
//
// When preparing to mount a CD, your first action will be reading the volume
// descriptors (specifically, you will be looking for the Primary Volume Descriptor).
// Since sectors 0x00-0x0F of the CD are reserved as System Area, the Volume
// Descriptors can be found starting at sector 0x10.
type VolumeDescriptor struct {
	// 0: BootRecord
	// 1: Primary Volume Descriptor
	// 2: Supplementary Volume Descriptor
	// 3: Volume Partition Descriptor
	// 4-254: Reserved
	// 255: Volume Descriptor Set Terminator
	Type byte
	// Always "CD001".
	StandardID [5]byte
	//Volume Descriptor Version (0x01).
	Version byte
}

// BootRecord identifies a system which can recognize and act upon the content
// of the field reserved for boot system use in the Boot Record, and shall
// contain information which is used to achieve a specific state for a system or
// for an application.
type BootRecord struct {
	VolumeDescriptor
	SystemID  [32]byte
	ID        [32]byte
	SystemUse [1977]byte
}

// Terminator indicates the termination of a Volume Descriptor Set.
type Terminator struct {
	VolumeDescriptor
	// All bytes of this field are set to (00).
	Reserved [2041]byte
}

// PrimaryVolumePart1 represents the Primary Volume Descriptor first half, before the
// root directory record. We are only reading big-endian values so placeholders
// are used for little-endian ones.
type PrimaryVolumePart1 struct {
	VolumeDescriptor
	// Unused
	_ byte // 00
	// The name of the system that can act upon sectors 0x00-0x0F for the volume.
	SystemID [32]byte
	// Identification of this volume.
	ID [32]byte
	//Unused2
	_ [8]byte
	// Amount of data available on the CD-ROM. Ignores little-endian order.
	// Takes big-endian encoded value.
	VolumeSpaceSizeLE int32
	VolumeSpaceSizeBE int32
	Unused2           [32]byte
	// The size of the set in this logical volume (number of disks). Ignores
	// little-endian order. Takes big-endian encoded value.
	VolumeSetSizeLE int16
	VolumeSetSizeBE int16
	// The number of this disk in the Volume Set. Ignores little-endian order.
	// Takes big-endian encoded value.
	VolumeSeqNumberLE int16
	VolumeSeqNumberBE int16
	// The size in bytes of a logical block. NB: This means that a logical block
	// on a CD could be something other than 2 KiB!
	LogicalBlkSizeLE int16
	LogicalBlkSizeBE int16
	// The size in bytes of the path table. Ignores little-endian order.
	// Takes big-endian encoded value.
	PathTableSizeLE int32
	PathTableSizeBE int32
	// LBA location of the path table. The path table pointed to contains only
	// little-endian values.
	LocPathTableLE int32
	// LBA location of the optional path table. The path table pointed to contains
	// only little-endian values. Zero means that no optional path table exists.
	LocOptPathTableLE int32
	// LBA location of the path table. The path table pointed to contains
	// only big-endian values.
	LocPathTableBE int32
	// LBA location of the optional path table. The path table pointed to contains
	// only big-endian values. Zero means that no optional path table exists.
	LocOptPathTableBE int32
}

// DirectoryRecord describes the characteristics of a file or directory,
// beginning with a length octet describing the size of the entire entry.
// Entries themselves are of variable length, up to 255 octets in size.
// Attributes for the file described by the directory entry are stored in the
// directory entry itself (unlike UNIX).
// The root directory entry is a variable length object, so that the name can be of variable length.
//
// Important: before each entry there can be "fake entries" to support the Long File Name.
//
// Even if a directory spans multiple sectors, the directory entries are not
// permitted to cross the sector boundary (unlike the path table). Where there
// is not enough space to record an entire directory entry at the end of a sector,
// that sector is zero-padded and the next consecutive sector is used.
// Unfortunately, the date/time format is different from that used in the Primary
// Volume Descriptor.
type DirectoryRecord struct {
	// Extended Attribute Record length, stored at the beginning of
	// the file's extent.
	ExtendedAttrLen byte
	// Location of extent (Logical Block Address) in both-endian format.
	ExtentLocationLE uint32
	ExtentLocationBE uint32
	// Data length (size of extent) in both-endian format.
	ExtentLengthLE uint32
	ExtentLengthBE uint32
	// Date and the time of the day at which the information in the Extent
	// described by the Directory Record was recorded.
	RecordedTime [7]byte
	// If this Directory Record identifies a directory then bit positions 2, 3
	// and 7 shall be set to ZERO. If no Extended Attribute Record is associated
	// with the File Section identified by this Directory Record then bit
	// positions 3 and 4 shall be set to ZERO. -- 9.1.6
	FileFlags byte
	// File unit size for files recorded in interleaved mode, zero otherwise.
	FileUnitSize byte
	// Interleave gap size for files recorded in interleaved mode, zero otherwise.
	InterleaveGapSize byte
	// Volume sequence number - the volume that this extent is recorded on, in
	// 16 bit both-endian format.
	VolumeSeqNumberLE uint16
	VolumeSeqNumberBE uint16
	// Length of file identifier (file name). This terminates with a ';'
	// character followed by the file ID number in ASCII coded decimal ('1').
	FileIDLength byte
	// The interpretation of this field depends as follows on the setting of the
	// Directory bit of the File Flags field. If set to ZERO, it shall mean:
	//
	// − The field shall specify an identification for the file.
	// − The characters in this field shall be d-characters or d1-characters, SEPARATOR 1, SEPARATOR 2.
	// − The field shall be recorded as specified in 7.5. If set to ONE, it shall mean:
	// − The field shall specify an identification for the directory.
	// − The characters in this field shall be d-characters or d1-characters, or only a (00) byte, or only a (01) byte.
	// − The field shall be recorded as specified in 7.6.
	// fileID string
}

// PrimaryVolumePart2 represents the Primary Volume Descriptor half after the
// root directory record.
type PrimaryVolumePart2 struct {
	// Identifier of the volume set of which this volume is a member.
	VolumeSetID [128]byte
	// The volume publisher. For extended publisher information, the first byte
	// should be 0x5F, followed by the filename of a file in the root directory.
	// If not specified, all bytes should be 0x20.
	PublisherID [128]byte
	// The identifier of the person(s) who prepared the data for this volume.
	// For extended preparation information, the first byte should be 0x5F,
	// followed by the filename of a file in the root directory. If not specified,
	// all bytes should be 0x20.
	DataPreparerID [128]byte
	// Identifies how the data are recorded on this volume. For extended information,
	// the first byte should be 0x5F, followed by the filename of a file in the root
	// directory. If not specified, all bytes should be 0x20.
	AppID [128]byte
	// Filename of a file in the root directory that contains copyright
	// information for this volume set. If not specified, all bytes should be 0x20.
	CopyrightFileID [37]byte
	// Filename of a file in the root directory that contains abstract information
	// for this volume set. If not specified, all bytes should be 0x20.
	AbstractFileID [37]byte
	// Filename of a file in the root directory that contains bibliographic
	// information for this volume set. If not specified, all bytes should be 0x20.
	BibliographicFileID [37]byte
	// The date and time of when the volume was created.
	CreationTime [17]byte
	// The date and time of when the volume was modified.
	ModificationTime [17]byte
	// The date and time after which this volume is considered to be obsolete.
	// If not specified, then the volume is never considered to be obsolete.
	ExpirationTime [17]byte
	// The date and time after which the volume may be used. If not specified,
	// the volume may be used immediately.
	EffectiveTime [17]byte
	// The directory records and path table version (always 0x01).
	FileStructVersion byte
	// Reserved. Always 0x00.
	_ byte
	// Contents not defined by ISO 9660.
	AppUse [512]byte
	// Reserved by ISO.
	_ [653]byte
}

// PrimaryVolume descriptor acts much like the superblock of the UNIX filesystem, providing
// details on the ISO-9660 compliant portions of the disk. While we can have
// many kinds of filesystems on a single ISO-9660 CD-ROM, we can have only one
// ISO-9660 file structure (found as the primary volume-descriptor type).
//
// Directory entries are successively stored within this region. Evaluation of
// the ISO 9660 filenames is begun at this location. The root directory is stored
// as an extent, or sequential series of sectors, that contains each of the
// directory entries appearing in the root.
//
// Since ISO 9660 works by segmenting the CD-ROM into logical blocks, the size
// of these blocks is found in the primary volume descriptor as well.
type PrimaryVolume struct {
	PrimaryVolumePart1
	DirectoryRecord DirectoryRecord
	PrimaryVolumePart2
}

// SupplementaryVolume is used by Joliet.
type SupplementaryVolume struct {
	VolumeDescriptor
	Flags               int    `struc:"int8"`
	SystemID            string `struc:"[32]byte"`
	ID                  string `struc:"[32]byte"`
	Unused              byte
	VolumeSpaceSize     int    `struc:"int32"`
	EscapeSequences     string `struc:"[32]byte"`
	VolumeSetSize       int    `struc:"int16"`
	VolumeSeqNumber     int    `struc:"int16"`
	LogicalBlkSize      int    `struc:"int16"`
	PathTableSize       int    `struc:"int32"`
	LocLPathTable       int    `struc:"int32"`
	LocOptLPathTable    int    `struc:"int32"`
	LocMPathTable       int    `struc:"int32"`
	LocOptMPathTable    int    `struc:"int32"`
	RootDirRecord       DirectoryRecord
	VolumeSetID         string `struc:"[128]byte"`
	PublisherID         string `struc:"[128]byte"`
	DataPreparerID      string `struc:"[128]byte"`
	AppID               string `struc:"[128]byte"`
	CopyrightFileID     string `struc:"[37]byte"`
	AbstractFileID      string `struc:"[37]byte"`
	BibliographicFileID string `struc:"[37]byte"`
	CreationTime        Timestamp
	ModificationTime    Timestamp
	ExpirationTime      Timestamp
	EffectiveTime       Timestamp
	FileStructVersion   int `struc:"int8"`
	Reserved            byte
	AppData             [512]byte
	Reserved2           byte
}

// PartitionVolume ...
type PartitionVolume struct {
	VolumeDescriptor
	Unused    byte
	SystemID  string `struc:"[32]byte"`
	ID        string `struc:"[32]byte"`
	Location  int    `struc:"int8"`
	Size      int    `struc:"int8"`
	SystemUse [1960]byte
}

// Timestamp ...
type Timestamp struct {
	Year        int `struc:"[4]byte"`
	Month       int `struc:"[2]byte"`
	DayOfMonth  int `struc:"[2]byte"`
	Hour        int `struc:"[2]byte"`
	Minute      int `struc:"[2]byte"`
	Second      int `struc:"[2]byte"`
	Millisecond int `struc:"[2]byte"`
	GMTOffset   int `struc:"uint8"`
}

// ExtendedAttrRecord are simply a way to extend the attributes of files.
// Since attributes vary according to the user, most everyone has a different
// opinion on what a file attribute should specify.
type ExtendedAttrRecord struct {
	OwnerID          int `struc:"int16"`
	GroupID          int `struc:"int16"`
	Permissions      int `struc:"int16"`
	CreationTime     Timestamp
	ModificationTime Timestamp
	ExpirationTime   Timestamp
	// Specifies the date and the time of the day at which the information in
	// the file may be used. If the date and time are not specified then the
	// information may be used at once.
	EffectiveTime   Timestamp
	Format          int    `struc:"uint8"`
	Attributes      int    `struc:"uint8"`
	Length          int    `struc:"int16"`
	SystemID        string `struc:"[32]byte"`
	SystemUse       [64]byte
	Version         int `struc:"uint8"`
	EscapeSeqLength int `struc:"uint8"`
	Reserved        [64]byte
	AppUseLength    int `struc:"int16,sizeof=AppUse"`
	AppUse          []byte
	EscapeSequences []byte `struc:"sizefrom=AppUseLength"`
}

// PathTable contains a well-ordered sequence of records describing every
// directory extent on the CD. There are some exceptions with this: the Path
// Table can only contain 65536 records, due to the length of the "Parent Directory Number" field.
// If there are more than this number of directories on the disc, some CD
// authoring software will ignore this limit and create a non-compliant
// CD (this applies to some earlier versions of Nero, for example). If your
// file system uses the path table, you should be aware of this possibility.
// Windows uses the Path Table and will fail with such non-compliant
// CD's (additional nodes exist but appear as zero-byte). Linux, which uses
// the directory tables is not affected by this issue. The location of the path
// tables can be found in the Primary Volume Descriptor. There are two table types
// - the L-Path table (relevant to x86) and the M-Path table. The only
// difference between these two tables is that multi-byte values in the L-Table
// are LSB-first and the values in the M-Table are MSB-first.
//
// The path table is in ascending order of directory level and is alphabetically
// sorted within each directory level.
type PathTable struct {
	DirIDLength      int `struc:"uint8,sizeof=DirName"`
	ExtendedAttrsLen int `struc:"uint8"`
	// Number the Logical Block Number of the first Logical Block allocated to
	// the Extent in which the directory is recorded.
	// This is in a different format depending on whether this is the L-Table or
	// M-Table (see explanation above).
	ExtentLocation int `struc:"int32"`
	// Number of record for parent directory (or 1 for the root directory), as a
	// word; the first record is number 1, the second record is number 2, etc.
	// Directory number of parent directory (an index in to the path table).
	// This is the field that limits the table to 65536 records.
	ParentDirNumber int `struc:"int16"`
	// Directory Identifier (name) in d-characters.
	DirName string
}
