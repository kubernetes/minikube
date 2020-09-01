// Copyright 2009 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import (
	"log"
	"os"
	"strconv"
	"syscall"
)

type pipeFid struct {
	path      string
	file      *os.File
	dirs      []os.FileInfo
	dirents   []byte
	diroffset uint64
	st        os.FileInfo
	data      []uint8
	eof       bool
}

type Pipefs struct {
	Srv
	Root string
}

func (fid *pipeFid) stat() *Error {
	var err error

	fid.st, err = os.Lstat(fid.path)
	if err != nil {
		return toError(err)
	}

	return nil
}

// Dir is an instantiation of the p.Dir structure
// that can act as a receiver for local methods.
type pipeDir struct {
	Dir
}

func (*Pipefs) ConnOpened(conn *Conn) {
	if conn.Srv.Debuglevel > 0 {
		log.Println("connected")
	}
}

func (*Pipefs) ConnClosed(conn *Conn) {
	if conn.Srv.Debuglevel > 0 {
		log.Println("disconnected")
	}
}

func (*Pipefs) FidDestroy(sfid *SrvFid) {
	var fid *pipeFid

	if sfid.Aux == nil {
		return
	}

	fid = sfid.Aux.(*pipeFid)
	if fid.file != nil {
		fid.file.Close()
	}
}

func (pipe *Pipefs) Attach(req *SrvReq) {
	if req.Afid != nil {
		req.RespondError(Enoauth)
		return
	}

	tc := req.Tc
	fid := new(pipeFid)
	if len(tc.Aname) == 0 {
		fid.path = pipe.Root
	} else {
		fid.path = tc.Aname
	}

	req.Fid.Aux = fid
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	qid := dir2Qid(fid.st)
	req.RespondRattach(qid)
}

func (*Pipefs) Flush(req *SrvReq) {}

func (*Pipefs) Walk(req *SrvReq) {
	fid := req.Fid.Aux.(*pipeFid)
	tc := req.Tc

	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	if req.Newfid.Aux == nil {
		req.Newfid.Aux = new(pipeFid)
	}

	nfid := req.Newfid.Aux.(*pipeFid)
	wqids := make([]Qid, len(tc.Wname))
	path := fid.path
	i := 0
	for ; i < len(tc.Wname); i++ {
		p := path + "/" + tc.Wname[i]
		st, err := os.Lstat(p)
		if err != nil {
			if i == 0 {
				req.RespondError(Enoent)
				return
			}

			break
		}

		wqids[i] = *dir2Qid(st)
		path = p
	}

	nfid.path = path
	req.RespondRwalk(wqids[0:i])
}

func (*Pipefs) Open(req *SrvReq) {
	fid := req.Fid.Aux.(*pipeFid)
	tc := req.Tc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	var e error
	fid.file, e = os.OpenFile(fid.path, omode2uflags(tc.Mode), 0)
	if e != nil {
		req.RespondError(toError(e))
		return
	}

	req.RespondRopen(dir2Qid(fid.st), 0)
}

func (*Pipefs) Create(req *SrvReq) {
	fid := req.Fid.Aux.(*pipeFid)
	tc := req.Tc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	path := fid.path + "/" + tc.Name
	var e error = nil
	var file *os.File = nil
	switch {
	case tc.Perm&DMDIR != 0:
		e = os.Mkdir(path, os.FileMode(tc.Perm&0o777))

	case tc.Perm&DMSYMLINK != 0:
		e = os.Symlink(tc.Ext, path)

	case tc.Perm&DMLINK != 0:
		n, e := strconv.ParseUint(tc.Ext, 10, 0)
		if e != nil {
			break
		}

		ofid := req.Conn.FidGet(uint32(n))
		if ofid == nil {
			req.RespondError(Eunknownfid)
			return
		}

		e = os.Link(ofid.Aux.(*pipeFid).path, path)
		ofid.DecRef()

	case tc.Perm&DMNAMEDPIPE != 0:
	case tc.Perm&DMDEVICE != 0:
		req.RespondError(&Error{"not implemented", EIO})
		return

	default:
		var mode uint32 = tc.Perm & 0o777
		if req.Conn.Dotu {
			if tc.Perm&DMSETUID > 0 {
				mode |= syscall.S_ISUID
			}
			if tc.Perm&DMSETGID > 0 {
				mode |= syscall.S_ISGID
			}
		}
		file, e = os.OpenFile(path, omode2uflags(tc.Mode)|os.O_CREATE, os.FileMode(mode))
	}

	if file == nil && e == nil {
		file, e = os.OpenFile(path, omode2uflags(tc.Mode), 0)
	}

	if e != nil {
		req.RespondError(toError(e))
		return
	}

	fid.path = path
	fid.file = file
	err = fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	req.RespondRcreate(dir2Qid(fid.st), 0)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (*Pipefs) Read(req *SrvReq) {
	fid := req.Fid.Aux.(*pipeFid)
	tc := req.Tc
	rc := req.Rc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	InitRread(rc, tc.Count)
	var count int
	var e error
	if fid.st.IsDir() {
		if tc.Offset == 0 {
			var e error
			// If we got here, it was open. Can't really seek
			// in most cases, just close and reopen it.
			fid.file.Close()
			if fid.file, e = os.OpenFile(fid.path, omode2uflags(req.Fid.Omode), 0); e != nil {
				req.RespondError(toError(e))
				return
			}

			if fid.dirs, e = fid.file.Readdir(-1); e != nil {
				req.RespondError(toError(e))
				return
			}
			fid.dirents = nil

			for i := 0; i < len(fid.dirs); i++ {
				path := fid.path + "/" + fid.dirs[i].Name()
				st, _ := dir2Dir(path, fid.dirs[i], req.Conn.Dotu, req.Conn.Srv.Upool)
				if st == nil {
					continue
				}
				b := PackDir(st, req.Conn.Dotu)
				fid.dirents = append(fid.dirents, b...)
				count += len(b)
			}
		}
		switch {
		case tc.Offset > uint64(len(fid.dirents)):
			count = 0
		case len(fid.dirents[tc.Offset:]) > int(tc.Count):
			count = int(tc.Count)
		default:
			count = len(fid.dirents[tc.Offset:])
		}

		copy(rc.Data, fid.dirents[tc.Offset:int(tc.Offset)+count])

	} else {
		if fid.eof {
			req.RespondError(toError(e))
			return
		}
		length := min(len(rc.Data), len(fid.data))
		count = length
		copy(rc.Data, fid.data[:length])
		fid.data = fid.data[length:]
	}

	SetRreadCount(rc, uint32(count))
	req.Respond()
}

func (*Pipefs) Write(req *SrvReq) {
	fid := req.Fid.Aux.(*pipeFid)
	tc := req.Tc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	fid.data = append(fid.data, tc.Data...)

	req.RespondRwrite(uint32(len(tc.Data)))
}

func (*Pipefs) Clunk(req *SrvReq) { req.RespondRclunk() }

func (*Pipefs) Remove(req *SrvReq) {
	fid := req.Fid.Aux.(*pipeFid)
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	e := os.Remove(fid.path)
	if e != nil {
		req.RespondError(toError(e))
		return
	}

	req.RespondRremove()
}

func (*Pipefs) Stat(req *SrvReq) {
	fid := req.Fid.Aux.(*pipeFid)
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	st, derr := dir2Dir(fid.path, fid.st, req.Conn.Dotu, req.Conn.Srv.Upool)
	if st == nil {
		req.RespondError(derr)
		return
	}

	req.RespondRstat(st)
}

func (*Pipefs) Wstat(req *SrvReq) {
	req.RespondError(Eperm)
}
