/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"k8s.io/klog/v2"
	"sigs.k8s.io/boskos/client"
	"sigs.k8s.io/kubetest2/pkg/boskos"
)

const (
	// gceProjectResourceType is called "gce" project in Boskos,
	// while it is called "gcp" project in this CLI.
	gceProjectResourceType = "gce-project"
	runDirSSHKeys          = "gce-ssh-keys"
	remoteUserName         = "minitest"
)

type MiniTestBosKosDeployer struct {
	ctx    context.Context
	config *MiniTestBoskosConfig
	isUp   bool

	id               string
	gcpProject       string
	remoteUserName   string
	networkName      string
	firewallRuleName string
	instanceName     string
	sshAddr          string

	boskosClient *client.Client
	// this channel serves as a signal channel for the heartbeat goroutine
	// so that it can be explicitly closed
	boskosHeartbeatClose chan struct{}
}

func NewMiniTestBosKosDeployerFromConfigFile(path string) MiniTestDeployer {
	config := MiniTestBoskosConfig{}
	data, err := os.ReadFile(path)
	if err != nil {
		klog.Fatalf("failed to read config file %s: %v", path, err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		klog.Fatalf("failed to parse config file %s: %v", path, err)
	}
	return NewMiniTestBosKosDeployer(&config)
}

func NewMiniTestBosKosDeployer(config *MiniTestBoskosConfig) MiniTestDeployer {
	boskosClient, err := boskos.NewClient(config.BoskosLocation)
	if err != nil {
		klog.Fatalf("failed to make boskos client: %v", err)
	}
	id := uuid.New().String()[:8]
	return &MiniTestBosKosDeployer{
		ctx:                  context.TODO(),
		config:               config,
		boskosClient:         boskosClient,
		boskosHeartbeatClose: make(chan struct{}),
		remoteUserName:       remoteUserName,
		id:                   id,
		networkName:          "minitest-network-" + id,
		firewallRuleName:     "minitest-firewall-" + id,
		instanceName:         "minitest-vm-" + id,
	}
}

func (m *MiniTestBosKosDeployer) Up() error {

	if err := m.requestGCPProject(); err != nil {
		klog.Errorf("Failed to request gcp project from boskos: %v", err)
		return err
	}
	if err := m.gcpVMSetUp(); err != nil {
		klog.Errorf("Failed to start a vm in gcp: %v", err)
		return err
	}
	if err := m.gcpSSHSetUp(); err != nil {
		klog.Errorf("Failed to set up ssh: %v", err)
		return err
	}
	if err := sshConnectionCheck(m.ctx, m.remoteUserName, m.sshAddr, nil); err != nil {
		klog.Errorf("Failed to conntect via ssh: %v", err)
		return err
	}

	klog.Infof("Successfully started vm in gcp: %s", m.instanceName)
	m.isUp = true
	return nil

}

func (m *MiniTestBosKosDeployer) Down() error {
	//todo: clean up the VM?

	klog.Info("Releasing bosko project")
	if m.boskosClient == nil {
		return fmt.Errorf("m.boskosClient not set")
	}
	err := boskos.Release(
		m.boskosClient,
		[]string{m.gcpProject},
		m.boskosHeartbeatClose,
	)
	if err != nil {
		return fmt.Errorf("down failed to release boskos project: %v", err)
	}
	m.isUp = false
	return nil

}

func (m *MiniTestBosKosDeployer) IsUp() (bool, error) {
	return m.isUp, nil
}
func (m *MiniTestBosKosDeployer) Execute(args ...string) error {
	return executeSSHCommand(m.ctx, m.remoteUserName, m.sshAddr, nil, args...)
}

func (m *MiniTestBosKosDeployer) SyncToRemote(src string, dst string, excludedPattern []string) error {
	excludedArgs := make([]string, 0, len(excludedPattern)*2)
	for _, pattern := range excludedPattern {
		excludedArgs = append(excludedArgs, "--exclude", pattern)
	}
	dstRemote := fmt.Sprintf("%s@%s:%s", m.remoteUserName, m.sshAddr, dst)
	return executeRsyncSSHCommand(m.ctx, nil, src, dstRemote, excludedArgs)
}

func (m *MiniTestBosKosDeployer) SyncToHost(src string, dst string, excludedPattern []string) error {
	excludedArgs := make([]string, 0, len(excludedPattern)*2)
	for _, pattern := range excludedPattern {
		excludedArgs = append(excludedArgs, "--exclude", pattern)
	}
	srcRemote := fmt.Sprintf("%s@%s:%s", m.remoteUserName, m.sshAddr, src)
	return executeRsyncSSHCommand(m.ctx, nil, srcRemote, dst, excludedArgs)
}

func (m *MiniTestBosKosDeployer) requestGCPProject() error {

	resource, err := boskos.Acquire(
		m.boskosClient,
		gceProjectResourceType,
		time.Duration(m.config.BoskosAcquireTimeoutSeconds)*time.Second,
		time.Duration(m.config.BoskosHeartbeatIntervalSeconds)*time.Second,
		m.boskosHeartbeatClose,
	)

	if err != nil {
		return fmt.Errorf("failed to get project from boskos: %v", err)
	}
	if resource.Name == "" {
		return fmt.Errorf("boskos returned an empty resource name, resource: %v", resource)
	}

	klog.Infof("Got project %q from boskos", resource.Name)
	m.gcpProject = resource.Name
	return nil
}

func (m *MiniTestBosKosDeployer) gcpVMSetUp() error {
	// ensure gcp zone is set to start a vm
	if m.config.GCPZone == "" {
		return fmt.Errorf("GCPZone is not set")
	}

	// execute gcloud commands to set environment up a vm
	if err := m.executeLocalGloudCommand("services", "enable", "compute.googleapis.com"); err != nil {
		klog.Warningf("failed to enable service: %v", err)
	}
	if err := m.executeLocalGloudCommand("compute", "networks", "create", m.networkName); err != nil {
		klog.Warningf("failed to set up network: %v", err)
	}
	if err := m.executeLocalGloudCommand("compute", "firewall-rules", "create", m.firewallRuleName, "--network="+m.networkName, "--allow=tcp:22"); err != nil {
		klog.Warningf("failed to set up firewalls: %v", err)
	}

	description := fmt.Sprintf("%s instance (login ID: %q)", m.instanceName, m.remoteUserName)
	instImgPair := strings.SplitN(m.config.InstanceImage, "/", 2)

	// Create the VM. vmMachineTypes returns the next machine to try, from best
	// to worst. Try each type in every us-central1 zone before the next type.
	var err error
	zones := vmZones(m.config.GCPZone)
	for _, machineType := range vmMachineTypes(m.config.InstanceType, m.config.EnableNestedVirtualization) {
		for _, zone := range zones {
			createArgs := []string{"compute", "instances", "create"}
			if m.config.EnableNestedVirtualization {
				createArgs = append(createArgs, "--enable-nested-virtualization")
			}
			createArgs = append(createArgs,
				"--zone="+zone,
				"--description="+description,
				"--network="+m.networkName,
				"--image-project="+instImgPair[0],
				"--image-family="+instImgPair[1],
				"--machine-type="+machineType,
				fmt.Sprintf("--boot-disk-size=%dGiB", m.config.DiskGiB),
				"--boot-disk-type=pd-ssd",
				"--metadata=block-project-ssh-keys=TRUE",
				m.instanceName,
			)
			if err = m.executeLocalGloudCommand(createArgs...); err != nil {
				klog.Warningf("failed to create %s in %s: %v", machineType, zone, err)
				continue
			}
			m.config.GCPZone = zone
			return nil
		}
	}
	klog.Errorf("failed to start a gcp vm: %v", err)
	return err

}

// vmMachineTypes returns machine types to try, from best to worst. The
// configured type is first.
// none falls back through e2-standard-4, n2-standard-4, n2d-standard-4, and
// n1-standard-4. e2-standard-4 is enough: none finishes in 30m or less.
// kvm falls back through n2-standard-8, n1-standard-8, n2-standard-4, and
// n1-standard-4. n2-standard-8 finished in about 45m, n2-standard-4 in about
// 75m, so the larger types come first.
// Machine types: https://cloud.google.com/compute/docs/general-purpose-machines
// kvm skips E2 and N2D: Compute Engine does not expose VT-x or AMD-V on those
// types. N4D does expose AMD-V, but it only boots from Hyperdisk, and this
// create path uses pd-ssd.
func vmMachineTypes(preferred string, nested bool) []string {
	candidates := []string{"e2-standard-4", "n2-standard-4", "n2d-standard-4", "n1-standard-4"}
	if nested {
		candidates = []string{"n2-standard-8", "n1-standard-8", "n2-standard-4", "n1-standard-4"}
	}
	if preferred == "" {
		return candidates
	}
	types := make([]string, 0, len(candidates)+1)
	types = append(types, preferred)
	for _, machineType := range candidates {
		if machineType != preferred {
			types = append(types, machineType)
		}
	}
	return types
}

// vmZones returns us-central1 zones to try. The configured zone is first.
// The rest are shuffled so jobs do not all fall over to the same zone.
// Zones: https://cloud.google.com/compute/docs/regions-zones
//
// Follow-up: check whether other regions, such as us-east1 or us-west1, are
// less prone to stockouts and whether these jobs can create the VM there.
func vmZones(preferred string) []string {
	candidates := []string{"us-central1-a", "us-central1-b", "us-central1-c", "us-central1-f"}
	if preferred != "" {
		zones := make([]string, 0, len(candidates))
		for _, zone := range candidates {
			if zone != preferred {
				zones = append(zones, zone)
			}
		}
		candidates = zones
	}
	rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
	if preferred == "" {
		return candidates
	}
	return append([]string{preferred}, candidates...)
}

func (m *MiniTestBosKosDeployer) gcpSSHSetUp() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to find home dir: %v", err)
	}
	gcePubPath := filepath.Join(home, ".ssh/google_compute_engine.pub")
	gcePubContent, err := os.ReadFile(gcePubPath)
	if err != nil {
		return fmt.Errorf("failed to read GCE public key: %v", err)
	}
	sshKeysContent := []byte(m.remoteUserName + ":" + string(gcePubContent))
	runDir, err := os.MkdirTemp("", "rundir")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	sshKeysPath := filepath.Join(runDir, runDirSSHKeys)
	_ = os.RemoveAll(sshKeysPath)
	if err = os.WriteFile(sshKeysPath, sshKeysContent, 0400); err != nil {
		return fmt.Errorf("failed to create new public key file: %v", err)
	}

	// set up ssh login with pub key
	if err := m.executeLocalGloudCommand("compute", "instances", "add-metadata", m.instanceName, "--zone="+m.config.GCPZone, "--metadata-from-file=ssh-keys="+sshKeysPath); err != nil {
		klog.Warningf("failed to add metadata: %v", err)
		// continue anyway
	}
	// update the local ssh config
	if err := m.executeLocalGloudCommand("compute", "config-ssh"); err != nil {
		klog.Warningf("failed to compute ssh: %v", err)
		// continue anyway
	}

	// set sshAddr field
	if m.instanceName == "" || m.config.GCPZone == "" || m.gcpProject == "" {
		return fmt.Errorf("gcp project configuration incorrect: %s.%s.%s", m.instanceName, m.config.GCPZone, m.gcpProject)
	}
	m.sshAddr = fmt.Sprintf("%s.%s.%s", m.instanceName, m.config.GCPZone, m.gcpProject)

	return nil
}

func (m *MiniTestBosKosDeployer) executeLocalGloudCommand(args ...string) error {

	if err := executeLocalCommand(m.ctx, "gcloud", append([]string{"--project=" + m.gcpProject}, args...)...); err != nil {
		return fmt.Errorf("failed to run %v: %v", args, err)
	}
	return nil
}
