// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clnt

import (
	"k8s.io/minikube/third_party/go9p/p"
)

var Eisdir = &p.Error{"file is a directory", p.EIO}
var Enegoff = &p.Error{"negative i/o offset", p.EIO}

// Seek sets the offset for the next Read or Write to offset,
// interpreted according to whence: 0 means relative to the origin of
// the file, 1 means relative to the current offset, and 2 means
// relative to the end.  Seek returns the new offset and an error, if
// any.
//
// Seeking to a negative offset is an error, and results in Enegoff.
// Seeking to 0 in a directory is only valid if whence is 0. Seek returns
// Eisdir otherwise.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	var off int64

	switch whence {
	case 0:
		// origin
		off = offset
		if f.fid.Qid.Type&p.QTDIR > 0 && off != 0 {
			return 0, Eisdir
		}

	case 1:
		// current
		if f.fid.Qid.Type&p.QTDIR > 0 {
			return 0, Eisdir
		}
		off = offset + int64(f.offset)

	case 2:
		// end
		if f.fid.Qid.Type&p.QTDIR > 0 {
			return 0, Eisdir
		}

		dir, err := f.fid.Clnt.Stat(f.fid)
		if err != nil {
			return 0, &p.Error{"stat error in seek: " + err.Error(), p.EIO}
		}
		off = int64(dir.Length) + offset

	default:
		return 0, &p.Error{"bad whence in seek", p.EIO}
	}

	if off < 0 {
		return 0, Enegoff
	}
	f.offset = uint64(off)

	return off, nil
}
