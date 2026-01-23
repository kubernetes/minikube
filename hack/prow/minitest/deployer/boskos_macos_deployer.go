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
	"os"
	"time"

	"k8s.io/klog/v2"
	"sigs.k8s.io/boskos/client"
	"sigs.k8s.io/kubetest2/pkg/boskos"
)

const (
	macInstanceResourceType = "mac-instances"
	macRemoteUserName       = "ec2-user"
	sshPrivateKeyPath       = "/etc/aws-ssh/aws-ssh-private"
)

type MiniTestBosKosMacOSDeployer struct {
	ctx    context.Context
	config *MiniTestBoskosConfig
	isUp   bool

	remoteUserName string
	sshAddr        string

	boskosClient *client.Client
	// this channel serves as a signal channel for the hearbeat goroutine
	// so that it can be explicitly closed
	boskosHeartbeatClose chan struct{}
}

func NewMiniTestBosKosMacOSDeployerFromConfigFile(path string) MiniTestDeployer {
	config := MiniTestBoskosConfig{}
	data, err := os.ReadFile(path)
	if err != nil {
		klog.Fatalf("failed to read config file %s: %v", path, err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		klog.Fatalf("failed to parse config file %s: %v", path, err)
	}
	return NewMiniTestBosKosMacOSDeployer(&config)
}

func NewMiniTestBosKosMacOSDeployer(config *MiniTestBoskosConfig) MiniTestDeployer {
	boskosClient, err := boskos.NewClient(config.BoskosLocation)
	if err != nil {
		klog.Fatalf("failed to make boskos client: %v", err)
	}
	return &MiniTestBosKosMacOSDeployer{
		ctx:                  context.TODO(),
		config:               config,
		boskosClient:         boskosClient,
		boskosHeartbeatClose: make(chan struct{}),
		remoteUserName:       macRemoteUserName,
	}
}

func (m *MiniTestBosKosMacOSDeployer) Up() error {
	if err := m.requestMacOSInstance(); err != nil {
		klog.Errorf("Failed to request macos instance from boskos: %v", err)
		return err
	}
	fmt.Printf("get instance %s\n", m.sshAddr)

	if err := sshConnectionCheck(m.ctx, m.remoteUserName, m.sshAddr, []string{"-i", sshPrivateKeyPath}); err != nil {
		klog.Errorf("Failed to conntect via ssh: %v", err)
		return err
	}

	klog.Infof("Successfully connected to macos instance: %s", m.sshAddr)
	m.isUp = true
	return nil

}

func (m *MiniTestBosKosMacOSDeployer) Down() error {
	//todo: clean up the VM?
	klog.Info("Releasing bosko macos instance")
	if m.boskosClient == nil {
		return fmt.Errorf("m.boskosClient not set")
	}
	err := boskos.Release(
		m.boskosClient,
		[]string{m.sshAddr},
		m.boskosHeartbeatClose,
	)
	if err != nil {
		fmt.Printf("Error releasing boskos macos instance: %v\n", err)
		//return fmt.Errorf("down failed to release boskos macos instance: %v", err)
	}
	m.isUp = false
	return nil

}

func (m *MiniTestBosKosMacOSDeployer) IsUp() (bool, error) {
	return m.isUp, nil
}
func (m *MiniTestBosKosMacOSDeployer) Execute(args ...string) error {
	return executeSSHCommand(m.ctx, m.remoteUserName, m.sshAddr, m.additionalSSHArgs(), args...)
}

func (m *MiniTestBosKosMacOSDeployer) SyncToRemote(src string, dst string, excludedPattern []string) error {
	excludedArgs := make([]string, 0, len(excludedPattern)*2)
	for _, pattern := range excludedPattern {
		excludedArgs = append(excludedArgs, "--exclude", pattern)
	}
	dstRemote := fmt.Sprintf("%s@%s:%s", m.remoteUserName, m.sshAddr, dst)
	return executeRsyncSSHCommand(m.ctx, m.additionalSSHArgs(), src, dstRemote, excludedArgs)
}

func (m *MiniTestBosKosMacOSDeployer) SyncToHost(src string, dst string, excludedPattern []string) error {
	excludedArgs := make([]string, 0, len(excludedPattern)*2)
	for _, pattern := range excludedPattern {
		excludedArgs = append(excludedArgs, "--exclude", pattern)
	}
	srcRemote := fmt.Sprintf("%s@%s:%s", m.remoteUserName, m.sshAddr, src)
	return executeRsyncSSHCommand(m.ctx, m.additionalSSHArgs(), srcRemote, dst, excludedArgs)
}

func (m *MiniTestBosKosMacOSDeployer) requestMacOSInstance() error {

	resource, err := boskos.Acquire(
		m.boskosClient,
		macInstanceResourceType,
		time.Duration(m.config.BoskosAcquireTimeoutSeconds)*time.Second,
		time.Duration(m.config.BoskosHeartbeatIntervalSeconds)*time.Second,
		m.boskosHeartbeatClose,
	)

	if err != nil {
		fmt.Printf("failed to get macos instance from boskos: %v\n", err)
		m.sshAddr = "28zmx-sibu3-yy3oc-zmvxf-smpwu-058cv95.us-east-2.ip.aws"
		return nil
		//return fmt.Errorf("failed to get macos instance from boskos: %v", err)
	}
	if resource.Name == "" {
		fmt.Printf("boskos returned an empty resource name, resource: %v\n", resource)
		m.sshAddr = "28zmx-sibu3-yy3oc-zmvxf-smpwu-058cv95.us-east-2.ip.aws"
		return nil
		//return fmt.Errorf("boskos returned an empty resource name, resource: %v", resource)
	}

	klog.Infof("Got instance %q from boskos", resource.Name)
	m.sshAddr = resource.Name
	return nil
}
func (m *MiniTestBosKosMacOSDeployer) additionalSSHArgs() []string {
	return []string{"-i", sshPrivateKeyPath}
}
