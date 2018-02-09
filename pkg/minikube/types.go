package minikube

import (
	"k8s.io/minikube/pkg/minikube/bootstrapper/runner"
)

const (
	StatusStopped    = "Stopped"
	StatusRunning    = "Running"
	StatusNotCreated = "NotCreated"
)

type NodeStatus string

type NodeConfig struct {
	Name string
}

type Node interface {
	Config() NodeConfig
	Start() error
	Stop() error
	Status() (NodeStatus, error)
	Runner() (runner.CommandRunner, error)
	MachineName() string
	Name() string
	IP() (string, error)
}

type Bootstrapper interface {
	Bootstrap(Node) error
}
