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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/golang/glog"
	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/delete"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
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

// runDelete handles the executes the flow of "minikube delete"
func runDelete(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		exit.UsageT("Usage: minikube delete")
	}
	profileFlag, err := cmd.Flags().GetString("profile")
	if err != nil {
		exit.WithError("Could not get profile flag", err)
	}

	if deleteAll {
		if profileFlag != constants.DefaultMachineName {
			exit.UsageT("usage: minikube delete --all")
		}

		validProfiles, invalidProfiles, err := config.ListProfiles()
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

		profileName := viper.GetString(config.MachineProfile)
		profile, err := config.LoadProfile(profileName)
		if err != nil {
			out.ErrT(out.Meh, `"{{.name}}" profile does not exist`, out.V{"name": profileName})
		}

		errs := DeleteProfiles([]*config.Profile{profile})
		if len(errs) > 0 {
			HandleDeletionErrors(errs)
		} else {
			out.T(out.DeletingHost, "Successfully deleted profile \"{{.name}}\"", out.V{"name": profileName})
		}
	}
}

// Deletes one or more profiles
func DeleteProfiles(profiles []*config.Profile) []error {
	var errs []error
	for _, profile := range profiles {
		err := deleteProfile(profile)

		if err != nil {
			mm, loadErr := cluster.LoadMachine(profile.Name)

			if !profile.IsValid() || (loadErr != nil || !mm.IsValid()) {
				invalidProfileDeletionErrs := delete.DeleteInvalidProfile(profile)
				if len(invalidProfileDeletionErrs) > 0 {
					errs = append(errs, invalidProfileDeletionErrs...)
				}
			} else {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

func deleteProfile(profile *config.Profile) error {
	viper.Set(config.MachineProfile, profile.Name)

	api, err := machine.NewAPIClient()
	if err != nil {
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("error getting client %v", err))
		return delete.DeletionError{Err: delErr, ErrorType: delete.Fatal}
	}
	defer api.Close()

	cc, err := config.Load()
	if err != nil && !os.IsNotExist(err) {
		out.ErrT(out.Sad, "Error loading profile {{.name}}: {{.error}}", out.V{"name": profile, "error": err})
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("error loading profile config: %v", err))
		return delete.DeletionError{Err: delErr, ErrorType: delete.MissingProfile}
	}

	// In the case of "none", we want to uninstall Kubernetes as there is no VM to delete
	if err == nil && cc.MachineConfig.VMDriver == constants.DriverNone {
		if err := delete.UninstallKubernetes(api, cc.KubernetesConfig, viper.GetString(cmdcfg.Bootstrapper)); err != nil {
			deletionError, ok := err.(delete.DeletionError)
			if ok {
				delErr := profileDeletionErr(profile.Name, fmt.Sprintf("%v", err))
				deletionError.Err = delErr
				return deletionError
			}
			return err
		}
	}

	if err := killMountProcess(); err != nil {
		out.T(out.FailureType, "Failed to kill mount process: {{.error}}", out.V{"error": err})
	}

	if err = cluster.DeleteHost(api); err != nil {
		switch errors.Cause(err).(type) {
		case mcnerror.ErrHostDoesNotExist:
			out.T(out.Meh, `"{{.name}}" cluster does not exist. Proceeding ahead with cleanup.`, out.V{"name": profile.Name})
		default:
			out.T(out.FailureType, "Failed to delete cluster: {{.error}}", out.V{"error": err})
			out.T(out.Notice, `You may need to manually remove the "{{.name}}" VM from your hypervisor`, out.V{"name": profile.Name})
		}
	}

	// In case DeleteHost didn't complete the job.
	delete.DeleteProfileDirectory(profile.Name)

	if err := config.DeleteProfile(profile.Name); err != nil {
		if os.IsNotExist(err) {
			delErr := profileDeletionErr(profile.Name, fmt.Sprintf("\"%s\" profile does not exist", profile.Name))
			return delete.DeletionError{Err: delErr, ErrorType: delete.MissingProfile}
		}
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("failed to remove profile %v", err))
		return delete.DeletionError{Err: delErr, ErrorType: delete.Fatal}
	}

	out.T(out.Crushed, `The "{{.name}}" cluster has been deleted.`, out.V{"name": profile.Name})

	machineName := config.GetMachineName()
	if err := kubeconfig.DeleteContext(constants.KubeconfigPath, machineName); err != nil {
		return delete.DeletionError{Err: fmt.Errorf("update config: %v", err), ErrorType: delete.Fatal}
	}

	if err := cmdcfg.Unset(config.MachineProfile); err != nil {
		return delete.DeletionError{Err: fmt.Errorf("unset minikube profile: %v", err), ErrorType: delete.Fatal}
	}
	return nil
}

func profileDeletionErr(profileName string, additionalInfo string) error {
	return fmt.Errorf("error deleting profile \"%s\": %s", profileName, additionalInfo)
}

// Handles deletion error from DeleteProfiles
func HandleDeletionErrors(errors []error) {
	if len(errors) == 1 {
		handleSingleDeletionError(errors[0])
	} else {
		handleMultipleDeletionErrors(errors)
	}
}

func handleSingleDeletionError(err error) {
	deletionError, ok := err.(delete.DeletionError)

	if ok {
		switch deletionError.ErrorType {
		case delete.Fatal:
			out.FatalT(deletionError.Error())
		case delete.MissingProfile:
			out.ErrT(out.Sad, deletionError.Error())
		case delete.MissingCluster:
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
		deletionError, ok := err.(delete.DeletionError)

		if ok {
			glog.Errorln(deletionError.Error())
		} else {
			exit.WithError("Could not process errors from failed deletion", err)
		}
	}
}

// killMountProcess kills the mount process, if it is running
func killMountProcess() error {
	pidPath := filepath.Join(localpath.MiniPath(), constants.MountProcessFileName)
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		return nil
	}

	glog.Infof("Found %s ...", pidPath)
	out, err := ioutil.ReadFile(pidPath)
	if err != nil {
		return errors.Wrap(err, "ReadFile")
	}
	glog.Infof("pidfile contents: %s", out)
	pid, err := strconv.Atoi(string(out))
	if err != nil {
		return errors.Wrap(err, "error parsing pid")
	}
	// os.FindProcess does not check if pid is running :(
	entry, err := ps.FindProcess(pid)
	if err != nil {
		return errors.Wrap(err, "ps.FindProcess")
	}
	if entry == nil {
		glog.Infof("Stale pid: %d", pid)
		if err := os.Remove(pidPath); err != nil {
			return errors.Wrap(err, "Removing stale pid")
		}
		return nil
	}

	// We found a process, but it still may not be ours.
	glog.Infof("Found process %d: %s", pid, entry.Executable())
	proc, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrap(err, "os.FindProcess")
	}

	glog.Infof("Killing pid %d ...", pid)
	if err := proc.Kill(); err != nil {
		glog.Infof("Kill failed with %v - removing probably stale pid...", err)
		if err := os.Remove(pidPath); err != nil {
			return errors.Wrap(err, "Removing likely stale unkillable pid")
		}
		return errors.Wrap(err, fmt.Sprintf("Kill(%d/%s)", pid, entry.Executable()))
	}
	return nil
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Set flag to delete all profiles")
	RootCmd.AddCommand(deleteCmd)
}
