package main

import (
	"flag"
	"k8s.io/minikube/third_party/go9p/p"
	"k8s.io/minikube/third_party/go9p/p/clnt"
	"log"
	"os"
	"strings"
)

var debuglevel = flag.Int("d", 0, "debuglevel")
var addr = flag.String("addr", "127.0.0.1:5640", "network address")

func main() {
	var user p.User
	var ba [][]byte
	var nreqs int
	var rchan chan *clnt.Req
	var tag *clnt.Tag
	var fid *clnt.Fid
	var wnames []string

	flag.Parse()
	user = p.OsUsers.Uid2User(os.Geteuid())
	clnt.DefaultDebuglevel = *debuglevel
	c, err := clnt.Mount("tcp", *addr, "", user)
	if err != nil {
		goto error
	}

	if flag.NArg() != 1 {
		log.Println("invalid arguments")
		return
	}

	ba = make([][]byte, 100)
	for i := 0; i < len(ba); i++ {
		ba[i] = make([]byte, 8192)
	}

	nreqs = 0
	rchan = make(chan *clnt.Req)
	tag = c.TagAlloc(rchan)

	// walk the file
	wnames = strings.Split(flag.Arg(0), "/")
	for wnames[0] == "" {
		wnames = wnames[1:]
	}

	fid = c.FidAlloc()
	for root := c.Root; len(wnames) > 0; root = fid {
		n := len(wnames)
		if n > 8 {
			n = 8
		}

		err = tag.Walk(root, fid, wnames[0:n])
		if err != nil {
			goto error
		}

		nreqs++
		wnames = wnames[n:]
	}
	err = tag.Open(fid, p.OREAD)
	if err != nil {
		goto error
	}

	for i := 0; i < len(ba); i++ {
		err = tag.Read(fid, uint64(i*8192), 8192)
		if err != nil {
			goto error
		}
		nreqs++
	}

	err = tag.Clunk(fid)

	// now start reading...
	for nreqs > 0 {
		r := <-rchan
		if r.Tc.Type == p.Tread {
			i := r.Tc.Offset / 8192
			copy(ba[i], r.Rc.Data)
			ba[i] = ba[i][0:r.Rc.Count]
		}
		nreqs--
	}

	for i := 0; i < len(ba); i++ {
		os.Stdout.Write(ba[i])
	}

	return

error:
	log.Println("error: ", err)
}
