// Copyright 2012 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ufs

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	"k8s.io/minikube/third_party/go9p/p"
	"k8s.io/minikube/third_party/go9p/p/srv"
)

func atime(stat *syscall.Stat_t) time.Time {
	return time.Unix(stat.Atim.Unix())
}

// IsBlock reports if the file is a block device
func isBlock(d os.FileInfo) bool {
	stat := d.Sys().(*syscall.Stat_t)
	return (stat.Mode & syscall.S_IFMT) == syscall.S_IFBLK
}

// IsChar reports if the file is a character device
func isChar(d os.FileInfo) bool {
	stat := d.Sys().(*syscall.Stat_t)
	return (stat.Mode & syscall.S_IFMT) == syscall.S_IFCHR
}

func dir2Qid(d os.FileInfo) *p.Qid {
	var qid p.Qid

	qid.Path = d.Sys().(*syscall.Stat_t).Ino
	qid.Version = uint32(d.ModTime().UnixNano() / 1000000)
	qid.Type = dir2QidType(d)

	return &qid
}

func dir2Dir(path string, d os.FileInfo, dotu bool, upool p.Users) *p.Dir {
	sysMode := d.Sys().(*syscall.Stat_t)

	dir := new(Dir)
	dir.Qid = *dir2Qid(d)
	dir.Mode = dir2Npmode(d, dotu)
	dir.Atime = uint32(atime(sysMode).Unix())
	dir.Mtime = uint32(d.ModTime().Unix())
	dir.Length = uint64(d.Size())
	dir.Name = path[strings.LastIndex(path, "/")+1:]

	if dotu {
		dir.dotu(path, d, upool, sysMode)
		return &dir.Dir
	}

	unixUid := int(sysMode.Uid)
	unixGid := int(sysMode.Gid)
	dir.Uid = strconv.Itoa(unixUid)
	dir.Gid = strconv.Itoa(unixGid)
	dir.Muid = "none"

	// BUG(akumar): LookupId will never find names for
	// groups, as it only operates on user ids.
	u, err := user.LookupId(dir.Uid)
	if err == nil {
		dir.Uid = u.Username
	}
	g, err := user.LookupId(dir.Gid)
	if err == nil {
		dir.Gid = g.Username
	}

	return &dir.Dir
}

func (dir *Dir) dotu(path string, d os.FileInfo, upool p.Users, sysMode *syscall.Stat_t) {
	u := upool.Uid2User(int(sysMode.Uid))
	g := upool.Gid2Group(int(sysMode.Gid))
	dir.Uid = u.Name()
	if dir.Uid == "" {
		dir.Uid = "none"
	}

	dir.Gid = g.Name()
	if dir.Gid == "" {
		dir.Gid = "none"
	}
	dir.Muid = "none"
	dir.Ext = ""
	dir.Uidnum = uint32(u.Id())
	dir.Gidnum = uint32(g.Id())
	dir.Muidnum = p.NOUID
	if d.Mode()&os.ModeSymlink != 0 {
		var err error
		dir.Ext, err = os.Readlink(path)
		if err != nil {
			dir.Ext = ""
		}
	} else if isBlock(d) {
		dir.Ext = fmt.Sprintf("b %d %d", sysMode.Rdev>>24, sysMode.Rdev&0xFFFFFF)
	} else if isChar(d) {
		dir.Ext = fmt.Sprintf("c %d %d", sysMode.Rdev>>24, sysMode.Rdev&0xFFFFFF)
	}
}

func (*Ufs) Wstat(req *srv.Req) {
	fid := req.Fid.Aux.(*Fid)
	err := fid.stat()
	if err != nil {
		req.RespondError(err)
		return
	}

	dir := &req.Tc.Dir
	if dir.Mode != 0xFFFFFFFF {
		mode := dir.Mode & 0777
		if req.Conn.Dotu {
			if dir.Mode&p.DMSETUID > 0 {
				mode |= syscall.S_ISUID
			}
			if dir.Mode&p.DMSETGID > 0 {
				mode |= syscall.S_ISGID
			}
		}
		e := os.Chmod(fid.path, os.FileMode(mode))
		if e != nil {
			req.RespondError(toError(e))
			return
		}
	}

	uid, gid := p.NOUID, p.NOUID
	if req.Conn.Dotu {
		uid = dir.Uidnum
		gid = dir.Gidnum
	}

	// Try to find local uid, gid by name.
	if (dir.Uid != "" || dir.Gid != "") && !req.Conn.Dotu {
		uid, err = lookup(dir.Uid, false)
		if err != nil {
			req.RespondError(err)
			return
		}

		// BUG(akumar): Lookup will never find gids
		// corresponding to group names, because
		// it only operates on user names.
		gid, err = lookup(dir.Gid, true)
		if err != nil {
			req.RespondError(err)
			return
		}
	}

	if uid != p.NOUID || gid != p.NOUID {
		e := os.Chown(fid.path, int(uid), int(gid))
		if e != nil {
			req.RespondError(toError(e))
			return
		}
	}

	if dir.Name != "" {
		path := fid.path[0:strings.LastIndex(fid.path, "/")+1] + "/" + dir.Name
		err := syscall.Rename(fid.path, path)
		if err != nil {
			req.RespondError(toError(err))
			return
		}
		fid.path = path
	}

	if dir.Length != 0xFFFFFFFFFFFFFFFF {
		e := os.Truncate(fid.path, int64(dir.Length))
		if e != nil {
			req.RespondError(toError(e))
			return
		}
	}

	// If either mtime or atime need to be changed, then
	// we must change both.
	if dir.Mtime != ^uint32(0) || dir.Atime != ^uint32(0) {
		mt, at := time.Unix(int64(dir.Mtime), 0), time.Unix(int64(dir.Atime), 0)
		if cmt, cat := (dir.Mtime == ^uint32(0)), (dir.Atime == ^uint32(0)); cmt || cat {
			st, e := os.Stat(fid.path)
			if e != nil {
				req.RespondError(toError(e))
				return
			}
			switch cmt {
			case true:
				mt = st.ModTime()
			default:
				at = atime(st.Sys().(*syscall.Stat_t))
			}
		}
		e := os.Chtimes(fid.path, at, mt)
		if e != nil {
			req.RespondError(toError(e))
			return
		}
	}

	req.RespondRwstat()
}
