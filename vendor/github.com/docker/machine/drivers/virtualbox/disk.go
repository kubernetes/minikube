package virtualbox

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
)

type VirtualDisk struct {
	UUID string
	Path string
}

type DiskCreator interface {
	Create(size int, publicSSHKeyPath, diskPath string) error
}

func NewDiskCreator() DiskCreator {
	return &defaultDiskCreator{}
}

type defaultDiskCreator struct{}

// Make a boot2docker VM disk image.
func (c *defaultDiskCreator) Create(size int, publicSSHKeyPath, diskPath string) error {
	log.Debugf("Creating %d MB hard disk image...", size)

	tarBuf, err := mcnutils.MakeDiskImage(publicSSHKeyPath)
	if err != nil {
		return err
	}

	log.Debug("Calling inner createDiskImage")

	return createDiskImage(diskPath, size, tarBuf)
}

// createDiskImage makes a disk image at dest with the given size in MB. If r is
// not nil, it will be read as a raw disk image to convert from.
func createDiskImage(dest string, size int, r io.Reader) error {
	// Convert a raw image from stdin to the dest VMDK image.
	sizeBytes := int64(size) << 20 // usually won't fit in 32-bit int (max 2GB)
	// FIXME: why isn't this just using the vbm*() functions?
	cmd := exec.Command(vboxManageCmd, "convertfromraw", "stdin", dest,
		fmt.Sprintf("%d", sizeBytes), "--format", "VMDK")

	log.Debug(cmd)

	if os.Getenv("MACHINE_DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	log.Debug("Starting command")

	if err := cmd.Start(); err != nil {
		return err
	}

	log.Debug("Copying to stdin")

	n, err := io.Copy(stdin, r)
	if err != nil {
		return err
	}

	log.Debug("Filling zeroes")

	// The total number of bytes written to stdin must match sizeBytes, or
	// VBoxManage.exe on Windows will fail. Fill remaining with zeros.
	if left := sizeBytes - n; left > 0 {
		if err := zeroFill(stdin, left); err != nil {
			return err
		}
	}

	log.Debug("Closing STDIN")

	// cmd won't exit until the stdin is closed.
	if err := stdin.Close(); err != nil {
		return err
	}

	log.Debug("Waiting on cmd")

	return cmd.Wait()
}

// zeroFill writes n zero bytes into w.
func zeroFill(w io.Writer, n int64) error {
	const blocksize = 32 << 10
	zeros := make([]byte, blocksize)
	var k int
	var err error
	for n > 0 {
		if n > blocksize {
			k, err = w.Write(zeros)
		} else {
			k, err = w.Write(zeros[:n])
		}
		if err != nil {
			return err
		}
		n -= int64(k)
	}
	return nil
}

func getVMDiskInfo(name string, vbox VBoxManager) (*VirtualDisk, error) {
	out, err := vbox.vbmOut("showvminfo", name, "--machinereadable")
	if err != nil {
		return nil, err
	}

	disk := &VirtualDisk{}

	err = parseKeyValues(out, reEqualQuoteLine, func(key, val string) error {
		switch key {
		case "SATA-1-0":
			disk.Path = val
		case "SATA-ImageUUID-1-0":
			disk.UUID = val
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return disk, nil
}
