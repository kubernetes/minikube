// +build !darwin

package parallels

import "github.com/docker/machine/libmachine/drivers"

func NewDriver(hostName, storePath string) drivers.Driver {
	return drivers.NewDriverNotSupported("parallels", hostName, storePath)
}
