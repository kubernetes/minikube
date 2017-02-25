// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clnt

import (
	"k8s.io/minikube/third_party/go9p/p"
	"net"
)

// Creates an authentication fid for the specified user. Returns the fid, if
// successful, or an Error.
func (clnt *Clnt) Auth(user p.User, aname string) (*Fid, error) {
	fid := clnt.FidAlloc()
	tc := clnt.NewFcall()
	err := p.PackTauth(tc, fid.Fid, user.Name(), aname, uint32(user.Id()), clnt.Dotu)
	if err != nil {
		return nil, err
	}

	_, err = clnt.Rpc(tc)
	if err != nil {
		return nil, err
	}

	fid.Iounit = clnt.Msize - p.IOHDRSZ
	fid.User = user
	fid.walked = true
	return fid, nil
}

// Creates a fid for the specified user that points to the root
// of the file server's file tree. Returns a Fid pointing to the root,
// if successful, or an Error.
func (clnt *Clnt) Attach(afid *Fid, user p.User, aname string) (*Fid, error) {
	var afno uint32

	if afid != nil {
		afno = afid.Fid
	} else {
		afno = p.NOFID
	}

	fid := clnt.FidAlloc()
	tc := clnt.NewFcall()
	err := p.PackTattach(tc, fid.Fid, afno, user.Name(), aname, uint32(user.Id()), clnt.Dotu)
	if err != nil {
		return nil, err
	}

	rc, err := clnt.Rpc(tc)
	if err != nil {
		return nil, err
	}

	fid.Qid = rc.Qid
	fid.User = user
	fid.walked = true
	return fid, nil
}

// Connects to a file server and attaches to it as the specified user.
func Mount(ntype, addr, aname string, user p.User) (*Clnt, error) {
	c, e := net.Dial(ntype, addr)
	if e != nil {
		return nil, &p.Error{e.Error(), p.EIO}
	}

	return MountConn(c, aname, user)
}

func MountConn(c net.Conn, aname string, user p.User) (*Clnt, error) {
	clnt, err := Connect(c, 8192+p.IOHDRSZ, true)
	if err != nil {
		return nil, err
	}

	fid, err := clnt.Attach(nil, user, aname)
	if err != nil {
		clnt.Unmount()
		return nil, err
	}

	clnt.Root = fid
	return clnt, nil
}

// Closes the connection to the file sever.
func (clnt *Clnt) Unmount() {
	clnt.Lock()
	clnt.err = &p.Error{"connection closed", p.EIO}
	clnt.conn.Close()
	clnt.Unlock()
}
