package cruntime

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
)

// listCRIContainers returns a list of containers using crictl
func listCRIContainers(_ CommandRunner, _ string) ([]string, error) {
	// Should use crictl ps -a, but needs some massaging and testing.
	return []string{}, fmt.Errorf("unimplemented")
}

// criCRIContainers kills a list of containers using crictl
func killCRIContainers(CommandRunner, []string) error {
	return fmt.Errorf("unimplemented")
}

// StopCRIContainers stops containers using crictl
func stopCRIContainers(CommandRunner, []string) error {
	return fmt.Errorf("unimplemented")
}

// populateCRIConfig sets up /etc/crictl.yaml
func populateCRIConfig(cr CommandRunner, socket string) error {
	cPath := "/etc/crictl.yaml"
	tmpl := `runtime-endpoint: unix://{{.Socket}}
image-endpoint: unix://{{.Socket}}
`
	t, err := template.New("crictl").Parse(tmpl)
	if err != nil {
		return err
	}
	opts := struct{ Socket string }{Socket: socket}
	var b bytes.Buffer
	if err := t.Execute(&b, opts); err != nil {
		return err
	}
	return cr.Run(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s", path.Dir(cPath), b.String(), cPath))
}
