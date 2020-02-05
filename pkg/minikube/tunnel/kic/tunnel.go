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

type sshTunnel struct {
	name string
	kill bool
	cmd  *exec.Cmd
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
			v.kill = true
		}

		for _, s := range services.Items {
			if s.Spec.Type == v1.ServiceTypeLoadBalancer {
				sshTunnel, ok := sshTunnels[s.Name]

				if ok {
					sshTunnel.kill = false
					continue
				}

				newSSHTunnel := t.createSSHTunnel(s.Name, s.Spec.ClusterIP, s.Spec.Ports)
				sshTunnels[newSSHTunnel.name] = newSSHTunnel
			}
		}

		for _, v := range sshTunnels {
			if v.kill {
				v.stop()
				delete(sshTunnels, v.name)
			}
		}

		// which time to use?
		time.Sleep(1 * time.Second)
	}
}

func (t *Tunnel) createSSHTunnel(name, clusterIP string, ports []v1.ServicePort) *sshTunnel {
	// extract sshArgs
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

	// TODO: name must be different, because if a service was changed,
	// we must remove the old process and create the new one
	s := &sshTunnel{
		name: fmt.Sprintf("%s", name),
		kill: false,
		cmd:  cmd,
	}

	go s.run()

	return s
}

func (s *sshTunnel) run() {
	fmt.Println("running", s.name)
	err := s.cmd.Start()
	if err != nil {
		// TODO: improve logging
		fmt.Println(err)
	}

	// we are ignoring wait return, because the process will be killed, once the tunnel is not needed.
	s.cmd.Wait()
}

func (s *sshTunnel) stop() {
	fmt.Println("stopping", s.name)
	err := s.cmd.Process.Kill()
	if err != nil {
		// TODO: improve logging
		fmt.Println(err)
	}
}
