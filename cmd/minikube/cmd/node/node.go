package node

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"k8s.io/minikube/cmd/minikube/profile"
	"k8s.io/minikube/pkg/minikube"
)

const internalErrorCode = -1

func NewCmdNode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node SUBCOMMAND [flags]",
		Short: "Control a minikube cluster's nodes",
		Long:  `Control a cluster's nodes using subcommands like "minikube node add <node_name>"`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(NewCmdAdd())
	cmd.AddCommand(NewCmdList())
	cmd.AddCommand(NewCmdRemove())
	cmd.AddCommand(NewCmdSsh())
	cmd.AddCommand(NewCmdStart())
	return cmd
}

func getMachineName(clusterName string, node minikube.NodeConfig) string {
	return fmt.Sprintf("%s-%s", clusterName, node.Name)
}

func getNode(clusterName, nodeName string) (minikube.NodeConfig, error) {
	cfg, err := profile.LoadConfigFromFile(clusterName)
	if err != nil && !os.IsNotExist(err) {
		return minikube.NodeConfig{}, errors.Errorf("Error loading profile config: %s", err)
	}

	for _, node := range cfg.Nodes {
		if node.Name == nodeName {
			return node, nil
		}
	}

	return minikube.NodeConfig{}, errors.Errorf("Node not found in cluster. cluster: %s node: %s", clusterName, nodeName)
}
