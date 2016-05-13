package crashreport

import "github.com/docker/machine/libmachine/log"

type logger struct{}

func (d *logger) Printf(fmtString string, args ...interface{}) {
	log.Debugf(fmtString, args)
}
