package hosttest

import (
	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/version"
)

const (
	DefaultHostName    = "test-host"
	HostTestCaCert     = "test-cert"
	HostTestPrivateKey = "test-key"
)

type DriverOptionsMock struct {
	Data map[string]interface{}
}

func (d DriverOptionsMock) String(key string) string {
	return d.Data[key].(string)
}

func (d DriverOptionsMock) StringSlice(key string) []string {
	return d.Data[key].([]string)
}

func (d DriverOptionsMock) Int(key string) int {
	return d.Data[key].(int)
}

func (d DriverOptionsMock) Bool(key string) bool {
	return d.Data[key].(bool)
}

func GetTestDriverFlags() *DriverOptionsMock {
	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"name":            DefaultHostName,
			"url":             "unix:///var/run/docker.sock",
			"swarm":           false,
			"swarm-host":      "",
			"swarm-master":    false,
			"swarm-discovery": "",
		},
	}
	return flags
}

func GetDefaultTestHost() (*host.Host, error) {
	hostOptions := &host.Options{
		EngineOptions: &engine.Options{},
		SwarmOptions:  &swarm.Options{},
		AuthOptions: &auth.Options{
			CaCertPath:       HostTestCaCert,
			CaPrivateKeyPath: HostTestPrivateKey,
		},
	}

	driver := none.NewDriver(DefaultHostName, "/tmp/artifacts")

	host := &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          DefaultHostName,
		Driver:        driver,
		DriverName:    "none",
		HostOptions:   hostOptions,
	}

	flags := GetTestDriverFlags()
	if err := host.Driver.SetConfigFromFlags(flags); err != nil {
		return nil, err
	}

	return host, nil
}
