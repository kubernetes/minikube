package cmd

import (
	"flag"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/templates"
	cmdfull "k8s.io/minikube/cmd/minikube/cmd"
	configCmd "k8s.io/minikube/cmd/minikube/cmd/config"

	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/translate"
)

const (
	MinikubeConfig = "minikube-config"
	KubeConfig     = "kube-config"
)

var RootCmd = &cobra.Command{
	Use:               "sudominikube",
	Short:             "sudominikube executes some of minikube's function without inputing sudo password everytime",
	Long:              "sudominikube executes some of minikube's function without inputing sudo password everytime",
	PersistentPreRun:  cmdfull.RootCmd.PersistentPreRun,
	PersistentPostRun: cmdfull.RootCmd.PersistentPostRun,
}

func init() {
	// it will be called in init() of minikube/cmd
	// klog.InitFlags(nil)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine) // avoid `generate-docs_test.go` complaining about "Docs are not updated"

	RootCmd.PersistentFlags().StringP(config.ProfileName, "p", constants.DefaultClusterName, `The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently.`)
	RootCmd.PersistentFlags().StringP(configCmd.Bootstrapper, "b", "kubeadm", "The name of the cluster bootstrapper that will set up the Kubernetes cluster.")
	RootCmd.PersistentFlags().String(config.UserFlag, "", "Specifies the user executing the operation. Useful for auditing operations executed by 3rd party tools. Defaults to the operating system username.")
	RootCmd.PersistentFlags().Bool(config.SkipAuditFlag, false, "Skip recording the current command in the audit logs.")
	RootCmd.PersistentFlags().Bool(config.Rootless, false, "Force to use rootless driver (docker and podman driver only)")

	// To make sudominikube can be pointed to correct minikube abd kubectl configs, two global additional arguments are added
	RootCmd.PersistentFlags().String(MinikubeConfig, "", "Specifies the location of minikube's home folder (MINIKUBE_HOME)")
	RootCmd.PersistentFlags().String(KubeConfig, "", "Specifies the location of kubectl's config file (KUBECONFIG)")

	groups := templates.CommandGroups{
		{
			Message: translate.T("Sudo Commands:"),
			Commands: []*cobra.Command{
				// add commands here
				TunnelCmd,
			},
		},
		{
			Message: translate.T("Basic Commands:"),
			Commands: []*cobra.Command{
				// add commands here
				versionCmd,
			},
		},
	}
	groups.Add(RootCmd)
	templates.ActsAsRootCommand(RootCmd, []string{"options"}, groups...)

	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		exit.Error(reason.InternalBindFlags, "Unable to bind flags", err)
	}
	// it will be called in init() of minikube/cmd
	// translate.DetermineLocale()
	// it will be called in init() of minikube/cmd
	// cobra.OnInitialize(cmdfull.InitConfig)
}

func Execute() {
	// Check whether this is a windows binary (.exe) running inisde WSL.
	if runtime.GOOS != "linux" {
		exit.Message(reason.DrvUnsupportedOS, "sudominikube only works in linux")
	}

	for _, c := range RootCmd.Commands() {
		c.Short = translate.T(c.Short)
		c.Long = translate.T(c.Long)
		c.Flags().VisitAll(func(f *pflag.Flag) {
			f.Usage = translate.T(f.Usage)
		})

		c.SetUsageTemplate(cmdfull.UsageTemplate())
	}
	RootCmd.Short = translate.T(RootCmd.Short)
	RootCmd.Long = translate.T(RootCmd.Long)
	RootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Usage = translate.T(f.Usage)
	})

	// Universally ensure that we never speak to the wrong DOCKER_HOST
	if err := oci.PointToHostDockerDaemon(); err != nil {
		klog.Errorf("oci env: %v", err)
	}

	if err := oci.PointToHostPodman(); err != nil {
		klog.Errorf("oci env: %v", err)
	}

	if err := RootCmd.Execute(); err != nil {
		// Cobra already outputs the error, typically because the user provided an unknown command.
		defer os.Exit(reason.ExProgramUsage)
	}
}
