package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
	config *MiniTestConfig
	isUp   bool

	id               string
	gcpProject       string
	remoteUserName   string
	networkName      string
	firewallRuleName string
	instanceName     string

	boskosClient *client.Client
	// this channel serves as a signal channel for the hearbeat goroutine
	// so that it can be explicitly closed
	boskosHeartbeatClose chan struct{}
}

func NewMiniTestBosKosDeployer(config *MiniTestConfig) MiniTestDeployer {
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

	if err:=m.sshConnectionCheck();err!=nil{
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

func (m *MiniTestBosKosDeployer) DumpLogs() error {
	return nil
}

func (m *MiniTestBosKosDeployer) SSHAddr() (string, error) {
	addr := fmt.Sprintf("%s.%s.%s", m.instanceName, m.config.GCPZone, m.gcpProject)
	if m.instanceName == "" || m.config.GCPZone == "" || m.gcpProject == "" {
		return "", fmt.Errorf("gcp project configuration incorrect: %s", addr)
	}
	return addr, nil
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
	if err := m.executeGloudCommand(m.ctx, "services", "enable", "compute.googleapis.com"); err != nil {
		klog.Warningf("failed to enable service: %v", err)
	}
	if err := m.executeGloudCommand(m.ctx, "compute", "networks", "create", m.networkName); err != nil {
		klog.Warningf("failed to set up network: %v", err)
	}
	if err := m.executeGloudCommand(m.ctx, "compute", "firewall-rules", "create", m.firewallRuleName, "--network="+m.networkName, "--allow=tcp:22"); err != nil {
		klog.Warningf("failed to set up firewalls: %v", err)
	}

	// create the vm
	description := fmt.Sprintf("%s instance (login ID: %q)", m.instanceName, m.remoteUserName)
	instImgPair := strings.SplitN(m.config.InstanceImage, "/", 2)
	if err := m.executeGloudCommand(m.ctx, "compute", "instances", "create",
		"--zone="+m.config.GCPZone,
		"--description="+description,
		"--network="+m.networkName,
		"--image-project="+instImgPair[0],
		"--image-family="+instImgPair[1],
		"--machine-type="+m.config.InstanceType,
		fmt.Sprintf("--boot-disk-size=%dGiB", m.config.DiskGiB),
		"--metadata=block-project-ssh-keys=TRUE",
		m.instanceName,
	); err != nil {
		klog.Errorf("failed to start a gcp vm: %v", err)
		return err
	}

	return nil

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
	if err := m.executeGloudCommand(m.ctx, "compute", "instance", "add-metadata", m.instanceName, "--zone="+m.config.GCPZone); err != nil {
		klog.Warningf("failed to add metadata: %v", err)
	}
	// update the local ssh config
	if err := m.executeGloudCommand(m.ctx, "compute", "config-ssh"); err != nil {
		klog.Warningf("failed to compute ssh: %v", err)
	}

	// check ssh connnectivity
	return nil
}

func (m *MiniTestBosKosDeployer) executeGloudCommand(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "gcloud", append([]string{"--project=" + m.gcpProject}, args...)...)
	klog.Infof("Executing: %v", cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run %v: %v", args, err)
	}
	return nil
}

func (m *MiniTestBosKosDeployer) sshConnectionCheck() error {
	addr, err := m.SSHAddr()
	if err != nil {
		return fmt.Errorf("failed to get ssh addr:%v", err)
	}
	for i := range 10 {
		//  cmd cannot be reused after its failure
		cmd := exec.CommandContext(m.ctx, addr,
			"-o",
			"StrictHostKeyChecking=no",
			"-o",
			"User="+m.remoteUserName,
			"--", "uname", "-a")
		klog.Infof("executing %v", cmd.Args)
		
		if err = cmd.Run(); err == nil {
			return nil
		}
		klog.Infof("[%d/10]ssh command failed with error: %v, command: %v", i, err, cmd)
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("failed to connect to vm: %v", err)
}
