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
	var user p.User
	var err error
	var c *clnt.Clnt
	var file *clnt.File
	var d []*p.Dir

	flag.Parse()
	user = p.OsUsers.Uid2User(os.Geteuid())
	clnt.DefaultDebuglevel = *debuglevel
	c, err = clnt.Mount("tcp", *addr, "", user)
	if err != nil {
		log.Fatal(err)
	}

	lsarg := "/"
	if flag.NArg() == 1 {
		lsarg = flag.Arg(0)
	} else if flag.NArg() > 1 {
		log.Fatal("error: only one argument expected")
	}

	file, err = c.FOpen(lsarg, p.OREAD)
	if err != nil {
		log.Fatal(err)
	}

	for {
		d, err = file.Readdir(0)
		if d == nil || len(d) == 0 || err != nil {
			break
		}

		for i := 0; i < len(d); i++ {
			os.Stdout.WriteString(d[i].Name + "\n")
		}
	}

	file.Close()
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	return
}
