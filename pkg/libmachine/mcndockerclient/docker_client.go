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
	"net/http"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
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

	return client.New(
		client.WithHost(url),
		client.WithHTTPClient(httpClient),
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
	out, err := docker.ImagePull(ctx, config.Image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("Unable to pull image: %s", err)
	}
	defer out.Close()
	if err := out.Wait(ctx); err != nil {
		return fmt.Errorf("Unable to pull image: %s", err)
	}

	resp, err := docker.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     config,
		HostConfig: hostConfig,
		Name:       name,
	})
	if err != nil {
		return fmt.Errorf("Error while creating container: %s", err)
	}

	if _, err = docker.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Error while starting container: %s", err)
	}

	return nil
}
