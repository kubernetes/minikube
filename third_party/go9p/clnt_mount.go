// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import (
	"net"
)

// Creates an authentication fid for the specified user. Returns the fid, if
// successful, or an Error.
func (clnt *Clnt) Auth(user User, aname string) (*Fid, error) {
	fid := clnt.FidAlloc()
	tc := clnt.NewFcall()
	err := PackTauth(tc, fid.Fid, user.Name(), aname, uint32(user.Id()), clnt.Dotu)
	if err != nil {
		return nil, err
	}

	_, err = clnt.Rpc(tc)
	if err != nil {
		return nil, err
	}

	fid.User = user
	fid.walked = true
	return fid, nil
}

// Creates a fid for the specified user that points to the root
// of the file server's file tree. Returns a Fid pointing to the root,
// if successful, or an Error.
func (clnt *Clnt) Attach(afid *Fid, user User, aname string) (*Fid, error) {
	var afno uint32

	if afid != nil {
		afno = afid.Fid
	} else {
		afno = NOFID
	}

	fid := clnt.FidAlloc()
	tc := clnt.NewFcall()
	err := PackTattach(tc, fid.Fid, afno, user.Name(), aname, uint32(user.Id()), clnt.Dotu)
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
func Mount(ntype, addr, aname string, msize uint32, user User) (*Clnt, error) {
	c, e := net.Dial(ntype, addr)
	if e != nil {
		return nil, &Error{e.Error(), EIO}
	}

	return MountConn(c, aname, msize, user)
}

func MountConn(c net.Conn, aname string, msize uint32, user User) (*Clnt, error) {
	clnt, err := Connect(c, msize+IOHDRSZ, true)
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

// Closes the connection to the file server.
func (clnt *Clnt) Unmount() {
	clnt.Lock()
	clnt.err = &Error{"connection closed", EIO}
	clnt.conn.Close()
	clnt.Unlock()
}
