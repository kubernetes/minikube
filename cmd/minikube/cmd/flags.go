package cmd

import (
	"fmt"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// Return a minikube command containing the current profile name
func minikubeCmd() string {
	cname := ClusterFlagValue()
	if cname != constants.DefaultClusterName {
		return fmt.Sprintf("minikube -p %s", cname)
	}
	return "minikube"
}

// ClusterFlagValue returns the current cluster name based on flags
func ClusterFlagValue() string {
	return viper.GetString(config.ProfileName)
}
