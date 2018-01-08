package cmd

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/cluster"
)

func startNodes(api libmachine.API, masterIP string, baseConfig cluster.Config, count int) error {
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s-%d", baseConfig.MachineConfig.MachineName, i+1)
		newConfig := newConfig(baseConfig.MachineConfig, name)
		glog.Infoln("Creating machine: %s", name)
		_, err := cluster.StartHost(api, newConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func newConfig(baseConfig cluster.MachineConfig, machineName string) cluster.MachineConfig {
	baseConfig.MachineName = machineName
	return baseConfig
}
