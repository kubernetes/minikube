package delete

import (
	"fmt"
	"github.com/docker/machine/libmachine"
	"k8s.io/minikube/pkg/minikube/config"
	"os"
	"path/filepath"

	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
)

type typeOfError int

type DeletionError struct {
	Err       error
	ErrorType typeOfError
}

func (error DeletionError) Error() string {
	return error.Err.Error()
}

const (
	Fatal          typeOfError = 0
	MissingProfile typeOfError = 1
	MissingCluster typeOfError = 2
)

func DeleteProfileDirectory(profile string) {
	machineDir := filepath.Join(localpath.MiniPath(), "machines", profile)
	if _, err := os.Stat(machineDir); err == nil {
		out.T(out.DeletingHost, `Removing {{.directory}} ...`, out.V{"directory": machineDir})
		err := os.RemoveAll(machineDir)
		if err != nil {
			exit.WithError("Unable to remove machine directory: %v", err)
		}
	}
}

func DeleteInvalidProfile(profile *config.Profile) []error {
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
