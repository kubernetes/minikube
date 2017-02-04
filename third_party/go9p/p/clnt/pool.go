// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clnt

var m2id = [...]uint8{
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 4,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 5,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 4,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 6,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 4,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 5,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 4,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 7,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 4,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 5,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 4,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 6,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 4,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 5,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 4,
	0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 0,
}

func newPool(maxid uint32) *pool {
	p := new(pool)
	p.maxid = maxid
	p.nchan = make(chan uint32)

	return p
}

func (p *pool) getId() uint32 {
	var n uint32 = 0
	var ret uint32

	p.Lock()
	for n = 0; n < uint32(len(p.imap)); n++ {
		if p.imap[n] != 0xFF {
			break
		}
	}

	if int(n) >= len(p.imap) {
		m := uint32(len(p.imap) + 32)
		if uint32(m*8) > p.maxid {
			m = p.maxid/8 + 1
		}

		b := make([]byte, m)
		copy(b, p.imap)
		p.imap = b
	}

	if n >= uint32(len(p.imap)) {
		p.need++
		p.Unlock()
		ret = <-p.nchan
	} else {
		ret = uint32(m2id[p.imap[n]])
		p.imap[n] |= 1 << ret
		ret += n * 8
		p.Unlock()
	}

	return ret
}

func (p *pool) putId(id uint32) {
	p.Lock()
	if p.need > 0 {
		p.nchan <- id
		p.need--
		p.Unlock()
		return
	}

	p.imap[id/8] &= ^(1 << (id % 8))
	p.Unlock()
}
