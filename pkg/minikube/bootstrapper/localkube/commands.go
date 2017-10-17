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

package localkube

import (
	"bytes"
	gflag "flag"
	"fmt"
	"strings"

	"text/template"

	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/constants"
)

// Kill any running instances.

var localkubeStartCmdTemplate = "/usr/local/bin/localkube {{.Flags}} --generate-certs=false --logtostderr=true --enable-dns=false"

var startCommandNoSystemdTemplate = `
# Run with nohup so it stays up. Redirect logs to useful places.
sudo sh -c 'PATH=/usr/local/sbin:$PATH GODEBUG=netdns=go nohup {{.LocalkubeStartCmd}} > {{.Stdout}} 2> {{.Stderr}} < /dev/null & echo $! > {{.Pidfile}} &'
`

var localkubeSystemdTmpl = `[Unit]
Description=Localkube
Documentation=https://github.com/kubernetes/minikube/tree/master/pkg/localkube

[Service]
Type=notify
Restart=always
RestartSec=3

Environment=GODEBUG=netdns=go

ExecStart={{.LocalkubeStartCmd}}

ExecReload=/bin/kill -s HUP $MAINPID

[Install]
WantedBy=multi-user.target
`

var startCommandTemplate = "if [[ `systemctl` =~ -\\.mount ]] &>/dev/null;" + `then
  {{.StartCommandSystemd}}
  sudo systemctl daemon-reload
  sudo systemctl enable localkube.service
  sudo systemctl restart localkube.service || true
else
  sudo killall localkube || true
  {{.StartCommandNoSystemd}}
fi
`

func GetStartCommand(kubernetesConfig bootstrapper.KubernetesConfig) (string, error) {
	localkubeStartCommand, err := GenLocalkubeStartCmd(kubernetesConfig)
	if err != nil {
		return "", err
	}
	startCommandNoSystemd, err := GetStartCommandNoSystemd(kubernetesConfig, localkubeStartCommand)
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
		StartCommandNoSystemd string
		StartCommandSystemd   string
	}{
		StartCommandNoSystemd: startCommandNoSystemd,
		StartCommandSystemd:   startCommandSystemd,
	}
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetStartCommandNoSystemd(kubernetesConfig bootstrapper.KubernetesConfig, localkubeStartCmd string) (string, error) {
	t := template.Must(template.New("startCommand").Parse(startCommandNoSystemdTemplate))
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

func GetStartCommandSystemd(kubernetesConfig bootstrapper.KubernetesConfig, localkubeStartCmd string) (string, error) {
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

func GenLocalkubeStartCmd(kubernetesConfig bootstrapper.KubernetesConfig) (string, error) {
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

	if kubernetesConfig.APIServerName != constants.APIServerName {
		flagVals = append(flagVals, "--apiserver-name="+kubernetesConfig.APIServerName)
	}

	if kubernetesConfig.DNSDomain != "" {
		flagVals = append(flagVals, "--dns-domain="+kubernetesConfig.DNSDomain)
	}

	if kubernetesConfig.NodeIP != "127.0.0.1" {
		flagVals = append(flagVals, "--node-ip="+kubernetesConfig.NodeIP)
	}

	for _, e := range kubernetesConfig.ExtraOptions {
		flagVals = append(flagVals, fmt.Sprintf("--extra-config=%s", e.String()))
	}
	flags := strings.Join(flagVals, " ")

	t := template.Must(template.New("localkubeStartCmd").Parse(localkubeStartCmdTemplate))
	buf := bytes.Buffer{}
	data := struct {
		Flags         string
		APIServerName string
	}{
		Flags:         flags,
		APIServerName: kubernetesConfig.APIServerName,
	}
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const logsTemplate = "if [[ `systemctl` =~ -\\.mount ]] &>/dev/null; " + `then
  sudo journalctl {{.Flags}} -u localkube
else
  tail -n +1 {{.Flags}} {{.RemoteLocalkubeErrPath}} {{.RemoteLocalkubeOutPath}} 
fi
`

func GetLogsCommand(follow bool) (string, error) {
	t, err := template.New("logsTemplate").Parse(logsTemplate)
	if err != nil {
		return "", err
	}
	var flags []string
	if follow {
		flags = append(flags, "-f")
	}

	buf := bytes.Buffer{}
	data := struct {
		RemoteLocalkubeErrPath string
		RemoteLocalkubeOutPath string
		Flags                  string
	}{
		RemoteLocalkubeErrPath: constants.RemoteLocalKubeErrPath,
		RemoteLocalkubeOutPath: constants.RemoteLocalKubeOutPath,
		Flags: strings.Join(flags, " "),
	}
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var localkubeStatusCommand = fmt.Sprintf("if [[ `systemctl` =~ -\\.mount ]] &>/dev/null; "+`then
  sudo systemctl is-active localkube &>/dev/null && echo "Running" || echo "Stopped"
else
  if ps $(cat %s) &>/dev/null; then
    echo "Running"
  else
    echo "Stopped"
  fi
fi
`, constants.LocalkubePIDPath)
