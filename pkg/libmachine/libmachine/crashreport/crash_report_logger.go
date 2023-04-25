package crashreport

import "k8s.io/minikube/pkg/libmachine/libmachine/log"

type logger struct{}

func (d *logger) Printf(fmtString string, args ...interface{}) {
	log.Debugf(fmtString, args)
}
