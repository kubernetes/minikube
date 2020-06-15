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
      log
      errors
      health {
         lameduck 5s
      }
      ready
      kubernetes cluster.local in-addr.arpa ip6.arpa {
         pods insecure
         fallthrough in-addr.arpa ip6.arpa
         ttl 30
      }
      rewrite name metadata.google.internal metadata.metadata.svc.cluster.local
      prometheus :9153
      forward . /etc/resolv.conf
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
      health {
         lameduck 5s
      }
      ready
      kubernetes cluster.local in-addr.arpa ip6.arpa {
         pods insecure
         fallthrough in-addr.arpa ip6.arpa
         ttl 30
      }
      prometheus :9153
      forward . /etc/resolv.conf
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
