package node

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"k8s.io/minikube/cmd/minikube/profile"
	cmdutil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/node"
)

func NewCmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists all nodes",
		Long:  "Lists all nodes",
		Run:   list,
	}
}

func list(cmd *cobra.Command, args []string) {
	configs, err := profile.LoadClusterConfigs()
	if err != nil {
		glog.Errorln("Error loading cluster configs: ", err)
		cmdutil.MaybeReportErrorAndExit(err)
	}

	api, err := machine.NewAPIClient()
	if err != nil {
		glog.Errorf("Error getting client: %s\n", err)
		os.Exit(1)
	}
	defer api.Close()

	fmt.Printf("%-20s %-20s %-16s %-20s\n", "CLUSTER", "NODE", "IP", "STATUS")

	nodesFound := false
	for _, c := range configs {
		for _, nc := range c.Nodes {
			nodesFound = true
			n := node.NewNode(nc, c.MachineConfig, c.ClusterName, api)
			status, err := n.Status()
			if err != nil {
				status = minikube.NodeStatus("Error: " + err.Error())
			}

			ip := ""
			if status == minikube.StatusRunning {
				ip, err = n.IP()
				if err != nil {
					glog.Errorf("Error getting IP address for node %s: %s", nc.Name, err)
					cmdutil.MaybeReportErrorAndExit(err)
				}
			}

			fmt.Printf("%-20s %-20s %-16s %-20s\n", c.ClusterName, nc.Name, ip, status)
		}
	}

	if !nodesFound {
		fmt.Println("No nodes found.")
	}
}
