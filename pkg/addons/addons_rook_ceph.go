/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package addons

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

const (
	// rookCephOSDImagePath is the path inside the minikube node where the sparse OSD image lives.
	rookCephOSDImagePath = "/data/rook/ceph-osd-data.img"
	// rookCephDataDir is the Rook data directory on the host (dataDirHostPath in CephCluster).
	rookCephDataDir = "/data/rook"
	// defaultOSDSize is the default sparse file size for the OSD loop device.
	defaultOSDSize = "6Gi"
	losetupBinary  = "/usr/sbin/losetup"
)

// enableOrDisableRookCeph handles rook-ceph addon enable/disable with special logic:
//   - On enable: creates a loop device from a sparse file so Ceph has an OSD to use.
//   - On disable: performs ordered teardown to avoid finalizer deadlocks.
func enableOrDisableRookCeph(cc *config.ClusterConfig, name, val string, options *run.CommandOptions) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("parsing bool: %s: %w", name, err)
	}

	if enable {
		return enableRookCeph(cc, name, val, options)
	}
	return disableRookCeph(cc, name, val, options)
}

// enableRookCeph sets up the loop device and then delegates to the standard EnableOrDisableAddon.
func enableRookCeph(cc *config.ClusterConfig, name, val string, options *run.CommandOptions) error {
	co := mustload.Running(cc.Name, options)
	runner := co.CP.Runner

	if _, err := runner.RunCmd(exec.Command("sudo", "mkdir", "-p", rookCephDataDir)); err != nil {
		return fmt.Errorf("creating rook data dir: %w", err)
	}

	// Check if the configured OSD device actually exists on the node.
	deviceExists := false
	if cc.RookCephOSDDevice != "" {
		// Does the block device file exist?
		if _, err := runner.RunCmd(exec.Command("sudo", "test", "-b", cc.RookCephOSDDevice)); err == nil {
			// Is device attached?
			rr, err := runner.RunCmd(exec.Command("sudo", "lsblk", "-dno", "SIZE", cc.RookCephOSDDevice))
			if err == nil {
				size := strings.TrimSpace(rr.Stdout.String())
				if size != "" && size != "0B" {
					deviceExists = true
				}
			}
		}
	}

	if !deviceExists {
		osdSize := cc.RookCephOSDSize
		if osdSize == "" {
			osdSize = defaultOSDSize
		}

		sizeBytes, err := parseK8sSize(osdSize)
		if err != nil {
			return fmt.Errorf("invalid OSD size %q: %w", osdSize, err)
		}

		out.Step(style.SubStep, "Preparing Ceph OSD loop device ({{.size}})...", out.V{"size": osdSize})

		// Check if the sparse file already exists; if not, create it.
		if _, err := runner.RunCmd(exec.Command("sudo", "test", "-f", rookCephOSDImagePath)); err != nil {
			klog.Infof("Creating sparse OSD image at %s (%d bytes)", rookCephOSDImagePath, sizeBytes)
			if _, err := runner.RunCmd(exec.Command("sudo", "truncate", "-s", fmt.Sprintf("%d", sizeBytes), rookCephOSDImagePath)); err != nil {
				return fmt.Errorf("creating sparse OSD image: %w", err)
			}
		} else {
			klog.Infof("OSD image already exists at %s, reusing", rookCephOSDImagePath)
		}

		// Check if the loop device is already attached.
		rr, err := runner.RunCmd(exec.Command("sudo", losetupBinary, "-j", rookCephOSDImagePath))
		if err == nil && strings.TrimSpace(rr.Stdout.String()) != "" {
			klog.Infof("Loop device already attached: %s", strings.TrimSpace(rr.Stdout.String()))
			lines := strings.Split(strings.TrimSpace(rr.Stdout.String()), "\n")
			parts := strings.Split(lines[0], ":")
			if len(parts) > 0 {
				cc.RookCephOSDDevice = strings.TrimSpace(parts[0])
			}
		} else {
			// Attach the sparse file as a loop device.
			rr, err = runner.RunCmd(exec.Command("sudo", losetupBinary, "--find", "--show", rookCephOSDImagePath))
			if err != nil {
				return fmt.Errorf("attaching loop device: %w", err)
			}
			loopDev := strings.TrimSpace(rr.Stdout.String())
			out.Step(style.SubStep, "Loop device created: {{.dev}}", out.V{"dev": loopDev})
			cc.RookCephOSDDevice = loopDev
		}
	} else {
		out.Step(style.SubStep, "Using configured Ceph OSD device: {{.dev}}", out.V{"dev": cc.RookCephOSDDevice})
	}

	if err := config.SaveProfile(cc.Name, cc); err != nil {
		klog.Warningf("Failed to save profile with loop device info: %v", err)
	}

	return EnableOrDisableAddon(cc, name, val, options)
}

// disableRookCeph performs an ordered teardown of rook-ceph resources to avoid finalizer deadlocks.
// The order is:
//  1. Patch CephCluster to allow cleanup
//  2. Delete Ceph CRs (cluster YAML) — wait for operator to process finalizers
//  3. Force-remove any remaining finalizers
//  4. Delete operator resources (operator YAML)
//  5. Delete CRDs (CRDs YAML)
//  6. Clean up loop device and sparse file
func disableRookCeph(cc *config.ClusterConfig, name, val string, options *run.CommandOptions) error {
	co := mustload.Running(cc.Name, options)
	runner := co.CP.Runner

	addon := assets.Addons[name]

	v := constants.DefaultKubernetesVersion
	if cc != nil {
		v = cc.KubernetesConfig.KubernetesVersion
	}
	kubectlBinary := kapi.KubectlBinaryPath(v)
	kubeconfigPath := path.Join(vmpath.GuestPersistentDir, "kubeconfig")

	out.Step(style.SubStep, "Shutting down Ceph cluster (this may take a minute)...")

	// Step 1: Patch CephCluster cleanupPolicy so operator knows it should tear down.
	patchJSON := `{"spec":{"cleanupPolicy":{"confirmation":"yes-really-destroy-data","allowUninstallWithVolumes":true}}}`
	_, err := runner.RunCmd(exec.Command("sudo",
		fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath), kubectlBinary,
		"patch", "cephcluster", "my-cluster", "-n", "rook-ceph",
		"--type", "merge", "-p", patchJSON,
	))
	if err != nil {
		klog.Warningf("Failed to patch CephCluster cleanupPolicy (may not exist yet): %v", err)
	}

	// Step 2: Delete Ceph CRs (the cluster YAML) first — the operator must be running to handle finalizers.
	clusterFile := path.Join(vmpath.GuestAddonsDir, "rook-ceph-cluster.yaml")
	deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer deleteCancel()

	klog.Infof("Deleting rook-ceph cluster resources from %s", clusterFile)
	_, err = runner.RunCmd(exec.CommandContext(deleteCtx, "sudo",
		fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath), kubectlBinary,
		"delete", "--ignore-not-found", "--wait=true", "--timeout=60s",
		"-f", clusterFile,
	))
	if err != nil {
		klog.Warningf("Cluster resource deletion returned error (will force-remove finalizers): %v", err)
	}

	// Step 3: Force-remove any stuck finalizers on Ceph CRs.
	out.Step(style.SubStep, "Cleaning up Ceph finalizers...")
	forceRemoveCephFinalizers(runner, kubectlBinary, kubeconfigPath)

	// Step 4: Delete operator resources (RBAC, deployment, namespace resources).
	operatorFile := path.Join(vmpath.GuestAddonsDir, "rook-ceph-operator.yaml")
	opCtx, opCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer opCancel()

	klog.Infof("Deleting rook-ceph operator resources from %s", operatorFile)
	_, err = runner.RunCmd(exec.CommandContext(opCtx, "sudo",
		fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath), kubectlBinary,
		"delete", "--ignore-not-found", "--wait=false",
		"-f", operatorFile,
	))
	if err != nil {
		klog.Warningf("Operator resource deletion error: %v", err)
	}

	// Step 5: Delete CRDs.
	crdsFile := path.Join(vmpath.GuestAddonsDir, "rook-ceph-crds.yaml")
	crdCtx, crdCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer crdCancel()

	klog.Infof("Deleting rook-ceph CRDs from %s", crdsFile)
	_, err = runner.RunCmd(exec.CommandContext(crdCtx, "sudo",
		fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath), kubectlBinary,
		"delete", "--ignore-not-found", "--wait=false",
		"-f", crdsFile,
	))
	if err != nil {
		klog.Warningf("CRD deletion error: %v", err)
	}

	// Step 6: Clean up loop device and sparse OSD image on the node.
	out.Step(style.SubStep, "Cleaning up Ceph OSD loop device and data...")
	cleanupLoopDevice(runner)

	// Clean up the addon YAML files from the guest.
	for _, a := range addon.Assets {
		var f assets.CopyableFile = a
		fPath := path.Join(f.GetTargetDir(), f.GetTargetName())
		klog.Infof("Removing addon file %s", fPath)
		if err := runner.Remove(f); err != nil {
			klog.Warningf("Error removing %s: %v", fPath, err)
		}
	}

	if cc != nil {
		cc.RookCephOSDDevice = ""
		if err := config.SaveProfile(cc.Name, cc); err != nil {
			klog.Warningf("Failed to clear loop device info on disable: %v", err)
		}
	}

	return nil
}

// forceRemoveCephFinalizers patches out finalizers from all Ceph CRs in the rook-ceph namespace.
func forceRemoveCephFinalizers(runner command.Runner, kubectlBinary, kubeconfigPath string) {
	cephResources := []string{
		"cephcluster",
		"cephblockpool",
		"cephfilesystem",
		"cephobjectstore",
		"cephnfs",
		"cephrbdmirror",
		"cephfilesystemmirror",
	}

	for _, resource := range cephResources {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		// Get all resource names in the namespace.
		rr, err := runner.RunCmd(exec.CommandContext(ctx, "sudo",
			fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath), kubectlBinary,
			"get", resource, "-n", "rook-ceph",
			"-o", "jsonpath={.items[*].metadata.name}",
		))
		cancel()

		if err != nil || strings.TrimSpace(rr.Stdout.String()) == "" {
			continue
		}

		names := strings.Fields(strings.TrimSpace(rr.Stdout.String()))
		for _, n := range names {
			patchCtx, patchCancel := context.WithTimeout(context.Background(), 10*time.Second)
			klog.Infof("Removing finalizers from %s/%s", resource, n)
			_, err := runner.RunCmd(exec.CommandContext(patchCtx, "sudo",
				fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath), kubectlBinary,
				"patch", resource, n, "-n", "rook-ceph",
				"--type", "merge", "-p", `{"metadata":{"finalizers":null}}`,
			))
			patchCancel()
			if err != nil {
				klog.Warningf("Failed to remove finalizers from %s/%s: %v", resource, n, err)
			}
		}
	}
}

// cleanupLoopDevice detaches loop devices backed by the OSD image and removes the image + data dir.
func cleanupLoopDevice(runner command.Runner) {
	rr, err := runner.RunCmd(exec.Command("sudo", losetupBinary, "-j", rookCephOSDImagePath))
	if err == nil && strings.TrimSpace(rr.Stdout.String()) != "" {
		for _, line := range strings.Split(rr.Stdout.String(), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 1 {
				continue
			}
			loopDev := strings.TrimSpace(parts[0])
			klog.Infof("Detaching loop device %s", loopDev)
			if _, err := runner.RunCmd(exec.Command("sudo", losetupBinary, "-d", loopDev)); err != nil {
				klog.Warningf("Failed to detach %s: %v", loopDev, err)
			}
		}
	}

	// Remove the sparse image file.
	if _, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", rookCephOSDImagePath)); err != nil {
		klog.Warningf("Failed to remove OSD image: %v", err)
	}

	if _, err := runner.RunCmd(exec.Command("sudo", "rm", "-rf", rookCephDataDir)); err != nil {
		klog.Warningf("Failed to remove rook data dir: %v", err)
	}
}

// parseK8sSize converts a Kubernetes-style size string (e.g., "6Gi", "10Gi") to bytes.
func parseK8sSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	multipliers := map[string]int64{
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"K":  1000,
		"M":  1000 * 1000,
		"G":  1000 * 1000 * 1000,
		"T":  1000 * 1000 * 1000 * 1000,
	}

	for _, suffix := range []string{"Ti", "Gi", "Mi", "Ki", "T", "G", "M", "K"} {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number %q: %w", numStr, err)
			}
			return int64(num * float64(multipliers[suffix])), nil
		}
	}

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}
	return num, nil
}
