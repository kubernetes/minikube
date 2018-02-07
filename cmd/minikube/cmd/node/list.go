package node

import (
	"os"
	"text/template"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/cmd/minikube/profile"
	cmdutil "k8s.io/minikube/cmd/util"
	cfg "k8s.io/minikube/pkg/minikube/config"
)

func list(cmd *cobra.Command, args []string) {
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
