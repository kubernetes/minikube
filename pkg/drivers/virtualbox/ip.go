package virtualbox

import (
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnutils"
)

// IPWaiter waits for an IP to be configured.
type IPWaiter interface {
	Wait(d *Driver) error
}

func NewIPWaiter() IPWaiter {
	return &sshIPWaiter{}
}

type sshIPWaiter struct{}

func (w *sshIPWaiter) Wait(d *Driver) error {
	// Wait for SSH over NAT to be available before returning to user
	if err := drivers.WaitForSSH(d); err != nil {
		return err
	}

	// Bail if we don't get an IP from DHCP after a given number of seconds.
	if err := mcnutils.WaitForSpecific(d.hostOnlyIPAvailable, 5, 4*time.Second); err != nil {
		return err
	}

	var err error
	d.IPAddress, err = d.GetIP()

	return err
}
