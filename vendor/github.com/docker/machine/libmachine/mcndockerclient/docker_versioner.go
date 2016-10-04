package mcndockerclient

import "fmt"

var CurrentDockerVersioner DockerVersioner = &defaultDockerVersioner{}

type DockerVersioner interface {
	DockerVersion(host DockerHost) (string, error)
}

func DockerVersion(host DockerHost) (string, error) {
	return CurrentDockerVersioner.DockerVersion(host)
}

type defaultDockerVersioner struct{}

func (dv *defaultDockerVersioner) DockerVersion(host DockerHost) (string, error) {
	client, err := DockerClient(host)
	if err != nil {
		return "", fmt.Errorf("Unable to query docker version: %s", err)
	}

	version, err := client.Version()
	if err != nil {
		return "", fmt.Errorf("Unable to query docker version: %s", err)
	}

	return version.Version, nil
}
