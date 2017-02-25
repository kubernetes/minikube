// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srv

import "fmt"
import "k8s.io/minikube/third_party/go9p/p"

// Respond to the request with Rerror message
func (req *Req) RespondError(err interface{}) {
	switch e := err.(type) {
	case *p.Error:
		p.PackRerror(req.Rc, e.Error(), uint32(e.Errornum), req.Conn.Dotu)
	case error:
		p.PackRerror(req.Rc, e.Error(), uint32(p.EIO), req.Conn.Dotu)
	default:
		p.PackRerror(req.Rc, fmt.Sprintf("%v", e), uint32(p.EIO), req.Conn.Dotu)
	}

	req.Respond()
}

// Respond to the request with Rversion message
func (req *Req) RespondRversion(msize uint32, version string) {
	err := p.PackRversion(req.Rc, msize, version)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rauth message
func (req *Req) RespondRauth(aqid *p.Qid) {
	err := p.PackRauth(req.Rc, aqid)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rflush message
func (req *Req) RespondRflush() {
	err := p.PackRflush(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rattach message
func (req *Req) RespondRattach(aqid *p.Qid) {
	err := p.PackRattach(req.Rc, aqid)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rwalk message
func (req *Req) RespondRwalk(wqids []p.Qid) {
	err := p.PackRwalk(req.Rc, wqids)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Ropen message
func (req *Req) RespondRopen(qid *p.Qid, iounit uint32) {
	err := p.PackRopen(req.Rc, qid, iounit)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rcreate message
func (req *Req) RespondRcreate(qid *p.Qid, iounit uint32) {
	err := p.PackRcreate(req.Rc, qid, iounit)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rread message
func (req *Req) RespondRread(data []byte) {
	err := p.PackRread(req.Rc, data)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rwrite message
func (req *Req) RespondRwrite(count uint32) {
	err := p.PackRwrite(req.Rc, count)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rclunk message
func (req *Req) RespondRclunk() {
	err := p.PackRclunk(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rremove message
func (req *Req) RespondRremove() {
	err := p.PackRremove(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rstat message
func (req *Req) RespondRstat(st *p.Dir) {
	err := p.PackRstat(req.Rc, st, req.Conn.Dotu)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rwstat message
func (req *Req) RespondRwstat() {
	err := p.PackRwstat(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}
