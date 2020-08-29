// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The p9 package go9provides the definitions and functions used to implement
// the 9P2000 protocol.
// TODO.
// All the packet conversion code in this file is crap and needs a rewrite.
package go9p

import (
	"fmt"
)

// 9P2000 message types
const (
	Tversion = 100 + iota
	Rversion
	Tauth
	Rauth
	Tattach
	Rattach
	Terror
	Rerror
	Tflush
	Rflush
	Twalk
	Rwalk
	Topen
	Ropen
	Tcreate
	Rcreate
	Tread
	Rread
	Twrite
	Rwrite
	Tclunk
	Rclunk
	Tremove
	Rremove
	Tstat
	Rstat
	Twstat
	Rwstat
	Tlast
)

const (
	MSIZE   = 1048576 + IOHDRSZ // default message size (1048576+IOHdrSz)
	IOHDRSZ = 24                // the non-data size of the Twrite messages
	PORT    = 564               // default port for 9P file servers
)

// Qid types
const (
	QTDIR     = 0x80 // directories
	QTAPPEND  = 0x40 // append only files
	QTEXCL    = 0x20 // exclusive use files
	QTMOUNT   = 0x10 // mounted channel
	QTAUTH    = 0x08 // authentication file
	QTTMP     = 0x04 // non-backed-up file
	QTSYMLINK = 0x02 // symbolic link (Unix, 9P2000.u)
	QTLINK    = 0x01 // hard link (Unix, 9P2000.u)
	QTFILE    = 0x00
)

// Flags for the mode field in Topen and Tcreate messages
const (
	OREAD   = 0  // open read-only
	OWRITE  = 1  // open write-only
	ORDWR   = 2  // open read-write
	OEXEC   = 3  // execute (== read but check execute permission)
	OTRUNC  = 16 // or'ed in (except for exec), truncate file first
	OCEXEC  = 32 // or'ed in, close on exec
	ORCLOSE = 64 // or'ed in, remove on close
)

// File modes
const (
	DMDIR       = 0x80000000 // mode bit for directories
	DMAPPEND    = 0x40000000 // mode bit for append only files
	DMEXCL      = 0x20000000 // mode bit for exclusive use files
	DMMOUNT     = 0x10000000 // mode bit for mounted channel
	DMAUTH      = 0x08000000 // mode bit for authentication file
	DMTMP       = 0x04000000 // mode bit for non-backed-up file
	DMSYMLINK   = 0x02000000 // mode bit for symbolic link (Unix, 9P2000.u)
	DMLINK      = 0x01000000 // mode bit for hard link (Unix, 9P2000.u)
	DMDEVICE    = 0x00800000 // mode bit for device file (Unix, 9P2000.u)
	DMNAMEDPIPE = 0x00200000 // mode bit for named pipe (Unix, 9P2000.u)
	DMSOCKET    = 0x00100000 // mode bit for socket (Unix, 9P2000.u)
	DMSETUID    = 0x00080000 // mode bit for setuid (Unix, 9P2000.u)
	DMSETGID    = 0x00040000 // mode bit for setgid (Unix, 9P2000.u)
	DMREAD      = 0x4        // mode bit for read permission
	DMWRITE     = 0x2        // mode bit for write permission
	DMEXEC      = 0x1        // mode bit for execute permission
)

const (
	NOTAG uint16 = 0xFFFF     // no tag specified
	NOFID uint32 = 0xFFFFFFFF // no fid specified
	NOUID uint32 = 0xFFFFFFFF // no uid specified
)

// Error values
const (
	EPERM   = 1
	ENOENT  = 2
	EIO     = 5
	EEXIST  = 17
	ENOTDIR = 20
	EINVAL  = 22
)

// Error represents a 9P2000 (and 9P2000.u) error
type Error struct {
	Err      string // textual representation of the error
	Errornum uint32 // numeric representation of the error (9P2000.u)
}

// File identifier
type Qid struct {
	Type    uint8  // type of the file (high 8 bits of the mode)
	Version uint32 // version number for the path
	Path    uint64 // server's unique identification of the file
}

// Dir describes a file
type Dir struct {
	Size   uint16 // size-2 of the Dir on the wire
	Type   uint16
	Dev    uint32
	Qid           // file's Qid
	Mode   uint32 // permissions and flags
	Atime  uint32 // last access time in seconds
	Mtime  uint32 // last modified time in seconds
	Length uint64 // file length in bytes
	Name   string // file name
	Uid    string // owner name
	Gid    string // group name
	Muid   string // name of the last user that modified the file

	/* 9P2000.u extension */
	Ext     string // special file's descriptor
	Uidnum  uint32 // owner ID
	Gidnum  uint32 // group ID
	Muidnum uint32 // ID of the last user that modified the file
}

// Fcall represents a 9P2000 message
type Fcall struct {
	Size    uint32   // size of the message
	Type    uint8    // message type
	Fid     uint32   // file identifier
	Tag     uint16   // message tag
	Msize   uint32   // maximum message size (used by Tversion, Rversion)
	Version string   // protocol version (used by Tversion, Rversion)
	Oldtag  uint16   // tag of the message to flush (used by Tflush)
	Error   string   // error (used by Rerror)
	Qid              // file Qid (used by Rauth, Rattach, Ropen, Rcreate)
	Iounit  uint32   // maximum bytes read without breaking in multiple messages (used by Ropen, Rcreate)
	Afid    uint32   // authentication fid (used by Tauth, Tattach)
	Uname   string   // user name (used by Tauth, Tattach)
	Aname   string   // attach name (used by Tauth, Tattach)
	Perm    uint32   // file permission (mode) (used by Tcreate)
	Name    string   // file name (used by Tcreate)
	Mode    uint8    // open mode (used by Topen, Tcreate)
	Newfid  uint32   // the fid that represents the file walked to (used by Twalk)
	Wname   []string // list of names to walk (used by Twalk)
	Wqid    []Qid    // list of Qids for the walked files (used by Rwalk)
	Offset  uint64   // offset in the file to read/write from/to (used by Tread, Twrite)
	Count   uint32   // number of bytes read/written (used by Tread, Rread, Twrite, Rwrite)
	Data    []uint8  // data read/to-write (used by Rread, Twrite)
	Dir              // file description (used by Rstat, Twstat)

	/* 9P2000.u extensions */
	Errornum uint32 // error code, 9P2000.u only (used by Rerror)
	Ext      string // special file description, 9P2000.u only (used by Tcreate)
	Unamenum uint32 // user ID, 9P2000.u only (used by Tauth, Tattach)

	Pkt []uint8 // raw packet data
	Buf []uint8 // buffer to put the raw data in
}

// Interface for accessing users and groups
type Users interface {
	Uid2User(uid int) User
	Uname2User(uname string) User
	Gid2Group(gid int) Group
	Gname2Group(gname string) Group
}

// Represents a user
type User interface {
	Name() string          // user name
	Id() int               // user id
	Groups() []Group       // groups the user belongs to (can return nil)
	IsMember(g Group) bool // returns true if the user is member of the specified group
}

// Represents a group of users
type Group interface {
	Name() string    // group name
	Id() int         // group id
	Members() []User // list of members that belong to the group (can return nil)
}

// minimum size of a 9P2000 message for a type
var minFcsize = [...]uint32{
	6,  /* Tversion msize[4] version[s] */
	6,  /* Rversion msize[4] version[s] */
	8,  /* Tauth fid[4] uname[s] aname[s] */
	13, /* Rauth aqid[13] */
	12, /* Tattach fid[4] afid[4] uname[s] aname[s] */
	13, /* Rattach qid[13] */
	0,  /* Terror */
	2,  /* Rerror ename[s] (ecode[4]) */
	2,  /* Tflush oldtag[2] */
	0,  /* Rflush */
	10, /* Twalk fid[4] newfid[4] nwname[2] */
	2,  /* Rwalk nwqid[2] */
	5,  /* Topen fid[4] mode[1] */
	17, /* Ropen qid[13] iounit[4] */
	11, /* Tcreate fid[4] name[s] perm[4] mode[1] */
	17, /* Rcreate qid[13] iounit[4] */
	16, /* Tread fid[4] offset[8] count[4] */
	4,  /* Rread count[4] */
	16, /* Twrite fid[4] offset[8] count[4] */
	4,  /* Rwrite count[4] */
	4,  /* Tclunk fid[4] */
	0,  /* Rclunk */
	4,  /* Tremove fid[4] */
	0,  /* Rremove */
	4,  /* Tstat fid[4] */
	4,  /* Rstat stat[n] */
	8,  /* Twstat fid[4] stat[n] */
	0,  /* Rwstat */
	20, /* Tbread fileid[8] offset[8] count[4] */
	4,  /* Rbread count[4] */
	20, /* Tbwrite fileid[8] offset[8] count[4] */
	4,  /* Rbwrite count[4] */
	16, /* Tbtrunc fileid[8] offset[8] */
	0,  /* Rbtrunc */
}

// minimum size of a 9P2000.u message for a type
var minFcusize = [...]uint32{
	6,  /* Tversion msize[4] version[s] */
	6,  /* Rversion msize[4] version[s] */
	12, /* Tauth fid[4] uname[s] aname[s] */
	13, /* Rauth aqid[13] */
	16, /* Tattach fid[4] afid[4] uname[s] aname[s] */
	13, /* Rattach qid[13] */
	0,  /* Terror */
	6,  /* Rerror ename[s] (ecode[4]) */
	2,  /* Tflush oldtag[2] */
	0,  /* Rflush */
	10, /* Twalk fid[4] newfid[4] nwname[2] */
	2,  /* Rwalk nwqid[2] */
	5,  /* Topen fid[4] mode[1] */
	17, /* Ropen qid[13] iounit[4] */
	13, /* Tcreate fid[4] name[s] perm[4] mode[1] */
	17, /* Rcreate qid[13] iounit[4] */
	16, /* Tread fid[4] offset[8] count[4] */
	4,  /* Rread count[4] */
	16, /* Twrite fid[4] offset[8] count[4] */
	4,  /* Rwrite count[4] */
	4,  /* Tclunk fid[4] */
	0,  /* Rclunk */
	4,  /* Tremove fid[4] */
	0,  /* Rremove */
	4,  /* Tstat fid[4] */
	4,  /* Rstat stat[n] */
	8,  /* Twstat fid[4] stat[n] */
	20, /* Tbread fileid[8] offset[8] count[4] */
	4,  /* Rbread count[4] */
	20, /* Tbwrite fileid[8] offset[8] count[4] */
	4,  /* Rbwrite count[4] */
	16, /* Tbtrunc fileid[8] offset[8] */
	0,  /* Rbtrunc */
}

func gint8(buf []byte) (uint8, []byte) { return buf[0], buf[1:] }

func gint16(buf []byte) (uint16, []byte) {
	return uint16(buf[0]) | (uint16(buf[1]) << 8), buf[2:]
}

func gint32(buf []byte) (uint32, []byte) {
	return uint32(buf[0]) | (uint32(buf[1]) << 8) | (uint32(buf[2]) << 16) |
			(uint32(buf[3]) << 24),
		buf[4:]
}

func Gint32(buf []byte) (uint32, []byte) { return gint32(buf) }

func gint64(buf []byte) (uint64, []byte) {
	return uint64(buf[0]) | (uint64(buf[1]) << 8) | (uint64(buf[2]) << 16) |
			(uint64(buf[3]) << 24) | (uint64(buf[4]) << 32) | (uint64(buf[5]) << 40) |
			(uint64(buf[6]) << 48) | (uint64(buf[7]) << 56),
		buf[8:]
}

func gstr(buf []byte) (string, []byte) {
	var n uint16

	if buf == nil || len(buf) < 2 {
		return "", nil
	}

	n, buf = gint16(buf)

	if int(n) > len(buf) {
		return "", nil
	}

	return string(buf[0:n]), buf[n:]
}

func gqid(buf []byte, qid *Qid) []byte {
	qid.Type, buf = gint8(buf)
	qid.Version, buf = gint32(buf)
	qid.Path, buf = gint64(buf)

	return buf
}

func gstat(buf []byte, d *Dir, dotu bool) ([]byte, error) {
	sz := len(buf)
	d.Size, buf = gint16(buf)
	d.Type, buf = gint16(buf)
	d.Dev, buf = gint32(buf)
	buf = gqid(buf, &d.Qid)
	d.Mode, buf = gint32(buf)
	d.Atime, buf = gint32(buf)
	d.Mtime, buf = gint32(buf)
	d.Length, buf = gint64(buf)
	d.Name, buf = gstr(buf)
	if buf == nil {
		s := fmt.Sprintf("Buffer too short for basic 9p: need %d, have %d",
			49, sz)
		return nil, &Error{s, EINVAL}
	}

	d.Uid, buf = gstr(buf)
	if buf == nil {
		return nil, &Error{"d.Uid failed", EINVAL}
	}
	d.Gid, buf = gstr(buf)
	if buf == nil {
		return nil, &Error{"d.Gid failed", EINVAL}
	}

	d.Muid, buf = gstr(buf)
	if buf == nil {
		return nil, &Error{"d.Muid failed", EINVAL}
	}

	if dotu {
		d.Ext, buf = gstr(buf)
		if buf == nil {
			return nil, &Error{"d.Ext failed", EINVAL}
		}

		d.Uidnum, buf = gint32(buf)
		d.Gidnum, buf = gint32(buf)
		d.Muidnum, buf = gint32(buf)
	} else {
		d.Uidnum = NOUID
		d.Gidnum = NOUID
		d.Muidnum = NOUID
	}

	return buf, nil
}

func pint8(val uint8, buf []byte) []byte {
	buf[0] = val
	return buf[1:]
}

func pint16(val uint16, buf []byte) []byte {
	buf[0] = uint8(val)
	buf[1] = uint8(val >> 8)
	return buf[2:]
}

func pint32(val uint32, buf []byte) []byte {
	buf[0] = uint8(val)
	buf[1] = uint8(val >> 8)
	buf[2] = uint8(val >> 16)
	buf[3] = uint8(val >> 24)
	return buf[4:]
}

func pint64(val uint64, buf []byte) []byte {
	buf[0] = uint8(val)
	buf[1] = uint8(val >> 8)
	buf[2] = uint8(val >> 16)
	buf[3] = uint8(val >> 24)
	buf[4] = uint8(val >> 32)
	buf[5] = uint8(val >> 40)
	buf[6] = uint8(val >> 48)
	buf[7] = uint8(val >> 56)
	return buf[8:]
}

func pstr(val string, buf []byte) []byte {
	n := uint16(len(val))
	buf = pint16(n, buf)
	b := []byte(val)
	copy(buf, b)
	return buf[n:]
}

func pqid(val *Qid, buf []byte) []byte {
	buf = pint8(val.Type, buf)
	buf = pint32(val.Version, buf)
	buf = pint64(val.Path, buf)

	return buf
}

func statsz(d *Dir, dotu bool) int {
	sz := 2 + 2 + 4 + 13 + 4 + 4 + 4 + 8 + 2 + 2 + 2 + 2 + len(d.Name) + len(d.Uid) + len(d.Gid) + len(d.Muid)
	if dotu {
		sz += 2 + 4 + 4 + 4 + len(d.Ext)
	}

	return sz
}

func pstat(d *Dir, buf []byte, dotu bool) []byte {
	sz := statsz(d, dotu)
	buf = pint16(uint16(sz-2), buf)
	buf = pint16(d.Type, buf)
	buf = pint32(d.Dev, buf)
	buf = pqid(&d.Qid, buf)
	buf = pint32(d.Mode, buf)
	buf = pint32(d.Atime, buf)
	buf = pint32(d.Mtime, buf)
	buf = pint64(d.Length, buf)
	buf = pstr(d.Name, buf)
	buf = pstr(d.Uid, buf)
	buf = pstr(d.Gid, buf)
	buf = pstr(d.Muid, buf)
	if dotu {
		buf = pstr(d.Ext, buf)
		buf = pint32(d.Uidnum, buf)
		buf = pint32(d.Gidnum, buf)
		buf = pint32(d.Muidnum, buf)
	}

	return buf
}

// Converts a Dir value to its on-the-wire representation and writes it to
// the buf. Returns the number of bytes written, 0 if there is not enough space.
func PackDir(d *Dir, dotu bool) []byte {
	sz := statsz(d, dotu)
	buf := make([]byte, sz)
	pstat(d, buf, dotu)
	return buf
}

// Converts the on-the-wire representation of a stat to Stat value.
// Returns an error if the conversion is impossible, otherwise
// a pointer to a Stat value.
func UnpackDir(buf []byte, dotu bool) (d *Dir, b []byte, amt int, err error) {
	sz := 2 + 2 + 4 + 13 + 4 + /* size[2] type[2] dev[4] qid[13] mode[4] */
		4 + 4 + 8 + /* atime[4] mtime[4] length[8] */
		2 + 2 + 2 + 2 /* name[s] uid[s] gid[s] muid[s] */

	if dotu {
		sz += 2 + 4 + 4 + 4 /* extension[s] n_uid[4] n_gid[4] n_muid[4] */
	}

	if len(buf) < sz {
		s := fmt.Sprintf("short buffer: Need %d and have %v", sz, len(buf))
		return nil, nil, 0, &Error{s, EINVAL}
	}

	d = new(Dir)
	b, err = gstat(buf, d, dotu)
	if err != nil {
		return nil, nil, 0, err
	}

	return d, b, len(buf) - len(b), nil
}

// Allocates a new Fcall.
func NewFcall(sz uint32) *Fcall {
	fc := new(Fcall)
	fc.Buf = make([]byte, sz)

	return fc
}

// Sets the tag of a Fcall.
func SetTag(fc *Fcall, tag uint16) {
	fc.Tag = tag
	pint16(tag, fc.Pkt[5:])
}

func packCommon(fc *Fcall, size int, id uint8) ([]byte, error) {
	size += 4 + 1 + 2 /* size[4] id[1] tag[2] */
	if len(fc.Buf) < int(size) {
		return nil, &Error{"buffer too small", EINVAL}
	}

	fc.Size = uint32(size)
	fc.Type = id
	fc.Tag = NOTAG
	p := fc.Buf
	p = pint32(uint32(size), p)
	p = pint8(id, p)
	p = pint16(NOTAG, p)
	fc.Pkt = fc.Buf[0:size]

	return p, nil
}

func (err *Error) Error() string {
	if err != nil {
		return err.Err
	}

	return ""
}
