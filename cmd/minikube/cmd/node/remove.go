package node

import (
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/cmd/minikube/profile"
	cmdutil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
)

func remove(cmd *cobra.Command, args []string) {
	clusterName := viper.GetString(cfg.MachineProfile)

	if len(args) == 0 || args[0] == "" {
		glog.Error("node_name is required.")
		os.Exit(1)
	}

	nodeName := args[0]

	cfg, err := profile.LoadConfigFromFile(clusterName)
	if err != nil && !os.IsNotExist(err) {
		glog.Errorln("Error loading profile config: ", err)
		cmdutil.MaybeReportErrorAndExit(err)
	}
	api, err := machine.NewAPIClient()
	if err != nil {
		glog.Errorf("Error getting client: %s\n", err)
		os.Exit(1)
	}
	defer api.Close()

	for i, node := range cfg.Nodes {
		if node.Name == nodeName {
			machineName := getMachineName(clusterName, node)
			exists, err := api.Exists(machineName)
			if err != nil {
				glog.Errorln("Error removing node: ", err)
				os.Exit(1)
			}

			if exists {
				if err := cluster.DeleteHost(machineName, api); err != nil {
					glog.Errorln("Error removing node: ", err)
					os.Exit(1)
				}
			}

			cfg.Nodes = append(cfg.Nodes[:i], cfg.Nodes[i+1:]...)
			break

		} else if i == len(cfg.Nodes)-1 {
			glog.Errorln("Node not found: ", nodeName)
			os.Exit(1)
		}
	}

	if err := profile.SaveConfig(clusterName, cfg); err != nil {
		glog.Errorln("Error saving profile cluster configuration: ", err)
		os.Exit(1)
	}
}
