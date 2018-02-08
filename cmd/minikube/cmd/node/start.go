package node

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/cmd/minikube/profile"
	cmdutil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/bootstrapper/kubeadm"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/node"
)

func startNode(cmd *cobra.Command, args []string) {
	if len(args) == 0 || args[0] == "" {
		glog.Error("node_name is required.")
		os.Exit(1)
	}

	nodeName := args[0]
	clusterName := viper.GetString(cfg.MachineProfile)

	cfg, err := profile.LoadConfigFromFile(clusterName)
	if err != nil {
		glog.Errorln("Error loading profile config: ", err)
		cmdutil.MaybeReportErrorAndExit(err)
	}

	fmt.Println("Starting nodes...")

	api, err := machine.NewAPIClient()
	if err != nil {
		glog.Errorf("Error getting client: %s\n", err)
		os.Exit(1)
	}
	defer api.Close()

	for _, nodeCfg := range cfg.Nodes {
		name := nodeCfg.Name
		if name != nodeName {
			continue
		}

		fmt.Printf("Starting node: %s\n", name)

		n := node.NewNode(nodeCfg, cfg.MachineConfig, clusterName, api)
		if err := n.Start(); err != nil {
			glog.Errorln("Error starting node machine: ", err)
			cmdutil.MaybeReportErrorAndExit(err)
		}

		b := kubeadm.NewNodeBootstrapper(cfg.KubernetesConfig, os.Stdout)
		if err := b.Bootstrap(n); err != nil {
			glog.Errorln("Error bootstrapping node: ", err)
			cmdutil.MaybeReportErrorAndExit(err)
		}
		fmt.Printf("Node %s started and configured.\n", n.Name())
	}
}
