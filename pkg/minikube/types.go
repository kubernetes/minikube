package minikube

import (
	"k8s.io/minikube/pkg/minikube/bootstrapper"
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
	Runner() (bootstrapper.CommandRunner, error)
	MachineName() string
	IP() (string, error)
}
