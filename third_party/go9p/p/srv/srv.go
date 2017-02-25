// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The srv package provides definitions and functions used to implement
// a 9P2000 file server.
package srv

import (
	"k8s.io/minikube/third_party/go9p/p"
	"net"
	"runtime"
	"sync"
)

type reqStatus int

const (
	reqFlush     reqStatus = (1 << iota) /* request is flushed (no response will be sent) */
	reqWork                              /* goroutine is currently working on it */
	reqResponded                         /* response is already produced */
	reqSaved                             /* no response was produced after the request is worked on */
)

// Debug flags
const (
	DbgPrintFcalls  = (1 << iota) // print all 9P messages on stderr
	DbgPrintPackets               // print the raw packets on stderr
	DbgLogFcalls                  // keep the last N 9P messages (can be accessed over http)
	DbgLogPackets                 // keep the last N 9P messages (can be accessed over http)
)

var Eunknownfid error = &p.Error{"unknown fid", p.EINVAL}
var Enoauth error = &p.Error{"no authentication required", p.EINVAL}
var Einuse error = &p.Error{"fid already in use", p.EINVAL}
var Ebaduse error = &p.Error{"bad use of fid", p.EINVAL}
var Eopen error = &p.Error{"fid already opened", p.EINVAL}
var Enotdir error = &p.Error{"not a directory", p.ENOTDIR}
var Eperm error = &p.Error{"permission denied", p.EPERM}
var Etoolarge error = &p.Error{"i/o count too large", p.EINVAL}
var Ebadoffset error = &p.Error{"bad offset in directory read", p.EINVAL}
var Edirchange error = &p.Error{"cannot convert between files and directories", p.EINVAL}
var Enouser error = &p.Error{"unknown user", p.EINVAL}
var Enotimpl error = &p.Error{"not implemented", p.EINVAL}

// Authentication operations. The file server should implement them if
// it requires user authentication. The authentication in 9P2000 is
// done by creating special authentication fids and performing I/O
// operations on them. Once the authentication is done, the authentication
// fid can be used by the user to get access to the actual files.
type AuthOps interface {
	// AuthInit is called when the user starts the authentication
	// process on Fid afid. The user that is being authenticated
	// is referred by afid.User. The function should return the Qid
	// for the authentication file, or an Error if the user can't be
	// authenticated
	AuthInit(afid *Fid, aname string) (*p.Qid, error)

	// AuthDestroy is called when an authentication fid is destroyed.
	AuthDestroy(afid *Fid)

	// AuthCheck is called after the authentication process is finished
	// when the user tries to attach to the file server. If the function
	// returns nil, the authentication was successful and the user has
	// permission to access the files.
	AuthCheck(fid *Fid, afid *Fid, aname string) error

	// AuthRead is called when the user attempts to read data from an
	// authentication fid.
	AuthRead(afid *Fid, offset uint64, data []byte) (count int, err error)

	// AuthWrite is called when the user attempts to write data to an
	// authentication fid.
	AuthWrite(afid *Fid, offset uint64, data []byte) (count int, err error)
}

// Connection operations. These should be implemented if the file server
// needs to be called when a connection is opened or closed.
type ConnOps interface {
	ConnOpened(*Conn)
	ConnClosed(*Conn)
}

// Fid operations. This interface should be implemented if the file server
// needs to be called when a Fid is destroyed.
type FidOps interface {
	FidDestroy(*Fid)
}

// Request operations. This interface should be implemented if the file server
// needs to bypass the default request process, or needs to perform certain
// operations before the (any) request is processed, or before (any) response
// sent back to the client.
type ReqProcessOps interface {
	// Called when a new request is received from the client. If the
	// interface is not implemented, (req *Req) srv.Process() method is
	// called. If the interface is implemented, it is the user's
	// responsibility to call srv.Process. If srv.Process isn't called,
	// Fid, Afid and Newfid fields in Req are not set, and the ReqOps
	// methods are not called.
	ReqProcess(*Req)

	// Called when a request is responded, i.e. when (req *Req)srv.Respond()
	// is called and before the response is sent. If the interface is not
	// implemented, (req *Req) srv.PostProcess() method is called to finalize
	// the request. If the interface is implemented and ReqProcess calls
	// the srv.Process method, ReqRespond should call the srv.PostProcess
	// method.
	ReqRespond(*Req)
}

// Flush operation. This interface should be implemented if the file server
// can flush pending requests. If the interface is not implemented, requests
// that were passed to the file server implementation won't be flushed.
// The flush method should call the (req *Req) srv.Flush() method if the flush
// was successful so the request can be marked appropriately.
type FlushOp interface {
	Flush(*Req)
}

// Request operations. This interface should be implemented by all file servers.
// The operations correspond directly to most of the 9P2000 message types.
type ReqOps interface {
	Attach(*Req)
	Walk(*Req)
	Open(*Req)
	Create(*Req)
	Read(*Req)
	Write(*Req)
	Clunk(*Req)
	Remove(*Req)
	Stat(*Req)
	Wstat(*Req)
}

type StatsOps interface {
	statsRegister()
	statsUnregister()
}

// The Srv type contains the basic fields used to control the 9P2000
// file server. Each file server implementation should create a value
// of Srv type, initialize the values it cares about and pass the
// struct to the (Srv *) srv.Start(ops) method together with the object
// that implements the file server operations.
type Srv struct {
	sync.Mutex
	Id         string  // Used for debugging and stats
	Msize      uint32  // Maximum size of the 9P2000 messages supported by the server
	Dotu       bool    // If true, the server supports the 9P2000.u extension
	Debuglevel int     // debug level
	Upool      p.Users // Interface for finding users and groups known to the file server
	Maxpend    int     // Maximum pending outgoing requests
	Log        *p.Logger

	ops   interface{}     // operations
	conns map[*Conn]*Conn // List of connections
}

// The Conn type represents a connection from a client to the file server
type Conn struct {
	sync.Mutex
	Srv        *Srv
	Msize      uint32 // maximum size of 9P2000 messages for the connection
	Dotu       bool   // if true, both the client and the server speak 9P2000.u
	Id         string // used for debugging and stats
	Debuglevel int

	conn    net.Conn
	fidpool map[uint32]*Fid
	reqs    map[uint16]*Req // all outstanding requests

	reqout chan *Req
	rchan  chan *p.Fcall
	done   chan bool

	// stats
	nreqs   int    // number of requests processed by the server
	tsz     uint64 // total size of the T messages received
	rsz     uint64 // total size of the R messages sent
	npend   int    // number of currently pending messages
	maxpend int    // maximum number of pending messages
	nreads  int    // number of reads
	nwrites int    // number of writes
}

// The Fid type identifies a file on the file server.
// A new Fid is created when the user attaches to the file server (the Attach
// operation), or when Walk-ing to a file. The Fid values are created
// automatically by the srv implementation. The FidDestroy operation is called
// when a Fid is destroyed.
type Fid struct {
	sync.Mutex
	fid       uint32
	refcount  int
	opened    bool        // True if the Fid is opened
	Fconn     *Conn       // Connection the Fid belongs to
	Omode     uint8       // Open mode (p.O* flags), if the fid is opened
	Type      uint8       // Fid type (p.QT* flags)
	Diroffset uint64      // If directory, the next valid read position
	User      p.User      // The Fid's user
	Aux       interface{} // Can be used by the file server implementation for per-Fid data
}

// The Req type represents a 9P2000 request. Each request has a
// T-message (Tc) and a R-message (Rc). If the ReqProcessOps don't
// override the default behavior, the implementation initializes Fid,
// Afid and Newfid values and automatically keeps track on when the Fids
// should be destroyed.
type Req struct {
	sync.Mutex
	Tc     *p.Fcall // Incoming 9P2000 message
	Rc     *p.Fcall // Outgoing 9P2000 response
	Fid    *Fid     // The Fid value for all messages that contain fid[4]
	Afid   *Fid     // The Fid value for the messages that contain afid[4] (Tauth and Tattach)
	Newfid *Fid     // The Fid value for the messages that contain newfid[4] (Twalk)
	Conn   *Conn    // Connection that the request belongs to

	status     reqStatus
	flushreq   *Req
	prev, next *Req
}

// The Start method should be called once the file server implementor
// initializes the Srv struct with the preferred values. It sets default
// values to the fields that are not initialized and creates the goroutines
// required for the server's operation. The method receives an empty
// interface value, ops, that should implement the interfaces the file server is
// interested in. Ops must implement the ReqOps interface.
func (srv *Srv) Start(ops interface{}) bool {
	if _, ok := (ops).(ReqOps); !ok {
		return false
	}

	srv.ops = ops
	if srv.Upool == nil {
		srv.Upool = p.OsUsers
	}

	if srv.Msize < p.IOHDRSZ {
		srv.Msize = p.MSIZE
	}

	if srv.Log == nil {
		srv.Log = p.NewLogger(1024)
	}

	if sop, ok := (interface{}(srv)).(StatsOps); ok {
		sop.statsRegister()
	}

	return true
}

func (srv *Srv) String() string {
	return srv.Id
}

func (req *Req) process() {
	req.Lock()
	flushed := (req.status & reqFlush) != 0
	if !flushed {
		req.status |= reqWork
	}
	req.Unlock()

	if flushed {
		req.Respond()
	}

	if rop, ok := (req.Conn.Srv.ops).(ReqProcessOps); ok {
		rop.ReqProcess(req)
	} else {
		req.Process()
	}

	req.Lock()
	req.status &= ^reqWork
	if !(req.status&reqResponded != 0) {
		req.status |= reqSaved
	}
	req.Unlock()
}

// Performs the default processing of a request. Initializes
// the Fid, Afid and Newfid fields and calls the appropriate
// ReqOps operation for the message. The file server implementer
// should call it only if the file server implements the ReqProcessOps
// within the ReqProcess operation.
func (req *Req) Process() {
	conn := req.Conn
	srv := conn.Srv
	tc := req.Tc

	if tc.Fid != p.NOFID && tc.Type != p.Tattach {
		srv.Lock()
		req.Fid = conn.FidGet(tc.Fid)
		srv.Unlock()
		if req.Fid == nil {
			req.RespondError(Eunknownfid)
			return
		}
	}

	switch req.Tc.Type {
	default:
		req.RespondError(&p.Error{"unknown message type", p.EINVAL})

	case p.Tversion:
		srv.version(req)

	case p.Tauth:
		if runtime.GOOS == "windows" {
			return
		}
		srv.auth(req)

	case p.Tattach:
		srv.attach(req)

	case p.Tflush:
		srv.flush(req)

	case p.Twalk:
		srv.walk(req)

	case p.Topen:
		srv.open(req)

	case p.Tcreate:
		srv.create(req)

	case p.Tread:
		srv.read(req)

	case p.Twrite:
		srv.write(req)

	case p.Tclunk:
		srv.clunk(req)

	case p.Tremove:
		srv.remove(req)

	case p.Tstat:
		srv.stat(req)

	case p.Twstat:
		srv.wstat(req)
	}
}

// Performs the post processing required if the (*Req) Process() method
// is called for a request. The file server implementer should call it
// only if the file server implements the ReqProcessOps within the
// ReqRespond operation.
func (req *Req) PostProcess() {
	srv := req.Conn.Srv

	/* call the post-handlers (if needed) */
	switch req.Tc.Type {
	case p.Tauth:
		srv.authPost(req)

	case p.Tattach:
		srv.attachPost(req)

	case p.Twalk:
		srv.walkPost(req)

	case p.Topen:
		srv.openPost(req)

	case p.Tcreate:
		srv.createPost(req)

	case p.Tread:
		srv.readPost(req)

	case p.Tclunk:
		srv.clunkPost(req)

	case p.Tremove:
		srv.removePost(req)
	}

	if req.Fid != nil {
		req.Fid.DecRef()
		req.Fid = nil
	}

	if req.Afid != nil {
		req.Afid.DecRef()
		req.Afid = nil
	}

	if req.Newfid != nil {
		req.Newfid.DecRef()
		req.Newfid = nil
	}
}

// The Respond method sends response back to the client. The req.Rc value
// should be initialized and contain valid 9P2000 message. In most cases
// the file server implementer shouldn't call this method directly. Instead
// one of the RespondR* methods should be used.
func (req *Req) Respond() {
	var flushreqs *Req

	conn := req.Conn
	req.Lock()
	status := req.status
	req.status |= reqResponded
	req.status &= ^reqWork
	req.Unlock()

	if (status & reqResponded) != 0 {
		return
	}

	/* remove the request and all requests flushing it */
	conn.Lock()
	nextreq := req.prev
	if nextreq != nil {
		nextreq.next = nil
		// if there are flush requests, move them to the next request
		if req.flushreq != nil {
			var p *Req = nil
			r := nextreq.flushreq
			for ; r != nil; p, r = r, r.flushreq {
			}

			if p == nil {
				nextreq.flushreq = req.flushreq
			} else {
				p.next = req.flushreq
			}
		}

		flushreqs = nil
	} else {
		delete(conn.reqs, req.Tc.Tag)
		flushreqs = req.flushreq
	}
	conn.Unlock()

	if rop, ok := (req.Conn.Srv.ops).(ReqProcessOps); ok {
		rop.ReqRespond(req)
	} else {
		req.PostProcess()
	}

	if (status & reqFlush) == 0 {
		conn.reqout <- req
	}

	// process the next request with the same tag (if available)
	if nextreq != nil {
		go nextreq.process()
	}

	// respond to the flush messages
	// can't send the responses directly to conn.reqout, because the
	// the flushes may be in a tag group too
	for freq := flushreqs; freq != nil; freq = freq.flushreq {
		freq.Respond()
	}
}

// Should be called to cancel a request. Should only be called
// from the Flush operation if the FlushOp is implemented.
func (req *Req) Flush() {
	req.Lock()
	req.status |= reqFlush
	req.Unlock()
	req.Respond()
}

// Lookup a Fid struct based on the 32-bit identifier sent over the wire.
// Returns nil if the fid is not found. Increases the reference count of
// the returned fid. The user is responsible to call DecRef once it no
// longer needs it.
func (conn *Conn) FidGet(fidno uint32) *Fid {
	conn.Lock()
	fid, present := conn.fidpool[fidno]
	conn.Unlock()
	if present {
		fid.IncRef()
	}

	return fid
}

// Creates a new Fid struct for the fidno integer. Returns nil
// if the Fid for that number already exists. The returned fid
// has reference count set to 1.
func (conn *Conn) FidNew(fidno uint32) *Fid {
	conn.Lock()
	_, present := conn.fidpool[fidno]
	if present {
		conn.Unlock()
		return nil
	}

	fid := new(Fid)
	fid.fid = fidno
	fid.refcount = 1
	fid.Fconn = conn
	conn.fidpool[fidno] = fid
	conn.Unlock()

	return fid
}

func (conn *Conn) String() string {
	return conn.Srv.Id + "/" + conn.Id
}

// Increase the reference count for the fid.
func (fid *Fid) IncRef() {
	fid.Lock()
	fid.refcount++
	fid.Unlock()
}

// Decrease the reference count for the fid. When the
// reference count reaches 0, the fid is no longer valid.
func (fid *Fid) DecRef() {
	fid.Lock()
	fid.refcount--
	n := fid.refcount
	fid.Unlock()

	if n > 0 {
		return
	}

	conn := fid.Fconn
	conn.Lock()
	delete(conn.fidpool, fid.fid)
	conn.Unlock()

	if fop, ok := (conn.Srv.ops).(FidOps); ok {
		fop.FidDestroy(fid)
	}
}
