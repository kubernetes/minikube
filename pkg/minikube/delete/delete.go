package delete

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
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

// killMountProcess kills the mount process, if it is running
func KillMountProcess() error {
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
