// package runtime contains code specific to container runtimes
package runtime

import (
	"fmt"

	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
)

// IsDocker returns whether or not a runtime is considered docker.
func IsDocker(s string) bool {
	return s == "" || s == constants.DockerRuntime
}

// IsCRIO returns whether or not a runtime is considered CRIO
func IsCRIO(s string) bool {
	return s == constants.CrioRuntime || s == constants.Cri_oRuntime
}

// Name is a human readable name for a runtime
func Name(s string) string {
	if IsCRIO(s) {
		return "CRIO"
	}
	if IsDocker(s) {
		return "Docker"
	}
	return s
}

// SocketPath returns the path to a socket file for a given runtime
func SocketPath(path string, r string) string {
	if path != "" {
		glog.Infof("Using supplied socket path: %s", path)
		return path
	}

	if IsCRIO(r) {
		return "/var/run/crio/crio.sock"
	}
	switch r {
	case "containerd":
		return "/run/containerd/containerd.sock"
	default:
		return ""
	}
}

// ConfigureHost configures a VM for the appropriate runtime
func ConfigureHost(h *host.Host, r string) error {
	var err error
	if !IsDocker(r) {
		glog.Infof("Shutting down docker ...")
		if _, err := h.RunSSHCommand("sudo systemctl stop docker"); err == nil {
			_, err = h.RunSSHCommand("sudo systemctl stop docker.socket")
		}
		if err != nil {
			return errors.Wrap(err, "docker stop")
		}
	}
	if !IsCRIO(r) {
		glog.Infof("Shutting down CRIO ...")
		if _, err := h.RunSSHCommand("sudo systemctl stop crio"); err != nil {
			return errors.Wrap(err, "crio stop")
		}
	}
	if r != constants.RktRuntime {
		glog.Infof("Shutting down rkt ...")
		if _, err := h.RunSSHCommand("sudo systemctl stop rkt-api"); err == nil {
			_, err = h.RunSSHCommand("sudo systemctl stop rkt-metadata")
		}
		if err != nil {
			return errors.Wrap(err, "rkt stop")
		}
	}

	if r == constants.ContainerdRuntime {
		glog.Infof("Restarting containerd ...")
		// restart containerd so that it can install all plugins
		if _, err := h.RunSSHCommand("sudo systemctl restart containerd"); err != nil {
			return errors.Wrap(err, "containerd stop")
		}
	}
	return nil
}

// KubeletOptions returns kubelet options for a runtime.
func KubeletOptions(r string, cfg map[string]string) map[string]string {
	// If the options already have a configured runtime, leave everything alone.
	if _, ok := cfg["container-runtime"]; ok {
		if r != cfg["container-runtime"] {
			glog.Warningf("container-runtime is already %q, ignoring value %q", cfg["container-runtime"], r)
		}
		return cfg
	}

	socket := SocketPath("", r)
	if IsCRIO(r) {
		cfg["container-runtime"] = "remote"
		cfg["container-runtime-endpoint"] = socket
		cfg["image-service-endpoint"] = socket
		cfg["runtime-request-timeout"] = "15m"
	}
	if r == "containerd" {
		cfg["container-runtime"] = "remote"
		cfg["container-runtime-endpoint"] = fmt.Sprintf("unix://%s", socket)
		cfg["image-service-endpoint"] = fmt.Sprintf("unix://%s", socket)
		cfg["runtime-request-timeout"] = "15m"
	}
	cfg["container-runtime"] = r
	return cfg
}
