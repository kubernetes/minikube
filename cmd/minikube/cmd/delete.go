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
	"github.com/golang/glog"
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
)

type DeletionError struct {
	Err     error
	Errtype typeOfError
}

func (error DeletionError) Error() string {
	return error.Err.Error()
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

		errs := DeleteProfiles(profilesToDelete)
		if len(errs) > 0 {
			HandleDeletionErrors(errs)
		} else {
			out.T(out.DeletingHost, "Successfully deleted all profiles")
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

		errs := DeleteProfiles([]*pkg_config.Profile{profile})
		if len(errs) > 0 {
			HandleDeletionErrors(errs)
		} else {
			out.T(out.DeletingHost, "Successfully deleted profile \"{{.name}}\"", out.V{"name": profileName})
		}
	}
}

func DeleteProfiles(profiles []*pkg_config.Profile) []error {
	var errs []error
	for _, profile := range profiles {
		err := deleteProfile(profile)

		var mm *cluster.Machine
		var loadErr error

		if err != nil {
			mm, loadErr = cluster.LoadMachine(profile.Name)
		}

		if (err != nil && !profile.IsValid()) || (loadErr != nil || !mm.IsValid()) {
			invalidProfileDeletionErrs := deleteInvalidProfile(profile)
			if len(invalidProfileDeletionErrs) > 0 {
				errs = append(errs, invalidProfileDeletionErrs...)
			}
		} else if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func deleteProfile(profile *pkg_config.Profile) error {
	viper.Set(pkg_config.MachineProfile, profile.Name)

	api, err := machine.NewAPIClient()
	if err != nil {
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("error getting client %v", err))
		return DeletionError{Err: delErr, Errtype: Fatal}
	}
	defer api.Close()

	cc, err := pkg_config.Load()
	if err != nil && !os.IsNotExist(err) {
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("error loading profile config: %v", err))
		return DeletionError{Err: delErr, Errtype: MissingProfile}
	}

	// In the case of "none", we want to uninstall Kubernetes as there is no VM to delete
	if err == nil && cc.MachineConfig.VMDriver == constants.DriverNone {
		if err := uninstallKubernetes(api, cc.KubernetesConfig, viper.GetString(cmdcfg.Bootstrapper)); err != nil {
			deletionError, ok := err.(DeletionError)
			if ok {
				delErr := profileDeletionErr(profile.Name, fmt.Sprintf("%v", err))
				deletionError.Err = delErr
				return deletionError
			}
			return err
		}
	}

	if err = cluster.DeleteHost(api); err != nil {
		switch err := errors.Cause(err).(type) {
		case mcnerror.ErrHostDoesNotExist:
			delErr := profileDeletionErr(profile.Name, fmt.Sprintf("\"%s\" cluster does not exist", profile.Name))
			return DeletionError{Err: delErr, Errtype: MissingCluster}
		default:
			delErr := profileDeletionErr(profile.Name, fmt.Sprintf("failed to delete cluster %v", err))
			return DeletionError{Err: delErr, Errtype: Fatal}
		}
	}

	if err := cmdUtil.KillMountProcess(); err != nil {
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("failed to kill mount process: %v", err))
		return DeletionError{Err: delErr, Errtype: Fatal}
	}

	if err := os.RemoveAll(constants.GetProfilePath(viper.GetString(pkg_config.MachineProfile))); err != nil {
		if os.IsNotExist(err) {
			delErr := profileDeletionErr(profile.Name, fmt.Sprintf("\"%s\" profile does not exist", profile.Name))
			return DeletionError{Err: delErr, Errtype: MissingProfile}
		}
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("failed to remove profile %v", err))
		return DeletionError{Err: delErr, Errtype: Fatal}
	}
	out.T(out.Crushed, `The "{{.cluster_name}}" cluster has been deleted.`, out.V{"cluster_name": profile.Name})

	machineName := pkg_config.GetMachineName()
	if err := pkgutil.DeleteKubeConfigContext(constants.KubeconfigPath, machineName); err != nil {
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("update config %v", err))
		return DeletionError{Err: delErr, Errtype: Fatal}
	}
	return nil
}

func deleteInvalidProfile(profile *pkg_config.Profile) []error {
	out.T(out.DeletingHost, "Trying to delete invalid profile {{.profile}}", out.V{"profile": profile.Name})

	var errs []error
	pathToProfile := constants.GetProfilePath(profile.Name, constants.GetMinipath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		err := os.RemoveAll(pathToProfile)
		if err != nil {
			errs = append(errs, DeletionError{err, Fatal})
		}
	}

	pathToMachine := constants.GetMachinePath(profile.Name, constants.GetMinipath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		err := os.RemoveAll(pathToMachine)
		if err != nil {
			errs = append(errs, DeletionError{err, Fatal})
		}
	}
	return errs
}

func profileDeletionErr(profileName string, additionalInfo string) error {
	return fmt.Errorf("error deleting profile \"%s\": %s", profileName, additionalInfo)
}

func uninstallKubernetes(api libmachine.API, kc pkg_config.KubernetesConfig, bsName string) error {
	out.T(out.Resetting, "Uninstalling Kubernetes {{.kubernetes_version}} using {{.bootstrapper_name}} ...", out.V{"kubernetes_version": kc.KubernetesVersion, "bootstrapper_name": bsName})
	clusterBootstrapper, err := getClusterBootstrapper(api, bsName)
	if err != nil {
		return DeletionError{Err: fmt.Errorf("unable to get bootstrapper: %v", err), Errtype: Fatal}
	} else if err = clusterBootstrapper.DeleteCluster(kc); err != nil {
		return DeletionError{Err: fmt.Errorf("failed to delete cluster: %v", err), Errtype: Fatal}
	}
	return nil
}

func HandleDeletionErrors(errors []error) {
	if len(errors) == 1 {
		handleSingleDeletionError(errors[0])
	} else {
		handleMultipleDeletionErrors(errors)
	}
}

func handleSingleDeletionError(err error) {
	deletionError, ok := err.(DeletionError)

	if ok {
		switch deletionError.Errtype {
		case Fatal:
			out.FatalT(deletionError.Error())
		case MissingProfile:
			out.ErrT(out.Sad, deletionError.Error())
		case MissingCluster:
			out.ErrT(out.Meh, deletionError.Error())
		default:
			out.FatalT(deletionError.Error())
		}
	} else {
		exit.WithError("Could not process error from failed deletion", err)
	}
}

func handleMultipleDeletionErrors(errors []error) {
	out.ErrT(out.Sad, "Multiple errors deleting profiles")

	for _, err := range errors {
		deletionError, ok := err.(DeletionError)

		if ok {
			glog.Errorln(deletionError.Error())
		} else {
			exit.WithError("Could not process errors from failed deletion", err)
		}
	}
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Set flag to delete all profiles")
	RootCmd.AddCommand(deleteCmd)
}
