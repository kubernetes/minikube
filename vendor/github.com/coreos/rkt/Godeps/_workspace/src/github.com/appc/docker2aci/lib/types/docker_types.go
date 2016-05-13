// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import "time"

// DockerImageData stores the JSON structure of a Docker image.
// Taken and adapted from upstream Docker.
type DockerImageData struct {
	ID              string             `json:"id"`
	Parent          string             `json:"parent,omitempty"`
	Comment         string             `json:"comment,omitempty"`
	Created         time.Time          `json:"created"`
	Container       string             `json:"container,omitempty"`
	ContainerConfig DockerImageConfig  `json:"container_config,omitempty"`
	DockerVersion   string             `json:"docker_version,omitempty"`
	Author          string             `json:"author,omitempty"`
	Config          *DockerImageConfig `json:"config,omitempty"`
	Architecture    string             `json:"architecture,omitempty"`
	OS              string             `json:"os,omitempty"`
	Checksum        string             `json:"checksum"`
}

// Note: the Config structure should hold only portable information about the container.
// Here, "portable" means "independent from the host we are running on".
// Non-portable information *should* appear in HostConfig.
// Taken and adapted from upstream Docker.
type DockerImageConfig struct {
	Hostname        string
	Domainname      string
	User            string
	Memory          int64  // Memory limit (in bytes)
	MemorySwap      int64  // Total memory usage (memory + swap); set `-1' to disable swap
	CpuShares       int64  // CPU shares (relative weight vs. other containers)
	Cpuset          string // Cpuset 0-2, 0,1
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	PortSpecs       []string // Deprecated - Can be in the format of 8080/tcp
	ExposedPorts    map[string]struct{}
	Tty             bool // Attach standard streams to a tty, including stdin if it is not closed.
	OpenStdin       bool // Open stdin
	StdinOnce       bool // If true, close stdin after the 1 attached client disconnects.
	Env             []string
	Cmd             []string
	Image           string // Name of the image as it was passed by the operator (eg. could be symbolic)
	Volumes         map[string]struct{}
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
	MacAddress      string
	OnBuild         []string
}

// Taken from upstream Docker
type DockerAuthConfig struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Auth          string `json:"auth"`
	Email         string `json:"email"`
	ServerAddress string `json:"serveraddress,omitempty"`
}
