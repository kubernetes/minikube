// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clnt

import (
	"io"
	"k8s.io/minikube/third_party/go9p/p"
)

// Reads count bytes starting from offset from the file associated with the fid.
// Returns a slice with the data read, if the operation was successful, or an
// Error.
func (clnt *Clnt) Read(fid *Fid, offset uint64, count uint32) ([]byte, error) {
	if count > fid.Iounit {
		count = fid.Iounit
	}

	tc := clnt.NewFcall()
	err := p.PackTread(tc, fid.Fid, offset, count)
	if err != nil {
		return nil, err
	}

	rc, err := clnt.Rpc(tc)
	if err != nil {
		return nil, err
	}

	return rc.Data, nil
}

// Reads up to len(buf) bytes from the File. Returns the number
// of bytes read, or an Error.
func (file *File) Read(buf []byte) (int, error) {
	n, err := file.ReadAt(buf, int64(file.offset))
	if err == nil {
		file.offset += uint64(n)
	}

	return n, err
}

// Reads up to len(buf) bytes from the file starting from offset.
// Returns the number of bytes read, or an Error.
func (file *File) ReadAt(buf []byte, offset int64) (int, error) {
	b, err := file.fid.Clnt.Read(file.fid, uint64(offset), uint32(len(buf)))
	if err != nil {
		return 0, err
	}

	if len(b) == 0 {
		return 0, io.EOF
	}

	copy(buf, b)
	return len(b), nil
}

// Reads exactly len(buf) bytes from the File starting from offset.
// Returns the number of bytes read (could be less than len(buf) if
// end-of-file is reached), or an Error.
func (file *File) Readn(buf []byte, offset uint64) (int, error) {
	ret := 0
	for len(buf) > 0 {
		n, err := file.ReadAt(buf, int64(offset))
		if err != nil {
			return 0, err
		}

		if n == 0 {
			break
		}

		buf = buf[n:]
		offset += uint64(n)
		ret += n
	}

	return ret, nil
}

// Reads the content of the directory associated with the File.
// Returns an array of maximum num entries (if num is 0, returns
// all entries from the directory). If the operation fails, returns
// an Error.
func (file *File) Readdir(num int) ([]*p.Dir, error) {
	buf := make([]byte, file.fid.Clnt.Msize-p.IOHDRSZ)
	dirs := make([]*p.Dir, 32)
	pos := 0
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if n == 0 {
			break
		}

		for b := buf[0:n]; len(b) > 0; {
			d, perr := p.UnpackDir(b, file.fid.Clnt.Dotu)
			if perr != nil {
				return nil, perr
			}

			b = b[d.Size+2:]
			if pos >= len(dirs) {
				s := make([]*p.Dir, len(dirs)+32)
				copy(s, dirs)
				dirs = s
			}

			dirs[pos] = d
			pos++
			if num != 0 && pos >= num {
				break
			}
		}
	}

	return dirs[0:pos], nil
}
