// Copyright 2009 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import (
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"sort"
	"strconv"
	"syscall"
)

type ufsFid struct {
	path       string
	file       *os.File
	dirs       []os.FileInfo
	direntends []int
	dirents    []byte
	diroffset  uint64
	st         os.FileInfo
}

type Ufs struct {
	Srv
	Root string
}

func toError(err error) *Error {
	var ecode uint32

	ename := err.Error()
	if e, ok := err.(syscall.Errno); ok {
		ecode = uint32(e)
	} else {
		ecode = EIO
	}

	return &Error{ename, ecode}
}

func (fid *ufsFid) stat() *Error {
	var err error

	fid.st, err = os.Lstat(fid.path)
	if err != nil {
		return toError(err)
	}

	return nil
}

func omode2uflags(mode uint8) int {
	ret := int(0)
	switch mode & 3 {
	case OREAD:
		ret = os.O_RDONLY
		break

	case ORDWR:
		ret = os.O_RDWR
		break

	case OWRITE:
		ret = os.O_WRONLY
		break

	case OEXEC:
		ret = os.O_RDONLY
		break
	}

	if mode&OTRUNC != 0 {
		ret |= os.O_TRUNC
	}

	return ret
}

func dir2QidType(d os.FileInfo) uint8 {
	ret := uint8(0)
	if d.IsDir() {
		ret |= QTDIR
	}

	if d.Mode()&os.ModeSymlink != 0 {
		ret |= QTSYMLINK
	}

	return ret
}

func dir2Npmode(d os.FileInfo, dotu bool) uint32 {
	ret := uint32(d.Mode() & 0o777)
	if d.IsDir() {
		ret |= DMDIR
	}

	if dotu {
		mode := d.Mode()
		if mode&os.ModeSymlink != 0 {
			ret |= DMSYMLINK
		}

		if mode&os.ModeSocket != 0 {
			ret |= DMSOCKET
		}

		if mode&os.ModeNamedPipe != 0 {
			ret |= DMNAMEDPIPE
		}

		if mode&os.ModeDevice != 0 {
			ret |= DMDEVICE
		}

		if mode&os.ModeSetuid != 0 {
			ret |= DMSETUID
		}

		if mode&os.ModeSetgid != 0 {
			ret |= DMSETGID
		}
	}

	return ret
}

// Dir is an instantiation of the p.Dir structure
// that can act as a receiver for local methods.
type ufsDir struct {
	Dir
}

func (*Ufs) ConnOpened(conn *Conn) {
	if conn.Srv.Debuglevel > 0 {
		log.Println("connected")
	}
}

func (*Ufs) ConnClosed(conn *Conn) {
	if conn.Srv.Debuglevel > 0 {
		log.Println("disconnected")
	}
}

func (*Ufs) FidDestroy(sfid *SrvFid) {
	var fid *ufsFid

	if sfid.Aux == nil {
		return
	}

	fid = sfid.Aux.(*ufsFid)
	if fid.file != nil {
		fid.file.Close()
	}
}

func (ufs *Ufs) Attach(req *SrvReq) {
	if req.Afid != nil {
		req.RespondError(Enoauth)
		return
	}

	tc := req.Tc
	fid := new(ufsFid)
	// You can think of the ufs.Root as a 'chroot' of a sort.
	// clients attach are not allowed to go outside the
	// directory represented by ufs.Root
	fid.path = path.Join(ufs.Root, tc.Aname)

	req.Fid.Aux = fid
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	qid := dir2Qid(fid.st)
	req.RespondRattach(qid)
}

func (*Ufs) Flush(req *SrvReq) {}

func (*Ufs) Walk(req *SrvReq) {
	fid := req.Fid.Aux.(*ufsFid)
	tc := req.Tc

	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	if req.Newfid.Aux == nil {
		req.Newfid.Aux = new(ufsFid)
	}

	nfid := req.Newfid.Aux.(*ufsFid)
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

func (*Ufs) Open(req *SrvReq) {
	fid := req.Fid.Aux.(*ufsFid)
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

func (*Ufs) Create(req *SrvReq) {
	fid := req.Fid.Aux.(*ufsFid)
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

		e = os.Link(ofid.Aux.(*ufsFid).path, path)
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

func (*Ufs) Read(req *SrvReq) {
	fid := req.Fid.Aux.(*ufsFid)
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
			fid.direntends = nil
			for i := 0; i < len(fid.dirs); i++ {
				path := fid.path + "/" + fid.dirs[i].Name()
				st, _ := dir2Dir(path, fid.dirs[i], req.Conn.Dotu, req.Conn.Srv.Upool)
				if st == nil {
					continue
				}
				b := PackDir(st, req.Conn.Dotu)
				fid.dirents = append(fid.dirents, b...)
				count += len(b)
				fid.direntends = append(fid.direntends, count)
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

		if !*Akaros {
			nextend := sort.SearchInts(fid.direntends, int(tc.Offset)+count)
			if nextend < len(fid.direntends) {
				if fid.direntends[nextend] > int(tc.Offset)+count {
					if nextend > 0 {
						count = fid.direntends[nextend-1] - int(tc.Offset)
					} else {
						count = 0
					}
				}
			}
			if count == 0 && int(tc.Offset) < len(fid.dirents) && len(fid.dirents) > 0 {
				req.RespondError(&Error{"too small read size for dir entry", EINVAL})
				return
			}
		}

		copy(rc.Data, fid.dirents[tc.Offset:int(tc.Offset)+count])

	} else {
		count, e = fid.file.ReadAt(rc.Data, int64(tc.Offset))
		if e != nil && e != io.EOF {
			req.RespondError(toError(e))
			return
		}

	}

	SetRreadCount(rc, uint32(count))
	req.Respond()
}

func (*Ufs) Write(req *SrvReq) {
	fid := req.Fid.Aux.(*ufsFid)
	tc := req.Tc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	n, e := fid.file.WriteAt(tc.Data, int64(tc.Offset))
	if e != nil {
		req.RespondError(toError(e))
		return
	}

	req.RespondRwrite(uint32(n))
}

func (*Ufs) Clunk(req *SrvReq) { req.RespondRclunk() }

func (*Ufs) Remove(req *SrvReq) {
	fid := req.Fid.Aux.(*ufsFid)
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

func (*Ufs) Stat(req *SrvReq) {
	fid := req.Fid.Aux.(*ufsFid)
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

func lookup(uid string, group bool) (uint32, *Error) {
	if uid == "" {
		return NOUID, nil
	}
	usr, e := user.Lookup(uid)
	if e != nil {
		return NOUID, toError(e)
	}
	conv := usr.Uid
	if group {
		conv = usr.Gid
	}
	u, e := strconv.Atoi(conv)
	if e != nil {
		return NOUID, toError(e)
	}
	return uint32(u), nil
}

/* enables "Akaros" capabilities, which right now means
 * a sane error message format.
 */

// (r2d4): We don't want this exposed in minikube right now
// var Akaros = flag.Bool("akaros", false, "Akaros extensions")
var Akaros = boolPointer(false)

func boolPointer(b bool) *bool {
	return &b
}
