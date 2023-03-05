package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdfull "k8s.io/minikube/cmd/minikube/cmd"
)

var TunnelCmd = &cobra.Command{
	Use:              "tunnel",
	Short:            "Connect to LoadBalancer services",
	Long:             `tunnel creates a route to services deployed with type LoadBalancer and sets their Ingress to their ClusterIP. for a detailed example see https://minikube.sigs.k8s.io/docs/tasks/loadbalancer`,
	PersistentPreRun: cmdfull.TunnelCmd.PersistentPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		// point minikube to correct minikube abd kubectl configs
		minikubeConfig := viper.GetString(MinikubeConfig)
		kubeConfig := viper.GetString(KubeConfig)
		os.Setenv("MINIKUBE_HOME", minikubeConfig)
		os.Setenv("KUBECONFIG", kubeConfig)
		cmdfull.TunnelCmd.Run(cmd, args)
	},
}
