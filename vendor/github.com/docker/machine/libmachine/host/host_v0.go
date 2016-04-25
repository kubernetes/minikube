package host

import "github.com/docker/machine/libmachine/drivers"

type V0 struct {
	Name          string `json:"-"`
	Driver        drivers.Driver
	DriverName    string
	ConfigVersion int
	HostOptions   *Options

	StorePath      string
	CaCertPath     string
	PrivateKeyPath string
	ServerCertPath string
	ServerKeyPath  string
	ClientCertPath string
	SwarmHost      string
	SwarmMaster    bool
	SwarmDiscovery string
	ClientKeyPath  string
}

type MetadataV0 struct {
	HostOptions Options
	DriverName  string

	ConfigVersion  int
	StorePath      string
	CaCertPath     string
	PrivateKeyPath string
	ServerCertPath string
	ServerKeyPath  string
	ClientCertPath string
}
