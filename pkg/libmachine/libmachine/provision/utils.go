/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/cert"
	"k8s.io/minikube/pkg/libmachine/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/minikube/assets"
)

func setRemoteAuthOptions(p Provisioner) *auth.Options {
	dockerDir := p.GetDockerOptionsDir()
	authOptions := p.GetAuthOptions()

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	authOptions.CaCertRemotePath = path.Join(dockerDir, "ca.pem")
	authOptions.ServerCertRemotePath = path.Join(dockerDir, "server.pem")
	authOptions.ServerKeyRemotePath = path.Join(dockerDir, "server-key.pem")

	return &authOptions
}

// x7TODO: fix this -- we're making use of getSSHHostname method from the driver
// This logic needs to be replaces by the more generic counterpart
func ConfigureAuth(p Provisioner) error {
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: configureAuth took %s", time.Since(start))
	}()

	var (
		err error
	)

	driver := p.GetDriver()
	machineName := driver.GetMachineName()
	authOptions := p.GetAuthOptions()
	// x7NOTE: minikube doesn't want this.. we may remove it in a later commit
	swarmOptions := p.GetSwarmOptions()
	org := mcnutils.GetUsername() + "." + machineName
	bits := 2048

	ip, err := driver.GetIP()
	if err != nil {
		return errors.Wrap(err, "error getting ip during provisioning")
	}

	log.Info("Copying certs to the local machine directory...")

	if err := copyHostCerts(authOptions); err != nil {
		return errors.Wrap(err, "error while copying certs into the machine")
	}

	hostIP, err := driver.GetSSHHostname()
	if err != nil {
		return errors.Wrap(err, "while getting ssh hostname during provisioning")
	}

	// The Host IP is always added to the certificate's SANs list
	hosts := append(authOptions.ServerCertSANs, ip, "localhost", "127.0.0.1", hostIP, machineName)
	log.Debugf("generating server cert: %s ca-key=%s private-key=%s org=%s san=%s",
		authOptions.ServerCertPath,
		authOptions.CaCertPath,
		authOptions.CaPrivateKeyPath,
		org,
		hosts,
	)

	// TODO: Switch to passing just authOptions to this func
	// instead of all these individual fields
	err = cert.GenerateCert(&cert.Options{
		Hosts:     hosts,
		CertFile:  authOptions.ServerCertPath,
		KeyFile:   authOptions.ServerKeyPath,
		CAFile:    authOptions.CaCertPath,
		CAKeyFile: authOptions.CaPrivateKeyPath,
		Org:       org,
		Bits:      bits,
		// x7NOTE: minikube wants this removed
		SwarmMaster: swarmOptions.Master,
	})

	if err != nil {
		return fmt.Errorf("error generating server cert: %s", err)
	}

	return copyRemoteCerts(authOptions, driver)
}

// x7NOTE: added from minikube/pkg/provision/provision.go -- may remove if not getting anywhere..
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

	execRunner := runner.NewExecRunner(false)
	for src, dst := range hostCerts {
		f, err := assets.NewFileAsset(src, path.Dir(dst), filepath.Base(dst), "0777")
		if err != nil {
			return errors.Wrapf(err, "open cert file: %s", src)
		}
		defer func() {
			if err := f.Close(); err != nil {
				klog.Warningf("error closing the file %s: %v", f.GetSourcePath(), err)
			}
		}()

		if err := execRunner.CopyFile(f); err != nil {
			return errors.Wrapf(err, "transferring file: %+v", f)
		}
	}

	return nil
}

// x7NOTE: same as above...
func copyRemoteCerts(authOptions auth.Options, driver drivers.Driver) error {
	klog.Infof("copyRemoteCerts")

	remoteCerts := map[string]string{
		authOptions.CaCertPath:     authOptions.CaCertRemotePath,
		authOptions.ServerCertPath: authOptions.ServerCertRemotePath,
		authOptions.ServerKeyPath:  authOptions.ServerKeyRemotePath,
	}

	runner, err := driver.GetRunner()
	if err != nil {
		return errors.Wrapf(err, "while getting runner")
	}

	dirs := []string{}
	for _, dst := range remoteCerts {
		dirs = append(dirs, path.Dir(dst))
	}

	args := append([]string{"mkdir", "-p"}, dirs...)
	if _, err := runner.RunCmd(exec.Command("sudo", args...)); err != nil {
		return err
	}

	for src, dst := range remoteCerts {
		f, err := assets.NewFileAsset(src, path.Dir(dst), filepath.Base(dst), "0640")
		if err != nil {
			return errors.Wrapf(err, "error copying %s to %s", src, dst)
		}
		defer func() {
			if err := f.Close(); err != nil {
				klog.Warningf("error closing the file %s: %v", f.GetSourcePath(), err)
			}
		}()

		if err := runner.CopyFile(f); err != nil {
			return errors.Wrapf(err, "transferring file to machine %v", f)
		}
	}

	return nil
}

func matchNetstatOut(reDaemonListening, netstatOut string) bool {
	// TODO: I would really prefer this be a Scanner directly on
	// the STDOUT of the executed command than to do all the string
	// manipulation hokey-pokey.
	//
	// TODO: Unit test this matching.
	for _, line := range strings.Split(netstatOut, "\n") {
		match, err := regexp.MatchString(reDaemonListening, line)
		if err != nil {
			log.Warnf("Regex warning: %s", err)
		}
		if match && line != "" {
			return true
		}
	}

	return false
}

func decideStorageDriver(p Provisioner, defaultDriver, suppliedDriver string) (string, error) {
	if suppliedDriver != "" {
		return suppliedDriver, nil
	}
	bestSuitedDriver := ""

	defer func() {
		if bestSuitedDriver != "" {
			log.Debugf("No storagedriver specified, using %s\n", bestSuitedDriver)
		}
	}()

	if defaultDriver != "aufs" {
		bestSuitedDriver = defaultDriver
	} else {
		remoteFilesystemType, err := getFilesystemType(p, "/var/lib")
		if err != nil {
			return "", err
		}
		if remoteFilesystemType == "btrfs" {
			bestSuitedDriver = "btrfs"
		} else {
			bestSuitedDriver = defaultDriver
		}
	}
	return bestSuitedDriver, nil

}

func getFilesystemType(p Provisioner, directory string) (string, error) {
	statCommandOutput, err := p.RunCmd(exec.Command("stat", "-f", "-c", "%T", directory))
	if err != nil {
		err = fmt.Errorf("Error looking up filesystem type: %s", err)
		return "", err
	}

	fstype := strings.TrimSpace(statCommandOutput.Stdout.String())
	return fstype, nil
}

func checkDaemonUp(p Provisioner, dockerPort int) func() bool {
	reDaemonListening := fmt.Sprintf(":%d\\s+.*:.*", dockerPort)
	return func() bool {
		// HACK: Check netstat's output to see if anyone's listening on the Docker API port.
		cmd := "if ! type netstat 1>/dev/null; then ss -tln; else netstat -tln; fi"
		netstatOut, err := p.RunCmd(exec.Command("bash", "-c", cmd))
		if err != nil {
			log.Warnf("Error running SSH command: %s", err)
			return false
		}

		return matchNetstatOut(reDaemonListening, netstatOut.Stdout.String())
	}
}

func WaitForDocker(p Provisioner, dockerPort int) error {
	if err := mcnutils.WaitForSpecific(checkDaemonUp(p, dockerPort), 10, 3*time.Second); err != nil {
		return NewErrDaemonAvailable(err)
	}

	return nil
}

// DockerClientVersion returns the version of the Docker client on the host
// that ssh is connected to, e.g. "1.12.1".
func DockerClientVersion(ssh Commander) (string, error) {
	// `docker version --format {{.Client.Version}}` would be preferable, but
	// that fails if the server isn't running yet.
	//
	// output is expected to be something like
	//
	//     Docker version 1.12.1, build 7a86f89
	output, err := ssh.RunCmd(exec.Command("docker", "--version"))
	if err != nil {
		return "", err
	}

	words := strings.Fields(output.Stdout.String())
	if len(words) < 3 || words[0] != "Docker" || words[1] != "version" {
		return "", fmt.Errorf("DockerClientVersion: cannot parse version string from %q", output)
	}

	return strings.TrimRight(words[2], ","), nil
}

func waitForLockAptGetUpdate(ssh Commander) error {
	return waitForLock(ssh, exec.Command("sudo", "apt-get", "update"))
}

func waitForLock(ssh Commander, cmd *exec.Cmd) error {
	var sshErr error
	err := mcnutils.WaitFor(func() bool {
		_, sshErr = ssh.RunCmd(cmd)
		if sshErr != nil {
			if strings.Contains(sshErr.Error(), "Could not get lock") {
				sshErr = nil
				return false
			}
			return true
		}
		return true
	})
	if sshErr != nil {
		return fmt.Errorf("Error running %q: %s", cmd, sshErr)
	}
	if err != nil {
		return fmt.Errorf("Failed to obtain lock: %s", err)
	}
	return nil
}
