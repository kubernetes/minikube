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
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/sshutil"
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

func configureAuth(p miniProvisioner) error {
	log.Infof("configureAuth start")
	start := time.Now()
	defer func() {
		log.Infof("configureAuth took %s", time.Since(start))
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

	if err := copyHostCerts(authOptions); err != nil {
		return err
	}

	// The Host IP is always added to the certificate's SANs list
	hosts := append(authOptions.ServerCertSANs, ip, "localhost", "127.0.0.1")
	log.Debugf("generating server cert: %s ca-key=%s private-key=%s org=%s san=%s",
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
	log.Infof("copyHostCerts")

	err := os.MkdirAll(authOptions.StorePath, 0700)
	if err != nil {
		log.Errorf("mkdir failed: %v", err)
	}

	hostCerts := map[string]string{
		authOptions.CaCertPath:     path.Join(authOptions.StorePath, "ca.pem"),
		authOptions.ClientCertPath: path.Join(authOptions.StorePath, "cert.pem"),
		authOptions.ClientKeyPath:  path.Join(authOptions.StorePath, "key.pem"),
	}

	execRunner := command.NewExecRunner()
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
	log.Infof("copyRemoteCerts")

	remoteCerts := map[string]string{
		authOptions.CaCertPath:     authOptions.CaCertRemotePath,
		authOptions.ServerCertPath: authOptions.ServerCertRemotePath,
		authOptions.ServerKeyPath:  authOptions.ServerKeyRemotePath,
	}

	sshClient, err := sshutil.NewSSHClient(driver)
	if err != nil {
		return errors.Wrap(err, "provisioning: error getting ssh client")
	}
	sshRunner := command.NewSSHRunner(sshClient)

	dirs := []string{}
	for _, dst := range remoteCerts {
		dirs = append(dirs, path.Dir(dst))
	}

	args := append([]string{"mkdir", "-p"}, dirs...)
	if _, err = sshRunner.RunCmd(exec.Command("sudo", args...)); err != nil {
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
		return nil
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
