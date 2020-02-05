package kic

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Tunnel ...
type Tunnel struct {
	sshPort string
	sshKey  string
	v1Core  typed_core.CoreV1Interface
}

// NewTunnel ...
func NewTunnel(sshPort, sshKey string, v1Core typed_core.CoreV1Interface) *Tunnel {
	return &Tunnel{
		sshPort: sshPort,
		sshKey:  sshKey,
		v1Core:  v1Core,
	}
}

// Start ...
func (t *Tunnel) Start() error {
	sshTunnels := make(map[string]*sshTunnel)

	for {
		services, err := t.v1Core.Services("").List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, v := range sshTunnels {
			v.sigkill = true
		}

		for _, s := range services.Items {
			if s.Spec.Type == v1.ServiceTypeLoadBalancer {
				sshTunnel, ok := sshTunnels[s.Name]

				if ok {
					sshTunnel.sigkill = false
					continue
				}

				newSSHTunnel := createSSHTunnel(t, s.Name, s.Spec.ClusterIP, s.Spec.Ports)
				sshTunnels[newSSHTunnel.name] = newSSHTunnel
			}
		}

		for _, v := range sshTunnels {
			if v.sigkill {
				v.stop()
				delete(sshTunnels, v.name)
			}
		}

		// which time to use?
		time.Sleep(1 * time.Second)
	}
}
