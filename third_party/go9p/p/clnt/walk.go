// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clnt

import (
	"k8s.io/minikube/third_party/go9p/p"
	"strings"
)

// Starting from the file associated with fid, walks all wnames in
// sequence and associates the resulting file with newfid. If no wnames
// were walked successfully, an Error is returned. Otherwise a slice with a
// Qid for each walked name is returned.
func (clnt *Clnt) Walk(fid *Fid, newfid *Fid, wnames []string) ([]p.Qid, error) {
	tc := clnt.NewFcall()
	err := p.PackTwalk(tc, fid.Fid, newfid.Fid, wnames)
	if err != nil {
		return nil, err
	}

	rc, err := clnt.Rpc(tc)
	if err != nil {
		return nil, err
	}

	newfid.walked = true
	return rc.Wqid, nil
}

// Walks to a named file. Returns a Fid associated with the file,
// or an Error.
func (clnt *Clnt) FWalk(path string) (*Fid, error) {
	var err error = nil

	var i, m int
	for i = 0; i < len(path); i++ {
		if path[i] != '/' {
			break
		}
	}

	if i > 0 {
		path = path[i:]
	}

	wnames := strings.Split(path, "/")
	newfid := clnt.FidAlloc()
	fid := clnt.Root
	newfid.User = fid.User

	/* get rid of the empty names */
	for i, m = 0, 0; i < len(wnames); i++ {
		if wnames[i] != "" {
			wnames[m] = wnames[i]
			m++
		}
	}

	wnames = wnames[0:m]
	for {
		n := len(wnames)
		if n > 16 {
			n = 16
		}

		tc := clnt.NewFcall()
		err = p.PackTwalk(tc, fid.Fid, newfid.Fid, wnames[0:n])
		if err != nil {
			goto error
		}

		var rc *p.Fcall
		rc, err = clnt.Rpc(tc)
		if err != nil {
			goto error
		}

		newfid.walked = true
		if len(rc.Wqid) != n {
			err = &p.Error{"file not found", p.ENOENT}
			goto error
		}

		if len(rc.Wqid) > 0 {
			newfid.Qid = rc.Wqid[len(rc.Wqid)-1]
		} else {
			newfid.Qid = fid.Qid
		}

		wnames = wnames[n:]
		fid = newfid
		if len(wnames) == 0 {
			break
		}
	}

	return newfid, nil

error:
	clnt.Clunk(newfid)
	return nil, err
}
