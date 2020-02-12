package kic

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/minikube/pkg/minikube/tunnel"
)

// Tunnel ...
type Tunnel struct {
	sshPort              string
	sshKey               string
	v1Core               typed_core.CoreV1Interface
	LoadBalancerEmulator tunnel.LoadBalancerEmulator
}

// NewTunnel ...
func NewTunnel(sshPort, sshKey string, v1Core typed_core.CoreV1Interface) *Tunnel {
	return &Tunnel{
		sshPort:              sshPort,
		sshKey:               sshKey,
		v1Core:               v1Core,
		LoadBalancerEmulator: tunnel.NewLoadBalancerEmulator(v1Core),
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
				name := sshTunnelName(s)
				existingSSHTunnel, ok := sshTunnels[name]

				if ok {
					existingSSHTunnel.sigkill = false
					continue
				}

				newSSHTunnel := createSSHTunnel(name, t.sshPort, t.sshKey, s)
				sshTunnels[newSSHTunnel.name] = newSSHTunnel
				_, err := t.LoadBalancerEmulator.PatchServicesIP("127.0.0.1")
				if err != nil {
					fmt.Println(err)
				}
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
func sshTunnelName(service v1.Service) string {
	n := []string{
		service.Name,
		"-",
		service.Spec.ClusterIP,
	}

	for _, port := range service.Spec.Ports {
		n = append(n, fmt.Sprintf("-%d", port.Port))
	}

	return strings.Join(n, "")
}
