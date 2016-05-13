package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

var (
	errWrongNumberArguments = errors.New("Improper number of arguments")

	// TODO: possibly move this to ssh package
	baseSSHArgs = []string{
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=quiet", // suppress "Warning: Permanently added '[localhost]:2022' (ECDSA) to the list of known hosts."
	}
)

// HostInfo gives the mandatory information to connect to a host.
type HostInfo interface {
	GetMachineName() string

	GetIP() (string, error)

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

func cmdScp(c CommandLine, api libmachine.API) error {
	args := c.Args()
	if len(args) != 2 {
		c.ShowHelp()
		return errWrongNumberArguments
	}

	src := args[0]
	dest := args[1]

	hostInfoLoader := &storeHostInfoLoader{api}

	cmd, err := getScpCmd(src, dest, c.Bool("recursive"), hostInfoLoader)
	if err != nil {
		return err
	}

	return runCmdWithStdIo(*cmd)
}

func getScpCmd(src, dest string, recursive bool, hostInfoLoader HostInfoLoader) (*exec.Cmd, error) {
	cmdPath, err := exec.LookPath("scp")
	if err != nil {
		return nil, errors.New("Error: You must have a copy of the scp binary locally to use the scp feature.")
	}

	srcHost, srcPath, srcOpts, err := getInfoForScpArg(src, hostInfoLoader)
	if err != nil {
		return nil, err
	}

	destHost, destPath, destOpts, err := getInfoForScpArg(dest, hostInfoLoader)
	if err != nil {
		return nil, err
	}

	// TODO: Check that "-3" flag is available in user's version of scp.
	// It is on every system I've checked, but the manual mentioned it's "newer"
	sshArgs := baseSSHArgs
	sshArgs = append(sshArgs, "-3")
	if recursive {
		sshArgs = append(sshArgs, "-r")
	}

	// Append needed -i / private key flags to command.
	sshArgs = append(sshArgs, srcOpts...)
	sshArgs = append(sshArgs, destOpts...)

	// Append actual arguments for the scp command (i.e. docker@<ip>:/path)
	locationArg, err := generateLocationArg(srcHost, srcPath)
	if err != nil {
		return nil, err
	}

	sshArgs = append(sshArgs, locationArg)
	locationArg, err = generateLocationArg(destHost, destPath)
	if err != nil {
		return nil, err
	}
	sshArgs = append(sshArgs, locationArg)

	cmd := exec.Command(cmdPath, sshArgs...)
	log.Debug(*cmd)
	return cmd, nil
}

func getInfoForScpArg(hostAndPath string, hostInfoLoader HostInfoLoader) (HostInfo, string, []string, error) {
	// Local path.  e.g. "/tmp/foo"
	if !strings.Contains(hostAndPath, ":") {
		return nil, hostAndPath, nil, nil
	}

	// Path with hostname.  e.g. "hostname:/usr/bin/cmatrix"
	parts := strings.SplitN(hostAndPath, ":", 2)
	hostName := parts[0]
	path := parts[1]
	if hostName == "localhost" {
		return nil, path, nil, nil
	}

	// Remote path
	hostInfo, err := hostInfoLoader.load(hostName)
	if err != nil {
		return nil, "", nil, fmt.Errorf("Error loading host: %s", err)
	}

	args := []string{}
	if hostInfo.GetSSHKeyPath() != "" {
		args = append(args, "-i", hostInfo.GetSSHKeyPath())
	}

	return hostInfo, path, args, nil
}

func generateLocationArg(hostInfo HostInfo, path string) (string, error) {
	if hostInfo == nil {
		return path, nil
	}

	ip, err := hostInfo.GetIP()
	if err != nil {
		return "", err
	}

	location := fmt.Sprintf("%s@%s:%s", hostInfo.GetSSHUsername(), ip, path)
	return location, nil
}

func runCmdWithStdIo(cmd exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
