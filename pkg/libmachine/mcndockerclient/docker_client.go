/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package mcndockerclient

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"k8s.io/minikube/pkg/libmachine/cert"
)

// DockerClient creates a docker client for a given host.
func DockerClient(dockerHost DockerHost) (*client.Client, error) {
	url, err := dockerHost.URL()
	if err != nil {
		return nil, err
	}

	tlsConfig, err := cert.ReadTLSConfig(url, dockerHost.AuthOptions())
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLS config: %s", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client.NewClientWithOpts(
		client.WithHost(url),
		client.WithHTTPClient(httpClient),
		client.WithAPIVersionNegotiation(),
	)
}

// CreateContainer creates a docker container.
func CreateContainer(dockerHost DockerHost, config *container.Config, hostConfig *container.HostConfig, name string) error {
	ctx := context.Background()
	docker, err := DockerClient(dockerHost)
	if err != nil {
		return err
	}
	defer docker.Close()

	// Pull plugin image
	out, err := docker.ImagePull(ctx, config.Image, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("Unable to pull image: %s", err)
	}
	defer out.Close()
	// consume output to ensure pull is finished
	_, _ = io.Copy(io.Discard, out)

	resp, err := docker.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return fmt.Errorf("Error while creating container: %s", err)
	}

	if err = docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("Error while starting container: %s", err)
	}

	return nil
}
