/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	goflag "flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/kubectl/pkg/util/templates"
	configCmd "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/translate"
)

var dirs = [...]string{
	localpath.MiniPath(),
	localpath.MakeMiniPath("certs"),
	localpath.MakeMiniPath("machines"),
	localpath.MakeMiniPath("cache"),
	localpath.MakeMiniPath("cache", "iso"),
	localpath.MakeMiniPath("config"),
	localpath.MakeMiniPath("addons"),
	localpath.MakeMiniPath("files"),
	localpath.MakeMiniPath("logs"),
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "minikube",
	Short: "minikube quickly sets up a local Kubernetes cluster",
	Long:  `minikube provisions and manages local Kubernetes clusters optimized for development workflows.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		for _, path := range dirs {
			if err := os.MkdirAll(path, 0777); err != nil {
				exit.WithError("Error creating minikube directory", err)
			}
		}

		logDir := pflag.Lookup("log_dir")
		if !logDir.Changed {
			if err := logDir.Value.Set(localpath.MakeMiniPath("logs")); err != nil {
				exit.WithError("logdir set failed", err)
			}
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	_, callingCmd := filepath.Split(os.Args[0])
	if callingCmd == "kubectl" {
		os.Args = append([]string{"minikube", callingCmd}, os.Args[1:]...)
	}
	for _, c := range RootCmd.Commands() {
		c.Short = translate.T(c.Short)
		c.Long = translate.T(c.Long)
		c.Flags().VisitAll(func(flag *pflag.Flag) {
			flag.Usage = translate.T(flag.Usage)
		})

		c.SetUsageTemplate(usageTemplate())
	}
	RootCmd.Short = translate.T(RootCmd.Short)
	RootCmd.Long = translate.T(RootCmd.Long)
	RootCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = translate.T(flag.Usage)
	})

	if runtime.GOOS != "windows" {
		// add minikube binaries to the path
		targetDir := localpath.MakeMiniPath("bin")
		addToPath(targetDir)
	}

	// Universally ensure that we never speak to the wrong DOCKER_HOST
	if err := oci.PointToHostDockerDaemon(); err != nil {
		glog.Errorf("oci env: %v", err)
	}

	if err := oci.PointToHostPodman(); err != nil {
		glog.Errorf("oci env: %v", err)
	}

	if err := RootCmd.Execute(); err != nil {
		// Cobra already outputs the error, typically because the user provided an unknown command.
		os.Exit(exit.BadUsage)
	}
}

// usageTemplate just calls translate.T on the default usage template
// explicitly using the raw string instead of calling c.UsageTemplate()
// so the extractor can find this monstrosity of a string
func usageTemplate() string {
	return fmt.Sprintf(`%s:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

%s:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

%s{{end}}
`, translate.T("Usage"), translate.T("Aliases"), translate.T("Examples"), translate.T("Available Commands"), translate.T("Flags"), translate.T("Global Flags"), translate.T("Additional help topics"), translate.T(`Use "{{.CommandPath}} [command] --help" for more information about a command.`))
}

// Handle config values for flags used in external packages (e.g. glog)
// by setting them directly, using values from viper when not passed in as args
func setFlagsUsingViper() {
	for _, config := range []string{"alsologtostderr", "log_dir", "v"} {
		var a = pflag.Lookup(config)
		viper.SetDefault(a.Name, a.DefValue)
		// If the flag is set, override viper value
		if a.Changed {
			viper.Set(a.Name, a.Value.String())
		}
		// Viper will give precedence first to calls to the Set command,
		// then to values from the config.yml
		if err := a.Value.Set(viper.GetString(a.Name)); err != nil {
			exit.WithError(fmt.Sprintf("failed to set value for %q", a.Name), err)
		}
		a.Changed = true
	}
}

func init() {
	translate.DetermineLocale()
	RootCmd.PersistentFlags().StringP(config.ProfileName, "p", constants.DefaultClusterName, `The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently.`)
	RootCmd.PersistentFlags().StringP(configCmd.Bootstrapper, "b", "kubeadm", "The name of the cluster bootstrapper that will set up the Kubernetes cluster.")

	groups := templates.CommandGroups{
		{
			Message: translate.T("Basic Commands:"),
			Commands: []*cobra.Command{
				startCmd,
				statusCmd,
				stopCmd,
				deleteCmd,
				dashboardCmd,
				pauseCmd,
				unpauseCmd,
			},
		},
		{
			Message: translate.T("Images Commands:"),
			Commands: []*cobra.Command{
				dockerEnvCmd,
				podmanEnvCmd,
				cacheCmd,
			},
		},
		{
			Message: translate.T("Configuration and Management Commands:"),
			Commands: []*cobra.Command{
				configCmd.AddonsCmd,
				configCmd.ConfigCmd,
				configCmd.ProfileCmd,
				updateContextCmd,
			},
		},
		{
			Message: translate.T("Networking and Connectivity Commands:"),
			Commands: []*cobra.Command{
				serviceCmd,
				tunnelCmd,
			},
		},
		{
			Message: translate.T("Advanced Commands:"),
			Commands: []*cobra.Command{
				mountCmd,
				sshCmd,
				kubectlCmd,
				nodeCmd,
			},
		},
		{
			Message: translate.T("Troubleshooting Commands:"),
			Commands: []*cobra.Command{
				sshKeyCmd,
				ipCmd,
				logsCmd,
				updateCheckCmd,
				versionCmd,
				optionsCmd,
			},
		},
	}
	groups.Add(RootCmd)

	// Ungrouped commands will show up in the "Other Commands" section
	RootCmd.AddCommand(completionCmd)
	templates.ActsAsRootCommand(RootCmd, []string{"options"}, groups...)

	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		exit.WithError("Unable to bind flags", err)
	}
	cobra.OnInitialize(initConfig)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configPath := localpath.ConfigFile()
	viper.SetConfigFile(configPath)
	viper.SetConfigType("json")
	if err := viper.ReadInConfig(); err != nil {
		// This config file is optional, so don't emit errors if missing
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			glog.Warningf("Error reading config file at %s: %v", configPath, err)
		}
	}
	setupViper()
}

func setupViper() {
	viper.SetEnvPrefix(minikubeEnvPrefix)
	// Replaces '-' in flags with '_' in env variables
	// e.g. iso-url => $ENVPREFIX_ISO_URL
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetDefault(config.WantUpdateNotification, true)
	viper.SetDefault(config.ReminderWaitPeriodInHours, 24)
	viper.SetDefault(config.WantReportError, false)
	viper.SetDefault(config.WantReportErrorPrompt, true)
	viper.SetDefault(config.WantKubectlDownloadMsg, true)
	viper.SetDefault(config.WantNoneDriverWarning, true)
	viper.SetDefault(config.ShowDriverDeprecationNotification, true)
	viper.SetDefault(config.ShowBootstrapperDeprecationNotification, true)
	setFlagsUsingViper()
}

func addToPath(dir string) {
	new := fmt.Sprintf("%s:%s", dir, os.Getenv("PATH"))
	glog.Infof("Updating PATH: %s", dir)
	os.Setenv("PATH", new)
}
