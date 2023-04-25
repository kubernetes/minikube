//+build windows

// Windows API functions for manipulating ACLs.
package api

import (
	"golang.org/x/sys/windows"
)

var advapi32 = windows.MustLoadDLL("advapi32.dll")
