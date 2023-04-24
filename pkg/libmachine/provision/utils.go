package provision

import (
	"fmt"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"

	"k8s.io/minikube/pkg/libmachine/auth"
)

type ContainerRuntimeOptions struct {
	EngineOptions     string
	EngineOptionsPath string
}

func installDockerGeneric(p Provisioner, baseURL string) error {
	// install docker - until cloudinit we use ubuntu everywhere so we
	// just install it using the docker repos
	if output, err := p.RunCmd(exec.Command(fmt.Sprintf("if ! type docker; then curl -sSL %s | sh -; fi", baseURL))); err != nil {
		return fmt.Errorf("error installing docker: %s", output)
	}

	return nil
}

func makeDockerOptionsDir(p Provisioner) error {
	dockerDir := p.GetCRuntimeOptionsDir()
	if _, err := p.RunCmd(exec.Command(fmt.Sprintf("sudo mkdir -p %s", dockerDir))); err != nil {
		return err
	}

	return nil
}

func setRemoteAuthOptions(p Provisioner) auth.Options {
	dockerDir := p.GetCRuntimeOptionsDir()
	authOptions := p.GetAuthOptions()

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	authOptions.CaCertRemotePath = path.Join(dockerDir, "ca.pem")
	authOptions.ServerCertRemotePath = path.Join(dockerDir, "server.pem")
	authOptions.ServerKeyRemotePath = path.Join(dockerDir, "server-key.pem")

	return authOptions
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
	statCommandOutput, err := p.RunCmd(exec.Command("stat -f -c %T " + directory))
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
		netstatOut, err := p.RunCmd(exec.Command("if ! type netstat 1>/dev/null; then ss -tln; else netstat -tln; fi"))
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
