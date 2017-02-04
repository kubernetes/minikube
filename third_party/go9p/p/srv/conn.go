// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package srv

import (
	"fmt"
	"k8s.io/minikube/third_party/go9p/p"
	"log"
	"net"
)

func (srv *Srv) NewConn(c net.Conn) {
	conn := new(Conn)
	conn.Srv = srv
	conn.Msize = srv.Msize
	conn.Dotu = srv.Dotu
	conn.Debuglevel = srv.Debuglevel
	conn.conn = c
	conn.fidpool = make(map[uint32]*Fid)
	conn.reqs = make(map[uint16]*Req)
	conn.reqout = make(chan *Req, srv.Maxpend)
	conn.done = make(chan bool)
	conn.rchan = make(chan *p.Fcall, 64)

	srv.Lock()
	if srv.conns == nil {
		srv.conns = make(map[*Conn]*Conn)
	}
	srv.conns[conn] = conn
	srv.Unlock()

	conn.Id = c.RemoteAddr().String()
	if op, ok := (conn.Srv.ops).(ConnOps); ok {
		op.ConnOpened(conn)
	}

	if sop, ok := (interface{}(conn)).(StatsOps); ok {
		sop.statsRegister()
	}

	go conn.recv()
	go conn.send()
}

func (conn *Conn) close() {
	conn.done <- true
	conn.Srv.Lock()
	delete(conn.Srv.conns, conn)
	conn.Srv.Unlock()

	if sop, ok := (interface{}(conn)).(StatsOps); ok {
		sop.statsUnregister()
	}
	if op, ok := (conn.Srv.ops).(ConnOps); ok {
		op.ConnClosed(conn)
	}

	/* call FidDestroy for all remaining fids */
	if op, ok := (conn.Srv.ops).(FidOps); ok {
		for _, fid := range conn.fidpool {
			op.FidDestroy(fid)
		}
	}
}

func (conn *Conn) recv() {
	var err error
	var n int

	buf := make([]byte, conn.Msize*8)
	pos := 0
	for {
		if len(buf) < int(conn.Msize) {
			b := make([]byte, conn.Msize*8)
			copy(b, buf[0:pos])
			buf = b
			b = nil
		}

		n, err = conn.conn.Read(buf[pos:])
		if err != nil || n == 0 {
			conn.close()
			return
		}

		pos += n
		for pos > 4 {
			sz, _ := p.Gint32(buf)
			if sz > conn.Msize {
				log.Println("bad client connection: ", conn.conn.RemoteAddr())
				conn.conn.Close()
				conn.close()
				return
			}
			if pos < int(sz) {
				if len(buf) < int(sz) {
					b := make([]byte, conn.Msize*8)
					copy(b, buf[0:pos])
					buf = b
					b = nil
				}

				break
			}
			fc, err, fcsize := p.Unpack(buf, conn.Dotu)
			if err != nil {
				log.Println(fmt.Sprintf("invalid packet : %v %v", err, buf))
				conn.conn.Close()
				conn.close()
				return
			}

			tag := fc.Tag
			req := new(Req)
			select {
			case req.Rc = <-conn.rchan:
				break
			default:
				req.Rc = p.NewFcall(conn.Msize)
			}

			req.Conn = conn
			req.Tc = fc
			//			req.Rc = rc
			if conn.Debuglevel > 0 {
				conn.logFcall(req.Tc)
				if conn.Debuglevel&DbgPrintPackets != 0 {
					log.Println(">->", conn.Id, fmt.Sprint(req.Tc.Pkt))
				}

				if conn.Debuglevel&DbgPrintFcalls != 0 {
					log.Println(">>>", conn.Id, req.Tc.String())
				}
			}

			conn.Lock()
			conn.nreqs++
			conn.tsz += uint64(fc.Size)
			conn.npend++
			if conn.npend > conn.maxpend {
				conn.maxpend = conn.npend
			}

			req.next = conn.reqs[tag]
			conn.reqs[tag] = req
			process := req.next == nil
			if req.next != nil {
				req.next.prev = req
			}
			conn.Unlock()
			if process {
				go req.process()
			}

			buf = buf[fcsize:]
			pos -= fcsize
		}
	}

	panic("unreached")
}

func (conn *Conn) send() {
	for {
		select {
		case <-conn.done:
			return

		case req := <-conn.reqout:
			p.SetTag(req.Rc, req.Tc.Tag)
			conn.Lock()
			conn.rsz += uint64(req.Rc.Size)
			conn.npend--
			conn.Unlock()
			if conn.Debuglevel > 0 {
				conn.logFcall(req.Rc)
				if conn.Debuglevel&DbgPrintPackets != 0 {
					log.Println("<-<", conn.Id, fmt.Sprint(req.Rc.Pkt))
				}

				if conn.Debuglevel&DbgPrintFcalls != 0 {
					log.Println("<<<", conn.Id, req.Rc.String())
				}
			}

			for buf := req.Rc.Pkt; len(buf) > 0; {
				n, err := conn.conn.Write(buf)
				if err != nil {
					/* just close the socket, will get signal on conn.done */
					log.Println("error while writing")
					conn.conn.Close()
					break
				}

				buf = buf[n:]
			}

			select {
			case conn.rchan <- req.Rc:
				break
			default:
			}
		}
	}

	panic("unreached")
}

func (conn *Conn) RemoteAddr() net.Addr {
	return conn.conn.RemoteAddr()
}

func (conn *Conn) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}

func (conn *Conn) logFcall(fc *p.Fcall) {
	if conn.Debuglevel&DbgLogPackets != 0 {
		pkt := make([]byte, len(fc.Pkt))
		copy(pkt, fc.Pkt)
		conn.Srv.Log.Log(pkt, conn, DbgLogPackets)
	}

	if conn.Debuglevel&DbgLogFcalls != 0 {
		f := new(p.Fcall)
		*f = *fc
		f.Pkt = nil
		conn.Srv.Log.Log(f, conn, DbgLogFcalls)
	}
}

func (srv *Srv) StartNetListener(ntype, addr string) error {
	l, err := net.Listen(ntype, addr)
	if err != nil {
		return &p.Error{err.Error(), p.EIO}
	}

	return srv.StartListener(l)
}

// Start listening on the specified network and address for incoming
// connections. Once a connection is established, create a new Conn
// value, read messages from the socket, send them to the specified
// server, and send back responses received from the server.
func (srv *Srv) StartListener(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return &p.Error{err.Error(), p.EIO}
		}

		srv.NewConn(c)
	}
	return nil
}
