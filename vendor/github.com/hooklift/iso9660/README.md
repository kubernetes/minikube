# iso9660
[![GoDoc](https://godoc.org/github.com/hooklift/iso9660?status.svg)](https://godoc.org/github.com/hooklift/iso9660)
[![Build Status](https://travis-ci.org/hooklift/iso9660.svg?branch=master)](https://travis-ci.org/hooklift/iso9660)

Go library and CLI to extract data from ISO9660 images.

### CLI
```
$ ./iso9660
Usage:
  iso9660 <image-path> <destination-path>
  iso9660 -h | --help
  iso9660 --version
```

### Library
An example on how to use the library to extract files and directories from an ISO image can be found in our CLI source code at:
https://github.com/hooklift/iso9660/blob/master/cmd/iso9660/main.go

### Not supported
* Reading files recorded in interleave mode
* Multi-extent or reading individual files larger than 4GB
* Joliet extensions, meaning that file names longer than 32 characters are going to be truncated. Unicode characters are not going to be properly decoded either.
* Multisession extension
* Rock Ridge extension, making unable to return recorded POSIX permissions, timestamps as well as owner and group for files and directories.
