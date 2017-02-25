package main

import (
	"flag"
	"io"
	"k8s.io/minikube/third_party/go9p/p"
	"k8s.io/minikube/third_party/go9p/p/clnt"
	"log"
	"os"
)

var debuglevel = flag.Int("d", 0, "debuglevel")
var addr = flag.String("addr", "127.0.0.1:5640", "network address")

func main() {
	var n int
	var user p.User
	var err error
	var c *clnt.Clnt
	var file *clnt.File
	var buf []byte

	flag.Parse()
	user = p.OsUsers.Uid2User(os.Geteuid())
	clnt.DefaultDebuglevel = *debuglevel
	c, err = clnt.Mount("tcp", *addr, "", user)
	if err != nil {
		goto error
	}

	if flag.NArg() != 1 {
		log.Println("invalid arguments")
		return
	}

	file, err = c.FOpen(flag.Arg(0), p.OREAD)
	if err != nil {
		goto error
	}

	buf = make([]byte, 8192)
	for {
		n, err = file.Read(buf)
		if n == 0 {
			break
		}

		os.Stdout.Write(buf[0:n])
	}

	file.Close()

	if err != nil && err != io.EOF {
		goto error
	}

	return

error:
	log.Println("Error", err)
}
