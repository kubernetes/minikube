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

package cluster

import (
	"bytes"
	gflag "flag"
	"fmt"
	"net"
	"strings"
	"text/template"

	"k8s.io/minikube/pkg/minikube/constants"
)

// Kill any running instances.

var localkubeStartCmdTemplate = "/usr/local/bin/localkube {{.Flags}} --generate-certs=false --logtostderr=true --enable-dns=false --node-ip={{.NodeIP}}"

var startCommandB2DTemplate = `
# Run with nohup so it stays up. Redirect logs to useful places.
sudo sh -c 'PATH=/usr/local/sbin:$PATH nohup {{.LocalkubeStartCmd}} > {{.Stdout}} 2> {{.Stderr}} < /dev/null & echo $! > {{.Pidfile}} &'
`

var localkubeSystemdTmpl = `[Unit]
Description=Localkube
Documentation=https://github.com/kubernetes/minikube/tree/master/pkg/localkube

Wants=network-online.target
After=network-online.target

[Service]
Type=notify
Restart=always
RestartSec=3

ExecStart={{.LocalkubeStartCmd}}

ExecReload=/bin/kill -s HUP $MAINPID

[Install]
WantedBy=multi-user.target
`

var startCommandTemplate = `
if which systemctl 2>&1 1>/dev/null; then
  {{.StartCommandSystemd}}
  sudo systemctl daemon-reload
  sudo systemctl enable localkube.service
  sudo systemctl restart localkube.service || true
else
  sudo killall localkube || true
  {{.StartCommandB2D}}
fi
`

func GetStartCommand(kubernetesConfig KubernetesConfig) (string, error) {
	localkubeStartCommand, err := GenLocalkubeStartCmd(kubernetesConfig)
	if err != nil {
		return "", err
	}
	startCommandB2D, err := GetStartCommandB2D(kubernetesConfig, localkubeStartCommand)
	if err != nil {
		return "", err
	}
	startCommandSystemd, err := GetStartCommandSystemd(kubernetesConfig, localkubeStartCommand)
	if err != nil {
		return "", err
	}
	t := template.Must(template.New("startCommand").Parse(startCommandTemplate))
	buf := bytes.Buffer{}
	data := struct {
		StartCommandB2D     string
		StartCommandSystemd string
	}{
		StartCommandB2D:     startCommandB2D,
		StartCommandSystemd: startCommandSystemd,
	}
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetStartCommandB2D(kubernetesConfig KubernetesConfig, localkubeStartCmd string) (string, error) {
	t := template.Must(template.New("startCommand").Parse(startCommandB2DTemplate))
	buf := bytes.Buffer{}
	data := struct {
		LocalkubeStartCmd string
		Stdout            string
		Stderr            string
		Pidfile           string
	}{
		LocalkubeStartCmd: localkubeStartCmd,
		Stdout:            constants.RemoteLocalKubeOutPath,
		Stderr:            constants.RemoteLocalKubeErrPath,
		Pidfile:           constants.LocalkubePIDPath,
	}
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetStartCommandSystemd(kubernetesConfig KubernetesConfig, localkubeStartCmd string) (string, error) {
	t, err := template.New("localkubeConfig").Parse(localkubeSystemdTmpl)
	if err != nil {
		return "", err
	}
	buf := bytes.Buffer{}
	data := struct {
		LocalkubeStartCmd string
	}{
		LocalkubeStartCmd: localkubeStartCmd,
	}
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return fmt.Sprintf("printf %%s \"%s\" | sudo tee %s", buf.String(),
		constants.LocalkubeServicePath), nil
}

func GenLocalkubeStartCmd(kubernetesConfig KubernetesConfig) (string, error) {
	flagVals := make([]string, len(constants.LogFlags))
	for _, logFlag := range constants.LogFlags {
		if logVal := gflag.Lookup(logFlag); logVal != nil && logVal.Value.String() != logVal.DefValue {
			flagVals = append(flagVals, fmt.Sprintf("--%s %s", logFlag, logVal.Value.String()))
		}
	}

	if kubernetesConfig.ContainerRuntime != "" {
		flagVals = append(flagVals, "--container-runtime="+kubernetesConfig.ContainerRuntime)
	}

	if kubernetesConfig.NetworkPlugin != "" {
		flagVals = append(flagVals, "--network-plugin="+kubernetesConfig.NetworkPlugin)
	}

	if kubernetesConfig.FeatureGates != "" {
		flagVals = append(flagVals, "--feature-gates="+kubernetesConfig.FeatureGates)
	}

	for _, e := range kubernetesConfig.ExtraOptions {
		flagVals = append(flagVals, fmt.Sprintf("--extra-config=%s", e.String()))
	}
	flags := strings.Join(flagVals, " ")

	t := template.Must(template.New("localkubeStartCmd").Parse(localkubeStartCmdTemplate))
	buf := bytes.Buffer{}
	data := struct {
		Flags  string
		NodeIP string
	}{
		Flags:  flags,
		NodeIP: kubernetesConfig.NodeIP,
	}
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var logsCommand = fmt.Sprintf(`
if which systemctl 2>&1 1>/dev/null; then
  sudo journalctl -u localkube
else
  tail -n +1 %s %s
fi
`, constants.RemoteLocalKubeErrPath, constants.RemoteLocalKubeOutPath)

var localkubeStatusCommand = fmt.Sprintf(`
if which systemctl 2>&1 1>/dev/null; then
  sudo systemctl is-active localkube 2>&1 1>/dev/null && echo "Running" || echo "Stopped"
else
  if ps $(cat %s) 2>&1 1>/dev/null; then
    echo "Running"
  else
    echo "Stopped"
  fi
fi
`, constants.LocalkubePIDPath)

func GetMount9pCommand(ip net.IP) string {
	return fmt.Sprintf(`
sudo mkdir /mount-9p;
sudo mount -t 9p -o trans=tcp -o port=5640 %s /mount-9p;
sudo chmod 775 /mount-9p;`, ip)
}
