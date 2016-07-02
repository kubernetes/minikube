// +build darwin dragonfly freebsd linux,!arm64 netbsd openbsd

package panicwrap

import (
	"syscall"
)

func dup2(oldfd, newfd int) error {
	return syscall.Dup2(oldfd, newfd)
}
