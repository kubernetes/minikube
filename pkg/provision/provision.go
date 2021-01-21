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

package provision

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
)

// generic interface for minikube provisioner
type miniProvisioner interface {
	String() string
	CompatibleWithHost() bool
	GenerateDockerOptions(int) (*provision.DockerOptions, error)
	Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error
	GetDriver() drivers.Driver
	GetAuthOptions() auth.Options
	SSHCommand(string) (string, error)
}

// for escaping systemd template specifiers (e.g. '%i'), which are not supported by minikube
var systemdSpecifierEscaper = strings.NewReplacer("%", "%%")

func init() {
	provision.Register("Buildroot", &provision.RegisteredProvisioner{
		New: NewBuildrootProvisioner,
	})
	provision.Register("Ubuntu", &provision.RegisteredProvisioner{
		New: NewUbuntuProvisioner,
	})

}

// NewSystemdProvisioner is our fork of the same name in the upstream provision library, without the packages
func NewSystemdProvisioner(osReleaseID string, d drivers.Driver) provision.SystemdProvisioner {
	return provision.SystemdProvisioner{
		GenericProvisioner: provision.GenericProvisioner{
			SSHCommander:      provision.GenericSSHCommander{Driver: d},
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/systemd/system/docker.service.d/10-machine.conf",
			OsReleaseID:       osReleaseID,
			Driver:            d,
		},
	}
}

func configureAuth(p miniProvisioner) error {
	klog.Infof("configureAuth start")
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: configureAuth took %s", time.Since(start))
	}()

	driver := p.GetDriver()
	machineName := driver.GetMachineName()
	authOptions := p.GetAuthOptions()
	org := mcnutils.GetUsername() + "." + machineName
	bits := 2048

	ip, err := driver.GetIP()
	if err != nil {
		return errors.Wrap(err, "error getting ip during provisioning")
	}

	hostIP, err := driver.GetSSHHostname()
	if err != nil {
		return errors.Wrap(err, "error getting ssh hostname during provisioning")
	}

	if err := copyHostCerts(authOptions); err != nil {
		return err
	}

	// The Host IP is always added to the certificate's SANs list
	hosts := append(authOptions.ServerCertSANs, ip, hostIP, "localhost", "127.0.0.1", "minikube", machineName)
	klog.Infof("generating server cert: %s ca-key=%s private-key=%s org=%s san=%s",
		authOptions.ServerCertPath,
		authOptions.CaCertPath,
		authOptions.CaPrivateKeyPath,
		org,
		hosts,
	)

	err = cert.GenerateCert(&cert.Options{
		Hosts:     hosts,
		CertFile:  authOptions.ServerCertPath,
		KeyFile:   authOptions.ServerKeyPath,
		CAFile:    authOptions.CaCertPath,
		CAKeyFile: authOptions.CaPrivateKeyPath,
		Org:       org,
		Bits:      bits,
	})

	if err != nil {
		return fmt.Errorf("error generating server cert: %v", err)
	}

	return copyRemoteCerts(authOptions, driver)
}

func copyHostCerts(authOptions auth.Options) error {
	klog.Infof("copyHostCerts")

	err := os.MkdirAll(authOptions.StorePath, 0700)
	if err != nil {
		klog.Errorf("mkdir failed: %v", err)
	}

	hostCerts := map[string]string{
		authOptions.CaCertPath:     path.Join(authOptions.StorePath, "ca.pem"),
		authOptions.ClientCertPath: path.Join(authOptions.StorePath, "cert.pem"),
		authOptions.ClientKeyPath:  path.Join(authOptions.StorePath, "key.pem"),
	}

	execRunner := command.NewExecRunner(false)
	for src, dst := range hostCerts {
		f, err := assets.NewFileAsset(src, path.Dir(dst), filepath.Base(dst), "0777")
		if err != nil {
			return errors.Wrapf(err, "open cert file: %s", src)
		}
		if err := execRunner.Copy(f); err != nil {
			return errors.Wrapf(err, "transferring file: %+v", f)
		}
	}

	return nil
}

func copyRemoteCerts(authOptions auth.Options, driver drivers.Driver) error {
	klog.Infof("copyRemoteCerts")

	remoteCerts := map[string]string{
		authOptions.CaCertPath:     authOptions.CaCertRemotePath,
		authOptions.ServerCertPath: authOptions.ServerCertRemotePath,
		authOptions.ServerKeyPath:  authOptions.ServerKeyRemotePath,
	}

	sshRunner := command.NewSSHRunner(driver)

	dirs := []string{}
	for _, dst := range remoteCerts {
		dirs = append(dirs, path.Dir(dst))
	}

	args := append([]string{"mkdir", "-p"}, dirs...)
	if _, err := sshRunner.RunCmd(exec.Command("sudo", args...)); err != nil {
		return err
	}

	for src, dst := range remoteCerts {
		f, err := assets.NewFileAsset(src, path.Dir(dst), filepath.Base(dst), "0640")
		if err != nil {
			return errors.Wrapf(err, "error copying %s to %s", src, dst)
		}
		if err := sshRunner.Copy(f); err != nil {
			return errors.Wrapf(err, "transferring file to machine %v", f)
		}
	}

	return nil
}

func setRemoteAuthOptions(p provision.Provisioner) auth.Options {
	dockerDir := p.GetDockerOptionsDir()
	authOptions := p.GetAuthOptions()

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	authOptions.CaCertRemotePath = path.Join(dockerDir, "ca.pem")
	authOptions.ServerCertRemotePath = path.Join(dockerDir, "server.pem")
	authOptions.ServerKeyRemotePath = path.Join(dockerDir, "server-key.pem")

	return authOptions
}

func setContainerRuntimeOptions(name string, p miniProvisioner) error {
	c, err := config.Load(name)
	if err != nil {
		return errors.Wrap(err, "getting cluster config")
	}

	switch c.KubernetesConfig.ContainerRuntime {
	case "crio", "cri-o":
		return setCrioOptions(p)
	case "containerd":
		return setContainerdOptions(p)
	default:
		_, err := p.GenerateDockerOptions(engine.DefaultPort)
		return err
	}
}

func setCrioOptions(p provision.SSHCommander) error {
	// pass through --insecure-registry
	var (
		crioOptsTmpl = `
CRIO_MINIKUBE_OPTIONS='{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}'
`
		crioOptsPath = "/etc/sysconfig/crio.minikube"
	)
	t, err := template.New("crioOpts").Parse(crioOptsTmpl)
	if err != nil {
		return err
	}
	var crioOptsBuf bytes.Buffer
	if err := t.Execute(&crioOptsBuf, p); err != nil {
		return err
	}

	if _, err = p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s", path.Dir(crioOptsPath), crioOptsBuf.String(), crioOptsPath)); err != nil {
		return err
	}

	return nil
}

func setContainerdOptions(p provision.SSHCommander) error {
	// pass through --insecure-registry
	var (
		containerdConfigTmpl = `[plugins]
  [plugins.cri]
    [plugins.cri.registry]
      [plugins.cri.registry.mirrors]
	{{ range .EngineOptions.InsecureRegistry -}}
        [plugins.cri.registry.mirrors.\"{{. -}}\"]
		  endpoint = [\"{{. -}}\"]
        {{ end -}}`
		containerdConfigPath = "/etc/containerd/config.minikube.toml"
	)
	t, err := template.New("containerdConfigPath").Parse(containerdConfigTmpl)
	if err != nil {
		return err
	}
	var containerdConfigBuf bytes.Buffer
	if err := t.Execute(&containerdConfigBuf, p); err != nil {
		return err
	}

	if _, err = p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s", path.Dir(containerdConfigPath), containerdConfigBuf.String(), containerdConfigPath)); err != nil {
		return err
	}

	return nil
}

func rootFileSystemType(p provision.SSHCommander) (string, error) {
	fs, err := p.SSHCommand("df --output=fstype / | tail -n 1")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(fs), nil
}

// escapeSystemdDirectives escapes special characters in the input variables used to create the
// systemd unit file, which would otherwise be interpreted as systemd directives. An example
// are template specifiers (e.g. '%i') which are predefined variables that get evaluated dynamically
// (see systemd man pages for more info). This is not supported by minikube, thus needs to be escaped.
func escapeSystemdDirectives(engineConfigContext *provision.EngineConfigContext) {
	// escape '%' in Environment option so that it does not evaluate into a template specifier
	engineConfigContext.EngineOptions.Env = replaceChars(engineConfigContext.EngineOptions.Env, systemdSpecifierEscaper)
	// input might contain whitespaces, wrap it in quotes
	engineConfigContext.EngineOptions.Env = concatStrings(engineConfigContext.EngineOptions.Env, "\"", "\"")
}

// replaceChars returns a copy of the src slice with each string modified by the replacer
func replaceChars(src []string, replacer *strings.Replacer) []string {
	ret := make([]string, len(src))
	for i, s := range src {
		ret[i] = replacer.Replace(s)
	}
	return ret
}

// concatStrings concatenates each string in the src slice with prefix and postfix and returns a new slice
func concatStrings(src []string, prefix string, postfix string) []string {
	var buf bytes.Buffer
	ret := make([]string, len(src))
	for i, s := range src {
		buf.WriteString(prefix)
		buf.WriteString(s)
		buf.WriteString(postfix)
		ret[i] = buf.String()
		buf.Reset()
	}
	return ret
}

// updateUnit efficiently updates a systemd unit file
func updateUnit(p provision.SSHCommander, name string, content string, dst string) error {
	klog.Infof("Updating %s unit: %s ...", name, dst)

	if _, err := p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s.new", path.Dir(dst), content, dst)); err != nil {
		return err
	}
	if _, err := p.SSHCommand(fmt.Sprintf("sudo diff -u %s %s.new || { sudo mv %s.new %s; sudo systemctl -f daemon-reload && sudo systemctl -f enable %s && sudo systemctl -f restart %s; }", dst, dst, dst, dst, name, name)); err != nil {
		return err
	}
	return nil
}
