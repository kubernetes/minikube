// Copyright 2009 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ufs

import (
	"io"
	"log"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"k8s.io/minikube/third_party/go9p/p"
	"k8s.io/minikube/third_party/go9p/p/srv"
)

type Fid struct {
	path      string
	file      *os.File
	dirs      []os.FileInfo
	diroffset uint64
	st        os.FileInfo
}

type Ufs struct {
	srv.Srv
}

var addr string
var debug int
var root string
var Enoent = &p.Error{"file not found", p.ENOENT}

func toError(err error) *p.Error {
	var ecode uint32

	ename := err.Error()
	if e, ok := err.(syscall.Errno); ok {
		ecode = uint32(e)
	} else {
		ecode = p.EIO
	}

	return &p.Error{ename, ecode}
}

func (fid *Fid) stat() *p.Error {
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
	case p.OREAD:
		ret = os.O_RDONLY
		break

	case p.ORDWR:
		ret = os.O_RDWR
		break

	case p.OWRITE:
		ret = os.O_WRONLY
		break

	case p.OEXEC:
		ret = os.O_RDONLY
		break
	}

	if mode&p.OTRUNC != 0 {
		ret |= os.O_TRUNC
	}

	return ret
}

func dir2QidType(d os.FileInfo) uint8 {
	ret := uint8(0)
	if d.IsDir() {
		ret |= p.QTDIR
	}

	if d.Mode()&os.ModeSymlink != 0 {
		ret |= p.QTSYMLINK
	}

	return ret
}

func dir2Npmode(d os.FileInfo, dotu bool) uint32 {
	ret := uint32(d.Mode() & 0777)
	if d.IsDir() {
		ret |= p.DMDIR
	}

	if dotu {
		mode := d.Mode()
		if mode&os.ModeSymlink != 0 {
			ret |= p.DMSYMLINK
		}

		if mode&os.ModeSocket != 0 {
			ret |= p.DMSOCKET
		}

		if mode&os.ModeNamedPipe != 0 {
			ret |= p.DMNAMEDPIPE
		}

		if mode&os.ModeDevice != 0 {
			ret |= p.DMDEVICE
		}

		if mode&os.ModeSetuid != 0 {
			ret |= p.DMSETUID
		}

		if mode&os.ModeSetgid != 0 {
			ret |= p.DMSETGID
		}
	}

	return ret
}

// Dir is an instantiation of the p.Dir structure
// that can act as a receiver for local methods.
type Dir struct {
	p.Dir
}

func (*Ufs) ConnOpened(conn *srv.Conn) {
	if conn.Srv.Debuglevel > 0 {
		log.Println("connected")
	}
}

func (*Ufs) ConnClosed(conn *srv.Conn) {
	if conn.Srv.Debuglevel > 0 {
		log.Println("disconnected")
	}
}

func (*Ufs) FidDestroy(sfid *srv.Fid) {
	var fid *Fid

	if sfid.Aux == nil {
		return
	}

	fid = sfid.Aux.(*Fid)
	if fid.file != nil {
		fid.file.Close()
	}
}

func (*Ufs) Attach(req *srv.Req) {
	if req.Afid != nil {
		req.RespondError(srv.Enoauth)
		return
	}

	tc := req.Tc
	fid := new(Fid)
	if len(tc.Aname) == 0 {
		fid.path = root
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

func (*Ufs) Flush(req *srv.Req) {}

func (*Ufs) Walk(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	tc := req.Tc

	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	if req.Newfid.Aux == nil {
		req.Newfid.Aux = new(Fid)
	}

	nfid := req.Newfid.Aux.(*Fid)
	wqids := make([]p.Qid, len(tc.Wname))
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

func (*Ufs) Open(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
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

func (*Ufs) Create(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
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
	case tc.Perm&p.DMDIR != 0:
		e = os.Mkdir(path, os.FileMode(tc.Perm&0777))

	case tc.Perm&p.DMSYMLINK != 0:
		e = os.Symlink(tc.Ext, path)

	case tc.Perm&p.DMLINK != 0:
		n, e := strconv.ParseUint(tc.Ext, 10, 0)
		if e != nil {
			break
		}

		ofid := req.Conn.FidGet(uint32(n))
		if ofid == nil {
			req.RespondError(srv.Eunknownfid)
			return
		}

		e = os.Link(ofid.Aux.(*Fid).path, path)
		ofid.DecRef()

	case tc.Perm&p.DMNAMEDPIPE != 0:
	case tc.Perm&p.DMDEVICE != 0:
		req.RespondError(&p.Error{"not implemented", p.EIO})
		return

	default:
		var mode uint32 = tc.Perm & 0777
		if req.Conn.Dotu {
			if tc.Perm&p.DMSETUID > 0 {
				mode |= syscall.S_ISUID
			}
			if tc.Perm&p.DMSETGID > 0 {
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

func (*Ufs) Read(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	tc := req.Tc
	rc := req.Rc
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	p.InitRread(rc, tc.Count)
	var count int
	var e error
	if fid.st.IsDir() {
		b := rc.Data
		if tc.Offset == 0 {
			fid.file.Close()
			fid.file, e = os.OpenFile(fid.path, omode2uflags(req.Fid.Omode), 0)
			if e != nil {
				req.RespondError(toError(e))
				return
			}
		}

		for len(b) > 0 {
			if fid.dirs == nil {
				fid.dirs, e = fid.file.Readdir(16)
				if e != nil && e != io.EOF {
					req.RespondError(toError(e))
					return
				}

				if len(fid.dirs) == 0 {
					break
				}
			}

			var i int
			for i = 0; i < len(fid.dirs); i++ {
				path := fid.path + "/" + fid.dirs[i].Name()
				st := dir2Dir(path, fid.dirs[i], req.Conn.Dotu, req.Conn.Srv.Upool)
				sz := p.PackDir(st, b, req.Conn.Dotu)
				if sz == 0 {
					break
				}

				b = b[sz:]
				count += sz
			}

			if i < len(fid.dirs) {
				fid.dirs = fid.dirs[i:]
				break
			} else {
				fid.dirs = nil
			}
		}
	} else {
		count, e = fid.file.ReadAt(rc.Data, int64(tc.Offset))
		if e != nil && e != io.EOF {
			req.RespondError(toError(e))
			return
		}
	}

	p.SetRreadCount(rc, uint32(count))
	req.Respond()
}

func (*Ufs) Write(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
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

func (*Ufs) Clunk(req *srv.Req) { req.RespondRclunk() }

func (*Ufs) Remove(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
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

func (*Ufs) Stat(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	st := dir2Dir(fid.path, fid.st, req.Conn.Dotu, req.Conn.Srv.Upool)
	req.RespondRstat(st)
}

func lookup(uid string, group bool) (uint32, *p.Error) {
	if uid == "" {
		return p.NOUID, nil
	}
	usr, e := user.Lookup(uid)
	if e != nil {
		return p.NOUID, toError(e)
	}
	conv := usr.Uid
	if group {
		conv = usr.Gid
	}
	u, e := strconv.Atoi(conv)
	if e != nil {
		return p.NOUID, toError(e)
	}
	return uint32(u), nil
}

func StartServer(addrVal string, debugVal int, rootVal string) {
	addr = addrVal
	debug = debugVal
	root = rootVal
	ufs := new(Ufs)
	ufs.Dotu = true
	ufs.Id = "ufs"
	ufs.Debuglevel = debug
	ufs.Start(ufs)

	err := ufs.StartNetListener("tcp", addr)
	if err != nil {
		log.Println(err)
	}
}
