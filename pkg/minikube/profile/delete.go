/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
)

type typeOfError int

// DeletionError can be returned from DeleteAll
type DeletionError struct {
	Err       error
	ErrorType typeOfError
}

func (error DeletionError) Error() string {
	return error.Err.Error()
}

const (
	// Fatal is a type of DeletionError
	Fatal typeOfError = 0
	// MissingProfile is a type of DeletionError
	MissingProfile typeOfError = 1
	// MissingCluster is a type of DeletionError
	MissingCluster typeOfError = 2
)

// DeleteAll deletes one or more profiles
func DeleteAll(profiles []*config.Profile) []error {
	var errs []error
	for _, profile := range profiles {
		err := delete(profile)

		if err != nil {
			mm, loadErr := cluster.LoadMachine(profile.Name)

			if !profile.IsValid() || (loadErr != nil || !mm.IsValid()) {
				invalidProfileDeletionErrs := DeleteInvalid(profile)
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

func delete(profile *config.Profile) error {
	viper.Set(config.MachineProfile, profile.Name)

	api, err := machine.NewAPIClient()
	if err != nil {
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("error getting client %v", err))
		return DeletionError{Err: delErr, ErrorType: Fatal}
	}
	defer api.Close()

	cc, err := config.Load()
	if err != nil && !os.IsNotExist(err) {
		out.ErrT(out.Sad, "Error loading profile {{.name}}: {{.error}}", out.V{"name": profile, "error": err})
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("error loading profile config: %v", err))
		return DeletionError{Err: delErr, ErrorType: MissingProfile}
	}

	if err == nil && driver.BareMetal(cc.MachineConfig.VMDriver) {
		if err := UninstallKubernetes(api, cc.KubernetesConfig, viper.GetString(cmdcfg.Bootstrapper)); err != nil {
			deletionError, ok := err.(DeletionError)
			if ok {
				delErr := profileDeletionErr(profile.Name, fmt.Sprintf("%v", err))
				deletionError.Err = delErr
				return deletionError
			}
			return err
		}
	}

	if err := cluster.KillMountProcess(); err != nil {
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
	DeleteDirectoryOfProfile(profile.Name)

	if err := config.DeleteProfileDirectory(profile.Name); err != nil {
		if os.IsNotExist(err) {
			delErr := profileDeletionErr(profile.Name, fmt.Sprintf("\"%s\" profile does not exist", profile.Name))
			return DeletionError{Err: delErr, ErrorType: MissingProfile}
		}
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("failed to remove profile %v", err))
		return DeletionError{Err: delErr, ErrorType: Fatal}
	}

	out.T(out.Crushed, `The "{{.name}}" cluster has been deleted.`, out.V{"name": profile.Name})

	machineName := config.GetMachineName()
	if err := kubeconfig.DeleteContext(constants.KubeconfigPath, machineName); err != nil {
		return DeletionError{Err: fmt.Errorf("update config: %v", err), ErrorType: Fatal}
	}

	if err := cmdcfg.Unset(config.MachineProfile); err != nil {
		return DeletionError{Err: fmt.Errorf("unset minikube profile: %v", err), ErrorType: Fatal}
	}
	return nil
}

func DeleteInvalid(profile *config.Profile) []error {
	out.T(out.DeletingHost, "Trying to delete invalid profile {{.profile}}", out.V{"profile": profile.Name})

	var errs []error
	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		err := os.RemoveAll(pathToProfile)
		if err != nil {
			errs = append(errs, DeletionError{err, Fatal})
		}
	}

	pathToMachine := cluster.MachinePath(profile.Name, localpath.MiniPath())
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

// Handles deletion error from DeleteAll
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
		switch deletionError.ErrorType {
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

func DeleteDirectoryOfProfile(profile string) {
	machineDir := filepath.Join(localpath.MiniPath(), "machines", profile)
	if _, err := os.Stat(machineDir); err == nil {
		out.T(out.DeletingHost, `Removing {{.directory}} ...`, out.V{"directory": machineDir})
		err := os.RemoveAll(machineDir)
		if err != nil {
			exit.WithError("Unable to remove machine directory: %v", err)
		}
	}
}

func UninstallKubernetes(api libmachine.API, kc config.KubernetesConfig, bsName string) error {
	out.T(out.Resetting, "Uninstalling Kubernetes {{.kubernetes_version}} using {{.bootstrapper_name}} ...", out.V{"kubernetes_version": kc.KubernetesVersion, "bootstrapper_name": bsName})
	clusterBootstrapper, err := cluster.GetClusterBootstrapper(api, bsName)
	if err != nil {
		return DeletionError{Err: fmt.Errorf("unable to get bootstrapper: %v", err), ErrorType: Fatal}
	} else if err = clusterBootstrapper.DeleteCluster(kc); err != nil {
		return DeletionError{Err: fmt.Errorf("failed to delete cluster: %v", err), ErrorType: Fatal}
	}
	return nil
}
