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

var (
	NodeCmd *cobra.Command
)

func init() {
	NodeCmd = &cobra.Command{
		Use:   "node SUBCOMMAND [flags]",
		Short: "Control a minikube cluster's nodes",
		Long:  `Control a cluster's nodes using subcommands like "minikube node add <node_name>"`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	NodeCmd.AddCommand(&cobra.Command{
		Use:   "add [flags]",
		Short: "Adds a node to the cluster",
		Long:  "Adds a node tot the cluster",
		Run:   add,
	})
	NodeCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Lists all nodes",
		Long:  "Lists all nodes",
		Run:   list,
	})
	NodeCmd.AddCommand(&cobra.Command{
		Use:   "remove <node_name>",
		Short: "Removes a node from the cluster",
		Long:  "Removes a node from the cluster",
		Run:   remove,
	})
	NodeCmd.AddCommand(&cobra.Command{
		Use:   "ssh",
		Short: "Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'",
		Long:  "Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'.",
		Run:   ssh,
	})
	NodeCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts all nodes",
		Long:  "Starts all nodes",
		Run:   startNode,
	})
}

func getMachineName(clusterName string, node minikube.NodeConfig) string {
	return fmt.Sprintf("%s-node-%s", clusterName, node.Name)
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
