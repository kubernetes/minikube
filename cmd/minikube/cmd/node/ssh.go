package node

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/pkg/minikube/cluster"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
)

func ssh(cmd *cobra.Command, args []string) {
	clusterName := viper.GetString(cfg.MachineProfile)

	if len(args) == 0 || args[0] == "" {
		glog.Error("node_name is required.")
		os.Exit(1)
	}

	nodeName := args[0]
	args = args[1:]

	node, err := getNode(clusterName, nodeName)
	if err != nil {
		glog.Error("Error loading node: ", err)
		os.Exit(1)
	}

	api, err := machine.NewAPIClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
		os.Exit(1)
	}
	defer api.Close()

	machineName := getMachineName(clusterName, node)
	host, err := cluster.CheckIfApiExistsAndLoadByName(machineName, api)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting host: %s\n", err)
		os.Exit(1)
	}
	if host.Driver.DriverName() == "none" {
		fmt.Println(`'none' driver does not support 'minikube ssh' command`)
		os.Exit(0)
	}
	err = cluster.CreateSSHShellByName(machineName, api, args)
	if err != nil {
		glog.Errorln(errors.Wrap(err, "Error attempting to ssh/run-ssh-command"))
		os.Exit(1)
	}
}
