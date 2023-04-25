package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/minikube/pkg/libmachine/libmachine"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
)

var (
	// TODO: possibly move this to ssh package
	baseSSHFSArgs = []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=quiet", // suppress "Warning: Permanently added '[localhost]:2022' (ECDSA) to the list of known hosts."
	}
)

func cmdMount(c CommandLine, api libmachine.API) error {
	args := c.Args()
	if len(args) < 1 || len(args) > 2 {
		c.ShowHelp()
		return errWrongNumberArguments
	}

	src := args[0]
	dest := ""
	if len(args) > 1 {
		dest = args[1]
	}

	hostInfoLoader := &storeHostInfoLoader{api}

	cmd, err := getMountCmd(src, dest, c.Bool("unmount"), hostInfoLoader)
	if err != nil {
		return err
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func getMountCmd(src, dest string, unmount bool, hostInfoLoader HostInfoLoader) (*exec.Cmd, error) {
	var cmdPath string
	var err error
	if !unmount {
		cmdPath, err = exec.LookPath("sshfs")
		if err != nil {
			return nil, errors.New("You must have a copy of the sshfs binary locally to use the mount feature")
		}
	} else {
		cmdPath, err = exec.LookPath("fusermount")
		if err != nil {
			return nil, errors.New("You must have a copy of the fusermount binary locally to use the unmount option")
		}
	}

	srcHost, srcUser, srcPath, srcOpts, err := getInfoForSshfsArg(src, hostInfoLoader)
	if err != nil {
		return nil, err
	}

	if dest == "" {
		dest = srcPath
	}

	sshArgs := baseSSHFSArgs
	if srcHost.GetSSHKeyPath() != "" {
		sshArgs = append(sshArgs, "-o", "IdentitiesOnly=yes")
	}

	// Append needed -i / private key flags to command.
	sshArgs = append(sshArgs, srcOpts...)

	// Append actual arguments for the sshfs command (i.e. docker@<ip>:/path)
	locationArg, err := generateLocationArg(srcHost, srcUser, srcPath)
	if err != nil {
		return nil, err
	}

	if !unmount {
		sshArgs = append(sshArgs, locationArg)
		sshArgs = append(sshArgs, dest)
	} else {
		sshArgs = []string{"-u"}
		sshArgs = append(sshArgs, dest)
	}

	cmd := exec.Command(cmdPath, sshArgs...)
	log.Debug(*cmd)
	return cmd, nil
}

func getInfoForSshfsArg(hostAndPath string, hostInfoLoader HostInfoLoader) (h HostInfo, user string, path string, args []string, err error) {
	// Path with hostname.  e.g. "hostname:/usr/bin/cmatrix"
	var hostName string
	if parts := strings.SplitN(hostAndPath, ":", 2); len(parts) < 2 {
		hostName = defaultMachineName
		path = parts[0]
	} else {
		hostName = parts[0]
		path = parts[1]
	}
	if hParts := strings.SplitN(hostName, "@", 2); len(hParts) == 2 {
		user, hostName = hParts[0], hParts[1]
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
		args = append(args, "-o", fmt.Sprintf("IdentityFile=%s", h.GetSSHKeyPath()))
	}

	if user == "" {
		user = h.GetSSHUsername()
	}

	return
}
