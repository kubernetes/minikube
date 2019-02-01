package cruntime

import (
	"bytes"
	"fmt"
	"html/template"
	"path"

	"github.com/golang/glog"
)

// CRIO contains CRIO runtime state
type CRIO struct {
	config Config
}

// Name is a human readable name for CRIO
func (r *CRIO) Name() string {
	return "CRIO"
}

// SocketPath returns the path to the socket file for CRIO
func (r *CRIO) SocketPath() string {
	if r.config.Socket != "" {
		return r.config.Socket
	}
	return "/var/run/crio/crio.sock"
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *CRIO) Available(cr CommandRunner) error {
	return cr.Run("command -v crio")
}

// Active returns if CRIO is active on the host
func (r *CRIO) Active(cr CommandRunner) bool {
	err := cr.Run("systemctl is-active --quiet service crio")
	return err == nil
}

// createConfigFile runs the commands necessary to create crictl.yaml
func (r *CRIO) createConfigFile(cr CommandRunner) error {
	var (
		crictlYamlTmpl = `runtime-endpoint: {{.RuntimeEndpoint}}
image-endpoint: {{.ImageEndpoint}}
`
		crictlYamlPath = "/etc/crictl.yaml"
	)
	t, err := template.New("crictlYaml").Parse(crictlYamlTmpl)
	if err != nil {
		return err
	}
	opts := struct {
		RuntimeEndpoint string
		ImageEndpoint   string
	}{
		RuntimeEndpoint: r.SocketPath(),
		ImageEndpoint:   r.SocketPath(),
	}
	var crictlYamlBuf bytes.Buffer
	if err := t.Execute(&crictlYamlBuf, opts); err != nil {
		return err
	}
	return cr.Run(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s",
		path.Dir(crictlYamlPath), crictlYamlBuf.String(), crictlYamlPath))
}

// Enable idempotently enables CRIO on a host
func (r *CRIO) Enable(cr CommandRunner) error {
	if err := disableOthers(r, cr); err != nil {
		glog.Warningf("disableOthers: %v", err)
	}
	if err := r.createConfigFile(cr); err != nil {
		return err
	}
	if err := enableIPForwarding(cr); err != nil {
		return err
	}
	return cr.Run("sudo systemctl restart crio")
}

// Disable idempotently disables CRIO on a host
func (r *CRIO) Disable(cr CommandRunner) error {
	return cr.Run("sudo systemctl stop crio")
}

// LoadImage loads an image into this runtime
func (r *CRIO) LoadImage(cr CommandRunner, path string) error {
	return cr.Run(fmt.Sprintf("sudo podman load -i %s", path))
}

// KubeletOptions returns kubelet options for a runtime.
func (r *CRIO) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime":          "remote",
		"container-runtime-endpoint": r.SocketPath(),
		"image-service-endpoint":     r.SocketPath(),
		"runtime-request-timeout":    "15m",
	}
}

// ListContainers returns a list of managed by this container runtime
func (r *CRIO) ListContainers(cr CommandRunner, filter string) ([]string, error) {
	return listCRIContainers(cr, filter)
}

// KillContainers removes containers based on ID
func (r *CRIO) KillContainers(cr CommandRunner, ids []string) error {
	return killCRIContainers(cr, ids)
}

// StopContainers stops containers based on ID
func (r *CRIO) StopContainers(cr CommandRunner, ids []string) error {
	return stopCRIContainers(cr, ids)
}
