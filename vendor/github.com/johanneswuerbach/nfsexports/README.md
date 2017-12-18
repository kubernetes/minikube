# nfsexports [![Build Status](https://travis-ci.org/johanneswuerbach/nfsexports.svg?branch=master)](https://travis-ci.org/johanneswuerbach/nfsexports)

Go util to manage NFS exports `/etc/exports`.

## Features

* Add and remove exports
* Verify the added export is valid, before updating
* Auto-creates missing exports file, which auto starts nfsd on OS X
* Handles missing line breaks

```go
package main

import (
    "github.com/johanneswuerbach/nfsexports"
)

func main() {
  newExports, err := nfsexports.Add("", "myExport", "/Users 192.168.64.16 -alldirs -maproot=root")
  if err != nil {
      panic(err)
  }

  newExports, err := nfsexports.Remove("", "myExport")
  if err != nil {
      panic(err)
  }

  if err = nfsexports.ReloadDaemon(); err != nil {
      panic(err)
  }
}
```
