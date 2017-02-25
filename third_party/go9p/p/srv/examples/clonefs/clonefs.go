// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// A synthetic filesystem emulating a persistent cloning interface
// from Plan 9. Reading the /clone file creates new entries in the filesystem
// each containing unique information/data. Clone files remember what it written
// to them. Removing a clone file does what is expected.

package main

import (
	"flag"
	"fmt"
	"k8s.io/minikube/third_party/go9p/p"
	"k8s.io/minikube/third_party/go9p/p/srv"
	"log"
	"os"
	"strconv"
	"time"
)

type ClFile struct {
	srv.File
	created string
	id      int
	data    []byte
}

type Clone struct {
	srv.File
	clones int
}

var addr = flag.String("addr", ":5640", "network address")
var debug = flag.Bool("d", false, "print debug messages")

var root *srv.File

func (cl *ClFile) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	var b []byte
	if len(cl.data) == 0 {
		str := strconv.Itoa(cl.id) + " created on:" + cl.created
		b = []byte(str)
	} else {
		b = cl.data
	}
	n := len(b)
	if offset >= uint64(n) {
		return 0, nil
	}

	b = b[int(offset):n]
	n -= int(offset)
	if len(buf) < n {
		n = len(buf)
	}

	copy(buf[offset:int(offset)+n], b[offset:])
	return n, nil
}

func (cl *ClFile) Write(fid *srv.FFid, data []byte, offset uint64) (int, error) {
	n := uint64(len(cl.data))
	nlen := offset + uint64(len(data))
	if nlen > n {
		ndata := make([]byte, nlen)
		copy(ndata, cl.data[0:n])
		cl.data = ndata
	}

	copy(cl.data[offset:], data[offset:])
	return len(data), nil
}

func (cl *ClFile) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil
}

func (cl *ClFile) Remove(fid *srv.FFid) error {
	return nil
}

func (cl *Clone) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	// we only allow a single read from us, change the offset and we're done
	if offset > uint64(0) {
		return 0, nil
	}

	cl.clones += 1
	ncl := new(ClFile)
	ncl.id = cl.clones
	ncl.created = time.Now().String()
	name := strconv.Itoa(ncl.id)

	err := ncl.Add(root, name, p.OsUsers.Uid2User(os.Geteuid()), nil, 0666, ncl)
	if err != nil {
		return 0, &p.Error{"can not create file", 0}
	}

	b := []byte(name)
	if len(buf) < len(b) {
		// cleanup
		ncl.File.Remove()
		return 0, &p.Error{"not enough buffer space for result", 0}
	}

	copy(buf, b)
	return len(b), nil
}

func main() {
	var err error
	var cl *Clone
	var s *srv.Fsrv

	flag.Parse()
	user := p.OsUsers.Uid2User(os.Geteuid())
	root = new(srv.File)
	err = root.Add(nil, "/", user, nil, p.DMDIR|0777, nil)
	if err != nil {
		goto error
	}

	cl = new(Clone)
	err = cl.Add(root, "clone", p.OsUsers.Uid2User(os.Geteuid()), nil, 0444, cl)
	if err != nil {
		goto error
	}

	s = srv.NewFileSrv(root)
	s.Dotu = true

	if *debug {
		s.Debuglevel = 1
	}

	s.Start(s)
	err = s.StartNetListener("tcp", *addr)
	if err != nil {
		goto error
	}
	return

error:
	log.Println(fmt.Sprintf("Error: %s", err))
}
