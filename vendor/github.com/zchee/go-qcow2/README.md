go-qcow2
========

[![GoDoc](https://godoc.org/github.com/zchee/go-qcow2?status.svg)](https://godoc.org/github.com/zchee/go-qcow2)

Manage the QEMU qcow2 image format written in Go.

Project Goals
=============

Fully implement the management of the qcow2 image format written in Go.  
Without importing the C(cgo) files related to the QEMU.

Mainly, this package was written for [docker-machine-driver-xhyve](https://github.com/zchee/docker-machine-driver-xhyve).

License
=======

This project is released under the BSD license. Same as the typical Go packages.

qcow2 image format specifications is under the QEMU license.
