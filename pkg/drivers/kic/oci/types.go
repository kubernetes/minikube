/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
)

const (
	// DefaultBindIPV4 is The default IP the container will listen on.
	DefaultBindIPV4 = "127.0.0.1"
	// Docker is docker
	Docker = "docker"
	// Podman is podman
	Podman = "podman"
	// ProfileLabelKey is applied to any container or volume created by a specific minikube profile name.minikube.sigs.k8s.io=PROFILE_NAME
	ProfileLabelKey = "name.minikube.sigs.k8s.io"
	// NodeLabelKey is applied to each volume so it can be referred to by name
	NodeLabelKey = "mode.minikube.sigs.k8s.io"
	// NodeRoleKey is used to identify if it is control plane or worker
	nodeRoleLabelKey = "role.minikube.sigs.k8s.io"
	// CreatedByLabelKey is applied to any container/volume that is created by minikube created_by.minikube.sigs.k8s.io=true
	CreatedByLabelKey = "created_by.minikube.sigs.k8s.io"
)

// CreateParams are parameters needed to create a container
type CreateParams struct {
	ClusterName   string            // cluster(profile name) that this container belongs to
	Name          string            // used for container name and hostname
	Image         string            // container image to use to create the node.
	ClusterLabel  string            // label the clusters we create using minikube so we can clean up
	NodeLabel     string            // label the nodes so we can clean up by node name
	Role          string            // currently only role supported is control-plane
	Mounts        []Mount           // volume mounts
	APIServerPort int               // Kubernetes api server port
	PortMappings  []PortMapping     // ports to map to container from host
	CPUs          string            // number of cpu cores assign to container
	Memory        string            // memory (mbs) to assign to the container
	Envs          map[string]string // environment variables to pass to the container
	ExtraArgs     []string          // a list of any extra option to pass to oci binary during creation time, for example --expose 8080...
	OCIBinary     string            // docker or podman
	Network       string            // network name that the container will attach to
	IP            string            // static IP to assign for th container in the cluster network
}

// createOpt is an option for Create
type createOpt func(*createOpts) *createOpts

// actual options struct
type createOpts struct {
	RunArgs       []string
	ContainerArgs []string
	Mounts        []Mount
	PortMappings  []PortMapping
}

/*
These types are from
https://github.com/kubernetes/kubernetes/blob/063e7ff358fdc8b0916e6f39beedc0d025734cb1/pkg/kubelet/apis/cri/runtime/v1alpha2/api.pb.go#L183
*/

// Mount specifies a host volume to mount into a container.
// This is a close copy of the upstream cri Mount type
// see: k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2
// It additionally serializes the "propagation" field with the string enum
// names on disk as opposed to the int32 values, and the serlialzed field names
// have been made closer to core/v1 VolumeMount field names
// In yaml this looks like:
//  containerPath: /foo
//  hostPath: /bar
//  readOnly: true
//  selinuxRelabel: false
//  propagation: None
// Propagation may be one of: None, HostToContainer, Bidirectional
type Mount struct {
	// Path of the mount within the container.
	ContainerPath string `protobuf:"bytes,1,opt,name=container_path,json=containerPath,proto3" json:"containerPath,omitempty"`
	// Path of the mount on the host. If the hostPath doesn't exist, then runtimes
	// should report error. If the hostpath is a symbolic link, runtimes should
	// follow the symlink and mount the real destination to container.
	HostPath string `protobuf:"bytes,2,opt,name=host_path,json=hostPath,proto3" json:"hostPath,omitempty"`
	// If set, the mount is read-only.
	Readonly bool `protobuf:"varint,3,opt,name=readonly,proto3,json=readOnly,proto3" json:"readOnly,omitempty"`
	// If set, the mount needs SELinux relabeling.
	SelinuxRelabel bool `protobuf:"varint,4,opt,name=selinux_relabel,json=selinuxRelabel,proto3" json:"selinuxRelabel,omitempty"`
	// Requested propagation mode.
	Propagation MountPropagation `protobuf:"varint,5,opt,name=propagation,proto3,enum=runtime.v1alpha2.MountPropagation" json:"propagation,omitempty"`
}

// ParseMountString parses a mount string of format:
// '[host-path:]container-path[:<options>]' The comma-delimited 'options' are
// [rw|ro], [Z], [srhared|rslave|rprivate].
func ParseMountString(spec string) (m Mount, err error) {
	f := strings.Split(spec, ":")
	fields := f
	// suppressing err is safe here since the regex will always compile
	windows, _ := regexp.MatchString(`^[A-Z]:\\*`, spec)
	if windows {
		// Recreate the host path that got split above since
		// Windows paths look like C:\path
		hpath := fmt.Sprintf("%s:%s", f[0], f[1])
		fields = []string{hpath}
		fields = append(fields, f[2:]...)
	}
	switch len(fields) {
	case 0:
		err = errors.New("invalid empty spec")
	case 1:
		m.ContainerPath = fields[0]
	case 3:
		for _, opt := range strings.Split(fields[2], ",") {
			switch opt {
			case "Z":
				m.SelinuxRelabel = true
			case "ro":
				m.Readonly = true
			case "rw":
				m.Readonly = false
			case "rslave":
				m.Propagation = MountPropagationHostToContainer
			case "rshared":
				m.Propagation = MountPropagationBidirectional
			case "private":
				m.Propagation = MountPropagationNone
			default:
				err = fmt.Errorf("unknown mount option: '%s'", opt)
			}
		}
		fallthrough
	case 2:
		m.HostPath, m.ContainerPath = fields[0], fields[1]
		if !path.IsAbs(m.ContainerPath) {
			err = fmt.Errorf("'%s' container path must be absolute", m.ContainerPath)
		}
	default:
		err = errors.New("spec must be in form: <host path>:<container path>[:<options>]")
	}
	return m, err
}

// PortMapping specifies a host port mapped into a container port.
// In yaml this looks like:
//  containerPort: 80
//  hostPort: 8000
//  listenAddress: 127.0.0.1
type PortMapping struct {
	// Port within the container.
	ContainerPort int32 `protobuf:"varint,1,opt,name=container_port,json=containerPort,proto3" json:"containerPort,omitempty"`
	// Port on the host.
	HostPort      int32  `protobuf:"varint,2,opt,name=host_path,json=hostPort,proto3" json:"hostPort,omitempty"`
	ListenAddress string `protobuf:"bytes,3,opt,name=listenAddress,json=hostPort,proto3" json:"listenAddress,omitempty"`
}

// MountPropagation represents an "enum" for mount propagation options,
// see also Mount.
type MountPropagation int32

const (
	// MountPropagationNone specifies that no mount propagation
	// ("private" in Linux terminology).
	MountPropagationNone MountPropagation = 0
	// MountPropagationHostToContainer specifies that mounts get propagated
	// from the host to the container ("rslave" in Linux).
	MountPropagationHostToContainer MountPropagation = 1
	// MountPropagationBidirectional specifies that mounts get propagated from
	// the host to the container and from the container to the host
	// ("rshared" in Linux).
	MountPropagationBidirectional MountPropagation = 2
)

// MountPropagationValueToName is a map of valid MountPropagation values to
// their string names
var MountPropagationValueToName = map[MountPropagation]string{
	MountPropagationNone:            "None",
	MountPropagationHostToContainer: "HostToContainer",
	MountPropagationBidirectional:   "Bidirectional",
}

// MountPropagationNameToValue is a map of valid MountPropagation names to
// their values
var MountPropagationNameToValue = map[string]MountPropagation{
	"None":            MountPropagationNone,
	"HostToContainer": MountPropagationHostToContainer,
	"Bidirectional":   MountPropagationBidirectional,
}
