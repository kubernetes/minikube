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
	"k8s.io/minikube/pkg/minikube/profile"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
)

var deleteAll bool
var purge bool

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a local kubernetes cluster",
	Long: `Deletes a local kubernetes cluster. This command deletes the VM, and removes all
associated files.`,
	Run: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Set flag to delete all profiles")
	deleteCmd.Flags().BoolVar(&purge, "purge", false, "Set this flag to delete the '.minikube' folder from your user directory.")

	if err := viper.BindPFlags(deleteCmd.Flags()); err != nil {
		exit.WithError("unable to bind flags", err)
	}
	RootCmd.AddCommand(deleteCmd)
}

// runDelete handles the executes the flow of "minikube delete"
func runDelete(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		exit.UsageT("Usage: minikube delete")
	}
	profileFlag, err := cmd.Flags().GetString("profile")
	if err != nil {
		exit.WithError("Could not get profile flag", err)
	}

	validProfiles, invalidProfiles, err := config.ListProfiles()
	profilesToDelete := append(validProfiles, invalidProfiles...)

	// If the purge flag is set, go ahead and delete the .minikube directory.
	if purge && len(profilesToDelete) > 1 && !deleteAll {
		out.ErrT(out.Notice, "Multiple minikube profiles were found - ")
		for _, p := range profilesToDelete {
			out.T(out.Notice, "    - {{.profile}}", out.V{"profile": p.Name})
		}
		exit.UsageT("Usage: minikube delete --all --purge")
	}

	if deleteAll {
		if profileFlag != constants.DefaultMachineName {
			exit.UsageT("usage: minikube delete --all")
		}

		if err != nil {
			exit.WithError("Error getting profiles to delete", err)
		}

		errs := profile.DeleteAll(profilesToDelete)
		if len(errs) > 0 {
			profile.HandleDeletionErrors(errs)
		} else {
			out.T(out.DeletingHost, "Successfully deleted all profiles")
		}
	} else {
		if len(args) > 0 {
			exit.UsageT("usage: minikube delete")
		}

		profileName := viper.GetString(config.MachineProfile)
		p, err := config.LoadProfile(profileName)
		if err != nil {
			out.ErrT(out.Meh, `"{{.name}}" p does not exist`, out.V{"name": profileName})
		}

		errs := profile.DeleteAll([]*config.Profile{p})
		if len(errs) > 0 {
			profile.HandleDeletionErrors(errs)
		} else {
			out.T(out.DeletingHost, "Successfully deleted p \"{{.name}}\"", out.V{"name": profileName})
		}
	}

	// If the purge flag is set, go ahead and delete the .minikube directory.
	if purge {
		glog.Infof("Purging the '.minikube' directory located at %s", localpath.MiniPath())
		if err := os.RemoveAll(localpath.MiniPath()); err != nil {
			exit.WithError("unable to delete minikube config folder", err)
		}
		out.T(out.Crushed, "Successfully purged minikube directory located at - [{{.minikubeDirectory}}]", out.V{"minikubeDirectory": localpath.MiniPath()})
	}
}
