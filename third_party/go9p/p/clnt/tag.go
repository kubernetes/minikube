// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clnt

import "k8s.io/minikube/third_party/go9p/p"

type Tag struct {
	clnt     *Clnt
	tag      uint16
	reqchan  chan *Req
	respchan chan *Req
	donechan chan bool
}

func (clnt *Clnt) TagAlloc(reqchan chan *Req) *Tag {
	tag := new(Tag)
	tag.clnt = clnt
	tag.tag = uint16(clnt.tagpool.getId())
	tag.reqchan = reqchan
	tag.respchan = make(chan *Req, 16)
	tag.donechan = make(chan bool)
	go tag.reqproc()

	return tag
}

func (clnt *Clnt) TagFree(tag *Tag) {
	tag.donechan <- true
	clnt.tagpool.putId(uint32(tag.tag))
}

func (tag *Tag) reqAlloc() *Req {
	r := new(Req)
	r.tag = tag.tag
	r.Clnt = tag.clnt
	r.Done = tag.respchan
	r.Tc = tag.clnt.NewFcall()

	return r
}

func (tag *Tag) ReqFree(r *Req) {
	tag.clnt.FreeFcall(r.Tc)
}

func (tag *Tag) reqproc() {
	for {
		select {
		case <-tag.donechan:
			return

		case r := <-tag.respchan:
			rc := r.Rc
			fid := r.fid
			err := r.Rc.Type == p.Rerror

			switch r.Tc.Type {
			case p.Tauth:
				if err {
					fid.User = nil
				}

			case p.Tattach:
				if !err {
					fid.Qid = rc.Qid
				} else {
					fid.User = nil
				}

			case p.Twalk:
				if !err {
					fid.walked = true
					if len(rc.Wqid) > 0 {
						fid.Qid = rc.Wqid[len(rc.Wqid)-1]
					}
				} else {
					fid.User = nil
				}

			case p.Topen:
			case p.Tcreate:
				if !err {
					fid.Iounit = rc.Iounit
					fid.Qid = rc.Qid
				} else {
					fid.Mode = 0
				}

			case p.Tclunk:
			case p.Tremove:
				tag.clnt.fidpool.putId(fid.Fid)
			}

			tag.reqchan <- r
		}
	}
}

func (tag *Tag) Auth(afid *Fid, user p.User, aname string) error {
	req := tag.reqAlloc()
	req.fid = afid
	err := p.PackTauth(req.Tc, afid.Fid, user.Name(), aname, uint32(user.Id()), tag.clnt.Dotu)
	if err != nil {
		return err
	}

	afid.User = user
	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Attach(fid, afid *Fid, user p.User, aname string) error {
	var afno uint32

	if afid != nil {
		afno = afid.Fid
	} else {
		afno = p.NOFID
	}

	req := tag.reqAlloc()
	req.fid = fid
	err := p.PackTattach(req.Tc, fid.Fid, afno, user.Name(), aname, uint32(user.Id()), tag.clnt.Dotu)
	if err != nil {
		return err
	}

	fid.User = user
	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Walk(fid *Fid, newfid *Fid, wnames []string) error {
	req := tag.reqAlloc()
	req.fid = newfid
	if len(wnames) == 0 {
		newfid.Qid = fid.Qid
	}

	err := p.PackTwalk(req.Tc, fid.Fid, newfid.Fid, wnames)
	if err != nil {
		return err
	}

	newfid.User = fid.User
	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Open(fid *Fid, mode uint8) error {
	req := tag.reqAlloc()
	req.fid = fid
	err := p.PackTopen(req.Tc, fid.Fid, mode)
	if err != nil {
		return err
	}

	fid.Mode = mode
	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Create(fid *Fid, name string, perm uint32, mode uint8, ext string) error {
	req := tag.reqAlloc()
	req.fid = fid
	err := p.PackTcreate(req.Tc, fid.Fid, name, perm, mode, ext, tag.clnt.Dotu)
	if err != nil {
		return err
	}

	fid.Mode = mode
	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Read(fid *Fid, offset uint64, count uint32) error {
	req := tag.reqAlloc()
	req.fid = fid
	err := p.PackTread(req.Tc, fid.Fid, offset, count)
	if err != nil {
		return err
	}

	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Write(fid *Fid, data []byte, offset uint64) error {
	req := tag.reqAlloc()
	req.fid = fid
	err := p.PackTwrite(req.Tc, fid.Fid, offset, uint32(len(data)), data)
	if err != nil {
		return err
	}

	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Clunk(fid *Fid) error {
	req := tag.reqAlloc()
	req.fid = fid
	err := p.PackTclunk(req.Tc, fid.Fid)
	if err != nil {
		return err
	}

	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Remove(fid *Fid) error {
	req := tag.reqAlloc()
	req.fid = fid
	err := p.PackTremove(req.Tc, fid.Fid)
	if err != nil {
		return err
	}

	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Stat(fid *Fid) error {
	req := tag.reqAlloc()
	req.fid = fid
	err := p.PackTstat(req.Tc, fid.Fid)
	if err != nil {
		return err
	}

	return tag.clnt.Rpcnb(req)
}

func (tag *Tag) Wstat(fid *Fid, dir *p.Dir) error {
	req := tag.reqAlloc()
	req.fid = fid
	err := p.PackTwstat(req.Tc, fid.Fid, dir, tag.clnt.Dotu)
	if err != nil {
		return err
	}

	return tag.clnt.Rpcnb(req)
}
