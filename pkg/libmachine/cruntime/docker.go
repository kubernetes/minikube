package cruntime

import (
	"bytes"
	"html/template"
)

const (
	dockerEngineConfigTemplate = `
DOCKER_OPTS='
-H tcp://0.0.0.0:{{.port}}
-H unix:///var/run/docker.sock
--storage-driver {{.EngineOptions.StorageDriver}}
--tlsverify
--tlscacert {{.AuthOptions.CaCertRemotePath}}
--tlscert {{.AuthOptions.ServerCertRemotePath}}
--tlskey {{.AuthOptions.ServerKeyRemotePath}}
{{ range .EngineOptions.Labels }}--label {{.}}
{{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}}
{{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}}
{{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}}
{{ end }}
'
{{range .EngineOptions.Env}}export \"{{ printf "%q" . }}\"
{{end}}
`
)

type DockerRuntime struct {
	EngineOptions Options
}

func (dr *DockerRuntime) GenConfigFile(opt Options) (string, error) {
	var engineCfg bytes.Buffer

	t, err := template.New("engineConfig").Parse(dockerEngineConfigTemplate)
	if err != nil {
		return "", err
	}

	t.Execute(&engineCfg, dr.EngineOptions)

	return engineCfg.String(), nil
}
