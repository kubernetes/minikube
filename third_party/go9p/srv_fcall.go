// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import "runtime"

func (srv *Srv) version(req *SrvReq) {
	tc := req.Tc
	conn := req.Conn

	if tc.Msize < IOHDRSZ {
		req.RespondError(&Error{"msize too small", EINVAL})
		return
	}

	if tc.Msize < conn.Msize {
		conn.Msize = tc.Msize
	}

	conn.Dotu = tc.Version == "9P2000.u" && srv.Dotu
	ver := "9P2000"
	if conn.Dotu {
		ver = "9P2000.u"
	}

	/* make sure that the responses of all current requests will be ignored */
	conn.Lock()
	for tag, r := range conn.reqs {
		if tag == NOTAG {
			continue
		}

		for rr := r; rr != nil; rr = rr.next {
			rr.Lock()
			rr.status |= reqFlush
			rr.Unlock()
		}
	}
	conn.Unlock()

	req.RespondRversion(conn.Msize, ver)
}

func (srv *Srv) auth(req *SrvReq) {
	tc := req.Tc
	conn := req.Conn
	if tc.Afid == NOFID {
		req.RespondError(Eunknownfid)
		return
	}

	req.Afid = conn.FidNew(tc.Afid)
	if req.Afid == nil {
		req.RespondError(Einuse)
		return
	}

	var user User = nil
	if tc.Unamenum != NOUID || conn.Dotu {
		user = srv.Upool.Uid2User(int(tc.Unamenum))
	} else if tc.Uname != "" {
		user = srv.Upool.Uname2User(tc.Uname)
	}

	if runtime.GOOS != "windows" {
		if user == nil {
			req.RespondError(Enouser)
			return
		}
	}

	req.Afid.User = user
	req.Afid.Type = QTAUTH
	if aop, ok := (srv.ops).(AuthOps); ok {
		aqid, err := aop.AuthInit(req.Afid, tc.Aname)
		if err != nil {
			req.RespondError(err)
		} else {
			aqid.Type |= QTAUTH // just in case
			req.RespondRauth(aqid)
		}
	} else {
		req.RespondError(Enoauth)
	}

}

func (srv *Srv) authPost(req *SrvReq) {
	if req.Rc != nil && req.Rc.Type == Rauth {
		req.Afid.IncRef()
	}
}

func (srv *Srv) attach(req *SrvReq) {
	tc := req.Tc
	conn := req.Conn
	if tc.Fid == NOFID {
		req.RespondError(Eunknownfid)
		return
	}

	req.Fid = conn.FidNew(tc.Fid)
	if req.Fid == nil {
		req.RespondError(Einuse)
		return
	}

	if tc.Afid != NOFID {
		req.Afid = conn.FidGet(tc.Afid)
		if req.Afid == nil {
			req.RespondError(Eunknownfid)
		}
	}

	var user User = nil
	if tc.Unamenum != NOUID || conn.Dotu {
		user = srv.Upool.Uid2User(int(tc.Unamenum))
	} else if tc.Uname != "" {
		user = srv.Upool.Uname2User(tc.Uname)
	}

	if runtime.GOOS != "windows" {
		if user == nil {
			req.RespondError(Enouser)
			return
		}
	}

	req.Fid.User = user
	if aop, ok := (srv.ops).(AuthOps); ok {
		err := aop.AuthCheck(req.Fid, req.Afid, tc.Aname)
		if err != nil {
			req.RespondError(err)
			return
		}
	}

	(srv.ops).(SrvReqOps).Attach(req)
}

func (srv *Srv) attachPost(req *SrvReq) {
	if req.Rc != nil && req.Rc.Type == Rattach {
		req.Fid.Type = req.Rc.Qid.Type
		req.Fid.IncRef()
	}
}

func (srv *Srv) flush(req *SrvReq) {
	conn := req.Conn
	tag := req.Tc.Oldtag
	PackRflush(req.Rc)
	conn.Lock()
	r := conn.reqs[tag]
	if r != nil {
		req.flushreq = r.flushreq
		r.flushreq = req
	}
	conn.Unlock()

	if r == nil {
		// there are no requests with that tag
		req.Respond()
		return
	}

	r.Lock()
	status := r.status
	if (status & (reqWork | reqSaved)) == 0 {
		/* the request is not worked on yet */
		r.status |= reqFlush
	}
	r.Unlock()

	if (status & (reqWork | reqSaved)) == 0 {
		r.Respond()
	} else {
		if op, ok := (srv.ops).(FlushOp); ok {
			op.Flush(r)
		}
	}
}

func (srv *Srv) walk(req *SrvReq) {
	conn := req.Conn
	tc := req.Tc
	fid := req.Fid

	/* we can't walk regular files, only clone them */
	if len(tc.Wname) > 0 && (fid.Type&QTDIR) == 0 {
		req.RespondError(Enotdir)
		return
	}

	/* we can't walk open files */
	if fid.opened {
		req.RespondError(Ebaduse)
		return
	}

	if tc.Fid != tc.Newfid {
		req.Newfid = conn.FidNew(tc.Newfid)
		if req.Newfid == nil {
			req.RespondError(Einuse)
			return
		}

		req.Newfid.User = fid.User
		req.Newfid.Type = fid.Type
	} else {
		req.Newfid = req.Fid
		req.Newfid.IncRef()
	}

	(req.Conn.Srv.ops).(SrvReqOps).Walk(req)
}

func (srv *Srv) walkPost(req *SrvReq) {
	rc := req.Rc
	if rc == nil || rc.Type != Rwalk || req.Newfid == nil {
		return
	}

	n := len(rc.Wqid)
	if n > 0 {
		req.Newfid.Type = rc.Wqid[n-1].Type
	} else {
		req.Newfid.Type = req.Fid.Type
	}

	// Don't retain the fid if only a partial walk succeeded
	if n != len(req.Tc.Wname) {
		return
	}

	if req.Newfid.fid != req.Fid.fid {
		req.Newfid.IncRef()
	}
}

func (srv *Srv) open(req *SrvReq) {
	fid := req.Fid
	tc := req.Tc
	if fid.opened {
		req.RespondError(Eopen)
		return
	}

	if (fid.Type&QTDIR) != 0 && tc.Mode != OREAD {
		req.RespondError(Eperm)
		return
	}

	fid.Omode = tc.Mode
	(req.Conn.Srv.ops).(SrvReqOps).Open(req)
}

func (srv *Srv) openPost(req *SrvReq) {
	if req.Fid != nil {
		req.Fid.opened = req.Rc != nil && req.Rc.Type == Ropen
	}
}

func (srv *Srv) create(req *SrvReq) {
	fid := req.Fid
	tc := req.Tc
	if fid.opened {
		req.RespondError(Eopen)
		return
	}

	if (fid.Type & QTDIR) == 0 {
		req.RespondError(Enotdir)
		return
	}

	/* can't open directories for other than reading */
	if (tc.Perm&DMDIR) != 0 && tc.Mode != OREAD {
		req.RespondError(Eperm)
		return
	}

	/* can't create special files if not 9P2000.u */
	if (tc.Perm&(DMNAMEDPIPE|DMSYMLINK|DMLINK|DMDEVICE|DMSOCKET)) != 0 && !req.Conn.Dotu {
		req.RespondError(Eperm)
		return
	}

	fid.Omode = tc.Mode
	(req.Conn.Srv.ops).(SrvReqOps).Create(req)
}

func (srv *Srv) createPost(req *SrvReq) {
	if req.Rc != nil && req.Rc.Type == Rcreate && req.Fid != nil {
		req.Fid.Type = req.Rc.Qid.Type
		req.Fid.opened = true
	}
}

func (srv *Srv) read(req *SrvReq) {
	tc := req.Tc
	fid := req.Fid
	if tc.Count+IOHDRSZ > req.Conn.Msize {
		req.RespondError(Etoolarge)
		return
	}

	if (fid.Type & QTAUTH) != 0 {
		var n int

		rc := req.Rc
		err := InitRread(rc, tc.Count)
		if err != nil {
			req.RespondError(err)
			return
		}

		if op, ok := (req.Conn.Srv.ops).(AuthOps); ok {
			n, err = op.AuthRead(fid, tc.Offset, rc.Data)
			if err != nil {
				req.RespondError(err)
				return
			}

			SetRreadCount(rc, uint32(n))
			req.Respond()
		} else {
			req.RespondError(Enotimpl)
		}

		return
	}

	if (fid.Type & QTDIR) != 0 {
		if tc.Offset == 0 {
			fid.Diroffset = 0
		} else if tc.Offset != fid.Diroffset {
			fid.Diroffset = tc.Offset
		}
	}

	(req.Conn.Srv.ops).(SrvReqOps).Read(req)
}

func (srv *Srv) readPost(req *SrvReq) {
	if req.Rc != nil && req.Rc.Type == Rread && (req.Fid.Type&QTDIR) != 0 {
		req.Fid.Diroffset += uint64(req.Rc.Count)
	}
}

func (srv *Srv) write(req *SrvReq) {
	fid := req.Fid
	tc := req.Tc
	if (fid.Type & QTAUTH) != 0 {
		tc := req.Tc
		if op, ok := (req.Conn.Srv.ops).(AuthOps); ok {
			n, err := op.AuthWrite(req.Fid, tc.Offset, tc.Data)
			if err != nil {
				req.RespondError(err)
			} else {
				req.RespondRwrite(uint32(n))
			}
		} else {
			req.RespondError(Enotimpl)
		}

		return
	}

	if !fid.opened || (fid.Type&QTDIR) != 0 || (fid.Omode&3) == OREAD {
		req.RespondError(Ebaduse)
		return
	}

	if tc.Count+IOHDRSZ > req.Conn.Msize {
		req.RespondError(Etoolarge)
		return
	}

	(req.Conn.Srv.ops).(SrvReqOps).Write(req)
}

func (srv *Srv) clunk(req *SrvReq) {
	fid := req.Fid
	if (fid.Type & QTAUTH) != 0 {
		if op, ok := (req.Conn.Srv.ops).(AuthOps); ok {
			op.AuthDestroy(fid)
			req.RespondRclunk()
		} else {
			req.RespondError(Enotimpl)
		}

		return
	}

	(req.Conn.Srv.ops).(SrvReqOps).Clunk(req)
}

func (srv *Srv) clunkPost(req *SrvReq) {
	if req.Rc != nil && req.Rc.Type == Rclunk && req.Fid != nil {
		req.Fid.DecRef()
	}
}

func (srv *Srv) remove(req *SrvReq) { (req.Conn.Srv.ops).(SrvReqOps).Remove(req) }

func (srv *Srv) removePost(req *SrvReq) {
	if req.Rc != nil && req.Fid != nil {
		req.Fid.DecRef()
	}
}

func (srv *Srv) stat(req *SrvReq) { (req.Conn.Srv.ops).(SrvReqOps).Stat(req) }

func (srv *Srv) wstat(req *SrvReq) {
	/*
		fid := req.Fid
		d := &req.Tc.Dir
		if d.Type != uint16(0xFFFF) || d.Dev != uint32(0xFFFFFFFF) || d.Version != uint32(0xFFFFFFFF) ||
			d.Path != uint64(0xFFFFFFFFFFFFFFFF) {
			req.RespondError(Eperm)
			return
		}

		if (d.Mode != 0xFFFFFFFF) && (((fid.Type&QTDIR) != 0 && (d.Mode&DMDIR) == 0) ||
			((d.Type&QTDIR) == 0 && (d.Mode&DMDIR) != 0)) {
			req.RespondError(Edirchange)
			return
		}
	*/

	(req.Conn.Srv.ops).(SrvReqOps).Wstat(req)
}
