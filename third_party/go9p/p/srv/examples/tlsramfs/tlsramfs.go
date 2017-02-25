// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Listen on SSL connection, can be used as an example with p/clnt/examples/tls.go
// Sample certificate was copied from the Go source code
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"fmt"
	"k8s.io/minikube/third_party/go9p/p"
	"k8s.io/minikube/third_party/go9p/p/srv"
	"log"
	"math/big"
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

	flag.Parse()
	rsrv.user = p.OsUsers.Uid2User(os.Geteuid())
	rsrv.group = p.OsUsers.Gid2Group(os.Getegid())
	rsrv.blksz = *blksize
	rsrv.blkchan = make(chan []byte, 2048)
	rsrv.zero = make([]byte, rsrv.blksz)

	root := new(RFile)
	err = root.Add(nil, "/", rsrv.user, nil, p.DMDIR|0777, root)
	if err != nil {
		log.Println(fmt.Sprintf("Error: %s", err))
		return
	}

	l := p.NewLogger(*logsz)
	rsrv.srv = srv.NewFileSrv(&root.File)
	rsrv.srv.Dotu = true
	rsrv.srv.Debuglevel = *debug
	rsrv.srv.Start(rsrv.srv)
	rsrv.srv.Id = "ramfs"
	rsrv.srv.Log = l

	cert := make([]tls.Certificate, 1)
	cert[0].Certificate = [][]byte{testCertificate}
	cert[0].PrivateKey = testPrivateKey

	ls, oerr := tls.Listen("tcp", *addr, &tls.Config{
		Rand:               rand.Reader,
		Certificates:       cert,
		CipherSuites:       []uint16{tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA},
		InsecureSkipVerify: true,
	})
	if oerr != nil {
		log.Println("can't listen:", oerr)
		return
	}

	err = rsrv.srv.StartListener(ls)
	if err != nil {
		log.Println(fmt.Sprintf("Error: %s", err))
		return
	}
	return
}

// Copied from crypto/tls/handshake_server_test.go from the Go repository
func bigFromString(s string) *big.Int {
	ret := new(big.Int)
	ret.SetString(s, 10)
	return ret
}

func fromHex(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

var testCertificate = fromHex("308202b030820219a00302010202090085b0bba48a7fb8ca300d06092a864886f70d01010505003045310b3009060355040613024155311330110603550408130a536f6d652d53746174653121301f060355040a1318496e7465726e6574205769646769747320507479204c7464301e170d3130303432343039303933385a170d3131303432343039303933385a3045310b3009060355040613024155311330110603550408130a536f6d652d53746174653121301f060355040a1318496e7465726e6574205769646769747320507479204c746430819f300d06092a864886f70d010101050003818d0030818902818100bb79d6f517b5e5bf4610d0dc69bee62b07435ad0032d8a7a4385b71452e7a5654c2c78b8238cb5b482e5de1f953b7e62a52ca533d6fe125c7a56fcf506bffa587b263fb5cd04d3d0c921964ac7f4549f5abfef427100fe1899077f7e887d7df10439c4a22edb51c97ce3c04c3b326601cfafb11db8719a1ddbdb896baeda2d790203010001a381a73081a4301d0603551d0e04160414b1ade2855acfcb28db69ce2369ded3268e18883930750603551d23046e306c8014b1ade2855acfcb28db69ce2369ded3268e188839a149a4473045310b3009060355040613024155311330110603550408130a536f6d652d53746174653121301f060355040a1318496e7465726e6574205769646769747320507479204c746482090085b0bba48a7fb8ca300c0603551d13040530030101ff300d06092a864886f70d010105050003818100086c4524c76bb159ab0c52ccf2b014d7879d7a6475b55a9566e4c52b8eae12661feb4f38b36e60d392fdf74108b52513b1187a24fb301dbaed98b917ece7d73159db95d31d78ea50565cd5825a2d5a5f33c4b6d8c97590968c0f5298b5cd981f89205ff2a01ca31b9694dda9fd57e970e8266d71999b266e3850296c90a7bdd9")

var testPrivateKey = &rsa.PrivateKey{
	PublicKey: rsa.PublicKey{
		N: bigFromString("131650079503776001033793877885499001334664249354723305978524647182322416328664556247316495448366990052837680518067798333412266673813370895702118944398081598789828837447552603077848001020611640547221687072142537202428102790818451901395596882588063427854225330436740647715202971973145151161964464812406232198521"),
		E: 65537,
	},
	D: bigFromString("29354450337804273969007277378287027274721892607543397931919078829901848876371746653677097639302788129485893852488285045793268732234230875671682624082413996177431586734171663258657462237320300610850244186316880055243099640544518318093544057213190320837094958164973959123058337475052510833916491060913053867729"),
	Primes: []*big.Int{
		bigFromString("11969277782311800166562047708379380720136961987713178380670422671426759650127150688426177829077494755200794297055316163155755835813760102405344560929062149"),
		bigFromString("10998999429884441391899182616418192492905073053684657075974935218461686523870125521822756579792315215543092255516093840728890783887287417039645833477273829"),
	},
}
