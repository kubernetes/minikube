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
	"fmt"
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	pkg_config "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	pkgutil "k8s.io/minikube/pkg/util"
)

var deleteAll bool

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a local kubernetes cluster",
	Long: `Deletes a local kubernetes cluster. This command deletes the VM, and removes all
associated files.`,
	Run: runDelete,
}

type typeOfError int

const (
	Fatal          typeOfError = 0
	MissingProfile typeOfError = 1
	MissingCluster typeOfError = 2
	Usage          typeOfError = 3
)

type deletionError struct {
	err     error
	errtype typeOfError
}

func (error deletionError) Error() string {
	return error.err.Error()
}

// runDelete handles the executes the flow of "minikube delete"
func runDelete(cmd *cobra.Command, args []string) {
	profileFlag, err := cmd.Flags().GetString("profile")
	if err != nil {
		exit.WithError("Could not get profile flag", err)
	}

	if deleteAll {
		if profileFlag != constants.DefaultMachineName {
			exit.UsageT("usage: minikube delete --all")
		}

		validProfiles, invalidProfiles, err := pkg_config.ListProfiles()
		profilesToDelete := append(validProfiles, invalidProfiles...)

		if err != nil {
			exit.WithError("Error getting profiles to delete", err)
		}

		errs := deleteAllProfiles(profilesToDelete)
		if len(errs) > 0 {
			handleDeletionErrors(errs)
		}
	} else {
		if len(args) > 0 {
			exit.UsageT("usage: minikube delete")
		}

		profileName := viper.GetString(pkg_config.MachineProfile)
		profile, err := pkg_config.LoadProfile(profileName)
		if err != nil {
			out.ErrT(out.Meh, `"{{.name}}" profile does not exist`, out.V{"name": profileName})
		}

		err = deleteProfile(profile)
		if err != nil {
			handleDeletionErrors([]error{err})
		}
	}
}

func handleDeletionErrors(errors []error) {
	for _, err := range errors {
		deletionError, ok := err.(deletionError)
		if ok {
			switch deletionError.errtype {
			case Fatal:
				out.FatalT(deletionError.Error())
			case MissingProfile:
				out.ErrT(out.Sad, deletionError.Error())
			case MissingCluster:
				out.ErrT(out.Meh, deletionError.Error())
			case Usage:
				out.ErrT(out.Usage, "usage: minikube delete or minikube delete -p foo or minikube delete --all")
			default:
				out.FatalT(deletionError.Error())
			}
		} else {
			exit.WithError("Could not process errors from failed deletion", err)
		}
	}
}

func deleteAllProfiles(profiles []*pkg_config.Profile) []error {
	var errs []error
	for _, profile := range profiles {
		err := deleteProfile(profile)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func deleteProfile(profile *pkg_config.Profile) error {
	viper.Set(pkg_config.MachineProfile, profile.Name)

	api, err := machine.NewAPIClient()
	if err != nil {
		return deletionError{fmt.Errorf("error deleting profile \"%s\": error getting client %v", profile.Name, err), Fatal}
	}
	defer api.Close()

	cc, err := pkg_config.Load()
	if err != nil && !os.IsNotExist(err) {
		return deletionError{fmt.Errorf("error deleting profile \"%s\": error loading profile config: %s", profile.Name, profile.Name), Usage}
	}

	// In the case of "none", we want to uninstall Kubernetes as there is no VM to delete
	if err == nil && cc.MachineConfig.VMDriver == constants.DriverNone {
		if err := uninstallKubernetes(api, cc.KubernetesConfig, viper.GetString(cmdcfg.Bootstrapper)); err != nil {
			deletionError, ok := err.(deletionError)
			if ok {
				deletionError.err = fmt.Errorf("error deleting profile \"%s\": %v", profile.Name, err)
				return deletionError
			}
			return err
		}
	}

	if err = cluster.DeleteHost(api); err != nil {
		switch err := errors.Cause(err).(type) {
		case mcnerror.ErrHostDoesNotExist:
			return deletionError{fmt.Errorf("error deleting profile \"%s\": \"%s\" cluster does not exist", profile.Name, profile.Name), MissingCluster}
		default:
			return deletionError{fmt.Errorf("error deleting profile \"%s\": failed to delete cluster %v", profile.Name, err), Fatal}
		}
	}

	if err := cmdUtil.KillMountProcess(); err != nil {
		return deletionError{fmt.Errorf("error deleting profile \"%s\": failed to kill mount process: %v", profile.Name, err), Fatal}
	}

	if err := os.RemoveAll(constants.GetProfilePath(viper.GetString(pkg_config.MachineProfile))); err != nil {
		if os.IsNotExist(err) {
			return deletionError{fmt.Errorf("error deleting profile \"%s\": %s profile does not exist", profile.Name, profile.Name), MissingProfile}
		}
		return deletionError{fmt.Errorf("error deleting profile \"%s\": failed to remove profile %v", profile.Name, err), Fatal}
	}
	out.T(out.Crushed, `The "{{.cluster_name}}" cluster has been deleted.`, out.V{"cluster_name": profile.Name})

	machineName := pkg_config.GetMachineName()
	if err := pkgutil.DeleteKubeConfigContext(constants.KubeconfigPath, machineName); err != nil {
		return deletionError{fmt.Errorf("error deleting profile \"%s\": update config %v", profile.Name, err), Fatal}
	}
	return nil
}

func uninstallKubernetes(api libmachine.API, kc pkg_config.KubernetesConfig, bsName string) error {
	out.T(out.Resetting, "Uninstalling Kubernetes {{.kubernetes_version}} using {{.bootstrapper_name}} ...", out.V{"kubernetes_version": kc.KubernetesVersion, "bootstrapper_name": bsName})
	clusterBootstrapper, err := getClusterBootstrapper(api, bsName)
	if err != nil {
		return deletionError{fmt.Errorf("unable to get bootstrapper: %v", err), Fatal}
	} else if err = clusterBootstrapper.DeleteCluster(kc); err != nil {
		return deletionError{fmt.Errorf("failed to delete cluster: %v", err), Fatal}
	}
	return nil
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Set flag to delete all profiles")
	RootCmd.AddCommand(deleteCmd)
}
