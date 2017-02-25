// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srv

import (
	"k8s.io/minikube/third_party/go9p/p"
	"log"
	"sync"
	"time"
)

// The FStatOp interface provides a single operation (Stat) that will be
// called before a file stat is sent back to the client. If implemented,
// the operation should update the data in the File struct.
type FStatOp interface {
	Stat(fid *FFid) error
}

// The FWstatOp interface provides a single operation (Wstat) that will be
// called when the client requests the File metadata to be modified. If
// implemented, the operation will be called when Twstat message is received.
// If not implemented, "permission denied" error will be sent back. If the
// operation returns an Error, the error is send back to the client.
type FWstatOp interface {
	Wstat(*FFid, *p.Dir) error
}

// If the FReadOp interface is implemented, the Read operation will be called
// to read from the file. If not implemented, "permission denied" error will
// be send back. The operation returns the number of bytes read, or the
// error occured while reading.
type FReadOp interface {
	Read(fid *FFid, buf []byte, offset uint64) (int, error)
}

// If the FWriteOp interface is implemented, the Write operation will be called
// to write to the file. If not implemented, "permission denied" error will
// be send back. The operation returns the number of bytes written, or the
// error occured while writing.
type FWriteOp interface {
	Write(fid *FFid, data []byte, offset uint64) (int, error)
}

// If the FCreateOp interface is implemented, the Create operation will be called
// when the client attempts to create a file in the File implementing the interface.
// If not implemented, "permission denied" error will be send back. If successful,
// the operation should call (*File)Add() to add the created file to the directory.
// The operation returns the created file, or the error occured while creating it.
type FCreateOp interface {
	Create(fid *FFid, name string, perm uint32) (*File, error)
}

// If the FRemoveOp interface is implemented, the Remove operation will be called
// when the client attempts to create a file in the File implementing the interface.
// If not implemented, "permission denied" error will be send back.
// The operation returns nil if successful, or the error that occured while removing
// the file.
type FRemoveOp interface {
	Remove(*FFid) error
}

type FOpenOp interface {
	Open(fid *FFid, mode uint8) error
}

type FClunkOp interface {
	Clunk(fid *FFid) error
}

type FDestroyOp interface {
	FidDestroy(fid *FFid)
}

type FFlags int

const (
	Fremoved FFlags = 1 << iota
)

// The File type represents a file (or directory) served by the file server.
type File struct {
	sync.Mutex
	p.Dir
	flags FFlags

	Parent        *File // parent
	next, prev    *File // siblings, guarded by parent.Lock
	cfirst, clast *File // children (if directory)
	Ops           interface{}
}

type FFid struct {
	F    *File
	Fid  *Fid
	dirs []*File // used for readdir
}

// The Fsrv can be used to create file servers that serve
// simple trees of synthetic files.
type Fsrv struct {
	Srv
	Root *File
}

var lock sync.Mutex
var qnext uint64
var Eexist = &p.Error{"file already exists", p.EEXIST}
var Enoent = &p.Error{"file not found", p.ENOENT}
var Enotempty = &p.Error{"directory not empty", p.EPERM}

// Creates a file server with root as root directory
func NewFileSrv(root *File) *Fsrv {
	srv := new(Fsrv)
	srv.Root = root
	root.Parent = root // make sure we can .. in root

	return srv
}

// Initializes the fields of a file and add it to a directory.
// Returns nil if successful, or an error.
func (f *File) Add(dir *File, name string, uid p.User, gid p.Group, mode uint32, ops interface{}) error {

	lock.Lock()
	qpath := qnext
	qnext++
	lock.Unlock()

	f.Qid.Type = uint8(mode >> 24)
	f.Qid.Version = 0
	f.Qid.Path = qpath
	f.Mode = mode
	f.Atime = uint32(time.Now().Unix())
	f.Mtime = f.Atime
	f.Length = 0
	f.Name = name
	if uid != nil {
		f.Uid = uid.Name()
		f.Uidnum = uint32(uid.Id())
	} else {
		f.Uid = "none"
		f.Uidnum = p.NOUID
	}

	if gid != nil {
		f.Gid = gid.Name()
		f.Gidnum = uint32(gid.Id())
	} else {
		f.Gid = "none"
		f.Gidnum = p.NOUID
	}

	f.Muid = ""
	f.Muidnum = p.NOUID
	f.Ext = ""

	if dir != nil {
		f.Parent = dir
		dir.Lock()
		for p := dir.cfirst; p != nil; p = p.next {
			if name == p.Name {
				dir.Unlock()
				return Eexist
			}
		}

		if dir.clast != nil {
			dir.clast.next = f
		} else {
			dir.cfirst = f
		}

		f.prev = dir.clast
		f.next = nil
		dir.clast = f
		dir.Unlock()
	} else {
		f.Parent = f
	}

	f.Ops = ops
	return nil
}

// Removes a file from its parent directory.
func (f *File) Remove() {
	f.Lock()
	if (f.flags & Fremoved) != 0 {
		f.Unlock()
		return
	}

	f.flags |= Fremoved
	f.Unlock()

	p := f.Parent
	p.Lock()
	if f.next != nil {
		f.next.prev = f.prev
	} else {
		p.clast = f.prev
	}

	if f.prev != nil {
		f.prev.next = f.next
	} else {
		p.cfirst = f.next
	}

	f.next = nil
	f.prev = nil
	p.Unlock()
}

func (f *File) Rename(name string) error {
	p := f.Parent
	p.Lock()
	defer p.Unlock()
	for c := p.cfirst; c != nil; c = c.next {
		if name == c.Name {
			return Eexist
		}
	}

	f.Name = name
	return nil
}

// Looks for a file in a directory. Returns nil if the file is not found.
func (p *File) Find(name string) *File {
	var f *File

	p.Lock()
	for f = p.cfirst; f != nil; f = f.next {
		if name == f.Name {
			break
		}
	}
	p.Unlock()
	return f
}

// Checks if the specified user has permission to perform
// certain operation on a file. Perm contains one or more
// of p.DMREAD, p.DMWRITE, and p.DMEXEC.
func (f *File) CheckPerm(user p.User, perm uint32) bool {
	if user == nil {
		return false
	}

	perm &= 7

	/* other permissions */
	fperm := f.Mode & 7
	if (fperm & perm) == perm {
		return true
	}

	/* user permissions */
	if f.Uid == user.Name() || f.Uidnum == uint32(user.Id()) {
		fperm |= (f.Mode >> 6) & 7
	}

	if (fperm & perm) == perm {
		return true
	}

	/* group permissions */
	groups := user.Groups()
	if groups != nil && len(groups) > 0 {
		for i := 0; i < len(groups); i++ {
			if f.Gid == groups[i].Name() || f.Gidnum == uint32(groups[i].Id()) {
				fperm |= (f.Mode >> 3) & 7
				break
			}
		}
	}

	if (fperm & perm) == perm {
		return true
	}

	return false
}

func (s *Fsrv) Attach(req *Req) {
	fid := new(FFid)
	fid.F = s.Root
	fid.Fid = req.Fid
	req.Fid.Aux = fid
	req.RespondRattach(&s.Root.Qid)
}

func (*Fsrv) Walk(req *Req) {
	fid := req.Fid.Aux.(*FFid)
	tc := req.Tc

	if req.Newfid.Aux == nil {
		nfid := new(FFid)
		nfid.Fid = req.Newfid
		req.Newfid.Aux = nfid
	}

	nfid := req.Newfid.Aux.(*FFid)
	wqids := make([]p.Qid, len(tc.Wname))
	i := 0
	f := fid.F
	for ; i < len(tc.Wname); i++ {
		if tc.Wname[i] == ".." {
			// handle dotdot
			f = f.Parent
			wqids[i] = f.Qid
			continue
		}
		if (wqids[i].Type & p.QTDIR) > 0 {
			if !f.CheckPerm(req.Fid.User, p.DMEXEC) {
				break
			}
		}

		p := f.Find(tc.Wname[i])
		if p == nil {
			break
		}

		f = p
		wqids[i] = f.Qid
	}

	if len(tc.Wname) > 0 && i == 0 {
		req.RespondError(Enoent)
		return
	}

	nfid.F = f
	req.RespondRwalk(wqids[0:i])
}

func mode2Perm(mode uint8) uint32 {
	var perm uint32 = 0

	switch mode & 3 {
	case p.OREAD:
		perm = p.DMREAD
	case p.OWRITE:
		perm = p.DMWRITE
	case p.ORDWR:
		perm = p.DMREAD | p.DMWRITE
	}

	if (mode & p.OTRUNC) != 0 {
		perm |= p.DMWRITE
	}

	return perm
}

func (*Fsrv) Open(req *Req) {
	fid := req.Fid.Aux.(*FFid)
	tc := req.Tc

	if !fid.F.CheckPerm(req.Fid.User, mode2Perm(tc.Mode)) {
		req.RespondError(Eperm)
		return
	}

	if op, ok := (fid.F.Ops).(FOpenOp); ok {
		err := op.Open(fid, tc.Mode)
		if err != nil {
			req.RespondError(err)
			return
		}
	}
	req.RespondRopen(&fid.F.Qid, 0)
}

func (*Fsrv) Create(req *Req) {
	fid := req.Fid.Aux.(*FFid)
	tc := req.Tc

	dir := fid.F
	if !dir.CheckPerm(req.Fid.User, p.DMWRITE) {
		req.RespondError(Eperm)
		return
	}

	if cop, ok := (dir.Ops).(FCreateOp); ok {
		f, err := cop.Create(fid, tc.Name, tc.Perm)
		if err != nil {
			req.RespondError(err)
		} else {
			fid.F = f
			req.RespondRcreate(&fid.F.Qid, 0)
		}
	} else {
		req.RespondError(Eperm)
	}
}

func (*Fsrv) Read(req *Req) {
	var i, n int
	var err error

	fid := req.Fid.Aux.(*FFid)
	f := fid.F
	tc := req.Tc
	rc := req.Rc
	p.InitRread(rc, tc.Count)

	if f.Mode&p.DMDIR != 0 {
		// directory
		if tc.Offset == 0 {
			var g *File
			f.Lock()
			for n, g = 0, f.cfirst; g != nil; n, g = n+1, g.next {
			}

			fid.dirs = make([]*File, n)
			for n, g = 0, f.cfirst; g != nil; n, g = n+1, g.next {
				fid.dirs[n] = g
			}
			f.Unlock()
		}

		n = 0
		b := rc.Data
		for i = 0; i < len(fid.dirs); i++ {
			g := fid.dirs[i]
			g.Lock()
			if (g.flags & Fremoved) != 0 {
				g.Unlock()
				continue
			}

			sz := p.PackDir(&g.Dir, b, req.Conn.Dotu)
			g.Unlock()
			if sz == 0 {
				break
			}

			b = b[sz:]
			n += sz
		}

		fid.dirs = fid.dirs[i:]
	} else {
		// file
		if rop, ok := f.Ops.(FReadOp); ok {
			n, err = rop.Read(fid, rc.Data, tc.Offset)
			if err != nil {
				req.RespondError(err)
				return
			}
		} else {
			req.RespondError(Eperm)
			return
		}
	}

	p.SetRreadCount(rc, uint32(n))
	req.Respond()
}

func (*Fsrv) Write(req *Req) {
	fid := req.Fid.Aux.(*FFid)
	f := fid.F
	tc := req.Tc

	if wop, ok := (f.Ops).(FWriteOp); ok {
		n, err := wop.Write(fid, tc.Data, tc.Offset)
		if err != nil {
			req.RespondError(err)
		} else {
			req.RespondRwrite(uint32(n))
		}
	} else {
		req.RespondError(Eperm)
	}

}

func (*Fsrv) Clunk(req *Req) {
	fid := req.Fid.Aux.(*FFid)

	if op, ok := (fid.F.Ops).(FClunkOp); ok {
		err := op.Clunk(fid)
		if err != nil {
			req.RespondError(err)
		}
	}
	req.RespondRclunk()
}

func (*Fsrv) Remove(req *Req) {
	fid := req.Fid.Aux.(*FFid)
	f := fid.F
	f.Lock()
	if f.cfirst != nil {
		f.Unlock()
		req.RespondError(Enotempty)
		return
	}
	f.Unlock()

	if rop, ok := (f.Ops).(FRemoveOp); ok {
		err := rop.Remove(fid)
		if err != nil {
			req.RespondError(err)
		} else {
			f.Remove()
			req.RespondRremove()
		}
	} else {
		log.Println("remove not implemented")
		req.RespondError(Eperm)
	}
}

func (*Fsrv) Stat(req *Req) {
	fid := req.Fid.Aux.(*FFid)
	f := fid.F

	if sop, ok := (f.Ops).(FStatOp); ok {
		err := sop.Stat(fid)
		if err != nil {
			req.RespondError(err)
		} else {
			req.RespondRstat(&f.Dir)
		}
	} else {
		req.RespondRstat(&f.Dir)
	}
}

func (*Fsrv) Wstat(req *Req) {
	tc := req.Tc
	fid := req.Fid.Aux.(*FFid)
	f := fid.F

	if wop, ok := (f.Ops).(FWstatOp); ok {
		err := wop.Wstat(fid, &tc.Dir)
		if err != nil {
			req.RespondError(err)
		} else {
			req.RespondRwstat()
		}
	} else {
		req.RespondError(Eperm)
	}
}

func (*Fsrv) FidDestroy(ffid *Fid) {
	if ffid.Aux == nil {
		return
	}
	fid := ffid.Aux.(*FFid)
	f := fid.F

	if f == nil {
		return // otherwise errs in bad walks
	}

	if op, ok := (f.Ops).(FDestroyOp); ok {
		op.FidDestroy(fid)
	}
}
