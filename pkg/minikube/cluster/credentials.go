package cluster

import (
	"net"

	"k8s.io/minikube/pkg/util"
)

func GenerateCerts(pub, priv string, ip net.IP) error {
	ips := []net.IP{ip}
	if err := util.GenerateSelfSignedCert(pub, priv, ips, util.GetAlternateDNS(util.DNSDomain)); err != nil {
		return err
	}
	return nil
}
