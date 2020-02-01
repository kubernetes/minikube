package kic

import (
	"fmt"
	"os/exec"
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
	for {
		services, err := t.v1Core.Services("").List(metav1.ListOptions{})
		if err != nil {
			// do I return error, or log and continue?
			return err
		}

		for _, s := range services.Items {
			if s.Spec.Type == v1.ServiceTypeLoadBalancer {
				t.createTunnel(s.Name, s.Spec.ClusterIP, s.Spec.Ports)
			}
		}

		// which time to use?
		time.Sleep(10 * time.Second)
	}
}

func (t *Tunnel) createTunnel(name, clusterIP string, ports []v1.ServicePort) {
	sshArgs := []string{
		"-N",
		"docker@127.0.0.1",
		"-p", t.sshPort,
		"-i", t.sshKey,
	}

	for _, port := range ports {
		arg := fmt.Sprintf(
			"-L %d:%s:%d",
			port.Port,
			clusterIP,
			port.Port,
		)

		sshArgs = append(sshArgs, arg)
	}

	cmd := exec.Command("ssh", sshArgs...)
	err := cmd.Run()
	fmt.Println(err)
}
