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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/phayes/freeport"
	gossh "golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
)

var sshSetupScript = `
#!/bin/bash
USERNAME=%s
apt update
apt install -y openssh-server rsync
service ssh start
useradd -m -s /bin/bash ${USERNAME}
passwd -d  ${USERNAME}

USER_HOME=$(eval echo "~${USERNAME}")
mkdir -p $USER_HOME/.ssh
chmod 700 $USER_HOME/.ssh
echo "%s" >> $USER_HOME/.ssh/authorized_keys
chmod 600 $USER_HOME/.ssh/authorized_keys
chown -R $USERNAME:$USERNAME "$USER_HOME/.ssh"
`

type MiniTestDockerDeployer struct {
	ctx    context.Context
	config *MiniTestDockerConfig
	isUp   bool

	dockerClient *client.Client
	containerSHA string

	sshPrivateKeyFile   string
	sshPublicKeyFile    string
	sshPublicKeyContent string
	sshPort             string

	sshTempDir     string
	remoteUserName string
}

func NewMiniTestDockerDeployerFromConfigFile(path string) MiniTestDeployer {
	config := MiniTestDockerConfig{}
	data, err := os.ReadFile(path)
	if err != nil {
		klog.Fatalf("failed to read config file %s: %v", path, err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		klog.Fatalf("failed to parse config file %s: %v", path, err)
	}
	return NewMiniTestDockerDeployer(&config)

}

func NewMiniTestDockerDeployer(config *MiniTestDockerConfig) MiniTestDeployer {
	return &MiniTestDockerDeployer{
		ctx:            context.TODO(),
		config:         config,
		isUp:           false,
		remoteUserName: remoteUserName,
	}
}

func (m *MiniTestDockerDeployer) Up() error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker api client: %v", err)
	}
	m.dockerClient = dockerClient

	// pull the image
	reader, err := dockerClient.ImagePull(m.ctx, m.config.Image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %v", m.config.Image, err)
	}
	defer reader.Close()
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return fmt.Errorf("failed to read image pull response: %v", err)
	}
	// find a free port for ssh
	port, err := freeport.GetFreePort()
	if err != nil {
		return fmt.Errorf("failed to get a free port for ssh: %v", err)
	}
	m.sshPort = strconv.Itoa(port)
	klog.Infof("Using port %s for ssh", m.sshPort)

	// start the container
	id := uuid.New().String()[:8]
	//todo: remove hard coded ports
	exposedPorts, portBindings, _ := nat.ParsePortSpecs([]string{m.sshPort + ":22"})
	response, err := dockerClient.ContainerCreate(
		m.ctx,
		&container.Config{
			Image:     m.config.Image,
			Tty:       true, //-t
			OpenStdin: true, // -i
			//Entrypoint:   []string{"/usr/sbin/init"},
			ExposedPorts: exposedPorts,
		},

		&container.HostConfig{
			Privileged:   true,
			PortBindings: portBindings,
		}, nil, nil, "minitest-"+id)
	if err != nil {
		klog.Errorf("failed to create container from image %s: %v", m.config.Image, err)
		return fmt.Errorf("failed to create container from image %s: %v", m.config.Image, err)
	}
	m.containerSHA = response.ID

	err = dockerClient.ContainerStart(m.ctx, m.containerSHA,
		client.ContainerStartOptions{})
	if err != nil {
		klog.Errorf("failed to start container %s: %v", m.containerSHA, err)
		return fmt.Errorf("failed to start container %s: %v", m.containerSHA, err)
	}

	// set up ssh keys
	if err := m.sshSetUp(); err != nil {
		klog.Errorf("failed to set up ssh: %v", err)
		return fmt.Errorf("failed to set up ssh: %v", err)
	}

	// set up sshd server
	err = m.executeDockerShellCommand("root", fmt.Sprintf(sshSetupScript, m.remoteUserName, m.sshPublicKeyContent))
	if err != nil {
		klog.Errorf("failed to set up sshd server in container %s: %v", m.containerSHA, err)
		return fmt.Errorf("failed to set up sshd server in container %s: %v", m.containerSHA, err)
	}

	// check ssh connectivity
	if err := sshConnectionCheck(m.ctx, m.remoteUserName, "localhost", m.sshAdditionalArgs()); err != nil {
		klog.Errorf("Failed to connect via ssh: %v", err)
		return fmt.Errorf("Failed to connect via ssh: %v", err)
	}

	m.isUp = true
	return nil
}

func (m *MiniTestDockerDeployer) Down() error {
	os.RemoveAll(m.sshTempDir)

	if m.dockerClient == nil {
		klog.Errorf("m.dockerClient not set")
		return fmt.Errorf("m.dockerClient not set")
	}
	if err := m.dockerClient.ContainerRemove(m.ctx, m.containerSHA, client.ContainerRemoveOptions{Force: true}); err != nil {
		klog.Errorf("failed to remove container %s: %v", m.containerSHA, err)
		return fmt.Errorf("failed to remove container %s: %v", m.containerSHA, err)
	}
	// close the docker client
	m.dockerClient.Close()
	m.isUp = false
	klog.Infof("Successfully removed container %s", m.containerSHA)
	return nil
}
func (m *MiniTestDockerDeployer) IsUp() (bool, error) {
	return m.isUp, nil
}

func (m *MiniTestDockerDeployer) Execute(args ...string) error {
	return executeSSHCommand(m.ctx, m.remoteUserName, "localhost", m.sshAdditionalArgs(), args...)
}

func (m *MiniTestDockerDeployer) SyncToRemote(src string, dst string, excludedPattern []string) error {
	excludedArgs := make([]string, 0, len(excludedPattern)*2)
	for _, pattern := range excludedPattern {
		excludedArgs = append(excludedArgs, "--exclude", pattern)
	}
	dstRemote := fmt.Sprintf("%s@%s:%s", m.remoteUserName, "localhost", dst)
	return executeRsyncSSHCommand(m.ctx, m.sshAdditionalArgs(), src, dstRemote, excludedArgs)
}

func (m *MiniTestDockerDeployer) SyncToHost(src string, dst string, excludedPattern []string) error {
	excludedArgs := make([]string, 0, len(excludedPattern)*2)
	for _, pattern := range excludedPattern {
		excludedArgs = append(excludedArgs, "--exclude", pattern)
	}
	srcRemote := fmt.Sprintf("%s@%s:%s", m.remoteUserName, "localhost", src)
	return executeRsyncSSHCommand(m.ctx, m.sshAdditionalArgs(), srcRemote, dst, excludedArgs)
	//return executeScpCommand(m.ctx, m.remoteUserName, "localhost", m.scpAdditionalArgs(), src, dst)
}

func (m *MiniTestDockerDeployer) executeDockerShellCommand(user string, args ...string) error {
	execResp, err := m.dockerClient.ContainerExecCreate(m.ctx, m.containerSHA, container.ExecOptions{
		Cmd:          append([]string{"/bin/bash", "-c"}, args...),
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		User:         user,
	})
	if err != nil {
		return fmt.Errorf("failed to create exec instance: %v", err)
	}

	attachResp, err := m.dockerClient.ContainerExecAttach(
		m.ctx, execResp.ID,
		container.ExecAttachOptions{Tty: true})
	if err != nil {
		return fmt.Errorf("failed to attach to exec instance: %v", err)
	}
	defer attachResp.Close()
	if _, err = io.Copy(os.Stdout, attachResp.Reader); err != nil {
		return fmt.Errorf("failed to read exec output: %v", err)
	}
	return nil

}

func (m *MiniTestDockerDeployer) sshSetUp() error {
	// we generate a ssh key pair and add the public key to the container's authorized_keys

	sshTempDir, err := os.MkdirTemp("", "minitest-ssh-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for ssh keys: %v", err)
	}
	m.sshTempDir = sshTempDir
	klog.Info("Created temp dir for ssh keys: ", sshTempDir)

	// create private key file
	sshPrivateKeyFile, err := os.CreateTemp(sshTempDir, "id_rsa")
	if err != nil {
		return fmt.Errorf("failed to create temp file for private key: %v", err)
	}
	if sshPrivateKeyFile.Chmod(0600) != nil {
		return fmt.Errorf("failed to chmod private key file: %v", err)
	}
	m.sshPrivateKeyFile = sshPrivateKeyFile.Name()
	klog.Info("Created temp file for private key: ", m.sshPrivateKeyFile)

	// create public key file
	sshPublicKeyFile, err := os.CreateTemp(sshTempDir, "id_rsa.pub")
	if err != nil {
		return fmt.Errorf("failed to create temp file for public key: %v", err)
	}
	if sshPublicKeyFile.Chmod(0644) != nil {
		return fmt.Errorf("failed to chmod public key file: %v", err)
	}
	m.sshPublicKeyFile = sshPublicKeyFile.Name()
	klog.Info("Created temp file for public key: ", m.sshPublicKeyFile)

	// generate private key and convert to PEM format
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	}

	if err := pem.Encode(sshPrivateKeyFile, privBlock); err != nil {
		sshPrivateKeyFile.Close()
		return fmt.Errorf("failed to write private key to file: %v", err)
	}

	pub, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to generate public key: %v", err)
	}
	pubBytes := gossh.MarshalAuthorizedKey(pub)

	if _, err := sshPublicKeyFile.Write(pubBytes); err != nil {
		return fmt.Errorf("failed to write public key to file: %v", err)
	}
	sshPublicKeyFile.Close()
	m.sshPublicKeyContent = string(pubBytes)
	klog.Infof("Generated ssh public key:%s ", m.sshPublicKeyContent)
	return nil
}

func (m *MiniTestDockerDeployer) sshAdditionalArgs() []string {
	return []string{"-i", m.sshPrivateKeyFile, "-p", m.sshPort}
}

func (m *MiniTestDockerDeployer) scpAdditionalArgs() []string {
	return []string{"-i", m.sshPrivateKeyFile, "-P", m.sshPort}
}
