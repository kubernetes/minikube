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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"

	"github.com/docker/machine/libmachine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	pkgProfile "k8s.io/minikube/pkg/minikube/profile"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var (
	deleteAll bool
	purge     bool
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a local Kubernetes cluster",
	Long: `Deletes a local Kubernetes cluster. This command deletes the VM, and removes all
associated files.`,
	Run: runDelete,
}

type typeOfError int

const (
	// Fatal is a type of DeletionError
	Fatal typeOfError = 0
	// MissingProfile is a type of DeletionError
	MissingProfile typeOfError = 1
	// MissingCluster is a type of DeletionError
	MissingCluster typeOfError = 2
)

// DeletionError can be returned from DeleteProfiles
type DeletionError struct {
	Err     error
	Errtype typeOfError
}

func (error DeletionError) Error() string {
	return error.Err.Error()
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Set flag to delete all profiles")
	deleteCmd.Flags().BoolVar(&purge, "purge", false, "Set this flag to delete the '.minikube' folder from your user directory.")

	if err := viper.BindPFlags(deleteCmd.Flags()); err != nil {
		exit.Error(reason.InternalBindFlags, "unable to bind flags", err)
	}
}

// shotgun cleanup to delete orphaned docker container data
func deleteContainersAndVolumes(ctx context.Context, ociBin string) {
	if _, err := exec.LookPath(ociBin); err != nil {
		klog.Infof("skipping deleteContainersAndVolumes for %s: %v", ociBin, err)
		return
	}

	klog.Infof("deleting containers and volumes ...")

	delLabel := fmt.Sprintf("%s=%s", oci.CreatedByLabelKey, "true")
	errs := oci.DeleteContainersByLabel(ociBin, delLabel)
	if len(errs) > 0 { // it will error if there is no container to delete
		klog.Infof("error delete containers by label %q (might be okay): %+v", delLabel, errs)
	}

	errs = oci.DeleteAllVolumesByLabel(ctx, ociBin, delLabel)
	if len(errs) > 0 { // it will not error if there is nothing to delete
		klog.Warningf("error delete volumes by label %q (might be okay): %+v", delLabel, errs)
	}

	if ociBin == oci.Podman {
		// podman prune does not support --filter
		return
	}

	errs = oci.PruneAllVolumesByLabel(ctx, ociBin, delLabel)
	if len(errs) > 0 { // it will not error if there is nothing to delete
		klog.Warningf("error pruning volumes by label %q (might be okay): %+v", delLabel, errs)
	}
}

// runDelete handles the executes the flow of "minikube delete"
func runDelete(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		exit.Message(reason.Usage, "Usage: minikube delete")
	}
	// register.SetEventLogPath(localpath.EventLog(ClusterFlagValue()))
	register.Reg.SetStep(register.Deleting)

	validProfiles, invalidProfiles, err := config.ListProfiles()
	if err != nil {
		klog.Warningf("'error loading profiles in minikube home %q: %v", localpath.MiniPath(), err)
	}
	profilesToDelete := append(validProfiles, invalidProfiles...)
	// in the case user has more than 1 profile and runs --purge
	// to prevent abandoned VMs/containers, force user to run with delete --all
	if purge && len(profilesToDelete) > 1 && !deleteAll {
		out.ErrT(style.Notice, "Multiple minikube profiles were found - ")
		for _, p := range profilesToDelete {
			out.Step(style.Notice, "    - {{.profile}}", out.V{"profile": p.Name})
		}
		exit.Message(reason.Usage, "Usage: minikube delete --all --purge")
	}
	delCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if deleteAll {
		deleteContainersAndVolumes(delCtx, oci.Docker)
		deleteContainersAndVolumes(delCtx, oci.Podman)

		errs := DeleteProfiles(profilesToDelete)
		register.Reg.SetStep(register.Done)

		if len(errs) > 0 {
			HandleDeletionErrors(errs)
		} else {
			out.Step(style.DeletingHost, "Successfully deleted all profiles")
		}
	} else {
		if len(args) > 0 {
			exit.Message(reason.Usage, "usage: minikube delete")
		}

		cname := ClusterFlagValue()
		profile, err := config.LoadProfile(cname)
		orphan := false

		if err != nil {
			out.ErrT(style.Meh, `"{{.name}}" profile does not exist, trying anyways.`, out.V{"name": cname})
			orphan = true
		}

		errs := DeleteProfiles([]*config.Profile{profile})
		register.Reg.SetStep(register.Done)

		if len(errs) > 0 {
			HandleDeletionErrors(errs)
		}

		if orphan {
			pkgProfile.DeletePossibleLeftOvers(delCtx, cname, driver.Docker)
			pkgProfile.DeletePossibleLeftOvers(delCtx, cname, driver.Podman)
		}
	}

	// If the purge flag is set, go ahead and delete the .minikube directory.
	if purge {
		purgeMinikubeDirectory()
	}
}

func purgeMinikubeDirectory() {
	klog.Infof("Purging the '.minikube' directory located at %s", localpath.MiniPath())
	if err := os.RemoveAll(localpath.MiniPath()); err != nil {
		exit.Error(reason.HostPurge, "unable to delete minikube config folder", err)
	}
	out.Step(style.Deleted, "Successfully purged minikube directory located at - [{{.minikubeDirectory}}]", out.V{"minikubeDirectory": localpath.MiniPath()})
}

// DeleteProfiles deletes one or more profiles
func DeleteProfiles(profiles []*config.Profile) []error {
	klog.Infof("DeleteProfiles")
	var errs []error
	for _, profile := range profiles {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		err := deleteProfile(ctx, profile)
		if err != nil {
			mm, loadErr := machine.LoadMachine(profile.Name)

			if !profile.IsValid() || (loadErr != nil || !mm.IsValid()) {
				invalidProfileDeletionErrs := deleteInvalidProfile(profile)
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

func deleteProfile(ctx context.Context, profile *config.Profile) error {
	klog.Infof("Deleting %s", profile.Name)
	register.Reg.SetStep(register.Deleting)

	viper.Set(config.ProfileName, profile.Name)
	if profile.Config != nil {
		klog.Infof("%s configuration: %+v", profile.Name, profile.Config)

		// if driver is oci driver, delete containers and volumes
		if driver.IsKIC(profile.Config.Driver) {
			out.Step(style.DeletingHost, `Deleting "{{.profile_name}}" in {{.driver_name}} ...`, out.V{"profile_name": profile.Name, "driver_name": profile.Config.Driver})
			for _, n := range profile.Config.Nodes {
				machineName := config.MachineName(*profile.Config, n)
				pkgProfile.DeletePossibleLeftOvers(ctx, machineName, profile.Config.Driver)
			}
		}
	} else {
		klog.Infof("%s has no configuration, will try to make it work anyways", profile.Name)
	}

	api, err := machine.NewAPIClient()
	if err != nil {
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("error getting client %v", err))
		return DeletionError{Err: delErr, Errtype: Fatal}
	}
	defer api.Close()

	cc, err := config.Load(profile.Name)
	if err != nil && !config.IsNotExist(err) {
		delErr := profileDeletionErr(profile.Name, fmt.Sprintf("error loading profile config: %v", err))
		return DeletionError{Err: delErr, Errtype: MissingProfile}
	}

	if err == nil && (driver.BareMetal(cc.Driver) || driver.IsSSH(cc.Driver)) {
		if err := uninstallKubernetes(api, *cc, cc.Nodes[0], viper.GetString(cmdcfg.Bootstrapper)); err != nil {
			deletionError, ok := err.(DeletionError)
			if ok {
				delErr := profileDeletionErr(profile.Name, fmt.Sprintf("%v", err))
				deletionError.Err = delErr
				return deletionError
			}
			return err
		}
	}

	if err := killMountProcess(); err != nil {
		out.FailureT("Failed to kill mount process: {{.error}}", out.V{"error": err})
	}

	deleteHosts(api, cc)

	// In case DeleteHost didn't complete the job.
	deleteProfileDirectory(profile.Name)
	deleteMachineDirectories(cc)

	if err := deleteConfig(profile.Name); err != nil {
		return err
	}

	if err := deleteContext(profile.Name); err != nil {
		return err
	}
	out.Step(style.Deleted, `Removed all traces of the "{{.name}}" cluster.`, out.V{"name": profile.Name})
	return nil
}

func deleteHosts(api libmachine.API, cc *config.ClusterConfig) {
	register.Reg.SetStep(register.Deleting)

	if cc != nil {
		for _, n := range cc.Nodes {
			machineName := config.MachineName(*cc, n)
			if err := machine.DeleteHost(api, machineName); err != nil {
				switch errors.Cause(err).(type) {
				case mcnerror.ErrHostDoesNotExist:
					klog.Infof("Host %s does not exist. Proceeding ahead with cleanup.", machineName)
				default:
					out.FailureT("Failed to delete cluster: {{.error}}", out.V{"error": err})
					out.Step(style.Notice, `You may need to manually remove the "{{.name}}" VM from your hypervisor`, out.V{"name": machineName})
				}
			}
		}
	}
}

func deleteConfig(cname string) error {
	if err := config.DeleteProfile(cname); err != nil {
		if config.IsNotExist(err) {
			delErr := profileDeletionErr(cname, fmt.Sprintf("\"%s\" profile does not exist", cname))
			return DeletionError{Err: delErr, Errtype: MissingProfile}
		}
		delErr := profileDeletionErr(cname, fmt.Sprintf("failed to remove profile %v", err))
		return DeletionError{Err: delErr, Errtype: Fatal}
	}
	return nil
}

func deleteContext(machineName string) error {
	if err := kubeconfig.DeleteContext(machineName); err != nil {
		return DeletionError{Err: fmt.Errorf("update config: %v", err), Errtype: Fatal}
	}

	if err := cmdcfg.Unset(config.ProfileName); err != nil {
		return DeletionError{Err: fmt.Errorf("unset minikube profile: %v", err), Errtype: Fatal}
	}
	return nil
}

func deleteInvalidProfile(profile *config.Profile) []error {
	out.Step(style.DeletingHost, "Trying to delete invalid profile {{.profile}}", out.V{"profile": profile.Name})

	var errs []error
	pathToProfile := config.ProfileFolderPath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToProfile); !os.IsNotExist(err) {
		err := os.RemoveAll(pathToProfile)
		if err != nil {
			errs = append(errs, DeletionError{err, Fatal})
		}
	}

	pathToMachine := localpath.MachinePath(profile.Name, localpath.MiniPath())
	if _, err := os.Stat(pathToMachine); !os.IsNotExist(err) {
		err := os.RemoveAll(pathToMachine)
		if err != nil {
			errs = append(errs, DeletionError{err, Fatal})
		}
	}
	return errs
}

func profileDeletionErr(cname string, additionalInfo string) error {
	return fmt.Errorf("error deleting profile \"%s\": %s", cname, additionalInfo)
}

func uninstallKubernetes(api libmachine.API, cc config.ClusterConfig, n config.Node, bsName string) error {
	out.Step(style.Resetting, "Uninstalling Kubernetes {{.kubernetes_version}} using {{.bootstrapper_name}} ...", out.V{"kubernetes_version": cc.KubernetesConfig.KubernetesVersion, "bootstrapper_name": bsName})
	host, err := machine.LoadHost(api, config.MachineName(cc, n))
	if err != nil {
		return DeletionError{Err: fmt.Errorf("unable to load host: %v", err), Errtype: MissingCluster}
	}

	r, err := machine.CommandRunner(host)
	if err != nil {
		return DeletionError{Err: fmt.Errorf("unable to get command runner %v", err), Errtype: MissingCluster}
	}

	clusterBootstrapper, err := cluster.Bootstrapper(api, bsName, cc, r)
	if err != nil {
		return DeletionError{Err: fmt.Errorf("unable to get bootstrapper: %v", err), Errtype: Fatal}
	}

	cr, err := cruntime.New(cruntime.Config{Type: cc.KubernetesConfig.ContainerRuntime, Runner: r})
	if err != nil {
		return DeletionError{Err: fmt.Errorf("unable to get runtime: %v", err), Errtype: Fatal}
	}

	// Unpause the cluster if necessary to avoid hung kubeadm
	_, err = cluster.Unpause(cr, r, nil)
	if err != nil {
		klog.Errorf("unpause failed: %v", err)
	}

	if err = clusterBootstrapper.DeleteCluster(cc.KubernetesConfig); err != nil {
		return DeletionError{Err: fmt.Errorf("failed to delete cluster: %v", err), Errtype: Fatal}
	}
	return nil
}

// HandleDeletionErrors handles deletion errors from DeleteProfiles
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
			out.ErrT(style.Sad, deletionError.Error())
		case MissingCluster:
			out.ErrT(style.Meh, deletionError.Error())
		default:
			out.FatalT(deletionError.Error())
		}
	} else {
		exit.Error(reason.GuestDeletion, "Could not process error from failed deletion", err)
	}
}

func handleMultipleDeletionErrors(errors []error) {
	out.ErrT(style.Sad, "Multiple errors deleting profiles")

	for _, err := range errors {
		deletionError, ok := err.(DeletionError)

		if ok {
			klog.Errorln(deletionError.Error())
		} else {
			exit.Error(reason.GuestDeletion, "Could not process errors from failed deletion", err)
		}
	}
}

func deleteProfileDirectory(profile string) {
	machineDir := filepath.Join(localpath.MiniPath(), "machines", profile)
	if _, err := os.Stat(machineDir); err == nil {
		out.Step(style.DeletingHost, `Removing {{.directory}} ...`, out.V{"directory": machineDir})
		err := os.RemoveAll(machineDir)
		if err != nil {
			exit.Error(reason.GuestProfileDeletion, "Unable to remove machine directory", err)
		}
	}
}

func deleteMachineDirectories(cc *config.ClusterConfig) {
	if cc != nil {
		for _, n := range cc.Nodes {
			machineName := config.MachineName(*cc, n)
			deleteProfileDirectory(machineName)
		}
	}
}

// killMountProcess kills the mount process, if it is running
func killMountProcess() error {
	pidPath := filepath.Join(localpath.MiniPath(), constants.MountProcessFileName)
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		return nil
	}

	klog.Infof("Found %s ...", pidPath)
	out, err := ioutil.ReadFile(pidPath)
	if err != nil {
		return errors.Wrap(err, "ReadFile")
	}
	klog.Infof("pidfile contents: %s", out)
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
		klog.Infof("Stale pid: %d", pid)
		if err := os.Remove(pidPath); err != nil {
			return errors.Wrap(err, "Removing stale pid")
		}
		return nil
	}

	// We found a process, but it still may not be ours.
	klog.Infof("Found process %d: %s", pid, entry.Executable())
	proc, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrap(err, "os.FindProcess")
	}

	klog.Infof("Killing pid %d ...", pid)
	if err := proc.Kill(); err != nil {
		klog.Infof("Kill failed with %v - removing probably stale pid...", err)
		if err := os.Remove(pidPath); err != nil {
			return errors.Wrap(err, "Removing likely stale unkillable pid")
		}
		return errors.Wrap(err, fmt.Sprintf("Kill(%d/%s)", pid, entry.Executable()))
	}
	return nil
}
