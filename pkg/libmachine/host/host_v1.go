package host

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
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
