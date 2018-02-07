package node

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/util"
)

func NewNode(
	config minikube.NodeConfig,
	baseConfig cluster.MachineConfig,
	clusterName string,
	api libmachine.API,
) minikube.Node {
	return &node{
		api:         api,
		config:      config,
		baseConfig:  baseConfig,
		clusterName: clusterName,
	}
}

type node struct {
	api         libmachine.API
	config      minikube.NodeConfig
	baseConfig  cluster.MachineConfig
	clusterName string
}

func (n *node) Config() minikube.NodeConfig {
	return n.config
}

func (n *node) IP() (string, error) {
	host, err := cluster.CheckIfApiExistsAndLoadByName(n.MachineName(), n.api)
	if err != nil {
		return "", err
	}

	ip, err := host.Driver.GetIP()
	return ip, errors.Wrap(err, "Error getting IP")
}

func (n *node) MachineName() string {
	return fmt.Sprintf("%s-node-%s", n.clusterName, n.config.Name)
}

func (n *node) Start() error {
	_, err := cluster.StartHost(n.api, n.machineConfig())
	if err != nil {
		return err
	}
	return nil
}

func (n *node) Stop() error {
	return fmt.Errorf("Not implemented yet")
}

func (n *node) Status() (minikube.NodeStatus, error) {
	s, err := n.status()
	return s, errors.Wrap(err, "Error getting node status.")
}

func (n *node) Runner() (bootstrapper.CommandRunner, error) {
	h, err := n.api.Load(n.MachineName())
	if err != nil {
		return nil, errors.Wrap(err, "Error loading host")
	}

	// The none driver executes commands directly on the host
	if h.Driver.DriverName() == constants.DriverNone {
		return &bootstrapper.ExecRunner{}, nil
	}
	client, err := sshutil.NewSSHClient(h.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "getting ssh client")
	}
	return bootstrapper.NewSSHRunner(client), nil
}

func (n *node) machineConfig() cluster.MachineConfig {
	cfg := n.baseConfig
	cfg.Downloader = util.DefaultDownloader{}
	cfg.MachineName = n.MachineName()
	return cfg
}

func (n *node) status() (minikube.NodeStatus, error) {
	if exists, err := n.api.Exists(n.MachineName()); err == nil && !exists {
		return minikube.StatusNotCreated, nil
	} else if err != nil {
		return minikube.NodeStatus(""), err
	}

	host, err := n.api.Load(n.MachineName())
	if err != nil {
		return minikube.NodeStatus(""), err
	}

	s, err := host.Driver.GetState()
	if err != nil {
		return minikube.NodeStatus(""), err
	}

	switch s {
	case state.Running:
		return minikube.StatusRunning, nil
	case state.Starting:
		return minikube.StatusRunning, nil
	case state.Stopping:
		return minikube.StatusRunning, nil
	case state.Stopped:
		return minikube.StatusStopped, nil
	case state.Paused:
		return minikube.StatusStopped, nil
	case state.Saved:
		return minikube.StatusStopped, nil
	case state.Error:
		return minikube.NodeStatus(""), errors.Errorf("Error state %s from libmachine", s)
	case state.Timeout:
		return minikube.NodeStatus(""), errors.Errorf("Error state %s from libmachine", s)
	default:
		return minikube.NodeStatus(""), errors.Errorf("Unknown state %s from libmachine", s)
	}
}
