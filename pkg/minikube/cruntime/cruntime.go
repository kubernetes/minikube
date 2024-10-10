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

// Package cruntime contains code specific to container runtimes
package cruntime

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

// ContainerState is the run state of a container
type ContainerState int

const (
	// All is all of the states
	All ContainerState = iota
	// Running is only running
	Running
	// Paused is only paused
	Paused
)

func (cs ContainerState) String() string {
	return [...]string{"all", "running", "paused"}[cs]
}

// ValidRuntimes lists the supported container runtimes
func ValidRuntimes() []string {
	return []string{"docker", "cri-o", "containerd"}
}

// CommandRunner is the subset of command.Runner this package consumes
type CommandRunner interface {
	// RunCmd is a blocking method that runs a command
	// Use this if you don't need to stream stdout and stderr in real-time
	RunCmd(cmd *exec.Cmd) (*command.RunResult, error)
	// StartCmd is a non-blocking method that starts a command
	// Use WaitCmd to block until the command is complete
	// Use this if you need to stream stdout and/or stderr in real-time
	StartCmd(cmd *exec.Cmd) (*command.StartedCmd, error)
	// WaitCmd blocks until the started command completes
	WaitCmd(sc *command.StartedCmd) (*command.RunResult, error)
	// Copy is a convenience method that runs a command to copy a file
	Copy(assets.CopyableFile) error
	// CopyFrom is a convenience method that runs a command to copy a file back
	CopyFrom(assets.CopyableFile) error
	// Remove is a convenience method that runs a command to remove a file
	Remove(assets.CopyableFile) error

	ReadableFile(sourcePath string) (assets.ReadableFile, error)
}

// Manager is a common interface for container runtimes
type Manager interface {
	// Name is a human readable name for a runtime
	Name() string
	// Version retrieves the current version of this runtime
	Version() (string, error)
	// Enable idempotently enables this runtime on a host
	Enable(bool, string, bool) error
	// Disable idempotently disables this runtime on a host
	Disable() error
	// Active returns whether or not a runtime is active on a host
	Active() bool
	// Available returns an error if it is not possible to use this runtime on a host
	Available() error
	// Style is an associated StyleEnum for Name()
	Style() style.Enum

	// CGroupDriver returns cgroup driver ("cgroupfs" or "systemd")
	CGroupDriver() (string, error)
	// KubeletOptions returns kubelet options for a runtime.
	KubeletOptions() map[string]string
	// SocketPath returns the path to the socket file for a given runtime
	SocketPath() string

	// Load an image idempotently into the runtime on a host
	LoadImage(string) error
	// Pull an image to the runtime from the container registry
	PullImage(string) error
	// Build an image idempotently into the runtime on a host
	BuildImage(string, string, string, bool, []string, []string) error
	// Save an image from the runtime on a host
	SaveImage(string, string) error
	// Tag an image
	TagImage(string, string) error
	// Push an image from the runtime to the container registry
	PushImage(string) error

	// ImageExists takes image name and optionally image sha to check if an image exists
	ImageExists(string, string) bool
	// ListImages returns a list of images managed by this container runtime
	ListImages(ListImagesOptions) ([]ListImage, error)

	// RemoveImage remove image based on name
	RemoveImage(string) error

	// ListContainers returns a list of containers managed by this container runtime
	ListContainers(ListContainersOptions) ([]string, error)
	// KillContainers removes containers based on ID
	KillContainers([]string) error
	// StopContainers stops containers based on ID
	StopContainers([]string) error
	// PauseContainers pauses containers based on ID
	PauseContainers([]string) error
	// UnpauseContainers unpauses containers based on ID
	UnpauseContainers([]string) error
	// ContainerLogCmd returns the command to retrieve the log for a container based on ID
	ContainerLogCmd(string, int, bool) string
	// SystemLogCmd returns the command to return the system logs
	SystemLogCmd(int) string
	// Preload preloads the container runtime with k8s images
	Preload(config.ClusterConfig) error
	// ImagesPreloaded returns true if all images have been preloaded
	ImagesPreloaded([]string) bool
}

// Config is runtime configuration
type Config struct {
	// Type of runtime to create ("docker, "crio", etc)
	Type string
	// Custom path to a socket file
	Socket string
	// Runner is the CommandRunner object to execute commands with
	Runner CommandRunner
	// NetworkPlugin name of networking plugin ("cni")
	NetworkPlugin string
	// ImageRepository image repository to download image from
	ImageRepository string
	// KubernetesVersion Kubernetes version
	KubernetesVersion semver.Version
	// InsecureRegistry list of insecure registries
	InsecureRegistry []string
	// GPUs add GPU devices to the container
	GPUs string
}

// ListContainersOptions are the options to use for listing containers
type ListContainersOptions struct {
	// State is the container state to filter by (All, Running, Paused)
	State ContainerState
	// Name is a name filter
	Name string
	// Namespaces is the namespaces to look into
	Namespaces []string
}

// ListImagesOptions are the options to use for listing images
type ListImagesOptions struct {
}

type ListImage struct {
	ID          string   `json:"id" yaml:"id"`
	RepoDigests []string `json:"repoDigests" yaml:"repoDigests"`
	RepoTags    []string `json:"repoTags" yaml:"repoTags"`
	Size        string   `json:"size" yaml:"size"`
}

// ErrContainerRuntimeNotRunning is thrown when container runtime is not running
var ErrContainerRuntimeNotRunning = errors.New("container runtime is not running")

// ErrServiceVersion is the error returned when disk image has incompatible version of service
type ErrServiceVersion struct {
	// Service is the name of the incompatible service
	Service string
	// Installed is the installed version of Service
	Installed string
	// Required is the minimum required version of Service
	Required string
}

// NewErrServiceVersion creates a new ErrServiceVersion
func NewErrServiceVersion(svc, required, installed string) *ErrServiceVersion {
	return &ErrServiceVersion{
		Service:   svc,
		Installed: installed,
		Required:  required,
	}
}

func (e ErrServiceVersion) Error() string {
	return fmt.Sprintf("service %q version is %v. Required: %v",
		e.Service, e.Installed, e.Required)
}

// New returns an appropriately configured runtime
func New(c Config) (Manager, error) {
	sm := sysinit.New(c.Runner)

	switch c.Type {
	case "", "docker":
		sp := c.Socket
		cs := ""
		// There is no more dockershim socket, in Kubernetes version 1.24 and beyond
		if sp == "" && c.KubernetesVersion.GTE(semver.MustParse("1.24.0-alpha.0")) {
			sp = ExternalDockerCRISocket
			cs = "cri-docker.socket"
		}
		return &Docker{
			Socket:            sp,
			Runner:            c.Runner,
			NetworkPlugin:     c.NetworkPlugin,
			ImageRepository:   c.ImageRepository,
			KubernetesVersion: c.KubernetesVersion,
			Init:              sm,
			UseCRI:            (sp != ""), // !dockershim
			CRIService:        cs,
			GPUs:              c.GPUs,
		}, nil
	case "crio", "cri-o":
		return &CRIO{
			Socket:            c.Socket,
			Runner:            c.Runner,
			ImageRepository:   c.ImageRepository,
			KubernetesVersion: c.KubernetesVersion,
			Init:              sm,
		}, nil
	case "containerd":
		return &Containerd{
			Socket:            c.Socket,
			Runner:            c.Runner,
			ImageRepository:   c.ImageRepository,
			KubernetesVersion: c.KubernetesVersion,
			Init:              sm,
			InsecureRegistry:  c.InsecureRegistry,
		}, nil
	default:
		return nil, fmt.Errorf("unknown runtime type: %q", c.Type)
	}
}

// ContainerStatusCommand works across container runtimes with good formatting
func ContainerStatusCommand() string {
	// Fallback to 'docker ps' if it fails (none driver)
	return "sudo `which crictl || echo crictl` ps -a || sudo docker ps -a"
}

// disableOthers disables all other runtimes except for me.
func disableOthers(me Manager, cr CommandRunner) error {
	// valid values returned by manager.Name()
	runtimes := []string{"containerd", "crio", "docker"}
	for _, name := range runtimes {
		r, err := New(Config{Type: name, Runner: cr})
		if err != nil {
			return fmt.Errorf("runtime(%s): %v", name, err)
		}

		// Don't disable myself.
		if r.Name() == me.Name() {
			continue
		}

		// Don't disable containerd if we are bound to it
		if me.Name() == "Docker" && r.Name() == "containerd" && dockerBoundToContainerd(cr) {
			klog.Infof("skipping containerd shutdown because we are bound to it")
			continue
		}
		// in case of docker, if other runtime are already not active we are sure it is disabled, nothing to do.
		// because #11515 for non-docker runtimes, we gotta ensure Docker is disabled and can not just check if it is not active
		// since it is enabled by default in the current base image and it keeps coming back to life
		if me.Name() == "Docker" && !r.Active() {
			continue
		}

		if err = r.Disable(); err != nil {
			klog.Warningf("disable failed: %v", err)
		}

		// Validate that the runtime really is offline - and that Active & Disable are properly written.
		if r.Active() {
			return fmt.Errorf("%s is still active", r.Name())
		}
	}
	return nil
}

var requiredContainerdVersion = semver.MustParse("1.4.0")

// compatibleWithVersion checks if current version of "runtime" is compatible with version "v"
func compatibleWithVersion(runtime, v string) error {
	if runtime == "containerd" {
		vv, err := semver.Make(v)
		if err != nil {
			return err
		}
		if requiredContainerdVersion.GT(vv) {
			return NewErrServiceVersion(runtime, requiredContainerdVersion.String(), vv.String())
		}
	}
	return nil
}

// CheckCompatibility checks if the container runtime managed by "cr" is compatible with current minikube code
// returns: NewErrServiceVersion if not
func CheckCompatibility(cr Manager) error {
	v, err := cr.Version()
	if err != nil {
		return errors.Wrap(err, "Failed to check container runtime version")
	}
	return compatibleWithVersion(cr.Name(), v)
}

// CheckKernelCompatibility returns an error when the kernel is older than the specified version.
func CheckKernelCompatibility(cr CommandRunner, major, minor int) error {
	expected := fmt.Sprintf("%d.%d", major, minor)
	unameRes, err := cr.RunCmd(exec.Command("uname", "-r"))
	if err != nil {
		return err
	}
	actual := strings.TrimSpace(unameRes.Stdout.String())
	sortRes, err := cr.RunCmd(exec.Command("sh", "-euc", fmt.Sprintf(`(echo %s; echo %s) | sort -V | head -n1`, actual, expected)))
	if err != nil {
		return err
	}
	comparison := strings.TrimSpace(sortRes.Stdout.String())
	if comparison != expected {
		return NewErrServiceVersion("kernel", expected, actual)
	}
	return nil
}
