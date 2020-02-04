package cluster

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
)

// CreateSSHShell creates a new SSH shell / client
func CreateSSHShell(api libmachine.API, args []string) error {
	machineName := viper.GetString(config.MachineProfile)
	host, err := CheckIfHostExistsAndLoad(api, machineName)
	if err != nil {
		return errors.Wrap(err, "host exists and load")
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		return errors.Wrap(err, "state")
	}

	if currentState != state.Running {
		return errors.Errorf("%q is not running", machineName)
	}

	client, err := host.CreateSSHClient()
	if err != nil {
		return errors.Wrap(err, "Creating ssh client")
	}
	return client.Shell(args...)
}
