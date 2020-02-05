package kic

import (
	"fmt"
	"strings"
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

		for _, sshTunnel := range sshTunnels {
			sshTunnel.sigkill = true
		}

		for _, s := range services.Items {
			if s.Spec.Type == v1.ServiceTypeLoadBalancer {
				name := sshTunnelName(s.Name, s.Spec.ClusterIP, s.Spec.Ports)
				existingSSHTunnel, ok := sshTunnels[name]

				if ok {
					existingSSHTunnel.sigkill = false
					continue
				}

				newSSHTunnel := createSSHTunnel(t, name, s.Spec.ClusterIP, s.Spec.Ports)
				sshTunnels[newSSHTunnel.name] = newSSHTunnel
			}
		}

		for _, sshTunnel := range sshTunnels {
			if sshTunnel.sigkill {
				sshTunnel.stop()
				delete(sshTunnels, sshTunnel.name)
			}
		}

		// TODO: which time to use?
		time.Sleep(1 * time.Second)
	}
}

// sshTunnelName creaets a uniq name for the tunnel, using its name/clusterIP/ports.
// This allows a new process to be created if an existing service was changed,
// the new process will support the IP/Ports change ocurred.
func sshTunnelName(name, clusterIP string, ports []v1.ServicePort) string {
	n := []string{
		name,
		"-",
		clusterIP,
	}

	for _, port := range ports {
		n = append(n, fmt.Sprintf("-%d", port.Port))
	}

	return strings.Join(n, "")
}
