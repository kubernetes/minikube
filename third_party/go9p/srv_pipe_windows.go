// Copyright 2009 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import (
	"os"
	"syscall"
)

func (dir *pipeDir) dotu(path string, d os.FileInfo, upool Users, sysMode *syscall.Win32FileAttributeData) {
	// u := upool.Uid2User(int(sysMode.Uid))
	// g := upool.Gid2Group(int(sysMode.Gid))
	// dir.Uid = u.Name()
	// if dir.Uid == "" {
	// 	dir.Uid = "none"
	// }

	// dir.Gid = g.Name()
	// if dir.Gid == "" {
	// 	dir.Gid = "none"
	// }
	// dir.Muid = "none"
	// dir.Ext = ""
	// dir.Uidnum = uint32(u.Id())
	// dir.Gidnum = uint32(g.Id())
	// dir.Muidnum = NOUID
	// if d.Mode()&os.ModeSymlink != 0 {
	// 	var err error
	// 	dir.Ext, err = os.Readlink(path)
	// 	if err != nil {
	// 		dir.Ext = ""
	// 	}
	// } else if isBlock(d) {
	// 	dir.Ext = fmt.Sprintf("b %d %d", sysMode.Rdev>>24, sysMode.Rdev&0xFFFFFF)
	// } else if isChar(d) {
	// 	dir.Ext = fmt.Sprintf("c %d %d", sysMode.Rdev>>24, sysMode.Rdev&0xFFFFFF)
	// }

	dir.Uid = "none"
	dir.Gid = "none"
	dir.Muid = "none"
	dir.Uidnum = 0
	dir.Gidnum = 0
	dir.Muidnum = NOUID
	dir.Ext = ""
}
