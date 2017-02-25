// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"k8s.io/minikube/third_party/go9p/p"
	"k8s.io/minikube/third_party/go9p/p/srv"
	"log"
	"os"
)

type Ramfs struct {
	srv     *srv.Fsrv
	user    p.User
	group   p.Group
	blksz   int
	blkchan chan []byte
	zero    []byte // blksz array of zeroes
}

type RFile struct {
	srv.File
	data [][]byte
}

var addr = flag.String("addr", ":5640", "network address")
var debug = flag.Int("d", 0, "debuglevel")
var blksize = flag.Int("b", 8192, "block size")
var logsz = flag.Int("l", 2048, "log size")
var rsrv Ramfs

func (f *RFile) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	f.Lock()
	defer f.Unlock()

	if offset > f.Length {
		return 0, nil
	}

	count := uint32(len(buf))
	if offset+uint64(count) > f.Length {
		count = uint32(f.Length - offset)
	}

	for n, off, b := offset/uint64(rsrv.blksz), offset%uint64(rsrv.blksz), buf[0:count]; len(b) > 0; n++ {
		m := rsrv.blksz - int(off)
		if m > len(b) {
			m = len(b)
		}

		blk := rsrv.zero
		if len(f.data[n]) != 0 {
			blk = f.data[n]
		}

		//		log.Stderr("read block", n, "off", off, "len", m, "l", len(blk), "ll", len(b))
		copy(b, blk[off:off+uint64(m)])
		b = b[m:]
		off = 0
	}

	return int(count), nil
}

func (f *RFile) Write(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	f.Lock()
	defer f.Unlock()

	// make sure the data array is big enough
	sz := offset + uint64(len(buf))
	if f.Length < sz {
		f.expand(sz)
	}

	count := 0
	for n, off := offset/uint64(rsrv.blksz), offset%uint64(rsrv.blksz); len(buf) > 0; n++ {
		blk := f.data[n]
		if len(blk) == 0 {
			select {
			case blk = <-rsrv.blkchan:
				break
			default:
				blk = make([]byte, rsrv.blksz)
			}

			//			if off>0 {
			copy(blk, rsrv.zero /*[0:off]*/)
			//			}

			f.data[n] = blk
		}

		m := copy(blk[off:], buf)
		buf = buf[m:]
		count += m
		off = 0
	}

	return count, nil
}

func (f *RFile) Create(fid *srv.FFid, name string, perm uint32) (*srv.File, error) {
	ff := new(RFile)
	err := ff.Add(&f.File, name, rsrv.user, rsrv.group, perm, ff)
	return &ff.File, err
}

func (f *RFile) Remove(fid *srv.FFid) error {
	f.trunc(0)

	return nil
}

func (f *RFile) Wstat(fid *srv.FFid, dir *p.Dir) error {
	var uid, gid uint32

	f.Lock()
	defer f.Unlock()

	up := rsrv.srv.Upool
	uid = dir.Uidnum
	gid = dir.Gidnum
	if uid == p.NOUID && dir.Uid != "" {
		user := up.Uname2User(dir.Uid)
		if user == nil {
			return srv.Enouser
		}

		f.Uidnum = uint32(user.Id())
	}

	if gid == p.NOUID && dir.Gid != "" {
		group := up.Gname2Group(dir.Gid)
		if group == nil {
			return srv.Enouser
		}

		f.Gidnum = uint32(group.Id())
	}

	if dir.Mode != 0xFFFFFFFF {
		f.Mode = (f.Mode &^ 0777) | (dir.Mode & 0777)
	}

	if dir.Name != "" {
		if err := f.Rename(dir.Name); err != nil {
			return err
		}
	}

	if dir.Length != 0xFFFFFFFFFFFFFFFF {
		f.trunc(dir.Length)
	}

	return nil
}

// called with f locked
func (f *RFile) trunc(sz uint64) {
	if f.Length == sz {
		return
	}

	if f.Length > sz {
		f.shrink(sz)
	} else {
		f.expand(sz)
	}
}

// called with f lock held
func (f *RFile) shrink(sz uint64) {
	blknum := sz / uint64(rsrv.blksz)
	off := sz % uint64(rsrv.blksz)
	if off > 0 {
		if len(f.data[blknum]) > 0 {
			copy(f.data[blknum][off:], rsrv.zero)
		}

		blknum++
	}

	for i := blknum; i < uint64(len(f.data)); i++ {
		if len(f.data[i]) == 0 {
			continue
		}

		select {
		case rsrv.blkchan <- f.data[i]:
			break
		default:
		}
	}

	f.data = f.data[0:blknum]
	f.Length = sz
}

func (f *RFile) expand(sz uint64) {
	blknum := sz / uint64(rsrv.blksz)
	if sz%uint64(rsrv.blksz) != 0 {
		blknum++
	}

	data := make([][]byte, blknum)
	if f.data != nil {
		copy(data, f.data)
	}

	f.data = data
	f.Length = sz
}

func main() {
	var err error
	var l *p.Logger

	flag.Parse()
	rsrv.user = p.OsUsers.Uid2User(os.Geteuid())
	rsrv.group = p.OsUsers.Gid2Group(os.Getegid())
	rsrv.blksz = *blksize
	rsrv.blkchan = make(chan []byte, 2048)
	rsrv.zero = make([]byte, rsrv.blksz)

	root := new(RFile)
	err = root.Add(nil, "/", rsrv.user, nil, p.DMDIR|0777, root)
	if err != nil {
		goto error
	}

	l = p.NewLogger(*logsz)
	rsrv.srv = srv.NewFileSrv(&root.File)
	rsrv.srv.Dotu = true
	rsrv.srv.Debuglevel = *debug
	rsrv.srv.Start(rsrv.srv)
	rsrv.srv.Id = "ramfs"
	rsrv.srv.Log = l

	err = rsrv.srv.StartNetListener("tcp", *addr)
	if err != nil {
		goto error
	}
	return

error:
	log.Println(fmt.Sprintf("Error: %s", err))
}
