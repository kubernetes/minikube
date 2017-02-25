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
	var n, m int
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

	file, err = c.FOpen(flag.Arg(0), p.OWRITE|p.OTRUNC)
	if err != nil {
		file, err = c.FCreate(flag.Arg(0), 0666, p.OWRITE)
		if err != nil {
			goto error
		}
	}

	buf = make([]byte, 8192)
	for {
		n, err = os.Stdin.Read(buf)
		if err != nil && err != io.EOF {
			goto error
		}

		if n == 0 {
			break
		}

		m, err = file.Write(buf[0:n])
		if err != nil {
			goto error
		}

		if m != n {
			err = &p.Error{"short write", 0}
			goto error
		}
	}

	file.Close()
	return

error:
	log.Println("Error", err)
}
