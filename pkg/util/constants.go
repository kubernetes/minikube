package util

import "fmt"

const (
	LocalkubeDirectory = "/var/lib/localkube"
	DNSDomain          = "cluster.local"
	CertPath           = "/var/lib/localkube/certs/"
)

func GetAlternateDNS(domain string) []string {
	return []string{fmt.Sprintf("%s.%s", "kubernetes.default.svc", domain), "kubernetes.default.svc", "kubernetes.default", "kubernetes"}
}
