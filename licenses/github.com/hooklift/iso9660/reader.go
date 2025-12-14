package iso9660

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/c4milo/gotoolkit"
)

var (
	// ErrInvalidImage is returned when an attempt to unpack the image Primary Volume Descriptor failed or
	// when the end of the image was reached without finding a primary volume descriptor
	ErrInvalidImage = func(err error) error { return fmt.Errorf("invalid-iso9660-image: %s", err) }
	// ErrCorruptedImage is returned when a seek operation, on the image, failed.
	ErrCorruptedImage = func(err error) error { return fmt.Errorf("corrupted-image: %s", err) }
)

// Reader defines the state of the ISO9660 image reader. It needs to be instantiated
// from its constructor.
type Reader struct {
	// File descriptor to the opened ISO image
	image io.ReadSeeker
	// Copy of unencoded Primary Volume Descriptor
	pvd PrimaryVolume
	// Queue used to walk through file system iteratively
	queue gotoolkit.Queue
	// Current sector
	sector uint32
	// Current bytes read from current sector.
	read uint32
}

// NewReader creates a new ISO 9660 image reader.
func NewReader(rs io.ReadSeeker) (*Reader, error) {
	// Starts reading from image data area
	sector := dataAreaSector
	// Iterates over volume descriptors until it finds the primary volume descriptor
	// or an error condition.
	for {
		offset, err := rs.Seek(int64(sector*sectorSize), os.SEEK_SET)
		if err != nil {
			return nil, ErrCorruptedImage(err)
		}

		var volDesc VolumeDescriptor
		if err := binary.Read(rs, binary.BigEndian, &volDesc); err != nil {
			return nil, ErrCorruptedImage(err)
		}

		if volDesc.Type == primaryVol {
			// backs up to the beginning of the sector again in order to unpack
			// the entire primary volume descriptor more easily.
			if _, err := rs.Seek(offset, os.SEEK_SET); err != nil {
				return nil, ErrCorruptedImage(err)
			}

			reader := new(Reader)
			reader.image = rs
			reader.queue = new(gotoolkit.SliceQueue)

			if err := reader.unpackPVD(); err != nil {
				return nil, ErrCorruptedImage(err)
			}

			return reader, nil
		}

		if volDesc.Type == volSetTerminator {
			return nil, ErrInvalidImage(errors.New("Volume Set Terminator reached. A Primary Volume Descriptor was not found."))
		}
		sector++
	}
}

// Skip skips the given number of directory records.
func (r *Reader) Skip(n int) error {
	var drecord File
	var len byte
	var err error
	for i := 0; i < n; i++ {
		if len, err = r.unpackDRecord(&drecord); err != nil {
			return err
		}
		r.read += uint32(len)
	}
	return nil
}

// Next moves onto the next directory record present in the image.
// It does not use the Path Table since the goal is to read everything
// from the ISO image.
func (r *Reader) Next() (os.FileInfo, error) {
	if r.queue.IsEmpty() {
		return nil, io.EOF
	}

	// We only dequeue the directory when it does not contain more children
	// or when it is empty and there is no children to iterate over.
	item, err := r.queue.Peek()
	if err != nil {
		panic(err)
	}

	f := item.(File)
	if r.sector == 0 {
		r.sector = f.ExtentLocationBE
		_, err := r.image.Seek(int64(r.sector*sectorSize), os.SEEK_SET)
		if err != nil {
			return nil, ErrCorruptedImage(err)
		}

		// Skips . and .. directories
		if err = r.Skip(2); err != nil {
			return nil, ErrCorruptedImage(err)
		}
	}

	var drecord File
	var len byte
	if (r.read % sectorSize) == 0 {
		r.sector++
		_, err := r.image.Seek(int64(r.sector*sectorSize), os.SEEK_SET)
		if err != nil {
			return nil, ErrCorruptedImage(err)
		}
	}

	if len, err = r.unpackDRecord(&drecord); err != nil && err != io.EOF {
		return nil, ErrCorruptedImage(err)
	}
	r.read += uint32(len)

	if err == io.EOF {
		// directory record is empty, sector space wasted, move onto next sector.
		rsize := (sectorSize - (r.read % sectorSize))
		buf := make([]byte, rsize)
		if err := binary.Read(r.image, binary.BigEndian, buf); err != nil {
			return nil, ErrCorruptedImage(err)
		}
		r.read += rsize
	}

	// If there is no more entries in the current directory, dequeue it.
	if r.read >= f.ExtentLengthBE {
		r.read = 0
		r.sector = 0
		r.queue.Dequeue()
	}

	// End of directory listing, drecord is empty so we don't bother
	// to return it and keep iterating to look for the next actual
	// directory or file.
	if drecord.fileID == "" {
		return r.Next()
	}

	parent := f.Name()
	if parent == "\x00" {
		parent = "/"
	}
	drecord.fileID = filepath.Join(parent, drecord.fileID)

	if drecord.IsDir() {
		r.queue.Enqueue(drecord)
	} else {
		drecord.image = r.image
	}

	return &drecord, nil
}

// unpackDRecord unpacks directory record bits into Go's struct
func (r *Reader) unpackDRecord(f *File) (byte, error) {
	// Gets the directory record length
	var len byte
	if err := binary.Read(r.image, binary.BigEndian, &len); err != nil {
		return len, ErrCorruptedImage(err)
	}

	if len == 0 {
		return len + 1, io.EOF
	}

	// Reads directory record into Go struct
	var drecord DirectoryRecord
	if err := binary.Read(r.image, binary.BigEndian, &drecord); err != nil {
		return len, ErrCorruptedImage(err)
	}

	f.DirectoryRecord = drecord
	// Gets the name
	name := make([]byte, drecord.FileIDLength)
	if err := binary.Read(r.image, binary.BigEndian, name); err != nil {
		return len, ErrCorruptedImage(err)
	}
	f.fileID = string(name)

	// Padding field as per section 9.1.12 in ECMA-119
	if (drecord.FileIDLength % 2) == 0 {
		var zero byte
		if err := binary.Read(r.image, binary.BigEndian, &zero); err != nil {
			return len, ErrCorruptedImage(err)
		}
	}

	// System use field as per section 9.1.13 in ECMA-119
	// Directory record has 34 bytes in addition to the name's
	// variable length and the padding field mentioned in section 9.1.12
	totalLen := 34 + drecord.FileIDLength - (drecord.FileIDLength % 2)
	sysUseLen := int64(len - totalLen)
	if sysUseLen > 0 {
		sysData := make([]byte, sysUseLen)
		if err := binary.Read(r.image, binary.BigEndian, sysData); err != nil {
			return len, ErrCorruptedImage(err)
		}
	}
	return len, nil
}

// unpackPVD unpacks Primary Volume Descriptor in three phases. This is
// because the root directory record is a variable-length record and Go's binary
// package doesn't support unpacking variable-length structs easily.
func (r *Reader) unpackPVD() error {
	// Unpack first half
	var pvd1 PrimaryVolumePart1
	if err := binary.Read(r.image, binary.BigEndian, &pvd1); err != nil {
		return ErrCorruptedImage(err)
	}
	r.pvd.PrimaryVolumePart1 = pvd1

	// Unpack root directory record
	var drecord File
	if _, err := r.unpackDRecord(&drecord); err != nil {
		return ErrCorruptedImage(err)
	}
	r.pvd.DirectoryRecord = drecord.DirectoryRecord
	r.queue.Enqueue(drecord)

	// Unpack second half
	var pvd2 PrimaryVolumePart2
	if err := binary.Read(r.image, binary.BigEndian, &pvd2); err != nil {
		return ErrCorruptedImage(err)
	}
	r.pvd.PrimaryVolumePart2 = pvd2

	return nil
}
