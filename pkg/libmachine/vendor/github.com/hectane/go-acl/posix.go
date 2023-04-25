//+build !windows

package acl

import "os"

// Chmod is os.Chmod.
var Chmod = os.Chmod
