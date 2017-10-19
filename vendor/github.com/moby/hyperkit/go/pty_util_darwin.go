package hyperkit

/*
Most of this code was copied and adjusted from:
https://github.com/containerd/console
which is under Apache License Version 2.0, January 2004
*/
import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func tcget(fd uintptr, p *unix.Termios) error {
	return ioctl(fd, unix.TIOCGETA, uintptr(unsafe.Pointer(p)))
}

func tcset(fd uintptr, p *unix.Termios) error {
	return ioctl(fd, unix.TIOCSETA, uintptr(unsafe.Pointer(p)))
}

func ioctl(fd, flag, data uintptr) error {
	if _, _, err := unix.Syscall(unix.SYS_IOCTL, fd, flag, data); err != 0 {
		return err
	}
	return nil
}

func saneTerminal(f *os.File) error {
	// Go doesn't have a wrapper for any of the termios ioctls.
	var termios unix.Termios
	if err := tcget(f.Fd(), &termios); err != nil {
		return err
	}
	// Set -onlcr so we don't have to deal with \r.
	termios.Oflag &^= unix.ONLCR
	return tcset(f.Fd(), &termios)
}

func setRaw(f *os.File) error {
	var termios unix.Termios
	if err := tcget(f.Fd(), &termios); err != nil {
		return err
	}
	termios = cfmakeraw(termios)
	termios.Oflag = termios.Oflag | unix.OPOST
	return tcset(f.Fd(), &termios)
}

// isTerminal checks if the provided file is a terminal
func isTerminal(f *os.File) bool {
	var termios unix.Termios
	if tcget(f.Fd(), &termios) != nil {
		return false
	}
	return true
}

func cfmakeraw(t unix.Termios) unix.Termios {
	t.Iflag = uint64(uint32(t.Iflag) & ^uint32((unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON)))
	t.Oflag = uint64(uint32(t.Oflag) & ^uint32(unix.OPOST))
	t.Lflag = uint64(uint32(t.Lflag) & ^(uint32(unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN)))
	t.Cflag = uint64(uint32(t.Cflag) & ^(uint32(unix.CSIZE | unix.PARENB)))
	t.Cflag = t.Cflag | unix.CS8
	t.Cc[unix.VMIN] = 1
	t.Cc[unix.VTIME] = 0

	return t
}
