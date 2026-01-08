package host

import (
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/swarm"
)

type AuthOptionsV1 struct {
	StorePath            string
	CaCertPath           string
	CaCertRemotePath     string
	ServerCertPath       string
	ServerKeyPath        string
	ClientKeyPath        string
	ServerCertRemotePath string
	ServerKeyRemotePath  string
	PrivateKeyPath       string
	ClientCertPath       string
}

type OptionsV1 struct {
	Driver        string
	Memory        int
	Disk          int
	EngineOptions *engine.Options
	SwarmOptions  *swarm.Options
	AuthOptions   *AuthOptionsV1
}

type V1 struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *OptionsV1
	Name          string `json:"-"`
	StorePath     string
}
