package node

import (
	"fmt"
	"math/rand"
	"os"
	"text/template"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/cmd/minikube/profile"
	cmdutil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/bootstrapper/kubeadm"
	"k8s.io/minikube/pkg/minikube/cluster"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
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

	addCmd := &cobra.Command{
		Use:   "add [flags]",
		Short: "Adds a node to the cluster",
		Long:  "Adds a node tot the cluster",
		Run:   addNode,
	}
	NodeCmd.AddCommand(addCmd)

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Starts all nodes",
		Long:  "Starts all nodes",
		Run:   startNode,
	}
	NodeCmd.AddCommand(startCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all nodes",
		Long:  "Lists all nodes",
		Run:   listNodes,
	}
	NodeCmd.AddCommand(listCmd)

	removeCmd := &cobra.Command{
		Use:   "remove <node_name>",
		Short: "Removes a node from the cluster",
		Long:  "Removes a node from the cluster",
		Run:   removeNode,
	}
	NodeCmd.AddCommand(removeCmd)

	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'",
		Long:  "Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'.",
		Run:   ssh,
	}
	NodeCmd.AddCommand(sshCmd)
}

func addNode(cmd *cobra.Command, args []string) {
	// TODO Make clusterName into `--cluster=` flag
	clusterName := viper.GetString(cfg.MachineProfile)

	cfg, err := profile.LoadConfigFromFile(clusterName)
	if err != nil {
		glog.Errorln("Error loading profile config: ", err)
		cmdutil.MaybeReportErrorAndExit(err)
	}

	s := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	node := cluster.Node{
		Name: fmt.Sprintf("node-%d", s.Uint32()),
	}

	cfg.Nodes = append(cfg.Nodes, node)

	if err := profile.SaveConfig(clusterName, cfg); err != nil {
		glog.Errorln("Error saving profile cluster configuration: ", err)
		os.Exit(1)
	}

	fmt.Println("Added node: ", node.Name)
}

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

	for _, node := range cfg.Nodes {
		if node.Name != nodeName {
			continue
		}

		fmt.Printf("Starting node: %s\n", node.Name)

		machineCfg := cfg.MachineConfig
		machineCfg.MachineName = getMachineName(clusterName, node)

		host, err := cluster.StartHost(api, machineCfg)
		if err != nil {
			glog.Errorln("Error starting node machine: ", err)
			cmdutil.MaybeReportErrorAndExit(err)
		}

		k8sBootstrapper, err := kubeadm.NewKubeadmBootstrapperForMachine(machineCfg.MachineName, api)
		if err != nil {
			glog.Errorln("Error getting kubeadm bootstrapper: ", err)
			cmdutil.MaybeReportErrorAndExit(err)
		}

		fmt.Println("Moving assets into node...")
		if err := k8sBootstrapper.UpdateNode(cfg.KubernetesConfig); err != nil {
			glog.Errorln("Error updating node: ", err)
			cmdutil.MaybeReportErrorAndExit(err)
		}
		fmt.Println("Setting up certs...")
		if err := k8sBootstrapper.SetupCerts(cfg.KubernetesConfig); err != nil {
			glog.Errorln("Error configuring authentication: ", err)
			cmdutil.MaybeReportErrorAndExit(err)
		}

		fmt.Println("Joining node to cluster...")
		if err := k8sBootstrapper.JoinNode(cfg.KubernetesConfig); err != nil {
			glog.Errorln("Error joining node to cluster: ", err)
			cmdutil.MaybeReportErrorAndExit(err)
		}

		ip, err := host.Driver.GetIP()
		if err != nil {
			glog.Errorln("Error getting machine IP: ", err)
			cmdutil.MaybeReportErrorAndExit(err)
		}
		fmt.Printf("Node started. IP: %s\n", ip)
	}
}

func removeNode(cmd *cobra.Command, args []string) {
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

func getMachineName(clusterName string, node cluster.Node) string {
	return fmt.Sprintf("%s-node-%s", clusterName, node.Name)
}

func listNodes(cmd *cobra.Command, args []string) {
	clusterName := viper.GetString(cfg.MachineProfile)

	cfg, err := profile.LoadConfigFromFile(clusterName)
	if err != nil && !os.IsNotExist(err) {
		glog.Errorln("Error loading profile config: ", err)
		cmdutil.MaybeReportErrorAndExit(err)
	}

	tmpl, err := template.New("nodeeList").Parse("{{range .}}{{ .Name }}\n{{end}}")
	if err != nil {
		glog.Errorln("Error creating nodeList template:", err)
		os.Exit(internalErrorCode)
	}

	err = tmpl.Execute(os.Stdout, cfg.Nodes)
	if err != nil {
		glog.Errorln("Error executing nodeList template:", err)
		os.Exit(internalErrorCode)
	}
}

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

func getNode(clusterName, nodeName string) (cluster.Node, error) {
	cfg, err := profile.LoadConfigFromFile(clusterName)
	if err != nil && !os.IsNotExist(err) {
		return cluster.Node{}, errors.Errorf("Error loading profile config: %s", err)
	}

	for _, node := range cfg.Nodes {
		if node.Name == nodeName {
			return node, nil
		}
	}

	return cluster.Node{}, errors.Errorf("Node not found in cluster. cluster: %s node: %s", clusterName, nodeName)
}
