/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package ktmpl

import "text/template"

// KubeletSystemdTemplate hosts the override kubelet flags, written to kubeletSystemdConfFile
var KubeletSystemdTemplate = template.Must(template.New("kubeletSystemdTemplate").Parse(`[Unit]
{{if or (eq .ContainerRuntime "cri-o") (eq .ContainerRuntime "cri")}}Wants=crio.service{{else if eq .ContainerRuntime "containerd"}}Wants=containerd.service{{else}}Wants=docker.socket{{end}}

[Service]
ExecStart=
ExecStart={{.KubeletPath}}{{if .ExtraOptions}} {{.ExtraOptions}}{{end}}

[Install]
`))

// KubeletServiceTemplate is the base kubelet systemd template, written to kubeletServiceFile
var KubeletServiceTemplate = template.Must(template.New("kubeletServiceTemplate").Parse(`[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=http://kubernetes.io/docs/
StartLimitIntervalSec=0

[Service]
ExecStart={{.KubeletPath}}
Restart=always
# Tuned for local dev: faster than upstream default (10s), but slower than systemd default (100ms)
RestartSec=600ms

[Install]
WantedBy=multi-user.target
`))
