package metadata

import (
	"bytes"
	"fmt"
	"os/exec"
)

var metadataCorefileConfigmap = `apiVersion: v1
data:
  Corefile: |
    .:53 {
        errors
        health
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        rewrite name metadata.google.internal metadata.metadata.svc.cluster.local
        prometheus :9153
        proxy . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
`

var originalCorefileConfigmap = `apiVersion: v1
data:
  Corefile: |
    .:53 {
        errors
        health
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        proxy . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
`

func updateConfigmap(data string) error {
	// get current configmap
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	reader := bytes.NewReader([]byte(data))
	cmd.Stdin = reader

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(string(output))
		return err
	}
	return nil
}
