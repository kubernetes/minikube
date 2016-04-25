package mcndockerclient

import (
	"fmt"

	"github.com/docker/machine/libmachine/cert"
	"github.com/samalba/dockerclient"
)

// DockerClient creates a docker client for a given host.
func DockerClient(dockerHost DockerHost) (*dockerclient.DockerClient, error) {
	url, err := dockerHost.URL()
	if err != nil {
		return nil, err
	}

	tlsConfig, err := cert.ReadTLSConfig(url, dockerHost.AuthOptions())
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLS config: %s", err)
	}

	return dockerclient.NewDockerClient(url, tlsConfig)
}

// CreateContainer creates a docker container.
func CreateContainer(dockerHost DockerHost, config *dockerclient.ContainerConfig, name string) error {
	docker, err := DockerClient(dockerHost)
	if err != nil {
		return err
	}

	if err = docker.PullImage(config.Image, nil); err != nil {
		return fmt.Errorf("Unable to pull image: %s", err)
	}

	var authConfig *dockerclient.AuthConfig
	containerID, err := docker.CreateContainer(config, name, authConfig)
	if err != nil {
		return fmt.Errorf("Error while creating container: %s", err)
	}

	if err = docker.StartContainer(containerID, &config.HostConfig); err != nil {
		return fmt.Errorf("Error while starting container: %s", err)
	}

	return nil
}
