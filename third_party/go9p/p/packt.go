// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

// Create a Tversion message in the specified Fcall.
func PackTversion(fc *Fcall, msize uint32, version string) error {
	size := 4 + 2 + len(version) /* msize[4] version[s] */
	p, err := packCommon(fc, size, Tversion)
	if err != nil {
		return err
	}

	fc.Msize = msize
	fc.Version = version
	p = pint32(msize, p)
	p = pstr(version, p)

	return nil
}

// Create a Tauth message in the specified Fcall.
func PackTauth(fc *Fcall, fid uint32, uname string, aname string, unamenum uint32, dotu bool) error {
	size := 4 + 2 + 2 + len(uname) + len(aname) /* fid[4] uname[s] aname[s] */
	if dotu {
		size += 4 /* n_uname[4] */
	}

	p, err := packCommon(fc, size, Tauth)
	if err != nil {
		return err
	}

	fc.Fid = fid
	fc.Uname = uname
	fc.Aname = aname
	p = pint32(fid, p)
	p = pstr(uname, p)
	p = pstr(aname, p)
	if dotu {
		fc.Unamenum = unamenum
		p = pint32(unamenum, p)
	}

	return nil
}

// Create a Tflush message in the specified Fcall.
func PackTflush(fc *Fcall, oldtag uint16) error {
	p, err := packCommon(fc, 2, Tflush)
	if err != nil {
		return err
	}

	fc.Oldtag = oldtag
	p = pint16(oldtag, p)
	return nil
}

// Create a Tattach message in the specified Fcall. If dotu is true,
// the function will create 9P2000.u including the nuname value, otherwise
// nuname is ignored.
func PackTattach(fc *Fcall, fid uint32, afid uint32, uname string, aname string, unamenum uint32, dotu bool) error {
	size := 4 + 4 + 2 + len(uname) + 2 + len(aname) /* fid[4] afid[4] uname[s] aname[s] */
	if dotu {
		size += 4
	}

	p, err := packCommon(fc, size, Tattach)
	if err != nil {
		return err
	}

	fc.Fid = fid
	fc.Afid = afid
	fc.Uname = uname
	fc.Aname = aname
	p = pint32(fid, p)
	p = pint32(afid, p)
	p = pstr(uname, p)
	p = pstr(aname, p)
	if dotu {
		fc.Unamenum = unamenum
		p = pint32(unamenum, p)
	}

	return nil
}

// Create a Twalk message in the specified Fcall.
func PackTwalk(fc *Fcall, fid uint32, newfid uint32, wnames []string) error {
	nwname := len(wnames)
	size := 4 + 4 + 2 + nwname*2 /* fid[4] newfid[4] nwname[2] nwname*wname[s] */
	for i := 0; i < nwname; i++ {
		size += len(wnames[i])
	}

	p, err := packCommon(fc, size, Twalk)
	if err != nil {
		return err
	}

	fc.Fid = fid
	fc.Newfid = newfid
	p = pint32(fid, p)
	p = pint32(newfid, p)
	p = pint16(uint16(nwname), p)
	fc.Wname = make([]string, nwname)
	for i := 0; i < nwname; i++ {
		fc.Wname[i] = wnames[i]
		p = pstr(wnames[i], p)
	}

	return nil
}

// Create a Topen message in the specified Fcall.
func PackTopen(fc *Fcall, fid uint32, mode uint8) error {
	size := 4 + 1 /* fid[4] mode[1] */
	p, err := packCommon(fc, size, Topen)
	if err != nil {
		return err
	}

	fc.Fid = fid
	fc.Mode = mode
	p = pint32(fid, p)
	p = pint8(mode, p)
	return nil
}

// Create a Tcreate message in the specified Fcall. If dotu is true,
// the function will create a 9P2000.u message that includes ext.
// Otherwise the ext value is ignored.
func PackTcreate(fc *Fcall, fid uint32, name string, perm uint32, mode uint8, ext string, dotu bool) error {
	size := 4 + 2 + len(name) + 4 + 1 /* fid[4] name[s] perm[4] mode[1] */

	if dotu {
		size += 2 + len(ext)
	}

	p, err := packCommon(fc, size, Tcreate)
	if err != nil {
		return err
	}

	fc.Fid = fid
	fc.Name = name
	fc.Perm = perm
	fc.Mode = mode
	p = pint32(fid, p)
	p = pstr(name, p)
	p = pint32(perm, p)
	p = pint8(mode, p)

	if dotu {
		fc.Ext = ext
		p = pstr(ext, p)
	}

	return nil
}

// Create a Tread message in the specified Fcall.
func PackTread(fc *Fcall, fid uint32, offset uint64, count uint32) error {
	size := 4 + 8 + 4 /* fid[4] offset[8] count[4] */
	p, err := packCommon(fc, size, Tread)
	if err != nil {
		return err
	}

	fc.Fid = fid
	fc.Offset = offset
	fc.Count = count
	p = pint32(fid, p)
	p = pint64(offset, p)
	p = pint32(count, p)
	return nil
}

// Create a Twrite message in the specified Fcall.
func PackTwrite(fc *Fcall, fid uint32, offset uint64, count uint32, data []byte) error {
	c := len(data)
	size := 4 + 8 + 4 + c /* fid[4] offset[8] count[4] data[count] */
	p, err := packCommon(fc, size, Twrite)
	if err != nil {
		return err
	}

	fc.Fid = fid
	fc.Offset = offset
	fc.Count = count
	p = pint32(fid, p)
	p = pint64(offset, p)
	p = pint32(count, p)
	fc.Data = p
	copy(fc.Data, data)
	return nil
}

// Create a Tclunk message in the specified Fcall.
func PackTclunk(fc *Fcall, fid uint32) error {
	p, err := packCommon(fc, 4, Tclunk) /* fid[4] */
	if err != nil {
		return err
	}

	fc.Fid = fid
	p = pint32(fid, p)
	return nil
}

// Create a Tremove message in the specified Fcall.
func PackTremove(fc *Fcall, fid uint32) error {
	p, err := packCommon(fc, 4, Tremove) /* fid[4] */
	if err != nil {
		return err
	}

	fc.Fid = fid
	p = pint32(fid, p)
	return nil
}

// Create a Tstat message in the specified Fcall.
func PackTstat(fc *Fcall, fid uint32) error {
	p, err := packCommon(fc, 4, Tstat) /* fid[4] */
	if err != nil {
		return err
	}

	fc.Fid = fid
	p = pint32(fid, p)
	return nil
}

// Create a Twstat message in the specified Fcall. If dotu is true
// the function will create 9P2000.u message, otherwise the 9P2000.u
// specific fields from the Stat value will be ignored.
func PackTwstat(fc *Fcall, fid uint32, d *Dir, dotu bool) error {
	stsz := statsz(d, dotu)
	size := 4 + 2 + stsz /* fid[4] stat[n] */
	p, err := packCommon(fc, size, Twstat)
	if err != nil {
		return err
	}

	fc.Fid = fid
	fc.Dir = *d
	p = pint32(fid, p)
	p = pint16(uint16(stsz), p)
	p = pstat(d, p, dotu)
	return nil
}
