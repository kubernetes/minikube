package bsutil

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/command"
)

// AdjustResourceLimits makes fine adjustments to pod resources that aren't possible via kubeadm config.
func AdjustResourceLimits(c command.Runner) error {
	rr, err := c.RunCmd(exec.Command("/bin/bash", "-c", "cat /proc/$(pgrep kube-apiserver)/oom_adj"))
	if err != nil {
		return errors.Wrapf(err, "oom_adj check cmd %s. ", rr.Command())
	}
	glog.Infof("apiserver oom_adj: %s", rr.Stdout.String())
	// oom_adj is already a negative number
	if strings.HasPrefix(rr.Stdout.String(), "-") {
		return nil
	}
	glog.Infof("adjusting apiserver oom_adj to -10")

	// Prevent the apiserver from OOM'ing before other pods, as it is our gateway into the cluster.
	// It'd be preferable to do this via Kubernetes, but kubeadm doesn't have a way to set pod QoS.
	if _, err = c.RunCmd(exec.Command("/bin/bash", "-c", "echo -10 | sudo tee /proc/$(pgrep kube-apiserver)/oom_adj")); err != nil {
		return errors.Wrap(err, fmt.Sprintf("oom_adj adjust"))
	}
	return nil
}

// ExistingConfig checks if there are config files from possible previous kubernets cluster
func ExistingConfig(c command.Runner) error {
	args := append([]string{"ls"}, expectedRemoteArtifacts...)
	_, err := c.RunCmd(exec.Command("sudo", args...))
	return err
}
