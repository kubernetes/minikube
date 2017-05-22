package hyperkit

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/cloudflare/cfssl/log"
	"github.com/docker/machine/libmachine/mcnutils"
)

func createDiskImage(sshKeyPath, diskPath string, diskSize int) error {
	tarBuf, err := mcnutils.MakeDiskImage(sshKeyPath)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(diskPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Seek(0, os.SEEK_SET)

	if _, err := file.Write(tarBuf.Bytes()); err != nil {
		return err
	}
	file.Close()

	if err := os.Truncate(diskPath, int64(diskSize*1048576)); err != nil {
		return err
	}
	return nil
}

func fixPermissions(path string) error {
	os.Chown(path, syscall.Getuid(), syscall.Getegid())
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		fp := filepath.Join(path, f.Name())
		log.Debugf(fp)
		if err := os.Chown(fp, syscall.Getuid(), syscall.Getegid()); err != nil {
			return err
		}
	}
	return nil
}
