package kubeadm

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
)

type NodeBootstrapper struct {
	config bootstrapper.KubernetesConfig
	ui     io.Writer
}

func NewNodeBootstrapper(c bootstrapper.KubernetesConfig, ui io.Writer) *NodeBootstrapper {
	return &NodeBootstrapper{config: c, ui: ui}
}

func (nb *NodeBootstrapper) Bootstrap(n minikube.Node) error {
	ip, err := n.IP()
	if err != nil {
		return errors.Wrap(err, "Error getting node's IP")
	}

	runner, err := n.Runner()
	if err != nil {
		return errors.Wrap(err, "Error getting node's runner")
	}

	b := NewKubeadmBootstrapperForRunner(n.MachineName(), ip, runner)

	fmt.Fprintln(nb.ui, "Moving assets into node...")
	if err := b.UpdateNode(nb.config); err != nil {
		return errors.Wrap(err, "Error updating node")
	}
	fmt.Fprintln(nb.ui, "Setting up certs...")
	if err := b.SetupCerts(nb.config); err != nil {
		return errors.Wrap(err, "Error configuring authentication")
	}

	fmt.Fprintln(nb.ui, "Joining node to cluster...")
	if err := b.JoinNode(nb.config); err != nil {
		return errors.Wrap(err, "Error joining node to cluster")
	}
	return nil
}
