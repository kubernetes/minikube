// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"encoding/json"
	"io"
	"time"
)

// ConfigFile is the configuration file that holds the metadata describing
// how to launch a container.  The names of the fields are chosen to reflect
// the JSON payload of the ConfigFile as defined here: https://git.io/vrAEY
type ConfigFile struct {
	Architecture    string    `json:"architecture"`
	Container       string    `json:"container"`
	Created         time.Time `json:"created"`
	DockerVersion   string    `json:"docker_version"`
	History         []History `json:"history"`
	OS              string    `json:"os"`
	RootFS          RootFS    `json:"rootfs"`
	Config          Config    `json:"config"`
	ContainerConfig Config    `json:"container_config"`
	OSVersion       string    `json:"osversion"`
}

// History is one entry of a list recording how this container image was built.
type History struct {
	Author     string    `json:"author"`
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"created_by"`
	Comment    string    `json:"comment"`
	EmptyLayer bool      `json:"empty_layer,omitempty"`
}

// RootFS holds the ordered list of file system deltas that comprise the
// container image's root filesystem.
type RootFS struct {
	Type    string `json:"type"`
	DiffIDs []Hash `json:"diff_ids"`
}

// Config is a submessage of the config file described as:
//   The execution parameters which SHOULD be used as a base when running
//   a container using the image.
// The names of the fields in this message are chosen to reflect the JSON
// payload of the Config as defined here:
// https://git.io/vrAET
// and
// https://github.com/opencontainers/image-spec/blob/master/config.md
type Config struct {
	AttachStderr    bool
	AttachStdin     bool
	AttachStdout    bool
	Cmd             []string
	Domainname      string
	Entrypoint      []string
	End             []string
	Hostname        string
	Image           string
	Labels          map[string]string
	OnBuild         []string
	OpenStdin       bool
	StdinOnce       bool
	Tty             bool
	User            string
	Volumes         map[string]struct{}
	WorkingDir      string
	ExposedPorts    map[string]struct{}
	ArgsEscaped     bool
	NetworkDisabled bool
	MacAddress      string
	StopSignal      string
	Shell           []string
}

// ParseConfigFile parses the io.ReadCloser's contents into a ConfigFile.
func ParseConfigFile(r io.ReadCloser) (*ConfigFile, error) {
	defer r.Close()
	cf := ConfigFile{}
	if err := json.NewDecoder(r).Decode(&cf); err != nil {
		return nil, err
	}
	return &cf, nil
}
