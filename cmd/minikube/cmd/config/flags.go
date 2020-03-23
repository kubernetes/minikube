package config

import (
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
)

// ClusterFlagValue returns the current cluster name based on flags
func ClusterFlagValue() string {
	return viper.GetString(config.ProfileName)
}
