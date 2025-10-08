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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/templates"
	configCmd "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/audit"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/translate"
	"k8s.io/minikube/pkg/version"
)

var (
	dirs = [...]string{
		localpath.MiniPath(),
		localpath.MakeMiniPath("certs"),
		localpath.MakeMiniPath("machines"),
		localpath.MakeMiniPath("cache"),
		localpath.MakeMiniPath("config"),
		localpath.MakeMiniPath("addons"),
		localpath.MakeMiniPath("files"),
		localpath.MakeMiniPath("logs"),
	}
	auditID string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "minikube",
	Short: "minikube quickly sets up a local Kubernetes cluster",
	Long:  `minikube provisions and manages local Kubernetes clusters optimized for development workflows.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		for _, path := range dirs {
			if err := os.MkdirAll(path, 0777); err != nil {
				exit.Error(reason.HostHomeMkdir, "Error creating minikube directory", err)
			}
		}
		userName := viper.GetString(config.UserFlag)
		if !validateUsername(userName) {
			out.WarningT("User name '{{.username}}' is not valid", out.V{"username": userName})
			exit.Message(reason.Usage, "User name must be 60 chars or less.")
		}
		var err error
		auditID, err = audit.LogCommandStart()
		if err != nil {
			klog.Warningf("failed to log command start to audit: %v", err)
		}
		// viper maps $MINIKUBE_ROOTLESS to "rootless" property automatically, but it does not do vice versa,
		// so we map "rootless" property to $MINIKUBE_ROOTLESS expliclity here.
		// $MINIKUBE_ROOTLESS is referred by KIC runner, which is decoupled from viper.
		if viper.GetBool(config.Rootless) {
			os.Setenv(constants.MinikubeRootlessEnv, "true")
		}
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		if err := audit.LogCommandEnd(auditID); err != nil {
			klog.Warningf("failed to log command end to audit: %v", err)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Check whether this is a windows binary (.exe) running inisde WSL.
	if runtime.GOOS == "windows" && detect.IsMicrosoftWSL() {
		if !slices.Contains(os.Args, "--force") {
			exit.Message(reason.WrongBinaryWSL, "You are trying to run a windows .exe binary inside WSL. For better integration please use a Linux binary instead (Download at https://minikube.sigs.k8s.io/docs/start/.). Otherwise if you still want to do this, you can do it using --force")
		}
	}

	if runtime.GOOS == "darwin" && detect.IsAmd64M1Emulation() {
		out.Boxed("You are trying to run the amd64 binary on an M1 system.\nPlease consider running the darwin/arm64 binary instead.\nDownload at {{.url}}",
			out.V{"url": notify.DownloadURL(version.GetVersion(), "darwin", "arm64")})
	}

	_, callingCmd := filepath.Split(os.Args[0])
	callingCmd = strings.TrimSuffix(callingCmd, ".exe")

	if callingCmd == "kubectl" {
		// If the user is using the minikube binary as kubectl, allow them to specify the kubectl context without also specifying minikube profile
		profile := ""
		for i, a := range os.Args {
			if a == "--context" {
				if len(os.Args) > i+1 {
					profile = fmt.Sprintf("--profile=%s", os.Args[i+1])
				}
				break
			} else if strings.HasPrefix(a, "--context=") {
				context := strings.Split(a, "=")[1]
				profile = fmt.Sprintf("--profile=%s", context)
				break
			}
		}
		if profile != "" {
			os.Args = append([]string{RootCmd.Use, callingCmd, profile, "--"}, os.Args[1:]...)
		} else {
			os.Args = append([]string{RootCmd.Use, callingCmd, "--"}, os.Args[1:]...)
		}
	}

	applyToAllCommands(RootCmd, func(c *cobra.Command) {
		c.Short = translate.T(c.Short)
		c.Long = translate.T(c.Long)
		c.Flags().VisitAll(func(f *pflag.Flag) {
			f.Usage = translate.T(f.Usage)
		})

		c.SetUsageTemplate(usageTemplate())
	})

	RootCmd.Short = translate.T(RootCmd.Short)
	RootCmd.Long = translate.T(RootCmd.Long)
	RootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Usage = translate.T(f.Usage)
	})

	if runtime.GOOS != "windows" {
		// add minikube binaries to the path
		targetDir := localpath.MakeMiniPath("bin")
		addToPath(targetDir)
	}

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

func init() {
	klog.InitFlags(nil)
	// preset logtostderr and alsologtostderr only for test runs, for normal runs consider flags in main()
	if strings.HasPrefix(filepath.Base(os.Args[0]), "e2e-") || strings.HasSuffix(os.Args[0], "test") {
		if err := flag.Set("logtostderr", "false"); err != nil {
			klog.Warningf("Unable to set default flag value for logtostderr: %v", err)
		}
		if err := flag.Set("alsologtostderr", "false"); err != nil {
			klog.Warningf("Unable to set default flag value for alsologtostderr: %v", err)
		}
	}
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine) // avoid `generate-docs_test.go` complaining about "Docs are not updated"

	RootCmd.PersistentFlags().StringP(config.ProfileName, "p", constants.DefaultClusterName, `The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently.`)
	RootCmd.PersistentFlags().StringP(configCmd.Bootstrapper, "b", "kubeadm", "The name of the cluster bootstrapper that will set up the Kubernetes cluster.")
	RootCmd.PersistentFlags().String(config.UserFlag, "", "Specifies the user executing the operation. Useful for auditing operations executed by 3rd party tools. Defaults to the operating system username.")
	RootCmd.PersistentFlags().Bool(config.SkipAuditFlag, false, "Skip recording the current command in the audit logs.")
	RootCmd.PersistentFlags().Bool(config.Rootless, false, "Force to use rootless driver (docker and podman driver only)")

	translate.DetermineLocale()

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
				imageCmd,
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
				cpCmd,
			},
		},
		{
			Message: translate.T("Troubleshooting Commands:"),
			Commands: []*cobra.Command{
				sshKeyCmd,
				sshHostCmd,
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
	RootCmd.AddCommand(licenseCmd)
	templates.ActsAsRootCommand(RootCmd, []string{"options"}, groups...)

	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		exit.Error(reason.InternalBindFlags, "Unable to bind flags", err)
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
			klog.Warningf("Error reading config file at %s: %v", configPath, err)
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

	viper.RegisterAlias(config.EmbedCerts, embedCerts)
	viper.SetDefault(config.WantUpdateNotification, true)
	viper.SetDefault(config.ReminderWaitPeriodInHours, 24)
	viper.SetDefault(config.WantNoneDriverWarning, true)
	viper.SetDefault(config.WantVirtualBoxDriverWarning, true)
	viper.SetDefault(config.MaxAuditEntries, 1000)
	viper.SetDefault(config.SkipAuditFlag, false)
}

func addToPath(dir string) {
	path := fmt.Sprintf("%s:%s", dir, os.Getenv("PATH"))
	klog.Infof("Updating PATH: %s", dir)
	os.Setenv("PATH", path)
}

func validateUsername(name string) bool {
	return len(name) <= 60
}

// applyToAllCommands applies the provided func to all commands including sub commands
func applyToAllCommands(cmd *cobra.Command, f func(subCmd *cobra.Command)) {
	for _, c := range cmd.Commands() {
		f(c)
		if c.HasSubCommands() {
			applyToAllCommands(c, f)
		}
	}
}
