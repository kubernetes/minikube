// Copyright 2009 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ufs

import (
	"fmt"
	"log"

	"k8s.io/minikube/third_party/go9p"
)

func StartServer(addrVal string, debugVal int, rootVal string) {
	ufs := new(go9p.Ufs)
	ufs.Dotu = true
	ufs.Id = "ufs"
	ufs.Root = rootVal
	ufs.Debuglevel = debugVal
	ufs.Start(ufs)

	fmt.Print("ufs starting\n")
	// determined by build tags
	// extraFuncs()
	err := ufs.StartNetListener("tcp", addrVal)
	if err != nil {
		log.Println(err)
	}
}
