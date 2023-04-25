package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/persist"
)

var (
	errWrongNumberArguments = errors.New("Improper number of arguments")

	// TODO: possibly move this to ssh package
	baseSSHArgs = []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=quiet", // suppress "Warning: Permanently added '[localhost]:2022' (ECDSA) to the list of known hosts."
	}
)

// HostInfo gives the mandatory information to connect to a host.
type HostInfo interface {
	GetMachineName() string

	GetSSHHostname() (string, error)

	GetSSHPort() (int, error)

	GetSSHUsername() string

	GetSSHKeyPath() string
}

// HostInfoLoader loads host information.
type HostInfoLoader interface {
	load(name string) (HostInfo, error)
}

type storeHostInfoLoader struct {
	store persist.Store
}

func (s *storeHostInfoLoader) load(name string) (HostInfo, error) {
	host, err := s.store.Load(name)
	if err != nil {
		return nil, fmt.Errorf("Error loading host: %s", err)
	}

	return host.Driver, nil
}

func getScpCmd(src, dest string, recursive bool, delta bool, quiet bool, hostInfoLoader HostInfoLoader) (*exec.Cmd, error) {
	var cmdPath string
	var err error
	if !delta {
		cmdPath, err = exec.LookPath("scp")
		if err != nil {
			return nil, errors.New("You must have a copy of the scp binary locally to use the scp feature")
		}
	} else {
		cmdPath, err = exec.LookPath("rsync")
		if err != nil {
			return nil, errors.New("You must have a copy of the rsync binary locally to use the --delta option")
		}
	}

	srcHost, srcUser, srcPath, srcOpts, err := getInfoForScpArg(src, hostInfoLoader)
	if err != nil {
		return nil, err
	}

	destHost, destUser, destPath, destOpts, err := getInfoForScpArg(dest, hostInfoLoader)
	if err != nil {
		return nil, err
	}

	// TODO: Check that "-3" flag is available in user's version of scp.
	// It is on every system I've checked, but the manual mentioned it's "newer"
	sshArgs := baseSSHArgs
	if !delta {
		sshArgs = append(sshArgs, "-3")
		if recursive {
			sshArgs = append(sshArgs, "-r")
		}
		if quiet {
			sshArgs = append(sshArgs, "-q")
		}
	}

	// Don't use ssh-agent if both hosts have explicit ssh keys
	if !missesExplicitSSHKey(srcHost) && !missesExplicitSSHKey(destHost) {
		sshArgs = append(sshArgs, "-o", "IdentitiesOnly=yes")
	}

	// Append needed -i / private key flags to command.
	sshArgs = append(sshArgs, srcOpts...)
	sshArgs = append(sshArgs, destOpts...)

	// Append actual arguments for the scp command (i.e. docker@<ip>:/path)
	locationArg, err := generateLocationArg(srcHost, srcUser, srcPath)
	if err != nil {
		return nil, err
	}

	// TODO: Check that "--progress" flag is available in user's version of rsync.
	// Use quiet mode as a workaround, if it should happen to not be supported...
	if delta {
		sshArgs = append([]string{"-e"}, "ssh "+strings.Join(sshArgs, " "))
		if !quiet {
			sshArgs = append([]string{"--progress"}, sshArgs...)
		}
		if recursive {
			sshArgs = append(sshArgs, "-r")
		}
	}

	sshArgs = append(sshArgs, locationArg)
	locationArg, err = generateLocationArg(destHost, destUser, destPath)
	if err != nil {
		return nil, err
	}
	sshArgs = append(sshArgs, locationArg)

	cmd := exec.Command(cmdPath, sshArgs...)
	log.Debug(*cmd)
	return cmd, nil
}

func missesExplicitSSHKey(hostInfo HostInfo) bool {
	return hostInfo != nil && hostInfo.GetSSHKeyPath() == ""
}

func getInfoForScpArg(hostAndPath string, hostInfoLoader HostInfoLoader) (h HostInfo, user string, path string, args []string, err error) {
	// Local path.  e.g. "/tmp/foo"
	if !strings.Contains(hostAndPath, ":") {
		return nil, "", hostAndPath, nil, nil
	}

	// Path with hostname.  e.g. "hostname:/usr/bin/cmatrix"
	parts := strings.SplitN(hostAndPath, ":", 2)
	hostName := parts[0]
	if hParts := strings.SplitN(hostName, "@", 2); len(hParts) == 2 {
		user, hostName = hParts[0], hParts[1]
	}
	path = parts[1]
	if hostName == "localhost" {
		return nil, "", path, nil, nil
	}

	// Remote path
	h, err = hostInfoLoader.load(hostName)
	if err != nil {
		return nil, "", "", nil, fmt.Errorf("Error loading host: %s", err)
	}

	args = []string{}
	port, err := h.GetSSHPort()
	if err == nil && port > 0 {
		args = append(args, "-o", fmt.Sprintf("Port=%v", port))
	}

	if h.GetSSHKeyPath() != "" {
		args = append(args, "-o", fmt.Sprintf("IdentityFile=%q", h.GetSSHKeyPath()))
	}

	return
}

func generateLocationArg(hostInfo HostInfo, user, path string) (string, error) {
	if hostInfo == nil {
		return path, nil
	}

	hostname, err := hostInfo.GetSSHHostname()
	if err != nil {
		return "", err
	}

	if user == "" {
		user = hostInfo.GetSSHUsername()
	}
	location := fmt.Sprintf("%s@%s:%s", user, hostname, path)
	return location, nil
}

func runCmdWithStdIo(cmd exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
